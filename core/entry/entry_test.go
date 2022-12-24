package entry

import (
	"fmt"
	"testing"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/mocked"
)

func TestEntryController_Clear(t *testing.T) {
	_, log, qsoList, _, controller, _ := setupEntryTestWithClassicExchangeFields()
	log.Activate()
	log.On("NextNumber").Return(core.QSONumber(1)).Once()
	qsoList.Activate()
	qsoList.On("SelectLastQSO").Once()

	controller.Clear()

	assert.Equal(t, controller.input.myReport, "599", "my report")
	assert.Equal(t, controller.input.myNumber, "001", "my number")
	assert.Equal(t, controller.input.myExchange, []string{"599", "001", ""}, "my exchange")
	assert.Equal(t, controller.input.callsign, "", "callsign")
	assert.Equal(t, controller.input.theirReport, "599", "their report")
	assert.Equal(t, controller.input.theirNumber, "", "their number")
	assert.Equal(t, controller.input.theirExchange, []string{"599", "", ""}, "their exchange")
	assert.Equal(t, controller.input.band, "160m", "band")
	assert.Equal(t, controller.input.mode, "CW", "mode")
}

func TestEntryController_ClearView(t *testing.T) {
	_, log, qsoList, view, controller, _ := setupEntryTestWithClassicExchangeFields()
	log.Activate()
	log.On("NextNumber").Once().Return(core.QSONumber(1))
	qsoList.Activate()
	qsoList.On("SelectLastQSO").Once()

	view.Activate()
	view.On("SetTheirExchange", 1, "599").Once()
	view.On("SetTheirExchange", 2, "").Once()
	view.On("SetTheirExchange", 3, "").Once()
	view.On("SetMyExchange", 1, "599").Once()
	view.On("SetMyExchange", 2, "001").Once()
	view.On("SetMyExchange", 3, "").Once()
	view.On("SetMyCall", "DL0ABC").Once()
	view.On("SetFrequency", mock.Anything).Once()
	view.On("SetCallsign", "").Once()
	view.On("SetBand", "160m").Once()
	view.On("SetMode", "CW").Once()
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("SetDuplicateMarker", false).Once()
	view.On("SetEditingMarker", false).Once()
	view.On("ClearMessage").Once()

	controller.Clear()

	view.AssertExpectations(t)
}

func TestEntryController_SetLastSelectedBandAndModeOnClear(t *testing.T) {
	_, log, qsoList, _, controller, _ := setupEntryTest()
	log.Activate()
	log.On("NextNumber").Once().Return(core.QSONumber(1))
	qsoList.Activate()
	qsoList.On("SelectLastQSO").Once()

	controller.SetActiveField(core.BandField)
	controller.Enter("30m")
	controller.SetActiveField(core.ModeField)
	controller.Enter("RTTY")
	controller.Clear()

	assert.Equal(t, "30m", controller.input.band)
	assert.Equal(t, core.Band30m, controller.selectedBand)
	assert.Equal(t, "RTTY", controller.input.mode)
	assert.Equal(t, core.ModeRTTY, controller.selectedMode)
}

