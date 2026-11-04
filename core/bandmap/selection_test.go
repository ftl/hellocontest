package bandmap

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestSelection_SelectEntry_EmitsTheTargetVFO(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.targetVFO = core.VFO2
	context.focusedVFO = core.VFO1

	selection.SelectEntry(onlyEntryID(t, selection))

	require.Len(t, listener.selections, 1)
	assert.Equal(t, core.VFO2, listener.selections[0].vfo, "the target VFO works the spot")
}

func TestSelection_SelectByCallsign_EmitsTheTargetVFO(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.targetVFO = core.VFO2
	context.focusedVFO = core.VFO1

	selection.SelectByCallsign(core.MustParseCallsign("dl1abc"))

	require.Len(t, listener.selections, 1)
	assert.Equal(t, core.VFO2, listener.selections[0].vfo)
}

func TestSelection_RowNavigation_EmitsTheTargetVFO(t *testing.T) {
	tt := []struct {
		desc   string
		action func(*Selection)
	}{
		{desc: "next up", action: func(s *Selection) { s.SelectNextUp(3000000) }},
		{desc: "next down", action: func(s *Selection) { s.SelectNextDown(4000000) }},
		{desc: "highest value", action: func(s *Selection) { s.SelectHighestValue() }},
		{desc: "nearest", action: func(s *Selection) { s.SelectNearest(3500000) }},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			selection, listener, context := setupSelection(t)
			context.targetVFO = core.VFO2
			context.focusedVFO = core.VFO1

			tc.action(selection)

			require.Len(t, listener.selections, 1)
			assert.Equal(t, core.VFO2, listener.selections[0].vfo, "the same rules apply as when clicking the row")
		})
	}
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
			context.focusedVFO = core.VFO1
			// the spot is on 3535000
			selection.markers.AddNumbered(1, 3520000, core.Band80m, time.Now())

			tc.action(selection)

			if tc.wantMarker {
				require.Len(t, listener.markerSelections, 1, "the navigation stops at the marker")
				assert.Equal(t, core.Frequency(3520000), listener.markerSelections[0].marker.Frequency)
				assert.Equal(t, core.VFO1, listener.markerSelections[0].vfo, "the target VFO of the marker works it")
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
	context.focusedVFO = core.VFO1
	selection.markers.AddNumbered(1, 3520000, core.Band80m, time.Now())
	// the table shows only the spot, therefore the marker is no navigation target
	context.rows = []core.BandmapRow{testRow(t, selection, "dl9zzz", 3535000, core.Band80m)}

	selection.SelectNearest(3525000)

	assert.Empty(t, listener.markerSelections)
	require.Len(t, listener.selections, 1, "the navigation continues with the spots")
}

func TestSelection_SelectNearest_SkipsAMarkerOnTheFrequency(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = core.VFO1
	selection.markers.AddNumbered(1, 3520000, core.Band80m, time.Now())

	selection.SelectNearest(3520000)

	assert.Empty(t, listener.markerSelections, "the operator is already on the marker")
	require.Len(t, listener.selections, 1)
}

func TestSelection_SelectEntry_IgnoresAnInvisibleEntry(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.selectable = func(core.BandmapEntry) bool { return false }

	selection.SelectEntry(onlyEntryID(t, selection))

	assert.Empty(t, listener.selections)
}

func setupSelection(t *testing.T) (*Selection, *selectionListenerSpy, *selectionContextStub) {
	t.Helper()
	notifier := &Notifier{asyncRunner: func(f func()) { f() }}
	entries := NewEntries(notifier, countAllEntries)
	entries.Add(core.Spot{Call: core.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: core.Band80m, Time: time.Now()}, time.Now(), defaultWeights)

	markers := NewMarkers(entries)

	context := &selectionContextStub{entries: entries, markers: markers}
	selection := NewSelection(entries, markers, notifier, context)

	listener := new(selectionListenerSpy)
	notifier.Notify(listener)

	return selection, listener, context
}

