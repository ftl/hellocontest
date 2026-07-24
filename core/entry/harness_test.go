package entry_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ftl/conval"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/entry"
)

var scenarioFixedTime = time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)

// ---- Scenario ----------------------------------------------------------------

type Scenario struct {
	t           *testing.T
	controller  *entry.Controller
	view        *viewSpy
	logbook     *logbookSpy
	callinfo    *callinfoSpy
	listener    *listenerSpy
	vfo         *vfoSpy
	vfo2        *vfoSpy
	vfoSwitcher *vfoSwitcherSpy
	bandmap     *bandmapSpy
	esmView     *esmViewSpy
	keyer       *keyerSpy
	qsoList     *qsoListSpy
	seq         int // shared sequence counter for ordering assertions
}

func NewScenario(t *testing.T) *Scenario {
	t.Helper()
	s := &Scenario{
		t:        t,
		view:     &viewSpy{},
		logbook:  &logbookSpy{nextQSONumber: 1},
		callinfo: &callinfoSpy{},
		listener: &listenerSpy{},
		bandmap:  &bandmapSpy{},
		esmView:  &esmViewSpy{},
		keyer:    &keyerSpy{},
		qsoList:  &qsoListSpy{},
	}
	settings := &scenarioSettings{myCall: "DL0ABC"}
	s.controller = entry.NewController(
		settings,
		clock.Static(scenarioFixedTime),
		s.logbook,
		s.qsoList,
		s.bandmap,
		func(f func()) { f() },
	)
	s.vfo = &vfoSpy{ctrl: s.controller}
	s.controller.SetVFO(core.VFO1, s.vfo)
	s.controller.SetCallinfo(s.callinfo)
	s.controller.Notify(s.listener)
	s.controller.SetView(s.view) // triggers Clear() → vfo.Refresh()
	s.controller.SetESMEnabled(false)
	s.resetSpies()
	return s
}

func (s *Scenario) resetSpies() {
	s.seq = 0
	s.view.reset()
	s.logbook.resetCalls()
	s.callinfo.resetCalls()
	s.listener.resetCalls()
	s.vfo.reset()
	if s.vfo2 != nil {
		s.vfo2.reset()
	}
	if s.vfoSwitcher != nil {
		s.vfoSwitcher.reset()
	}
	s.bandmap.reset()
	s.esmView.reset()
	s.keyer.reset()
	s.qsoList.resetCalls()
}

// ---- Setup methods (no spy reset) -------------------------------------------

// WithClassicExchange configures RST + serial + generic-text exchange fields,
// with auto-generated serial numbers.
func (s *Scenario) WithClassicExchange() *Scenario {
	contest := core.Contest{
		Definition: scenarioFieldDefinition(
			conval.ExchangeField{conval.RSTProperty},
			conval.ExchangeField{conval.SerialNumberProperty},
			conval.ExchangeField{conval.GenericTextProperty},
		),
		GenerateSerialExchange: true,
		ExchangeValues:         []string{"599", "", ""},
	}
	contest.UpdateExchangeFields()
	s.controller.ContestChanged(contest)
	return s
}

// WithDuplicateQSO makes the logbook spy return qso for any FindDuplicateQSOs call.
func (s *Scenario) WithDuplicateQSO(qso core.QSO) *Scenario {
	s.logbook.duplicates = append(s.logbook.duplicates, qso)
	return s
}

// WithCallinfoFrame injects a prediction frame for vfo via CallinfoFrameChanged.
func (s *Scenario) WithCallinfoFrame(vfo core.VFOID, frame core.CallinfoFrame) *Scenario {
	s.controller.CallinfoFrameChanged(vfo, frame)
	return s
}

// WithClassicExchangeWithOptionalText configures RST + serial + optional-generic-text
// (EmptyAllowed) exchange fields. Needed for C8 prediction-fill path.
func (s *Scenario) WithClassicExchangeWithOptionalText() *Scenario {
	contest := core.Contest{
		Definition: scenarioFieldDefinition(
			conval.ExchangeField{conval.RSTProperty},
			conval.ExchangeField{conval.SerialNumberProperty},
			conval.ExchangeField{conval.GenericTextProperty, conval.EmptyProperty},
		),
		GenerateSerialExchange: true,
		ExchangeValues:         []string{"599", "", ""},
	}
	contest.UpdateExchangeFields()
	s.controller.ContestChanged(contest)
	return s
}

// WithESMEnabled enables ESM mode.
func (s *Scenario) WithESMEnabled() *Scenario {
	s.controller.SetESMEnabled(true)
	return s
}

// WithWorkmode sets the workmode without spy reset.
func (s *Scenario) WithWorkmode(w core.Workmode) *Scenario {
	s.controller.WorkmodeChanged(core.VFO1, w)
	return s
}

// WithKeyer wires the keyer spy into the controller without spy reset.
func (s *Scenario) WithKeyer() *Scenario {
	s.keyer.seq = &s.seq
	s.controller.SetKeyer(s.keyer)
	return s
}

// WithQSOListCallback sets a callback fired when SelectLastQSO is called. Not an action (no spy reset).
func (s *Scenario) WithQSOListCallback(f func()) *Scenario {
	s.qsoList.onSelectLast = f
	return s
}

// WithClassicExchangeAndReport configures RST + serial + generic-text exchange fields
// with both auto-generated serial numbers and report regeneration on mode change.
func (s *Scenario) WithClassicExchangeAndReport() *Scenario {
	contest := core.Contest{
		Definition: scenarioFieldDefinition(
			conval.ExchangeField{conval.RSTProperty},
			conval.ExchangeField{conval.SerialNumberProperty},
			conval.ExchangeField{conval.GenericTextProperty},
		),
		GenerateSerialExchange: true,
		GenerateReport:         true,
		ExchangeValues:         []string{"599", "", ""},
	}
	contest.UpdateExchangeFields()
	s.controller.ContestChanged(contest)
	return s
}

