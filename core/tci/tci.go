package tci

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/tci/client"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/network"
)

const retryInterval = 10 * time.Second

func NewClient(address string, trx int, trx2 int, bandplan bandplan.Bandplan) (*Client, error) {
	host, err := network.ParseTCPAddr(address)
	if err != nil {
		return nil, err
	}

	result := &Client{
		bandplan:  bandplan,
		singleVFO: true,
	}
	result.trx = &trxListener{
		client: result,
		trx:    trx,
		vfo:    core.VFO1,
	}
	if trx2 >= 0 && trx2 != trx {
		result.singleVFO = false
		result.trx2 = &trxListener{
			client: result,
			trx:    trx2,
			vfo:    core.VFO2,
		}
	}
	result.resetSpots()
	result.client = client.KeepOpen(host, retryInterval, false)
	result.client.Notify(result.trx)
	if !result.singleVFO {
		result.client.Notify(result.trx2)
	}

	return result, nil
}

type Client struct {
	client   *client.Client
	bandplan bandplan.Bandplan

	singleVFO bool

	sendSpots      bool
	lastHeardSpots map[string]time.Time
	sentMarkers    map[string]core.BandmapMarker

	trx       *trxListener
	trx2      *trxListener
	connected bool

	currentVFO core.VFOID
	txVFO      core.VFOID

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

func (c *Client) SingleVFO() bool {
	return c.singleVFO
}

func (c *Client) toTCITRX(vfo core.VFOID) int {
	if !c.singleVFO && vfo == core.VFO2 {
		return c.trx2.trx
	}
	return c.trx.trx
}

func (c *Client) SetCurrentVFO(vfo core.VFOID) {
	if c.singleVFO {
		c.currentVFO = core.VFO1
		return
	}
	c.currentVFO = vfo
}

func (c *Client) SetTXVFO(vfo core.VFOID) {
	if c.singleVFO {
		c.txVFO = core.VFO1
		return
	}
	c.txVFO = vfo
	c.emitTXVFOChanged(vfo)
}

func (c *Client) MuteAudio(_ core.VFOID) {
	// TCI can only mute audio completely
	err := c.client.SetMute(true)
	if err != nil {
		log.Printf("tci: cannot mute: %v", err)
	}
}

func (c *Client) UnmuteAudio(_ core.VFOID) {
	// TCI can only mute audio completely
	err := c.client.SetMute(false)
	if err != nil {
		log.Printf("tci: cannot unmute: %v", err)
	}
}

func (c *Client) ToggleAudio(_ core.VFOID) {
	// TCI can only mute audio completely
	muted, err := c.client.Mute()
	if err != nil {
		log.Printf("tci: cannot read mute state: %v", err)
	}
	err = c.client.SetMute(!muted)
	if err != nil {
		log.Printf("tci: cannot set mute state: %v", err)
	}
}

func (c *Client) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) emitConnectionChanged(connected bool) {
	core.Emit(c.listeners, func(listener core.ConnectionChangedListener) {
		listener.ConnectionChanged(connected)
	})
}

func (c *Client) emitFrequencyChanged(vfo core.VFOID, frequency core.Frequency) {
	core.Emit(c.listeners, func(listener core.VFOFrequencyListener) {
		listener.VFOFrequencyChanged(vfo, frequency)
	})
}

func (c *Client) emitBandChanged(vfo core.VFOID, band core.Band) {
	core.Emit(c.listeners, func(listener core.VFOBandListener) {
		listener.VFOBandChanged(vfo, band)
	})
}

func (c *Client) emitModeChanged(vfo core.VFOID, mode core.Mode) {
	core.Emit(c.listeners, func(listener core.VFOModeListener) {
		listener.VFOModeChanged(vfo, mode)
	})
}

func (c *Client) emitIncrementalTuningChanged(vfo core.VFOID, kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	core.Emit(c.listeners, func(listener core.VFOIncrementalTuningListener) {
		listener.VFOIncrementalTuningChanged(vfo, kind, active, offset)
	})
}

