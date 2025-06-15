package tci

import (
	"fmt"
	"log"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/tci/client"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/network"
)

const retryInterval = 10 * time.Second

func NewClient(address string, trx int, bandplan bandplan.Bandplan) (*Client, error) {
	host, err := network.ParseTCPAddr(address)
	if err != nil {
		return nil, err
	}

	result := &Client{
		bandplan: bandplan,
	}
	result.trx = &trxListener{
		client: result,
		trx:    trx,
	}
	result.resetSpots()
	result.client = client.KeepOpen(host, retryInterval, false)
	result.client.Notify(result.trx)

	return result, nil
}

type Client struct {
	client   *client.Client
	bandplan bandplan.Bandplan

	sendSpots      bool
	lastHeardSpots map[string]time.Time

	trx       *trxListener
	connected bool

	listeners []any
}

func (c *Client) Connect() error {
	if !c.connected {
		return fmt.Errorf("cannot connect to TCI host")
	}
	return nil
}

func (c *Client) Disconnect() {
	c.client.Disconnect()
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) Active() bool {
	return c.connected
}

func (c *Client) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) emitConnectionChanged(connected bool) {
	type listenerType interface {
		ConnectionChanged(bool)
	}
	for _, listener := range c.listeners {
		if typedListener, ok := listener.(listenerType); ok {
			typedListener.ConnectionChanged(connected)
		}
	}
}

func (c *Client) emitFrequencyChanged(f core.Frequency) {
	for _, listener := range c.listeners {
		if frequencyListener, ok := listener.(core.VFOFrequencyListener); ok {
			frequencyListener.VFOFrequencyChanged(f)
		}
	}
}

func (c *Client) emitBandChanged(b core.Band) {
	for _, listener := range c.listeners {
		if bandListener, ok := listener.(core.VFOBandListener); ok {
			bandListener.VFOBandChanged(b)
		}
	}
}

func (c *Client) emitModeChanged(m core.Mode) {
	for _, listener := range c.listeners {
		if modeListener, ok := listener.(core.VFOModeListener); ok {
			modeListener.VFOModeChanged(m)
		}
	}
}

func (c *Client) emitXITChanged(active bool, offset core.Frequency) {
	for _, listener := range c.listeners {
		if xitListener, ok := listener.(core.VFOXITListener); ok {
			xitListener.VFOXITChanged(active, offset)
		}
	}
}

func (c *Client) emitPTTChanged(active bool) {
	for _, listener := range c.listeners {
		if xitListener, ok := listener.(core.VFOPTTListener); ok {
			xitListener.VFOPTTChanged(active)
		}
	}
}

func (c *Client) Speed(wpm int) {
	err := c.client.SetCWMacrosSpeed(wpm)
	if err != nil {
		log.Printf("cannot set CW speed: %v", err)
	}
}

func (c *Client) Send(text string) {
	err := c.client.SendCWMacro(c.trx.trx, text)
	if err != nil {
		log.Printf("cannot send CW: %v", err)
	}
}

func (c *Client) Abort() {
	err := c.client.StopCW()
	if err != nil {
		log.Printf("cannot abort CW: %v", err)
	}
}

func (c *Client) SetFrequency(frequency core.Frequency) {
	err := c.client.SetVFOFrequency(c.trx.trx, client.VFOA, int(frequency))
	if err != nil {
		log.Printf("cannot set VFO frequency: %v", err)
	}
}

func (c *Client) SetBand(band core.Band) {
	bandplanBand := c.bandplan[toBandplanBandName(band)]
	frequency := findModePortionCenter(c.bandplan, int(bandplanBand.Center()), toBandplanMode(c.trx.mode))
	err := c.client.SetVFOFrequency(c.trx.trx, client.VFOA, frequency)
	if err != nil {
		log.Printf("cannot switch to band %s: %v", band, err)
	}
}

