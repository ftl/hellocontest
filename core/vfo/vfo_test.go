package vfo

import (
	"testing"
	"time"

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

func (s *visibilitySpy) VFOIncrementalTuningVisibilityChanged(vfo core.VFOID, kind core.IncrementalTuningKind, visible bool) {
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
	assert.True(t, spy.visible[core.XIT], "available XIT visible in S&P")
	assert.False(t, spy.visible[core.RIT], "unavailable RIT hidden")

	v.WorkmodeChanged(core.VFO1, core.Run)
	assert.False(t, spy.visible[core.XIT], "XIT hidden in Run")
}

func TestToggleAndShiftAvailableIncrementalTuning(t *testing.T) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	v.WorkmodeChanged(core.VFO1, core.SearchPounce)

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

// laggingClient simulates a rig that does not report the commanded frequency back immediately.
type laggingClient struct {
	commanded []core.Frequency
}

func (c *laggingClient) Notify(any)                    {}
func (c *laggingClient) Active() bool                  { return true }
func (c *laggingClient) Refresh()                      {}
func (c *laggingClient) SetCurrentVFO(core.VFOID)      {}
func (c *laggingClient) SetTXVFO(core.VFOID)           {}
func (c *laggingClient) SetBand(core.VFOID, core.Band) {}
func (c *laggingClient) SetMode(core.VFOID, core.Mode) {}
func (c *laggingClient) IncrementalTuningPerVFO() bool { return false }
func (c *laggingClient) MuteAudio(core.VFOID)          {}
func (c *laggingClient) UnmuteAudio(core.VFOID)        {}
func (c *laggingClient) ToggleAudio(core.VFOID)        {}
func (c *laggingClient) SetIncrementalTuning(core.VFOID, core.IncrementalTuningKind, bool, core.Frequency) {
}

func (c *laggingClient) SetFrequency(_ core.VFOID, frequency core.Frequency) {
	c.commanded = append(c.commanded, frequency)
}

func TestShiftFrequencyAccumulatesWhileTheRigLags(t *testing.T) {
	client := new(laggingClient)
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetClient(client)
	now := time.Now()
	start := v.currentFrequency()

	for i := range 5 {
		v.shiftFrequency(100, now.Add(time.Duration(i)*10*time.Millisecond))
	}

	assert.Equal(t, []core.Frequency{start + 100, start + 200, start + 300, start + 400, start + 500}, client.commanded)
}

func TestShiftFrequencyUsesTheRigFrequencyAfterTheTimeout(t *testing.T) {
	client := new(laggingClient)
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetClient(client)
	now := time.Now()
	start := v.currentFrequency()

	v.shiftFrequency(100, now)
	v.shiftFrequency(100, now.Add(pendingFrequencyTimeout))

	assert.Equal(t, []core.Frequency{start + 100, start + 100}, client.commanded)
}

func TestShiftFrequencyUsesTheRigFrequencyOnceItCaughtUp(t *testing.T) {
	client := new(laggingClient)
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetClient(client)
	now := time.Now()
	start := v.currentFrequency()

	v.shiftFrequency(100, now)
	v.VFOFrequencyChanged(core.VFO1, start+100)
	v.VFOFrequencyChanged(core.VFO1, start+700) // the operator turned the knob at the rig
	v.shiftFrequency(100, now.Add(10*time.Millisecond))

	assert.Equal(t, []core.Frequency{start + 100, start + 800}, client.commanded)
}

type frequencySpy struct {
	reported []core.Frequency
	tuned    []core.Frequency
}

func (s *frequencySpy) VFOFrequencyChanged(_ core.VFOID, frequency core.Frequency) {
	s.reported = append(s.reported, frequency)
}

func (s *frequencySpy) VFOTuned(_ core.VFOID, frequency core.Frequency) {
	s.tuned = append(s.tuned, frequency)
}

func TestVFOFrequencyChanged_WithoutCommand_IsATuning(t *testing.T) {
	v, spy := setupFrequencySpy()

	v.vfoFrequencyChanged(core.VFO1, 14100000, time.Now())

	assert.Equal(t, []core.Frequency{14100000}, spy.tuned, "the operator tuned the VFO")
	assert.Equal(t, []core.Frequency{14100000}, spy.reported, "the new frequency is reported")
}

func TestVFOFrequencyChanged_ConfirmingACommand_IsNoTuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.setFrequency(14200000, now)
	spy.reported = nil // the offline client confirms the command on its own

	v.vfoFrequencyChanged(core.VFO1, 14200000, now.Add(100*time.Millisecond))

	assert.Empty(t, spy.tuned, "the application commanded this frequency")
	assert.Equal(t, []core.Frequency{14200000}, spy.reported, "the new frequency is reported")
}

func TestVFOFrequencyChanged_StaleReportDuringACommand_IsIgnored(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(core.VFO1, 14050000, now) // the rig is here before the command
	v.setFrequency(14200000, now)
	spy.reported = nil
	spy.tuned = nil

	// a poll that started before the command replies with the old frequency
	v.vfoFrequencyChanged(core.VFO1, 14050000, now.Add(100*time.Millisecond))

	assert.Empty(t, spy.tuned, "a stale report is no tuning")
	assert.Empty(t, spy.reported, "a stale report reaches nobody")
}

