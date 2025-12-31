package logbook

import (
	"fmt"
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

func TestLogbook_Load_loadsQTCs(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	header := core.QTCHeader{SeriesNumber: 1, QTCCount: 2}
	now := time.Now()
	qtcs := []core.QTC{
		{Kind: core.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-2 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000},
		{Kind: core.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-1 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000},
		{Kind: core.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: header, Timestamp: now, Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000},
	}
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)

	logbook.Load(nil, qtcs)

	assert.Equal(t, qtcs, logbook.AllQTCs())
}

func TestLogbook_Load_emitsQTCAddedForAllQTCs(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	header := core.QTCHeader{SeriesNumber: 1, QTCCount: 2}
	now := time.Now()
	qtc1 := core.QTC{Kind: core.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-2 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000}
	qtc2 := core.QTC{Kind: core.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-1 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000}
	qtc3 := core.QTC{Kind: core.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: header, Timestamp: now, Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000}
	loadedQTCs := []core.QTC{}
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	logbook.Notify(QTCAddedListenerFunc(func(qtc core.QTC) {
		loadedQTCs = append(loadedQTCs, qtc)
	}))

	logbook.Load(nil, []core.QTC{qtc1, qtc2, qtc3})

	assert.Equal(t, []core.QTC{qtc1, qtc2, qtc3}, loadedQTCs)
}

func TestLogbook_FindDuplicateQSOs_bandRules(t *testing.T) {
	tests := []struct {
		name      string
		bandRule  conval.BandRule
		band      core.Band
		mode      core.Mode
		duplicate bool
	}{
		{
			name:      "once, same band, same mode",
			bandRule:  conval.Once,
			band:      core.Band80m,
			mode:      core.ModeCW,
			duplicate: true,
		},
		{
			name:      "once, same band, different mode",
			bandRule:  conval.Once,
			band:      core.Band80m,
			mode:      core.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once, different band, same mode",
			bandRule:  conval.Once,
			band:      core.Band40m,
			mode:      core.ModeCW,
			duplicate: true,
		},
		{
			name:      "once, different band, different mode",
			bandRule:  conval.Once,
			band:      core.Band40m,
			mode:      core.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once per band, same band, same mode",
			bandRule:  conval.OncePerBand,
			band:      core.Band80m,
			mode:      core.ModeCW,
			duplicate: true,
		},
		{
			name:      "once per band, same band, different mode",
			bandRule:  conval.OncePerBand,
			band:      core.Band80m,
			mode:      core.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once per band, different band, same mode",
			bandRule:  conval.OncePerBand,
			band:      core.Band40m,
			mode:      core.ModeCW,
			duplicate: false,
		},
		{
			name:      "once per band, different band, different mode",
			bandRule:  conval.OncePerBand,
			band:      core.Band40m,
			mode:      core.ModeSSB,
			duplicate: false,
		},
		{
			name:      "once per band and mode, same band, same mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      core.Band80m,
			mode:      core.ModeCW,
			duplicate: true,
		},
		{
			name:      "once per band and mode, same band, different mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      core.Band80m,
			mode:      core.ModeSSB,
			duplicate: false,
		},
		{
			name:      "once per band and mode, different band, same mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      core.Band40m,
			mode:      core.ModeCW,
			duplicate: false,
		},
		{
			name:      "once per band and mode, different band, different mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      core.Band40m,
			mode:      core.ModeSSB,
			duplicate: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dl1abc := callsign.MustParse("DL1ABC")
			qso := core.QSO{MyNumber: 1, Callsign: dl1abc, Band: core.Band80m, Mode: core.ModeCW}
			logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
			logbook.bandRule = test.bandRule
			logbook.AddQSO(qso)

			duplicateQSOs := logbook.FindDuplicateQSOs(dl1abc, test.band, test.mode)

			if test.duplicate {
				assert.Equal(t, []core.QSO{qso}, duplicateQSOs)
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
		band      core.Band
		mode      core.Mode
		duplicate bool
	}{
		{
			name:      "once, same band, same mode",
			bandRule:  conval.Once,
			band:      core.Band80m,
			mode:      core.ModeCW,
			duplicate: true,
		},
		{
			name:      "once, same band, different mode",
			bandRule:  conval.Once,
			band:      core.Band80m,
			mode:      core.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once, different band, same mode",
			bandRule:  conval.Once,
			band:      core.Band40m,
			mode:      core.ModeCW,
			duplicate: true,
		},
		{
			name:      "once, different band, different mode",
			bandRule:  conval.Once,
			band:      core.Band40m,
			mode:      core.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once per band, same band, same mode",
			bandRule:  conval.OncePerBand,
			band:      core.Band80m,
			mode:      core.ModeCW,
			duplicate: true,
		},
		{
			name:      "once per band, same band, different mode",
			bandRule:  conval.OncePerBand,
			band:      core.Band80m,
			mode:      core.ModeSSB,
			duplicate: true,
		},
		{
			name:      "once per band, different band, same mode",
			bandRule:  conval.OncePerBand,
			band:      core.Band40m,
			mode:      core.ModeCW,
			duplicate: false,
		},
		{
			name:      "once per band, different band, different mode",
			bandRule:  conval.OncePerBand,
			band:      core.Band40m,
			mode:      core.ModeSSB,
			duplicate: false,
		},
		{
			name:      "once per band and mode, same band, same mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      core.Band80m,
			mode:      core.ModeCW,
			duplicate: true,
		},
		{
			name:      "once per band and mode, same band, different mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      core.Band80m,
			mode:      core.ModeSSB,
			duplicate: false,
		},
		{
			name:      "once per band and mode, different band, same mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      core.Band40m,
			mode:      core.ModeCW,
			duplicate: false,
		},
		{
			name:      "once per band and mode, different band, different mode",
			bandRule:  conval.OncePerBandAndMode,
			band:      core.Band40m,
			mode:      core.ModeSSB,
			duplicate: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dl1abc := callsign.MustParse("DL1ABC")
			qso := core.QSO{MyNumber: 1, Callsign: dl1abc, Band: core.Band80m, Mode: core.ModeCW}
			logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
			logbook.bandRule = test.bandRule
			logbook.AddQSO(qso)

			duplicateQSOs, duplicate := logbook.FindWorkedQSOs(dl1abc, test.band, test.mode)

			assert.Equal(t, []core.QSO{qso}, duplicateQSOs)
			assert.Equal(t, test.duplicate, duplicate)
		})
	}
}

func TestLogbook_Find(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	dl2abc := callsign.MustParse("DL2ABC")
	n1mm := callsign.MustParse("N1MM")
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	logbook.bandRule = conval.Once
	logbook.AddQSO(core.QSO{Callsign: dl1abc, MyNumber: 1})
	logbook.AddQSO(core.QSO{Callsign: dl2abc, MyNumber: 3})
	logbook.AddQSO(core.QSO{Callsign: n1mm, MyNumber: 5})

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
	dl1abc := callsign.MustParse("DL1ABC")
	header := core.QTCHeader{SeriesNumber: 1, QTCCount: 2}
	now := time.Now()
	series := core.QTCSeries{
		TheirCallsign: dl1abc,
		Header:        header,
		QTCs: []core.QTC{
			{Kind: core.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-2 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000},
			{Kind: core.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-1 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000},
			{Kind: core.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: header, Timestamp: now, Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000},
		},
	}
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)

	logbook.AddQTCSeries(series)

	assert.Equal(t, series.QTCs, logbook.AllQTCs())
}

