package vfo

import (
	"testing"
	"time"

	"github.com/ftl/hamradio/bandplan"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/model"
)

func TestSetBandplanSwapsThePlan(t *testing.T) {
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, nil)
	assert.Equal(t, 1810000.0, float64(v.bandplan[bandplan.Band160m].From))

	v.SetBandplan(bandplan.IARURegion2)
	assert.Equal(t, 1800000.0, float64(v.bandplan[bandplan.Band160m].From))
}

func TestBandNameConversion(t *testing.T) {
	bndpln := bandplan.IARURegion1

	for band, plan := range bndpln {
		assert.Equal(t, band, plan.Name)
	}

	for _, band := range model.Bands {
		plan, ok := bndpln[bandplan.BandName(band)]
		assert.True(t, ok, band)
		assert.Equal(t, string(band), string(plan.Name))
	}

}

func TestShiftFrequency(t *testing.T) {
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	start := v.currentFrequency()

	v.ShiftFrequency(250)
	assert.Equal(t, start+250, v.currentFrequency())

	v.ShiftFrequency(-250)
	assert.Equal(t, start, v.currentFrequency())
}

func TestShiftIncrementalTuning(t *testing.T) {
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })

	v.ShiftIncrementalTuning(model.XIT, 100)
	assert.Equal(t, model.Frequency(100), v.state[model.XIT].offset)

	v.ShiftIncrementalTuning(model.XIT, -30)
	assert.Equal(t, model.Frequency(70), v.state[model.XIT].offset)

	v.ShiftIncrementalTuning(model.RIT, 40)
	assert.Equal(t, model.Frequency(40), v.state[model.RIT].offset)
	assert.Equal(t, model.Frequency(70), v.state[model.XIT].offset)
}

type visibilitySpy struct {
	visible map[model.IncrementalTuningKind]bool
}

func (s *visibilitySpy) VFOIncrementalTuningVisibilityChanged(vfo model.VFOID, kind model.IncrementalTuningKind, visible bool) {
	if s.visible == nil {
		s.visible = make(map[model.IncrementalTuningKind]bool)
	}
	s.visible[kind] = visible
}

func TestIncrementalTuningVisibility(t *testing.T) {
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	spy := &visibilitySpy{}
	v.Notify(spy)

	v.WorkmodeChanged(model.VFO1, model.SearchPounce)
	assert.True(t, spy.visible[model.XIT], "available XIT visible in S&P")
	assert.False(t, spy.visible[model.RIT], "unavailable RIT hidden")

	v.WorkmodeChanged(model.VFO1, model.Run)
	assert.False(t, spy.visible[model.XIT], "XIT hidden in Run")
}

func TestToggleAndShiftAvailableIncrementalTuning(t *testing.T) {
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	v.WorkmodeChanged(model.VFO1, model.SearchPounce)

	kind, ok := v.availableIncrementalTuningKind()
	assert.True(t, ok)
	assert.Equal(t, model.XIT, kind)

	v.ToggleAvailableIncrementalTuning()
	assert.True(t, v.IncrementalTuningActive(model.XIT), "toggle enables the available kind")
	v.ToggleAvailableIncrementalTuning()
	assert.False(t, v.IncrementalTuningActive(model.XIT), "toggle again disables it")

	v.ShiftAvailableIncrementalTuning(1)
	assert.Equal(t, core.DefaultXITShift, v.state[model.XIT].offset)
	v.ShiftAvailableIncrementalTuning(-1)
	assert.Equal(t, model.Frequency(0), v.state[model.XIT].offset)
}

func TestWorkmodeGatesIncrementalTuning(t *testing.T) {
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetIncrementalTuningActive(model.XIT, true)
	v.SetIncrementalTuningActive(model.RIT, true)

	v.WorkmodeChanged(model.VFO1, model.SearchPounce)
	assert.True(t, v.state[model.XIT].actualActive, "XIT active in S&P")
	assert.False(t, v.state[model.RIT].actualActive, "RIT forced off in S&P")

	v.WorkmodeChanged(model.VFO1, model.Run)
	assert.False(t, v.state[model.XIT].actualActive, "XIT forced off in Run")
	assert.True(t, v.state[model.RIT].actualActive, "RIT active in Run")

	assert.True(t, v.IncrementalTuningActive(model.XIT), "XIT intent survives the round-trip")
	assert.True(t, v.IncrementalTuningActive(model.RIT), "RIT intent survives the round-trip")
}