func TestVFOFrequencyChanged_StaleReportAfterTheConfirmation_IsIgnored(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(core.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	v.vfoFrequencyChanged(core.VFO1, 14200000, now.Add(100*time.Millisecond))
	spy.reported = nil
	spy.tuned = nil

	// a poll that started before the command replies after the rig confirmed it
	v.vfoFrequencyChanged(core.VFO1, 14050000, now.Add(200*time.Millisecond))

	assert.Empty(t, spy.tuned, "a stale report is no tuning")
	assert.Empty(t, spy.reported, "a stale report reaches nobody")
}

func TestVFOFrequencyChanged_RoundedConfirmation_IsNoTuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(core.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	spy.reported = nil
	spy.tuned = nil

	// the rig rounds the commanded frequency to its own raster
	v.vfoFrequencyChanged(core.VFO1, 14199990, now.Add(100*time.Millisecond))

	assert.Empty(t, spy.tuned, "the rig follows the command")
	assert.Equal(t, []core.Frequency{14199990}, spy.reported, "the frequency of the rig is reported")
}

func TestVFOFrequencyChanged_TuningAfterARoundedConfirmation_IsATuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(core.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	v.vfoFrequencyChanged(core.VFO1, 14199990, now.Add(100*time.Millisecond)) // rounded
	spy.tuned = nil

	// the operator tunes away after the rig settled
	v.vfoFrequencyChanged(core.VFO1, 14205000, now.Add(2*time.Second))

	assert.Equal(t, []core.Frequency{14205000}, spy.tuned, "the operator tuned the VFO")
}

func TestVFOFrequencyChanged_AfterTheTimeout_IsATuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.setFrequency(14200000, now)
	spy.reported = nil // the offline client confirms the command on its own

	// the rig did not follow the command within the timeout
	v.vfoFrequencyChanged(core.VFO1, 14050000, now.Add(pendingFrequencyTimeout+time.Millisecond))

	assert.Equal(t, []core.Frequency{14050000}, spy.tuned, "the command timed out")
}

func TestVFOFrequencyChanged_TuningIsReportedBeforeTheFrequency(t *testing.T) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	order := make([]string, 0, 2)
	v.Notify(&orderSpy{order: &order})

	v.vfoFrequencyChanged(core.VFO1, 14100000, time.Now())

	// the listeners must be able to compare the new frequency with the one they know
	assert.Equal(t, []string{"tuned", "frequency"}, order)
}

func TestVFOFrequencyChanged_OfAnotherVFO_IsIgnored(t *testing.T) {
	v, spy := setupFrequencySpy()

	v.vfoFrequencyChanged(core.VFO2, 14100000, time.Now())

	assert.Empty(t, spy.tuned)
	assert.Empty(t, spy.reported)
}

func setupFrequencySpy() (*VFO, *frequencySpy) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	spy := new(frequencySpy)
	v.Notify(spy)
	return v, spy
}

type orderSpy struct {
	order *[]string
}

func (s *orderSpy) VFOFrequencyChanged(core.VFOID, core.Frequency) {
	*s.order = append(*s.order, "frequency")
}

func (s *orderSpy) VFOTuned(core.VFOID, core.Frequency) {
	*s.order = append(*s.order, "tuned")
}

func TestVFOFrequencyChanged_TuningRightAfterTheCommand_IsATuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(core.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	v.vfoFrequencyChanged(core.VFO1, 14200000, now.Add(100*time.Millisecond)) // the rig follows
	spy.tuned = nil

	// the operator tunes away immediately, well inside the command timeout
	v.vfoFrequencyChanged(core.VFO1, 14205000, now.Add(200*time.Millisecond))

	assert.Equal(t, []core.Frequency{14205000}, spy.tuned, "the rig settled, this is the operator")
}

func TestVFOFrequencyChanged_StaleReportAfterATuning_IsIgnored(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(core.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	v.vfoFrequencyChanged(core.VFO1, 14200000, now.Add(100*time.Millisecond))
	v.vfoFrequencyChanged(core.VFO1, 14205000, now.Add(200*time.Millisecond)) // the operator
	spy.tuned = nil
	spy.reported = nil

	// the late reply of a poll that started before the command
	v.vfoFrequencyChanged(core.VFO1, 14050000, now.Add(300*time.Millisecond))

	assert.Empty(t, spy.tuned, "a stale report is no tuning")
	assert.Empty(t, spy.reported, "a stale report reaches nobody")
}

func TestShiftFrequency_IsATuningOfTheOperator(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(core.VFO1, 14050000, now)
	spy.tuned = nil

	// a dial or the keyboard shifts the frequency, the rig follows
	v.shiftFrequency(1000, now)
	v.vfoFrequencyChanged(core.VFO1, 14051000, now.Add(100*time.Millisecond))

	assert.Equal(t, []core.Frequency{14051000}, spy.tuned, "the operator tuned with the dial")
}

func TestShiftFrequency_EndsACommandInFlight(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(core.VFO1, 14050000, now)
	v.setFrequency(14200000, now) // a spot was selected
	spy.tuned = nil

	// the operator grabs the dial before the rig confirmed the commanded frequency
	v.shiftFrequency(1000, now.Add(100*time.Millisecond))
	v.vfoFrequencyChanged(core.VFO1, 14201000, now.Add(200*time.Millisecond))

	assert.Equal(t, []core.Frequency{14201000}, spy.tuned, "the operator took over")
}
