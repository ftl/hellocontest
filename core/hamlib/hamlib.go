package hamlib

import (
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/hl-go"

	"github.com/ftl/hellocontest/core"
)

const (
	doBufferSize = 10
)

type Client struct {
	client *hl.RigClient

	bandplan bandplan.Bandplan
	vfos     []hl.VFO

	listeners []any

	pollingInterval time.Duration
	requestTimeout  time.Duration
	do              chan func()
	done            chan struct{}
	loopRunning     *atomic.Bool
	loopStopped     chan struct{}

	currentVFO hl.VFO
	singleVFO  bool
	lastState  []vfoState
}

type vfoState struct {
	vfo          hl.VFO
	frequency    core.Frequency
	band         core.Band
	mode         core.Mode
	txVFO        bool
	xitActive    bool
	xitOffset    core.Frequency
	xitAvailable bool
	ptt          bool
	pttAvailable bool
}

func New(address string, bandplan bandplan.Bandplan, vfo1, vfo2 string) *Client {
	vfos, singleVFO := sanitizeVFOs(vfo1, vfo2)
	result := &Client{
		client:          hl.NewRigClient(address),
		bandplan:        bandplan,
		vfos:            vfos,
		pollingInterval: 500 * time.Millisecond,
		requestTimeout:  500 * time.Millisecond,
		do:              make(chan func(), doBufferSize),
		done:            make(chan struct{}),
		loopRunning:     new(atomic.Bool),
		loopStopped:     make(chan struct{}),
		currentVFO:      hl.CurrVFO,
		singleVFO:       singleVFO,
		lastState:       make([]vfoState, int(core.VFOCount)),
	}
	result.client.Notify(result)
	if singleVFO {
		log.Printf("hamlib: using SINGLE VFO: %s", result.vfos[core.VFO1])
	} else {
		log.Printf("hamlib: using VFOs: %v", result.vfos)
	}
	return result
}

func sanitizeVFOs(vfo1, vfo2 string) ([]hl.VFO, bool) {
	result := []hl.VFO{sanitizeHamlibVFO(vfo1), sanitizeHamlibVFO(vfo2)}
	singleVFO := false
	if result[core.VFO1] == "" {
		result[core.VFO1] = hl.CurrVFO
		singleVFO = true
	}
	return result, singleVFO
}

func sanitizeHamlibVFO(s string) hl.VFO {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" { // shortcut
		return ""
	}

	validVFOs := []hl.VFO{hl.MainVFO, hl.SubVFO, hl.VFOA, hl.VFOB, hl.VFOC, hl.MainAVFO, hl.SubAVFO, hl.MainBVFO, hl.SubBVFO}
	for _, vfo := range validVFOs {
		if strings.ToLower(string(vfo)) == s {
			return vfo
		}
	}

	log.Printf("hamlib: invalid VFO: %s", s)
	return ""
}

func (c *Client) toVFOID(vfo hl.VFO) (core.VFOID, bool) {
	normalizedVFO := strings.ToLower(string(vfo))
	for id := range c.vfos {
		if strings.ToLower(string(c.vfos[id])) == normalizedVFO {
			return core.VFOID(id), true
		}
	}
	return -1, false
}

func (c *Client) run() {
	go func() {
		defer close(c.loopStopped)
		defer c.loopRunning.Store(false)
		c.loopRunning.Store(true)
		currentState := make([]vfoState, core.VFOCount)
		for {
			currentVFO, currentState, err := c.poll(currentState)
			if err != nil {
				log.Printf("hamlib: cannot poll: %v", err)
			} else {
				lastVFO := c.currentVFO
				c.currentVFO = currentVFO
				c.emitCurrentVFOChanged(lastVFO, currentVFO)

				for vfo := range core.VFOCount {
					c.emitChangeNotifications(vfo, c.lastState[vfo], currentState[vfo])
					c.lastState[vfo] = currentState[vfo]
				}
			}

			select {
			case f := <-c.do:
				f()
			case <-time.After(c.pollingInterval):
			case <-c.done:
				return
			}
		}
	}()
}

