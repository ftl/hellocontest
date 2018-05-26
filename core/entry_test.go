package core

import (
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEntryController_Reset(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	log.On("GetNextNumber").Once().Return(QSONumber(1))
	view.On("SetMyReport", "599").Once()
	view.On("SetMyNumber", "001").Once()
	view.On("SetCallsign", "").Once()
	view.On("SetTheirReport", "599").Once()
	view.On("SetTheirNumber", "").Once()
	view.On("SetActiveField", CallsignField).Once()
	view.On("SetDuplicateMarker", false).Once()
	view.On("ClearError").Once()

	controller.Reset()

	view.AssertExpectations(t)
}

func TestEntryController_GotoNextField(t *testing.T) {
	_, _, view, controller := setupEntryTest()

	view.On("GetCallsign").Return("").Maybe()
	view.On("SetActiveField", mock.Anything).Times(6)

	assert.Equal(t, CallsignField, controller.GetActiveField(), "callsign should be active at start")

	testCases := []struct {
		active, next EntryField
	}{
		{CallsignField, TheirReportField},
		{TheirReportField, TheirNumberField},
		{TheirNumberField, CallsignField},
		{MyReportField, CallsignField},
		{MyNumberField, CallsignField},
		{OtherField, CallsignField},
	}
	for _, tc := range testCases {
		controller.SetActiveField(tc.active)
		actual := controller.GotoNextField()
		assert.Equal(t, tc.next, actual)
		assert.Equal(t, tc.next, controller.GetActiveField())
	}

	view.AssertExpectations(t)
}

func TestEntryController_EnterNewCallsign(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	log.On("Find", mock.Anything).Return(QSO{}, false)
	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("SetDuplicateMarker", false).Once()
	view.On("SetActiveField", TheirReportField).Once()

	controller.GotoNextField()

	log.AssertExpectations(t)
	view.AssertExpectations(t)
}

func TestEntryController_EnterDuplicateCallsign(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := QSO{
		Callsign:    dl1abc,
		TheirReport: RST("599"),
		TheirNumber: 12,
		MyReport:    RST("559"),
		MyNumber:    1,
	}

	log.On("Find", dl1abc).Return(qso, true)
	view.On("SetDuplicateMarker", true)
	view.On("SetActiveField", TheirReportField)
	view.On("GetCallsign").Return("DL1ABC")
	view.On("SetTheirReport", "599")
	view.On("SetTheirNumber", "012")
	view.On("SetMyReport", "559")
	view.On("SetMyNumber", "001")

	controller.GotoNextField()

	log.AssertExpectations(t)
	view.AssertExpectations(t)
}

func TestEntryController_LogNewQSO(t *testing.T) {
	clock, log, view, controller := setupEntryTest()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := QSO{
		Callsign:    dl1abc,
		Time:        clock.Now(),
		TheirReport: RST("559"),
		TheirNumber: 12,
		MyReport:    RST("579"),
		MyNumber:    1,
	}

	log.On("Find", dl1abc).Once().Return(QSO{}, false)
	view.On("SetDuplicateMarker", false).Once()
	view.On("SetActiveField", TheirReportField).Once()
	view.On("GetCallsign").Once().Return("DL1ABC")
	controller.GotoNextField()

	log.On("Find", dl1abc).Once().Return(QSO{}, false)
	log.On("GetNextNumber").Once().Return(QSONumber(1))
	log.On("Log", qso).Once()
	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("GetTheirReport").Once().Return("559")
	view.On("GetTheirNumber").Once().Return("012")
	view.On("GetMyReport").Once().Return("579")
	view.On("GetMyNumber").Once().Return("001")
	view.On("SetMyReport", "599").Once()
	view.On("SetMyNumber", "001").Once()
	view.On("SetCallsign", "").Once()
	view.On("SetTheirReport", "599").Once()
	view.On("SetTheirNumber", "").Once()
	view.On("SetActiveField", CallsignField).Once()
	view.On("SetDuplicateMarker", false).Once()
	view.On("ClearError").Once()

	controller.Log()

	log.AssertExpectations(t)
	assert.Equal(t, CallsignField, controller.GetActiveField())
}

func TestEntryController_LogWithWrongCallsign(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	view.On("GetCallsign").Once().Return("DL")
	view.On("SetActiveField", CallsignField).Once()
	view.On("ShowError", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, CallsignField, controller.GetActiveField())
}

func TestEntryController_LogWithWrongTheirNumber(t *testing.T) {
	_, log, view, controller := setupEntryTest()

	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("GetTheirReport").Once().Return("559")
	view.On("GetTheirNumber").Once().Return("abc")
	view.On("SetActiveField", TheirNumberField).Once()
	view.On("ShowError", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, TheirNumberField, controller.GetActiveField())
}

func TestEntryController_LogDuplicateBeforeCheckForDuplicate(t *testing.T) {
	clock, log, view, controller := setupEntryTest()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := QSO{
		Callsign:    dl1abc,
		Time:        clock.Now(),
		TheirReport: RST("559"),
		TheirNumber: 12,
		MyReport:    RST("579"),
		MyNumber:    12,
	}

	log.On("Find", dl1abc).Once().Return(qso, true)
	view.On("GetCallsign").Once().Return("DL1ABC")
	view.On("GetTheirReport").Once().Return("599")
	view.On("GetTheirNumber").Once().Return("1")
	view.On("GetMyReport").Once().Return("579")
	view.On("GetMyNumber").Once().Return("013")
	view.On("SetActiveField", CallsignField).Once()
	view.On("ShowError", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)

}

// Helpers

func setupEntryTest() (Clock, *mockedLog, *mockedEntryView, EntryController) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := &staticClock{now}
	log := new(mockedLog)
	view := new(mockedEntryView)
	controller := NewEntryController(clock, log)

	log.On("GetNextNumber").Once().Return(QSONumber(1))
	view.On("SetEntryController", controller).Once()
	view.On("SetMyReport", "599").Once()
	view.On("SetMyNumber", "001").Once()
	view.On("SetCallsign", "").Once()
	view.On("SetTheirReport", "599").Once()
	view.On("SetTheirNumber", "").Once()
	view.On("SetActiveField", CallsignField).Once()
	view.On("SetDuplicateMarker", false).Once()
	view.On("ClearError").Once()
	controller.SetView(view)

	return clock, log, view, controller
}

// Mock

type mockedEntryView struct {
	mock.Mock
}

func (m *mockedEntryView) SetEntryController(controller EntryController) {
	m.Called(controller)
}

func (m *mockedEntryView) GetCallsign() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockedEntryView) SetCallsign(callsign string) {
	m.Called(callsign)
}

func (m *mockedEntryView) GetTheirReport() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockedEntryView) SetTheirReport(report string) {
	m.Called(report)
}

func (m *mockedEntryView) GetTheirNumber() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockedEntryView) SetTheirNumber(number string) {
	m.Called(number)
}

func (m *mockedEntryView) GetMyReport() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockedEntryView) SetMyReport(report string) {
	m.Called(report)
}

func (m *mockedEntryView) GetMyNumber() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockedEntryView) SetMyNumber(number string) {
	m.Called(number)
}

func (m *mockedEntryView) SetActiveField(field EntryField) {
	m.Called(field)
}

func (m *mockedEntryView) SetDuplicateMarker(active bool) {
	m.Called(active)
}

func (m *mockedEntryView) ShowError(err error) {
	m.Called(err)
}

func (m *mockedEntryView) ClearError() {
	m.Called()
}
