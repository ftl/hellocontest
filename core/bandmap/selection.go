package bandmap

import (
	"slices"

	"github.com/ftl/hellocontest/core"
)

type SelectionContext interface {
	VisibleRows() []core.BandmapRow
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

func (s *Selection) clear() {
	s.selectedRow = core.BandmapRow{}
	s.selected = false
	// TODO??? s.notifier.emitEntrySelected(s.selectedRow.Entry)
}

// findAndSelectRow navigates through the rows that the spots table shows, therefore the
// operator reaches a marker with the same actions as a spot. The focused VFO goes to the
// row, because the operator searches from the frequency of that VFO.
func (s *Selection) findAndSelectRow(frequency core.Frequency, descendingID bool, matches func(core.BandmapRow) bool) {
	rows := make([]core.BandmapRow, 0)
	for _, row := range navigableRows(s.context.VisibleRows()) {
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

	s.selectRowOnFocusedVFO(rows[0])
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

// SelectHighestValue takes all spots that the spots table shows into account, on every band,
// and the target VFO works the spot, with the same rules as when the operator clicks a row.
func (s *Selection) SelectHighestValue() {
	var best core.BandmapRow
	found := false
	for _, row := range navigableRows(s.context.VisibleRows()) {
		if row.Kind != core.SpotRow || core.IsWorkedSpot(row.Entry) {
			continue
		}
		if found && core.BandmapByValue(row.Entry, best.Entry) <= 0 {
			continue
		}
		best = row
		found = true
	}
	if !found {
		return
	}

	s.selectRowOnTargetVFO(best)
}

func (s *Selection) SelectNearest(frequency core.Frequency) {
	s.findAndSelectRow(frequency, false, func(row core.BandmapRow) bool {
		return !rowOnFrequency(row, frequency)
	})
}

func (s *Selection) SelectNextUp(frequency core.Frequency) {
	s.selectRowByOffset(frequency, -1)
}

func (s *Selection) SelectNextDown(frequency core.Frequency) {
	s.selectRowByOffset(frequency, 1)
}

// selectRowByOffset moves the selection through the rows that the spots table shows, in the
// order of the table. Without a selection it starts at the row that is closest to frequency,
// and it stops at both ends of the table.
func (s *Selection) selectRowByOffset(frequency core.Frequency, offset int) {
	rows := navigableRows(s.context.VisibleRows())
	if len(rows) == 0 {
		return
	}

	index := s.indexOfSelectedRow(rows)
	if index < 0 {
		index = indexOfNearestRow(rows, frequency)
	} else {
		index += offset
	}
	if index < 0 || index >= len(rows) {
		return
	}

	s.selectRowOnTargetVFO(rows[index])
}

// navigableRows removes the CQ marker, because the CQ frequency has its own action and the
// navigation does not stop there.
func navigableRows(rows []core.BandmapRow) []core.BandmapRow {
	result := make([]core.BandmapRow, 0, len(rows))
	for _, row := range rows {
		if row.Kind == core.MarkerRow && row.Marker.Kind == core.CQMarker {
			continue
		}
		result = append(result, row)
	}
	return result
}

func (s *Selection) indexOfSelectedRow(rows []core.BandmapRow) int {
	if !s.selected {
		return -1
	}
	for i, row := range rows {
		if row.Kind == s.selectedRow.Kind && row.ID() == s.selectedRow.ID() {
			return i
		}
	}
	return -1
}

func indexOfNearestRow(rows []core.BandmapRow, frequency core.Frequency) int {
	result := 0
	for i, row := range rows {
		if distance(row, frequency) < distance(rows[result], frequency) {
			result = i
		}
	}
	return result
}

// selectRowOnTargetVFO selects a row with the same rules that apply when the operator clicks
// the row in the spots table: the target VFO works the spot, also on another band.
func (s *Selection) selectRowOnTargetVFO(row core.BandmapRow) {
	if row.Kind == core.MarkerRow {
		s.selectMarker(s.context.TargetVFOForMarker(row.Marker.Band), row.Marker)
		return
	}
	s.selectEntry(s.context.TargetVFO(row.Entry), row.Entry)
}

func (s *Selection) selectRowOnFocusedVFO(row core.BandmapRow) {
	vfo := s.context.FocusedVFO()
	if row.Kind == core.MarkerRow {
		s.selectMarker(vfo, row.Marker)
		return
	}
	s.selectEntry(vfo, row.Entry)
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
