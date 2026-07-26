package radio

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

type vfoSpy struct {
	setActive    bool
	activeKind   core.IncrementalTuningKind
	active       bool
	toggled      bool
	shiftedAvail core.Frequency
}

func (v *vfoSpy) Name() string                                                          { return "" }
func (v *vfoSpy) Notify(any)                                                            {}
func (v *vfoSpy) Refresh()                                                              {}
func (v *vfoSpy) SetFrequency(core.Frequency)                                           {}
func (v *vfoSpy) ShiftFrequency(core.Frequency)                                         {}
func (v *vfoSpy) SetBand(core.Band)                                                     {}
func (v *vfoSpy) SetMode(core.Mode)                                                     {}
func (v *vfoSpy) SetIncrementalTuning(core.IncrementalTuningKind, bool, core.Frequency) {}
func (v *vfoSpy) ShiftOffset(core.IncrementalTuningKind, core.Frequency)                {}
func (v *vfoSpy) IncrementalTuningActive(core.IncrementalTuningKind) bool               { return v.active }
func (v *vfoSpy) SetIncrementalTuningActive(kind core.IncrementalTuningKind, active bool) {
	v.setActive = true
	v.activeKind = kind
	v.active = active
}
func (v *vfoSpy) ToggleIncrementalTuning()                            { v.toggled = true }
func (v *vfoSpy) ShiftAvailableIncrementalTuning(sign core.Frequency) { v.shiftedAvail = sign }

func newTestController(v1, v2 core.VFO) *Controller {
	c := NewController(nil, nil, nil)
	c.SetVFO(core.VFO1, v1)
	c.SetVFO(core.VFO2, v2)
	return c
}

func TestIncrementalTuningRouting_MainVFOOnly(t *testing.T) {
	v1, v2 := &vfoSpy{}, &vfoSpy{}
	c := newTestController(v1, v2)
	c.SetCurrentVFO(core.VFO2)

	c.SetIncrementalTuningActive(core.VFO2, core.XIT, true)
	assert.False(t, v1.setActive, "explicit setting on VFO2 is ignored")
	assert.False(t, v2.setActive)

	c.ToggleIncrementalTuning()
	assert.False(t, v1.toggled, "toggle on VFO2 is ignored")
	assert.False(t, v2.toggled)
}

func TestIncrementalTuningRouting_PerVFO(t *testing.T) {
	v1, v2 := &vfoSpy{}, &vfoSpy{}
	c := newTestController(v1, v2)
	c.incrementalTuningPerVFO = true
	c.SetCurrentVFO(core.VFO2)

	c.SetIncrementalTuningActive(core.VFO2, core.XIT, true)
	assert.True(t, v2.setActive, "explicit VFO2 targets VFO2 when flag is on")
	assert.False(t, v1.setActive)

	c.ToggleIncrementalTuning()
	assert.True(t, v2.toggled, "focused action targets the current VFO")

	c.IncrementalTuningUp()
	assert.Equal(t, core.Frequency(1), v2.shiftedAvail)
}
