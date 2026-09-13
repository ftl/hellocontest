package bandmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestMarkers_AddGivesEveryMarkerAnOwnID(t *testing.T) {
	markers, now := setupMarkers()

	first := markers.AddText("one", 3500000, core.Band80m, now)
	second := markers.AddText("two", 3510000, core.Band80m, now)

	assert.NotEqual(t, first.ID, second.ID)
	assert.Equal(t, 2, markers.Len())
}

func TestMarkers_ThereIsOnlyOneCQMarker(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddText("keep me", 3500000, core.Band80m, now)

	markers.SetCQ(3510000, core.Band80m, now)
	second := markers.SetCQ(3520000, core.Band80m, now.Add(time.Minute))

	all := markers.All()
	require.Len(t, all, 2, "the text marker survives")
	cqMarkers := 0
	for _, marker := range all {
		if marker.Kind == core.CQMarker {
			cqMarkers++
			assert.Equal(t, second.ID, marker.ID, "the last CQ call wins")
			assert.Equal(t, core.Frequency(3520000), marker.Frequency)
		}
	}
	assert.Equal(t, 1, cqMarkers, "there is either none or only one CQ marker")
}

func TestMarkers_CQMarkerAlwaysCarriesTheSameText(t *testing.T) {
	markers, now := setupMarkers()

	marker := markers.SetCQ(3510000, core.Band80m, now)

	assert.Equal(t, core.CQMarker, marker.Kind)
	assert.Equal(t, "CQ", marker.Text)
}

func TestMarkers_NumberedMarkerCarriesItsNumberAsText(t *testing.T) {
	markers, now := setupMarkers()

	marker := markers.AddNumbered(7, 3510000, core.Band80m, now)

	assert.Equal(t, core.NumberedMarker, marker.Kind)
	assert.Equal(t, 7, marker.Number)
	assert.Equal(t, "7", marker.Text, "the spot list shows and sorts the number as the text")
}

func TestMarkers_TextMarkerCarriesItsText(t *testing.T) {
	markers, now := setupMarkers()

	marker := markers.AddText("beacon", 3510000, core.Band80m, now)

	assert.Equal(t, core.TextMarker, marker.Kind)
	assert.Equal(t, "beacon", marker.Text)
	assert.Equal(t, 0, marker.Number, "a text marker has no number")
}

func TestMarkers_Remove(t *testing.T) {
	markers, now := setupMarkers()
	marker := markers.AddText("one", 3500000, core.Band80m, now)
	markers.AddText("two", 3510000, core.Band80m, now)

	markers.Remove(marker.ID)

	require.Len(t, markers.All(), 1)
	assert.Equal(t, "two", markers.All()[0].Text)
}

func TestMarkers_OrderedFiltersByBand(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddText("eighty", 3500000, core.Band80m, now)
	markers.AddText("twenty", 14000000, core.Band20m, now)

	result := markers.Ordered(func(marker core.BandmapMarker) bool {
		return marker.Band == core.Band80m
	}, core.SortSpotsByFrequency, false)

	require.Len(t, result, 1)
	assert.Equal(t, "eighty", result[0].Text)
}

func TestMarkers_OrderedByFrequency(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddText("high", 3520000, core.Band80m, now)
	markers.SetCQ(3510000, core.Band80m, now)
	markers.AddText("low", 3500000, core.Band80m, now)

	assert.Equal(t, []string{"low", "CQ", "high"}, markerTexts(markers, core.SortSpotsByFrequency, false))
	assert.Equal(t, []string{"high", "CQ", "low"}, markerTexts(markers, core.SortSpotsByFrequency, true))
}

func TestMarkers_OrderedByLastSeenUsesTheCreationTime(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddText("second", 3520000, core.Band80m, now.Add(time.Minute))
	markers.AddText("first", 3500000, core.Band80m, now)
	markers.AddText("third", 3510000, core.Band80m, now.Add(2*time.Minute))

	assert.Equal(t, []string{"first", "second", "third"}, markerTexts(markers, core.SortSpotsByLastSeen, false))
	assert.Equal(t, []string{"third", "second", "first"}, markerTexts(markers, core.SortSpotsByLastSeen, true))
}

