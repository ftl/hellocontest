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
	online := v.online() // v.online() might change concurrently
	log.Printf("%s set frequency %f (online: %t)", v.Name, frequency, online)
	if online {
		v.client.SetFrequency(frequency)
	}
	v.offlineClient.SetFrequency(frequency, !online)
}

func (v *VFO) SetBand(band core.Band) {
	online := v.online() // v.online() might change concurrently
	log.Printf("%s set band %s (online: %t)", v.Name, band, online)
	if online {
		v.client.SetBand(band)
	}
	v.offlineClient.SetBand(band, !online)
}

func (v *VFO) SetMode(mode core.Mode) {
	online := v.online() // v.online() might change concurrently
	log.Printf("%s set mode %s (online: %t)", v.Name, mode, online)
	if online {
		v.client.SetMode(mode)
	}
	v.offlineClient.SetMode(mode, !online)
}

func (v *VFO) emitFrequencyChanged(f core.Frequency) {
	for _, listener := range v.listeners {
		if frequencyListener, ok := listener.(core.VFOFrequencyListener); ok {
			frequencyListener.VFOFrequencyChanged(f)
		}
	}
}

func (v *VFO) emitBandChanged(b core.Band) {
	for _, listener := range v.listeners {
		if bandListener, ok := listener.(core.VFOBandListener); ok {
			bandListener.VFOBandChanged(b)
		}
	}
}

func (v *VFO) emitModeChanged(m core.Mode) {
	for _, listener := range v.listeners {
		if modeListener, ok := listener.(core.VFOModeListener); ok {
			modeListener.VFOModeChanged(m)
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
		log.Printf("last state found for %s: %f %s", band, result.frequency, result.mode)
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
	log.Printf("last state %s: %f %s", band, result.frequency, result.mode)
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

func (c *offlineClient) SetFrequency(frequency core.Frequency, offline bool) {
	planband := c.vfo.bandplan.ByFrequency(hamradio.Frequency(frequency))
	newBand := core.Band(planband.Name)

	state := c.lastState(newBand)
	state.frequency = frequency
	c.lastStates[newBand] = state
	if offline {
		c.vfo.emitFrequencyChanged(frequency)
	}

	if newBand == c.currentBand {
		return
	}
	c.currentBand = newBand
	if offline {
		c.vfo.emitBandChanged(c.currentBand)
	}
}

func (c *offlineClient) SetBand(band core.Band, offline bool) {
	log.Printf("offline client SetBand %s (offline %t)", band, offline)
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
	if offline {
		c.vfo.emitFrequencyChanged(state.frequency)
	}

	c.currentBand = newBand
	if offline {
		c.vfo.emitBandChanged(c.currentBand)
	}
}

func (c *offlineClient) SetMode(mode core.Mode, offline bool) {
	state := c.lastState(c.currentBand)
	state.mode = mode
	c.lastStates[c.currentBand] = state
	if offline {
		c.vfo.emitModeChanged(mode)
	}
}
