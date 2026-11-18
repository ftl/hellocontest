package demoradio

import (
	"testing"

	"github.com/ftl/hamradio/bandplan"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core/model"
)

func testClient(singleVFO bool) (*Client, *listener) {
	client := New(bandplan.IARURegion1, singleVFO, func(f func()) { f() })
	l := &listener{}
	client.Notify(l)
	return client, l
}

type listener struct {
	frequencies map[model.VFOID]model.Frequency
	bands       map[model.VFOID]model.Band
	modes       map[model.VFOID]model.Mode
}

func (l *listener) VFOFrequencyChanged(vfo model.VFOID, frequency model.Frequency) {
	if l.frequencies == nil {
		l.frequencies = make(map[model.VFOID]model.Frequency)
	}
	l.frequencies[vfo] = frequency
}

func (l *listener) VFOBandChanged(vfo model.VFOID, band model.Band) {
	if l.bands == nil {
		l.bands = make(map[model.VFOID]model.Band)
	}
	l.bands[vfo] = band
}

func (l *listener) VFOModeChanged(vfo model.VFOID, mode model.Mode) {
	if l.modes == nil {
		l.modes = make(map[model.VFOID]model.Mode)
	}
	l.modes[vfo] = mode
}

func TestClient_SetFrequency(t *testing.T) {
	tt := []struct {
		desc         string
		frequency    model.Frequency
		expectedBand model.Band
	}{
		{desc: "40m", frequency: 7028000, expectedBand: model.Band40m},
		{desc: "80m", frequency: 3535000, expectedBand: model.Band80m},
		{desc: "out of any band", frequency: 5000000, expectedBand: model.NoBand},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			client, l := testClient(true)

			client.SetFrequency(model.VFO1, tc.frequency)

			assert.Equal(t, tc.frequency, l.frequencies[model.VFO1])
			assert.Equal(t, tc.expectedBand, l.bands[model.VFO1])
		})
	}
}

func TestClient_SetBand(t *testing.T) {
	client, l := testClient(true)

	client.SetBand(model.VFO1, model.Band80m)

	assert.Equal(t, model.Band80m, l.bands[model.VFO1])
	assert.Equal(t, model.Frequency(3520000), l.frequencies[model.VFO1])
}

func TestClient_SetMode(t *testing.T) {
	client, l := testClient(true)

	client.SetMode(model.VFO1, model.ModeSSB)

	assert.Equal(t, model.ModeSSB, l.modes[model.VFO1])
}

func TestClient_SingleVFOIgnoresVFO2(t *testing.T) {
	client, l := testClient(true)

	client.SetFrequency(model.VFO2, 14030000)

	assert.True(t, client.SingleVFO())
	assert.NotContains(t, l.frequencies, model.VFO2)
}

func TestClient_SO2VAcceptsVFO2(t *testing.T) {
	client, l := testClient(false)

	client.SetFrequency(model.VFO2, 14030000)

	assert.False(t, client.SingleVFO())
	assert.Equal(t, model.Frequency(14030000), l.frequencies[model.VFO2])
}

// the demo radio must confirm a frequency that it already has, otherwise the VFO never
// leaves the state of a commanded, unconfirmed frequency
func TestClient_ConfirmsTheFrequencyItAlreadyHas(t *testing.T) {
	client, _ := testClient(true)
	counter := &countingListener{}
	client.Notify(counter)

	client.SetFrequency(model.VFO1, DefaultFrequency)

	assert.Equal(t, 1, counter.changes)
}

type countingListener struct {
	changes int
}

func (l *countingListener) VFOFrequencyChanged(model.VFOID, model.Frequency) {
	l.changes++
}