// poll must only be called from the run goroutine!
func (c *Client) poll(state []vfoState) (hl.VFO, []vfoState, error) {
	if c.singleVFO {
		vfoState, err := c.pollSingleVFO(hl.CurrVFO)
		if err != nil {
			return hl.CurrVFO, state, err
		}
		state[core.VFO1] = vfoState
		return hl.CurrVFO, state, nil
	}

	return c.pollDualVFO(state)
}

// pollSingleVFO must only be called from the run goroutine!
func (c *Client) pollSingleVFO(vfo hl.VFO) (vfoState, error) {
	if vfo == "" {
		return vfoState{}, nil
	}

	frequency, mode, _, _, _, err := c.client.GetVFOInfo(vfo)
	if err != nil {
		return vfoState{}, err
	}

	result := vfoState{
		vfo:       vfo,
		frequency: core.Frequency(frequency),
		band:      toCoreBand(c.bandplan.ByFrequency(hamradio.Frequency(frequency)).Name),
		mode:      toCoreMode(mode),
		txVFO:     true,
	}
	result = c.pollVFOAdditional(vfo, result)

	return result, nil
}

// pollDualVFO must only be called from the run goroutine!
func (c *Client) pollDualVFO(state []vfoState) (hl.VFO, []vfoState, error) {
	rigInfo, err := c.client.GetRigInfo()
	if err != nil {
		// TODO: be a bit smarter than just giving up
		log.Printf("hamlib: cannot retrieve RigInfo: %v", err)
		return c.currentVFO, state, err
	}
	// log.Printf("hamlib: RigInfo: %v", rigInfo)

	for _, vfoInfo := range rigInfo.VFOs {
		vfoID, ok := c.toVFOID(vfoInfo.VFO)
		if !ok {
			continue
		}
		vfoState := state[vfoID]
		vfoState.vfo = vfoInfo.VFO
		vfoState.frequency = core.Frequency(vfoInfo.Frequency)
		vfoState.band = toCoreBand(c.bandplan.ByFrequency(hamradio.Frequency(vfoInfo.Frequency)).Name)
		vfoState.mode = toCoreMode(vfoInfo.Mode)
		vfoState.txVFO = vfoInfo.TXActive
		state[vfoID] = vfoState
	}

	var currentVFOID core.VFOID
	var currentVFOOK bool
	currentVFO, err := c.client.GetVFO()
	if err != nil {
		// GetVFO is not supported on every radio, fallback to VFO1
		// log.Printf("hamlib: get_vfo not supported, using VFO %s", c.currentVFO)
		currentVFO = c.currentVFO
		currentVFOID = core.VFO1
		currentVFOOK = true
	} else {
		currentVFOID, currentVFOOK = c.toVFOID(currentVFO)
	}
	if currentVFOOK {
		vfoState := state[currentVFOID]
		vfoState = c.pollVFOAdditional(vfoState.vfo, vfoState)
		state[currentVFOID] = vfoState
	}

	return currentVFO, state, nil
}

// pollVFOAdditional must only be called from the run goroutine!
func (c *Client) pollVFOAdditional(vfo hl.VFO, state vfoState) vfoState {
	result := state
	var err error

	result.xitAvailable = true
	result.xitActive, err = c.client.GetFunc(vfo, hl.XITFunction)
	if err != nil {
		result.xitActive = false
		result.xitAvailable = false
	}
	var xitOffset hl.Frequency
	if result.xitActive {
		xitOffset, err = c.client.GetXIT(vfo)
		if err != nil {
			result.xitActive = false
			result.xitAvailable = false
			xitOffset = 0
		}
	} else {
		xitOffset = 0
	}
	result.xitOffset = core.Frequency(xitOffset)

	pttStatus, err := c.client.GetPTT(vfo)
	result.pttAvailable = (err == nil)
	if err != nil {
		pttStatus = hl.PTTOff
	}
	result.ptt = (pttStatus != hl.PTTOff)

	return result
}