func onlyEntryID(t *testing.T, selection *Selection) core.BandmapEntryID {
	t.Helper()
	var result core.BandmapEntryID
	found := false
	selection.entries.ForEach(func(entry core.BandmapEntry) bool {
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
	rows             []core.BandmapRow
	selectable       core.BandmapFilter
	navigation       core.BandmapFilter
	markerNavigation core.BandmapMarkerFilter
	targetVFO        core.VFOID
	focusedVFO       core.VFOID
}

// VisibleRows returns the rows of the spots table. Without an explicit setup it returns all
// markers and entries ordered by frequency, like the default sorting of the table.
func (s *selectionContextStub) VisibleRows() []core.BandmapRow {
	if s.rows != nil {
		return s.rows
	}

	result := make([]core.BandmapRow, 0)
	for _, marker := range s.markers.Ordered(func(core.BandmapMarker) bool { return true }, core.SortSpotsByFrequency, false) {
		result = append(result, core.MarkerBandmapRow(marker))
	}
	for _, entry := range s.entries.Query(core.BandmapByFrequency) {
		result = append(result, core.SpotBandmapRow(entry))
	}
	slices.SortStableFunc(result, func(a, b core.BandmapRow) int {
		return core.Compare(a.Frequency(), b.Frequency())
	})
	return result
}

func (s *selectionContextStub) SelectableFilter() core.BandmapFilter {
	if s.selectable == nil {
		return func(core.BandmapEntry) bool { return true }
	}
	return s.selectable
}

func (s *selectionContextStub) NavigationFilter() core.BandmapFilter {
	if s.navigation == nil {
		return func(core.BandmapEntry) bool { return true }
	}
	return s.navigation
}

func (s *selectionContextStub) MarkerNavigationFilter() core.BandmapMarkerFilter {
	if s.markerNavigation == nil {
		return func(core.BandmapMarker) bool { return true }
	}
	return s.markerNavigation
}

func (s *selectionContextStub) TargetVFO(core.BandmapEntry) core.VFOID {
	return s.targetVFO
}

func (s *selectionContextStub) TargetVFOForMarker(core.Band) core.VFOID {
	return s.targetVFO
}

func (s *selectionContextStub) FocusedVFO() core.VFOID {
	return s.focusedVFO
}

type selectionCall struct {
	vfo   core.VFOID
	entry core.BandmapEntry
}

type markerSelectionCall struct {
	vfo    core.VFOID
	marker core.BandmapMarker
}

type selectionListenerSpy struct {
	selections       []selectionCall
	markerSelections []markerSelectionCall
}

func (s *selectionListenerSpy) EntrySelected(vfo core.VFOID, entry core.BandmapEntry) {
	s.selections = append(s.selections, selectionCall{vfo: vfo, entry: entry})
}

func (s *selectionListenerSpy) MarkerSelected(vfo core.VFOID, marker core.BandmapMarker) {
	s.markerSelections = append(s.markerSelections, markerSelectionCall{vfo: vfo, marker: marker})
}

func TestSelection_RowNavigation_FollowsTheOrderOfTheTable(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = core.VFO1
	// the table sorts by value, therefore the rows are in no frequency order
	context.rows = []core.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, core.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, core.Band40m),
		testRow(t, selection, "dl3qrs", 21300000, core.Band15m),
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
	context.focusedVFO = core.VFO1
	context.rows = []core.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, core.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, core.Band40m),
	}

	selection.SelectNextUp(7016000)

	require.Len(t, listener.selections, 1)
	assert.Equal(t, "DL2XYZ", listener.selections[0].entry.Call.String())
}

func TestSelection_RowNavigation_StopsAtTheEndsOfTheTable(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = core.VFO1
	context.rows = []core.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, core.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, core.Band40m),
	}

	selection.SelectEntry(context.rows[0].Entry.ID)
	selection.SelectNextUp(0)

	assert.Len(t, listener.selections, 1, "the first row is the end of the table")

	selection.SelectEntry(context.rows[1].Entry.ID)
	selection.SelectNextDown(0)

	assert.Len(t, listener.selections, 2, "the last row is the end of the table")
}

func testRow(t *testing.T, selection *Selection, call string, frequency core.Frequency, band core.Band) core.BandmapRow {
	t.Helper()
	selection.entries.Add(core.Spot{
		Call:      core.MustParseCallsign(call),
		Frequency: frequency,
		Band:      band,
		Time:      time.Now(),
	}, time.Now(), defaultWeights)

	var result core.BandmapRow
	found := false
	selection.entries.ForEach(func(entry core.BandmapEntry) bool {
		if entry.Call.String() != strings.ToUpper(call) {
			return false
		}
		result = core.SpotBandmapRow(entry)
		found = true
		return true
	})
	require.True(t, found, "the scenario needs the entry of %s", call)
	return result
}

func TestSelection_SelectHighestValue_TakesTheSpotsOfEveryBand(t *testing.T) {
	selection, listener, context := setupSelection(t)
	context.focusedVFO = core.VFO1
	rows := []core.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, core.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, core.Band40m),
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
	context.focusedVFO = core.VFO1
	rows := []core.BandmapRow{
		testRow(t, selection, "dl1abc", 14020000, core.Band20m),
		testRow(t, selection, "dl2xyz", 7015000, core.Band40m),
	}
	rows[0].Entry.Info.WeightedValue = 10
	rows[1].Entry.Info.WeightedValue = 20
	rows[1].Entry.Source = core.WorkedSpot
	context.rows = rows

	selection.SelectHighestValue()

	require.Len(t, listener.selections, 1)
	assert.Equal(t, "DL1ABC", listener.selections[0].entry.Call.String(), "a worked spot is no goal")
}
