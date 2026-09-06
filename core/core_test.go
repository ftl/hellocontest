package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDescending_InvertsTheOrder(t *testing.T) {
	low := BandmapEntry{ID: 1, Info: Callinfo{WeightedValue: 1}}
	high := BandmapEntry{ID: 2, Info: Callinfo{WeightedValue: 9}}

	assert.Negative(t, BandmapByValue(low, high), "ascending puts the low value first")
	assert.Positive(t, Descending(BandmapByValue)(low, high), "descending puts the high value first")
	assert.Negative(t, Descending(BandmapByValue)(high, low), "descending puts the high value first")
}

func TestDescending_InvertsEqualValuesByID(t *testing.T) {
	first := BandmapEntry{ID: 1, Info: Callinfo{WeightedValue: 5}}
	second := BandmapEntry{ID: 2, Info: Callinfo{WeightedValue: 5}}

	assert.Negative(t, BandmapByValue(first, second))
	assert.Positive(t, Descending(BandmapByValue)(first, second))
}

func TestBandmapByCallsign(t *testing.T) {
	a := BandmapEntry{ID: 1, Call: MustParseCallsign("dl1abc")}
	b := BandmapEntry{ID: 2, Call: MustParseCallsign("dl2abc")}
	sameAsA := BandmapEntry{ID: 3, Call: MustParseCallsign("dl1abc")}

	assert.Negative(t, BandmapByCallsign(a, b), "DL1ABC before DL2ABC")
	assert.Positive(t, BandmapByCallsign(b, a), "DL2ABC after DL1ABC")
	assert.Negative(t, BandmapByCallsign(a, sameAsA), "the same call falls back to the ID")
}

func TestBandmapByLastSeen(t *testing.T) {
	now := time.Now()
	older := BandmapEntry{ID: 1, LastHeard: now.Add(-time.Minute)}
	newer := BandmapEntry{ID: 2, LastHeard: now}
	sameAsOlder := BandmapEntry{ID: 3, LastHeard: now.Add(-time.Minute)}

	assert.Negative(t, BandmapByLastSeen(older, newer), "the older entry comes first")
	assert.Positive(t, BandmapByLastSeen(newer, older), "the newer entry comes last")
	assert.Negative(t, BandmapByLastSeen(older, sameAsOlder), "the same time falls back to the ID")
}

func TestSpotFilterBand_StringRoundTrip(t *testing.T) {
	tt := []SpotFilterBand{
		{Kind: SpotFilterAll},
		{Kind: SpotFilterVFO1},
		{Kind: SpotFilterVFO2},
		{Kind: SpotFilterFocused},
		{Kind: SpotFilterContest},
		FixedSpotFilterBand(Band20m),
	}
	for _, band := range tt {
		t.Run(band.String(), func(t *testing.T) {
			assert.Equal(t, band, ParseSpotFilterBand(band.String()))
		})
	}
}

func TestSpotFilterMode_StringRoundTrip(t *testing.T) {
	tt := []SpotFilterMode{
		{Kind: SpotFilterAll},
		{Kind: SpotFilterVFO1},
		{Kind: SpotFilterVFO2},
		{Kind: SpotFilterFocused},
		{Kind: SpotFilterContest},
		FixedSpotFilterMode(ModeCW),
	}
	for _, mode := range tt {
		t.Run(mode.String(), func(t *testing.T) {
			assert.Equal(t, mode, ParseSpotFilterMode(mode.String()))
		})
	}
}

func TestSpotSortColumn_StringRoundTrip(t *testing.T) {
	for _, column := range SpotSortColumns {
		t.Run(column.String(), func(t *testing.T) {
			assert.Equal(t, column, ParseSpotSortColumn(column.String()))
		})
	}
}

func TestParseSpotFilter_UnknownValueFallsBackToAll(t *testing.T) {
	assert.Equal(t, SpotFilterBand{Kind: SpotFilterAll}, ParseSpotFilterBand("nonsense"))
	assert.Equal(t, SpotFilterMode{Kind: SpotFilterAll}, ParseSpotFilterMode("nonsense"))
	assert.Equal(t, SortSpotsByFrequency, ParseSpotSortColumn("nonsense"))
}
