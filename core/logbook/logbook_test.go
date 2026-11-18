package logbook

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/conval"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/model"
)

func TestLogbook_AddQSO(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso1 := model.QSO{MyNumber: 1}
	qso2 := model.QSO{MyNumber: 2}
	qso3 := model.QSO{MyNumber: 3}

	require.Empty(t, 0, logbook.AllQSOs())

	logbook.AddQSO(qso1)
	assert.Equal(t, []model.QSO{qso1}, logbook.AllQSOs())

	logbook.AddQSO(qso3)
	assert.Equal(t, []model.QSO{qso1, qso3}, logbook.AllQSOs())

	logbook.AddQSO(qso2)
	assert.Equal(t, []model.QSO{qso1, qso2, qso3}, logbook.AllQSOs())
}

func TestLogbook_addQSOLocked_cannotAddAgain(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := model.QSO{MyNumber: 1}

	_, _, _, _, err := logbook.addQSOLocked(qso)
	require.NoError(t, err)

	_, _, _, _, err = logbook.addQSOLocked(qso)
	assert.Error(t, err)
}

func TestLogbook_AddQSO_addsLogTimestamp(t *testing.T) {
	now := time.Now()
	c := clock.Static(now)
	logbook := NewLogbook(c, new(testSettings), testEntity)
	qso := model.QSO{MyNumber: 1}

	logbook.AddQSO(qso)
	loggedQSO := logbook.AllQSOs()[0]

	assert.Equal(t, now, loggedQSO.LogTimestamp)
}

