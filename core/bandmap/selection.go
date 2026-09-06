package bandmap

import (
	"github.com/ftl/hellocontest/core"
)

type SelectionContext interface {
	SelectableFilter() core.BandmapFilter
	NavigationFilter() core.BandmapFilter
	TargetVFO(core.BandmapEntry) core.VFOID
	FocusedVFO() core.VFOID
}

type Selection struct {
	selectedEntry core.BandmapEntry
	selected      bool

	entries  *Entries
	notifier *Notifier
	context  SelectionContext
}

func NewSelection(entries *Entries, notifier *Notifier, context SelectionContext) *Selection {
	return &Selection{
		entries:  entries,
		notifier: notifier,
		context:  context,
	}
}

func (s *Selection) selectEntry(vfo core.VFOID, entry core.BandmapEntry) {
	s.selectedEntry = entry
	s.selected = true
	s.notifier.emitEntrySelected(vfo, s.selectedEntry)
}

func (s *Selection) clear() {
	s.selectedEntry = core.BandmapEntry{}
	s.selected = false
	// TODO??? s.notifier.emitEntrySelected(s.selectedEntry)
}

func (s *Selection) findAndSelect(order core.BandmapOrder, filters ...core.BandmapFilter) {
	entries := s.entries.Query(order, filters...)
	if len(entries) > 0 {
		s.selectEntry(s.context.FocusedVFO(), entries[0])
	}
}

func (s *Selection) SelectedEntry() (core.BandmapEntry, bool) {
	return s.selectedEntry, s.selected
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
	s.findAndSelect(
		core.BandmapByDistance(frequency),
		s.context.NavigationFilter(),
		core.Not(core.OnFrequency(frequency)),
	)
}

func (s *Selection) SelectNextUp(frequency core.Frequency) {
	s.findAndSelect(
		core.BandmapByDistance(frequency),
		s.context.NavigationFilter(),
		func(entry core.BandmapEntry) bool {
			return (entry.Frequency > frequency) ||
				(s.selected && entry.Frequency == frequency && entry.ID > s.selectedEntry.ID)
		},
	)
}

func (s *Selection) SelectNextDown(frequency core.Frequency) {
	s.findAndSelect(
		core.BandmapByDistanceAndDescendingID(frequency),
		s.context.NavigationFilter(),
		func(entry core.BandmapEntry) bool {
			return (entry.Frequency < frequency) ||
				(s.selected && entry.Frequency == frequency && entry.ID < s.selectedEntry.ID)
		},
	)
}
