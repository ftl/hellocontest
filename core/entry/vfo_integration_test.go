package entry_test

import (
	"testing"

	"github.com/ftl/hamradio/bandplan"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/vfo"
)

// The spot list, the VFO and the entry controller work together here, because the
// interesting decisions happen between the components, not inside one of them.

func TestVFOAndEntry_TuningAfterASpotSelection_ClearsTheInput(t *testing.T) {
	s, v, rig := setupEntryWithVFO(t)

	rig.reportFrequency(14050000)
	s.EntrySelected(core.VFO1, spotAt(14200000))
	rig.reportFrequency(14200000) // the rig follows the command

	s.resetSpies()
	rig.reportFrequency(14205000) // the operator tunes away

	assert.True(t, s.view.wasCalledWith("SetCallsign", core.VFO1, ""),
		"the input belongs to the frequency of the spot, therefore it gets cleared")
	_ = v
}

func TestVFOAndEntry_SpotSelection_KeepsTheInput(t *testing.T) {
	s, _, rig := setupEntryWithVFO(t)

	rig.reportFrequency(14050000)
	s.EntrySelected(core.VFO1, spotAt(14200000))

	s.resetSpies()
	rig.reportFrequency(14200000) // the rig follows the command
	rig.reportFrequency(14050000) // a poll that started before the command replies late

	assert.False(t, s.view.wasCalledWith("SetCallsign", core.VFO1, ""),
		"neither the confirmation nor a stale report clears the input")
}

func spotAt(frequency core.Frequency) core.BandmapEntry {
	return core.BandmapEntry{
		Call:      core.MustParseCallsign("DL1ABC"),
		Frequency: frequency,
		Band:      core.Band20m,
		Mode:      core.ModeCW,
	}
}

func setupEntryWithVFO(t *testing.T) (*Scenario, *vfo.VFO, *rigStub) {
	t.Helper()
	s := NewScenario(t).WithClassicExchange()

	rig := new(rigStub)
	v := vfo.NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, func(f func()) { f() })
	rig.vfo = v
	v.SetClient(rig)
	v.Notify(s.controller)
	s.controller.SetVFO(core.VFO1, v)

	return s, v, rig
}

type rigStub struct {
	vfo       *vfo.VFO
	frequency core.Frequency
}

func (r *rigStub) reportFrequency(frequency core.Frequency) {
	r.frequency = frequency
	r.vfo.VFOFrequencyChanged(core.VFO1, frequency)
}

func (r *rigStub) Notify(any)               {}
func (r *rigStub) Active() bool             { return true }
func (r *rigStub) Refresh()                 { r.vfo.VFOFrequencyChanged(core.VFO1, r.frequency) }
func (r *rigStub) SetCurrentVFO(core.VFOID) {}
func (r *rigStub) SetTXVFO(core.VFOID)      {}
func (r *rigStub) SetFrequency(_ core.VFOID, frequency core.Frequency) {
	// a real rig needs time and answers with its own poll, see reportFrequency
}
func (r *rigStub) SetBand(core.VFOID, core.Band) {}
func (r *rigStub) SetMode(core.VFOID, core.Mode) {}
func (r *rigStub) IncrementalTuningPerVFO() bool { return false }
func (r *rigStub) SetIncrementalTuning(core.VFOID, core.IncrementalTuningKind, bool, core.Frequency) {
}
func (r *rigStub) MuteAudio(core.VFOID)   {}
func (r *rigStub) UnmuteAudio(core.VFOID) {}
func (r *rigStub) ToggleAudio(core.VFOID) {}
