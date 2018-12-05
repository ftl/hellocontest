package log

import (
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/mocked"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	log := New(clock.New())

	assert.Equal(t, core.QSONumber(1), log.GetNextNumber(), "next number of empty log should be 1")
	assert.Empty(t, log.GetQsosByMyNumber(), "empty log should not contain any QSO")
}

func TestLoad(t *testing.T) {
	reader := new(mocked.Reader)
	reader.On("ReadAll").Return([]core.QSO{
		{MyNumber: 123},
	}, nil)

	log, err := Load(clock.New(), reader)
	require.NoError(t, err)

	assert.Equal(t, core.QSONumber(124), log.GetNextNumber())
}

func TestLog_Log(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	log := New(clock)

	qso := core.QSO{MyNumber: 1}
	log.Log(qso)

	require.Equal(t, 1, len(log.GetQsosByMyNumber()), "after logging one QSO, the log should have one item")
	loggedQso := log.GetQsosByMyNumber()[0]
	assert.Equal(t, now, loggedQso.LogTimestamp, "LogTimestamp is wrong")
}

func TestLog_LogAgain(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	then := time.Date(2006, time.January, 2, 15, 5, 0, 0, time.UTC)
	clock := new(mocked.Clock)
	log := New(clock)

	clock.On("Now").Once().Return(now)
	qso := core.QSO{MyNumber: 1, TheirNumber: 1}
	log.Log(qso)

	clock.On("Now").Once().Return(then)
	qso.TheirNumber = 2
	log.Log(qso)

	require.Equal(t, 2, len(log.GetQsosByMyNumber()), "log should have two items")
	lastQso := log.GetQsosByMyNumber()[1]
	assert.Equal(t, then, lastQso.LogTimestamp, "last item should have last timestamp")
	assert.Equal(t, core.QSONumber(2), lastQso.TheirNumber, "last item should have latest data")
}

func TestLog_EmitRowAdded(t *testing.T) {
	now := time.Date(2006, time.January, 2, 15, 4, 5, 6, time.UTC)
	clock := clock.Static(now)
	log := New(clock)
	emitted := false
	log.OnRowAdded(func(core.QSO) error {
		emitted = true
		return nil
	})

	qso := core.QSO{MyNumber: 1}
	log.Log(qso)

	assert.True(t, emitted)
}

func TestLog_GetNextNumber(t *testing.T) {
	log := New(clock.New())

	qso := core.QSO{MyNumber: 123}
	log.Log(qso)

	assert.Equal(t, core.QSONumber(124), log.GetNextNumber(), "next number should be the highest existing number + 1")
}

func TestLog_Find(t *testing.T) {
	log := New(clock.New())
	aa3b, _ := callsign.Parse("AA3B")
	qso := core.QSO{Callsign: aa3b}

	log.Log(qso)
	loggedQso, ok := log.Find(aa3b)
	assert.True(t, ok, "qso found")
	assert.Equal(t, aa3b, loggedQso.Callsign, "callsign")
}