// ---- Action methods (reset spies before executing) --------------------------

// Enter sets the focused VFO's active field content.
func (s *Scenario) Enter(text string) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.Enter(text)
	return s
}

// GotoNextField advances the active field per the transition map.
func (s *Scenario) GotoNextField() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.GotoNextField()
	return s
}

// SelectQSO triggers QSOSelected, entering edit mode with qso.
func (s *Scenario) SelectQSO(qso core.QSO) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.QSOSelected(qso)
	return s
}

// SetActiveField sets the active field on the focused VFO.
func (s *Scenario) SetActiveField(field core.EntryField) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.SetActiveField(field)
	return s
}

// GotoNextPlaceholder moves focus to the callsign field and selects placeholder text.
func (s *Scenario) GotoNextPlaceholder() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.GotoNextPlaceholder()
	return s
}

// Log invokes the Log use case directly, bypassing EnterPressed dispatch.
func (s *Scenario) Log() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.Log()
	return s
}

// PressEnter dispatches EnterPressed on the controller.
func (s *Scenario) PressEnter() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.EnterPressed()
	return s
}

// SetXITActive toggles XIT active state on VFO1.
func (s *Scenario) SetXITActive(active bool) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.SetXITActive(active)
	return s
}

// Clear invokes controller.Clear().
func (s *Scenario) Clear() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.Clear()
	return s
}

// VFOFrequencyChanged fires the rig frequency-change event.
func (s *Scenario) VFOFrequencyChanged(vfo core.VFOID, freq core.Frequency) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.VFOFrequencyChanged(vfo, freq)
	return s
}

// VFOBandChanged fires the rig band-change event.
func (s *Scenario) VFOBandChanged(vfo core.VFOID, band core.Band) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.VFOBandChanged(vfo, band)
	return s
}

// VFOModeChanged fires the rig mode-change event.
func (s *Scenario) VFOModeChanged(vfo core.VFOID, mode core.Mode) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.VFOModeChanged(vfo, mode)
	return s
}

// WorkmodeChanged updates workmode (spy reset → action).
func (s *Scenario) WorkmodeChanged(w core.Workmode) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.WorkmodeChanged(core.VFO1, w)
	return s
}

// EditLastQSO triggers controller.EditLastQSO.
func (s *Scenario) EditLastQSO() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.EditLastQSO()
	return s
}

// MarkInBandmap adds the current callsign as a manual spot.
func (s *Scenario) MarkInBandmap() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.MarkInBandmap()
	return s
}

// EntrySelected simulates a bandmap entry being selected.
func (s *Scenario) EntrySelected(entry core.BandmapEntry) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.EntrySelected(entry)
	return s
}

// SelectBestMatchOnFrequency picks the top supercheck match.
func (s *Scenario) SelectBestMatchOnFrequency() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.SelectBestMatchOnFrequency()
	return s
}

// SelectMatch picks the supercheck match at the given index.
func (s *Scenario) SelectMatch(index int) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.SelectMatch(index)
	return s
}

// RefreshPrediction re-notifies callinfo with the current callsign.
func (s *Scenario) RefreshPrediction() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.RefreshPrediction()
	return s
}

// NextESMStep executes the next ESM step.
func (s *Scenario) NextESMStep() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.NextESMStep()
	return s
}

// ConnectESMView wires the esmViewSpy into the controller (spy reset first).
func (s *Scenario) ConnectESMView() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.SetESMView(s.esmView)
	return s
}

// ToggleESM enables or disables ESM (spy reset → action).
func (s *Scenario) ToggleESM(enabled bool) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.SetESMEnabled(enabled)
	return s
}

// ---- Assertion methods (no spy reset) ---------------------------------------

// AssertCallsignEntered asserts the CallsignEntered listener was notified with callsign.
func (s *Scenario) AssertCallsignEntered(callsign string) *Scenario {
	s.t.Helper()
	assert.Contains(s.t, s.listener.callsignsEntered, callsign,
		"CallsignEntered listener not notified with %q", callsign)
	return s
}

// AssertCallinfoNotified asserts callinfo.InputChanged was called with callsign.
func (s *Scenario) AssertCallinfoNotified(callsign string) *Scenario {
	s.t.Helper()
	assert.Contains(s.t, s.callinfo.callsigns, callsign,
		"callinfo InputChanged not called with callsign %q", callsign)
	return s
}

// AssertCallinfoCleared asserts callinfo.InputChanged(vfo, "") was called for the given VFO.
func (s *Scenario) AssertCallinfoCleared(vfo core.VFOID) *Scenario {
	s.t.Helper()
	found := false
	for _, inp := range s.callinfo.inputs {
		if inp.vfo == vfo && inp.call == "" {
			found = true
			break
		}
	}
	assert.True(s.t, found,
		"expected callinfo.InputChanged(%v, \"\") to be called (clear callinfo for VFO)", vfo)
	return s
}

// AssertMessageCleared asserts view.ClearMessage(vfo) was called and ShowMessage was not.
func (s *Scenario) AssertMessageCleared(vfo core.VFOID) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "ClearMessage", vfo)
	assert.False(s.t, s.view.wasCalled("ShowMessage"),
		"expected ShowMessage not to be called")
	return s
}

// AssertMessageShown asserts view.ShowMessage was called for vfo.
func (s *Scenario) AssertMessageShown(vfo core.VFOID) *Scenario {
	s.t.Helper()
	found := false
	for _, c := range s.view.calls {
		if c.method == "ShowMessage" && len(c.args) > 0 && c.args[0] == vfo {
			found = true
			break
		}
	}
	assert.True(s.t, found, "expected ShowMessage(%v, ...) to be called", vfo)
	return s
}