func TestEntryController_UpdateExchangeFields(t *testing.T) {
	tt := []struct {
		desc                   string
		value                  *conval.Definition
		generateSerialExchange bool
		expectedMyFields       []core.ExchangeField
		expectedTheirFields    []core.ExchangeField
	}{
		{
			desc:                "no definition",
			value:               nil,
			expectedMyFields:    nil,
			expectedTheirFields: nil,
		},
		{
			desc: "rst and member number",
			value: fieldDefinition(
				conval.ExchangeField{conval.RSTProperty},
				conval.ExchangeField{conval.MemberNumberProperty, conval.NoMemberProperty},
			),
			expectedMyFields: []core.ExchangeField{
				{Field: "myExchange_1", Short: "rst", Properties: conval.ExchangeField{conval.RSTProperty}},
				{Field: "myExchange_2", Short: "member_number/nm", Properties: conval.ExchangeField{conval.MemberNumberProperty, conval.NoMemberProperty}},
			},
			expectedTheirFields: []core.ExchangeField{
				{Field: "theirExchange_1", Short: "rst", Properties: conval.ExchangeField{conval.RSTProperty}},
				{Field: "theirExchange_2", Short: "member_number/nm", Properties: conval.ExchangeField{conval.MemberNumberProperty, conval.NoMemberProperty}},
			},
		},
		{
			desc: "rst and dok or serial number",
			value: fieldDefinition(
				conval.ExchangeField{conval.RSTProperty},
				conval.ExchangeField{conval.SerialNumberProperty, conval.NoMemberProperty, conval.WAGDOKProperty},
			),
			expectedMyFields: []core.ExchangeField{
				{Field: "myExchange_1", Short: "rst", Properties: conval.ExchangeField{conval.RSTProperty}},
				{Field: "myExchange_2", Short: "serial/nm/wag_dok", Properties: conval.ExchangeField{conval.SerialNumberProperty, conval.NoMemberProperty, conval.WAGDOKProperty}, CanContainSerial: true},
			},
			expectedTheirFields: []core.ExchangeField{
				{Field: "theirExchange_1", Short: "rst", Properties: conval.ExchangeField{conval.RSTProperty}},
				{Field: "theirExchange_2", Short: "serial/nm/wag_dok", Properties: conval.ExchangeField{conval.SerialNumberProperty, conval.NoMemberProperty, conval.WAGDOKProperty}, CanContainSerial: true},
			},
		},
		{
			desc: "rst and serial number",
			value: fieldDefinition(
				conval.ExchangeField{conval.RSTProperty},
				conval.ExchangeField{conval.SerialNumberProperty, conval.NoMemberProperty, conval.WAGDOKProperty},
			),
			generateSerialExchange: true,
			expectedMyFields: []core.ExchangeField{
				{Field: "myExchange_1", Short: "rst", Properties: conval.ExchangeField{conval.RSTProperty}},
				{Field: "myExchange_2", Short: "#", Hint: "Serial Number", Properties: conval.ExchangeField{conval.SerialNumberProperty, conval.NoMemberProperty, conval.WAGDOKProperty}, CanContainSerial: true, ReadOnly: true},
			},
			expectedTheirFields: []core.ExchangeField{
				{Field: "theirExchange_1", Short: "rst", Properties: conval.ExchangeField{conval.RSTProperty}},
				{Field: "theirExchange_2", Short: "serial/nm/wag_dok", Properties: conval.ExchangeField{conval.SerialNumberProperty, conval.NoMemberProperty, conval.WAGDOKProperty}, CanContainSerial: true},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			_, _, _, view, controller, _ := setupEntryTest()
			view.Activate()
			view.On("SetMyExchangeFields", tc.expectedMyFields).Once()
			view.On("SetTheirExchangeFields", tc.expectedTheirFields).Once()

			controller.updateExchangeFields(tc.value, tc.generateSerialExchange, make([]string, len(tc.expectedMyFields)))

			view.AssertExpectations(t)
		})
	}
}

