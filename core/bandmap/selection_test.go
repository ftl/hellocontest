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

	context := &selectionContextStub{}
	selection := NewSelection(entries, notifier, context)

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
	selectable core.BandmapFilter
	navigation core.BandmapFilter
	targetVFO  core.VFOID
	focusedVFO core.VFOID
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

func (s *selectionContextStub) TargetVFO(core.BandmapEntry) core.VFOID {
	return s.targetVFO
}

func (s *selectionContextStub) FocusedVFO() core.VFOID {
	return s.focusedVFO
}

type selectionCall struct {
	vfo   core.VFOID
	entry core.BandmapEntry
}

type selectionListenerSpy struct {
	selections []selectionCall
}

func (s *selectionListenerSpy) EntrySelected(vfo core.VFOID, entry core.BandmapEntry) {
	s.selections = append(s.selections, selectionCall{vfo: vfo, entry: entry})
}
