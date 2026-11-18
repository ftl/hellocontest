package bandmap

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core/model"
)

func TestSelection_SelectEntry_EmitsTheTargetVFO(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.targetVFO = model.VFO2
	context.focusedVFO = model.VFO1

	selection.SelectEntry(onlyEntryID(t, selection))

	require.Len(t, listener.selections, 1)
	assert.Equal(t, model.VFO2, listener.selections[0].vfo, "the target VFO works the spot")
}

func TestSelection_SelectByCallsign_EmitsTheTargetVFO(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.targetVFO = model.VFO2
	context.focusedVFO = model.VFO1

	selection.SelectByCallsign(model.MustParseCallsign("dl1abc"))

	require.Len(t, listener.selections, 1)
	assert.Equal(t, model.VFO2, listener.selections[0].vfo)
}

func TestSelection_RowNavigation_EmitsTheTargetVFO(t *testing.T) {
	tt := []struct {
		desc   string
		action func(*Selection)
	}{
		{desc: "next up", action: func(s *Selection) { s.SelectNextUp(3000000) }},
		{desc: "next down", action: func(s *Selection) { s.SelectNextDown(4000000) }},
		{desc: "highest value", action: func(s *Selection) { s.SelectHighestValue() }},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			selection, listener, context := setupSelection(t)
			context.targetVFO = model.VFO2
			context.focusedVFO = model.VFO1

			tc.action(selection)

			require.Len(t, listener.selections, 1)
			assert.Equal(t, model.VFO2, listener.selections[0].vfo, "the same rules apply as when clicking the row")
		})
	}
}

func TestSelection_SelectNearest_EmitsTheFocusedVFO(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.targetVFO = model.VFO2
	context.focusedVFO = model.VFO1

	selection.SelectNearest(3500000)

	require.Len(t, listener.selections, 1)
	assert.Equal(t, model.VFO1, listener.selections[0].vfo, "the operator searches from the frequency of the focused VFO, therefore that VFO goes to the spot")
}

func TestSelection_Navigation_IncludesTheMarkers(t *testing.T) {
	tt := []struct {
		desc       string
		action     func(*Selection)
		wantMarker bool
	}{
		{desc: "the nearest is the marker", action: func(s *Selection) { s.SelectNearest(3525000) }, wantMarker: true},
		{desc: "the nearest is the spot", action: func(s *Selection) { s.SelectNearest(3533000) }},
		{desc: "the row nearest to the marker", action: func(s *Selection) { s.SelectNextUp(3510000) }, wantMarker: true},
		{desc: "the row nearest to the spot", action: func(s *Selection) { s.SelectNextDown(3540000) }},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			selection, listener, context := setupSelection(t)
			context.focusedVFO = model.VFO1
			// the spot is on 3535000
			selection.markers.AddNumbered(1, 3520000, model.Band80m, time.Now())

			tc.action(selection)

			if tc.wantMarker {
				require.Len(t, listener.markerSelections, 1, "the navigation stops at the marker")
				assert.Equal(t, model.Frequency(3520000), listener.markerSelections[0].marker.Frequency)
				assert.Equal(t, model.VFO1, listener.markerSelections[0].vfo, "the target VFO of the marker works it")
				assert.Empty(t, listener.selections)
			} else {
				require.Len(t, listener.selections, 1, "the spot is closer than the marker")
				assert.Empty(t, listener.markerSelections)
			}
		})
	}
}

func TestSelection_Navigation_IgnoresAnInvisibleMarker(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = model.VFO1
	selection.markers.AddNumbered(1, 3520000, model.Band80m, time.Now())
	// the table shows only the spot, therefore the marker is no navigation target
	context.rows = []model.BandmapRow{testRow(t, selection, "dl9zzz", 3535000, model.Band80m)}

	selection.SelectNearest(3525000)

	assert.Empty(t, listener.markerSelections)
	require.Len(t, listener.selections, 1, "the navigation continues with the spots")
}

