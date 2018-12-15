package entry

import (
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/mocked"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEntryController_Reset(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	log.On("NextNumber").Once().Return(core.QSONumber(1))
	view.On("SetMyReport", "599").Once()
	view.On("SetMyNumber", "001").Once()
	view.On("SetCallsign", "").Once()
	view.On("SetTheirReport", "599").Once()
	view.On("SetTheirNumber", "").Once()
	view.On("SetTheirXchange", "").Once()
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("SetDuplicateMarker", false).Once()
	view.On("ClearMessage").Once()

	controller.Reset()

	view.AssertExpectations(t)
}

func TestEntryController_SetLastSelectedBandAndModeOnReset(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	log.On("NextNumber").Once().Return(core.QSONumber(1))
	view.On("SetBand", "30m").Once()
	view.On("SetMode", "RTTY").Once()
	view.On("SetMyReport", "599").Twice()
	view.On("SetTheirReport", "599").Twice()
	view.On("SetMyNumber", "001").Once()
	view.On("SetCallsign", "").Once()
	view.On("SetTheirNumber", "").Once()
	view.On("SetTheirXchange", "").Once()
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("SetDuplicateMarker", false).Once()
	view.On("ClearMessage").Once()

	controller.BandSelected("30m")
	controller.ModeSelected("RTTY")
	controller.Reset()

	view.AssertExpectations(t)
}

func TestEntryController_GotoNextField(t *testing.T) {
	_, _, view, controller := setupEntryTest()

	view.On("GetCallsign").Return("").Maybe()
	view.On("SetActiveField", mock.Anything).Times(10)

	assert.Equal(t, core.CallsignField, controller.GetActiveField(), "callsign should be active at start")

	testCases := []struct {
		enterTheirNumber, enterTheirXchange bool
		active, next                        core.EntryField
	}{
		{true, true, core.CallsignField, core.TheirReportField},
		{true, true, core.TheirReportField, core.TheirNumberField},
		{false, true, core.TheirReportField, core.TheirXchangeField},
		{false, false, core.TheirReportField, core.CallsignField},
		{true, true, core.TheirNumberField, core.TheirXchangeField},
		{true, false, core.TheirNumberField, core.CallsignField},
		{true, true, core.TheirXchangeField, core.CallsignField},
		{true, true, core.MyReportField, core.CallsignField},
		{true, true, core.MyNumberField, core.CallsignField},
		{true, true, core.OtherField, core.CallsignField},
	}
	for _, tc := range testCases {
		controller.enterTheirNumber = tc.enterTheirNumber
		controller.enterTheirXchange = tc.enterTheirXchange
		controller.SetActiveField(tc.active)
		actual := controller.GotoNextField()
		assert.Equal(t, tc.next, actual)
		assert.Equal(t, tc.next, controller.GetActiveField())
	}

	view.AssertExpectations(t)
}

func TestEntryController_EnterNewCallsign(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	log.On("Find", mock.Anything).Return(core.QSO{}, false)
	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("SetDuplicateMarker", false).Once()
	view.On("SetActiveField", core.TheirReportField).Once()

	controller.GotoNextField()

	log.AssertExpectations(t)
	view.AssertExpectations(t)
}

func TestEntryController_EnterDuplicateCallsign(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := core.QSO{
		Callsign:    dl1abc,
		Band:        core.Band20m,
		Mode:        core.ModeSSB,
		TheirReport: core.RST("599"),
		TheirNumber: 12,
		MyReport:    core.RST("559"),
		MyNumber:    1,
	}

	log.On("Find", dl1abc).Return(qso, true).Once()
	view.On("SetDuplicateMarker", true).Once()
	view.On("SetActiveField", core.TheirReportField).Once()
	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("SetBand", "20m").Once()
	view.On("SetMode", "SSB").Once()
	view.On("SetTheirReport", "599").Once()
	view.On("SetTheirNumber", "012").Once()
	view.On("SetTheirXchange", "").Once()
	view.On("SetMyReport", "559").Once()
	view.On("SetMyNumber", "001").Once()
	view.On("SetMyXchange", "").Once()

	controller.GotoNextField()

	log.AssertExpectations(t)
	view.AssertExpectations(t)
}

