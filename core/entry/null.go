package entry

import (
	"time"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

// Null implementations of the interfaces to avoid nil checks.

type nullView struct{}

func (n *nullView) SetUTC(string)                                                        {}
func (n *nullView) SetMyCall(string)                                                     {}
func (n *nullView) SetFrequency(core.Frequency)                                          {}
func (n *nullView) SetCallsign(string)                                                   {}
func (n *nullView) SetBand(text string)                                                  {}
func (n *nullView) SetMode(text string)                                                  {}
func (n *nullView) SetXITActive(active bool)                                             {}
func (n *nullView) SetXIT(active bool, offset core.Frequency)                            {}
func (n *nullView) SetTXState(ptt bool, parrotActive bool, parrotTimeLeft time.Duration) {}
func (n *nullView) SetMyExchange(int, string)                                            {}
func (n *nullView) SetTheirExchange(int, string)                                         {}
func (n *nullView) SetMyExchangeFields([]core.ExchangeField)                             {}
func (n *nullView) SetTheirExchangeFields([]core.ExchangeField)                          {}
func (n *nullView) SetActiveField(core.EntryField)                                       {}
func (n *nullView) SelectText(core.EntryField, string)                                   {}
func (n *nullView) SetDuplicateMarker(bool)                                              {}
func (n *nullView) SetEditingMarker(bool)                                                {}
func (n *nullView) ShowMessage(...any)                                                   {}
func (n *nullView) ClearMessage()                                                        {}

type nullVFO struct{}

func (n *nullVFO) Notify(any)                  {}
func (n *nullVFO) Active() bool                { return false }
func (n *nullVFO) Refresh()                    {}
func (n *nullVFO) SetFrequency(core.Frequency) {}
func (n *nullVFO) SetBand(core.Band)           {}
func (n *nullVFO) SetMode(core.Mode)           {}
func (n *nullVFO) SetXIT(bool, core.Frequency) {}
func (n *nullVFO) XITActive() bool             { return false }
func (n *nullVFO) SetXITActive(bool)           {}

type nullLogbook struct{}

func (n *nullLogbook) NextNumber() core.QSONumber { return 0 }
func (n *nullLogbook) LastBand() core.Band        { return core.NoBand }
func (n *nullLogbook) LastMode() core.Mode        { return core.NoMode }
func (n *nullLogbook) LastExchange() []string     { return nil }
func (n *nullLogbook) LogQSO(core.QSO)               {}

type nullCallinfo struct{}

func (n *nullCallinfo) InputChanged(string, core.Band, core.Mode, []string) {}

type nullBandmap struct{}

func (n *nullBandmap) Add(core.Spot)                      {}
func (n *nullBandmap) SelectByCallsign(callsign.Callsign) {}

type nullESMView struct{}

func (n *nullESMView) SetESMEnabled(enabled bool) {}
func (n *nullESMView) SetMessage(message string)  {}