func TestSelection_SelectNearest_SkipsAMarkerOnTheFrequency(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = model.VFO1
	selection.markers.AddNumbered(1, 3520000, model.Band80m, time.Now())

	selection.SelectNearest(3520000)

	assert.Empty(t, listener.markerSelections, "the operator is already on the marker")
	require.Len(t, listener.selections, 1)
}

func TestSelection_SelectEntry_IgnoresAnInvisibleEntry(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.selectable = func(model.BandmapEntry) bool { return false }

	selection.SelectEntry(onlyEntryID(t, selection))

	assert.Empty(t, listener.selections)
}

func setupSelection(t *testing.T) (*Selection, *selectionListenerSpy, *selectionContextStub) {
	t.Helper()
	notifier := &Notifier{asyncRunner: func(f func()) { f() }}
	entries := NewEntries(notifier, countAllEntries)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Time: time.Now()}, time.Now(), defaultWeights)

	markers := NewMarkers(entries)

	context := &selectionContextStub{entries: entries, markers: markers}
	selection := NewSelection(entries, markers, notifier, context)

	listener := new(selectionListenerSpy)
	notifier.Notify(listener)

	return selection, listener, context
}

func onlyEntryID(t *testing.T, selection *Selection) model.BandmapEntryID {
	t.Helper()
	var result model.BandmapEntryID
	found := false
	selection.entries.ForEach(func(entry model.BandmapEntry) bool {
		result = entry.ID
		found = true
		return true
	})
	require.True(t, found, "the scenario needs one entry")
	return result
}

type selectionContextStub struct {
	entries          *Entries
	markers          *Markers
	rows             []model.BandmapRow
	selectable       model.BandmapFilter
	navigation       model.BandmapFilter
	markerNavigation model.BandmapMarkerFilter
	targetVFO        model.VFOID
	focusedVFO       model.VFOID
}

// VisibleRows returns the rows of the spots table. Without an explicit setup it returns all
// markers and entries ordered by frequency, like the default sorting of the table.
func (s *selectionContextStub) VisibleRows() []model.BandmapRow {
	if s.rows != nil {
		return s.rows
	}

	result := make([]model.BandmapRow, 0)
	for _, marker := range s.markers.Ordered(func(model.BandmapMarker) bool { return true }, model.SortSpotsByFrequency, false) {
		result = append(result, model.MarkerBandmapRow(marker))
	}
	for _, entry := range s.entries.Query(model.BandmapByFrequency) {
		result = append(result, model.SpotBandmapRow(entry))
	}
	slices.SortStableFunc(result, func(a, b model.BandmapRow) int {
		return model.Compare(a.Frequency(), b.Frequency())
	})
	return result
}

func (s *selectionContextStub) SelectableFilter() model.BandmapFilter {
	if s.selectable == nil {
		return func(model.BandmapEntry) bool { return true }
	}
	return s.selectable
}

func (s *selectionContextStub) NavigationFilter() model.BandmapFilter {
	if s.navigation == nil {
		return func(model.BandmapEntry) bool { return true }
	}
	return s.navigation
}

func (s *selectionContextStub) MarkerNavigationFilter() model.BandmapMarkerFilter {
	if s.markerNavigation == nil {
		return func(model.BandmapMarker) bool { return true }
	}
	return s.markerNavigation
}

func (s *selectionContextStub) TargetVFO(model.BandmapEntry) model.VFOID {
	return s.targetVFO
}

func (s *selectionContextStub) TargetVFOForMarker(model.Band) model.VFOID {
	return s.targetVFO
}

func (s *selectionContextStub) FocusedVFO() model.VFOID {
	return s.focusedVFO
}

type selectionCall struct {
	vfo   model.VFOID
	entry model.BandmapEntry
}

type markerSelectionCall struct {
	vfo    model.VFOID
	marker model.BandmapMarker
}

type selectionListenerSpy struct {
	selections       []selectionCall
	markerSelections []markerSelectionCall
}

