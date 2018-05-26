package core

import (
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestParseRST(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		valid    bool
		expected RST
	}{
		{"valid CW report", "599", true, "599"},
		{"valid SSB report", "59", true, "59"},
		{"valid FM repeater report", "5", true, "5"},
		{"with whitespace", " 599 ", true, "599"},
		{"empty string is invalid", "", false, ""},
		{"single digit out of range", "6", false, ""},
		{"double digit out of range", "40", false, ""},
		{"trible digit out of range", "480", false, ""},
		{"invalid characters", "a-b", false, ""},
		{"too long", "1234", false, ""},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual, err := ParseRST(tC.value)
			if err != nil && tC.valid {
				t.Errorf("expected to be valid, but got error %v", err)
			}
			if err == nil && !tC.valid {
				t.Errorf("%q should not be parsed successfully", tC.value)
			}
			if tC.valid && actual != tC.expected {
				t.Errorf("%q: expected %v but got %v", tC.value, tC.expected, actual)
			}
		})
	}
}

func TestNewLog(t *testing.T) {
	log := NewLog(NewClock())

	assert.Equal(t, QSONumber(1), log.GetNextNumber(), "next number of empty log should be 1")
	assert.Empty(t, log.GetQsosByMyNumber(), "empty log should not contain any QSO")
}

func TestLog_Log(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := staticClock{now}
	log := NewLog(clock)

	qso := QSO{MyNumber: 1}
	log.Log(qso)

	require.Equal(t, 1, len(log.GetQsosByMyNumber()), "after logging one QSO, the log should have one item")
	loggedQso := log.GetQsosByMyNumber()[0]
	assert.Equal(t, now, loggedQso.LogTimestamp, "LogTimestamp is wrong")
}

func TestLog_LogAgain(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	then := time.Date(2006, time.January, 2, 15, 5, 0, 0, time.UTC)
	clock := &staticClock{now}
	log := NewLog(clock)

	qso := QSO{MyNumber: 1, TheirNumber: 1}
	log.Log(qso)

	clock.time = then
	qso.TheirNumber = 2
	log.Log(qso)

	require.Equal(t, 2, len(log.GetQsosByMyNumber()), "log should have two items")
	lastQso := log.GetQsosByMyNumber()[1]
	assert.Equal(t, then, lastQso.LogTimestamp, "last item should have last timestamp")
	assert.Equal(t, QSONumber(2), lastQso.TheirNumber, "last item should have latest data")
}

func TestLog_GetNextNumber(t *testing.T) {
	log := NewLog(NewClock())

	qso := QSO{MyNumber: 123}
	log.Log(qso)

	assert.Equal(t, QSONumber(124), log.GetNextNumber(), "next number should be the highest existing number + 1")
}

func TestLog_Find(t *testing.T) {
	log := NewLog(NewClock())
	aa3b, _ := callsign.Parse("AA3B")
	qso := QSO{Callsign: aa3b}

	log.Log(qso)
	loggedQso, ok := log.Find(aa3b)
	assert.True(t, ok, "qso found")
	assert.Equal(t, aa3b, loggedQso.Callsign, "callsign")
}

// Mock

type mockedLog struct {
	mock.Mock
}

func (m *mockedLog) SetView(view LogView) {
	m.Called(view)
}

func (m *mockedLog) GetNextNumber() QSONumber {
	args := m.Called()
	return args.Get(0).(QSONumber)
}

func (m *mockedLog) Log(qso QSO) {
	m.Called(qso)
}

func (m *mockedLog) Find(callsign callsign.Callsign) (QSO, bool) {
	args := m.Called(callsign)
	return args.Get(0).(QSO), args.Bool(1)
}

func (m *mockedLog) GetQsosByMyNumber() []QSO {
	args := m.Called()
	return args.Get(0).([]QSO)
}

type mockedLogView struct {
	mock.Mock
}

func (m *mockedLogView) SetLog(log Log) {
	m.Called(log)
}

func (m *mockedLogView) UpdateAllRows(qsos []QSO) {
	m.Called(qsos)
}

func (m *mockedLogView) RowAdded(qso QSO) {
	m.Called(qso)
}