func TestEntryController_GotoNextField(t *testing.T) {
	_, _, _, view, controller, config := setupEntryTestWithExchangeFields(3)

	assert.Equal(t, core.CallsignField, controller.activeField, "callsign should be active at start")

	testCases := []struct {
		active, next core.EntryField
	}{
		{core.CallsignField, core.TheirExchangeField(1)},
		{core.OtherField, core.CallsignField},
		{core.MyExchangeField(1), core.CallsignField},
		{core.MyExchangeField(2), core.CallsignField},
		{core.MyExchangeField(3), core.CallsignField},
		{core.TheirExchangeField(1), core.TheirExchangeField(2)},
		{core.TheirExchangeField(2), core.TheirExchangeField(3)},
		{core.TheirExchangeField(3), core.CallsignField},
	}
	view.Activate()
	view.On("Callsign").Return("").Maybe()
	view.On("SetActiveField", mock.Anything).Times(len(testCases))
	view.On("SetMyExchangeFields", mock.Anything).Times(len(testCases))
	view.On("SetTheirExchangeFields", mock.Anything).Times(len(testCases))
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s -> %s", tc.active, tc.next), func(t *testing.T) {
			controller.ContestChanged(config.Contest())
			controller.SetActiveField(tc.active)
			actual := controller.GotoNextField()
			assert.Equal(t, tc.next, actual)
			assert.Equal(t, tc.next, controller.activeField)
		})
	}

	view.AssertExpectations(t)
}

func TestEntryController_EnterNewCallsign(t *testing.T) {
	_, _, qsoList, view, controller, _ := setupEntryTestWithClassicExchangeFields()
	qsoList.Activate()
	qsoList.On("FindDuplicateQSOs", mock.Anything, mock.Anything, mock.Anything).Return([]core.QSO{})

	view.Activate()
	view.On("SetDuplicateMarker", false).Once()
	view.On("ClearMessage").Once()
	view.On("SetActiveField", core.TheirExchangeField(1)).Once()
	// view.On("SetTheirExchange", mock.Anything, mock.Anything).Once() // TODO implement the prediction with the new exchange fields

	controller.Enter("DL1ABC")
	controller.GotoNextField()

	assert.Equal(t, "DL1ABC", controller.input.callsign)

	qsoList.AssertExpectations(t)
	view.AssertExpectations(t)
}

func TestEntryController_EnterDuplicateCallsign(t *testing.T) {
	_, _, qsoList, view, controller, _ := setupEntryTestWithClassicExchangeFields()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		Band:          core.Band20m,
		Mode:          core.ModeSSB,
		TheirReport:   core.RST("599"),
		TheirNumber:   12,
		TheirExchange: []string{"599", "012", ""},
		MyReport:      core.RST("559"),
		MyNumber:      1,
		MyExchange:    []string{"599", "001", ""},
	}

	qsoList.Activate()
	qsoList.On("FindDuplicateQSOs", dl1abc, core.Band160m, core.ModeCW).Return([]core.QSO{qso}).Twice()

	view.Activate()
	view.On("SetDuplicateMarker", true).Once()
	view.On("ShowMessage", mock.Anything).Once()
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("SetActiveField", core.TheirExchangeField(1)).Once()
	// view.On("SetTheirExchange", mock.Anything, mock.Anything).Once() // TODO implement the prediction with the new exchange fields

	controller.Enter("DL1ABC")
	controller.GotoNextField()

	assert.False(t, controller.editing)
	qsoList.AssertExpectations(t)
	view.AssertExpectations(t)
}

func TestEntryController_EnterFrequency(t *testing.T) {
	_, _, _, view, controller, _ := setupEntryTest()

	view.Activate()
	view.On("SetCallsign", "").Once()
	view.On("SetFrequency", core.Frequency(7028000)).Once()

	controller.Enter("7028")
	controller.Log()

	assert.Equal(t, "", controller.input.callsign)

	view.AssertExpectations(t)
}

