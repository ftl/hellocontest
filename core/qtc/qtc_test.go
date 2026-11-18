package qtc

import (
	"testing"
	"time"

	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/model"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	c := newController().Build()
	assert.NotNil(t, c)
}

func TestOfferQTC_HappyPath(t *testing.T) {
	now := time.Now()
	view := new(fakeView)
	keyer := new(fakeKeyer)
	theirCallsign := model.MustParseCallsign("DL1ABC")
	logbook := &fakeLogbook{
		nextSeriesNumber: 4,
		lastCallsign:     theirCallsign,
		availableQTCs: qtcsFor(model.SentQTC, theirCallsign).
			Add("0123", "DK1AB", 1).
			Add("24", "DK2AB", 2).
			Build(),
	}
	c := newController().
		WithClock(clock.Static(now)).
		WithKeyer(keyer).
		WithLogbook(logbook).
		WithVFOState(7020000, model.Band40m, model.ModeCW).
		Build()
	c.SetView(view)

	c.OfferQTC()
	assert.Equal(t, "qtc", keyer.lastTransmission)
	assert.Equal(t, model.QTCStart, c.activePhase)
	assert.True(t, view.visible)
	assert.Equal(t, model.ProvideQTC, view.mode)
	assert.Equal(t, theirCallsign, view.series.TheirCallsign)
	assert.Equal(t, "4/2", view.series.Header.String())
	assert.Equal(t, 2, len(view.series.QTCs))
	assert.Equal(t, "DK1AB", view.series.QTCs[0].QTCCallsign.String())
	assert.Equal(t, "DK2AB", view.series.QTCs[1].QTCCallsign.String())
	assert.False(t, view.series.QTCs[0].Confirmed)
	assert.False(t, view.series.QTCs[1].Confirmed)

	c.ConfirmStart()
	assert.Equal(t, "qtc 4/2", keyer.lastTransmission)
	assert.Equal(t, model.QTCExchangeHeader, c.activePhase)
	assert.True(t, view.visible)

	c.ConfirmHeader()
	assert.Equal(t, "0123 DK1AB 001", keyer.lastTransmission)
	assert.Equal(t, model.QTCExchangeData, c.activePhase)
	assert.True(t, view.visible)
	assert.False(t, view.series.QTCs[0].Confirmed)
	assert.False(t, view.series.QTCs[1].Confirmed)

	c.ConfirmData()
	assert.Equal(t, "24 DK2AB 002", keyer.lastTransmission)
	assert.Equal(t, model.QTCExchangeData, c.activePhase)
	assert.True(t, view.visible)
	assert.True(t, view.series.QTCs[0].Confirmed)
	assert.False(t, view.series.QTCs[1].Confirmed)
	assert.Equal(t, model.Frequency(7020000), c.currentSeries.QTCs[0].Frequency)
	assert.Equal(t, model.Band40m, c.currentSeries.QTCs[0].Band)
	assert.Equal(t, model.ModeCW, c.currentSeries.QTCs[0].Mode)

	c.ConfirmData()
	assert.Equal(t, "24 DK2AB 002", keyer.lastTransmission)
	assert.Equal(t, model.QTCFinish, c.activePhase)
	assert.True(t, view.visible)
	assert.True(t, view.series.QTCs[0].Confirmed)
	assert.True(t, view.series.QTCs[1].Confirmed)
	assert.Equal(t, model.Frequency(7020000), c.currentSeries.QTCs[1].Frequency)
	assert.Equal(t, model.Band40m, c.currentSeries.QTCs[1].Band)
	assert.Equal(t, model.ModeCW, c.currentSeries.QTCs[1].Mode)

	c.CompleteQTCSeries()
	assert.Equal(t, "tu", keyer.lastTransmission)
	assert.False(t, view.visible)
	assert.Equal(t, 2, len(logbook.loggedQTCs))
	assert.Equal(t, model.QTC{
		Kind:          model.SentQTC,
		TheirCallsign: theirCallsign,
		Header:        model.QTCHeader{SeriesNumber: 4, QTCCount: 2},
		Timestamp:     now,
		Confirmed:     true,
		Frequency:     model.Frequency(7020000),
		Band:          model.Band40m,
		Mode:          model.ModeCW,
		QTCTime:       model.QTCTime{Hour: 1, Minute: 23},
		QTCCallsign:   model.MustParseCallsign("DK1AB"),
		QTCNumber:     model.QSONumber(1),
	}, logbook.loggedQTCs[0])
	assert.Equal(t, model.QTC{
		Kind:          model.SentQTC,
		TheirCallsign: theirCallsign,
		Header:        model.QTCHeader{SeriesNumber: 4, QTCCount: 2},
		Timestamp:     now,
		Confirmed:     true,
		Frequency:     model.Frequency(7020000),
		Band:          model.Band40m,
		Mode:          model.ModeCW,
		QTCTime:       model.QTCTime{Hour: 1, Minute: 24},
		QTCCallsign:   model.MustParseCallsign("DK2AB"),
		QTCNumber:     model.QSONumber(2),
	}, logbook.loggedQTCs[1])
}

