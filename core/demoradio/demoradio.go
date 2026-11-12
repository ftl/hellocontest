// Package demoradio provides a radio that exists only in the application. It answers every
// command immediately and keeps the state that it was given, so that a script can put the
// user interface into a defined state without a real transceiver.
package demoradio

import (
	"log"
	"sync"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"

	"github.com/ftl/hellocontest/core"
)

const (
	DefaultFrequency = core.Frequency(7028000)
	DefaultMode      = core.ModeCW
)

type Client struct {
	bandplan    bandplan.Bandplan
	asyncRunner core.AsyncRunner

	mutex      sync.RWMutex
	listeners  []any
	vfos       [core.VFOCount]vfoState
	currentVFO core.VFOID
	txVFO      core.VFOID
	singleVFO  bool
	connected  bool
}

type vfoState struct {
	frequency core.Frequency
	band      core.Band
	mode      core.Mode
}

func New(bandplan bandplan.Bandplan, singleVFO bool, asyncRunner core.AsyncRunner) *Client {
	result := &Client{
		bandplan:    bandplan,
		asyncRunner: asyncRunner,
		singleVFO:   singleVFO,
		connected:   true,
	}
	for i := range result.vfos {
		result.vfos[i] = vfoState{
			frequency: DefaultFrequency,
			band:      result.bandOf(DefaultFrequency),
			mode:      DefaultMode,
		}
	}
	return result
}

func (c *Client) SetBandplan(bandplan bandplan.Bandplan) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.bandplan = bandplan
}

func (c *Client) bandOf(frequency core.Frequency) core.Band {
	band := c.bandplan.ByFrequency(hamradio.Frequency(frequency))
	if band.Name == bandplan.BandUnknown {
		return core.NoBand
	}
	return core.Band(band.Name)
}

func (c *Client) Notify(listener any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.listeners = append(c.listeners, listener)
}

func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}

func (c *Client) Active() bool {
	return c.IsConnected()
}

func (c *Client) SingleVFO() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.singleVFO
}

func (c *Client) Disconnect() {
	c.mutex.Lock()
	c.connected = false
	c.mutex.Unlock()
	c.emitConnectionChanged(false)
}

func (c *Client) SetCurrentVFO(vfo core.VFOID) {
	if !c.knownVFO(vfo) {
		return
	}
	c.mutex.Lock()
	changed := c.currentVFO != vfo
	c.currentVFO = vfo
	c.mutex.Unlock()
	if !changed {
		return
	}
	c.asyncRunner(func() {
		core.Emit(c.listeners, func(listener core.CurrentVFOListener) {
			listener.CurrentVFOChanged(vfo)
		})
	})
}

func (c *Client) SetTXVFO(vfo core.VFOID) {
	if !c.knownVFO(vfo) {
		return
	}
	c.mutex.Lock()
	changed := c.txVFO != vfo
	c.txVFO = vfo
	c.mutex.Unlock()
	if !changed {
		return
	}
	c.asyncRunner(func() {
		core.Emit(c.listeners, func(listener core.TXVFOListener) {
			listener.TXVFOChanged(vfo)
		})
	})
}

func (c *Client) SetFrequency(vfo core.VFOID, frequency core.Frequency) {
	if !c.knownVFO(vfo) {
		return
	}
	c.mutex.Lock()
	c.vfos[vfo].frequency = frequency
	c.vfos[vfo].band = c.bandOf(frequency)
	current := c.vfos[vfo]
	c.mutex.Unlock()
	c.echo(vfo, current)
}

func (c *Client) SetBand(vfo core.VFOID, band core.Band) {
	if !c.knownVFO(vfo) {
		return
	}
	frequency, ok := c.frequencyInBand(band)
	if !ok {
		log.Printf("demo radio: unknown band %s", band)
		return
	}
	c.SetFrequency(vfo, frequency)
}

// frequencyInBand puts the VFO into the lower portion of the band, where a CW contest
// takes place.
func (c *Client) frequencyInBand(band core.Band) (core.Frequency, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	b, ok := c.bandplan[bandplan.BandName(band)]
	if !ok {
		return 0, false
	}
	return core.Frequency(b.From) + 20000, true
}

func (c *Client) SetMode(vfo core.VFOID, mode core.Mode) {
	if !c.knownVFO(vfo) {
		return
	}
	c.mutex.Lock()
	c.vfos[vfo].mode = mode
	current := c.vfos[vfo]
	c.mutex.Unlock()
	c.echo(vfo, current)
}

func (c *Client) SetIncrementalTuning(vfo core.VFOID, kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	if !c.knownVFO(vfo) {
		return
	}
	c.asyncRunner(func() {
		core.Emit(c.listeners, func(listener core.VFOIncrementalTuningListener) {
			listener.VFOIncrementalTuningChanged(vfo, kind, active, offset)
		})
	})
}

func (c *Client) Refresh() {
	c.mutex.RLock()
	vfos := c.vfos
	currentVFO := c.currentVFO
	txVFO := c.txVFO
	c.mutex.RUnlock()

	c.emitConnectionChanged(true)
	c.asyncRunner(func() {
		core.Emit(c.listeners, func(listener core.CurrentVFOListener) {
			listener.CurrentVFOChanged(currentVFO)
		})
		core.Emit(c.listeners, func(listener core.TXVFOListener) {
			listener.TXVFOChanged(txVFO)
		})
		for vfo, state := range vfos {
			c.emitState(core.VFOID(vfo), state)
		}
	})
}

func (c *Client) Speed(int) {}

func (c *Client) Send(string) {}

func (c *Client) Abort() {}

func (c *Client) knownVFO(vfo core.VFOID) bool {
	if vfo < 0 || int(vfo) >= len(c.vfos) {
		return false
	}
	return !c.singleVFO || vfo == core.VFO1
}

// echo answers every command, also when it does not change anything: the VFO waits for the
// confirmation of the radio before it shows a commanded value
func (c *Client) echo(vfo core.VFOID, state vfoState) {
	c.asyncRunner(func() {
		c.emitState(vfo, state)
	})
}

func (c *Client) emitState(vfo core.VFOID, state vfoState) {
	core.Emit(c.listeners, func(listener core.VFOFrequencyListener) {
		listener.VFOFrequencyChanged(vfo, state.frequency)
	})
	core.Emit(c.listeners, func(listener core.VFOBandListener) {
		listener.VFOBandChanged(vfo, state.band)
	})
	core.Emit(c.listeners, func(listener core.VFOModeListener) {
		listener.VFOModeChanged(vfo, state.mode)
	})
}

func (c *Client) emitConnectionChanged(connected bool) {
	c.asyncRunner(func() {
		core.Emit(c.listeners, func(listener core.ConnectionChangedListener) {
			listener.ConnectionChanged(connected)
		})
	})
}