func TestEntryController_LogNewQSO(t *testing.T) {
	clock, log, qsoList, _, controller, _ := setupEntryTestWithClassicExchangeFields()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		Time:          clock.Now(),
		Band:          core.Band40m,
		Mode:          core.ModeCW,
		TheirReport:   core.RST("559"),
		TheirNumber:   12,
		TheirExchange: []string{"559", "012", "thx"},
		MyReport:      core.RST("579"),
		MyNumber:      1,
		MyExchange:    []string{"579", "001", "myx"},
	}

	log.Activate()
	log.On("NextNumber").Return(core.QSONumber(1))
	log.On("Log", qso).Once()
	qsoList.Activate()
	qsoList.On("FindDuplicateQSOs", dl1abc, mock.Anything, mock.Anything).Return([]core.QSO{})
	qsoList.On("SelectLastQSO").Twice()

	controller.Clear()
	controller.SetActiveField(core.BandField)
	controller.Enter("40m")
	controller.SetActiveField(core.ModeField)
	controller.Enter("CW")
	controller.SetActiveField(core.MyExchangeField(1))
	controller.Enter("579")
	controller.SetActiveField(core.MyExchangeField(3))
	controller.Enter("myx")
	controller.GotoNextField()

	controller.Enter("DL1ABC")
	controller.GotoNextField()
	controller.Enter("559")
	controller.GotoNextField()
	controller.Enter("012")
	controller.GotoNextField()
	controller.Enter("thx")

	controller.Log()

	log.AssertExpectations(t)
	qsoList.AssertExpectations(t)
	assert.Equal(t, core.CallsignField, controller.activeField)
}

func TestEntryController_LogWithWrongCallsign(t *testing.T) {
	_, log, _, view, controller, _ := setupEntryTest()
	log.Activate()

	view.Activate()
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Enter("DL")
	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, core.CallsignField, controller.activeField)
}

func TestEntryController_LogWithInvalidTheirReport(t *testing.T) {
	_, log, _, view, controller, _ := setupEntryTestWithClassicExchangeFields()

	controller.SetActiveField(core.BandField)
	controller.Enter("40m")
	controller.SetActiveField(core.ModeField)
	controller.Enter("CW")
	controller.GotoNextField()
	controller.Enter("DL1ABC")
	controller.GotoNextField()
	controller.Enter("000")

	log.Activate()
	view.Activate()
	view.On("SetActiveField", core.TheirExchangeField(1)).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, core.TheirExchangeField(1), controller.activeField)
}

func TestEntryController_LogWithWrongTheirNumber(t *testing.T) {
	_, log, _, view, controller, _ := setupEntryTestWithClassicExchangeFields()

	controller.SetActiveField(core.BandField)
	controller.Enter("40m")
	controller.SetActiveField(core.ModeField)
	controller.Enter("CW")
	controller.GotoNextField()
	controller.Enter("DL1ABC")
	controller.GotoNextField()
	controller.Enter("559")
	controller.GotoNextField()
	controller.Enter("abc")

	log.Activate()
	view.Activate()
	view.On("SetActiveField", core.TheirExchangeField(2)).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, core.TheirExchangeField(2), controller.activeField)
}

func TestEntryController_LogWithoutMandatoryTheirNumber(t *testing.T) {
	_, log, _, view, controller, _ := setupEntryTestWithClassicExchangeFields()

	controller.SetActiveField(core.BandField)
	controller.Enter("40m")
	controller.SetActiveField(core.ModeField)
	controller.Enter("CW")
	controller.GotoNextField()
	controller.Enter("DL1ABC")
	controller.GotoNextField()
	controller.Enter("559")

	log.Activate()
	view.Activate()
	view.On("SetActiveField", core.TheirExchangeField(2)).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, core.TheirExchangeField(2), controller.activeField)
}

func TestEntryController_LogWithInvalidMyReport(t *testing.T) {
	_, log, _, view, controller, _ := setupEntryTestWithClassicExchangeFields()

	controller.SetActiveField(core.BandField)
	controller.Enter("40m")
	controller.SetActiveField(core.ModeField)
	controller.Enter("CW")
	controller.SetActiveField(core.MyExchangeField(1))
	controller.Enter("000")
	controller.GotoNextField()
	controller.Enter("DL1ABC")
	controller.GotoNextField()
	controller.Enter("559")
	controller.GotoNextField()
	controller.Enter("1")
	controller.GotoNextField()
	controller.Enter("abc")

	log.Activate()
	view.Activate()
	view.On("SetActiveField", core.MyExchangeField(1)).Once()
	view.On("ShowMessage", mock.Anything).Once()

	controller.Log()

	view.AssertExpectations(t)
	log.AssertNotCalled(t, "Log", mock.Anything)
	assert.Equal(t, core.MyExchangeField(1), controller.activeField)
}