// AssertSerialClaimed asserts a non-zero serial is assigned to the focused VFO.
// Note: in black-box context this checks CurrentValues().MyNumber != 0, which is
// true for both a held claim and an unclaimed preview when NextQSONumber > 0.
// True claim stickiness is covered by serial-claim use case tests (I1/I2).
func (s *Scenario) AssertSerialClaimed() *Scenario {
	s.t.Helper()
	assert.NotZero(s.t, s.controller.CurrentValues().MyNumber,
		"expected a serial to be assigned (MyNumber != 0)")
	return s
}

// AssertActiveField asserts view.SetActiveField(vfo, field) was called.
func (s *Scenario) AssertActiveField(vfo core.VFOID, field core.EntryField) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetActiveField", vfo, field)
	return s
}

// AssertTextSelected asserts view.SelectText(vfo, field, text) was called.
func (s *Scenario) AssertTextSelected(vfo core.VFOID, field core.EntryField, text string) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SelectText", vfo, field, text)
	return s
}

// AssertMyExchangeValue asserts the my-exchange slot at index (1-based) holds value.
// Reads via CurrentValues().MyExchanges since Enter on my-exchange fields has no direct view call.
func (s *Scenario) AssertMyExchangeValue(index int, value string) *Scenario {
	s.t.Helper()
	vals := s.controller.CurrentValues()
	if index < 1 || index > len(vals.MyExchanges) {
		s.t.Errorf("my-exchange index %d out of range (len=%d)", index, len(vals.MyExchanges))
		return s
	}
	assert.Equal(s.t, value, vals.MyExchanges[index-1],
		"expected MyExchanges[%d] = %q", index-1, value)
	return s
}

// AssertDuplicateMarker asserts view.SetDuplicateMarker(vfo, marked) was called.
func (s *Scenario) AssertDuplicateMarker(vfo core.VFOID, marked bool) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetDuplicateMarker", vfo, marked)
	return s
}

// AssertTheirExchangeSet asserts view.SetTheirExchange(vfo, index, value) was called.
// index is 1-based, matching the view interface.
func (s *Scenario) AssertTheirExchangeSet(vfo core.VFOID, index int, value string) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetTheirExchange", vfo, index, value)
	return s
}

// AssertQSOAdded asserts at least one QSO was added to the logbook.
func (s *Scenario) AssertQSOAdded() *Scenario {
	s.t.Helper()
	assert.NotEmpty(s.t, s.logbook.addedQSOs, "expected at least one QSO to be added to the logbook")
	return s
}

// AssertQSOAddedCallsign asserts the first added QSO has the given callsign string.
func (s *Scenario) AssertQSOAddedCallsign(callsign string) *Scenario {
	s.t.Helper()
	if assert.NotEmpty(s.t, s.logbook.addedQSOs, "expected at least one QSO to be added") {
		assert.Equal(s.t, callsign, s.logbook.addedQSOs[0].Callsign.String(),
			"expected added QSO callsign to be %q", callsign)
	}
	return s
}

// AssertCallsignLogged asserts the CallsignLogged listener was notified with the given callsign.
func (s *Scenario) AssertCallsignLogged(callsign string) *Scenario {
	s.t.Helper()
	assert.Contains(s.t, s.listener.callsignsLogged, callsign,
		"CallsignLogged listener not notified with %q", callsign)
	return s
}

// AssertVFOBand asserts the VFO rig was commanded with the given band.
func (s *Scenario) AssertVFOBand(band core.Band) *Scenario {
	s.t.Helper()
	assert.Equal(s.t, band, s.vfo.lastBand,
		"expected VFO to be commanded with band %v", band)
	return s
}

// AssertVFOMode asserts the VFO rig was commanded with the given mode.
func (s *Scenario) AssertVFOMode(mode core.Mode) *Scenario {
	s.t.Helper()
	assert.Equal(s.t, mode, s.vfo.lastMode,
		"expected VFO to be commanded with mode %v", mode)
	return s
}

// AssertVFOFrequency asserts the VFO rig was commanded with the given frequency.
func (s *Scenario) AssertVFOFrequency(freq core.Frequency) *Scenario {
	s.t.Helper()
	assert.Equal(s.t, freq, s.vfo.lastFreq,
		"expected VFO to be commanded with frequency %v", freq)
	return s
}

// AssertXITActiveCommanded asserts the VFO rig was commanded with the given XIT-active flag.
func (s *Scenario) AssertXITActiveCommanded(active bool) *Scenario {
	s.t.Helper()
	assert.Equal(s.t, active, s.vfo.xitActive,
		"expected VFO XIT-active to be commanded as %v", active)
	return s
}

// AssertCallsignView asserts view.SetCallsign(vfo, text) was called.
func (s *Scenario) AssertCallsignView(vfo core.VFOID, text string) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetCallsign", vfo, text)
	return s
}

// AssertMyExchangeView asserts view.SetMyExchange(index, value) was called (index is 1-based).
func (s *Scenario) AssertMyExchangeView(index int, value string) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetMyExchange", index, value)
	return s
}

// AssertBandmapSelected asserts bandmap.SelectByCallsign was called with the given callsign string.
func (s *Scenario) AssertBandmapSelected(callsign string) *Scenario {
	s.t.Helper()
	found := false
	for _, c := range s.bandmap.selectedCallsigns {
		if c.String() == callsign {
			found = true
			break
		}
	}
	assert.True(s.t, found, "expected bandmap.SelectByCallsign to be called with %q", callsign)
	return s
}

// AssertEditingMarker asserts view.SetEditingMarker(vfo, marked) was called.
func (s *Scenario) AssertEditingMarker(vfo core.VFOID, marked bool) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetEditingMarker", vfo, marked)
	return s
}