func (c *Client) doInLoop(f func()) {
	loopRunning := c.loopRunning.Load()
	if !loopRunning {
		return
	}
	c.do <- f
}

func (c *Client) KeepOpen() {
	err := c.connectAndRun(true)
	if err != nil {
		log.Printf("hamlib: connection error: %v", err)
	}
}

func (c *Client) Connect() error {
	return c.connectAndRun(false)
}

func (c *Client) connectAndRun(automaticReconnect bool) error {
	err := c.client.Open(automaticReconnect)
	if err != nil && !automaticReconnect {
		return err
	}
	c.run()
	return nil
}

func (c *Client) Disconnect() {
	select {
	case <-c.done:
		return
	default:
	}
	close(c.done)
	<-c.loopStopped
	c.client.Close()
}

func (c *Client) RigConnected(connected bool) {
	c.emitConnectionChanged(connected)
}

func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

func (c *Client) Active() bool {
	return c.IsConnected()
}

func (c *Client) SetFrequency(frequency core.Frequency) {
	// TODO: add the VFOID to all VFO-related Setters
	vfo := core.VFO1
	c.doInLoop(func() {
		err := c.client.SetFrequency(c.vfos[vfo], hl.Frequency(frequency))
		if err != nil {
			log.Printf("hamlib: cannot set frequency: %v", err)
		}
	})
}

func (c *Client) SetBand(band core.Band) {
	// TODO: add the VFOID to all VFO-related Setters
	vfo := core.VFO1

	outgoingBand, ok := c.bandplan[toBandplanBandName(band)]
	if !ok {
		log.Printf("hamlib: unknown band %v", band)
		return
	}

	c.doInLoop(func() {
		frequency := findModePortionCenter(c.bandplan, int(outgoingBand.Center()), toBandplanMode(c.lastState[vfo].mode))
		err := c.client.SetFrequency(c.vfos[vfo], hl.Frequency(frequency))
		if err != nil {
			log.Printf("hamlib: cannot switch to band: %v", err)
			return
		}
	})
}

func (c *Client) SetMode(mode core.Mode) {
	// TODO: add the VFOID to all VFO-related Setters
	vfo := core.VFO1
	c.doInLoop(func() {
		err := c.client.SetMode(c.vfos[vfo], toClientMode(mode), 0)
		if err != nil {
			log.Printf("hamlib: cannot switch to mode: %v", err)
			return
		}
	})
}

func (c *Client) SetXIT(active bool, offset core.Frequency) {
	// TODO: add the VFOID to all VFO-related Setters
	vfo := core.VFO1
	c.doInLoop(func() {
		if active == c.lastState[vfo].xitActive && offset == c.lastState[vfo].xitOffset {
			return
		}

		if active != c.lastState[vfo].xitActive {
			err := c.client.SetFunc(c.vfos[vfo], hl.XITFunction, active)
			if err != nil {
				log.Printf("hamlib: cannot set XIT function: %v", err)
				return
			}
		}

		if active && (offset != c.lastState[vfo].xitOffset) {
			err := c.client.SetXIT(c.vfos[vfo], hl.Frequency(offset))
			if err != nil {
				log.Printf("hamlib: cannot set XIT offset: %v", err)
				return
			}
		}
	})
}

func (c *Client) Refresh() {
	c.doInLoop(func() {
		for vfo := range core.VFOCount {
			c.emitChangeNotifications(vfo, vfoState{}, c.lastState[vfo])
		}
	})
}

func (c *Client) Speed(speed int) {
	c.doInLoop(func() {
		err := c.client.SetLevel(hl.CurrVFO, hl.KeyerSpeedLevel, float64(speed))
		if err != nil {
			log.Printf("hamlib: cannot set morse speed: %v", err)
		}
	})
}

func (c *Client) Send(text string) {
	c.doInLoop(func() {
		err := c.client.SendMorse(text)
		if err != nil {
			log.Printf("hamlib: cannot send morse text: %v", err)
		}
	})
}