func TestEntryController_EnterCallsignCheckForDuplicateAndShowMessage(t *testing.T) {
	clock, _, qsoList, view, controller, _ := setupEntryTest()
	qsoList.Activate()
	view.Activate()

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

	qsoList.On("FindDuplicateQSOs", dl1ab, mock.Anything, mock.Anything).Once().Return([]core.QSO{qso})
	view.On("ShowMessage", mock.Anything).Once()
	view.On("SetActiveField", mock.Anything).Once()
	controller.Enter("DL1AB")
	view.AssertExpectations(t)

	qsoList.On("FindDuplicateQSOs", dl1abc, mock.Anything, mock.Anything).Once().Return([]core.QSO{})
	view.On("ClearMessage").Once()
	controller.Enter("DL1ABC")
	view.AssertExpectations(t)
}

func TestEntryController_LogDuplicateQSO(t *testing.T) {
	clock, log, qsoList, _, controller, _ := setupEntryTestWithClassicExchangeFields()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		Time:          clock.Now().Add(-1 * time.Minute),
		Band:          core.Band40m,
		Mode:          core.ModeCW,
		TheirReport:   core.RST("559"),
		TheirNumber:   12,
		TheirExchange: []string{"559", "012", "thx"},
		MyReport:      core.RST("579"),
		MyNumber:      1,
		MyExchange:    []string{"579", "001", "myx"},
	}
	dupe := core.QSO{
		Callsign:      dl1abc,
		Time:          clock.Now(),
		Band:          core.Band40m,
		Mode:          core.ModeCW,
		TheirReport:   core.RST("569"),
		TheirNumber:   12,
		TheirExchange: []string{"569", "012", "thx"},
		MyReport:      core.RST("579"),
		MyNumber:      2,
		MyExchange:    []string{"579", "002", "myx"},
	}

	log.Activate()
	log.On("NextNumber").Return(core.QSONumber(2))
	log.On("Log", dupe).Once()
	qsoList.Activate()
	qsoList.On("FindDuplicateQSOs", dl1abc, mock.Anything, mock.Anything).Return([]core.QSO{qso})
	qsoList.On("SelectLastQSO").Twice()

	controller.Clear()
	controller.SetActiveField(core.BandField)
	controller.Enter("40m")
	controller.SetActiveField(core.ModeField)
	controller.Enter("CW")
	controller.SetActiveField(core.MyExchangeField(1))
	controller.Enter("579")
	controller.SetActiveField(core.MyExchangeField(3))
	controller.Enter("myx")
	controller.GotoNextField()

	controller.Enter("DL1ABC")
	controller.GotoNextField()
	controller.Enter("569")
	controller.GotoNextField()
	controller.Enter("012")
	controller.GotoNextField()
	controller.Enter("thx")

	controller.Log()

	log.AssertExpectations(t)
	qsoList.AssertExpectations(t)
	assert.Equal(t, core.CallsignField, controller.activeField)
}