func TestRequestQTC_HappyPath(t *testing.T) {
	now := time.Now()
	view := new(fakeView)
	keyer := new(fakeKeyer)
	theirCallsign := model.MustParseCallsign("DL1ABC")
	logbook := &fakeLogbook{lastCallsign: theirCallsign}
	c := newController().
		WithClock(clock.Static(now)).
		WithKeyer(keyer).
		WithLogbook(logbook).
		WithVFOState(7020000, model.Band40m, model.ModeCW).
		Build()
	c.SetView(view)

	c.RequestQTC()
	assert.Equal(t, model.QTCStart, c.activePhase)
	assert.True(t, view.visible)
	assert.Equal(t, model.ReceiveQTC, view.mode)
	assert.Equal(t, theirCallsign, view.series.TheirCallsign)
	assert.Equal(t, "", keyer.lastTransmission)

	c.StartAction()
	assert.Equal(t, model.QTCStart, c.activePhase)
	assert.Equal(t, "qtc?", keyer.lastTransmission)

	c.ConfirmStart() // qrv
	assert.Equal(t, "qrv", keyer.lastTransmission)
	assert.Equal(t, model.QTCExchangeHeader, c.activePhase)
	assert.Equal(t, model.QTCHeaderField, c.activeField)
	assert.Equal(t, model.QTCHeaderField, view.field)

	c.Enter("4/2") // header
	assert.Equal(t, "4/2", c.currentInput[model.QTCHeaderField])

	c.ConfirmHeader() // r
	assert.Equal(t, "r", keyer.lastTransmission)
	assert.Equal(t, model.QTCExchangeData, c.activePhase)
	assert.Equal(t, model.QTCTimestampField, c.activeField)
	assert.Equal(t, model.QTCTimestampField, view.field)
	assert.Equal(t, 0, c.currentQTC)
	assert.Equal(t, model.QTCHeader{SeriesNumber: 4, QTCCount: 2}, c.currentSeries.Header)

	c.Enter("0123")
	assert.Equal(t, "0123", c.currentInput[model.QTCTimestampField])

	c.GotoNextField()
	assert.Equal(t, model.QTCCallsignField, c.activeField)
	assert.Equal(t, model.QTCCallsignField, view.field)

	c.Enter("DK1AB")
	assert.Equal(t, "DK1AB", c.currentInput[model.QTCCallsignField])

	c.GotoNextField()
	assert.Equal(t, model.QTCExchangeField, c.activeField)
	assert.Equal(t, model.QTCExchangeField, view.field)

	c.Enter("1")
	assert.Equal(t, "1", c.currentInput[model.QTCExchangeField])

	c.ConfirmData() // r
	assert.Equal(t, "r", keyer.lastTransmission)
	assert.Equal(t, model.QTCExchangeData, c.activePhase)
	assert.Equal(t, model.QTCTimestampField, c.activeField)
	assert.Equal(t, model.QTCTimestampField, view.field)
	assert.Equal(t, 1, c.currentQTC)
	assert.Equal(t, model.QTC{
		Kind:        model.ReceivedQTC,
		Timestamp:   now,
		Confirmed:   true,
		QTCTime:     model.QTCTime{Hour: 1, Minute: 23},
		QTCCallsign: model.MustParseCallsign("DK1AB"),
		QTCNumber:   model.QSONumber(1),
	}, c.currentSeries.QTCs[0])
	assert.Equal(t, model.QTC{
		Kind:        model.ReceivedQTC,
		Timestamp:   now,
		Confirmed:   true,
		QTCTime:     model.QTCTime{Hour: 1, Minute: 23},
		QTCCallsign: model.MustParseCallsign("DK1AB"),
		QTCNumber:   model.QSONumber(1),
	}, view.series.QTCs[0])
	assert.Equal(t, "", c.currentInput[model.QTCTimestampField])
	assert.Equal(t, "", c.currentInput[model.QTCCallsignField])
	assert.Equal(t, "", c.currentInput[model.QTCExchangeField])

	c.Enter("24")
	assert.Equal(t, "24", c.currentInput[model.QTCTimestampField])

	c.GotoNextField()
	assert.Equal(t, model.QTCCallsignField, c.activeField)
	assert.Equal(t, model.QTCCallsignField, view.field)

	c.Enter("DK2AB")
	assert.Equal(t, "DK2AB", c.currentInput[model.QTCCallsignField])

	c.GotoNextField()
	assert.Equal(t, model.QTCExchangeField, c.activeField)
	assert.Equal(t, model.QTCExchangeField, view.field)

	c.Enter("2")
	assert.Equal(t, "2", c.currentInput[model.QTCExchangeField])

	c.ConfirmData() // r
	assert.Equal(t, "r", keyer.lastTransmission)
	assert.Equal(t, model.QTCFinish, c.activePhase)
	assert.Equal(t, 2, c.currentQTC)
	assert.Equal(t, model.QTC{
		Kind:        model.ReceivedQTC,
		Timestamp:   now,
		Confirmed:   true,
		QTCTime:     model.QTCTime{Hour: 1, Minute: 24},
		QTCCallsign: model.MustParseCallsign("DK2AB"),
		QTCNumber:   model.QSONumber(2),
	}, c.currentSeries.QTCs[1])
	assert.Equal(t, model.QTC{
		Kind:        model.ReceivedQTC,
		Timestamp:   now,
		Confirmed:   true,
		QTCTime:     model.QTCTime{Hour: 1, Minute: 24},
		QTCCallsign: model.MustParseCallsign("DK2AB"),
		QTCNumber:   model.QSONumber(2),
	}, view.series.QTCs[1])
	assert.Equal(t, "", c.currentInput[model.QTCTimestampField])
	assert.Equal(t, "", c.currentInput[model.QTCCallsignField])
	assert.Equal(t, "", c.currentInput[model.QTCExchangeField])

	keyer.lastTransmission = ""
	c.CompleteQTCSeries()
	assert.Equal(t, "", keyer.lastTransmission)
	assert.False(t, view.visible)
	assert.Equal(t, 2, len(logbook.loggedQTCs))
	assert.Equal(t, model.QTC{
		Kind:          model.ReceivedQTC,
		TheirCallsign: theirCallsign,
		Header:        model.QTCHeader{SeriesNumber: 4, QTCCount: 2},
		Timestamp:     now,
		Confirmed:     true,
		Frequency:     model.Frequency(7020000),
		Band:          model.Band40m,
		Mode:          model.ModeCW,
		QTCTime:       model.QTCTime{Hour: 1, Minute: 23},
		QTCCallsign:   model.MustParseCallsign("DK1AB"),
		QTCNumber:     model.QSONumber(1),
	}, logbook.loggedQTCs[0])
	assert.Equal(t, model.QTC{
		Kind:          model.ReceivedQTC,
		TheirCallsign: theirCallsign,
		Header:        model.QTCHeader{SeriesNumber: 4, QTCCount: 2},
		Timestamp:     now,
		Confirmed:     true,
		Frequency:     model.Frequency(7020000),
		Band:          model.Band40m,
		Mode:          model.ModeCW,
		QTCTime:       model.QTCTime{Hour: 1, Minute: 24},
		QTCCallsign:   model.MustParseCallsign("DK2AB"),
		QTCNumber:     model.QSONumber(2),
	}, logbook.loggedQTCs[1])
}