// laggingClient simulates a rig that does not report the commanded frequency back immediately.
type laggingClient struct {
	commanded []model.Frequency
}

func (c *laggingClient) Notify(any)                      {}
func (c *laggingClient) Active() bool                    { return true }
func (c *laggingClient) Refresh()                        {}
func (c *laggingClient) SetCurrentVFO(model.VFOID)       {}
func (c *laggingClient) SetTXVFO(model.VFOID)            {}
func (c *laggingClient) SetBand(model.VFOID, model.Band) {}
func (c *laggingClient) SetMode(model.VFOID, model.Mode) {}
func (c *laggingClient) IncrementalTuningPerVFO() bool   { return false }
func (c *laggingClient) MuteAudio(model.VFOID)           {}
func (c *laggingClient) UnmuteAudio(model.VFOID)         {}
func (c *laggingClient) ToggleAudio(model.VFOID)         {}
func (c *laggingClient) SetIncrementalTuning(model.VFOID, model.IncrementalTuningKind, bool, model.Frequency) {
}

func (c *laggingClient) SetFrequency(_ model.VFOID, frequency model.Frequency) {
	c.commanded = append(c.commanded, frequency)
}

func TestShiftFrequencyAccumulatesWhileTheRigLags(t *testing.T) {
	client := new(laggingClient)
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetClient(client)
	now := time.Now()
	start := v.currentFrequency()

	for i := range 5 {
		v.shiftFrequency(100, now.Add(time.Duration(i)*10*time.Millisecond))
	}

	assert.Equal(t, []model.Frequency{start + 100, start + 200, start + 300, start + 400, start + 500}, client.commanded)
}

func TestShiftFrequencyUsesTheRigFrequencyAfterTheTimeout(t *testing.T) {
	client := new(laggingClient)
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetClient(client)
	now := time.Now()
	start := v.currentFrequency()

	v.shiftFrequency(100, now)
	v.shiftFrequency(100, now.Add(pendingFrequencyTimeout))

	assert.Equal(t, []model.Frequency{start + 100, start + 100}, client.commanded)
}

func TestShiftFrequencyUsesTheRigFrequencyOnceItCaughtUp(t *testing.T) {
	client := new(laggingClient)
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetClient(client)
	now := time.Now()
	start := v.currentFrequency()

	v.shiftFrequency(100, now)
	v.VFOFrequencyChanged(model.VFO1, start+100)
	v.VFOFrequencyChanged(model.VFO1, start+700) // the operator turned the knob at the rig
	v.shiftFrequency(100, now.Add(10*time.Millisecond))

	assert.Equal(t, []model.Frequency{start + 100, start + 800}, client.commanded)
}

type frequencySpy struct {
	reported []model.Frequency
	tuned    []model.Frequency
}

func (s *frequencySpy) VFOFrequencyChanged(_ model.VFOID, frequency model.Frequency) {
	s.reported = append(s.reported, frequency)
}

func (s *frequencySpy) VFOTuned(_ model.VFOID, frequency model.Frequency) {
	s.tuned = append(s.tuned, frequency)
}

func TestVFOFrequencyChanged_WithoutCommand_IsATuning(t *testing.T) {
	v, spy := setupFrequencySpy()

	v.vfoFrequencyChanged(model.VFO1, 14100000, time.Now())

	assert.Equal(t, []model.Frequency{14100000}, spy.tuned, "the operator tuned the VFO")
	assert.Equal(t, []model.Frequency{14100000}, spy.reported, "the new frequency is reported")
}

func TestVFOFrequencyChanged_ConfirmingACommand_IsNoTuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.setFrequency(14200000, now)
	spy.reported = nil // the offline client confirms the command on its own

	v.vfoFrequencyChanged(model.VFO1, 14200000, now.Add(100*time.Millisecond))

	assert.Empty(t, spy.tuned, "the application commanded this frequency")
	assert.Equal(t, []model.Frequency{14200000}, spy.reported, "the new frequency is reported")
}