func TestEntryController_SelectRowForEditing(t *testing.T) {
	clock, _, _, view, controller, _ := setupEntryTestWithClassicExchangeFields()
	view.Activate()

	dl1abc, _ := callsign.Parse("DL1ABC")
	qso := core.QSO{
		Band:          core.Band80m,
		Mode:          core.ModeCW,
		Callsign:      dl1abc,
		Time:          clock.Now(),
		TheirReport:   core.RST("559"),
		TheirNumber:   12,
		TheirExchange: []string{"559", "012", "A01"},
		MyReport:      core.RST("579"),
		MyNumber:      34,
		MyExchange:    []string{"579", "034", "B36"},
	}

	view.On("SetBand", "80m").Once()
	view.On("SetMode", "CW").Once()
	view.On("SetCallsign", "DL1ABC").Once()
	view.On("SetTheirExchange", 1, "559").Once()
	view.On("SetTheirExchange", 2, "012").Once()
	view.On("SetTheirExchange", 3, "A01").Once()
	view.On("SetMyExchange", 1, "579").Once()
	view.On("SetMyExchange", 2, "034").Once()
	view.On("SetMyExchange", 3, "B36").Once()
	view.On("SetActiveField", core.CallsignField).Once()
	view.On("SetEditingMarker", true).Once()

	controller.QSOSelected(qso)

	assertQSOInput(t, qso, controller)

	view.AssertExpectations(t)
}

func TestEntryController_EditQSO(t *testing.T) {
	clock, log, _, _, controller, _ := setupEntryTestWithClassicExchangeFields()

	dl1abc, _ := callsign.Parse("DL1ABC")
	dl2abc, _ := callsign.Parse("DL2ABC")
	qso := core.QSO{
		Band:          core.Band80m,
		Mode:          core.ModeCW,
		Callsign:      dl1abc,
		Time:          clock.Now(),
		TheirReport:   core.RST("559"),
		TheirNumber:   12,
		TheirExchange: []string{"559", "012", "A01"},
		MyReport:      core.RST("579"),
		MyNumber:      34,
		MyExchange:    []string{"579", "034", "B36"},
	}
	changedQSO := qso
	changedQSO.Callsign = dl2abc
	changedQSO.TheirExchange = []string{"559", "012", "B02"}

	controller.QSOSelected(qso)
	controller.SetActiveField(core.CallsignField)
	controller.Enter("DL2ABC")
	controller.SetActiveField(core.TheirExchangeField(3))
	controller.Enter("B02")

	log.Activate()
	log.On("Log", changedQSO).Once()
	log.On("NextNumber").Return(core.QSONumber(35))
	controller.Log()

	log.AssertExpectations(t)
}

// Helpers

func setupEntryTest() (core.Clock, *mocked.Log, *mocked.QSOList, *mocked.EntryView, *Controller, *testSettings) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	log := new(mocked.Log)
	qsoList := new(mocked.QSOList)
	view := new(mocked.EntryView)
	settings := &testSettings{myCall: "DL0ABC", enterTheirNumber: true, enterTheirXchange: true}
	controller := NewController(settings, clock, qsoList, testIgnoreAsync)
	controller.SetLogbook(log)
	controller.SetView(view)

	return clock, log, qsoList, view, controller, settings
}

func setupEntryTestWithClassicExchangeFields() (core.Clock, *mocked.Log, *mocked.QSOList, *mocked.EntryView, *Controller, *testSettings) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	log := new(mocked.Log)
	qsoList := new(mocked.QSOList)
	view := new(mocked.EntryView)
	exchangeFields := []conval.ExchangeField{{conval.RSTProperty}, {conval.SerialNumberProperty}, {conval.GenericTextProperty}}
	settings := &testSettings{myCall: "DL0ABC", enterTheirNumber: true, enterTheirXchange: true, exchangeFields: exchangeFields, exchangeValues: []string{"599", "", ""}, generateSerialExchange: true}
	controller := NewController(settings, clock, qsoList, testIgnoreAsync)
	controller.SetLogbook(log)
	controller.SetView(view)

	return clock, log, qsoList, view, controller, settings
}