func (s *selectionListenerSpy) EntrySelected(vfo model.VFOID, entry model.BandmapEntry) {
	s.selections = append(s.selections, selectionCall{vfo: vfo, entry: entry})
}

func (s *selectionListenerSpy) MarkerSelected(vfo model.VFOID, marker model.BandmapMarker) {
	s.markerSelections = append(s.markerSelections, markerSelectionCall{vfo: vfo, marker: marker})
}

func TestSelection_RowNavigation_FollowsTheOrderOfTheTable(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = model.VFO1
	// the table sorts by value, therefore the rows are in no frequency order
	context.rows = []model.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, model.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, model.Band40m),
		testRow(t, selection, "dl3qrs", 21300000, model.Band15m),
	}

	selection.SelectEntry(context.rows[1].Entry.ID)
	selection.SelectNextUp(0)

	require.Len(t, listener.selections, 2)
	assert.Equal(t, "DL1ABC", listener.selections[1].entry.Call.String(), "up goes to the row above, on another band")

	selection.SelectNextDown(0)

	require.Len(t, listener.selections, 3)
	assert.Equal(t, "DL2XYZ", listener.selections[2].entry.Call.String(), "down goes to the row below")
}

func TestSelection_RowNavigation_StartsAtTheRowNearestToTheVFO(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = model.VFO1
	context.rows = []model.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, model.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, model.Band40m),
	}

	selection.SelectNextUp(7016000)

	require.Len(t, listener.selections, 1)
	assert.Equal(t, "DL2XYZ", listener.selections[0].entry.Call.String())
}

func TestSelection_RowNavigation_StopsAtTheEndsOfTheTable(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = model.VFO1
	context.rows = []model.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, model.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, model.Band40m),
	}

	selection.SelectEntry(context.rows[0].Entry.ID)
	selection.SelectNextUp(0)

	assert.Len(t, listener.selections, 1, "the first row is the end of the table")

	selection.SelectEntry(context.rows[1].Entry.ID)
	selection.SelectNextDown(0)

	assert.Len(t, listener.selections, 2, "the last row is the end of the table")
}

func testRow(t *testing.T, selection *Selection, call string, frequency model.Frequency, band model.Band) model.BandmapRow {
	t.Helper()
	selection.entries.Add(model.Spot{
		Call:      model.MustParseCallsign(call),
		Frequency: frequency,
		Band:      band,
		Time:      time.Now(),
	}, time.Now(), defaultWeights)

	var result model.BandmapRow
	found := false
	selection.entries.ForEach(func(entry model.BandmapEntry) bool {
		if entry.Call.String() != strings.ToUpper(call) {
			return false
		}
		result = model.SpotBandmapRow(entry)
		found = true
		return true
	})
	require.True(t, found, "the scenario needs the entry of %s", call)
	return result
}

func TestSelection_SelectHighestValue_TakesTheSpotsOfEveryBand(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = model.VFO1
	rows := []model.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, model.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, model.Band40m),
	}
	rows[0].Entry.Info.WeightedValue = 10
	rows[1].Entry.Info.WeightedValue = 20
	context.rows = rows

	selection.SelectHighestValue()

	require.Len(t, listener.selections, 1)
	assert.Equal(t, "DL2XYZ", listener.selections[0].entry.Call.String(), "the spot on the other band has the highest value")
}

func TestSelection_SelectHighestValue_IgnoresTheWorkedSpots(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = model.VFO1
	rows := []model.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, model.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, model.Band40m),
	}
	rows[0].Entry.Info.WeightedValue = 10
	rows[1].Entry.Info.WeightedValue = 20
	rows[1].Entry.Source = model.WorkedSpot
	context.rows = rows

	selection.SelectHighestValue()

	require.Len(t, listener.selections, 1)
	assert.Equal(t, "DL1ABC", listener.selections[0].entry.Call.String(), "a worked spot is no goal")
}