func (c *Client) Abort() {
	c.doInLoop(func() {
		err := c.client.StopMorse()
		if err != nil {
			log.Printf("hamlib: cannot stop morse transmission: %v", err)
		}
	})
}

func (c *Client) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) emitConnectionChanged(connected bool) {
	type listenerType interface {
		ConnectionChanged(bool)
	}
	core.Emit(c.listeners, func(listener listenerType) {
		listener.ConnectionChanged(connected)
	})
}

func (c *Client) emitCurrentVFOChanged(last, current hl.VFO) {
	if last == current {
		return
	}

	vfoID, ok := c.toVFOID(current)
	if !ok {
		log.Printf("VFO change ignored: %s -> %s", last, current)
		return
	}
	log.Printf("current vfo changed: %d = %s", vfoID, c.vfos[vfoID])

	go func() {
		core.Emit(c.listeners, func(listener core.CurrentVFOListener) {
			listener.CurrentVFOChanged(vfoID)
		})
	}()
}

func (c *Client) emitChangeNotifications(vfo core.VFOID, last, current vfoState) {
	go func() {
		if last.frequency != current.frequency {
			c.emitFrequencyChanged(vfo, current.frequency)
		}
		if last.band != current.band {
			c.emitBandChanged(vfo, current.band)
		}
		if last.mode != current.mode {
			c.emitModeChanged(vfo, current.mode)
		}
		if (last.xitActive != current.xitActive) || (current.xitActive && (last.xitOffset != current.xitOffset)) {
			c.emitXITChanged(vfo, current.xitActive, current.xitOffset)
		}
		if last.ptt != current.ptt {
			c.emitPTTChanged(vfo, current.ptt)
		}
	}()
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

func (c *Client) emitXITChanged(vfo core.VFOID, active bool, offset core.Frequency) {
	core.Emit(c.listeners, func(listener core.VFOXITListener) {
		listener.VFOXITChanged(vfo, active, offset)
	})
}

func (c *Client) emitPTTChanged(vfo core.VFOID, active bool) {
	core.Emit(c.listeners, func(listener core.VFOPTTListener) {
		listener.VFOPTTChanged(vfo, active)
	})
}

func toCoreBand(bandName bandplan.BandName) core.Band {
	if bandName == bandplan.BandUnknown {
		return core.NoBand
	}
	return core.Band(bandName)
}

func toCoreMode(mode hl.Mode) core.Mode {
	switch mode {
	case hl.ModeUSB, hl.ModeLSB:
		return core.ModeSSB
	case hl.ModeCW, hl.ModeCWR:
		return core.ModeCW
	case hl.ModeRTTY, hl.ModeRTTYR:
		return core.ModeRTTY
	case hl.ModeFM, hl.ModeWFM:
		return core.ModeFM
	case hl.ModePKTLSB, hl.ModePKTUSB, hl.ModePKTFM, hl.ModeECSSLSB, hl.ModeECSSUSB, hl.ModeSAM, hl.ModeSAL, hl.ModeSAH:
		return core.ModeDigital
	default:
		return core.NoMode
	}
}

func toClientMode(mode core.Mode) hl.Mode {
	switch mode {
	case core.ModeCW:
		return hl.ModeCW
	case core.ModeSSB:
		return hl.ModeUSB // TODO make this dependent of the current frequency either LSB or USB
	case core.ModeFM:
		return hl.ModeFM
	case core.ModeRTTY:
		return hl.ModeRTTY
	case core.ModeDigital:
		return hl.ModePKTUSB
	default:
		return ""
	}
}

func findModePortionCenter(bp bandplan.Bandplan, f int, mode bandplan.Mode) core.Frequency {
	log.Printf("find mode portion center: %d %s", f, mode)
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
		return core.Frequency(currentPortion.Center())
	}
	if modePortion.Mode == mode {
		return core.Frequency(modePortion.Center())
	}
	return core.Frequency(band.Center())
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

func toBandplanBandName(band core.Band) bandplan.BandName {
	if band == core.NoBand {
		return bandplan.BandUnknown
	}
	return bandplan.BandName(band)
}
