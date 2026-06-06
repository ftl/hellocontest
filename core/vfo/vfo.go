package vfo

import (
	"log"
	"sync"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/hellocontest/core"
)

type Client interface {
	Notify(any)
	Active() bool
	Refresh()
	SetTXVFO(core.VFOID)
	SetFrequency(core.VFOID, core.Frequency)
	SetBand(core.VFOID, core.Band)
	SetMode(core.VFOID, core.Mode)
	SetXIT(bool, core.Frequency)
	MuteAudio(core.VFOID)
	UnmuteAudio(core.VFOID)
	ToggleAudio(core.VFOID)
}

type Logbook interface {
	LastBand() core.Band
	LastMode() core.Mode
}

type VFO struct {
	XITControl

	id   core.VFOID
	name string

	bandplan      bandplan.Bandplan
	logbook       Logbook
	client        Client
	offlineClient *offlineClient
	refreshing    bool
	asyncRunner   core.AsyncRunner

	listeners []any
}

func NewVFO(id core.VFOID, name string, bandplan bandplan.Bandplan, logbook Logbook, asyncRunner core.AsyncRunner) *VFO {
	result := &VFO{
		id:          id,
		name:        name,
		bandplan:    bandplan,
		logbook:     logbook,
		asyncRunner: asyncRunner,
	}
	result.XITControl = XITControl{
		vfo: result,
	}
	result.offlineClient = newOfflineClient(result)
	result.SetClient(nil)

	return result
}

func (v *VFO) SetClient(client Client) {
	v.client = client
	if client != nil {
		v.client.Notify(v)
	}
}

func (v *VFO) Name() string {
	return v.name
}

func (v *VFO) Notify(listener any) {
	v.listeners = append(v.listeners, listener)
}

func (v *VFO) online() bool {
	return v.client != nil && v.client.Active()
}

func (v *VFO) Refresh() {
	v.refreshing = true
	if !v.online() {
		v.offlineClient.Refresh()
		return
	}
	v.client.Refresh()
	v.refreshing = false
}

func (v *VFO) SetFrequency(frequency core.Frequency) {
	if v.online() {
		v.client.SetFrequency(v.id, frequency)
	} else {
		v.offlineClient.SetFrequency(frequency)
	}
}

func (v *VFO) SetBand(band core.Band) {
	if v.online() {
		v.client.SetBand(v.id, band)
	} else {
		v.offlineClient.SetBand(band)
	}
}

func (v *VFO) SetMode(mode core.Mode) {
	if v.online() {
		v.client.SetMode(v.id, mode)
	} else {
		v.offlineClient.SetMode(mode)
	}
}

func (v *VFO) SetXIT(active bool, offset core.Frequency) {
	if v.online() {
		v.client.SetXIT(active, offset)
	} else {
		v.offlineClient.SetXIT(active, offset)
	}
}

func (v *VFO) MuteAudio() {
	if v.online() {
		v.client.MuteAudio(v.id)
	} else {
		v.offlineClient.MuteAudio()
	}
}

func (v *VFO) UnmuteAudio() {
	if v.online() {
		v.client.UnmuteAudio(v.id)
	} else {
		v.offlineClient.UnmuteAudio()
	}
}

func (v *VFO) ToggleAudio() {
	if v.online() {
		v.client.ToggleAudio(v.id)
	} else {
		v.offlineClient.ToggleAudio()
	}
}

func (v *VFO) LogbookLoaded() {
	log.Printf("VFO logbook changed")

	if v.online() {
		return
	}

	lastBand := v.logbook.LastBand()
	if lastBand != core.NoBand {
		v.offlineClient.SetBand(lastBand)
	}

	lastMode := v.logbook.LastMode()
	if lastMode != core.NoMode {
		v.offlineClient.SetMode(lastMode)
	}

	v.Refresh()
}

func (v *VFO) VFOFrequencyChanged(vfo core.VFOID, frequency core.Frequency) {
	if vfo != v.id {
		return
	}
	v.offlineClient.SetFrequency(frequency)
}

func (v *VFO) VFOBandChanged(vfo core.VFOID, band core.Band) {
	if vfo != v.id {
		return
	}
	v.offlineClient.SetBand(band)
}

func (v *VFO) VFOModeChanged(vfo core.VFOID, mode core.Mode) {
	if vfo != v.id {
		return
	}
	v.offlineClient.SetMode(mode)
}

func (v *VFO) VFOXITChanged(vfo core.VFOID, active bool, offset core.Frequency) {
	if vfo != v.id {
		return
	}
	v.XITControl.VFOXITChanged(vfo, active, offset)
	v.offlineClient.SetXIT(active, offset)
}

func (v *VFO) VFOPTTChanged(vfo core.VFOID, active bool) {
	if vfo != v.id {
		return
	}
	v.offlineClient.SetPTT(active)
}

