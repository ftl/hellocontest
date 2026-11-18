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

	"github.com/ftl/hellocontest/core/model"
	"github.com/ftl/hellocontest/core/network"
)

const retryInterval = 10 * time.Second

func NewClient(address string, trx int, trx2 int, bandplan bandplan.Bandplan, asyncRunner model.AsyncRunner) (*Client, error) {
	host, err := network.ParseTCPAddr(address)
	if err != nil {
		return nil, err
	}
	if asyncRunner == nil {
		asyncRunner = func(f func()) { f() }
	}

	result := &Client{
		bandplan:    bandplan,
		singleVFO:   true,
		asyncRunner: asyncRunner,
	}
	result.trx = &trxListener{
		client: result,
		trx:    trx,
		vfo:    model.VFO1,
	}
	if trx2 >= 0 && trx2 != trx {
		result.singleVFO = false
		result.trx2 = &trxListener{
			client: result,
			trx:    trx2,
			vfo:    model.VFO2,
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
	client      *client.Client
	bandplan    bandplan.Bandplan
	asyncRunner model.AsyncRunner

	singleVFO bool

	sendSpots      bool
	lastHeardSpots map[string]time.Time
	sentMarkers    map[string]model.BandmapMarker

	trx       *trxListener
	trx2      *trxListener
	connected bool

	currentVFO model.VFOID
	txVFO      model.VFOID

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

func (c *Client) toTCITRX(vfo model.VFOID) int {
	if !c.singleVFO && vfo == model.VFO2 {
		return c.trx2.trx
	}
	return c.trx.trx
}

func (c *Client) SetCurrentVFO(vfo model.VFOID) {
	if c.singleVFO {
		c.currentVFO = model.VFO1
		return
	}
	c.currentVFO = vfo
}

func (c *Client) SetTXVFO(vfo model.VFOID) {
	if c.singleVFO {
		c.txVFO = model.VFO1
		return
	}
	c.txVFO = vfo
	c.emitTXVFOChanged(vfo)
}

func (c *Client) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) emitConnectionChanged(connected bool) {
	c.asyncRunner(func() {
		model.Emit(c.listeners, func(listener model.ConnectionChangedListener) {
			listener.ConnectionChanged(connected)
		})
	})
}

func (c *Client) emitFrequencyChanged(vfo model.VFOID, frequency model.Frequency) {
	c.asyncRunner(func() {
		model.Emit(c.listeners, func(listener model.VFOFrequencyListener) {
			listener.VFOFrequencyChanged(vfo, frequency)
		})
	})
}

func (c *Client) emitBandChanged(vfo model.VFOID, band model.Band) {
	c.asyncRunner(func() {
		model.Emit(c.listeners, func(listener model.VFOBandListener) {
			listener.VFOBandChanged(vfo, band)
		})
	})
}

func (c *Client) emitModeChanged(vfo model.VFOID, mode model.Mode) {
	c.asyncRunner(func() {
		model.Emit(c.listeners, func(listener model.VFOModeListener) {
			listener.VFOModeChanged(vfo, mode)
		})
	})
}

func (c *Client) emitIncrementalTuningChanged(vfo model.VFOID, kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	c.asyncRunner(func() {
		model.Emit(c.listeners, func(listener model.VFOIncrementalTuningListener) {
			listener.VFOIncrementalTuningChanged(vfo, kind, active, offset)
		})
	})
}

func (c *Client) emitPTTChanged(vfo model.VFOID, active bool) {
	c.asyncRunner(func() {
		model.Emit(c.listeners, func(listener model.VFOPTTListener) {
			listener.VFOPTTChanged(vfo, active)
		})
	})
}

func (c *Client) emitTXVFOChanged(vfo model.VFOID) {
	c.asyncRunner(func() {
		model.Emit(c.listeners, func(listener model.TXVFOListener) {
			listener.TXVFOChanged(vfo)
		})
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

func (c *Client) SetFrequency(vfo model.VFOID, frequency model.Frequency) {
	err := c.client.SetVFOFrequency(c.toTCITRX(vfo), client.VFOA, int(frequency))
	if err != nil {
		log.Printf("cannot set VFO frequency: %v", err)
	}
}

func (c *Client) SetBand(vfo model.VFOID, band model.Band) {
	bandplanBand := c.bandplan[toBandplanBandName(band)]
	frequency := findModePortionCenter(c.bandplan, int(bandplanBand.Center()), toBandplanMode(c.modeOf(vfo)))
	log.Printf("tci: switching TRX%d to %s at %d", c.toTCITRX(vfo), band, frequency)
	err := c.client.SetVFOFrequency(c.toTCITRX(vfo), client.VFOA, frequency)
	if err != nil {
		log.Printf("cannot switch to band %s: %v", band, err)
	}
}

func (c *Client) modeOf(vfo model.VFOID) model.Mode {
	if !c.singleVFO && vfo == model.VFO2 {
		return c.trx2.mode
	}
	return c.trx.mode
}

func (c *Client) SetMode(vfo model.VFOID, mode model.Mode) {
	err := c.client.SetMode(c.toTCITRX(vfo), toClientMode(mode))
	if err != nil {
		log.Printf("cannot set mode: %v", err)
	}
}

func (c *Client) SetIncrementalTuning(vfo model.VFOID, kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	setEnable := c.client.SetXITEnable
	setOffset := c.client.SetXITOffset
	if kind == model.RIT {
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

var spotColors = map[model.SpotType]client.ARGB{
	model.WorkedSpot:  client.NewARGB(255, 128, 128, 128),
	model.ManualSpot:  client.NewARGB(255, 255, 255, 255),
	model.SkimmerSpot: client.NewARGB(255, 255, 255, 255),
	model.RBNSpot:     client.NewARGB(255, 255, 255, 255),
	model.ClusterSpot: client.NewARGB(255, 255, 255, 255),
}

// a marker is no spot, therefore it has its own colors: the CQ frequency in dark red, the
// numbered and text markers in the light blue of the spot list
var markerColors = map[model.BandmapMarkerKind]client.ARGB{
	model.CQMarker:       client.NewARGB(255, 139, 0, 0),
	model.NumberedMarker: client.NewARGB(255, 173, 216, 230),
	model.TextMarker:     client.NewARGB(255, 173, 216, 230),
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

func (c *Client) activeBands() []model.Band {
	trxs := c.activeTRXs()
	result := make([]model.Band, 0, len(trxs))
	for _, trx := range trxs {
		result = append(result, trx.band)
	}
	return result
}

func (c *Client) trxOnBand(band model.Band) (*trxListener, bool) {
	for _, trx := range c.activeTRXs() {
		if trx.band == band {
			return trx, true
		}
	}
	return nil, false
}

func (c *Client) spotVisible(band model.Band, mode model.Mode) bool {
	for _, trx := range c.activeTRXs() {
		if trx.band == band && trx.mode == mode {
			return true
		}
	}
	return false
}

func (c *Client) resetSpots() {
	c.lastHeardSpots = make(map[string]time.Time)
	c.sentMarkers = make(map[string]model.BandmapMarker)
}

func (c *Client) EntryAdded(entry model.BandmapEntry) {
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

func (c *Client) MarkersChanged(markers []model.BandmapMarker) {
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

func (c *Client) addMarkers(markers []model.BandmapMarker) {
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
func diffMarkers(sentMarkers map[string]model.BandmapMarker, markers []model.BandmapMarker, bands []model.Band) (map[string]model.BandmapMarker, []model.BandmapMarker, []string) {
	currentMarkers := make(map[string]model.BandmapMarker, len(markers))
	addedMarkers := make([]model.BandmapMarker, 0, len(markers))
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

func (c *Client) EntryUpdated(entry model.BandmapEntry) {
	c.EntryAdded(entry)
}

func (c *Client) EntryRemoved(entry model.BandmapEntry) {
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
	vfo       model.VFOID
	frequency model.Frequency
	band      model.Band
	mode      model.Mode
	xitActive bool
	xitOffset model.Frequency
	ritActive bool
	ritOffset model.Frequency
	ptt       bool
}

func (l *trxListener) Refresh() {
	l.client.emitFrequencyChanged(l.vfo, l.frequency)
	l.client.emitBandChanged(l.vfo, l.band)
	l.client.emitModeChanged(l.vfo, l.mode)
	l.client.emitIncrementalTuningChanged(l.vfo, model.XIT, l.xitActive, l.xitOffset)
	l.client.emitIncrementalTuningChanged(l.vfo, model.RIT, l.ritActive, l.ritOffset)
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
	incomingFrequency := model.Frequency(frequency)
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
	l.client.emitIncrementalTuningChanged(l.vfo, model.XIT, l.xitActive, l.xitOffset)
}

func (l *trxListener) SetXITOffset(trx int, offset int) {
	if trx != l.trx {
		return
	}
	incomingOffset := model.Frequency(offset)
	if incomingOffset == l.xitOffset {
		return
	}
	l.xitOffset = incomingOffset
	l.client.emitIncrementalTuningChanged(l.vfo, model.XIT, l.xitActive, l.xitOffset)
}

func (l *trxListener) SetRITEnable(trx int, active bool) {
	if trx != l.trx {
		return
	}
	if active == l.ritActive {
		return
	}
	l.ritActive = active
	l.client.emitIncrementalTuningChanged(l.vfo, model.RIT, l.ritActive, l.ritOffset)
}

func (l *trxListener) SetRITOffset(trx int, offset int) {
	if trx != l.trx {
		return
	}
	incomingOffset := model.Frequency(offset)
	if incomingOffset == l.ritOffset {
		return
	}
	l.ritOffset = incomingOffset
	l.client.emitIncrementalTuningChanged(l.vfo, model.RIT, l.ritActive, l.ritOffset)
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

func toCoreBand(bandName bandplan.BandName) model.Band {
	if bandName == bandplan.BandUnknown {
		return model.NoBand
	}
	return model.Band(bandName)
}

func toBandplanBandName(band model.Band) bandplan.BandName {
	if band == model.NoBand {
		return bandplan.BandUnknown
	}
	return bandplan.BandName(band)
}

func toCoreMode(mode client.Mode) model.Mode {
	switch mode {
	case client.ModeUSB, client.ModeLSB:
		return model.ModeSSB
	case client.ModeCW:
		return model.ModeCW
	case client.ModeNFM, client.ModeWFM:
		return model.ModeFM
	case client.ModeDIGL, client.ModeDIGU, client.ModeSPEC:
		return model.ModeDigital
	default:
		return model.NoMode
	}
}

func toClientMode(mode model.Mode) client.Mode {
	switch mode {
	case model.ModeCW:
		return client.ModeCW
	case model.ModeSSB:
		return client.ModeUSB // TODO make this dependent of the current frequency either LSB or USB
	case model.ModeFM:
		return client.ModeNFM
	case model.ModeRTTY:
		return client.ModeDIGU
	case model.ModeDigital:
		return client.ModeDIGU
	default:
		return client.ModeSPEC
	}
}

func toBandplanMode(mode model.Mode) bandplan.Mode {
	switch mode {
	case model.ModeCW:
		return bandplan.ModeCW
	case model.ModeSSB, model.ModeFM:
		return bandplan.ModePhone
	case model.ModeDigital, model.ModeRTTY:
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
