package vfo

import (
	"testing"

	"github.com/ftl/hamradio/bandplan"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestSetBandplanSwapsThePlan(t *testing.T) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, nil)
	assert.Equal(t, 1810000.0, float64(v.bandplan[bandplan.Band160m].From))

	v.SetBandplan(bandplan.IARURegion2)
	assert.Equal(t, 1800000.0, float64(v.bandplan[bandplan.Band160m].From))
}

func TestBandNameConversion(t *testing.T) {
	bndpln := bandplan.IARURegion1

	for band, plan := range bndpln {
		assert.Equal(t, band, plan.Name)
	}

	for _, band := range core.Bands {
		plan, ok := bndpln[bandplan.BandName(band)]
		assert.True(t, ok, band)
		assert.Equal(t, string(band), string(plan.Name))
	}

}

func TestShiftFrequency(t *testing.T) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	start := v.currentFrequency()

	v.ShiftFrequency(250)
	assert.Equal(t, start+250, v.currentFrequency())

	v.ShiftFrequency(-250)
	assert.Equal(t, start, v.currentFrequency())
}

func TestShiftIncrementalTuning(t *testing.T) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })

	v.ShiftIncrementalTuning(core.XIT, 100)
	assert.Equal(t, core.Frequency(100), v.state[core.XIT].offset)

	v.ShiftIncrementalTuning(core.XIT, -30)
	assert.Equal(t, core.Frequency(70), v.state[core.XIT].offset)

	v.ShiftIncrementalTuning(core.RIT, 40)
	assert.Equal(t, core.Frequency(40), v.state[core.RIT].offset)
	assert.Equal(t, core.Frequency(70), v.state[core.XIT].offset)
}

type visibilitySpy struct {
	visible map[core.IncrementalTuningKind]bool
}

func (s *visibilitySpy) IncrementalTuningVisibilityChanged(vfo core.VFOID, kind core.IncrementalTuningKind, visible bool) {
	if s.visible == nil {
		s.visible = make(map[core.IncrementalTuningKind]bool)
	}
	s.visible[kind] = visible
}

func TestIncrementalTuningVisibility(t *testing.T) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	spy := &visibilitySpy{}
	v.Notify(spy)

	v.WorkmodeChanged(core.VFO1, core.SearchPounce)
	v.IncrementalTuningAvailabilityChanged(core.VFO1, core.XIT, true)
	assert.True(t, spy.visible[core.XIT], "available XIT visible in S&P")
	assert.False(t, spy.visible[core.RIT], "unavailable RIT hidden")

	v.WorkmodeChanged(core.VFO1, core.Run)
	assert.False(t, spy.visible[core.XIT], "XIT hidden in Run")
}

func TestToggleAndShiftAvailableIncrementalTuning(t *testing.T) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	v.WorkmodeChanged(core.VFO1, core.SearchPounce)
	v.IncrementalTuningAvailabilityChanged(core.VFO1, core.XIT, true)

	kind, ok := v.availableIncrementalTuningKind()
	assert.True(t, ok)
	assert.Equal(t, core.XIT, kind)

	v.ToggleAvailableIncrementalTuning()
	assert.True(t, v.IncrementalTuningActive(core.XIT), "toggle enables the available kind")
	v.ToggleAvailableIncrementalTuning()
	assert.False(t, v.IncrementalTuningActive(core.XIT), "toggle again disables it")

	v.ShiftAvailableIncrementalTuning(1)
	assert.Equal(t, core.DefaultXITShift, v.state[core.XIT].offset)
	v.ShiftAvailableIncrementalTuning(-1)
	assert.Equal(t, core.Frequency(0), v.state[core.XIT].offset)
}

func TestWorkmodeGatesIncrementalTuning(t *testing.T) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetIncrementalTuningActive(core.XIT, true)
	v.SetIncrementalTuningActive(core.RIT, true)

	v.WorkmodeChanged(core.VFO1, core.SearchPounce)
	assert.True(t, v.state[core.XIT].actualActive, "XIT active in S&P")
	assert.False(t, v.state[core.RIT].actualActive, "RIT forced off in S&P")

	v.WorkmodeChanged(core.VFO1, core.Run)
	assert.False(t, v.state[core.XIT].actualActive, "XIT forced off in Run")
	assert.True(t, v.state[core.RIT].actualActive, "RIT active in Run")

	assert.True(t, v.IncrementalTuningActive(core.XIT), "XIT intent survives the round-trip")
	assert.True(t, v.IncrementalTuningActive(core.RIT), "RIT intent survives the round-trip")
}
