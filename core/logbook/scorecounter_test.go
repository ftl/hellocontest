package logbook

import (
	"testing"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestAddQSO_ScoreQSO(t *testing.T) {
	scoreCounter := newScoreCounter(new(testSettings), testEntity)
	scoreCounter.counter = &testConvalCounter{
		scores: map[string]conval.QSOScore{
			"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
			"K3LR":   {Points: 6, Multis: 7, Duplicate: false},
			"DK9ZZ":  {Points: 3, Multis: 4, Duplicate: false},
		},
	}

	scoredQSO := scoreCounter.AddQSO(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1})
	assert.Equal(t, 1, scoredQSO.Points)
	assert.Equal(t, 2, scoredQSO.Multis)

	scoredQSO = scoreCounter.AddQSO(core.QSO{Callsign: callsign.MustParse("K3LR"), MyNumber: 3})
	assert.Equal(t, 6, scoredQSO.Points)
	assert.Equal(t, 7, scoredQSO.Multis)

	scoredQSO = scoreCounter.AddQSO(core.QSO{Callsign: callsign.MustParse("DK9ZZ"), MyNumber: 2})
	assert.Equal(t, 3, scoredQSO.Points)
	assert.Equal(t, 4, scoredQSO.Multis)
}

func TestAddQTC_ScoreQTC(t *testing.T) {
	scoreCounter := newScoreCounter(new(testSettings), testEntity)
	scoreCounter.counter = &testConvalCounter{}
	dl1abc := callsign.MustParse("DL1ABC")
	header := core.QTCHeader{SeriesNumber: 1, QTCCount: 2}
	now := time.Now()

	scoreCounter.AddQTCSeries(core.QTCSeries{
		TheirCallsign: dl1abc,
		Header:        header,
		QTCs: []core.QTC{
			{Kind: core.SentQTC, QSONumber: 1, QTCNumber: 1, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-2 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000},
			{Kind: core.SentQTC, QSONumber: 2, QTCNumber: 2, TheirCallsign: dl1abc, Header: header, Timestamp: now.Add(-1 * time.Minute), Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000},
			{Kind: core.SentQTC, QSONumber: 3, QTCNumber: 3, TheirCallsign: dl1abc, Header: header, Timestamp: now, Band: core.Band80m, Mode: core.ModeCW, Frequency: 3500000},
		},
	})

	assert.Equal(t, 3, scoreCounter.Score().Result().QTCs, "total")
	assert.Equal(t, 3, scoreCounter.Score().ScorePerBand[core.Band80m].QTCs, "score per band")
}
