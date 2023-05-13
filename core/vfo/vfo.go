package vfo

import (
	"log"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/hellocontest/core"
)

type Client interface {
	Notify(any)
	Active() bool
	Refresh()
	SetFrequency(core.Frequency)
	SetBand(core.Band)
	SetMode(core.Mode)
}

type Logbook interface {
	LastBand() core.Band
	LastMode() core.Mode
}

type VFO struct {
	Name string

	bandplan      bandplan.Bandplan
	client        Client
	offlineClient *offlineClient

	listeners []any
}

func NewVFO(name string, bandplan bandplan.Bandplan) *VFO {
	result := &VFO{
		Name:     name,
		bandplan: bandplan,
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

func (v *VFO) Notify(listener any) {
	v.listeners = append(v.listeners, listener)
}

func (v *VFO) online() bool {
	return v.client != nil && v.client.Active()
}

func (v *VFO) Refresh() {
	if !v.online() {
		v.offlineClient.Refresh()
		return
	}
	v.client.Refresh()
}

func (v *VFO) SetFrequency(frequency core.Frequency) {
	if v.online() {
		v.client.SetFrequency(frequency)
	} else {
		v.offlineClient.SetFrequency(frequency)
	}
}

func (v *VFO) SetBand(band core.Band) {
	if v.online() {
		v.client.SetBand(band)
	} else {
		v.offlineClient.SetBand(band)
	}
}

func (v *VFO) SetMode(mode core.Mode) {
	if v.online() {
		v.client.SetMode(mode)
	} else {
		v.offlineClient.SetMode(mode)
	}
}

func (v *VFO) SetLogbook(logbook Logbook) {
	log.Printf("VFO logbook changed")

	if v.online() {
		return
	}

	lastBand := logbook.LastBand()
	if lastBand != core.NoBand {
		v.offlineClient.SetBand(lastBand)
	}

	lastMode := logbook.LastMode()
	if lastMode != core.NoMode {
		v.offlineClient.SetMode(lastMode)
	}

	v.Refresh()
}

func (v *VFO) VFOFrequencyChanged(frequency core.Frequency) {
	v.offlineClient.SetFrequency(frequency)
}

func (v *VFO) VFOBandChanged(band core.Band) {
	v.offlineClient.SetBand(band)
}

func (v *VFO) VFOModeChanged(mode core.Mode) {
	v.offlineClient.SetMode(mode)
}

func (v *VFO) emitFrequencyChanged(frequency core.Frequency) {
	for _, listener := range v.listeners {
		if frequencyListener, ok := listener.(core.VFOFrequencyListener); ok {
			frequencyListener.VFOFrequencyChanged(frequency)
		}
	}
}

func (v *VFO) emitBandChanged(band core.Band) {
	for _, listener := range v.listeners {
		if bandListener, ok := listener.(core.VFOBandListener); ok {
			bandListener.VFOBandChanged(band)
		}
	}
}

func (v *VFO) emitModeChanged(mode core.Mode) {
	for _, listener := range v.listeners {
		if modeListener, ok := listener.(core.VFOModeListener); ok {
			modeListener.VFOModeChanged(mode)
		}
	}
}

type bandState struct {
	frequency core.Frequency
	mode      core.Mode
}

type offlineClient struct {
	vfo         *VFO
	currentBand core.Band
	lastStates  map[core.Band]bandState
}

func newOfflineClient(vfo *VFO) *offlineClient {
	result := &offlineClient{
		vfo:         vfo,
		currentBand: core.Band160m,
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
	state := c.lastState(c.currentBand)

	c.vfo.emitFrequencyChanged(state.frequency)
	c.vfo.emitBandChanged(c.currentBand)
	c.vfo.emitModeChanged(state.mode)
}

func (c *offlineClient) SetFrequency(frequency core.Frequency) {
	planband := c.vfo.bandplan.ByFrequency(hamradio.Frequency(frequency))
	newBand := core.Band(planband.Name)

	state := c.lastState(newBand)
	state.frequency = frequency
	c.lastStates[newBand] = state
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
	if newBand == c.currentBand {
		log.Printf("Band %s already selected!", band)
		return
	}

	state := c.lastState(newBand)
	c.vfo.emitFrequencyChanged(state.frequency)

	c.currentBand = newBand
	c.vfo.emitBandChanged(c.currentBand)
}

func (c *offlineClient) SetMode(mode core.Mode) {
	state := c.lastState(c.currentBand)
	state.mode = mode
	c.lastStates[c.currentBand] = state
	c.vfo.emitModeChanged(mode)
}
