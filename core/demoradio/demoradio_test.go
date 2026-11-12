package demoradio

import (
	"testing"

	"github.com/ftl/hamradio/bandplan"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func testClient(singleVFO bool) (*Client, *listener) {
	client := New(bandplan.IARURegion1, singleVFO, func(f func()) { f() })
	l := &listener{}
	client.Notify(l)
	return client, l
}

type listener struct {
	frequencies map[core.VFOID]core.Frequency
	bands       map[core.VFOID]core.Band
	modes       map[core.VFOID]core.Mode
}

func (l *listener) VFOFrequencyChanged(vfo core.VFOID, frequency core.Frequency) {
	if l.frequencies == nil {
		l.frequencies = make(map[core.VFOID]core.Frequency)
	}
	l.frequencies[vfo] = frequency
}

func (l *listener) VFOBandChanged(vfo core.VFOID, band core.Band) {
	if l.bands == nil {
		l.bands = make(map[core.VFOID]core.Band)
	}
	l.bands[vfo] = band
}

func (l *listener) VFOModeChanged(vfo core.VFOID, mode core.Mode) {
	if l.modes == nil {
		l.modes = make(map[core.VFOID]core.Mode)
	}
	l.modes[vfo] = mode
}

func TestClient_SetFrequency(t *testing.T) {
	tt := []struct {
		desc         string
		frequency    core.Frequency
		expectedBand core.Band
	}{
		{desc: "40m", frequency: 7028000, expectedBand: core.Band40m},
		{desc: "80m", frequency: 3535000, expectedBand: core.Band80m},
		{desc: "out of any band", frequency: 5000000, expectedBand: core.NoBand},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			client, l := testClient(true)

			client.SetFrequency(core.VFO1, tc.frequency)

			assert.Equal(t, tc.frequency, l.frequencies[core.VFO1])
			assert.Equal(t, tc.expectedBand, l.bands[core.VFO1])
		})
	}
}

func TestClient_SetBand(t *testing.T) {
	client, l := testClient(true)

	client.SetBand(core.VFO1, core.Band80m)

	assert.Equal(t, core.Band80m, l.bands[core.VFO1])
	assert.Equal(t, core.Frequency(3520000), l.frequencies[core.VFO1])
}

func TestClient_SetMode(t *testing.T) {
	client, l := testClient(true)

	client.SetMode(core.VFO1, core.ModeSSB)

	assert.Equal(t, core.ModeSSB, l.modes[core.VFO1])
}

func TestClient_SingleVFOIgnoresVFO2(t *testing.T) {
	client, l := testClient(true)

	client.SetFrequency(core.VFO2, 14030000)

	assert.True(t, client.SingleVFO())
	assert.NotContains(t, l.frequencies, core.VFO2)
}

func TestClient_SO2VAcceptsVFO2(t *testing.T) {
	client, l := testClient(false)

	client.SetFrequency(core.VFO2, 14030000)

	assert.False(t, client.SingleVFO())
	assert.Equal(t, core.Frequency(14030000), l.frequencies[core.VFO2])
}

// the demo radio must confirm a frequency that it already has, otherwise the VFO never
// leaves the state of a commanded, unconfirmed frequency
func TestClient_ConfirmsTheFrequencyItAlreadyHas(t *testing.T) {
	client, _ := testClient(true)
	counter := &countingListener{}
	client.Notify(counter)

	client.SetFrequency(core.VFO1, DefaultFrequency)

	assert.Equal(t, 1, counter.changes)
}

type countingListener struct {
	changes int
}

func (l *countingListener) VFOFrequencyChanged(core.VFOID, core.Frequency) {
	l.changes++
}