func (c *Client) SetMode(mode core.Mode) {
	err := c.client.SetMode(c.trx.trx, toClientMode(mode))
	if err != nil {
		log.Printf("cannot set mode: %v", err)
	}
}

func (c *Client) SetXIT(active bool, offset core.Frequency) {
	err := c.client.SetXITEnable(c.trx.trx, active)
	if err != nil {
		log.Printf("cannot enable XIT: %v", err)
		return
	}

	err = c.client.SetXITOffset(c.trx.trx, int(offset))
	if err != nil {
		log.Printf("cannot set XIT offset: %v", err)
		return
	}
}

func (c *Client) Refresh() {
	c.trx.Refresh()
}

var spotColors = map[core.SpotType]client.ARGB{
	core.WorkedSpot:  client.NewARGB(255, 128, 128, 128),
	core.ManualSpot:  client.NewARGB(255, 255, 255, 255),
	core.SkimmerSpot: client.NewARGB(255, 255, 153, 255),
	core.RBNSpot:     client.NewARGB(255, 255, 255, 153),
	core.ClusterSpot: client.NewARGB(255, 153, 255, 255),
}

func (c *Client) SetSendSpots(sendSpots bool) {
	c.sendSpots = sendSpots
	c.resetSpots()
}

func (c *Client) resetSpots() {
	c.lastHeardSpots = make(map[string]time.Time)
}

func (c *Client) EntryAdded(entry core.BandmapEntry) {
	if !c.sendSpots {
		return
	}
	if !c.client.Connected() {
		return
	}

	if entry.Band != c.trx.band || entry.Mode != c.trx.mode {
		return
	}

	lastHeard, ok := c.lastHeardSpots[entry.Call.String()]
	if ok && !lastHeard.Before(entry.LastHeard) {
		return
	}

	// log.Printf("TCI: adding spot %s", entry.Call)
	c.lastHeardSpots[entry.Call.String()] = entry.LastHeard
	err := c.client.AddSpot(entry.Call.String(), toClientMode(entry.Mode), int(entry.Frequency), spotColors[entry.Source], "hellocontest")
	if err != nil {
		log.Printf("TCI: cannot add spot: %v", err)
	}
}

func (c *Client) EntryUpdated(entry core.BandmapEntry) {
	c.EntryAdded(entry)
}

func (c *Client) EntryRemoved(entry core.BandmapEntry) {
	if !c.sendSpots {
		return
	}

	err := c.client.DeleteSpot(entry.Call.String())
	if err != nil {
		log.Printf("TCI: cannot delete spot: %v", err)
	}
}

type trxListener struct {
	client    *Client
	trx       int
	frequency core.Frequency
	band      core.Band
	mode      core.Mode
	xitActive bool
	xitOffset core.Frequency
	ptt       bool
}

func (l *trxListener) Refresh() {
	l.client.emitFrequencyChanged(l.frequency)
	l.client.emitBandChanged(l.band)
	l.client.emitModeChanged(l.mode)
	l.client.emitXITChanged(l.xitActive, l.xitOffset)
	l.client.emitPTTChanged(l.ptt)
}

func (l *trxListener) Connected(connected bool) {
	l.client.connected = connected
	l.client.emitConnectionChanged(connected)
}

func (l *trxListener) SetVFOFrequency(trx int, vfo client.VFO, frequency int) {
	if trx != l.trx || vfo != client.VFOA {
		return
	}
	incomingFrequency := core.Frequency(frequency)
	if l.frequency == incomingFrequency {
		return
	}
	l.frequency = incomingFrequency
	l.client.emitFrequencyChanged(l.frequency)
	// log.Printf("incoming frequency: %s", l.frequency)

	band := l.client.bandplan.ByFrequency(hamradio.Frequency(frequency))
	incomingBand := toCoreBand(band.Name)
	if incomingBand == l.band {
		return
	}
	l.band = incomingBand
	l.client.resetSpots()
	l.client.emitBandChanged(l.band)
	// log.Printf("incoming band: %v", l.band)
}

