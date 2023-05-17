package core

import (
	"strconv"
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
)

func TestEntryField_ExchangeField(t *testing.T) {
	myExchange := MyExchangeField(1)
	assert.True(t, myExchange.IsMyExchange())

	theirExchange := TheirExchangeField(2)
	assert.True(t, theirExchange.IsTheirExchange())

	assert.False(t, CallsignField.IsMyExchange())
	assert.False(t, CallsignField.IsTheirExchange())
}

func TestEntryField_ExchangeIndex(t *testing.T) {
	myExchange := MyExchangeField(1)
	assert.Equal(t, 1, myExchange.ExchangeIndex())

	theirExchange := TheirExchangeField(2)
	assert.Equal(t, 2, theirExchange.ExchangeIndex())

	assert.Equal(t, -1, CallsignField.ExchangeIndex())
}

func TestEntryField_NextExchangeField(t *testing.T) {
	myExchange := MyExchangeField(1)
	assert.Equal(t, MyExchangeField(2), myExchange.NextExchangeField())

	theirExchange := TheirExchangeField(2)
	assert.Equal(t, TheirExchangeField(3), theirExchange.NextExchangeField())

	assert.Equal(t, EntryField(""), CallsignField.NextExchangeField())
}

func TestBandGraph_Bindex(t *testing.T) {
	tt := []struct {
		duration time.Duration
		value    time.Duration
		expected int
	}{
		{0, 1 * time.Second, 0},
		{2 * time.Hour, -1 * time.Second, -1},
		{2 * time.Hour, 0, 0},
		{2 * time.Hour, 1 * time.Second, 0},
		{2 * time.Hour, 1*time.Hour - 1*time.Second, 29},
		{2 * time.Hour, 1 * time.Hour, 30},
		{2 * time.Hour, 1*time.Hour + 1*time.Second, 30},
		{2 * time.Hour, 2*time.Hour - 1*time.Second, 59},
		{2 * time.Hour, 2 * time.Hour, -1},
		{2 * time.Hour, 2*time.Hour + 1*time.Second, -1},
	}
	startTime := time.Now()
	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			graph := NewBandGraph(NoBand, startTime, tc.duration)
			actual := graph.Bindex(startTime.Add(tc.value))
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestScore_StackedGraphPerBand(t *testing.T) {
	score := NewScore()
	score.GraphPerBand[Band160m] = BandGraph{
		Band:       Band160m,
		DataPoints: []BandScore{{16, 16, 16, 16}, {15, 15, 15, 15}, {14, 14, 14, 14}},
	}
	score.GraphPerBand[Band80m] = BandGraph{
		Band:       Band80m,
		DataPoints: []BandScore{{8, 8, 8, 8}, {7, 7, 7, 7}, {6, 6, 6, 6}},
	}
	score.GraphPerBand[Band40m] = BandGraph{
		Band:       Band40m,
		DataPoints: []BandScore{{4, 4, 4, 4}, {3, 3, 3, 3}, {2, 2, 2, 2}},
	}

	stackedGraphs := score.StackedGraphPerBand()

	assert.Equal(t, 3, len(stackedGraphs))

	assert.Equal(t, 16, stackedGraphs[0].DataPoints[0].QSOs)
	assert.Equal(t, 24, stackedGraphs[1].DataPoints[0].QSOs)
	assert.Equal(t, 28, stackedGraphs[2].DataPoints[0].QSOs)
}

func TestBandmapEntry_ProximityFactor(t *testing.T) {
	const frequency Frequency = 7035000
	tt := []struct {
		desc      string
		frequency Frequency
		expected  float64
	}{
		{
			desc:      "same frequency",
			frequency: frequency,
			expected:  1.0,
		},
		{
			desc:      "lower frequency in proximity",
			frequency: frequency - Frequency(spotFrequencyProximityThreshold/2),
			expected:  0.5,
		},
		{
			desc:      "higher frequency in proximity",
			frequency: frequency + Frequency(spotFrequencyProximityThreshold/2),
			expected:  -0.5,
		},
		{
			desc:      "frequency to low",
			frequency: frequency - Frequency(spotFrequencyProximityThreshold) - 1,
			expected:  0.0,
		},
		{
			desc:      "frequency to high",
			frequency: frequency + Frequency(spotFrequencyProximityThreshold) + 1,
			expected:  0.0,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			entry := BandmapEntry{
				Call:      callsign.MustParse("dl1abc"),
				Frequency: frequency,
			}

			actual := entry.ProximityFactor(tc.frequency)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