func TestMarkers_OrderedByTextKeepsTheCQMarkerOnTop(t *testing.T) {
	for _, column := range []core.SpotSortColumn{core.SortSpotsByCallsign, core.SortSpotsByValue} {
		t.Run(column.String(), func(t *testing.T) {
			markers, now := setupMarkers()
			markers.AddText("beacon", 3520000, core.Band80m, now)
			markers.AddText("Alpha", 3530000, core.Band80m, now)
			markers.SetCQ(3540000, core.Band80m, now)

			assert.Equal(t, []string{"CQ", "Alpha", "beacon"}, markerTexts(markers, column, false),
				"the text decides, the comparison ignores the case")
			assert.Equal(t, []string{"CQ", "beacon", "Alpha"}, markerTexts(markers, column, true),
				"the CQ marker keeps the first place in both directions")
		})
	}
}

func TestMarkers_OrderedByTextComparesNumbers(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddNumbered(10, 3520000, core.Band80m, now)
	markers.AddNumbered(2, 3530000, core.Band80m, now)
	markers.AddText("beacon", 3540000, core.Band80m, now)

	assert.Equal(t, []string{"2", "10", "beacon"}, markerTexts(markers, core.SortSpotsByCallsign, false),
		"a numbered marker compares as a number and comes before a text marker")
}

func markerTexts(markers *Markers, column core.SpotSortColumn, descending bool) []string {
	ordered := markers.Ordered(nil, column, descending)
	result := make([]string, len(ordered))
	for i, marker := range ordered {
		result[i] = marker.Text
	}
	return result
}

func setupMarkers() (*Markers, time.Time) {
	return NewMarkers(NewEntries(&Notifier{}, countAllEntries)), time.Now()
}

func TestMarkers_SetCQCreatesTheMarker(t *testing.T) {
	markers, now := setupMarkers()

	marker := markers.SetCQ(3560000, core.Band80m, now)

	require.Len(t, markers.All(), 1)
	assert.Equal(t, core.CQMarker, marker.Kind)
	assert.Equal(t, "CQ", marker.Text)
	assert.Equal(t, core.Frequency(3560000), marker.Frequency)
}

func TestMarkers_SetCQMovesTheExistingMarker(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddText("keep me", 3500000, core.Band80m, now)
	first := markers.SetCQ(3560000, core.Band80m, now)

	later := now.Add(time.Minute)
	second := markers.SetCQ(14060000, core.Band20m, later)

	assert.Equal(t, first.ID, second.ID, "the marker keeps its ID, the spot list moves the row")
	assert.Equal(t, core.Frequency(14060000), second.Frequency)
	assert.Equal(t, core.Band20m, second.Band)
	assert.Equal(t, later, second.CreatedAt)
	require.Len(t, markers.All(), 2, "the generic marker survives")
}

func TestMarkers_Get(t *testing.T) {
	markers, now := setupMarkers()
	marker := markers.AddText("one", 3500000, core.Band80m, now)

	found, ok := markers.Get(marker.ID)
	require.True(t, ok)
	assert.Equal(t, "one", found.Text)

	_, ok = markers.Get(marker.ID + 1)
	assert.False(t, ok, "an unknown ID gives no marker")
}

func TestMarkers_MarkWithNextNumber_UsesTheLowestAvailableNumber(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddNumbered(1, 3500000, core.Band80m, now)
	markers.AddNumbered(2, 3510000, core.Band80m, now)
	markers.AddNumbered(4, 3520000, core.Band80m, now)

	marker := markers.MarkWithNextNumber(3550000, core.Band80m, now)

	assert.Equal(t, 3, marker.Number, "the gap between 2 and 4 is filled first")
	assert.Equal(t, "3", marker.Text)
}

func TestMarkers_MarkWithNextNumber_StartsWithOne(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddText("beacon", 3500000, core.Band80m, now)
	markers.SetCQ(3510000, core.Band80m, now)

	marker := markers.MarkWithNextNumber(3550000, core.Band80m, now)

	assert.Equal(t, 1, marker.Number, "a text marker and the CQ marker use no number")
}