// AssertQSOUpdated asserts at least one QSO was updated in the logbook.
func (s *Scenario) AssertQSOUpdated() *Scenario {
	s.t.Helper()
	assert.NotEmpty(s.t, s.logbook.updatedQSOs, "expected at least one QSO to be updated in logbook")
	return s
}

// AssertQSOUpdatedCallsign asserts the first updated QSO has the given callsign string.
func (s *Scenario) AssertQSOUpdatedCallsign(callsign string) *Scenario {
	s.t.Helper()
	if assert.NotEmpty(s.t, s.logbook.updatedQSOs, "expected at least one QSO to be updated") {
		assert.Equal(s.t, callsign, s.logbook.updatedQSOs[0].Callsign.String(),
			"expected updated QSO callsign to be %q", callsign)
	}
	return s
}

// AssertNoQSOAdded asserts no QSO was added to the logbook.
func (s *Scenario) AssertNoQSOAdded() *Scenario {
	s.t.Helper()
	assert.Empty(s.t, s.logbook.addedQSOs, "expected no QSO added to logbook")
	return s
}

// AssertBandmapAddedCallsign asserts bandmap.Add was called with the given callsign.
func (s *Scenario) AssertBandmapAddedCallsign(callsign string) *Scenario {
	s.t.Helper()
	found := false
	for _, spot := range s.bandmap.addedSpots {
		if spot.Call.String() == callsign {
			found = true
			break
		}
	}
	assert.True(s.t, found, "expected bandmap.Add with callsign %q", callsign)
	return s
}

// AssertBandmapSpotSource asserts the first added bandmap spot has the given source.
func (s *Scenario) AssertBandmapSpotSource(source core.SpotType) *Scenario {
	s.t.Helper()
	if assert.NotEmpty(s.t, s.bandmap.addedSpots, "expected at least one bandmap spot") {
		assert.Equal(s.t, source, s.bandmap.addedSpots[0].Source,
			"expected bandmap spot source %v", source)
	}
	return s
}

// AssertKeyerSentMacro asserts keyer.SendMacro(index) was called with the given index.
func (s *Scenario) AssertKeyerSentMacro(index int) *Scenario {
	s.t.Helper()
	assert.Contains(s.t, s.keyer.sentIndices, index,
		"expected keyer.SendMacro(%d) to be called", index)
	return s
}

// AssertKeyerSentText asserts keyer.SendText was called at least once.
func (s *Scenario) AssertKeyerSentText() *Scenario {
	s.t.Helper()
	assert.NotEmpty(s.t, s.keyer.sentTexts, "expected keyer.SendText to be called")
	return s
}

// AssertNoKeyerText asserts keyer.SendText was NOT called.
func (s *Scenario) AssertNoKeyerText() *Scenario {
	s.t.Helper()
	assert.Empty(s.t, s.keyer.sentTexts, "expected keyer.SendText NOT to be called")
	return s
}

// AssertESMViewEnabled asserts the esmView received SetESMEnabled(enabled).
func (s *Scenario) AssertESMViewEnabled(enabled bool) *Scenario {
	s.t.Helper()
	assert.Equal(s.t, enabled, s.esmView.esmEnabled,
		"expected esmView.esmEnabled = %v", enabled)
	return s
}

// AssertESMViewMessage asserts the esmView received SetMessage(msg).
func (s *Scenario) AssertESMViewMessage(msg string) *Scenario {
	s.t.Helper()
	assert.Equal(s.t, msg, s.esmView.message, "expected esmView.message = %q", msg)
	return s
}

// AssertESMListenerNotified asserts ESMEnabled(enabled) was emitted to listeners.
func (s *Scenario) AssertESMListenerNotified(enabled bool) *Scenario {
	s.t.Helper()
	assert.Contains(s.t, s.listener.esmEnabledValues, enabled,
		"expected ESMEnabled(%v) to be notified", enabled)
	return s
}

// AssertViewNotCalledWith asserts view.method(args...) was NOT called.
func (s *Scenario) AssertViewNotCalledWith(method string, args ...any) *Scenario {
	s.t.Helper()
	assert.False(s.t, s.view.wasCalledWith(method, args...),
		"expected view.%s(%v) NOT to be called", method, args)
	return s
}

// AssertQSOListSelectLastCalled asserts qsoList.SelectLastQSO was called.
func (s *Scenario) AssertQSOListSelectLastCalled() *Scenario {
	s.t.Helper()
	assert.Positive(s.t, s.qsoList.selectLastCalled,
		"expected qsoList.SelectLastQSO to be called")
	return s
}

// AssertBandView asserts view.SetBand(vfo, text) was called.
func (s *Scenario) AssertBandView(vfo core.VFOID, text string) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetBand", vfo, text)
	return s
}

// AssertModeView asserts view.SetMode(vfo, text) was called.
func (s *Scenario) AssertModeView(vfo core.VFOID, text string) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetMode", vfo, text)
	return s
}

// AssertFrequencyView asserts view.SetFrequency(vfo, freq) was called.
func (s *Scenario) AssertFrequencyView(vfo core.VFOID, freq core.Frequency) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetFrequency", vfo, freq)
	return s
}

// AssertNoLogbookWrite asserts no QSO was added or updated in the logbook.
func (s *Scenario) AssertNoLogbookWrite() *Scenario {
	s.t.Helper()
	assert.Empty(s.t, s.logbook.addedQSOs, "expected no QSO added to logbook")
	assert.Empty(s.t, s.logbook.updatedQSOs, "expected no QSO updated in logbook")
	return s
}

// ---- scenarioSettings -------------------------------------------------------

type scenarioSettings struct {
	myCall string
}

