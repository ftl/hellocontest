package hamlib

import (
	"log"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/hl-go"

	"github.com/ftl/hellocontest/core"
)

type Client struct {
	client *hl.RigClient

	bandplan bandplan.Bandplan
	vfo1     hl.VFO
	vfo2     hl.VFO

	listeners []any

	pollingInterval time.Duration
	requestTimeout  time.Duration
	do              chan func()
	done            chan struct{}
	loopStopped     chan struct{}

	lastState vfoState
}

type vfoState struct {
	frequency core.Frequency
	band      core.Band
	mode      core.Mode
	xitActive bool
	xitOffset core.Frequency
	ptt       bool
}

func New(address string, bandplan bandplan.Bandplan, vfo1, vfo2 string) *Client {
	result := &Client{
		client:          hl.NewRigClient(address),
		bandplan:        bandplan,
		vfo1:            sanitizeHamlibVFO(core.VFO1, vfo1),
		vfo2:            sanitizeHamlibVFO(core.VFO2, vfo2),
		pollingInterval: 500 * time.Millisecond,
		requestTimeout:  500 * time.Millisecond,
		do:              make(chan func()),
		done:            make(chan struct{}),
		loopStopped:     make(chan struct{}),
	}
	result.client.Notify(result)
	return result
}

func sanitizeHamlibVFO(id core.VFOID, s string) hl.VFO {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" { // shortcut
		if id == core.VFO1 {
			return hl.CurrVFO
		}
		return ""
	}

	validVFOs := []hl.VFO{hl.MainVFO, hl.SubVFO, hl.VFOA, hl.VFOB, hl.VFOC, hl.MainAVFO, hl.SubAVFO, hl.MainBVFO, hl.SubBVFO}
	for _, vfo := range validVFOs {
		if strings.ToLower(string(vfo)) == s {
			return vfo
		}
	}
	log.Printf("hamlib: invalid VFO: %s", s)
	if id == core.VFO1 {
		return hl.CurrVFO
	}
	return ""
}

