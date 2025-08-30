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
		{2 * time.Hour, 1*time.Hour - 1*time.Second, 11},
		{2 * time.Hour, 1 * time.Hour, 12},
		{2 * time.Hour, 1*time.Hour + 1*time.Second, 12},
		{2 * time.Hour, 2*time.Hour - 1*time.Second, 23},
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

func TestBandScore_NoMultis(t *testing.T) {
	score := new(BandScore)
	score.AddQSO(QSOScore{Points: 2})
	score.AddQSO(QSOScore{Points: 2})
	score.AddQSO(QSOScore{Points: 2})
	score.AddQSO(QSOScore{Points: 2})
	assert.Equal(t, 8, score.Result())
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

func TestBandmapEntry_OnFrequency(t *testing.T) {
	const frequency Frequency = 7035000
	tt := []struct {
		desc      string
		frequency Frequency
		expected  bool
	}{
		{
			desc:      "same frequency",
			frequency: frequency,
			expected:  true,
		},
		{
			desc:      "lower frequency in proximity",
			frequency: frequency - Frequency(spotFrequencyDeltaThreshold-0.1),
			expected:  true,
		},
		{
			desc:      "higher frequency in proximity",
			frequency: frequency + Frequency(spotFrequencyDeltaThreshold-0.1),
			expected:  true,
		},
		{
			desc:      "frequency to low",
			frequency: frequency - Frequency(spotFrequencyDeltaThreshold+0.1),
			expected:  false,
		},
		{
			desc:      "frequency to high",
			frequency: frequency + Frequency(spotFrequencyDeltaThreshold+0.1),
			expected:  false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			entry := BandmapEntry{
				Call:      callsign.MustParse("dl1abc"),
				Frequency: frequency,
			}

			actual := entry.OnFrequency(tc.frequency)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestParseQTCHeader(t *testing.T) {
	tests := []struct {
		input    string
		expected QTCHeader
		invalid  bool
	}{
		{
			input:   "",
			invalid: true,
		},
		{
			input:   "1/20",
			invalid: true,
		},
		{
			input:   "1",
			invalid: true,
		},
		{
			input:    "1/1",
			expected: QTCHeader{SeriesNumber: 1, QTCCount: 1},
		},
		{
			input:    "1/10",
			expected: QTCHeader{SeriesNumber: 1, QTCCount: 10},
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual, err := ParseQTCHeader(test.input)
			if test.invalid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestParseQTCTime(t *testing.T) {
	tests := []struct {
		input      string
		relativeTo QTCTime
		expected   string
		invalid    bool
	}{
		{
			input:   "",
			invalid: true,
		},
		{
			input:   "12345",
			invalid: true,
		},
		{
			input:   "2806",
			invalid: true,
		},
		{
			input:   "1260",
			invalid: true,
		},
		{
			input:    "1",
			expected: "0001",
		},
		{
			input:    "12",
			expected: "0012",
		},
		{
			input:    "123",
			expected: "0123",
		},
		{
			input:    "1234",
			expected: "1234",
		},
		{
			input:      "1234",
			relativeTo: QTCTime{Hour: 13, Minute: 18},
			expected:   "1234",
		},
		{
			input:      "12",
			relativeTo: QTCTime{Hour: 13, Minute: 18},
			expected:   "1312",
		},
		{
			input:      "1",
			relativeTo: QTCTime{Hour: 13, Minute: 18},
			expected:   "1301",
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual, err := ParseQTCTime(test.input, test.relativeTo)
			if test.invalid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, actual.String())
			}
		})
	}
}
