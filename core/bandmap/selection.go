package bandmap

import (
	"slices"

	"github.com/ftl/hellocontest/core"
)

type SelectionContext interface {
	SelectableFilter() core.BandmapFilter
	NavigationFilter() core.BandmapFilter
	MarkerNavigationFilter() core.BandmapMarkerFilter
	TargetVFO(core.BandmapEntry) core.VFOID
	TargetVFOForMarker(core.Band) core.VFOID
	FocusedVFO() core.VFOID
}

type Selection struct {
	selectedRow core.BandmapRow
	selected    bool

	entries  *Entries
	markers  *Markers
	notifier *Notifier
	context  SelectionContext
}

func NewSelection(entries *Entries, markers *Markers, notifier *Notifier, context SelectionContext) *Selection {
	return &Selection{
		entries:  entries,
		markers:  markers,
		notifier: notifier,
		context:  context,
	}
}

func (s *Selection) selectEntry(vfo core.VFOID, entry core.BandmapEntry) {
	s.selectedRow = core.SpotBandmapRow(entry)
	s.selected = true
	s.notifier.emitEntrySelected(vfo, entry)
}

func (s *Selection) selectMarker(vfo core.VFOID, marker core.BandmapMarker) {
	s.selectedRow = core.MarkerBandmapRow(marker)
	s.selected = true
	s.notifier.emitMarkerSelected(vfo, marker)
}

func (s *Selection) selectRow(vfo core.VFOID, row core.BandmapRow) {
	if row.Kind == core.MarkerRow {
		s.selectMarker(vfo, row.Marker)
		return
	}
	s.selectEntry(vfo, row.Entry)
}

func (s *Selection) clear() {
	s.selectedRow = core.BandmapRow{}
	s.selected = false
	// TODO??? s.notifier.emitEntrySelected(s.selectedRow.Entry)
}

func (s *Selection) findAndSelect(order core.BandmapOrder, filters ...core.BandmapFilter) {
	entries := s.entries.Query(order, filters...)
	if len(entries) > 0 {
		s.selectEntry(s.context.FocusedVFO(), entries[0])
	}
}

// findAndSelectRow navigates through the spots and the markers together, therefore the
// operator reaches a marker with the same actions as a spot.
func (s *Selection) findAndSelectRow(frequency core.Frequency, descendingID bool, matches func(core.BandmapRow) bool) {
	rows := make([]core.BandmapRow, 0)
	for _, entry := range s.entries.Query(core.BandmapByFrequency, s.context.NavigationFilter()) {
		row := core.SpotBandmapRow(entry)
		if matches(row) {
			rows = append(rows, row)
		}
	}
	for _, marker := range s.markers.Ordered(s.context.MarkerNavigationFilter(), core.SortSpotsByFrequency, false) {
		row := core.MarkerBandmapRow(marker)
		if matches(row) {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return
	}

	slices.SortStableFunc(rows, func(a, b core.BandmapRow) int {
		result := core.Compare(distance(a, frequency), distance(b, frequency))
		if result != 0 {
			return result
		}
		return applyDirection(core.Compare(a.ID(), b.ID()), descendingID)
	})

	s.selectRow(s.context.FocusedVFO(), rows[0])
}

func distance(row core.BandmapRow, frequency core.Frequency) core.Frequency {
	result := row.Frequency() - frequency
	if result < 0 {
		return -result
	}
	return result
}

func (s *Selection) SelectedEntry() (core.BandmapEntry, bool) {
	if s.selectedRow.Kind != core.SpotRow {
		return core.BandmapEntry{}, false
	}
	return s.selectedRow.Entry, s.selected
}

func (s *Selection) SelectEntry(id core.BandmapEntryID) {
	found := false
	s.entries.ForEach(func(entry core.BandmapEntry) bool {
		if entry.ID == id && s.context.SelectableFilter()(entry) {
			s.selectEntry(s.context.TargetVFO(entry), entry)
			found = true
			return true
		}
		return false
	})
	if !found {
		s.clear()
	}
}

func (s *Selection) SelectByCallsign(call core.Callsign) {
	callStr := call.String()
	s.entries.ForEach(func(entry core.BandmapEntry) bool {
		if entry.Call.String() == callStr && s.context.SelectableFilter()(entry) {
			s.selectEntry(s.context.TargetVFO(entry), entry)
			return true
		}
		return false
	})
}

func (s *Selection) SelectHighestValue() {
	s.findAndSelect(
		core.Descending(core.BandmapByValue),
		s.context.NavigationFilter(),
		core.Not(core.IsWorkedSpot),
	)
}

func (s *Selection) SelectNearest(frequency core.Frequency) {
	s.findAndSelectRow(frequency, false, func(row core.BandmapRow) bool {
		return !rowOnFrequency(row, frequency)
	})
}

func (s *Selection) SelectNextUp(frequency core.Frequency) {
	s.findAndSelectRow(frequency, false, func(row core.BandmapRow) bool {
		return (row.Frequency() > frequency) ||
			(s.selected && row.Frequency() == frequency && row.ID() > s.selectedRow.ID())
	})
}

func (s *Selection) SelectNextDown(frequency core.Frequency) {
	s.findAndSelectRow(frequency, true, func(row core.BandmapRow) bool {
		return (row.Frequency() < frequency) ||
			(s.selected && row.Frequency() == frequency && row.ID() < s.selectedRow.ID())
	})
}

func rowOnFrequency(row core.BandmapRow, frequency core.Frequency) bool {
	if row.Kind == core.MarkerRow {
		return row.Marker.OnFrequency(frequency)
	}
	return row.Entry.OnFrequency(frequency)
}

func (s *Selection) SelectMarker(marker core.BandmapMarker) {
	s.selectMarker(s.context.TargetVFOForMarker(marker.Band), marker)
}
