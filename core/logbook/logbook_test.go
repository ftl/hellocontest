package logbook

import (
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/mocked"
)

func TestNew(t *testing.T) {
	logbook := New(clock.New())

	assert.Equal(t, core.QSONumber(1), logbook.NextNumber(), "next number of empty log should be 1")
	assert.Empty(t, logbook.AllQSOs(), "empty log should not contain any QSO")
}

func TestLoad(t *testing.T) {
	qsos := []core.QSO{
		{MyNumber: 123},
	}
	now := time.Now()
	qtcs := []core.QTC{
		{Header: core.QTCHeader{SeriesNumber: 1, QTCCount: 1}, Kind: core.SentQTC, QSONumber: 234, Timestamp: now.Add(-1 * time.Minute)},
		{Header: core.QTCHeader{SeriesNumber: 24, QTCCount: 1}, Kind: core.ReceivedQTC, QSONumber: 235, Timestamp: now},
	}

	logbook := Load(clock.New(), qsos, qtcs)

	assert.Equal(t, core.QSONumber(124), logbook.NextNumber())

	actualQTCs := logbook.AllQTCs()
	assert.Equal(t, qtcs, actualQTCs)
	assert.Equal(t, 2, logbook.NextSeriesNumber())
}

func TestLog_Log(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	logbook := New(clock)

	qso := core.QSO{MyNumber: 1}
	logbook.LogQSO(qso)

	require.Equal(t, 1, len(logbook.AllQSOs()), "after logging one QSO, the log should have one item")
	loggedQso := logbook.AllQSOs()[0]
	assert.Equal(t, now, loggedQso.LogTimestamp, "LogTimestamp is wrong")
}

func TestLog_LogAgain(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	then := time.Date(2006, time.January, 2, 15, 5, 0, 0, time.UTC)
	clock := new(mocked.Clock)
	logbook := New(clock)

	clock.On("Now").Once().Return(now)
	qso := core.QSO{MyNumber: 1, TheirNumber: 1}
	logbook.LogQSO(qso)

	clock.On("Now").Once().Return(then)
	qso.TheirNumber = 2
	logbook.LogQSO(qso)

	require.Equal(t, 2, len(logbook.AllQSOs()), "log should have two items")
	lastQso := logbook.AllQSOs()[1]
	assert.Equal(t, then, lastQso.LogTimestamp, "last item should have last timestamp")
	assert.Equal(t, core.QSONumber(2), lastQso.TheirNumber, "last item should have latest data")
}

func TestLog_EmitRowAdded(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	logbook := New(clock)
	emitted := false
	logbook.Notify(QSOAddedListenerFunc(func(core.QSO) {
		emitted = true
	}))

	qso := core.QSO{MyNumber: 1}
	logbook.LogQSO(qso)

	assert.True(t, emitted)
}

func TestLog_NextNumber(t *testing.T) {
	logbook := New(clock.New())

	qso := core.QSO{MyNumber: 123}
	logbook.LogQSO(qso)

	assert.Equal(t, core.QSONumber(124), logbook.NextNumber(), "next number should be the highest existing number + 1")
}

func TestLastBand(t *testing.T) {
	logbook := New(clock.New())
	assert.Equal(t, core.NoBand, logbook.LastBand())

	logbook.LogQSO(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1, Band: core.Band80m})
	assert.Equal(t, core.Band80m, logbook.LastBand())
}

func TestLastMode(t *testing.T) {
	logbook := New(clock.New())
	assert.Equal(t, core.NoMode, logbook.LastMode())

	logbook.LogQSO(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1, Mode: core.ModeDigital})
	assert.Equal(t, core.ModeDigital, logbook.LastMode())
}