func TestMarkers_MarkWithNextNumber_MovesAMarkerInProximity(t *testing.T) {
	tt := []struct {
		desc     string
		existing func(*Markers, time.Time) core.BandmapMarker
	}{
		{
			desc: "a numbered marker",
			existing: func(m *Markers, now time.Time) core.BandmapMarker {
				return m.AddNumbered(5, 3550000, core.Band80m, now)
			},
		},
		{
			desc: "a text marker",
			existing: func(m *Markers, now time.Time) core.BandmapMarker {
				return m.AddText("beacon", 3550000, core.Band80m, now)
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			markers, now := setupMarkers()
			existing := tc.existing(markers, now)
			later := now.Add(time.Minute)

			marker := markers.MarkWithNextNumber(3550100, core.Band80m, later)

			assert.Equal(t, existing.ID, marker.ID, "the existing marker moves")
			assert.Equal(t, existing.Kind, marker.Kind, "the kind stays")
			assert.Equal(t, existing.Text, marker.Text, "the text stays")
			assert.Equal(t, core.Frequency(3550100), marker.Frequency)
			assert.Equal(t, later, marker.CreatedAt, "an adjustment renews the creation time")
			assert.Len(t, markers.All(), 1, "no second marker on the same signal")
		})
	}
}

func TestMarkers_MarkWithNextNumber_IgnoresTheCQMarkerInProximity(t *testing.T) {
	markers, now := setupMarkers()
	markers.SetCQ(3550000, core.Band80m, now)

	marker := markers.MarkWithNextNumber(3550100, core.Band80m, now)

	assert.Equal(t, core.NumberedMarker, marker.Kind, "the CQ marker stays what it is")
	assert.Len(t, markers.All(), 2)
}

func TestMarkers_MarkWithNumber_MovesTheMarkerWithThatNumber(t *testing.T) {
	markers, now := setupMarkers()
	existing := markers.AddNumbered(3, 3500000, core.Band80m, now)
	later := now.Add(time.Minute)

	marker := markers.MarkWithNumber(3, 14060000, core.Band20m, later)

	assert.Equal(t, existing.ID, marker.ID)
	assert.Equal(t, core.Frequency(14060000), marker.Frequency)
	assert.Equal(t, core.Band20m, marker.Band)
	assert.Equal(t, later, marker.CreatedAt)
	assert.Len(t, markers.All(), 1)
}

func TestMarkers_MarkWithNumber_RenumbersAMarkerInProximity(t *testing.T) {
	markers, now := setupMarkers()
	existing := markers.AddNumbered(5, 3550000, core.Band80m, now)
	later := now.Add(time.Minute)

	marker := markers.MarkWithNumber(3, 3550100, core.Band80m, later)

	assert.Equal(t, existing.ID, marker.ID, "the marker in proximity takes the number over")
	assert.Equal(t, 3, marker.Number)
	assert.Equal(t, "3", marker.Text)
	assert.Equal(t, core.Frequency(3550100), marker.Frequency)
	assert.Equal(t, later, marker.CreatedAt)
	assert.Len(t, markers.All(), 1)
}

func TestMarkers_MarkWithNumber_ReplacesATextMarkerInProximity(t *testing.T) {
	markers, now := setupMarkers()
	existing := markers.AddText("beacon", 3550000, core.Band80m, now)
	later := now.Add(time.Minute)

	marker := markers.MarkWithNumber(3, 3550100, core.Band80m, later)

	assert.Equal(t, existing.ID, marker.ID)
	assert.Equal(t, core.NumberedMarker, marker.Kind, "the text marker becomes a numbered marker")
	assert.Equal(t, 3, marker.Number)
	assert.Equal(t, "3", marker.Text)
	assert.Equal(t, later, marker.CreatedAt)
	assert.Len(t, markers.All(), 1)
}

func TestMarkers_MarkWithNumber_PrefersANumberedMarkerInProximity(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddText("beacon", 3550000, core.Band80m, now)
	numbered := markers.AddNumbered(5, 3550050, core.Band80m, now)

	marker := markers.MarkWithNumber(3, 3550100, core.Band80m, now)

	assert.Equal(t, numbered.ID, marker.ID, "a numbered marker in proximity comes first")
	assert.Len(t, markers.All(), 2, "the text marker survives")
}

func TestMarkers_MarkWithNumber_CreatesAMarker(t *testing.T) {
	markers, now := setupMarkers()

	marker := markers.MarkWithNumber(3, 3550000, core.Band80m, now)

	assert.Equal(t, core.NumberedMarker, marker.Kind)
	assert.Equal(t, 3, marker.Number)
	assert.Len(t, markers.All(), 1)
}