func (l *trxListener) SetMode(trx int, mode client.Mode) {
	if trx != l.trx {
		return
	}
	incomingMode := toCoreMode(mode)
	if incomingMode == l.mode {
		return
	}
	l.mode = incomingMode
	l.client.emitModeChanged(l.mode)
	// log.Printf("incoming mode %v", incomingMode)
}

func (l *trxListener) SetXITEnable(trx int, active bool) {
	if trx != l.trx {
		return
	}
	incomingActive := active
	if incomingActive == l.xitActive {
		return
	}
	l.xitActive = incomingActive
	l.client.emitXITChanged(l.xitActive, l.xitOffset)
	// log.Printf("incoming XIT active %v", incomingActive)
}

func (l *trxListener) SetXITOffset(trx int, offset int) {
	if trx != l.trx {
		return
	}
	incomingOffset := core.Frequency(offset)
	if incomingOffset == l.xitOffset {
		return
	}
	l.xitOffset = incomingOffset
	l.client.emitXITChanged(l.xitActive, l.xitOffset)
	// log.Printf("incoming XIT offset %v", incomingOffset)
}

func (l *trxListener) SetTX(trx int, enable bool) {
	if trx != l.trx {
		return
	}
	if enable == l.ptt {
		return
	}
	l.ptt = enable
	l.client.emitPTTChanged(l.ptt)
	// log.Printf("incoming PTT %v", enable)
}

func toCoreBand(bandName bandplan.BandName) core.Band {
	if bandName == bandplan.BandUnknown {
		return core.NoBand
	}
	return core.Band(bandName)
}

func toBandplanBandName(band core.Band) bandplan.BandName {
	if band == core.NoBand {
		return bandplan.BandUnknown
	}
	return bandplan.BandName(band)
}

func toCoreMode(mode client.Mode) core.Mode {
	switch mode {
	case client.ModeUSB, client.ModeLSB:
		return core.ModeSSB
	case client.ModeCW:
		return core.ModeCW
	case client.ModeNFM, client.ModeWFM:
		return core.ModeFM
	case client.ModeDIGL, client.ModeDIGU, client.ModeSPEC:
		return core.ModeDigital
	default:
		return core.NoMode
	}
}

func toClientMode(mode core.Mode) client.Mode {
	switch mode {
	case core.ModeCW:
		return client.ModeCW
	case core.ModeSSB:
		return client.ModeUSB // TODO make this dependent of the current frequency either LSB or USB
	case core.ModeFM:
		return client.ModeNFM
	case core.ModeRTTY:
		return client.ModeDIGU
	case core.ModeDigital:
		return client.ModeDIGU
	default:
		return client.ModeSPEC
	}
}

func toBandplanMode(mode core.Mode) bandplan.Mode {
	switch mode {
	case core.ModeCW:
		return bandplan.ModeCW
	case core.ModeSSB, core.ModeFM:
		return bandplan.ModePhone
	case core.ModeDigital, core.ModeRTTY:
		return bandplan.ModeDigital
	default:
		return bandplan.ModeDigital
	}
}

func findModePortionCenter(bp bandplan.Bandplan, f int, mode bandplan.Mode) int {
	frequency := hamradio.Frequency(f)
	band := bp.ByFrequency(frequency)
	var modePortion bandplan.Portion
	var currentPortion bandplan.Portion
	for _, portion := range band.Portions {
		if (portion.Mode == mode && portion.From < frequency) || modePortion.Mode != mode {
			modePortion = portion
		}
		if portion.Contains(frequency) {
			currentPortion = portion
		}
		if modePortion.Mode == mode && currentPortion.Mode != "" {
			break
		}
	}
	if currentPortion.Mode == mode {
		return int(currentPortion.Center())
	}
	if modePortion.Mode == mode {
		return int(modePortion.Center())
	}
	return int(band.Center())
}
