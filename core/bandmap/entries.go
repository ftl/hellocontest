package bandmap

import (
	"log"
	"math"
	"time"

	"github.com/ftl/hellocontest/core"
	"golang.org/x/exp/slices"
)

const (
	// spots within this distance to an entry's frequency will be added to the entry
	spotFrequencyDeltaThreshold float64 = 500
	// the frequency of an entry is aligend to this is grid of accuracy
	spotFrequencyStep float64 = 10
)

type Entry struct {
	core.BandmapEntry

	spots   []core.Spot
	updated bool
}

func NewEntry(spot core.Spot) Entry {
	return Entry{
		BandmapEntry: core.BandmapEntry{
			Call:      spot.Call,
			Frequency: spot.Frequency,
			Band:      spot.Band,
			Mode:      spot.Mode,
			LastHeard: spot.Time,
			Source:    spot.Source,
			SpotCount: 1,
		},

		spots: []core.Spot{spot},
	}
}

func (e *Entry) Len() int {
	return len(e.spots)
}

func (e *Entry) Matches(spot core.Spot) bool {
	if spot.Call != e.Call {
		return false
	}
	if spot.Band != e.Band {
		return false
	}
	if spot.Mode != core.NoMode && e.Mode != core.NoMode && spot.Mode != e.Mode {
		return false
	}

	frequencyDelta := math.Abs(float64(e.Frequency - spot.Frequency))
	return frequencyDelta <= spotFrequencyDeltaThreshold
}

func (e *Entry) Add(spot core.Spot) bool {
	if !e.Matches(spot) {
		return false
	}

	e.spots = append(e.spots, spot)
	e.updateFrequency()
	if e.LastHeard.Before(spot.Time) {
		e.LastHeard = spot.Time
	}
	if e.Source.Priority() > spot.Source.Priority() {
		e.Source = spot.Source
	}

	return true
}

func (e *Entry) RemoveSpotsBefore(timestamp time.Time) bool {
	e.spots = filterSlice(e.spots, func(s core.Spot) bool {
		return !s.Time.Before(timestamp)
	})
	stillValid := len(e.spots) > 0

	if stillValid {
		e.update()
	}

	return stillValid
}

func (e *Entry) update() {
	frequencyUpdated := e.updateFrequency()

	lastHeard := time.Time{}
	var source core.SpotType
	for _, s := range e.spots {
		if lastHeard.Before(s.Time) {
			lastHeard = s.Time
		}
		if source.Priority() > s.Source.Priority() {
			source = s.Source
		}
	}

	e.updated = frequencyUpdated || (lastHeard != e.LastHeard) || (source != e.Source)
	e.LastHeard = lastHeard
	e.Source = source
	e.SpotCount = len(e.spots)
}

func (e *Entry) updateFrequency() bool {
	if len(e.spots) == 0 {
		e.Frequency = 0
		return true
	}

	var sum core.Frequency
	for _, s := range e.spots {
		sum += s.Frequency
	}
	downscaledMean := float64(sum) / (float64(len(e.spots)) * spotFrequencyStep)
	roundedMean := math.RoundToEven(downscaledMean)
	oldFrequency := e.Frequency
	e.Frequency = core.Frequency(roundedMean * spotFrequencyStep)

	return oldFrequency != e.Frequency
}

type EntryAddedListener interface {
	EntryAdded(core.BandmapEntry)
}

type EntryUpdatedListener interface {
	EntryUpdated(core.BandmapEntry)
}

type EntryRemovedListener interface {
	EntryRemoved(core.BandmapEntry)
}

type EntrySelectedListener interface {
	EntrySelected(core.BandmapEntry)
}

type Entries struct {
	entries         []*Entry
	bands           []core.Band
	summaries       map[core.Band]core.BandSummary
	order           core.BandmapOrder
	callinfo        Callinfo
	countEntryValue func(core.BandmapEntry) bool

	listeners []any
}

func NewEntries(countEntryValue func(core.BandmapEntry) bool) *Entries {
	result := &Entries{
		order:           core.BandmapByFrequency,
		callinfo:        new(nullCallinfo),
		countEntryValue: countEntryValue,
	}
	result.Clear()
	return result
}

func (l *Entries) SetBands(bands []core.Band) {
	l.bands = bands
	l.summaries = make(map[core.Band]core.BandSummary, len(bands))
}

func (l *Entries) SetCallinfo(callinfo Callinfo) {
	if callinfo == nil {
		l.callinfo = new(nullCallinfo)
		return
	}

	l.callinfo = callinfo
}

func (l *Entries) Notify(listener any) {
	l.listeners = append(l.listeners, listener)
}