func (c *Client) emitPTTChanged(vfo core.VFOID, active bool) {
	core.Emit(c.listeners, func(listener core.VFOPTTListener) {
		listener.VFOPTTChanged(vfo, active)
	})
}

func (c *Client) emitTXVFOChanged(vfo core.VFOID) {
	core.Emit(c.listeners, func(listener core.TXVFOListener) {
		listener.TXVFOChanged(vfo)
	})
}

func (c *Client) Speed(wpm int) {
	err := c.client.SetCWMacrosSpeed(wpm)
	if err != nil {
		log.Printf("cannot set CW speed: %v", err)
	}
}

func (c *Client) Send(text string) {
	err := c.client.SendCWMacro(c.toTCITRX(c.txVFO), text)
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

func (c *Client) SetFrequency(vfo core.VFOID, frequency core.Frequency) {
	err := c.client.SetVFOFrequency(c.toTCITRX(vfo), client.VFOA, int(frequency))
	if err != nil {
		log.Printf("cannot set VFO frequency: %v", err)
	}
}

func (c *Client) SetBand(vfo core.VFOID, band core.Band) {
	bandplanBand := c.bandplan[toBandplanBandName(band)]
	frequency := findModePortionCenter(c.bandplan, int(bandplanBand.Center()), toBandplanMode(c.modeOf(vfo)))
	log.Printf("tci: switching TRX%d to %s at %d", c.toTCITRX(vfo), band, frequency)
	err := c.client.SetVFOFrequency(c.toTCITRX(vfo), client.VFOA, frequency)
	if err != nil {
		log.Printf("cannot switch to band %s: %v", band, err)
	}
}

func (c *Client) modeOf(vfo core.VFOID) core.Mode {
	if !c.singleVFO && vfo == core.VFO2 {
		return c.trx2.mode
	}
	return c.trx.mode
}

func (c *Client) SetMode(vfo core.VFOID, mode core.Mode) {
	err := c.client.SetMode(c.toTCITRX(vfo), toClientMode(mode))
	if err != nil {
		log.Printf("cannot set mode: %v", err)
	}
}

func (c *Client) SetIncrementalTuning(vfo core.VFOID, kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	setEnable := c.client.SetXITEnable
	setOffset := c.client.SetXITOffset
	if kind == core.RIT {
		setEnable = c.client.SetRITEnable
		setOffset = c.client.SetRITOffset
	}

	err := setEnable(c.toTCITRX(vfo), active)
	if err != nil {
		log.Printf("cannot enable %s: %v", kind, err)
		return
	}

	err = setOffset(c.toTCITRX(vfo), int(offset))
	if err != nil {
		log.Printf("cannot set %s offset: %v", kind, err)
		return
	}
}

func (c *Client) Refresh() {
	c.trx.Refresh()
	if !c.singleVFO {
		c.trx2.Refresh()
	}
}

var spotColors = map[core.SpotType]client.ARGB{
	core.WorkedSpot:  client.NewARGB(255, 128, 128, 128),
	core.ManualSpot:  client.NewARGB(255, 255, 255, 255),
	core.SkimmerSpot: client.NewARGB(255, 255, 255, 255),
	core.RBNSpot:     client.NewARGB(255, 255, 255, 255),
	core.ClusterSpot: client.NewARGB(255, 255, 255, 255),
}

// a marker is no spot, therefore it has its own colors: the CQ frequency in dark red, the
// numbered and text markers in the light blue of the spot list
var markerColors = map[core.BandmapMarkerKind]client.ARGB{
	core.CQMarker:       client.NewARGB(255, 139, 0, 0),
	core.NumberedMarker: client.NewARGB(255, 173, 216, 230),
	core.TextMarker:     client.NewARGB(255, 173, 216, 230),
}

// SetBandplan swaps the bandplan used for band lookups. Subsequent lookups use
// the new plan.
func (c *Client) SetBandplan(bandplan bandplan.Bandplan) {
	c.bandplan = bandplan
}

func (c *Client) SetSendSpots(sendSpots bool) {
	c.sendSpots = sendSpots
	c.resetSpots()
}

// TCI shows the spots and the markers on the panorama of every TRX, therefore every TRX
// that this client uses contributes its band and mode
func (c *Client) activeTRXs() []*trxListener {
	if c.singleVFO {
		return []*trxListener{c.trx}
	}
	return []*trxListener{c.trx, c.trx2}
}

func (c *Client) activeBands() []core.Band {
	trxs := c.activeTRXs()
	result := make([]core.Band, 0, len(trxs))
	for _, trx := range trxs {
		result = append(result, trx.band)
	}
	return result
}

func (c *Client) trxOnBand(band core.Band) (*trxListener, bool) {
	for _, trx := range c.activeTRXs() {
		if trx.band == band {
			return trx, true
		}
	}
	return nil, false
}

func (c *Client) spotVisible(band core.Band, mode core.Mode) bool {
	for _, trx := range c.activeTRXs() {
		if trx.band == band && trx.mode == mode {
			return true
		}
	}
	return false
}

func (c *Client) resetSpots() {
	c.lastHeardSpots = make(map[string]time.Time)
	c.sentMarkers = make(map[string]core.BandmapMarker)
}

func (c *Client) EntryAdded(entry core.BandmapEntry) {
	if !c.sendSpots {
		return
	}
	if !c.client.Connected() {
		return
	}

	if !c.spotVisible(entry.Band, entry.Mode) {
		return
	}

	lastHeard, ok := c.lastHeardSpots[entry.Call.String()]
	if ok && !lastHeard.Before(entry.LastHeard) {
		return
	}

	c.lastHeardSpots[entry.Call.String()] = entry.LastHeard
	err := c.client.AddSpot(entry.Call.String(), toClientMode(entry.Mode), int(entry.Frequency), spotColors[entry.Source], "hellocontest")
	if err != nil && !isNotConnectedError(err) {
		log.Printf("TCI: cannot add spot: %v", err)
	}
}

func (c *Client) MarkersChanged(markers []core.BandmapMarker) {
	if !c.sendSpots {
		return
	}
	if !c.client.Connected() {
		return
	}

	currentMarkers, addedMarkers, removedMarkers := diffMarkers(c.sentMarkers, markers, c.activeBands())
	c.addMarkers(addedMarkers)
	c.deleteMarkers(removedMarkers)
	c.sentMarkers = currentMarkers
}

func (c *Client) addMarkers(markers []core.BandmapMarker) {
	for _, marker := range markers {
		// a marker has no mode, therefore the mode of the TRX on that band decides
		trx, onBand := c.trxOnBand(marker.Band)
		if !onBand {
			continue
		}
		err := c.client.AddSpot(marker.Text, toClientMode(trx.mode), int(marker.Frequency), markerColors[marker.Kind], "hellocontest")
		if err != nil && !isNotConnectedError(err) {
			log.Printf("TCI: cannot add marker: %v", err)
		}
	}
}

func (c *Client) deleteMarkers(texts []string) {
	for _, text := range texts {
		err := c.client.DeleteSpot(text)
		if err != nil && !isNotConnectedError(err) {
			log.Printf("TCI: cannot delete marker: %v", err)
		}
	}
}

// diffMarkers compares the markers on the bands of the TRXs with the markers that the
// client sent before, therefore the client sends only the changes.
func diffMarkers(sentMarkers map[string]core.BandmapMarker, markers []core.BandmapMarker, bands []core.Band) (map[string]core.BandmapMarker, []core.BandmapMarker, []string) {
	currentMarkers := make(map[string]core.BandmapMarker, len(markers))
	addedMarkers := make([]core.BandmapMarker, 0, len(markers))
	for _, marker := range markers {
		if !slices.Contains(bands, marker.Band) {
			continue
		}
		currentMarkers[marker.Text] = marker

		sentMarker, alreadySent := sentMarkers[marker.Text]
		if alreadySent && sentMarker.Frequency == marker.Frequency {
			continue
		}
		addedMarkers = append(addedMarkers, marker)
	}

	removedMarkers := make([]string, 0, len(sentMarkers))
	for text := range sentMarkers {
		_, stillThere := currentMarkers[text]
		if stillThere {
			continue
		}
		removedMarkers = append(removedMarkers, text)
	}
	slices.Sort(removedMarkers)

	return currentMarkers, addedMarkers, removedMarkers
}

func (c *Client) EntryUpdated(entry core.BandmapEntry) {
	c.EntryAdded(entry)
}

func (c *Client) EntryRemoved(entry core.BandmapEntry) {
	if !c.sendSpots {
		return
	}

	err := c.client.DeleteSpot(entry.Call.String())
	if err != nil && !isNotConnectedError(err) {
		log.Printf("TCI: cannot delete spot: %v", err)
	}
}

func isNotConnectedError(err error) bool {
	return strings.HasSuffix(err.Error(), "not connected")
}

type trxListener struct {
	client    *Client
	trx       int
	vfo       core.VFOID
	frequency core.Frequency
	band      core.Band
	mode      core.Mode
	xitActive bool
	xitOffset core.Frequency
	ritActive bool
	ritOffset core.Frequency
	ptt       bool
}

func (l *trxListener) Refresh() {
	l.client.emitFrequencyChanged(l.vfo, l.frequency)
	l.client.emitBandChanged(l.vfo, l.band)
	l.client.emitModeChanged(l.vfo, l.mode)
	l.client.emitIncrementalTuningChanged(l.vfo, core.XIT, l.xitActive, l.xitOffset)
	l.client.emitIncrementalTuningChanged(l.vfo, core.RIT, l.ritActive, l.ritOffset)
	l.client.emitPTTChanged(l.vfo, l.ptt)
}

func (l *trxListener) Connected(connected bool) {
	l.client.connected = connected
	l.client.emitConnectionChanged(connected)
}

func (l *trxListener) SetTRXCount(count int) {
	log.Printf("TCI TRX COUNT: %d", count)
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
	l.client.emitFrequencyChanged(l.vfo, l.frequency)
	// log.Printf("incoming frequency: %s", l.frequency)

	band := l.client.bandplan.ByFrequency(hamradio.Frequency(frequency))
	incomingBand := toCoreBand(band.Name)
	if incomingBand == l.band {
		return
	}
	l.band = incomingBand
	l.client.resetSpots()
	l.client.emitBandChanged(l.vfo, l.band)
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
	l.client.emitModeChanged(l.vfo, l.mode)
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
	l.client.emitIncrementalTuningChanged(l.vfo, core.XIT, l.xitActive, l.xitOffset)
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
	l.client.emitIncrementalTuningChanged(l.vfo, core.XIT, l.xitActive, l.xitOffset)
}

func (l *trxListener) SetRITEnable(trx int, active bool) {
	if trx != l.trx {
		return
	}
	if active == l.ritActive {
		return
	}
	l.ritActive = active
	l.client.emitIncrementalTuningChanged(l.vfo, core.RIT, l.ritActive, l.ritOffset)
}

func (l *trxListener) SetRITOffset(trx int, offset int) {
	if trx != l.trx {
		return
	}
	incomingOffset := core.Frequency(offset)
	if incomingOffset == l.ritOffset {
		return
	}
	l.ritOffset = incomingOffset
	l.client.emitIncrementalTuningChanged(l.vfo, core.RIT, l.ritActive, l.ritOffset)
}

func (l *trxListener) SetTX(trx int, enable bool) {
	if trx != l.trx {
		return
	}
	if enable == l.ptt {
		return
	}
	l.ptt = enable
	l.client.emitPTTChanged(l.vfo, l.ptt)
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
