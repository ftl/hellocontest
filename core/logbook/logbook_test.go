package logbook

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/mocked"
)

func TestNew(t *testing.T) {
	logbook := New(clock.New())

	assert.Equal(t, core.QSONumber(1), logbook.NextNumber(), "next number of empty log should be 1")
	assert.Empty(t, logbook.QsosOrderedByMyNumber(), "empty log should not contain any QSO")
}

func TestLoad(t *testing.T) {
	reader := new(mocked.Reader)
	reader.On("ReadAll").Return([]core.QSO{
		{MyNumber: 123},
	}, nil)

	logbook, err := Load(clock.New(), reader)
	require.NoError(t, err)

	assert.Equal(t, core.QSONumber(124), logbook.NextNumber())
}

func TestLog_Log(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	logbook := New(clock)

	qso := core.QSO{MyNumber: 1}
	logbook.Log(qso)

	require.Equal(t, 1, len(logbook.QsosOrderedByMyNumber()), "after logging one QSO, the log should have one item")
	loggedQso := logbook.QsosOrderedByMyNumber()[0]
	assert.Equal(t, now, loggedQso.LogTimestamp, "LogTimestamp is wrong")
}

func TestLog_LogAgain(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	then := time.Date(2006, time.January, 2, 15, 5, 0, 0, time.UTC)
	clock := new(mocked.Clock)
	logbook := New(clock)

	clock.On("Now").Once().Return(now)
	qso := core.QSO{MyNumber: 1, TheirNumber: 1}
	logbook.Log(qso)

	clock.On("Now").Once().Return(then)
	qso.TheirNumber = 2
	logbook.Log(qso)

	require.Equal(t, 2, len(logbook.QsosOrderedByMyNumber()), "log should have two items")
	lastQso := logbook.QsosOrderedByMyNumber()[1]
	assert.Equal(t, then, lastQso.LogTimestamp, "last item should have last timestamp")
	assert.Equal(t, core.QSONumber(2), lastQso.TheirNumber, "last item should have latest data")
}

func TestLog_EmitRowAdded(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	logbook := New(clock)
	emitted := false
	logbook.OnRowAdded(func(core.QSO) {
		emitted = true
	})

	qso := core.QSO{MyNumber: 1}
	logbook.Log(qso)

	assert.True(t, emitted)
}

func TestLog_NextNumber(t *testing.T) {
	logbook := New(clock.New())

	qso := core.QSO{MyNumber: 123}
	logbook.Log(qso)

	assert.Equal(t, core.QSONumber(124), logbook.NextNumber(), "next number should be the highest existing number + 1")
}

func TestLastBand(t *testing.T) {
	logbook := New(clock.New())
	assert.Equal(t, core.NoBand, logbook.LastBand())

	logbook.Log(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1, Band: core.Band80m})
	assert.Equal(t, core.Band80m, logbook.LastBand())
}

func TestLastMode(t *testing.T) {
	logbook := New(clock.New())
	assert.Equal(t, core.NoMode, logbook.LastMode())

	logbook.Log(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1, Mode: core.ModeDigital})
	assert.Equal(t, core.ModeDigital, logbook.LastMode())
}

func TestLog_UniqueQsosOrderedByMyNumber(t *testing.T) {
	logbook := New(clock.New())
	logbook.Log(core.QSO{Callsign: callsign.MustParse("AA3B"), MyNumber: core.QSONumber(4)})
	logbook.Log(core.QSO{Callsign: callsign.MustParse("AA1ZZZ"), MyNumber: core.QSONumber(1)})
	logbook.Log(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: core.QSONumber(3)})
	logbook.Log(core.QSO{Callsign: callsign.MustParse("S50A"), MyNumber: core.QSONumber(2)})

	actual := logbook.UniqueQsosOrderedByMyNumber()

	assert.Equal(t, core.QSONumber(1), actual[0].MyNumber)
	assert.Equal(t, core.QSONumber(2), actual[1].MyNumber)
	assert.Equal(t, core.QSONumber(3), actual[2].MyNumber)
	assert.Equal(t, core.QSONumber(4), actual[3].MyNumber)
}