func TestVFOFrequencyChanged_StaleReportDuringACommand_IsIgnored(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now) // the rig is here before the command
	v.setFrequency(14200000, now)
	spy.reported = nil
	spy.tuned = nil

	// a poll that started before the command replies with the old frequency
	v.vfoFrequencyChanged(model.VFO1, 14050000, now.Add(100*time.Millisecond))

	assert.Empty(t, spy.tuned, "a stale report is no tuning")
	assert.Empty(t, spy.reported, "a stale report reaches nobody")
}

func TestVFOFrequencyChanged_StaleReportAfterTheConfirmation_IsIgnored(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	v.vfoFrequencyChanged(model.VFO1, 14200000, now.Add(100*time.Millisecond))
	spy.reported = nil
	spy.tuned = nil

	// a poll that started before the command replies after the rig confirmed it
	v.vfoFrequencyChanged(model.VFO1, 14050000, now.Add(200*time.Millisecond))

	assert.Empty(t, spy.tuned, "a stale report is no tuning")
	assert.Empty(t, spy.reported, "a stale report reaches nobody")
}

func TestVFOFrequencyChanged_RoundedConfirmation_IsNoTuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	spy.reported = nil
	spy.tuned = nil

	// the rig rounds the commanded frequency to its own raster
	v.vfoFrequencyChanged(model.VFO1, 14199990, now.Add(100*time.Millisecond))

	assert.Empty(t, spy.tuned, "the rig follows the command")
	assert.Equal(t, []model.Frequency{14199990}, spy.reported, "the frequency of the rig is reported")
}

func TestVFOFrequencyChanged_TuningAfterARoundedConfirmation_IsATuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	v.vfoFrequencyChanged(model.VFO1, 14199990, now.Add(100*time.Millisecond)) // rounded
	spy.tuned = nil

	// the operator tunes away after the rig settled
	v.vfoFrequencyChanged(model.VFO1, 14205000, now.Add(2*time.Second))

	assert.Equal(t, []model.Frequency{14205000}, spy.tuned, "the operator tuned the VFO")
}

func TestVFOFrequencyChanged_AfterTheTimeout_IsATuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.setFrequency(14200000, now)
	spy.reported = nil // the offline client confirms the command on its own

	// the rig did not follow the command within the timeout
	v.vfoFrequencyChanged(model.VFO1, 14050000, now.Add(pendingFrequencyTimeout+time.Millisecond))

	assert.Equal(t, []model.Frequency{14050000}, spy.tuned, "the command timed out")
}

func TestVFOFrequencyChanged_TuningIsReportedBeforeTheFrequency(t *testing.T) {
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	order := make([]string, 0, 2)
	v.Notify(&orderSpy{order: &order})

	v.vfoFrequencyChanged(model.VFO1, 14100000, time.Now())

	// the listeners must be able to compare the new frequency with the one they know
	assert.Equal(t, []string{"tuned", "frequency"}, order)
}

func TestVFOFrequencyChanged_OfAnotherVFO_IsIgnored(t *testing.T) {
	v, spy := setupFrequencySpy()

	v.vfoFrequencyChanged(model.VFO2, 14100000, time.Now())

	assert.Empty(t, spy.tuned)
	assert.Empty(t, spy.reported)
}

func setupFrequencySpy() (*VFO, *frequencySpy) {
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	spy := new(frequencySpy)
	v.Notify(spy)
	return v, spy
}

type orderSpy struct {
	order *[]string
}

func (s *orderSpy) VFOFrequencyChanged(model.VFOID, model.Frequency) {
	*s.order = append(*s.order, "frequency")
}

func (s *orderSpy) VFOTuned(model.VFOID, model.Frequency) {
	*s.order = append(*s.order, "tuned")
}

func TestVFOFrequencyChanged_TuningRightAfterTheCommand_IsATuning(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	v.vfoFrequencyChanged(model.VFO1, 14200000, now.Add(100*time.Millisecond)) // the rig follows
	spy.tuned = nil

	// the operator tunes away immediately, well inside the command timeout
	v.vfoFrequencyChanged(model.VFO1, 14205000, now.Add(200*time.Millisecond))

	assert.Equal(t, []model.Frequency{14205000}, spy.tuned, "the rig settled, this is the operator")
}