func (l *Entries) emitEntryAdded(e Entry) {
	for _, listener := range l.listeners {
		if entryAddedListener, ok := listener.(EntryAddedListener); ok {
			entryAddedListener.EntryAdded(e.BandmapEntry)
		}
	}
}

func (l *Entries) emitEntryUpdated(e Entry) {
	for _, listener := range l.listeners {
		if entryUpdatedListener, ok := listener.(EntryUpdatedListener); ok {
			entryUpdatedListener.EntryUpdated(e.BandmapEntry)
		}
	}
}

func (l *Entries) emitEntryRemoved(e Entry) {
	for _, listener := range l.listeners {
		if entryRemovedListener, ok := listener.(EntryRemovedListener); ok {
			entryRemovedListener.EntryRemoved(e.BandmapEntry)
		}
	}
}

func (l *Entries) emitEntrySelected(e Entry) {
	for _, listener := range l.listeners {
		if entrySelectedListener, ok := listener.(EntrySelectedListener); ok {
			entrySelectedListener.EntrySelected(e.BandmapEntry)
		}
	}
}

func (l *Entries) Clear() {
	l.entries = make([]*Entry, 0, 100)
}

func (l *Entries) Len() int {
	return len(l.entries)
}

func (l *Entries) Add(spot core.Spot, now time.Time, weights core.BandmapWeights) {
	for _, e := range l.entries {
		if e.Add(spot) {
			e.Info = l.callinfo.GetInfo(spot.Call, spot.Band, spot.Mode, []string{})
			e.Info.WeightedValue = l.calculateWeightedValue(e, now, weights)
			l.emitEntryUpdated(*e)
			return
		}
	}

	newEntry := NewEntry(spot)
	if newEntry.Call.String() != "" {
		newEntry.Info = l.callinfo.GetInfo(newEntry.Call, newEntry.Band, newEntry.Mode, []string{})
		newEntry.Info.WeightedValue = l.calculateWeightedValue(&newEntry, now, weights)
	}
	l.insert(&newEntry)
	l.emitEntryAdded(newEntry)
}

func (l *Entries) insert(entry *Entry) {
	index := l.findIndexForInsert(entry)
	if index == len(l.entries) {
		l.entries = append(l.entries, entry)
		entry.Index = len(l.entries) - 1
		return
	}

	l.entries = append(l.entries, nil)
	copy(l.entries[index+1:], l.entries[index:])
	l.entries[index] = entry
	for i, e := range l.entries {
		e.Index = i
		l.entries[i] = e
	}
}

func (l *Entries) findIndexForInsert(entry *Entry) int {
	less := func(a, b *Entry) bool {
		return l.order(a.BandmapEntry, b.BandmapEntry)
	}
	left := 0
	right := len(l.entries) - 1
	for left <= right {
		pivot := (left + right) / 2
		if less(l.entries[pivot], entry) {
			left = pivot + 1
		} else if less(entry, l.entries[pivot]) {
			right = pivot - 1
		} else {
			return pivot
		}
	}
	return left
}

func (l *Entries) CleanOut(maximumAge time.Duration, now time.Time, weights core.BandmapWeights) {
	l.cleanOutOldEntries(maximumAge, now)
	l.cleanOutFalseEntries()

	l.summaries = make(map[core.Band]core.BandSummary, len(l.bands))
	for i, e := range l.entries {
		e.Index = i
		oldPoints, oldMultis, oldWeightedValue := e.Info.Points, e.Info.Multis, e.Info.WeightedValue
		e.Info.Points, e.Info.Multis, e.Info.MultiValues = l.callinfo.GetValue(e.Call, e.Band, e.Mode, []string{})
		e.Info.WeightedValue = l.calculateWeightedValue(e, now, weights)
		updated := e.updated || (oldPoints != e.Info.Points) || (oldMultis != e.Info.Multis) || (oldWeightedValue != e.Info.WeightedValue)
		e.updated = false
		l.entries[i] = e

		if updated {
			l.emitEntryUpdated(*e)
		}
		if l.countEntryValue(e.BandmapEntry) {
			l.addToSummary(e)
		}
	}
}

func (l *Entries) cleanOutOldEntries(maximumAge time.Duration, now time.Time) {
	deadline := now.Add(-maximumAge)
	removedEntries := make([]Entry, 0, len(l.entries))
	l.entries = filterSlice(l.entries, func(e *Entry) bool {
		stillValid := e.RemoveSpotsBefore(deadline)
		if !stillValid {
			removedEntries = append(removedEntries, *e)
		}
		return stillValid
	})
	for i, e := range removedEntries {
		e.Index -= i
		l.emitEntryRemoved(e)
	}
}

