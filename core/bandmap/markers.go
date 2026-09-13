package bandmap

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/hellocontest/core"
)

const cqMarkerText = "CQ"

type idSource interface {
	// nextID gives the markers the same IDs as the spots, so that the spot list can tell
	// its rows apart with one ID space
	nextID() core.BandmapEntryID
}

type Markers struct {
	ids     idSource
	markers []core.BandmapMarker
}

func NewMarkers(ids idSource) *Markers {
	return &Markers{
		ids: ids,
	}
}

func (m *Markers) AddNumbered(number int, frequency core.Frequency, band core.Band, now time.Time) core.BandmapMarker {
	return m.add(core.BandmapMarker{
		Kind:      core.NumberedMarker,
		Number:    number,
		Frequency: frequency,
		Band:      band,
		CreatedAt: now,
	})
}

func (m *Markers) AddText(text string, frequency core.Frequency, band core.Band, now time.Time) core.BandmapMarker {
	return m.add(core.BandmapMarker{
		Kind:      core.TextMarker,
		Text:      text,
		Frequency: frequency,
		Band:      band,
		CreatedAt: now,
	})
}

func (m *Markers) add(marker core.BandmapMarker) core.BandmapMarker {
	marker.ID = m.ids.nextID()
	marker.Text = markerText(marker)

	if marker.Kind == core.CQMarker {
		// there is either none or only one marker of this kind
		m.removeKind(core.CQMarker)
	}
	m.markers = append(m.markers, marker)

	return marker
}

func (m *Markers) SetCQ(frequency core.Frequency, band core.Band, now time.Time) core.BandmapMarker {
	for i, marker := range m.markers {
		if marker.Kind != core.CQMarker {
			continue
		}
		// the marker keeps its ID, so that the spot list moves the row instead of
		// removing and inserting it
		marker.Frequency = frequency
		marker.Band = band
		marker.CreatedAt = now
		m.markers[i] = marker
		return marker
	}

	return m.add(core.BandmapMarker{
		Kind:      core.CQMarker,
		Frequency: frequency,
		Band:      band,
		CreatedAt: now,
	})
}

func (m *Markers) MarkWithNextNumber(frequency core.Frequency, band core.Band, now time.Time) core.BandmapMarker {
	// a marker that is already close to that frequency moves instead, so that the operator
	// does not collect several markers on the same signal
	if i, ok := m.closestOnFrequency(frequency, core.NumberedMarker, core.TextMarker); ok {
		return m.moveMarker(i, frequency, band, now)
	}
	return m.AddNumbered(m.nextAvailableNumber(), frequency, band, now)
}

func (m *Markers) MarkWithNumber(number int, frequency core.Frequency, band core.Band, now time.Time) core.BandmapMarker {
	// the number always wins: an existing marker with that number moves, a marker close to
	// the frequency takes the number over, and a text marker close to the frequency becomes
	// a numbered marker
	if i, ok := m.indexOfNumber(number); ok {
		return m.moveMarker(i, frequency, band, now)
	}
	if i, ok := m.closestOnFrequency(frequency, core.NumberedMarker); ok {
		return m.renumberMarker(i, number, frequency, band, now)
	}
	if i, ok := m.closestOnFrequency(frequency, core.TextMarker); ok {
		return m.renumberMarker(i, number, frequency, band, now)
	}
	return m.AddNumbered(number, frequency, band, now)
}

func (m *Markers) MarkWithText(text string, frequency core.Frequency, band core.Band, now time.Time) core.BandmapMarker {
	// a marker that is already close to that frequency takes the new text, the text of a
	// marker does not have to be unique
	if i, ok := m.closestOnFrequency(frequency, core.TextMarker, core.NumberedMarker); ok {
		return m.retextMarker(i, text, frequency, band, now)
	}
	return m.AddText(text, frequency, band, now)
}

func (m *Markers) nextAvailableNumber() int {
	used := make(map[int]bool, len(m.markers))
	for _, marker := range m.markers {
		if marker.Kind == core.NumberedMarker {
			used[marker.Number] = true
		}
	}
	result := 1
	for used[result] {
		result++
	}
	return result
}

func (m *Markers) indexOfNumber(number int) (int, bool) {
	for i, marker := range m.markers {
		if marker.Kind == core.NumberedMarker && marker.Number == number {
			return i, true
		}
	}
	return 0, false
}

func (m *Markers) closestOnFrequency(frequency core.Frequency, kinds ...core.BandmapMarkerKind) (int, bool) {
	result := 0
	found := false
	for i, marker := range m.markers {
		if !slices.Contains(kinds, marker.Kind) {
			continue
		}
		if !marker.OnFrequency(frequency) {
			continue
		}
		if found && math.Abs(m.markers[result].ProximityFactor(frequency)) >= math.Abs(marker.ProximityFactor(frequency)) {
			continue
		}
		result = i
		found = true
	}
	return result, found
}

// the marker keeps its ID, so that the spot list moves the row instead of removing and
// inserting it. Every adjustment renews the creation time.
func (m *Markers) moveMarker(i int, frequency core.Frequency, band core.Band, now time.Time) core.BandmapMarker {
	marker := m.markers[i]
	marker.Frequency = frequency
	marker.Band = band
	marker.CreatedAt = now
	m.markers[i] = marker
	return marker
}