func TestVFOFrequencyChanged_StaleReportAfterATuning_IsIgnored(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now)
	v.setFrequency(14200000, now)
	v.vfoFrequencyChanged(model.VFO1, 14200000, now.Add(100*time.Millisecond))
	v.vfoFrequencyChanged(model.VFO1, 14205000, now.Add(200*time.Millisecond)) // the operator
	spy.tuned = nil
	spy.reported = nil

	// the late reply of a poll that started before the command
	v.vfoFrequencyChanged(model.VFO1, 14050000, now.Add(300*time.Millisecond))

	assert.Empty(t, spy.tuned, "a stale report is no tuning")
	assert.Empty(t, spy.reported, "a stale report reaches nobody")
}

func TestShiftFrequency_IsATuningOfTheOperator(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now)
	spy.tuned = nil

	// a dial or the keyboard shifts the frequency, the rig follows
	v.shiftFrequency(1000, now)
	v.vfoFrequencyChanged(model.VFO1, 14051000, now.Add(100*time.Millisecond))

	assert.Equal(t, []model.Frequency{14051000}, spy.tuned, "the operator tuned with the dial")
}

func TestShiftFrequency_EndsACommandInFlight(t *testing.T) {
	v, spy := setupFrequencySpy()
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now)
	v.setFrequency(14200000, now) // a spot was selected
	spy.tuned = nil

	// the operator grabs the dial before the rig confirmed the commanded frequency
	v.shiftFrequency(1000, now.Add(100*time.Millisecond))
	v.vfoFrequencyChanged(model.VFO1, 14201000, now.Add(200*time.Millisecond))

	assert.Equal(t, []model.Frequency{14201000}, spy.tuned, "the operator took over")
}

// refreshSpy is an active client that records what the VFO commands.
type refreshSpy struct {
	laggingClient
	refreshCount int
	bands        []model.Band
	modes        []model.Mode
}

func (c *refreshSpy) Refresh()                            { c.refreshCount++ }
func (c *refreshSpy) SetBand(_ model.VFOID, b model.Band) { c.bands = append(c.bands, b) }
func (c *refreshSpy) SetMode(_ model.VFOID, m model.Mode) { c.modes = append(c.modes, m) }

type logbookStub struct {
	band model.Band
	mode model.Mode
}

func (l *logbookStub) LastBand() model.Band { return l.band }
func (l *logbookStub) LastMode() model.Mode { return l.mode }

func TestLogbookLoadedRefreshesTheOnlineRigWithoutRetuningIt(t *testing.T) {
	logbook := &logbookStub{band: model.Band160m, mode: model.ModeCW}
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, logbook, nil)
	client := &refreshSpy{}
	v.SetClient(client)
	client.refreshCount = 0

	v.LogbookLoaded()

	assert.Equal(t, 1, client.refreshCount, "the listeners need the current state of the rig again")
	assert.Empty(t, client.bands, "the band of the logbook must not be sent to the rig")
	assert.Empty(t, client.modes, "the mode of the logbook must not be sent to the rig")
}

type bandSpy struct {
	bands []model.Band
}

func (s *bandSpy) VFOBandChanged(_ model.VFOID, band model.Band) {
	s.bands = append(s.bands, band)
}

func TestRefreshDoesNotLeaveTheVFOInRefreshingState(t *testing.T) {
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	spy := new(bandSpy)
	v.Notify(spy)

	v.Refresh()
	currentBand := spy.bands[len(spy.bands)-1]
	spy.bands = nil

	v.SetBand(currentBand)

	assert.Empty(t, spy.bands, "setting the band that is already selected must not emit a change")
}

func TestFrequencyIsTheCommandedFrequencyWhileTheRigLags(t *testing.T) {
	client := new(laggingClient)
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetClient(client)
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now)

	// a spot was selected, the rig did not report the new frequency yet
	v.setFrequency(14200000, now)

	assert.Equal(t, model.Frequency(14200000), v.baseFrequency(now.Add(100*time.Millisecond)), "the VFO is on its way to the commanded frequency")
}

func TestFrequencyIsTheRigFrequencyAfterTheTimeout(t *testing.T) {
	client := new(laggingClient)
	v := NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	v.SetClient(client)
	now := time.Now()
	v.vfoFrequencyChanged(model.VFO1, 14050000, now)

	v.setFrequency(14200000, now)

	assert.Equal(t, model.Frequency(14050000), v.baseFrequency(now.Add(pendingFrequencyTimeout+time.Millisecond)), "the rig did not follow the command")
}