func TestMarkers_MarkWithText_CreatesANewMarker(t *testing.T) {
	markers, now := setupMarkers()

	marker := markers.MarkWithText("beacon", 3550000, core.Band80m, now)

	assert.Equal(t, core.TextMarker, marker.Kind)
	assert.Equal(t, "beacon", marker.Text)
	assert.Equal(t, core.Frequency(3550000), marker.Frequency)
	assert.Len(t, markers.All(), 1)
}

func TestMarkers_MarkWithText_AllowsTheSameTextTwice(t *testing.T) {
	markers, now := setupMarkers()
	markers.MarkWithText("beacon", 3550000, core.Band80m, now)

	marker := markers.MarkWithText("beacon", 14060000, core.Band20m, now)

	assert.Equal(t, "beacon", marker.Text)
	assert.Len(t, markers.All(), 2, "the text of a marker does not have to be unique")
}

func TestMarkers_MarkWithText_ReplacesTheTextOfAMarkerInProximity(t *testing.T) {
	tt := []struct {
		desc     string
		existing func(*Markers, time.Time) core.BandmapMarker
	}{
		{
			desc: "a text marker",
			existing: func(m *Markers, now time.Time) core.BandmapMarker {
				return m.AddText("beacon", 3550000, core.Band80m, now)
			},
		},
		{
			desc: "a numbered marker",
			existing: func(m *Markers, now time.Time) core.BandmapMarker {
				return m.AddNumbered(5, 3550000, core.Band80m, now)
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			markers, now := setupMarkers()
			existing := tc.existing(markers, now)
			later := now.Add(time.Minute)

			marker := markers.MarkWithText("dx", 3550100, core.Band80m, later)

			assert.Equal(t, existing.ID, marker.ID, "the existing marker takes the text")
			assert.Equal(t, core.TextMarker, marker.Kind)
			assert.Equal(t, "dx", marker.Text)
			assert.Equal(t, 0, marker.Number, "a text marker has no number")
			assert.Equal(t, core.Frequency(3550100), marker.Frequency)
			assert.Equal(t, later, marker.CreatedAt, "an adjustment renews the creation time")
			assert.Len(t, markers.All(), 1, "no new marker")
		})
	}
}

func TestMarkers_MarkWithText_FreesTheNumberOfTheConvertedMarker(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddNumbered(1, 3550000, core.Band80m, now)

	markers.MarkWithText("dx", 3550100, core.Band80m, now)
	marker := markers.MarkWithNextNumber(3570000, core.Band80m, now)

	assert.Equal(t, 1, marker.Number, "the number of the converted marker is available again")
}

func TestMarkers_MarkWithText_IgnoresTheCQMarkerInProximity(t *testing.T) {
	markers, now := setupMarkers()
	markers.SetCQ(3550000, core.Band80m, now)

	marker := markers.MarkWithText("dx", 3550100, core.Band80m, now)

	assert.Equal(t, core.TextMarker, marker.Kind, "the CQ marker stays what it is")
	assert.Len(t, markers.All(), 2)
}

func TestMarkers_RemoveOnFrequency(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddNumbered(1, 3500000, core.Band80m, now)
	closest := markers.AddText("beacon", 3550000, core.Band80m, now)
	markers.AddNumbered(2, 3550500, core.Band80m, now)

	marker, found := markers.RemoveOnFrequency(3550100)

	require.True(t, found)
	assert.Equal(t, closest.ID, marker.ID, "the closest marker in proximity is deleted")
	assert.Len(t, markers.All(), 2)
}

func TestMarkers_RemoveOnFrequency_WithoutAMarkerInProximity(t *testing.T) {
	markers, now := setupMarkers()
	markers.AddNumbered(1, 3500000, core.Band80m, now)

	_, found := markers.RemoveOnFrequency(3550000)

	assert.False(t, found, "no marker in proximity, nothing happens")
	assert.Len(t, markers.All(), 1)
}

func TestMarkers_RemoveOnFrequency_KeepsTheCQMarker(t *testing.T) {
	markers, now := setupMarkers()
	markers.SetCQ(3550000, core.Band80m, now)

	_, found := markers.RemoveOnFrequency(3550100)

	assert.False(t, found, "the CQ marker is no numbered or text marker")
	assert.Len(t, markers.All(), 1)
}