func TestOfferQTC_ShowsFocusedVFONameOnlyInSO2V(t *testing.T) {
	theirCallsign := model.MustParseCallsign("DL1ABC")
	newViewFor := func(so2v bool) *fakeView {
		view := new(fakeView)
		logbook := &fakeLogbook{
			nextSeriesNumber: 1,
			lastCallsign:     theirCallsign,
			availableQTCs: qtcsFor(model.SentQTC, theirCallsign).
				Add("0123", "DK1AB", 1).
				Build(),
		}
		c := newController().
			WithKeyer(new(fakeKeyer)).
			WithLogbook(logbook).
			WithFocusedVFO("VFO 2", so2v).
			Build()
		c.SetView(view)
		c.OfferQTC()
		return view
	}

	assert.Equal(t, "VFO 2", newViewFor(true).vfoName)
	assert.Equal(t, "", newViewFor(false).vfoName)
}

func TestOfferQTC_SO2V_StampsFocusedVFOState(t *testing.T) {
	now := time.Now()
	view := new(fakeView)
	keyer := new(fakeKeyer)
	theirCallsign := model.MustParseCallsign("DL1ABC")
	logbook := &fakeLogbook{
		nextSeriesNumber: 1,
		lastCallsign:     theirCallsign,
		availableQTCs: qtcsFor(model.SentQTC, theirCallsign).
			Add("0123", "DK1AB", 1).
			Build(),
	}
	// the operator works the QTC series on the focused (sub) VFO, tuned to 15m
	c := newController().
		WithClock(clock.Static(now)).
		WithKeyer(keyer).
		WithLogbook(logbook).
		WithVFOState(21020000, model.Band15m, model.ModeCW).
		Build()
	c.SetView(view)

	c.OfferQTC()
	c.ConfirmStart()
	c.ConfirmHeader() // sends and stamps the first QTC
	c.ConfirmData()

	assert.Equal(t, model.Frequency(21020000), c.currentSeries.QTCs[0].Frequency)
	assert.Equal(t, model.Band15m, c.currentSeries.QTCs[0].Band)
	assert.Equal(t, model.ModeCW, c.currentSeries.QTCs[0].Mode)

	c.CompleteQTCSeries()
	assert.Equal(t, 1, len(logbook.loggedQTCs))
	assert.Equal(t, model.Frequency(21020000), logbook.loggedQTCs[0].Frequency)
	assert.Equal(t, model.Band15m, logbook.loggedQTCs[0].Band)
	assert.Equal(t, model.ModeCW, logbook.loggedQTCs[0].Mode)
}

