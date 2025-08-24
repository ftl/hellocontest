package bandmap

import (
	"math"
	"slices"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/texttheater/golang-levenshtein/levenshtein"

	"github.com/ftl/hellocontest/core"
)

const (
	// the frequency of an entry is aligend to this is grid of accuracy
	spotFrequencyStep float64 = 10
	// at least this number of spots of the same callsign on the same frequency are required for a valid spot
	spotValidThreshold = 3
	// if one callsign can be transformed into another with only this number of transformations, they are considered similar
	similarCallsignThreshold = 2
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

func (e *Entry) Matches(spot core.Spot) (core.SpotQuality, bool) {
	if spot.Band != e.Band {
		return core.UnknownSpotQuality, false
	}
	if spot.Mode != core.NoMode && e.Mode != core.NoMode && spot.Mode != e.Mode {
		return core.UnknownSpotQuality, false
	}

	var callsignDistance int
	if spot.Call == e.Call {
		callsignDistance = 0
	} else {
		callsignDistance = calculateCallsignDistance(spot.Call, e.Call)
	}

	onFrequency := e.OnFrequency(spot.Frequency)

	quality := core.UnknownSpotQuality
	if len(e.spots)+1 >= spotValidThreshold {
		quality = core.ValidSpotQuality
	}

	if callsignDistance == 0 && onFrequency {
		return quality, true
	} else if e.Quality == core.ValidSpotQuality && callsignDistance == 0 && !onFrequency {
		return core.QSYSpotQuality, false
	} else if e.Quality == core.ValidSpotQuality && callsignDistance <= similarCallsignThreshold && onFrequency {
		return core.BustedSpotQuality, false
	} else {
		return core.UnknownSpotQuality, false
	}
}

func calculateCallsignDistance(call1, call2 callsign.Callsign) int {
	options := levenshtein.DefaultOptions
	return levenshtein.DistanceForStrings([]rune(call1.String()), []rune(call2.String()), options)
}

func (e *Entry) Add(spot core.Spot) (core.SpotQuality, bool) {
	quality, match := e.Matches(spot)
	if !match {
		return quality, false
	}

	e.spots = append(e.spots, spot)
	e.updateFrequency()
	if e.LastHeard.Before(spot.Time) {
		e.LastHeard = spot.Time
	}
	if e.Source.Priority() > spot.Source.Priority() {
		e.Source = spot.Source
	}
	if quality == core.ValidSpotQuality {
		e.Quality = quality
	}

	return quality, true
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
	if e.Source != core.WorkedSpot {
		e.Source = source
	}
	e.SpotCount = len(e.spots)
	if e.SpotCount < spotValidThreshold && e.Quality == core.ValidSpotQuality {
		e.Quality = core.UnknownSpotQuality
	}
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

func (e *Entry) MatchesAllFilters(filters ...core.BandmapFilter) bool {
	for _, filter := range filters {
		if !filter(e.BandmapEntry) {
			return false
		}
	}
	return true
}

func (e *Entry) MatchesAnyFilter(filters ...core.BandmapFilter) bool {
	for _, filter := range filters {
		if filter(e.BandmapEntry) {
			return true
		}
	}
	return false
}

type Entries struct {
	entries         []*Entry
	bands           []core.Band
	summaries       map[core.Band]core.BandSummary
	order           core.BandmapOrder
	callinfo        Callinfo
	countEntryValue func(core.BandmapEntry) bool
	lastID          core.BandmapEntryID

	notifier *Notifier
}

func NewEntries(notifier *Notifier, countEntryValue func(core.BandmapEntry) bool) *Entries {
	result := &Entries{
		order:           core.BandmapByFrequency,
		callinfo:        new(nullCallinfo),
		countEntryValue: countEntryValue,
		notifier:        notifier,
	}
	result.Clear()
	return result
}

func (l *Entries) nextID() core.BandmapEntryID {
	l.lastID++
	return l.lastID
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

func (l *Entries) Clear() {
	l.entries = make([]*Entry, 0, 100)
}

func (l *Entries) Len() int {
	return len(l.entries)
}

func (l *Entries) Add(spot core.Spot, now time.Time, weights core.BandmapWeights) {
	entryQuality := core.UnknownSpotQuality
	for _, e := range l.entries {
		quality, added := e.Add(spot)
		if added {
			l.complementCallinfo(e, now, weights)
			l.notifier.emitEntryUpdated(e.BandmapEntry)
			return
		}
		if entryQuality < quality {
			entryQuality = quality
		}
	}

	newEntry := NewEntry(spot)
	newEntry.Quality = entryQuality
	l.complementCallinfo(&newEntry, now, weights)
	l.insert(&newEntry)
	l.notifier.emitEntryAdded(newEntry.BandmapEntry)
}

func (l *Entries) insert(entry *Entry) {
	entry.ID = l.nextID()

	index := l.findIndexForInsert(entry)
	if index == len(l.entries) {
		l.entries = append(l.entries, entry)
		return
	}

	l.entries = append(l.entries, nil)
	copy(l.entries[index+1:], l.entries[index:])
	l.entries[index] = entry
}

func (l *Entries) findIndexForInsert(entry *Entry) int {
	less := func(a, b *Entry) int {
		return l.order(b.BandmapEntry, a.BandmapEntry)
	}
	left := 0
	right := len(l.entries) - 1
	for left <= right {
		pivot := (left + right) / 2
		if less(l.entries[pivot], entry) < 0 {
			left = pivot + 1
		} else if less(entry, l.entries[pivot]) < 0 {
			right = pivot - 1
		} else {
			return pivot
		}
	}
	return left
}

func (l *Entries) MarkAsWorked(call callsign.Callsign, band core.Band, mode core.Mode) {
	for _, e := range l.entries {
		match := true
		if call != e.Call {
			match = false
		}
		if band != core.NoBand && band != e.Band {
			match = false
		}
		if mode != core.NoMode && mode != e.Mode {
			match = false
		}

		if !match {
			continue
		}

		e.Source = core.WorkedSpot
		e.update()
		l.notifier.emitEntryUpdated(e.BandmapEntry)
	}
}

func (l *Entries) CleanOut(maximumAge time.Duration, now time.Time, weights core.BandmapWeights) {
	l.cleanOutOldEntries(maximumAge, now)

	l.summaries = make(map[core.Band]core.BandSummary, len(l.bands))
	for i, e := range l.entries {
		oldPoints, oldMultis, oldWeightedValue := e.Info.Points, e.Info.Multis, e.Info.WeightedValue
		l.complementValue(e, now, weights)
		updated := e.updated || (oldPoints != e.Info.Points) || (oldMultis != e.Info.Multis) || (oldWeightedValue != e.Info.WeightedValue)
		e.updated = false
		l.entries[i] = e

		if updated {
			l.notifier.emitEntryUpdated(e.BandmapEntry)
		}
		if l.countEntryValue(e.BandmapEntry) {
			l.addToSummary(e)
		}
	}
}

func (l *Entries) cleanOutOldEntries(maximumAge time.Duration, now time.Time) {
	deadline := now.Add(-maximumAge)
	l.entries = filterSlice(l.entries, func(e *Entry) bool {
		stillValid := e.RemoveSpotsBefore(deadline)
		if !stillValid {
			l.notifier.emitEntryRemoved(e.BandmapEntry)
		}
		return stillValid
	})
}

func (l *Entries) complementCallinfo(entry *Entry, now time.Time, weights core.BandmapWeights) {
	if entry.Call.String() == "" {
		return
	}
	entry.Info = l.callinfo.GetInfo(entry.Call, entry.Band, entry.Mode, []string{})
	entry.Info.WeightedValue = l.calculateWeightedValue(entry, now, weights)
}

func (l *Entries) complementValue(entry *Entry, now time.Time, weights core.BandmapWeights) {
	if entry.Call.String() == "" {
		return
	}
	ok := l.callinfo.UpdateValue(&entry.Info, entry.Band, entry.Mode)
	if ok {
		entry.Info.WeightedValue = l.calculateWeightedValue(entry, now, weights)
	}
}

func (l *Entries) calculateWeightedValue(entry *Entry, now time.Time, weights core.BandmapWeights) float64 {
	if entry.Source == core.WorkedSpot {
		return 0
	}

	value := float64(entry.Info.Value)

	ageSeconds := now.Sub(entry.LastHeard).Seconds()
	spots := float64(entry.SpotCount)
	sourcePriority := float64(entry.Source.Priority())
	qualityFactor := 0.0
	if entry.Quality > core.BustedSpotQuality {
		qualityFactor = float64(entry.Quality)
	}
	weight := 1 + (ageSeconds * weights.AgeSeconds) + (spots * weights.Spots) + (sourcePriority * weights.Source) + (qualityFactor * weights.Quality)

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

func (l *Entries) DoOnEntry(id core.BandmapEntryID, f func(core.BandmapEntry)) {
	for _, entry := range l.entries {
		if entry.ID == id {
			f(entry.BandmapEntry)
			return
		}
	}
	f(core.BandmapEntry{})
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
	if order != nil {
		slices.SortStableFunc(result, order)
	}
	return result
}

func (l *Entries) Query(order core.BandmapOrder, filters ...core.BandmapFilter) []core.BandmapEntry {
	result := make([]core.BandmapEntry, 0, len(l.entries))
	for _, e := range l.entries {
		if e.MatchesAllFilters(filters...) {
			result = append(result, e.BandmapEntry)
		}
	}
	if order != nil {
		slices.SortStableFunc(result, order)
	}
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