func (v *VFO) emitFrequencyChanged(frequency core.Frequency) {
	core.Emit(v.listeners, func(listener core.VFOFrequencyListener) {
		v.asyncRunner(func() {
			listener.VFOFrequencyChanged(v.id, frequency)
		})
	})
}

func (v *VFO) emitBandChanged(band core.Band) {
	core.Emit(v.listeners, func(listener core.VFOBandListener) {
		v.asyncRunner(func() {
			listener.VFOBandChanged(v.id, band)
		})
	})
}

func (v *VFO) emitModeChanged(mode core.Mode) {
	core.Emit(v.listeners, func(listener core.VFOModeListener) {
		v.asyncRunner(func() {
			listener.VFOModeChanged(v.id, mode)
		})
	})
}

func (v *VFO) emitXITChanged(active bool, offset core.Frequency) {
	core.Emit(v.listeners, func(listener core.VFOXITListener) {
		v.asyncRunner(func() {
			listener.VFOXITChanged(v.id, active, offset)
		})
	})
}

func (v *VFO) emitPTTChanged(active bool) {
	core.Emit(v.listeners, func(listener core.VFOPTTListener) {
		v.asyncRunner(func() {
			listener.VFOPTTChanged(v.id, active)
		})
	})
}

type bandState struct {
	frequency core.Frequency
	mode      core.Mode
	xitActive bool
	xitOffset core.Frequency
}

type offlineClient struct {
	vfo         *VFO
	currentBand core.Band
	stateLock   *sync.RWMutex
	lastStates  map[core.Band]bandState
}

func newOfflineClient(vfo *VFO) *offlineClient {
	result := &offlineClient{
		vfo:         vfo,
		currentBand: core.Band160m,
		stateLock:   &sync.RWMutex{},
		lastStates:  make(map[core.Band]bandState),
	}
	_ = result.lastState(result.currentBand)
	return result
}

func (c *offlineClient) lastState(band core.Band) bandState {
	result, ok := c.lastStates[band]
	if ok {
		return result
	}

	plan, ok := c.vfo.bandplan[bandplan.BandName(band)]
	if !ok {
		log.Printf("Band %s not found in bandplan! (1)", band)
		return bandState{}
	}

	result = bandState{
		frequency: core.Frequency(plan.From),
		mode:      core.ModeCW,
	}
	c.lastStates[band] = result

	return result
}

func (c *offlineClient) Active() bool {
	return true
}

func (c *offlineClient) Refresh() {
	c.stateLock.RLock()
	state := c.lastState(c.currentBand)
	c.stateLock.RUnlock()

	c.vfo.emitFrequencyChanged(state.frequency)
	c.vfo.emitBandChanged(c.currentBand)
	c.vfo.emitModeChanged(state.mode)
}

func (c *offlineClient) SetFrequency(frequency core.Frequency) {
	planband := c.vfo.bandplan.ByFrequency(hamradio.Frequency(frequency))
	newBand := core.Band(planband.Name)

	c.stateLock.Lock()
	state := c.lastState(newBand)
	state.frequency = frequency
	c.lastStates[newBand] = state
	c.stateLock.Unlock()

	c.vfo.emitFrequencyChanged(frequency)

	if newBand == c.currentBand {
		return
	}
	c.currentBand = newBand
	c.vfo.emitBandChanged(c.currentBand)
}

func (c *offlineClient) SetBand(band core.Band) {
	plan, ok := c.vfo.bandplan[bandplan.BandName(band)]
	if !ok {
		log.Printf("Band %s not found in bandplan (2)", band)
		return
	}
	newBand := core.Band(plan.Name)
	if newBand == c.currentBand && !c.vfo.refreshing {
		return
	}

	c.stateLock.RLock()
	state := c.lastState(newBand)
	c.stateLock.RUnlock()

	c.vfo.emitFrequencyChanged(state.frequency)

	c.currentBand = newBand
	c.vfo.emitBandChanged(c.currentBand)
}

func (c *offlineClient) SetMode(mode core.Mode) {
	c.stateLock.Lock()
	state := c.lastState(c.currentBand)
	state.mode = mode
	c.lastStates[c.currentBand] = state
	c.stateLock.Unlock()

	c.vfo.emitModeChanged(mode)
}

func (c *offlineClient) SetXIT(active bool, offset core.Frequency) {
	c.stateLock.Lock()
	state := c.lastState(c.currentBand)
	state.xitActive = active
	state.xitOffset = offset
	c.lastStates[c.currentBand] = state
	c.stateLock.Unlock()

	c.vfo.emitXITChanged(state.xitActive, state.xitOffset)
}

func (c *offlineClient) SetPTT(active bool) {
	c.vfo.emitPTTChanged(active)
}

func (c *offlineClient) MuteAudio() {}

func (c *offlineClient) UnmuteAudio() {}

func (c *offlineClient) ToggleAudio() {}
