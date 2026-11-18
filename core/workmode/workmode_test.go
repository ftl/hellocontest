package workmode

import (
	"testing"

	"github.com/ftl/hellocontest/core/model"
	"github.com/stretchr/testify/assert"
)

type workmodeEvent struct {
	vfo      model.VFOID
	workmode model.Workmode
}

type listenerSpy struct {
	events []workmodeEvent
}

func (l *listenerSpy) WorkmodeChanged(vfo model.VFOID, workmode model.Workmode) {
	l.events = append(l.events, workmodeEvent{vfo, workmode})
}

func (l *listenerSpy) reset() {
	l.events = nil
}

func (l *listenerSpy) last() workmodeEvent {
	if len(l.events) == 0 {
		return workmodeEvent{}
	}
	return l.events[len(l.events)-1]
}

func newTestController() (*Controller, *listenerSpy) {
	c := NewController()
	l := &listenerSpy{}
	c.Notify(l)
	return c, l
}

func TestEffectiveWorkmode_SingleVFO(t *testing.T) {
	c := NewController()

	c.SetWorkmode(model.Run)
	assert.Equal(t, model.Run, c.EffectiveWorkmode(model.VFO1))
	assert.Equal(t, model.Run, c.EffectiveWorkmode(model.VFO2))

	c.SetWorkmode(model.SearchPounce)
	assert.Equal(t, model.SearchPounce, c.EffectiveWorkmode(model.VFO1))
	assert.Equal(t, model.SearchPounce, c.EffectiveWorkmode(model.VFO2))
}

func TestEffectiveWorkmode_SO2V_Run(t *testing.T) {
	c := NewController()
	c.RadioChanged("", false) // vfo2Enabled = true

	c.SetWorkmode(model.Run)
	assert.Equal(t, model.Run, c.EffectiveWorkmode(model.VFO1))
	assert.Equal(t, model.SearchPounce, c.EffectiveWorkmode(model.VFO2),
		"VFO2 must always be S&P in SO2V when global=Run")
}

func TestEffectiveWorkmode_SO2V_SP(t *testing.T) {
	c := NewController()
	c.RadioChanged("", false) // vfo2Enabled = true

	c.SetWorkmode(model.SearchPounce)
	assert.Equal(t, model.SearchPounce, c.EffectiveWorkmode(model.VFO1))
	assert.Equal(t, model.SearchPounce, c.EffectiveWorkmode(model.VFO2),
		"both VFOs must be S&P when global=S&P")
}

func TestSetWorkmode_SO2V_EmitsBothVFOs(t *testing.T) {
	c, l := newTestController()
	c.RadioChanged("", false) // SO2V

	l.reset()
	c.SetWorkmode(model.Run)

	assert.Len(t, l.events, 2)
	assert.Equal(t, workmodeEvent{model.VFO1, model.Run}, l.events[0])
	assert.Equal(t, workmodeEvent{model.VFO2, model.SearchPounce}, l.events[1])
}

func TestSetWorkmode_SingleVFO_EmitsVFO1Only(t *testing.T) {
	c, l := newTestController()
	// vfo2Enabled defaults to false

	l.reset()
	c.SetWorkmode(model.Run)

	assert.Len(t, l.events, 1)
	assert.Equal(t, workmodeEvent{model.VFO1, model.Run}, l.events[0])
}

func TestFocusChanged_SO2V_EmitsWorkmodes(t *testing.T) {
	c, l := newTestController()
	c.RadioChanged("", false) // SO2V
	c.SetWorkmode(model.Run)

	l.reset()
	c.FocusedVFOChanged(model.VFO2)

	assert.Len(t, l.events, 2)
	assert.Equal(t, workmodeEvent{model.VFO1, model.Run}, l.events[0])
	assert.Equal(t, workmodeEvent{model.VFO2, model.SearchPounce}, l.events[1])
}

func TestFocusChanged_SameVFO_NoEmit(t *testing.T) {
	c, l := newTestController()
	c.RadioChanged("", false)
	c.SetWorkmode(model.Run)

	l.reset()
	c.FocusedVFOChanged(model.VFO1) // already focused

	assert.Empty(t, l.events)
}

func TestRadioChanged_ToSingleVFO_ReEmits(t *testing.T) {
	c, l := newTestController()
	c.RadioChanged("", false) // SO2V
	c.SetWorkmode(model.Run)

	l.reset()
	c.RadioChanged("", true) // back to single VFO

	// VFO1 only, global workmode
	assert.Len(t, l.events, 1)
	assert.Equal(t, workmodeEvent{model.VFO1, model.Run}, l.events[0])
}