func (l *Entries) cleanOutFalseEntries() {
	removedEntries := make([]Entry, 0, len(l.entries))

	i := 0
	k := 0
	for i < len(l.entries) {
		entry1 := l.entries[i]
		if entry1 == nil {
			i++
			continue
		}

		for j := i + 1; j < len(l.entries); j++ {
			entry2 := l.entries[j]
			if entry2 == nil {
				continue
			}
			if !entry2.OnFrequency(entry1.Frequency) {
				break
			}

			switch CheckFalseEntry(entry1.BandmapEntry, entry2.BandmapEntry) {
			case DifferentEntries:
				continue
			case FirstIsFalse:
				removedEntries = append(removedEntries, *entry1)
				l.entries[i] = nil
				entry1 = nil
			case SecondIsFalse:
				removedEntries = append(removedEntries, *entry2)
				l.entries[j] = nil
			}
			if entry1 == nil {
				break
			}
		}

		if entry1 != nil {
			if i != k {
				l.entries[k] = entry1
			}
			k++
		}
		i++
	}
	l.entries = l.entries[:k]

	for i, e := range removedEntries {
		e.Index -= i
		l.emitEntryRemoved(e)
		log.Printf("false entry %s on %.2f kHz removed", e.Call, e.Frequency)
	}
}

func (l *Entries) calculateWeightedValue(entry *Entry, now time.Time, weights core.BandmapWeights) float64 {
	if entry.Source == core.WorkedSpot {
		return 0
	}

	points := float64(entry.Info.Points)
	multis := float64(entry.Info.Multis)
	value := (points * weights.TotalMultis) + (multis * weights.TotalPoints) + (points * multis)

	ageSeconds := now.Sub(entry.LastHeard).Seconds()
	spots := float64(entry.SpotCount)
	sourcePriority := float64(entry.Source.Priority())
	weight := 1 + (ageSeconds * weights.AgeSeconds) + (spots * weights.Spots) + (sourcePriority * weights.Source)

	return value * weight
}

func (l *Entries) addToSummary(entry *Entry) {
	summary, ok := l.summaries[entry.Band]
	if !ok {
		summary = core.BandSummary{Band: entry.Band}
	}
	summary.Points += entry.Info.Points
	summary.AddMultiValues(entry.Info.MultiValues)
	l.summaries[entry.Band] = summary
}

func (l *Entries) Bands(active, visible core.Band) []core.BandSummary {
	maxPointsIndex := 0
	maxPoints := 0
	maxMultisIndex := 0
	maxMultis := 0
	result := make([]core.BandSummary, len(l.bands))
	for i, band := range l.bands {
		summary, ok := l.summaries[band]
		if !ok {
			summary = core.BandSummary{
				Band: band,
			}
		}
		result[i] = summary
		result[i].Active = (summary.Band == active)
		result[i].Visible = (summary.Band == visible)
		if summary.Points > maxPoints {
			maxPoints = summary.Points
			maxPointsIndex = i
		}
		multis := summary.Multis()
		if multis > maxMultis {
			maxMultis = multis
			maxMultisIndex = i
		}
	}

	if maxPoints > 0 && maxPointsIndex < len(result) {
		result[maxPointsIndex].MaxPoints = true
	}
	if maxMultis > 0 && maxMultisIndex < len(result) {
		result[maxMultisIndex].MaxMultis = true
	}

	return result
}

func (l *Entries) DoOnEntry(index int, f func(core.BandmapEntry)) {
	if index < 0 || index >= len(l.entries) {
		f(core.BandmapEntry{})
		return
	}

	entry := l.entries[index]
	f(entry.BandmapEntry)
}

func (l *Entries) Select(index int) {
	if index < 0 || index >= len(l.entries) {
		return
	}

	entry := l.entries[index]
	l.emitEntrySelected(*entry)
}

func (l *Entries) All() []core.BandmapEntry {
	result := make([]core.BandmapEntry, len(l.entries))
	for i, e := range l.entries {
		result[i] = e.BandmapEntry
	}
	return result
}

func (l *Entries) AllBy(order core.BandmapOrder) []core.BandmapEntry {
	result := l.All()
	slices.SortStableFunc(result, order)
	return result
}

func (l *Entries) ForEach(f func(entry core.BandmapEntry) bool) {
	for _, entry := range l.entries {
		if f(entry.BandmapEntry) {
			return
		}
	}
}

func filterSlice[E any](slice []E, filter func(E) bool) []E {
	k := 0
	for i, e := range slice {
		if filter(e) {
			if i != k {
				slice[k] = e
			}
			k++
		}
	}
	return slice[:k]
}
