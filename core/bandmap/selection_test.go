package bandmap

import (
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

func TestSelection_Navigation_EmitsTheFocusedVFO(t *testing.T) {
	tt := []struct {
		desc   string
		action func(*Selection)
	}{
		{desc: "highest value", action: func(s *Selection) { s.SelectHighestValue() }},
		{desc: "nearest", action: func(s *Selection) { s.SelectNearest(3500000) }},
		{desc: "next up", action: func(s *Selection) { s.SelectNextUp(3000000) }},
		{desc: "next down", action: func(s *Selection) { s.SelectNextDown(4000000) }},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			selection, listener, context := setupSelection(t)
			context.targetVFO = core.VFO2
			context.focusedVFO = core.VFO1

			tc.action(selection)

			require.Len(t, listener.selections, 1)
			assert.Equal(t, core.VFO1, listener.selections[0].vfo, "the keyboard actions tune the focused VFO")
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
		{desc: "the next up is the marker", action: func(s *Selection) { s.SelectNextUp(3510000) }, wantMarker: true},
		{desc: "the next up is the spot", action: func(s *Selection) { s.SelectNextUp(3525000) }},
		{desc: "the next down is the marker", action: func(s *Selection) { s.SelectNextDown(3530000) }, wantMarker: true},
		{desc: "the next down is the spot", action: func(s *Selection) { s.SelectNextDown(3540000) }},
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
				assert.Equal(t, core.VFO1, listener.markerSelections[0].vfo, "the navigation tunes the focused VFO")
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
	context.markerNavigation = func(core.BandmapMarker) bool { return false }
	selection.markers.AddNumbered(1, 3520000, core.Band80m, time.Now())

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

	context := &selectionContextStub{}
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
	selectable       core.BandmapFilter
	navigation       core.BandmapFilter
	markerNavigation core.BandmapMarkerFilter
	targetVFO        core.VFOID
	focusedVFO       core.VFOID
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
