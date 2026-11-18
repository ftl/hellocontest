package bandmap

import (
	"slices"

	"github.com/ftl/hellocontest/core/model"
)

type SelectionContext interface {
	VisibleRows() []model.BandmapRow
	SelectableFilter() model.BandmapFilter
	NavigationFilter() model.BandmapFilter
	MarkerNavigationFilter() model.BandmapMarkerFilter
	TargetVFO(model.BandmapEntry) model.VFOID
	TargetVFOForMarker(model.Band) model.VFOID
	FocusedVFO() model.VFOID
}

type Selection struct {
	selectedRow model.BandmapRow
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

func (s *Selection) selectEntry(vfo model.VFOID, entry model.BandmapEntry) {
	s.selectedRow = model.SpotBandmapRow(entry)
	s.selected = true
	s.notifier.emitEntrySelected(vfo, entry)
}

func (s *Selection) selectMarker(vfo model.VFOID, marker model.BandmapMarker) {
	s.selectedRow = model.MarkerBandmapRow(marker)
	s.selected = true
	s.notifier.emitMarkerSelected(vfo, marker)
}

func (s *Selection) clear() {
	s.selectedRow = model.BandmapRow{}
	s.selected = false
	// TODO??? s.notifier.emitEntrySelected(s.selectedRow.Entry)
}

// findAndSelectRow navigates through the rows that the spots table shows, therefore the
// operator reaches a marker with the same actions as a spot. The focused VFO goes to the
// row, because the operator searches from the frequency of that VFO.
func (s *Selection) findAndSelectRow(frequency model.Frequency, descendingID bool, matches func(model.BandmapRow) bool) {
	rows := make([]model.BandmapRow, 0)
	for _, row := range navigableRows(s.context.VisibleRows()) {
		if matches(row) {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return
	}

	slices.SortStableFunc(rows, func(a, b model.BandmapRow) int {
		result := model.Compare(distance(a, frequency), distance(b, frequency))
		if result != 0 {
			return result
		}
		return applyDirection(model.Compare(a.ID(), b.ID()), descendingID)
	})

	s.selectRowOnFocusedVFO(rows[0])
}

func distance(row model.BandmapRow, frequency model.Frequency) model.Frequency {
	return row.Frequency().DistanceTo(frequency)
}

func (s *Selection) SelectedEntry() (model.BandmapEntry, bool) {
	if s.selectedRow.Kind != model.SpotRow {
		return model.BandmapEntry{}, false
	}
	return s.selectedRow.Entry, s.selected
}

func (s *Selection) SelectEntry(id model.BandmapEntryID) {
	found := false
	s.entries.ForEach(func(entry model.BandmapEntry) bool {
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

func (s *Selection) SelectByCallsign(call model.Callsign) {
	callStr := call.String()
	s.entries.ForEach(func(entry model.BandmapEntry) bool {
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
	var best model.BandmapRow
	found := false
	for _, row := range navigableRows(s.context.VisibleRows()) {
		if row.Kind != model.SpotRow || model.IsWorkedSpot(row.Entry) {
			continue
		}
		if found && model.BandmapByValue(row.Entry, best.Entry) <= 0 {
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

func (s *Selection) SelectNearest(frequency model.Frequency) {
	s.findAndSelectRow(frequency, false, func(row model.BandmapRow) bool {
		return !rowOnFrequency(row, frequency)
	})
}

func (s *Selection) SelectNextUp(frequency model.Frequency) {
	s.selectRowByOffset(frequency, -1)
}

func (s *Selection) SelectNextDown(frequency model.Frequency) {
	s.selectRowByOffset(frequency, 1)
}

// selectRowByOffset moves the selection through the rows that the spots table shows, in the
// order of the table. Without a selection it starts at the row that is closest to frequency,
// and it stops at both ends of the table.
func (s *Selection) selectRowByOffset(frequency model.Frequency, offset int) {
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
func navigableRows(rows []model.BandmapRow) []model.BandmapRow {
	result := make([]model.BandmapRow, 0, len(rows))
	for _, row := range rows {
		if row.Kind == model.MarkerRow && row.Marker.Kind == model.CQMarker {
			continue
		}
		result = append(result, row)
	}
	return result
}

func (s *Selection) indexOfSelectedRow(rows []model.BandmapRow) int {
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

func indexOfNearestRow(rows []model.BandmapRow, frequency model.Frequency) int {
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
func (s *Selection) selectRowOnTargetVFO(row model.BandmapRow) {
	if row.Kind == model.MarkerRow {
		s.selectMarker(s.context.TargetVFOForMarker(row.Marker.Band), row.Marker)
		return
	}
	s.selectEntry(s.context.TargetVFO(row.Entry), row.Entry)
}

func (s *Selection) selectRowOnFocusedVFO(row model.BandmapRow) {
	vfo := s.context.FocusedVFO()
	if row.Kind == model.MarkerRow {
		s.selectMarker(vfo, row.Marker)
		return
	}
	s.selectEntry(vfo, row.Entry)
}

func rowOnFrequency(row model.BandmapRow, frequency model.Frequency) bool {
	if row.Kind == model.MarkerRow {
		return row.Marker.OnFrequency(frequency)
	}
	return row.Entry.OnFrequency(frequency)
}

func (s *Selection) SelectMarker(marker model.BandmapMarker) {
	s.selectMarker(s.context.TargetVFOForMarker(marker.Band), marker)
}
