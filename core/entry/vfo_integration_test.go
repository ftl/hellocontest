package entry_test

import (
	"testing"

	"github.com/ftl/hamradio/bandplan"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/model"
	"github.com/ftl/hellocontest/core/vfo"
)

// The spot list, the VFO and the entry controller work together here, because the
// interesting decisions happen between the components, not inside one of them.

func TestVFOAndEntry_TuningAfterASpotSelection_ClearsTheInput(t *testing.T) {
	s, v, rig := setupEntryWithVFO(t)

	rig.reportFrequency(14050000)
	s.EntrySelected(model.VFO1, spotAt(14200000))
	rig.reportFrequency(14200000) // the rig follows the command

	s.resetSpies()
	rig.reportFrequency(14205000) // the operator tunes away

	assert.True(t, s.view.wasCalledWith("SetCallsign", model.VFO1, ""),
		"the input belongs to the frequency of the spot, therefore it gets cleared")
	_ = v
}

func TestVFOAndEntry_SpotSelection_KeepsTheInput(t *testing.T) {
	s, _, rig := setupEntryWithVFO(t)

	rig.reportFrequency(14050000)
	s.EntrySelected(model.VFO1, spotAt(14200000))

	s.resetSpies()
	rig.reportFrequency(14200000) // the rig follows the command
	rig.reportFrequency(14050000) // a poll that started before the command replies late

	assert.False(t, s.view.wasCalledWith("SetCallsign", model.VFO1, ""),
		"neither the confirmation nor a stale report clears the input")
}

func spotAt(frequency model.Frequency) model.BandmapEntry {
	return model.BandmapEntry{
		Call:      model.MustParseCallsign("DL1ABC"),
		Frequency: frequency,
		Band:      model.Band20m,
		Mode:      model.ModeCW,
	}
}

func setupEntryWithVFO(t *testing.T) (*Scenario, *vfo.VFO, *rigStub) {
	t.Helper()
	s := NewScenario(t).WithClassicExchange()

	rig := new(rigStub)
	v := vfo.NewVFO(model.VFO1, "VFO 1", clock.New(), bandplan.IARURegion1, nil, func(f func()) { f() })
	rig.vfo = v
	v.SetClient(rig)
	v.Notify(s.controller)
	s.controller.SetVFO(model.VFO1, v)

	return s, v, rig
}

type rigStub struct {
	vfo       *vfo.VFO
	frequency model.Frequency
}

func (r *rigStub) reportFrequency(frequency model.Frequency) {
	r.frequency = frequency
	r.vfo.VFOFrequencyChanged(model.VFO1, frequency)
}

func (r *rigStub) Notify(any)                {}
func (r *rigStub) Active() bool              { return true }
func (r *rigStub) Refresh()                  { r.vfo.VFOFrequencyChanged(model.VFO1, r.frequency) }
func (r *rigStub) SetCurrentVFO(model.VFOID) {}
func (r *rigStub) SetTXVFO(model.VFOID)      {}
func (r *rigStub) SetFrequency(_ model.VFOID, frequency model.Frequency) {
	// a real rig needs time and answers with its own poll, see reportFrequency
}
func (r *rigStub) SetBand(model.VFOID, model.Band) {}
func (r *rigStub) SetMode(model.VFOID, model.Mode) {}
func (r *rigStub) IncrementalTuningPerVFO() bool   { return false }
func (r *rigStub) SetIncrementalTuning(model.VFOID, model.IncrementalTuningKind, bool, model.Frequency) {
}
func (r *rigStub) MuteAudio(model.VFOID)   {}
func (r *rigStub) UnmuteAudio(model.VFOID) {}
func (r *rigStub) ToggleAudio(model.VFOID) {}