func TestLogbook_AddQSO_emitsQSOAdded(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := model.QSO{MyNumber: 1}
	addedQSO := model.QSO{}
	notificationReceived := false
	listener := QSOAddedListenerFunc(func(qso model.QSO) {
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
	qso := model.QSO{MyNumber: 1, Callsign: model.MustParseCallsign("DL1ABC")}
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
	qso := model.QSO{MyNumber: 1, Callsign: model.MustParseCallsign("DL1ABC")}
	var changedScore model.Score
	listener := ScoreChangedFunc(func(score model.Score) {
		changedScore = score
	})
	logbook.Notify(listener)

	logbook.AddQSO(qso)

	assert.Equal(t, 2, changedScore.Result().Result())
}

func TestLogbook_UpdateQSO(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qsoOld := model.QSO{MyNumber: 1, TheirNumber: 1}
	qsoNew := model.QSO{MyNumber: 1, TheirNumber: 2}

	logbook.AddQSO(qsoOld)
	logbook.UpdateQSO(qsoNew)

	assert.Equal(t, qsoNew, logbook.AllQSOs()[0])
}

func TestLogbook_updateQSOLocked_cannotUpdateNewQSO(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := model.QSO{MyNumber: 1}

	_, _, _, err := logbook.updateQSOLocked(qso)
	assert.Error(t, err)
}

func TestLogbook_UpdateQSO_updatesLogTimestamp(t *testing.T) {
	logbook := NewLogbook(clock.New(), new(testSettings), testEntity)

	logbook.AddQSO(model.QSO{MyNumber: 1, TheirNumber: 1})
	qsoOld := logbook.AllQSOs()[0]

	logbook.UpdateQSO(model.QSO{MyNumber: 1, TheirNumber: 2})
	qsoNew := logbook.AllQSOs()[0]

	assert.False(t, qsoOld.LogTimestamp.IsZero())
	assert.False(t, qsoNew.LogTimestamp.IsZero())
	assert.NotEqual(t, qsoOld.LogTimestamp, qsoNew.LogTimestamp)
}

func TestLogbook_UpdateQSO_emitsLogbookCleared(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := model.QSO{MyNumber: 1}
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
	qso1 := model.QSO{MyNumber: 1}
	qso2 := model.QSO{MyNumber: 2}
	qso3 := model.QSO{MyNumber: 3}
	updatedQSOs := []model.QSO{}
	listener := QSOAddedListenerFunc(func(qso model.QSO) {
		updatedQSOs = append(updatedQSOs, qso)
	})

	logbook.AddQSO(qso1)
	logbook.AddQSO(qso2)
	logbook.AddQSO(qso3)
	logbook.Notify(listener)
	logbook.UpdateQSO(qso3)

	assert.Equal(t, []model.QSO{qso1, qso2, qso3}, updatedQSOs)
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
	logbook.AddQSO(model.QSO{MyNumber: 1, Callsign: model.MustParseCallsign("DL1ABC")})
	require.Equal(t, 2, logbook.Score().Result().Result())

	logbook.UpdateQSO(model.QSO{MyNumber: 1, Callsign: model.MustParseCallsign("DL2ABC")})

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
	var changedScore model.Score
	listener := ScoreChangedFunc(func(score model.Score) {
		changedScore = score
	})
	logbook.Notify(listener)

	logbook.AddQSO(model.QSO{MyNumber: 1, Callsign: model.MustParseCallsign("DL1ABC")})
	logbook.UpdateQSO(model.QSO{MyNumber: 1, Callsign: model.MustParseCallsign("DL2ABC")})

	assert.Equal(t, 12, changedScore.Result().Result())
}

func TestLogbook_Load_QSOsOutOfOrder(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso1 := model.QSO{MyNumber: 1}
	qso2 := model.QSO{MyNumber: 2}
	qso3 := model.QSO{MyNumber: 3}

	err := logbook.Load(new(nullWriter), []model.QSO{qso1, qso3, qso2}, nil)
	assert.NoError(t, err)

	assert.Equal(t, []model.QSO{qso1, qso2, qso3}, logbook.AllQSOs())
}

func TestLogbook_Load_UpdatedQSOs(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qsoOld := model.QSO{MyNumber: 1, TheirNumber: 1}
	qsoNew := model.QSO{MyNumber: 1, TheirNumber: 2}

	err := logbook.Load(new(nullWriter), []model.QSO{qsoOld, qsoNew}, nil)
	assert.NoError(t, err)

	assert.Equal(t, []model.QSO{qsoNew}, logbook.AllQSOs())
}

func TestLogbook_Load_emitsLogbookCleared(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso := model.QSO{MyNumber: 1}
	notificationReceived := false
	listener := LogbookClearedFunc(func() {
		notificationReceived = true
	})

	logbook.Notify(listener)
	logbook.Load(new(nullWriter), []model.QSO{qso}, nil)

	assert.True(t, notificationReceived)
}

func TestLogbook_Load_emitsQSOAddedForAllQSOs(t *testing.T) {
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	qso1 := model.QSO{MyNumber: 1}
	qso2 := model.QSO{MyNumber: 2}
	qso3 := model.QSO{MyNumber: 3}
	loadedQSOs := []model.QSO{}
	listener := QSOAddedListenerFunc(func(qso model.QSO) {
		loadedQSOs = append(loadedQSOs, qso)
	})
	logbook.Notify(listener)

	logbook.Load(new(nullWriter), []model.QSO{qso1, qso3, qso2}, nil)

	assert.Equal(t, []model.QSO{qso1, qso2, qso3}, loadedQSOs)
}

func TestLogbook_Load_updatesTheScore(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	sentHeader := model.QTCHeader{SeriesNumber: 1, QTCCount: 3}
	dl2abc := model.MustParseCallsign("DL2ABC")
	receivedHeader := model.QTCHeader{SeriesNumber: 12, QTCCount: 3}
	now := time.Now()
	qtc1 := model.QTC{Kind: model.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: sentHeader, Timestamp: now.Add(-5 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc2 := model.QTC{Kind: model.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: sentHeader, Timestamp: now.Add(-4 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc3 := model.QTC{Kind: model.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: sentHeader, Timestamp: now.Add(-3 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc4 := model.QTC{Kind: model.ReceivedQTC, QTCNumber: 1, TheirCallsign: dl2abc, Header: receivedHeader, Timestamp: now.Add(-2 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc5 := model.QTC{Kind: model.ReceivedQTC, QTCNumber: 2, TheirCallsign: dl2abc, Header: receivedHeader, Timestamp: now.Add(-1 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc6 := model.QTC{Kind: model.ReceivedQTC, QTCNumber: 3, TheirCallsign: dl2abc, Header: receivedHeader, Timestamp: now, Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	logbook := withTestConvalCounter(
		NewLogbook(clock.Zero(), new(testSettings), testEntity),
		&testConvalCounter{
			scores: map[string]conval.QSOScore{
				"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
				"DL2ABC": {Points: 3, Multis: 4, Duplicate: false},
			},
		},
	)

	logbook.Load(new(nullWriter),
		[]model.QSO{
			{MyNumber: 1, Callsign: dl1abc, Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000},
			// update the callsign of the first QSO
			{MyNumber: 1, Callsign: dl2abc, Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000},
		},
		[]model.QTC{qtc1, qtc2, qtc3, qtc4, qtc5, qtc6},
	)

	assert.Equal(t, 36, logbook.Score().Result().Result())
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
	var changedScore model.Score
	listener := ScoreChangedFunc(func(score model.Score) {
		changedScore = score
	})
	logbook.Notify(listener)

	logbook.Load(new(nullWriter), []model.QSO{
		{MyNumber: 1, Callsign: model.MustParseCallsign("DL1ABC")},
		{MyNumber: 1, Callsign: model.MustParseCallsign("DL2ABC")},
	}, nil)

	assert.Equal(t, 12, changedScore.Result().Result())
}

func TestLogbook_Load_loadsQTCs(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	header := model.QTCHeader{SeriesNumber: 1, QTCCount: 2}
	now := time.Now()
	qtcs := []model.QTC{
		{Kind: model.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-2 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000},
		{Kind: model.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-1 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000},
		{Kind: model.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: header, Timestamp: now, Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000},
	}
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)

	logbook.Load(new(nullWriter), nil, qtcs)

	assert.Equal(t, qtcs, logbook.AllQTCs())
}

func TestLogbook_Load_emitsQTCAddedForAllQTCs(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	header := model.QTCHeader{SeriesNumber: 1, QTCCount: 2}
	now := time.Now()
	qtc1 := model.QTC{Kind: model.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-2 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc2 := model.QTC{Kind: model.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-1 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc3 := model.QTC{Kind: model.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: header, Timestamp: now, Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	loadedQTCs := []model.QTC{}
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	logbook.Notify(QTCAddedListenerFunc(func(qtc model.QTC) {
		loadedQTCs = append(loadedQTCs, qtc)
	}))

	logbook.Load(new(nullWriter), nil, []model.QTC{qtc1, qtc2, qtc3})

	assert.Equal(t, []model.QTC{qtc1, qtc2, qtc3}, loadedQTCs)
}

func TestLogbook_FindDuplicateQSOs_bandRules(t *testing.T) {
	tests := []struct {
		name      string
		bandRule  conval.BandRule
		band      model.Band
		mode      model.Mode
		duplicate bool
	}{
		{
			name:      "once, same band, same mode",
			bandRule:  conval.Once,
			band:      model.Band80m,
			mode:      model.ModeCW,
			duplicate: true,
		},
		{
			name:      "once, same band, different mode",
			bandRule:  conval.Once,
			band:      model.Band80m,
			mode:      model.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once, different band, same mode",
			bandRule:  conval.Once,
			band:      model.Band40m,
			mode:      model.ModeCW,
			duplicate: true,
		},
		{
			name:      "once, different band, different mode",
			bandRule:  conval.Once,
			band:      model.Band40m,
			mode:      model.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once per band, same band, same mode",
			bandRule:  conval.OncePerBand,
			band:      model.Band80m,
			mode:      model.ModeCW,
			duplicate: true,
		},
		{
			name:      "once per band, same band, different mode",
			bandRule:  conval.OncePerBand,
			band:      model.Band80m,
			mode:      model.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once per band, different band, same mode",
			bandRule:  conval.OncePerBand,
			band:      model.Band40m,
			mode:      model.ModeCW,
			duplicate: false,
		},
		{
			name:      "once per band, different band, different mode",
			bandRule:  conval.OncePerBand,
			band:      model.Band40m,
			mode:      model.ModeSSB,
			duplicate: false,
		},
		{
			name:      "once per band and mode, same band, same mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      model.Band80m,
			mode:      model.ModeCW,
			duplicate: true,
		},
		{
			name:      "once per band and mode, same band, different mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      model.Band80m,
			mode:      model.ModeSSB,
			duplicate: false,
		},
		{
			name:      "once per band and mode, different band, same mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      model.Band40m,
			mode:      model.ModeCW,
			duplicate: false,
		},
		{
			name:      "once per band and mode, different band, different mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      model.Band40m,
			mode:      model.ModeSSB,
			duplicate: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dl1abc := model.MustParseCallsign("DL1ABC")
			qso := model.QSO{MyNumber: 1, Callsign: dl1abc, Band: model.Band80m, Mode: model.ModeCW}
			logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
			logbook.bandRule = test.bandRule
			logbook.AddQSO(qso)

			duplicateQSOs := logbook.FindDuplicateQSOs(dl1abc, test.band, test.mode)

			if test.duplicate {
				assert.Equal(t, []model.QSO{qso}, duplicateQSOs)
			} else {
				assert.Empty(t, duplicateQSOs)
			}
		})
	}
}

func TestLogbook_FindWorkedQSOs_bandRules(t *testing.T) {
	tests := []struct {
		name      string
		bandRule  conval.BandRule
		band      model.Band
		mode      model.Mode
		duplicate bool
	}{
		{
			name:      "once, same band, same mode",
			bandRule:  conval.Once,
			band:      model.Band80m,
			mode:      model.ModeCW,
			duplicate: true,
		},
		{
			name:      "once, same band, different mode",
			bandRule:  conval.Once,
			band:      model.Band80m,
			mode:      model.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once, different band, same mode",
			bandRule:  conval.Once,
			band:      model.Band40m,
			mode:      model.ModeCW,
			duplicate: true,
		},
		{
			name:      "once, different band, different mode",
			bandRule:  conval.Once,
			band:      model.Band40m,
			mode:      model.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once per band, same band, same mode",
			bandRule:  conval.OncePerBand,
			band:      model.Band80m,
			mode:      model.ModeCW,
			duplicate: true,
		},
		{
			name:      "once per band, same band, different mode",
			bandRule:  conval.OncePerBand,
			band:      model.Band80m,
			mode:      model.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once per band, different band, same mode",
			bandRule:  conval.OncePerBand,
			band:      model.Band40m,
			mode:      model.ModeCW,
			duplicate: false,
		},
		{
			name:      "once per band, different band, different mode",
			bandRule:  conval.OncePerBand,
			band:      model.Band40m,
			mode:      model.ModeSSB,
			duplicate: false,
		},
		{
			name:      "once per band and mode, same band, same mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      model.Band80m,
			mode:      model.ModeCW,
			duplicate: true,
		},
		{
			name:      "once per band and mode, same band, different mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      model.Band80m,
			mode:      model.ModeSSB,
			duplicate: false,
		},
		{
			name:      "once per band and mode, different band, same mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      model.Band40m,
			mode:      model.ModeCW,
			duplicate: false,
		},
		{
			name:      "once per band and mode, different band, different mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      model.Band40m,
			mode:      model.ModeSSB,
			duplicate: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dl1abc := model.MustParseCallsign("DL1ABC")
			qso := model.QSO{MyNumber: 1, Callsign: dl1abc, Band: model.Band80m, Mode: model.ModeCW}
			logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
			logbook.bandRule = test.bandRule
			logbook.AddQSO(qso)

			duplicateQSOs, duplicate := logbook.FindWorkedQSOs(dl1abc, test.band, test.mode)

			assert.Equal(t, []model.QSO{qso}, duplicateQSOs)
			assert.Equal(t, test.duplicate, duplicate)
		})
	}
}

func TestLogbook_Find(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	dl2abc := model.MustParseCallsign("DL2ABC")
	n1mm := model.MustParseCallsign("N1MM")
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	logbook.bandRule = conval.Once
	logbook.AddQSO(model.QSO{Callsign: dl1abc, MyNumber: 1})
	logbook.AddQSO(model.QSO{Callsign: dl2abc, MyNumber: 3})
	logbook.AddQSO(model.QSO{Callsign: n1mm, MyNumber: 5})

	matches, err := logbook.Find("DL1ABC")
	require.NoError(t, err)
	assert.Len(t, matches, 2, "matches for DL1ABC")
	assert.Equal(t, dl1abc, matches[0].Callsign, "callsign 0 for DL1ABC")
	assert.Equal(t, dl2abc, matches[1].Callsign, "callsign 1 for DL1ABC")

	matches, err = logbook.Find("DL0ABC")
	require.NoError(t, err)
	assert.Len(t, matches, 2, "matches for DL0ABC")
	assert.Equal(t, dl1abc, matches[0].Callsign, "callsign 0 for DL0ABC")
	assert.Equal(t, dl2abc, matches[1].Callsign, "callsign 1 for DL0ABC")

	matches, err = logbook.Find("N1MM")
	require.NoError(t, err)
	assert.Len(t, matches, 1, "matches for N1MM")
	assert.Equal(t, n1mm, matches[0].Callsign, "callsign 0 for N1MM")
}

func TestLogbook_AddQTCSeries(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	header := model.QTCHeader{SeriesNumber: 1, QTCCount: 2}
	now := time.Now()
	series := model.QTCSeries{
		TheirCallsign: dl1abc,
		Header:        header,
		QTCs: []model.QTC{
			{Kind: model.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-2 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000},
			{Kind: model.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-1 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000},
			{Kind: model.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: header, Timestamp: now, Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000},
		},
	}
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)

	logbook.AddQTCSeries(series)

	assert.Equal(t, series.QTCs, logbook.AllQTCs())
}

func TestLogbook_AddQTCSeries_emitsQTCAddedForAllQTCs(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	header := model.QTCHeader{SeriesNumber: 1, QTCCount: 2}
	now := time.Now()
	qtc1 := model.QTC{Kind: model.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-2 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc2 := model.QTC{Kind: model.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-1 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc3 := model.QTC{Kind: model.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: header, Timestamp: now, Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	series := model.QTCSeries{
		TheirCallsign: dl1abc,
		Header:        header,
		QTCs:          []model.QTC{qtc1, qtc2, qtc3},
	}
	addedQTCs := []model.QTC{}
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	logbook.Notify(QTCAddedListenerFunc(func(qtc model.QTC) {
		addedQTCs = append(addedQTCs, qtc)
	}))

	logbook.AddQTCSeries(series)

	assert.Equal(t, []model.QTC{qtc1, qtc2, qtc3}, addedQTCs)
}

func TestLogbook_AddQTCSeries_updatesTheScore(t *testing.T) {
	now := time.Now()
	dl1abc := model.MustParseCallsign("DL1ABC")
	sentHeader := model.QTCHeader{SeriesNumber: 1, QTCCount: 3}
	qtc1 := model.QTC{Kind: model.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: sentHeader, Timestamp: now.Add(-5 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc2 := model.QTC{Kind: model.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: sentHeader, Timestamp: now.Add(-4 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc3 := model.QTC{Kind: model.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: sentHeader, Timestamp: now.Add(-3 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	sentSeries := model.QTCSeries{
		TheirCallsign: dl1abc,
		Header:        sentHeader,
		QTCs:          []model.QTC{qtc1, qtc2, qtc3},
	}
	dl2abc := model.MustParseCallsign("DL2ABC")
	receivedHeader := model.QTCHeader{SeriesNumber: 12, QTCCount: 3}
	qtc4 := model.QTC{Kind: model.ReceivedQTC, QTCNumber: 1, TheirCallsign: dl2abc, Header: receivedHeader, Timestamp: now.Add(-2 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc5 := model.QTC{Kind: model.ReceivedQTC, QTCNumber: 2, TheirCallsign: dl2abc, Header: receivedHeader, Timestamp: now.Add(-1 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc6 := model.QTC{Kind: model.ReceivedQTC, QTCNumber: 3, TheirCallsign: dl2abc, Header: receivedHeader, Timestamp: now, Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	receivedSeries := model.QTCSeries{
		TheirCallsign: dl2abc,
		Header:        receivedHeader,
		QTCs:          []model.QTC{qtc4, qtc5, qtc6},
	}
	logbook := withTestConvalCounter(NewLogbook(clock.Zero(), new(testSettings), testEntity), &testConvalCounter{})
	logbook.AddQTCSeries(sentSeries)
	logbook.AddQTCSeries(receivedSeries)

	assert.Equal(t, 6, logbook.Score().Result().Result())
}

func TestLogbook_AddQTCSeries_emitsScoreChanged(t *testing.T) {
	now := time.Now()
	dl1abc := model.MustParseCallsign("DL1ABC")
	sentHeader := model.QTCHeader{SeriesNumber: 1, QTCCount: 3}
	qtc1 := model.QTC{Kind: model.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: sentHeader, Timestamp: now.Add(-5 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc2 := model.QTC{Kind: model.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: sentHeader, Timestamp: now.Add(-4 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc3 := model.QTC{Kind: model.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: sentHeader, Timestamp: now.Add(-3 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	sentSeries := model.QTCSeries{
		TheirCallsign: dl1abc,
		Header:        sentHeader,
		QTCs:          []model.QTC{qtc1, qtc2, qtc3},
	}
	dl2abc := model.MustParseCallsign("DL2ABC")
	receivedHeader := model.QTCHeader{SeriesNumber: 12, QTCCount: 3}
	qtc4 := model.QTC{Kind: model.ReceivedQTC, QTCNumber: 1, TheirCallsign: dl2abc, Header: receivedHeader, Timestamp: now.Add(-2 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc5 := model.QTC{Kind: model.ReceivedQTC, QTCNumber: 2, TheirCallsign: dl2abc, Header: receivedHeader, Timestamp: now.Add(-1 * time.Minute), Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	qtc6 := model.QTC{Kind: model.ReceivedQTC, QTCNumber: 3, TheirCallsign: dl2abc, Header: receivedHeader, Timestamp: now, Band: model.Band80m, Mode: model.ModeCW, Frequency: 3500000}
	receivedSeries := model.QTCSeries{
		TheirCallsign: dl2abc,
		Header:        receivedHeader,
		QTCs:          []model.QTC{qtc4, qtc5, qtc6},
	}
	updatedScoreCount := 0
	updatedScore := model.Score{}
	logbook := withTestConvalCounter(NewLogbook(clock.Zero(), new(testSettings), testEntity), &testConvalCounter{})
	logbook.Notify(ScoreChangedFunc(func(score model.Score) {
		updatedScoreCount++
		updatedScore = score
	}))
	logbook.AddQTCSeries(sentSeries)
	logbook.AddQTCSeries(receivedSeries)

	assert.Equal(t, 2, updatedScoreCount, "count")
	assert.Equal(t, 6, updatedScore.Result().Result(), "result")
}

func TestLogbook_AvailableFor(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	logbook.qtcsEnabled = true

	available := logbook.AvailableFor(dl1abc)
	assert.Equal(t, 0, available, "fresh log")

	logbook.AddQSO(model.QSO{MyNumber: 1, Callsign: dl1abc})
	available = logbook.AvailableFor(dl1abc)
	assert.Equal(t, 0, available, "own")

	for i := range model.MaxQTCsPerCall {
		theirCall := model.MustParseCallsign(fmt.Sprintf("K%dAB", i+1))
		logbook.AddQSO(model.QSO{MyNumber: model.QSONumber(i + 2), Callsign: theirCall})
		available = logbook.AvailableFor(dl1abc)
		assert.Equal(t, i+1, available, theirCall.String())
	}

	logbook.AddQSO(model.QSO{MyNumber: model.QSONumber(model.MaxQTCsPerCall + 2), Callsign: model.MustParseCallsign("K1MORE")})
	available = logbook.AvailableFor(dl1abc)
	assert.Equal(t, model.MaxQTCsPerCall, available, "one more")
}

func TestLogbook_PrepareFor(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	logbook.qtcsEnabled = true

	qtcs := logbook.PrepareFor(dl1abc, model.MaxQTCsPerCall)
	assert.Equal(t, 0, len(qtcs), "fresh log")

	logbook.AddQSO(model.QSO{MyNumber: 1, Callsign: dl1abc})
	qtcs = logbook.PrepareFor(dl1abc, model.MaxQTCsPerCall)
	assert.Equal(t, 0, len(qtcs), "own")

	for i := range model.MaxQTCsPerCall {
		theirCall := model.MustParseCallsign(fmt.Sprintf("K%dAB", i+1))
		logbook.AddQSO(model.QSO{MyNumber: model.QSONumber(i + 2), Callsign: theirCall})
		qtcs = logbook.PrepareFor(dl1abc, model.MaxQTCsPerCall)
		assert.Equal(t, i+1, len(qtcs), theirCall.String())
	}

	logbook.AddQSO(model.QSO{MyNumber: model.QSONumber(model.MaxQTCsPerCall + 2), Callsign: model.MustParseCallsign("K1MORE")})
	qtcs = logbook.PrepareFor(dl1abc, model.MaxQTCsPerCall)
	assert.Equal(t, model.MaxQTCsPerCall, len(qtcs), "one more")

	qtcs = logbook.PrepareFor(dl1abc, 1)
	assert.Equal(t, 1, len(qtcs), "only one")
}

func TestLogbook_MarksDuplicates(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	dl2abc := model.MustParseCallsign("DL2ABC")
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	logbook.scoreCounter.counter = &testConvalCounter{}
	logbook.scoreCounter.counterFactory = func() convalCounter {
		return &testConvalCounter{}
	}

	logbook.AddQSO(model.QSO{Callsign: dl1abc, MyNumber: 1})
	assert.False(t, logbook.qsos[0].Duplicate, "first qso")

	logbook.AddQSO(model.QSO{Callsign: dl1abc, MyNumber: 3})
	assert.False(t, logbook.qsos[0].Duplicate, "first qso, duplicate")
	assert.True(t, logbook.qsos[1].Duplicate, "duplicate of first qso")

	logbook.UpdateQSO(model.QSO{Callsign: dl2abc, MyNumber: 1})
	assert.False(t, logbook.qsos[0].Duplicate, "first qso, edited")
	assert.False(t, logbook.qsos[1].Duplicate, "second qso, after edit")

	logbook.AddQSO(model.QSO{Callsign: dl1abc, MyNumber: 2})
	assert.False(t, logbook.qsos[0].Duplicate, "first qso, after insert")
	assert.False(t, logbook.qsos[1].Duplicate, "inserted qso")
	assert.True(t, logbook.qsos[2].Duplicate, "second qso, after insert")
}
