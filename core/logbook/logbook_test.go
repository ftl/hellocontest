package logbook

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
)

func TestLogbook_AddQSO(t *testing.T) {
	logbook := NewLogbook(clock.Zero())
	qso1 := core.QSO{MyNumber: 1}
	qso2 := core.QSO{MyNumber: 2}
	qso3 := core.QSO{MyNumber: 3}

	require.Empty(t, 0, logbook.AllQSOs())

	logbook.AddQSO(qso1)
	assert.Equal(t, []core.QSO{qso1}, logbook.AllQSOs())

	logbook.AddQSO(qso3)
	assert.Equal(t, []core.QSO{qso1, qso3}, logbook.AllQSOs())

	logbook.AddQSO(qso2)
	assert.Equal(t, []core.QSO{qso1, qso2, qso3}, logbook.AllQSOs())
}

func TestLogbook_addQSOLocked_cannotAddAgain(t *testing.T) {
	logbook := NewLogbook(clock.Zero())
	qso := core.QSO{MyNumber: 1}

	err := logbook.addQSOLocked(qso)
	require.NoError(t, err)

	err = logbook.addQSOLocked(qso)
	assert.Error(t, err)
}

func TestLogbook_AddQSO_addsLogTimestamp(t *testing.T) {
	now := time.Now()
	c := clock.Static(now)
	logbook := NewLogbook(c)
	qso := core.QSO{MyNumber: 1}

	logbook.AddQSO(qso)
	loggedQSO := logbook.AllQSOs()[0]

	assert.Equal(t, now, loggedQSO.LogTimestamp)
}

func TestLogbook_UpdateQSO(t *testing.T) {
	logbook := NewLogbook(clock.Zero())
	qsoOld := core.QSO{MyNumber: 1, TheirNumber: 1}
	qsoNew := core.QSO{MyNumber: 1, TheirNumber: 2}

	logbook.AddQSO(qsoOld)
	logbook.UpdateQSO(qsoNew)

	assert.Equal(t, qsoNew, logbook.AllQSOs()[0])
}

func TestLogbook_updateQSOLocked_cannotUpdateNewQSO(t *testing.T) {
	logbook := NewLogbook(clock.Zero())
	qso := core.QSO{MyNumber: 1}

	err := logbook.updateQSOLocked(qso)
	assert.Error(t, err)
}

func TestLogbook_UpdateQSO_updatesLogTimestamp(t *testing.T) {
	logbook := NewLogbook(clock.New())

	logbook.AddQSO(core.QSO{MyNumber: 1, TheirNumber: 1})
	qsoOld := logbook.AllQSOs()[0]

	logbook.UpdateQSO(core.QSO{MyNumber: 1, TheirNumber: 2})
	qsoNew := logbook.AllQSOs()[0]

	assert.False(t, qsoOld.LogTimestamp.IsZero())
	assert.False(t, qsoNew.LogTimestamp.IsZero())
	assert.NotEqual(t, qsoOld.LogTimestamp, qsoNew.LogTimestamp)
}

func TestLogbook_Load_QSOsOutOfOrder(t *testing.T) {
	logbook := NewLogbook(clock.Zero())
	qso1 := core.QSO{MyNumber: 1}
	qso2 := core.QSO{MyNumber: 2}
	qso3 := core.QSO{MyNumber: 3}

	err := logbook.Load([]core.QSO{qso1, qso3, qso2}, nil)
	assert.NoError(t, err)

	assert.Equal(t, []core.QSO{qso1, qso2, qso3}, logbook.AllQSOs())
}

func TestLogbook_Load_QSOsUpdated(t *testing.T) {
	logbook := NewLogbook(clock.Zero())
	qsoOld := core.QSO{MyNumber: 1, TheirNumber: 1}
	qsoNew := core.QSO{MyNumber: 1, TheirNumber: 2}

	err := logbook.Load([]core.QSO{qsoOld, qsoNew}, nil)
	assert.NoError(t, err)

	assert.Equal(t, []core.QSO{qsoNew}, logbook.AllQSOs())
}
