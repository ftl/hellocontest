package entry

import (
	"time"

	"github.com/ftl/hellocontest/core"
)

// Null implementations of the interfaces to avoid nil checks.

type nullView struct{}

func (n *nullView) SetMyCall(string)                                          {}
func (n *nullView) SetFrequency(core.VFOID, core.Frequency)                   {}
func (n *nullView) SetCallsign(core.VFOID, string)                            {}
func (n *nullView) SetBand(vfo core.VFOID, text string)                       {}
func (n *nullView) SetMode(vfo core.VFOID, text string)                       {}
func (n *nullView) SetTXState(vfo core.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration) {
}
func (n *nullView) SetMyExchange(int, string)                       {}
func (n *nullView) SetTheirExchange(core.VFOID, int, string)        {}
func (n *nullView) SetMyExchangeFields([]core.ExchangeField)        {}
func (n *nullView) SetTheirExchangeFields([]core.ExchangeField)     {}
func (n *nullView) SetSerialClaim(core.VFOID, core.QSONumber, bool) {}
func (n *nullView) SetActiveVFO(core.VFOID)                         {}
func (n *nullView) SetActiveField(core.VFOID, core.EntryField)      {}
func (n *nullView) SelectText(core.VFOID, core.EntryField, string)  {}
func (n *nullView) SetDuplicateMarker(core.VFOID, bool)             {}
func (n *nullView) SetEditingMarker(core.VFOID, bool)               {}
func (n *nullView) ShowMessage(core.VFOID, ...any)                  {}
func (n *nullView) ClearMessage(core.VFOID)                         {}
func (n *nullView) SetVFOEnabled(core.VFOID, bool)                  {}
func (n *nullView) SetVFOWorkmode(core.VFOID, core.Workmode)        {}
func (n *nullView) SetTXVFO(core.VFOID)                             {}

type nullVFO struct{}

func (n *nullVFO) Name() string                  { return "" }
func (n *nullVFO) Notify(any)                    {}
func (n *nullVFO) Active() bool                  { return false }
func (n *nullVFO) Refresh()                      {}
func (n *nullVFO) SetFrequency(core.Frequency)   {}
func (n *nullVFO) ShiftFrequency(core.Frequency) {}
func (n *nullVFO) SetBand(core.Band)             {}
func (n *nullVFO) SetMode(core.Mode)             {}
func (n *nullVFO) SetIncrementalTuning(core.IncrementalTuningKind, bool, core.Frequency) {}
func (n *nullVFO) ShiftOffset(core.IncrementalTuningKind, core.Frequency)                {}
func (n *nullVFO) ToggleIncrementalTuning()                                              {}
func (n *nullVFO) ShiftAvailableIncrementalTuning(core.Frequency)                        {}
func (n *nullVFO) IncrementalTuningActive(core.IncrementalTuningKind) bool               { return false }
func (n *nullVFO) SetIncrementalTuningActive(core.IncrementalTuningKind, bool)           {}

type nullLogbook struct{}

func (n *nullLogbook) NextNumber() core.QSONumber { return 0 }
func (n *nullLogbook) LastBand() core.Band        { return core.NoBand }
func (n *nullLogbook) LastMode() core.Mode        { return core.NoMode }
func (n *nullLogbook) LastExchange() []string     { return nil }
func (n *nullLogbook) AddQSO(core.QSO)            {}
func (n *nullLogbook) UpdateQSO(core.QSO)         {}

type nullCallinfo struct{}

func (n *nullCallinfo) InputChanged(core.VFOID, string, core.Band, core.Mode, []string) {}

type nullBandmap struct{}

func (n *nullBandmap) Add(core.Spot)                  {}
func (n *nullBandmap) SelectByCallsign(core.Callsign) {}

type nullESMView struct{}

func (n *nullESMView) SetESMEnabled(enabled bool) {}
func (n *nullESMView) SetMessage(message string)  {}
