package bandmap

import (
	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

type Selection struct {
	selectedEntry core.BandmapEntry
	selected      bool

	entries       *Entries
	notifier      *Notifier
	visibleFilter core.BandmapFilter
}

func NewSelection(entries *Entries, notifier *Notifier, visibleFilter core.BandmapFilter) *Selection {
	return &Selection{
		entries:       entries,
		notifier:      notifier,
		visibleFilter: visibleFilter,
	}
}

func (s *Selection) selectEntry(entry core.BandmapEntry) {
	s.selectedEntry = entry
	s.selected = true
	s.notifier.emitEntrySelected(s.selectedEntry)
}

func (s *Selection) clear() {
	s.selectedEntry = core.BandmapEntry{}
	s.selected = false
	// TODO??? s.notifier.emitEntrySelected(s.selectedEntry)
}

func (s *Selection) findAndSelect(order core.BandmapOrder, filters ...core.BandmapFilter) {
	entries := s.entries.Query(order, filters...)
	if len(entries) > 0 {
		s.selectEntry(entries[0])
	}
}

func (s *Selection) SelectedEntry() (core.BandmapEntry, bool) {
	return s.selectedEntry, s.selected
}

func (s *Selection) SelectEntry(id core.BandmapEntryID) {
	found := false
	s.entries.ForEach(func(entry core.BandmapEntry) bool {
		if entry.ID == id && s.visibleFilter(entry) {
			s.selectEntry(entry)
			found = true
			return true
		}
		return false
	})
	if !found {
		s.clear()
	}
}

func (s *Selection) SelectByCallsign(call callsign.Callsign) {
	callStr := call.String()
	s.entries.ForEach(func(entry core.BandmapEntry) bool {
		if entry.Call.String() == callStr && s.visibleFilter(entry) {
			s.selectEntry(entry)
			return true
		}
		return false
	})
}

func (s *Selection) SelectHighestValue() {
	s.findAndSelect(
		core.Descending(core.BandmapByValue),
		s.visibleFilter,
		core.Not(core.IsWorkedSpot),
	)
}

func (s *Selection) SelectNearest(frequency core.Frequency) {
	s.findAndSelect(
		core.BandmapByDistance(frequency),
		s.visibleFilter,
		core.Not(core.OnFrequency(frequency)),
	)
}

func (s *Selection) SelectNextUp(frequency core.Frequency) {
	s.findAndSelect(
		core.BandmapByDistance(frequency),
		s.visibleFilter,
		func(entry core.BandmapEntry) bool {
			return (entry.Frequency > frequency) ||
				(s.selected && entry.Frequency == frequency && entry.ID > s.selectedEntry.ID)
		},
	)
}

func (s *Selection) SelectNextDown(frequency core.Frequency) {
	s.findAndSelect(
		core.BandmapByDistanceAndDescendingID(frequency),
		s.visibleFilter,
		func(entry core.BandmapEntry) bool {
			return (entry.Frequency < frequency) ||
				(s.selected && entry.Frequency == frequency && entry.ID < s.selectedEntry.ID)
		},
	)
}
