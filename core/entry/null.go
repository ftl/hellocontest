package entry

import (
	"time"

	"github.com/ftl/hellocontest/core/model"
)

// Null implementations of the interfaces to avoid nil checks.

type nullView struct{}

func (n *nullView) SetFrequency(model.VFOID, model.Frequency) {}
func (n *nullView) SetCallsign(model.VFOID, string)           {}
func (n *nullView) SetBand(vfo model.VFOID, text string)      {}
func (n *nullView) SetMode(vfo model.VFOID, text string)      {}
func (n *nullView) SetTXState(vfo model.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration) {
}
func (n *nullView) SetMyExchange(int, string)                         {}
func (n *nullView) SetTheirExchange(model.VFOID, int, string)         {}
func (n *nullView) SetMyExchangeFields([]model.ExchangeField)         {}
func (n *nullView) SetTheirExchangeFields([]model.ExchangeField)      {}
func (n *nullView) SetSerialClaim(model.VFOID, model.QSONumber, bool) {}
func (n *nullView) SetActiveVFO(model.VFOID)                          {}
func (n *nullView) SetActiveField(model.VFOID, model.EntryField)      {}
func (n *nullView) SelectText(model.VFOID, model.EntryField, string)  {}
func (n *nullView) SetDuplicateMarker(model.VFOID, bool)              {}
func (n *nullView) SetEditingMarker(model.VFOID, bool)                {}
func (n *nullView) ShowMessage(model.VFOID, ...any)                   {}
func (n *nullView) ClearMessage(model.VFOID)                          {}
func (n *nullView) SetVFOEnabled(model.VFOID, bool)                   {}
func (n *nullView) SetVFOWorkmode(model.VFOID, model.Workmode)        {}
func (n *nullView) SetTXVFO(model.VFOID)                              {}

type nullVFO struct{}

func (n *nullVFO) Name() string                                                            { return "" }
func (n *nullVFO) Notify(any)                                                              {}
func (n *nullVFO) Active() bool                                                            { return false }
func (n *nullVFO) Refresh()                                                                {}
func (n *nullVFO) SetFrequency(model.Frequency)                                            {}
func (n *nullVFO) ShiftFrequency(model.Frequency)                                          {}
func (n *nullVFO) SetBand(model.Band)                                                      {}
func (n *nullVFO) SetMode(model.Mode)                                                      {}
func (n *nullVFO) IncrementalTuningActive(model.IncrementalTuningKind) bool                { return false }
func (n *nullVFO) SetIncrementalTuningActive(model.IncrementalTuningKind, bool)            {}
func (n *nullVFO) SetIncrementalTuning(model.IncrementalTuningKind, bool, model.Frequency) {}
func (n *nullVFO) ShiftIncrementalTuning(model.IncrementalTuningKind, model.Frequency)     {}
func (n *nullVFO) ToggleAvailableIncrementalTuning()                                       {}
func (n *nullVFO) ShiftAvailableIncrementalTuning(model.Frequency)                         {}

type nullLogbook struct{}

func (n *nullLogbook) NextNumber() model.QSONumber { return 0 }
func (n *nullLogbook) LastBand() model.Band        { return model.NoBand }
func (n *nullLogbook) LastMode() model.Mode        { return model.NoMode }
func (n *nullLogbook) LastExchange() []string      { return nil }
func (n *nullLogbook) AddQSO(model.QSO)            {}
func (n *nullLogbook) UpdateQSO(model.QSO)         {}

type nullCallinfo struct{}

func (n *nullCallinfo) InputChanged(model.VFOID, string, model.Band, model.Mode, []string) {}

type nullBandmap struct{}

func (n *nullBandmap) Add(model.Spot)                                                     {}
func (n *nullBandmap) SelectByCallsign(model.Callsign)                                    {}
func (n *nullBandmap) GotoCQMarker()                                                      {}
func (n *nullBandmap) GotoNumberedMarker(int)                                             {}
func (n *nullBandmap) MarkWithNextNumber(model.Frequency, model.Band, model.MarkerInput)  {}
func (n *nullBandmap) MarkWithNumber(int, model.Frequency, model.Band, model.MarkerInput) {}
func (n *nullBandmap) MarkWithText(string, model.Frequency, model.Band)                   {}

type nullESMView struct{}

func (n *nullESMView) SetESMEnabled(enabled bool) {}
func (n *nullESMView) SetMessage(message string)  {}