func (s *scenarioSettings) Station() core.Station {
	return core.Station{Callsign: core.MustParseCallsign(s.myCall)}
}

func (s *scenarioSettings) Contest() core.Contest {
	return core.Contest{}
}

// ---- vfoSpy -----------------------------------------------------------------

type vfoSpy struct {
	ctrl      *entry.Controller
	vfoID     core.VFOID
	lastBand  core.Band
	lastMode  core.Mode
	lastFreq  core.Frequency
	xitActive bool
}

func (v *vfoSpy) reset() {
	v.lastBand = core.NoBand
	v.lastMode = core.NoMode
	v.lastFreq = 0
	v.xitActive = false
}

func (v *vfoSpy) Name() string { return "STUB" }
func (v *vfoSpy) Notify(any)   {}
func (v *vfoSpy) Refresh() {
	v.ctrl.VFOFrequencyChanged(v.vfoID, 14050000)
	v.ctrl.VFOBandChanged(v.vfoID, core.Band20m)
	v.ctrl.VFOModeChanged(v.vfoID, core.ModeCW)
}
func (v *vfoSpy) SetFrequency(f core.Frequency) { v.lastFreq = f }
func (v *vfoSpy) SetBand(b core.Band)           { v.lastBand = b }
func (v *vfoSpy) SetMode(m core.Mode)           { v.lastMode = m }
func (v *vfoSpy) SetXIT(bool, core.Frequency)   {}
func (v *vfoSpy) XITActive() bool               { return v.xitActive }
func (v *vfoSpy) SetXITActive(active bool)      { v.xitActive = active }

// ---- qsoListSpy -------------------------------------------------------------

type qsoListSpy struct {
	selectLastCalled int
	onSelectLast     func() // optional: fired synchronously during SelectLastQSO
}

func (q *qsoListSpy) resetCalls() { q.selectLastCalled = 0 }

func (q *qsoListSpy) SelectLastQSO() {
	q.selectLastCalled++
	if q.onSelectLast != nil {
		q.onSelectLast()
	}
}

// ---- bandmapSpy -------------------------------------------------------------

type bandmapSpy struct {
	selectedCallsigns []core.Callsign
	addedSpots        []core.Spot
}

func (b *bandmapSpy) reset() {
	b.selectedCallsigns = nil
	b.addedSpots = nil
}

func (b *bandmapSpy) Add(spot core.Spot) { b.addedSpots = append(b.addedSpots, spot) }
func (b *bandmapSpy) SelectByCallsign(call core.Callsign) {
	b.selectedCallsigns = append(b.selectedCallsigns, call)
}

// ---- viewSpy ----------------------------------------------------------------

type viewCall struct {
	method string
	args   []any
}

type viewSpy struct {
	calls []viewCall
}

func (v *viewSpy) reset() { v.calls = nil }

func (v *viewSpy) record(method string, args ...any) {
	v.calls = append(v.calls, viewCall{method, args})
}

func (v *viewSpy) wasCalledWith(method string, args ...any) bool {
	for _, c := range v.calls {
		if c.method == method && reflect.DeepEqual(c.args, args) {
			return true
		}
	}
	return false
}

func (v *viewSpy) wasCalled(method string) bool {
	for _, c := range v.calls {
		if c.method == method {
			return true
		}
	}
	return false
}

func (v *viewSpy) assertCalledWith(t *testing.T, method string, args ...any) {
	t.Helper()
	assert.True(t, v.wasCalledWith(method, args...),
		"expected %s(%v) to be called", method, args)
}

func (v *viewSpy) SetMyCall(s string)            { v.record("SetMyCall", s) }
func (v *viewSpy) SetMyExchange(i int, s string) { v.record("SetMyExchange", i, s) }
func (v *viewSpy) SetFrequency(vfo core.VFOID, f core.Frequency) {
	v.record("SetFrequency", vfo, f)
}
func (v *viewSpy) SetBand(vfo core.VFOID, text string) { v.record("SetBand", vfo, text) }
func (v *viewSpy) SetMode(vfo core.VFOID, text string) { v.record("SetMode", vfo, text) }
func (v *viewSpy) SetXITActive(vfo core.VFOID, active bool) {
	v.record("SetXITActive", vfo, active)
}
func (v *viewSpy) SetXIT(vfo core.VFOID, active bool, offset core.Frequency) {
	v.record("SetXIT", vfo, active, offset)
}
func (v *viewSpy) SetTXState(vfo core.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration) {
	v.record("SetTXState", vfo, ptt, parrotActive, parrotTimeLeft)
}
func (v *viewSpy) SetCallsign(vfo core.VFOID, s string) { v.record("SetCallsign", vfo, s) }
func (v *viewSpy) SetTheirExchange(vfo core.VFOID, index int, text string) {
	v.record("SetTheirExchange", vfo, index, text)
}
func (v *viewSpy) SetSerialClaim(vfo core.VFOID, n core.QSONumber, committed bool) {
	v.record("SetSerialClaim", vfo, n, committed)
}
func (v *viewSpy) SetActiveVFO(vfo core.VFOID) {
	v.record("SetActiveVFO", vfo)
}
func (v *viewSpy) SetActiveField(vfo core.VFOID, field core.EntryField) {
	v.record("SetActiveField", vfo, field)
}
func (v *viewSpy) SelectText(vfo core.VFOID, field core.EntryField, s string) {
	v.record("SelectText", vfo, field, s)
}
func (v *viewSpy) SetDuplicateMarker(vfo core.VFOID, b bool) {
	v.record("SetDuplicateMarker", vfo, b)
}
func (v *viewSpy) SetEditingMarker(vfo core.VFOID, b bool) {
	v.record("SetEditingMarker", vfo, b)
}
func (v *viewSpy) ShowMessage(vfo core.VFOID, args ...any) {
	v.record("ShowMessage", append([]any{vfo}, args...)...)
}
func (v *viewSpy) ClearMessage(vfo core.VFOID)          { v.record("ClearMessage", vfo) }
func (v *viewSpy) SetVFOEnabled(vfo core.VFOID, b bool) { v.record("SetVFOEnabled", vfo, b) }
func (v *viewSpy) SetVFOWorkmode(vfo core.VFOID, w core.Workmode) {
	v.record("SetVFOWorkmode", vfo, w)
}