func (m *Markers) renumberMarker(i int, number int, frequency core.Frequency, band core.Band, now time.Time) core.BandmapMarker {
	marker := m.markers[i]
	marker.Kind = core.NumberedMarker
	marker.Number = number
	marker.Text = markerText(marker)
	marker.Frequency = frequency
	marker.Band = band
	marker.CreatedAt = now
	m.markers[i] = marker
	return marker
}

func (m *Markers) retextMarker(i int, text string, frequency core.Frequency, band core.Band, now time.Time) core.BandmapMarker {
	marker := m.markers[i]
	marker.Kind = core.TextMarker
	// a text marker has no number, therefore the number becomes available again
	marker.Number = 0
	marker.Text = text
	marker.Frequency = frequency
	marker.Band = band
	marker.CreatedAt = now
	m.markers[i] = marker
	return marker
}

func (m *Markers) Get(id core.BandmapEntryID) (core.BandmapMarker, bool) {
	for _, marker := range m.markers {
		if marker.ID == id {
			return marker, true
		}
	}
	return core.BandmapMarker{}, false
}

func (m *Markers) Numbered(number int) (core.BandmapMarker, bool) {
	i, found := m.indexOfNumber(number)
	if !found {
		return core.BandmapMarker{}, false
	}
	return m.markers[i], true
}

func (m *Markers) CQ() (core.BandmapMarker, bool) {
	for _, marker := range m.markers {
		if marker.Kind == core.CQMarker {
			return marker, true
		}
	}
	return core.BandmapMarker{}, false
}

func (m *Markers) RemoveOnFrequency(frequency core.Frequency) (core.BandmapMarker, bool) {
	// the CQ marker belongs to the running workmode, only a numbered or a text marker is
	// deleted this way
	i, found := m.closestOnFrequency(frequency, core.NumberedMarker, core.TextMarker)
	if !found {
		return core.BandmapMarker{}, false
	}
	marker := m.markers[i]
	m.markers = slices.Delete(m.markers, i, i+1)
	return marker, true
}

func (m *Markers) Remove(id core.BandmapEntryID) {
	m.markers = slices.DeleteFunc(m.markers, func(marker core.BandmapMarker) bool {
		return marker.ID == id
	})
}

func (m *Markers) Clear() {
	m.markers = nil
}

func (m *Markers) Len() int {
	return len(m.markers)
}

func (m *Markers) All() []core.BandmapMarker {
	return slices.Clone(m.markers)
}

func (m *Markers) Ordered(matches core.BandmapMarkerFilter, column core.SpotSortColumn, descending bool) []core.BandmapMarker {
	result := make([]core.BandmapMarker, 0, len(m.markers))
	for _, marker := range m.markers {
		if matches != nil && !matches(marker) {
			continue
		}
		result = append(result, marker)
	}
	slices.SortStableFunc(result, func(a, b core.BandmapMarker) int {
		return compareMarkers(a, b, column, descending)
	})
	return result
}

func (m *Markers) removeKind(kind core.BandmapMarkerKind) {
	m.markers = slices.DeleteFunc(m.markers, func(marker core.BandmapMarker) bool {
		return marker.Kind == kind
	})
}

func markerText(marker core.BandmapMarker) string {
	switch marker.Kind {
	case core.CQMarker:
		// the text of a CQ marker is always the same
		return cqMarkerText
	case core.NumberedMarker:
		// the number is the text of a numbered marker, therefore the spot list shows and
		// sorts it without a special case
		return strconv.Itoa(marker.Number)
	default:
		return marker.Text
	}
}

func blendsMarkersWithSpots(column core.SpotSortColumn) bool {
	// a marker has a frequency and a creation time, therefore it fits between the spots
	// with those two orders
	switch column {
	case core.SortSpotsByFrequency, core.SortSpotsByLastSeen:
		return true
	default:
		return false
	}
}

func compareRows(a, b core.BandmapRow, column core.SpotSortColumn, descending bool) int {
	var result int
	if column == core.SortSpotsByLastSeen {
		result = a.LastSeen().Compare(b.LastSeen())
	} else {
		result = core.Compare(a.Frequency(), b.Frequency())
	}
	if result == 0 {
		result = core.Compare(a.ID(), b.ID())
	}
	return applyDirection(result, descending)
}

func compareMarkers(a, b core.BandmapMarker, column core.SpotSortColumn, descending bool) int {
	switch column {
	case core.SortSpotsByCallsign, core.SortSpotsByValue:
		// a marker has no callsign and no value, therefore its text decides. The CQ marker
		// keeps the first place in both directions.
		if a.Kind != b.Kind {
			if a.Kind == core.CQMarker {
				return -1
			}
			if b.Kind == core.CQMarker {
				return 1
			}
		}
		return applyDirection(compareMarkerText(a.Text, b.Text), descending)
	case core.SortSpotsByLastSeen:
		// the creation time is the last seen value of a marker
		return applyDirection(a.CreatedAt.Compare(b.CreatedAt), descending)
	default:
		return applyDirection(core.Compare(a.Frequency, b.Frequency), descending)
	}
}

func compareMarkerText(a, b string) int {
	numberA, aIsNumber := markerTextNumber(a)
	numberB, bIsNumber := markerTextNumber(b)
	switch {
	case aIsNumber && bIsNumber:
		return core.Compare(numberA, numberB)
	case aIsNumber:
		return -1
	case bIsNumber:
		return 1
	default:
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	}
}

func markerTextNumber(text string) (int, bool) {
	number, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		return 0, false
	}
	return number, true
}

func applyDirection(comparison int, descending bool) int {
	if descending {
		return -comparison
	}
	return comparison
}