func (c *Client) run() {
	go func() {
		defer close(c.loopStopped)
		for {
			currentState, err := c.poll()
			if err != nil {
				continue
			}
			c.emitChangeNotifications(c.lastState, currentState)
			c.lastState = currentState

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
func (c *Client) poll() (vfoState, error) {
	frequency, err := c.client.GetFrequency(hl.CurrVFO)
	if err != nil {
		return vfoState{}, err
	}
	mode, _, err := c.client.GetMode(hl.CurrVFO)
	if err != nil {
		return vfoState{}, err
	}
	xitActive, err := c.client.GetFunc(hl.CurrVFO, hl.XITFunction)
	if err != nil {
		return vfoState{}, err
	}
	var xitOffset hl.Frequency
	if xitActive {
		xitOffset, err = c.client.GetXIT(hl.CurrVFO)
		if err != nil {
			return vfoState{}, err
		}
	} else {
		xitOffset = 0
	}
	pttStatus, err := c.client.GetPTT(hl.CurrVFO)
	if err != nil {
		return vfoState{}, err
	}

	return vfoState{
		frequency: core.Frequency(frequency),
		band:      toCoreBand(c.bandplan.ByFrequency(hamradio.Frequency(frequency)).Name),
		mode:      toCoreMode(mode),
		xitActive: xitActive,
		xitOffset: core.Frequency(xitOffset),
		ptt:       pttStatus != hl.PTTOff,
	}, nil
}

func (c *Client) doInLoop(f func()) {
	c.do <- f
}

func (c *Client) KeepOpen() {
	err := c.client.Open(true)
	if err != nil {
		log.Printf("hamlib: connection error: %v", err)
	}
	c.run()
}

func (c *Client) Connect() error {
	err := c.client.Open(false)
	if err != nil {
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
	c.doInLoop(func() {
		err := c.client.SetFrequency(hl.CurrVFO, hl.Frequency(frequency))
		if err != nil {
			log.Printf("hamlib: cannot set frequency: %v", err)
		}
		c.lastState.frequency = frequency
		c.lastState.band = toCoreBand(c.bandplan.ByFrequency(hamradio.Frequency(frequency)).Name)
	})
}

func (c *Client) SetBand(band core.Band) {
	outgoingBand, ok := c.bandplan[toBandplanBandName(band)]
	if !ok {
		log.Printf("hamlib: unknown band %v", band)
		return
	}

	c.doInLoop(func() {
		frequency := findModePortionCenter(c.bandplan, int(outgoingBand.Center()), toBandplanMode(c.lastState.mode))
		err := c.client.SetFrequency(hl.CurrVFO, hl.Frequency(frequency))
		if err != nil {
			log.Printf("hamlib: cannot switch to band: %v", err)
			return
		}
		c.lastState.frequency = core.Frequency(frequency)
		c.lastState.band = band
	})
}

func (c *Client) SetMode(mode core.Mode) {
	c.doInLoop(func() {
		err := c.client.SetMode(hl.CurrVFO, toClientMode(mode), 0)
		if err != nil {
			log.Printf("hamlib: cannot switch to mode: %v", err)
			return
		}
		c.lastState.mode = mode
	})
}

func (c *Client) SetXIT(active bool, offset core.Frequency) {
	c.doInLoop(func() {
		if active == c.lastState.xitActive && offset == c.lastState.xitOffset {
			return
		}

		if active != c.lastState.xitActive {
			err := c.client.SetFunc(hl.CurrVFO, hl.XITFunction, active)
			if err != nil {
				log.Printf("hamlib: cannot set XIT function: %v", err)
				return
			}
		}

		if active && (offset != c.lastState.xitOffset) {
			err := c.client.SetXIT(hl.CurrVFO, hl.Frequency(offset))
			if err != nil {
				log.Printf("hamlib: cannot set XIT offset: %v", err)
				return
			}
		}

		c.lastState.xitActive = active
		c.lastState.xitOffset = offset
	})
	// TODO: enable the XIT and set its offset
}

func (c *Client) Refresh() {
	c.doInLoop(func() {
		c.emitChangeNotifications(vfoState{}, c.lastState)
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

func (c *Client) emitChangeNotifications(last, current vfoState) {
	go func() {
		if last.frequency != current.frequency {
			c.emitFrequencyChanged(current.frequency)
		}
		if last.band != current.band {
			c.emitBandChanged(current.band)
		}
		if last.mode != current.mode {
			c.emitModeChanged(current.mode)
		}
		if (last.xitActive != current.xitActive) || (current.xitActive && (last.xitOffset != current.xitOffset)) {
			c.emitXITChanged(current.xitActive, current.xitOffset)
		}
		if last.ptt != current.ptt {
			c.emitPTTChanged(current.ptt)
		}
	}()
}

func (c *Client) emitFrequencyChanged(frequency core.Frequency) {
	core.Emit(c.listeners, func(listener core.VFOFrequencyListener) {
		listener.VFOFrequencyChanged(frequency)
	})
}

func (c *Client) emitBandChanged(band core.Band) {
	core.Emit(c.listeners, func(listener core.VFOBandListener) {
		listener.VFOBandChanged(band)
	})
}

func (c *Client) emitModeChanged(mode core.Mode) {
	core.Emit(c.listeners, func(listener core.VFOModeListener) {
		listener.VFOModeChanged(mode)
	})
}

func (c *Client) emitXITChanged(active bool, offset core.Frequency) {
	core.Emit(c.listeners, func(listener core.VFOXITListener) {
		listener.VFOXITChanged(active, offset)
	})
}

func (c *Client) emitPTTChanged(active bool) {
	core.Emit(c.listeners, func(listener core.VFOPTTListener) {
		listener.VFOPTTChanged(active)
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