func (v *viewSpy) SetTXVFO(vfo core.VFOID) {
	v.record("SetTXVFO", vfo)
}

// ---- logbookSpy -------------------------------------------------------------

type logbookSpy struct {
	nextQSONumber core.QSONumber
	lastBand      core.Band
	lastMode      core.Mode
	lastExchange  []string
	duplicates    []core.QSO
	addedQSOs     []core.QSO
	updatedQSOs   []core.QSO
}

func (l *logbookSpy) resetCalls() {
	l.addedQSOs = nil
	l.updatedQSOs = nil
}

func (l *logbookSpy) NextQSONumber() core.QSONumber { return l.nextQSONumber }
func (l *logbookSpy) LastBand() core.Band           { return l.lastBand }
func (l *logbookSpy) LastMode() core.Mode           { return l.lastMode }
func (l *logbookSpy) LastExchange() []string        { return l.lastExchange }
func (l *logbookSpy) AddQSO(qso core.QSO)           { l.addedQSOs = append(l.addedQSOs, qso) }
func (l *logbookSpy) UpdateQSO(qso core.QSO)        { l.updatedQSOs = append(l.updatedQSOs, qso) }
func (l *logbookSpy) FindDuplicateQSOs(_ core.Callsign, _ core.Band, _ core.Mode) []core.QSO {
	return l.duplicates
}

// ---- callinfoSpy ------------------------------------------------------------

type callinfoInput struct {
	vfo  core.VFOID
	call string
}

type callinfoSpy struct {
	callsigns []string
	inputs    []callinfoInput
}

func (c *callinfoSpy) resetCalls() {
	c.callsigns = nil
	c.inputs = nil
}

func (c *callinfoSpy) InputChanged(vfo core.VFOID, call string, _ core.Band, _ core.Mode, _ []string) {
	c.callsigns = append(c.callsigns, call)
	c.inputs = append(c.inputs, callinfoInput{vfo: vfo, call: call})
}

// ---- listenerSpy ------------------------------------------------------------

type listenerSpy struct {
	callsignsEntered []string
	callsignsLogged  []string
	esmEnabledValues []bool
}

func (l *listenerSpy) resetCalls() {
	l.callsignsEntered = nil
	l.callsignsLogged = nil
	l.esmEnabledValues = nil
}

func (l *listenerSpy) CallsignEntered(_ core.VFOID, callsign string) {
	l.callsignsEntered = append(l.callsignsEntered, callsign)
}

func (l *listenerSpy) CallsignLogged(callsign string, _ core.Frequency) {
	l.callsignsLogged = append(l.callsignsLogged, callsign)
}

func (l *listenerSpy) ESMEnabled(enabled bool) {
	l.esmEnabledValues = append(l.esmEnabledValues, enabled)
}

// ---- esmViewSpy -------------------------------------------------------------

type esmViewSpy struct {
	esmEnabled bool
	message    string
}

func (e *esmViewSpy) reset()                { e.esmEnabled = false; e.message = "" }
func (e *esmViewSpy) SetESMEnabled(en bool) { e.esmEnabled = en }
func (e *esmViewSpy) SetMessage(msg string) { e.message = msg }

// ---- keyerSpy ---------------------------------------------------------------

type keyerSpy struct {
	seq          *int // shared sequence counter (nil until WithKeyer+WithVFOSwitcher)
	sentIndices  []int
	sentTexts    []string
	sentQuestion string
	repeated     bool
	stopped      bool
	txSeq        int // sequence number of first TX call in this spy cycle
}

func (k *keyerSpy) reset() {
	k.sentIndices = nil
	k.sentTexts = nil
	k.sentQuestion = ""
	k.repeated = false
	k.stopped = false
	k.txSeq = 0
}

func (k *keyerSpy) nextSeq() int {
	if k.seq == nil {
		return 0
	}
	*k.seq++
	return *k.seq
}

func (k *keyerSpy) SendMacro(index int) {
	k.sentIndices = append(k.sentIndices, index)
	if k.txSeq == 0 {
		k.txSeq = k.nextSeq()
	}
}

func (k *keyerSpy) SendQuestion(q string) {
	k.sentQuestion = q
	if k.txSeq == 0 {
		k.txSeq = k.nextSeq()
	}
}

func (k *keyerSpy) GetText(_ core.Workmode, index int) (string, error) {
	return fmt.Sprintf("keyer[%d]", index), nil
}

func (k *keyerSpy) SendText(text string, _ ...any) {
	k.sentTexts = append(k.sentTexts, text)
	if k.txSeq == 0 {
		k.txSeq = k.nextSeq()
	}
}

func (k *keyerSpy) Repeat() {
	k.repeated = true
	if k.txSeq == 0 {
		k.txSeq = k.nextSeq()
	}
}
func (k *keyerSpy) Stop() { k.stopped = true }

// ---- helpers ----------------------------------------------------------------

func scenarioFieldDefinition(fields ...conval.ExchangeField) *conval.Definition {
	return &conval.Definition{
		Exchange: []conval.ExchangeDefinition{
			{Fields: fields},
		},
	}
}

// ---- vfoSwitcherSpy ---------------------------------------------------------

type vfoSwitcherSpy struct {
	seq               *int // shared sequence counter
	calledCurrentWith []core.VFOID
	calledTXWith      []core.VFOID
	lastSeq           int // sequence number of last SetTXVFO call
}

