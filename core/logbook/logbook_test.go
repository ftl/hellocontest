package logbook

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
)

func TestLogbook_AddQSO(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
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
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := core.QSO{MyNumber: 1}

	_, _, err := logbook.addQSOLocked(qso)
	require.NoError(t, err)

	_, _, err = logbook.addQSOLocked(qso)
	assert.Error(t, err)
}

func TestLogbook_AddQSO_addsLogTimestamp(t *testing.T) {
	now := time.Now()
	c := clock.Static(now)
	logbook := NewLogbook(c, new(testSettings), testEntity)
	qso := core.QSO{MyNumber: 1}

	logbook.AddQSO(qso)
	loggedQSO := logbook.AllQSOs()[0]

	assert.Equal(t, now, loggedQSO.LogTimestamp)
}

func TestLogbook_AddQSO_emitsQSOAdded(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := core.QSO{MyNumber: 1}
	addedQSO := core.QSO{}
	notificationReceived := false
	listener := QSOAddedListenerFunc(func(qso core.QSO) {
		notificationReceived = true
		addedQSO = qso
	})
	logbook.Notify(listener)

	logbook.AddQSO(qso)

	assert.True(t, notificationReceived)
	assert.Equal(t, qso, addedQSO)
}

func TestLogbook_AddQSO_updatesTheScore(t *testing.T) {
	logbook := withTestConvalCounter(
		NewLogbook(clock.Zero(), new(testSettings), testEntity),
		&testConvalCounter{
			scores: map[string]conval.QSOScore{
				"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
			},
		},
	)
	qso := core.QSO{MyNumber: 1, Callsign: callsign.MustParse("DL1ABC")}
	require.Equal(t, 0, logbook.Score().Result().Result())

	logbook.AddQSO(qso)

	assert.Equal(t, 2, logbook.Score().Result().Result())
}

func TestLogbook_AddQSO_emitsScoreChanged(t *testing.T) {
	logbook := withTestConvalCounter(
		NewLogbook(clock.Zero(), new(testSettings), testEntity),
		&testConvalCounter{
			scores: map[string]conval.QSOScore{
				"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
			},
		},
	)
	qso := core.QSO{MyNumber: 1, Callsign: callsign.MustParse("DL1ABC")}
	var changedScore core.Score
	listener := ScoreChangedFunc(func(score core.Score) {
		changedScore = score
	})
	logbook.Notify(listener)

	logbook.AddQSO(qso)

	assert.Equal(t, 2, changedScore.Result().Result())
}

func TestLogbook_UpdateQSO(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qsoOld := core.QSO{MyNumber: 1, TheirNumber: 1}
	qsoNew := core.QSO{MyNumber: 1, TheirNumber: 2}

	logbook.AddQSO(qsoOld)
	logbook.UpdateQSO(qsoNew)

	assert.Equal(t, qsoNew, logbook.AllQSOs()[0])
}

func TestLogbook_updateQSOLocked_cannotUpdateNewQSO(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := core.QSO{MyNumber: 1}

	_, _, _, err := logbook.updateQSOLocked(qso)
	assert.Error(t, err)
}

func TestLogbook_UpdateQSO_updatesLogTimestamp(t *testing.T) {
	logbook := NewLogbook(clock.New(), new(testSettings), testEntity)

	logbook.AddQSO(core.QSO{MyNumber: 1, TheirNumber: 1})
	qsoOld := logbook.AllQSOs()[0]

	logbook.UpdateQSO(core.QSO{MyNumber: 1, TheirNumber: 2})
	qsoNew := logbook.AllQSOs()[0]

	assert.False(t, qsoOld.LogTimestamp.IsZero())
	assert.False(t, qsoNew.LogTimestamp.IsZero())
	assert.NotEqual(t, qsoOld.LogTimestamp, qsoNew.LogTimestamp)
}

func TestLogbook_UpdateQSO_emitsLogbookCleared(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := core.QSO{MyNumber: 1}
	notificationReceived := false
	listener := LogbookClearedFunc(func() {
		notificationReceived = true
	})

	logbook.AddQSO(qso)
	logbook.Notify(listener)
	logbook.UpdateQSO(qso)

	assert.True(t, notificationReceived)
}

func TestLogbook_UpdateQSO_emitsQSOAddedForAllQSOs(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso1 := core.QSO{MyNumber: 1}
	qso2 := core.QSO{MyNumber: 2}
	qso3 := core.QSO{MyNumber: 3}
	updatedQSOs := []core.QSO{}
	listener := QSOAddedListenerFunc(func(qso core.QSO) {
		updatedQSOs = append(updatedQSOs, qso)
	})

	logbook.AddQSO(qso1)
	logbook.AddQSO(qso2)
	logbook.AddQSO(qso3)
	logbook.Notify(listener)
	logbook.UpdateQSO(qso3)

	assert.Equal(t, []core.QSO{qso1, qso2, qso3}, updatedQSOs)
}