func TestEntryController_LogNewQSO(t *testing.T) {
	clock, log, view, controller := setupEntryTest()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := core.QSO{
		Callsign:     dl1abc,
		Time:         clock.Now(),
		Band:         core.Band40m,
		Mode:         core.ModeCW,
		TheirReport:  core.RST("559"),
		TheirNumber:  12,
		TheirXchange: "thx",
		MyReport:     core.RST("579"),
		MyNumber:     1,
		MyXchange:    "myx",
	}

	log.On("Find", dl1abc).Once().Return(core.QSO{}, false)
	view.On("SetDuplicateMarker", false).Once()
	view.On("SetActiveField", core.TheirReportField).Once()
	view.On("GetCallsign").Once().Return("DL1ABC")
	controller.GotoNextField()

	log.On("Find", dl1abc).Once().Return(core.QSO{}, false)
	log.On("NextNumber").Once().Return(core.QSONumber(1))
	log.On("Log", qso).Once()
	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("GetBand").Once().Return("40m")
	view.On("GetMode").Once().Return("CW")
	view.On("GetTheirReport").Once().Return("559")
	view.On("GetTheirNumber").Once().Return("012")
	view.On("GetTheirXchange").Once().Return("thx")
	view.On("GetMyReport").Once().Return("579")
	view.On("GetMyNumber").Once().Return("001")
	view.On("GetMyXchange").Once().Return("myx")
	view.On("SetMyReport", "599").Once()
	view.On("SetMyNumber", "001").Once()
	view.On("SetCallsign", "").Once()
	view.On("SetTheirReport", "599").Once()
	view.On("SetTheirNumber", "").Once()
	view.On("SetTheirXchange", "").Once()
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("SetDuplicateMarker", false).Once()
	view.On("ClearMessage").Once()

	controller.Log()

	log.AssertExpectations(t)
	assert.Equal(t, core.CallsignField, controller.GetActiveField())
}

func TestEntryController_LogWithWrongCallsign(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	view.On("GetCallsign").Once().Return("DL")
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, core.CallsignField, controller.GetActiveField())
}

func TestEntryController_LogWithInvalidTheirReport(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("GetBand").Once().Return("40m")
	view.On("GetMode").Once().Return("CW")
	view.On("GetTheirReport").Once().Return("000")
	view.On("SetActiveField", core.TheirReportField).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, core.TheirReportField, controller.GetActiveField())
}

func TestEntryController_LogWithWrongTheirNumber(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("GetBand").Once().Return("40m")
	view.On("GetMode").Once().Return("CW")
	view.On("GetTheirReport").Once().Return("559")
	view.On("GetTheirNumber").Once().Return("abc")
	view.On("SetActiveField", core.TheirNumberField).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, core.TheirNumberField, controller.GetActiveField())
}

func TestEntryController_LogWithInvalidMyReport(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("GetBand").Once().Return("40m")
	view.On("GetMode").Once().Return("CW")
	view.On("GetTheirReport").Once().Return("599")
	view.On("GetTheirNumber").Once().Return("1")
	view.On("GetTheirXchange").Once().Return("abc")
	view.On("GetMyReport").Once().Return("000")
	view.On("SetActiveField", core.MyReportField).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, core.MyReportField, controller.GetActiveField())
}

func TestEntryController_LogDuplicateBeforeCheckForDuplicate(t *testing.T) {
	clock, log, view, controller := setupEntryTest()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := core.QSO{
		Callsign:     dl1abc,
		Time:         clock.Now(),
		TheirReport:  core.RST("559"),
		TheirNumber:  12,
		TheirXchange: "abc",
		MyReport:     core.RST("579"),
		MyNumber:     12,
		MyXchange:    "def",
	}

	log.On("Find", dl1abc).Once().Return(qso, true)
	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("GetBand").Once().Return("40m")
	view.On("GetMode").Once().Return("CW")
	view.On("GetTheirReport").Once().Return("599")
	view.On("GetTheirNumber").Once().Return("1")
	view.On("GetTheirXchange").Once().Return("abc")
	view.On("GetMyReport").Once().Return("579")
	view.On("GetMyNumber").Once().Return("013")
	view.On("GetMyXchange").Once().Return("def")
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
}

func TestEntryController_EnterCallsignCheckForDuplicateAndShowMessage(t *testing.T) {
	clock, log, view, controller := setupEntryTest()

	dl1ab, _ := callsign.Parse("DL1AB")
	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := core.QSO{
		Callsign:    dl1ab,
		Time:        clock.Now(),
		TheirReport: core.RST("559"),
		TheirNumber: 12,
		MyReport:    core.RST("579"),
		MyNumber:    12,
	}

	log.On("Find", dl1ab).Once().Return(qso, true)
	view.On("ShowMessage", mock.Anything).Once()
	controller.EnterCallsign("DL1AB")
	view.AssertExpectations(t)

	log.On("Find", dl1abc).Once().Return(qso, false)
	view.On("ClearMessage").Once()
	controller.EnterCallsign("DL1ABC")
	view.AssertExpectations(t)
}

// Helpers

func setupEntryTest() (core.Clock, *mocked.Log, *mocked.EntryView, *controller) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	log := new(mocked.Log)
	view := new(mocked.EntryView)
	controller := NewController(clock, log, true, true).(*controller)

	log.On("NextNumber").Once().Return(core.QSONumber(1))
	view.On("SetEntryController", controller).Once()
	view.On("EnableExchangeFields", true, true).Once()
	view.On("SetCallsign", "").Once()
	view.On("SetMyReport", "599").Once()
	view.On("SetTheirReport", "599").Once()
	view.On("SetMyNumber", "001").Once()
	view.On("SetTheirNumber", "").Once()
	view.On("SetTheirXchange", "").Once()
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("SetDuplicateMarker", false).Once()
	view.On("ClearMessage").Once()
	controller.SetView(view)

	return clock, log, view, controller
}