func (v *vfoSwitcherSpy) reset() {
	v.calledCurrentWith = nil
	v.calledTXWith = nil
	v.lastSeq = 0
}

func (v *vfoSwitcherSpy) SetCurrentVFO(vfo core.VFOID) {
	v.calledCurrentWith = append(v.calledCurrentWith, vfo)
	if v.seq != nil {
		*v.seq++
		v.lastSeq = *v.seq
	}
}

func (v *vfoSwitcherSpy) SetTXVFO(vfo core.VFOID) {
	v.calledTXWith = append(v.calledTXWith, vfo)
	if v.seq != nil {
		*v.seq++
		v.lastSeq = *v.seq
	}
}

// ---- Setup methods (H–N additions) -----------------------------------------

// WithVFO2 wires the vfo2 spy as VFO2 and enables dual-VFO mode.
func (s *Scenario) WithVFO2() *Scenario {
	s.vfo2 = &vfoSpy{ctrl: s.controller, vfoID: core.VFO2}
	s.controller.SetVFO(core.VFO2, s.vfo2)
	s.controller.RadioChanged("", false) // singleVFO=false → vfo2Enabled=true
	return s
}

// WithVFOSwitcher wires the vfoSwitcherSpy into the controller.
func (s *Scenario) WithVFOSwitcher() *Scenario {
	s.vfoSwitcher = &vfoSwitcherSpy{seq: &s.seq}
	s.controller.SetVFOSwitcher(s.vfoSwitcher)
	return s
}

// WithLastBand sets the logbook spy's last band (used by LogbookLoaded).
func (s *Scenario) WithLastBand(band core.Band) *Scenario {
	s.logbook.lastBand = band
	return s
}

// WithLastMode sets the logbook spy's last mode (used by LogbookLoaded).
func (s *Scenario) WithLastMode(mode core.Mode) *Scenario {
	s.logbook.lastMode = mode
	return s
}

// WithLastExchange sets the logbook spy's last exchange (seeded into my-exchange on Clear).
func (s *Scenario) WithLastExchange(exchange []string) *Scenario {
	s.logbook.lastExchange = exchange
	return s
}

// WithNextQSONumber sets the logbook spy's next QSO number.
func (s *Scenario) WithNextQSONumber(n core.QSONumber) *Scenario {
	s.logbook.nextQSONumber = n
	return s
}

// ---- Action methods (H–N additions) ----------------------------------------

// SetFocusedVFO switches the focused VFO.
func (s *Scenario) SetFocusedVFO(vfo core.VFOID) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.SetFocusedVFO(vfo)
	return s
}

// ToggleFocusedVFO flips the focused VFO between VFO1 and VFO2.
func (s *Scenario) ToggleFocusedVFO() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.ToggleFocusedVFO()
	return s
}

// FocusVFO1 sets focus to VFO1.
func (s *Scenario) FocusVFO1() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.FocusVFO1()
	return s
}

// FocusVFO2 sets focus to VFO2.
func (s *Scenario) FocusVFO2() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.FocusVFO2()
	return s
}

// LogVFO focuses vfo and logs the QSO.
func (s *Scenario) LogVFO(vfo core.VFOID) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.LogVFO(vfo)
	return s
}

// ClearVFO focuses vfo and clears the row.
func (s *Scenario) ClearVFO(vfo core.VFOID) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.ClearVFO(vfo)
	return s
}

// RadioChanged fires the radio-changed event.
func (s *Scenario) RadioChanged(singleVFO bool) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.RadioChanged("", singleVFO)
	return s
}

// CurrentVFOChanged fires the rig-side VFO change event.
func (s *Scenario) CurrentVFOChanged(vfo core.VFOID) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.CurrentVFOChanged(vfo)
	return s
}

// VFOXITChanged fires the XIT-changed event.
func (s *Scenario) VFOXITChanged(vfo core.VFOID, active bool, offset core.Frequency) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.VFOXITChanged(vfo, active, offset)
	return s
}

// XITActiveChanged fires the XIT-active event.
func (s *Scenario) XITActiveChanged(active bool) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.XITActiveChanged(active)
	return s
}

// VFOPTTChanged fires the PTT-changed event.
func (s *Scenario) VFOPTTChanged(vfo core.VFOID, active bool) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.VFOPTTChanged(vfo, active)
	return s
}

// SendQuestion dispatches SendQuestion on the controller.
func (s *Scenario) SendQuestion() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.SendQuestion()
	return s
}

// RepeatLastTransmission dispatches RepeatLastTransmission.
func (s *Scenario) RepeatLastTransmission() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.RepeatLastTransmission()
	return s
}

// StopTX dispatches StopTX.
func (s *Scenario) StopTX() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.StopTX()
	return s
}

// ParrotActive fires the parrot-active event.
func (s *Scenario) ParrotActive(active bool) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.ParrotActive(active)
	return s
}

// ParrotTimeLeft fires the parrot-time-left event.
func (s *Scenario) ParrotTimeLeft(d time.Duration) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.ParrotTimeLeft(d)
	return s
}

// SerialSent fires the serial-sent event.
func (s *Scenario) SerialSent() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.SerialSent()
	return s
}

// StationChanged fires the station-changed event with the given callsign.
func (s *Scenario) StationChanged(callsign string) *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.StationChanged(core.Station{Callsign: core.MustParseCallsign(callsign)})
	return s
}

// LogbookLoaded fires the logbook-loaded event.
func (s *Scenario) LogbookLoaded() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.LogbookLoaded()
	return s
}

// RefreshView pushes current input to the view.
func (s *Scenario) RefreshView() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.RefreshView()
	return s
}