func TestLogbook_UpdateQSO_updatesTheScore(t *testing.T) {
	logbook := withTestConvalCounter(
		NewLogbook(clock.Zero(), new(testSettings), testEntity),
		&testConvalCounter{
			scores: map[string]conval.QSOScore{
				"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
				"DL2ABC": {Points: 3, Multis: 4, Duplicate: false},
			},
		},
	)
	logbook.AddQSO(core.QSO{MyNumber: 1, Callsign: callsign.MustParse("DL1ABC")})
	require.Equal(t, 2, logbook.Score().Result().Result())

	logbook.UpdateQSO(core.QSO{MyNumber: 1, Callsign: callsign.MustParse("DL2ABC")})

	assert.Equal(t, 12, logbook.Score().Result().Result())
}

func TestLogbook_UpdateQSO_emitsScoreChanged(t *testing.T) {
	logbook := withTestConvalCounter(
		NewLogbook(clock.Zero(), new(testSettings), testEntity),
		&testConvalCounter{
			scores: map[string]conval.QSOScore{
				"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
				"DL2ABC": {Points: 3, Multis: 4, Duplicate: false},
			},
		},
	)
	var changedScore core.Score
	listener := ScoreChangedFunc(func(score core.Score) {
		changedScore = score
	})
	logbook.Notify(listener)

	logbook.AddQSO(core.QSO{MyNumber: 1, Callsign: callsign.MustParse("DL1ABC")})
	logbook.UpdateQSO(core.QSO{MyNumber: 1, Callsign: callsign.MustParse("DL2ABC")})

	assert.Equal(t, 12, changedScore.Result().Result())
}

func TestLogbook_Load_QSOsOutOfOrder(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso1 := core.QSO{MyNumber: 1}
	qso2 := core.QSO{MyNumber: 2}
	qso3 := core.QSO{MyNumber: 3}

	err := logbook.Load([]core.QSO{qso1, qso3, qso2}, nil)
	assert.NoError(t, err)

	assert.Equal(t, []core.QSO{qso1, qso2, qso3}, logbook.AllQSOs())
}

func TestLogbook_Load_UpdatedQSOs(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qsoOld := core.QSO{MyNumber: 1, TheirNumber: 1}
	qsoNew := core.QSO{MyNumber: 1, TheirNumber: 2}

	err := logbook.Load([]core.QSO{qsoOld, qsoNew}, nil)
	assert.NoError(t, err)

	assert.Equal(t, []core.QSO{qsoNew}, logbook.AllQSOs())
}

func TestLogbook_Load_emitsLogbookCleared(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := core.QSO{MyNumber: 1}
	notificationReceived := false
	listener := LogbookClearedFunc(func() {
		notificationReceived = true
	})

	logbook.Notify(listener)
	logbook.Load([]core.QSO{qso}, nil)

	assert.True(t, notificationReceived)
}

func TestLogbook_Load_emitsQSOAddedForAllQSOs(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso1 := core.QSO{MyNumber: 1}
	qso2 := core.QSO{MyNumber: 2}
	qso3 := core.QSO{MyNumber: 3}
	loadedQSOs := []core.QSO{}
	listener := QSOAddedListenerFunc(func(qso core.QSO) {
		loadedQSOs = append(loadedQSOs, qso)
	})
	logbook.Notify(listener)

	logbook.Load([]core.QSO{qso1, qso3, qso2}, nil)

	assert.Equal(t, []core.QSO{qso1, qso2, qso3}, loadedQSOs)
}

func TestLogbook_Load_updatesTheScore(t *testing.T) {
	logbook := withTestConvalCounter(
		NewLogbook(clock.Zero(), new(testSettings), testEntity),
		&testConvalCounter{
			scores: map[string]conval.QSOScore{
				"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
				"DL2ABC": {Points: 3, Multis: 4, Duplicate: false},
			},
		},
	)

	logbook.Load([]core.QSO{
		{MyNumber: 1, Callsign: callsign.MustParse("DL1ABC")},
		{MyNumber: 1, Callsign: callsign.MustParse("DL2ABC")},
	}, nil)

	assert.Equal(t, 12, logbook.Score().Result().Result())
}

func TestLogbook_Load_emitsScoreChanged(t *testing.T) {
	logbook := withTestConvalCounter(
		NewLogbook(clock.Zero(), new(testSettings), testEntity),
		&testConvalCounter{
			scores: map[string]conval.QSOScore{
				"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
				"DL2ABC": {Points: 3, Multis: 4, Duplicate: false},
			},
		},
	)
	var changedScore core.Score
	listener := ScoreChangedFunc(func(score core.Score) {
		changedScore = score
	})
	logbook.Notify(listener)

	logbook.Load([]core.QSO{
		{MyNumber: 1, Callsign: callsign.MustParse("DL1ABC")},
		{MyNumber: 1, Callsign: callsign.MustParse("DL2ABC")},
	}, nil)

	assert.Equal(t, 12, changedScore.Result().Result())
}
