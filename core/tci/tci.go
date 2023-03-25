package tci

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/tci/client"

	"github.com/ftl/hellocontest/core"
)

const retryInterval = 10 * time.Second

type VFOController interface {
	SetFrequency(core.Frequency)
	SetBand(core.Band)
	SetMode(core.Mode)
}

func NewClient(address string) (*Client, error) {
	host, err := parseTCPAddr(address)
	if err != nil {
		return nil, err
	}

	result := &Client{
		controller: new(nullController),
		bandplan:   bandplan.IARURegion1,
	}
	result.trx = &trxListener{
		client: result,
	}
	result.client = client.KeepOpen(host, retryInterval, false)
	result.client.Notify(result.trx)

	return result, nil
}

type Client struct {
	client     *client.Client
	controller VFOController
	bandplan   bandplan.Bandplan

	trx       *trxListener
	connected bool

	listeners []interface{}
}

func (c *Client) Disconnect() {
	c.client.Disconnect()
}

func (c *Client) SetVFOController(controller VFOController) {
	if controller == nil {
		c.controller = new(nullController)
	}
	c.controller = controller
}

func (c *Client) Notify(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) emitStatusChanged(available bool) {
	for _, listener := range c.listeners {
		if serviceStatusListener, ok := listener.(core.ServiceStatusListener); ok {
			serviceStatusListener.StatusChanged(core.TCIService, available)
			serviceStatusListener.StatusChanged(core.CWDaemonService, available)
		}
	}
}

func (c *Client) Connect() error {
	if !c.connected {
		return fmt.Errorf("cannot connect to TCI host")
	}
	return nil
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) Active() bool {
	return c.connected
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
	frequency := findModePortionCenter(int(bandplanBand.Center()), toBandplanMode(c.trx.mode))
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

func (c *Client) Refresh() {
	c.trx.Refresh()
}

var spotColors = map[core.SpotSource]client.ARGB{
	core.ManualSpot:  client.NewARGB(255, 255, 255, 255),
	core.SkimmerSpot: client.NewARGB(255, 255, 153, 255),
	core.RBNSpot:     client.NewARGB(255, 255, 255, 153),
	core.ClusterSpot: client.NewARGB(255, 153, 255, 255),
}

func (c *Client) EntryAdded(e core.BandmapEntry) {
	err := c.client.AddSpot(e.Call.String(), toClientMode(e.Mode), int(e.Frequency), spotColors[e.Source], "hellocontest")
	if err != nil {
		log.Printf("cannot add spot: %v", err)
	}
}

func (c *Client) EntryUpdated(e core.BandmapEntry) {
	c.EntryRemoved(e)
	c.EntryAdded(e)
}

func (c *Client) EntryRemoved(e core.BandmapEntry) {
	err := c.client.DeleteSpot(e.Call.String())
	if err != nil {
		log.Printf("cannot delete spot: %v", err)
	}
}

type trxListener struct {
	client    *Client
	trx       int
	frequency core.Frequency
	band      core.Band
	mode      core.Mode
}

func (l *trxListener) Refresh() {
	l.client.controller.SetFrequency(l.frequency)
	l.client.controller.SetBand(l.band)
	l.client.controller.SetMode(l.mode)
}

func (l *trxListener) Connected(connected bool) {
	l.client.connected = connected
	l.client.emitStatusChanged(connected)
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
	l.client.controller.SetFrequency(l.frequency)
	log.Printf("incoming frequency: %s", l.frequency)

	band := l.client.bandplan.ByFrequency(hamradio.Frequency(frequency))
	incomingBand := toCoreBand(band.Name)
	if incomingBand == l.band {
		return
	}
	l.band = incomingBand
	l.client.controller.SetBand(l.band)
	log.Printf("incoming band: %v", l.band)

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
	l.client.controller.SetMode(l.mode)
	log.Printf("incoming mode %v", incomingMode)
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

func findModePortionCenter(f int, mode bandplan.Mode) int {
	frequency := hamradio.Frequency(f)
	band := bandplan.IARURegion1.ByFrequency(frequency)
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

func parseTCPAddr(arg string) (*net.TCPAddr, error) {
	host, port := splitHostPort(arg)
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = strconv.Itoa(client.DefaultPort)
	}

	return net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host, port))
}

func splitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

type nullController struct{}

func (*nullController) SetFrequency(core.Frequency) {}
func (*nullController) SetBand(core.Band)           {}
func (*nullController) SetMode(core.Mode)           {}