func TestLog_UniqueQsosOrderedByMyNumber_Multiband(t *testing.T) {
	logbook := New(clock.New())
	logbook.Log(core.QSO{Callsign: callsign.MustParse("AA3B"), MyNumber: core.QSONumber(4)})
	logbook.Log(core.QSO{Callsign: callsign.MustParse("AA1ZZZ"), MyNumber: core.QSONumber(1), Band: core.Band80m})
	logbook.Log(core.QSO{Callsign: callsign.MustParse("AA1ZZZ"), MyNumber: core.QSONumber(5), Band: core.Band40m})
	logbook.Log(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: core.QSONumber(3)})
	logbook.Log(core.QSO{Callsign: callsign.MustParse("S50A"), MyNumber: core.QSONumber(2)})

	actual := logbook.UniqueQsosOrderedByMyNumber()

	assert.Equal(t, core.QSONumber(1), actual[0].MyNumber)
	assert.Equal(t, core.QSONumber(2), actual[1].MyNumber)
	assert.Equal(t, core.QSONumber(3), actual[2].MyNumber)
	assert.Equal(t, core.QSONumber(4), actual[3].MyNumber)
	assert.Equal(t, core.QSONumber(5), actual[4].MyNumber)
}

func TestUnique(t *testing.T) {
	c := clock.New()
	aa3b := callsign.MustParse("AA3B")
	qso1 := core.QSO{
		MyNumber:     core.QSONumber(1),
		Callsign:     aa3b,
		LogTimestamp: c.Now().Add(-10 * time.Minute),
	}
	qso2 := core.QSO{
		MyNumber:     core.QSONumber(1),
		Callsign:     aa3b,
		LogTimestamp: c.Now(),
	}

	actual := unique([]core.QSO{qso1, qso2})

	assert.Equal(t, 1, len(actual))
	assert.Equal(t, qso2, actual[0])
}

func TestUnique_Multiband(t *testing.T) {
	c := clock.New()
	aa3b := callsign.MustParse("AA3B")
	qso1 := core.QSO{
		MyNumber:     core.QSONumber(1),
		Callsign:     aa3b,
		Band:         core.Band20m,
		LogTimestamp: c.Now().Add(-10 * time.Minute),
	}
	qso2 := core.QSO{
		MyNumber:     core.QSONumber(1),
		Callsign:     aa3b,
		Band:         core.Band40m,
		LogTimestamp: c.Now().Add(-5 * time.Minute),
	}
	qso3 := core.QSO{
		MyNumber:     core.QSONumber(2),
		Callsign:     aa3b,
		Band:         core.Band80m,
		LogTimestamp: c.Now(),
	}

	actual := byMyNumber(unique([]core.QSO{qso1, qso2, qso3}))

	require.Equal(t, 2, len(actual))
	assert.Equal(t, qso2, actual[0])
	assert.Equal(t, qso3, actual[1])
}

func TestUnique_Multimode(t *testing.T) {
	c := clock.New()
	aa3b := callsign.MustParse("AA3B")
	qso1 := core.QSO{
		MyNumber:     core.QSONumber(1),
		Callsign:     aa3b,
		Band:         core.Band20m,
		Mode:         core.ModeCW,
		LogTimestamp: c.Now().Add(-10 * time.Minute),
	}
	qso2 := core.QSO{
		MyNumber:     core.QSONumber(1),
		Callsign:     aa3b,
		Band:         core.Band20m,
		Mode:         core.ModeSSB,
		LogTimestamp: c.Now().Add(-5 * time.Minute),
	}
	qso3 := core.QSO{
		MyNumber:     core.QSONumber(2),
		Callsign:     aa3b,
		Band:         core.Band20m,
		Mode:         core.ModeDigital,
		LogTimestamp: c.Now(),
	}

	actual := byMyNumber(unique([]core.QSO{qso1, qso2, qso3}))

	require.Equal(t, 2, len(actual))
	assert.Equal(t, qso2, actual[0])
	assert.Equal(t, qso3, actual[1])
}

func TestByMyNumber(t *testing.T) {
	c := clock.New()
	qsos := []core.QSO{
		{MyNumber: core.QSONumber(3), LogTimestamp: c.Now().Add(-1 * time.Minute)},
		{MyNumber: core.QSONumber(2), LogTimestamp: c.Now().Add(-2 * time.Minute)},
		{MyNumber: core.QSONumber(1), LogTimestamp: c.Now().Add(-3 * time.Minute)},
		{MyNumber: core.QSONumber(4), LogTimestamp: c.Now().Add(-50 * time.Second)},
	}

	actual := byMyNumber(qsos)

	assert.Equal(t, core.QSONumber(1), actual[0].MyNumber)
	assert.Equal(t, core.QSONumber(2), actual[1].MyNumber)
	assert.Equal(t, core.QSONumber(3), actual[2].MyNumber)
	assert.Equal(t, core.QSONumber(4), actual[3].MyNumber)
}