func setupEntryTestWithExchangeFields(exchangeFieldCount int) (core.Clock, *mocked.Log, *mocked.QSOList, *mocked.EntryView, *Controller, *testSettings) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	log := new(mocked.Log)
	qsoList := new(mocked.QSOList)
	view := new(mocked.EntryView)
	exchangeFields := make([]conval.ExchangeField, exchangeFieldCount)
	exchangeValues := make([]string, exchangeFieldCount)
	for i := range exchangeFields {
		exchangeFields[i] = conval.ExchangeField{conval.GenericTextProperty}
	}
	settings := &testSettings{myCall: "DL0ABC", enterTheirNumber: true, enterTheirXchange: true, exchangeFields: exchangeFields, exchangeValues: exchangeValues}
	controller := NewController(settings, clock, qsoList, testIgnoreAsync)
	controller.SetLogbook(log)
	controller.SetView(view)

	return clock, log, qsoList, view, controller, settings
}

func setupEntryWithOnlyNumberTest() (core.Clock, *mocked.Log, *mocked.QSOList, *mocked.EntryView, *Controller) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	log := new(mocked.Log)
	qsoList := new(mocked.QSOList)
	view := new(mocked.EntryView)
	settings := &testSettings{myCall: "DL0ABC", enterTheirNumber: true, enterTheirXchange: false}
	controller := NewController(settings, clock, qsoList, testIgnoreAsync)
	controller.SetLogbook(log)
	controller.SetView(view)

	return clock, log, qsoList, view, controller
}

func setupEntryWithOnlyExchangeTest() (core.Clock, *mocked.Log, *mocked.QSOList, *mocked.EntryView, *Controller) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	log := new(mocked.Log)
	qsoList := new(mocked.QSOList)
	view := new(mocked.EntryView)
	settings := &testSettings{myCall: "DL0ABC", enterTheirNumber: false, enterTheirXchange: true, requireTheirXchange: true}
	controller := NewController(settings, clock, qsoList, testIgnoreAsync)
	controller.SetLogbook(log)
	controller.SetView(view)

	return clock, log, qsoList, view, controller
}

func assertQSOInput(t *testing.T, qso core.QSO, controller *Controller) {
	assert.Equal(t, qso.Callsign.String(), controller.input.callsign, "callsign")
	assert.Equal(t, qso.TheirReport.String(), controller.input.theirReport, "their report")
	assert.Equal(t, qso.TheirNumber.String(), controller.input.theirNumber, "their number")
	assert.Equal(t, qso.TheirExchange, controller.input.theirExchange, "their exchange")
	assert.Equal(t, qso.MyReport.String(), controller.input.myReport, "my report")
	assert.Equal(t, qso.MyNumber.String(), controller.input.myNumber, "my number")
	assert.Equal(t, qso.MyExchange, controller.input.myExchange, "my exchange")
	assert.Equal(t, qso.Band.String(), controller.input.band, "input band")
	assert.Equal(t, qso.Band, controller.selectedBand, "selected band")
	assert.Equal(t, qso.Mode.String(), controller.input.mode, "input mode")
	assert.Equal(t, qso.Mode, controller.selectedMode, "selected mode")
}

type testSettings struct {
	myCall                 string
	enterTheirNumber       bool
	enterTheirXchange      bool
	requireTheirXchange    bool
	exchangeFields         []conval.ExchangeField
	exchangeValues         []string
	generateSerialExchange bool
}

func (s *testSettings) Station() core.Station {
	return core.Station{
		Callsign: callsign.MustParse(s.myCall),
	}
}

func (s *testSettings) Contest() core.Contest {
	return core.Contest{
		EnterTheirNumber:       s.enterTheirNumber,
		EnterTheirXchange:      s.enterTheirXchange,
		RequireTheirXchange:    s.requireTheirXchange,
		Definition:             fieldDefinition(s.exchangeFields...),
		GenerateSerialExchange: s.generateSerialExchange,
		ExchangeValues:         s.exchangeValues,
	}
}

func testIgnoreAsync(f func()) {}

func fieldDefinition(fields ...conval.ExchangeField) *conval.Definition {
	return &conval.Definition{
		Exchange: []conval.ExchangeDefinition{
			{
				Fields: fields,
			},
		},
	}
}
