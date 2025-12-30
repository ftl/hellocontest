package logbook

import (
	"testing"

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