func TestLogbook_AddQTCSeries_emitsQTCAddedForAllQTCs(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	header := core.QTCHeader{SeriesNumber: 1, QTCCount: 2}
	now := time.Now()
	qtc1 := core.QTC{Kind: core.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-2 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000}
	qtc2 := core.QTC{Kind: core.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-1 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000}
	qtc3 := core.QTC{Kind: core.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: header, Timestamp: now, Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000}
	series := core.QTCSeries{
		TheirCallsign: dl1abc,
		Header:        header,
		QTCs:          []core.QTC{qtc1, qtc2, qtc3},
	}
	addedQTCs := []core.QTC{}
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)
	logbook.Notify(QTCAddedListenerFunc(func(qtc core.QTC) {
		addedQTCs = append(addedQTCs, qtc)
	}))

	logbook.AddQTCSeries(series)

	assert.Equal(t, []core.QTC{qtc1, qtc2, qtc3}, addedQTCs)
}

func TestLogbook_AvailableFor(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)

	available := logbook.AvailableFor(dl1abc)
	assert.Equal(t, 0, available, "fresh log")

	logbook.AddQSO(core.QSO{MyNumber: 1, Callsign: dl1abc})
	available = logbook.AvailableFor(dl1abc)
	assert.Equal(t, 0, available, "own")

	for i := range core.MaxQTCsPerCall {
		theirCall := callsign.MustParse(fmt.Sprintf("K%dAB", i+1))
		logbook.AddQSO(core.QSO{MyNumber: core.QSONumber(i + 2), Callsign: theirCall})
		available = logbook.AvailableFor(dl1abc)
		assert.Equal(t, i+1, available, theirCall.String())
	}

	logbook.AddQSO(core.QSO{MyNumber: core.QSONumber(core.MaxQTCsPerCall + 2), Callsign: callsign.MustParse("K1MORE")})
	available = logbook.AvailableFor(dl1abc)
	assert.Equal(t, core.MaxQTCsPerCall, available, "one more")
}

func TestLogbook_PrepareFor(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	logbook := NewLogbook(clock.Zero(), new(testSettings), testEntity)

	qtcs := logbook.PrepareFor(dl1abc, core.MaxQTCsPerCall)
	assert.Equal(t, 0, len(qtcs), "fresh log")

	logbook.AddQSO(core.QSO{MyNumber: 1, Callsign: dl1abc})
	qtcs = logbook.PrepareFor(dl1abc, core.MaxQTCsPerCall)
	assert.Equal(t, 0, len(qtcs), "own")

	for i := range core.MaxQTCsPerCall {
		theirCall := callsign.MustParse(fmt.Sprintf("K%dAB", i+1))
		logbook.AddQSO(core.QSO{MyNumber: core.QSONumber(i + 2), Callsign: theirCall})
		qtcs = logbook.PrepareFor(dl1abc, core.MaxQTCsPerCall)
		assert.Equal(t, i+1, len(qtcs), theirCall.String())
	}

	logbook.AddQSO(core.QSO{MyNumber: core.QSONumber(core.MaxQTCsPerCall + 2), Callsign: callsign.MustParse("K1MORE")})
	qtcs = logbook.PrepareFor(dl1abc, core.MaxQTCsPerCall)
	assert.Equal(t, core.MaxQTCsPerCall, len(qtcs), "one more")

	qtcs = logbook.PrepareFor(dl1abc, 1)
	assert.Equal(t, 1, len(qtcs), "only one")
}