// Activate reapplies the active field to the view.
func (s *Scenario) Activate() *Scenario {
	s.t.Helper()
	s.resetSpies()
	s.controller.Activate()
	return s
}

// ---- Assertion methods (H–N additions) -------------------------------------

// AssertVFOSwitcherTXCalled asserts the vfoSwitcher was commanded with the given VFO.
func (s *Scenario) AssertVFOSwitcherTXCalled(vfo core.VFOID) *Scenario {
	s.t.Helper()
	if s.vfoSwitcher == nil {
		s.t.Error("vfoSwitcher spy not wired (call WithVFOSwitcher() in setup)")
		return s
	}
	found := false
	for _, v := range s.vfoSwitcher.calledTXWith {
		if v == vfo {
			found = true
			break
		}
	}
	assert.True(s.t, found, "expected vfoSwitcher.SetTXVFO(%v) to be called", vfo)
	return s
}

// AssertTXVFOBeforeKeyer asserts SetTXVFO(vfo) was called before any keyer TX method.
func (s *Scenario) AssertTXVFOBeforeKeyer(vfo core.VFOID) *Scenario {
	s.t.Helper()
	if s.vfoSwitcher == nil {
		s.t.Error("vfoSwitcher spy not wired (call WithVFOSwitcher() in setup)")
		return s
	}
	s.AssertVFOSwitcherTXCalled(vfo)
	assert.NotZero(s.t, s.vfoSwitcher.lastSeq, "SetTXVFO was not called")
	assert.NotZero(s.t, s.keyer.txSeq, "no keyer TX method was called")
	assert.Less(s.t, s.vfoSwitcher.lastSeq, s.keyer.txSeq,
		"SetTXVFO must be called before keyer TX (vfoSwitch seq=%d, keyer seq=%d)",
		s.vfoSwitcher.lastSeq, s.keyer.txSeq)
	return s
}

// AssertVFO2Enabled asserts view.SetVFOEnabled(VFO2, enabled) was called.
func (s *Scenario) AssertVFO2Enabled(enabled bool) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetVFOEnabled", core.VFO2, enabled)
	return s
}

// AssertXITView asserts view.SetXIT(vfo, active, offset) was called.
func (s *Scenario) AssertXITView(vfo core.VFOID, active bool, offset core.Frequency) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetXIT", vfo, active, offset)
	return s
}

// AssertXITActiveView asserts view.SetXITActive(vfo, active) was called.
func (s *Scenario) AssertXITActiveView(vfo core.VFOID, active bool) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetXITActive", vfo, active)
	return s
}

// AssertTXStateView asserts view.SetTXState(vfo, ptt, parrotActive, parrotTimeLeft) was called.
func (s *Scenario) AssertTXStateView(vfo core.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetTXState", vfo, ptt, parrotActive, parrotTimeLeft)
	return s
}

// AssertMyCallView asserts view.SetMyCall(call) was called.
func (s *Scenario) AssertMyCallView(call string) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetMyCall", call)
	return s
}

// AssertActiveVFO asserts view.SetActiveVFO(vfo) was called.
func (s *Scenario) AssertActiveVFO(vfo core.VFOID) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetActiveVFO", vfo)
	return s
}

// AssertVFOSwitcherCurrentCalled asserts vfoSwitcher.SetCurrentVFO(vfo) was called.
func (s *Scenario) AssertVFOSwitcherCurrentCalled(vfo core.VFOID) *Scenario {
	s.t.Helper()
	if s.vfoSwitcher == nil {
		s.t.Error("vfoSwitcher spy not wired (call WithVFOSwitcher() in setup)")
		return s
	}
	found := false
	for _, v := range s.vfoSwitcher.calledCurrentWith {
		if v == vfo {
			found = true
			break
		}
	}
	assert.True(s.t, found, "expected vfoSwitcher.SetCurrentVFO(%v) to be called", vfo)
	return s
}

// AssertSerialClaimView asserts view.SetSerialClaim(vfo, serial, committed) was called.
func (s *Scenario) AssertSerialClaimView(vfo core.VFOID, serial core.QSONumber, committed bool) *Scenario {
	s.t.Helper()
	s.view.assertCalledWith(s.t, "SetSerialClaim", vfo, serial, committed)
	return s
}

// AssertSerialCommitted asserts the serial claim for vfo is committed.
func (s *Scenario) AssertSerialCommitted(vfo core.VFOID) *Scenario {
	s.t.Helper()
	assert.True(s.t, s.controller.IsSerialCommitted(vfo),
		"expected serial on VFO %d to be committed", vfo)
	return s
}

// AssertSerialNotCommitted asserts the serial claim for vfo is not committed.
func (s *Scenario) AssertSerialNotCommitted(vfo core.VFOID) *Scenario {
	s.t.Helper()
	assert.False(s.t, s.controller.IsSerialCommitted(vfo),
		"expected serial on VFO %d to NOT be committed", vfo)
	return s
}

// AssertKeyerSentQuestion asserts keyer.SendQuestion(q) was called.
func (s *Scenario) AssertKeyerSentQuestion(q string) *Scenario {
	s.t.Helper()
	assert.Equal(s.t, q, s.keyer.sentQuestion,
		"expected keyer.SendQuestion(%q)", q)
	return s
}

// AssertKeyerRepeated asserts keyer.Repeat() was called.
func (s *Scenario) AssertKeyerRepeated() *Scenario {
	s.t.Helper()
	assert.True(s.t, s.keyer.repeated, "expected keyer.Repeat() to be called")
	return s
}

// AssertKeyerStopped asserts keyer.Stop() was called.
func (s *Scenario) AssertKeyerStopped() *Scenario {
	s.t.Helper()
	assert.True(s.t, s.keyer.stopped, "expected keyer.Stop() to be called")
	return s
}