func TestRequestQTC_HeaderErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "the header field is empty",
		},
		{
			name:     "only one number",
			input:    "4",
			expected: "\"4\" is not a valid QTC header",
		},
		{
			name:     "some letter",
			input:    "a",
			expected: "\"a\" is not a valid QTC header",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			now := time.Now()
			view := new(fakeView)
			keyer := new(fakeKeyer)
			theirCallsign := model.MustParseCallsign("DL1ABC")
			logbook := &fakeLogbook{lastCallsign: theirCallsign}
			c := newController().
				WithClock(clock.Static(now)).
				WithKeyer(keyer).
				WithLogbook(logbook).
				Build()
			c.SetView(view)
			c.RequestQTC()
			c.StartAction()
			c.ConfirmStart()

			c.Enter(test.input)
			c.ConfirmHeader()
			assert.Equal(t, test.expected, view.errorMessage)
		})
	}
}

func TestRequestQTC_DataErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         [3]string
		expectedError string
		expectedField model.QTCField
	}{
		{
			name:          "empty timestamp",
			input:         [...]string{"", "", ""},
			expectedError: "the QTC timestamp field is empty",
			expectedField: model.QTCTimestampField,
		},
		{
			name:          "too long timestamp",
			input:         [...]string{"12345", "", ""},
			expectedError: "cannot parse QTC time: \"12345\" is invalid",
			expectedField: model.QTCTimestampField,
		},
		{
			name:          "timestamp with letter",
			input:         [...]string{"ab", "", ""},
			expectedError: "cannot parse QTC time: \"ab\" is not a number",
			expectedField: model.QTCTimestampField,
		},
		{
			name:          "timestamp with invalid hour",
			input:         [...]string{"2400", "", ""},
			expectedError: "cannot parse QTC time: 24 is not valid for the hour section",
			expectedField: model.QTCTimestampField,
		},
		{
			name:          "timestamp with invalid minutes",
			input:         [...]string{"99", "", ""},
			expectedError: "cannot parse QTC time: 99 is not valid for the minute section",
			expectedField: model.QTCTimestampField,
		},
		{
			name:          "empty callsign",
			input:         [...]string{"0123", "", ""},
			expectedError: "the QTC callsign field is empty",
			expectedField: model.QTCCallsignField,
		},
		{
			name:          "invalid callsign",
			input:         [...]string{"0123", "12345", ""},
			expectedError: "\"12345\" is not a valid callsign",
			expectedField: model.QTCCallsignField,
		},
		{
			name:          "empty exchange",
			input:         [...]string{"0123", "DK1AB", ""},
			expectedError: "the QTC exchange field is empty",
			expectedField: model.QTCExchangeField,
		},
		{
			name:          "a letter as exchange",
			input:         [...]string{"0123", "DK1AB", "a"},
			expectedError: "\"a\" is not a valid QTC exchange",
			expectedField: model.QTCExchangeField,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			now := time.Now()
			view := new(fakeView)
			keyer := new(fakeKeyer)
			theirCallsign := model.MustParseCallsign("DL1ABC")
			logbook := &fakeLogbook{lastCallsign: theirCallsign}
			c := newController().
				WithClock(clock.Static(now)).
				WithKeyer(keyer).
				WithLogbook(logbook).
				Build()
			c.SetView(view)
			c.RequestQTC()
			c.StartAction()
			c.ConfirmStart()
			c.Enter("4/2")
			c.ConfirmHeader()

			c.Enter(test.input[0])
			c.GotoNextField()
			c.Enter(test.input[1])
			c.GotoNextField()
			c.Enter(test.input[2])
			c.ConfirmData()
			assert.Equal(t, test.expectedError, view.errorMessage)
			assert.Equal(t, test.expectedField, c.activeField)
			assert.Equal(t, test.expectedField, view.field)
		})
	}
}
