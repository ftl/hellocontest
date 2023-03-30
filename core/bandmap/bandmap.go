package bandmap

import (
	"log"
	"math"
	"time"

	"github.com/ftl/hamradio/callsign"
	"golang.org/x/exp/slices"

	"github.com/ftl/hellocontest/core"
)

const (
	// DefaultUpdatePeriod: the bandmap is updated with this period
	DefaultUpdatePeriod time.Duration = 10 * time.Second
	// DefaultMaximumAge of entries in the bandmap
	// entries that were not seen within this period are removed from the bandmap
	DefaultMaximumAge time.Duration = 10 * time.Minute
)

type View interface {
	Show(frame BandmapFrame)
}

type DupeChecker interface {
	FindWorkedQSOs(callsign.Callsign, core.Band, core.Mode) ([]core.QSO, bool)
}

type BandmapFrame struct {
	VFO     core.Frequency
	Entries []Entry
}

type Bandmap struct {
	entries *Entries
	clock   core.Clock

	dupeChecker DupeChecker
	view        View

	updatePeriod time.Duration
	maximumAge   time.Duration

	spots  chan core.Spot
	do     chan bandmapCommand
	closed chan struct{}
}

type bandmapCommand func()

func NewDefaultBandmap(clock core.Clock, dupeChecker DupeChecker) *Bandmap {
	return NewBandmap(clock, dupeChecker, DefaultUpdatePeriod, DefaultMaximumAge)
}

func NewBandmap(clock core.Clock, dupeChecker DupeChecker, updatePeriod time.Duration, maximumAge time.Duration) *Bandmap {
	result := &Bandmap{
		entries: NewEntries(),
		clock:   clock,

		dupeChecker: dupeChecker,

		updatePeriod: updatePeriod,
		maximumAge:   maximumAge,

		spots:  make(chan core.Spot),
		do:     make(chan bandmapCommand),
		closed: make(chan struct{}),
	}

	go result.run()

	return result
}

func (m *Bandmap) run() {
	updateTicker := time.NewTicker(m.updatePeriod)
	for {
		select {
		case <-m.closed:
			return
		case spot := <-m.spots:
			m.entries.Add(spot)
			m.update()
		case command := <-m.do:
			command()
		case <-updateTicker.C:
			m.update()
		}
	}
}

func (m *Bandmap) update() {
	m.entries.CleanOut(m.maximumAge, m.clock.Now())

	// TODO: calculate the current frame
	frame := BandmapFrame{}

	if m.view != nil {
		m.view.Show(frame)
	}
}

func (m *Bandmap) Close() {
	select {
	case <-m.closed:
		return
	default:
		close(m.closed)
	}
}

func (m *Bandmap) SetView(v View) {
	m.view = v
	m.do <- m.update
}

func (m *Bandmap) Notify(listener any) {
	m.do <- func() {
		m.entries.Notify(listener)
	}
}

func (m *Bandmap) Clear() {
	m.do <- func() {
		m.entries.Clear()
	}
}

func (m *Bandmap) Add(spot core.Spot) {
	_, worked := m.dupeChecker.FindWorkedQSOs(spot.Call, spot.Band, spot.Mode)
	if worked {
		spot.Source = core.WorkedSpot
	}
	m.spots <- spot
}

func (m *Bandmap) AllByProximity(f core.Frequency) []core.BandmapEntry {
	result := make(chan []core.BandmapEntry)
	m.do <- func() {
		result <- toBandmapEntries(m.entries.AllByProximity(f))
	}
	return <-result
}

func (m *Bandmap) AllByDistance(f core.Frequency) []core.BandmapEntry {
	result := make(chan []core.BandmapEntry)
	m.do <- func() {
		result <- toBandmapEntries(m.entries.AllByDistance(f))
	}
	return <-result
}

const (
	// spots within this distance to an entry's frequency will be added to the entry
	spotFrequencyDeltaThreshold float64 = 25
	// frequencies within this distance to an entry's frequency will be recognized as "in proximity"
	spotFrequencyProximityThreshold float64 = 500
	// the frequency of an entry is aligend to this is grid of accuracy
	spotFrequencyStep float64 = 10
)

type Entry struct {
	core.BandmapEntry

	spots []core.Spot
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
	if spot.Mode != e.Mode {
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
	if e.Source > spot.Source {
		e.Source = spot.Source
	}

	return true
}

func (e *Entry) RemoveSpotsBefore(timestamp time.Time) bool {
	e.spots = filterSlice(e.spots, func(s core.Spot) bool {
		return !s.Time.Before(timestamp)
	})

	e.update()

	return len(e.spots) > 0
}

// ProximityFactor increases the closer the given frequency is to this entry's frequency.
// 0.0 = not in proximity, 1.0 = exactly on frequency
func (e *Entry) ProximityFactor(f core.Frequency) float64 {
	frequencyDelta := math.Abs(float64(e.Frequency - f))
	if frequencyDelta > spotFrequencyProximityThreshold {
		return 0.0
	}

	return 1.0 - (frequencyDelta / spotFrequencyProximityThreshold)
}

func (e *Entry) update() {
	e.updateFrequency()

	lastHeard := time.Time{}
	source := core.MaxSpotType
	for _, s := range e.spots {
		if lastHeard.Before(s.Time) {
			lastHeard = s.Time
		}
		if source > s.Source {
			source = s.Source
		}
	}
	e.LastHeard = lastHeard
	e.Source = source
}

func (e *Entry) updateFrequency() {
	if len(e.spots) == 0 {
		e.Frequency = 0
		return
	}

	var sum core.Frequency
	for _, s := range e.spots {
		sum += s.Frequency
	}
	downscaledMean := float64(sum) / (float64(len(e.spots)) * spotFrequencyStep)
	roundedMean := math.RoundToEven(downscaledMean)
	e.Frequency = core.Frequency(roundedMean * spotFrequencyStep)
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

type Entries struct {
	entries []*Entry

	listeners []any
}

func NewEntries() *Entries {
	result := new(Entries)
	result.Clear()
	return result
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

func (l *Entries) Clear() {
	l.entries = make([]*Entry, 0, 100)
}

func (l *Entries) Len() int {
	return len(l.entries)
}

func (l *Entries) Add(spot core.Spot) {
	for _, e := range l.entries {
		if e.Add(spot) {
			l.emitEntryUpdated(*e)
			return
		}
	}
	newEntry := NewEntry(spot)
	l.entries = append(l.entries, &newEntry)
	l.emitEntryAdded(newEntry)
}

func (l *Entries) CleanOut(maximumAge time.Duration, now time.Time) {
	deadline := now.Add(-maximumAge)
	l.entries = filterSlice(l.entries, func(e *Entry) bool {
		matches := e.RemoveSpotsBefore(deadline)
		if !matches {
			l.emitEntryRemoved(*e)
		}
		return matches
	})
}

func (l *Entries) sorted(less func(Entry, Entry) bool) []Entry {
	result := make([]Entry, len(l.entries))
	for i, e := range l.entries {
		result[i] = *e
	}
	slices.SortStableFunc(result, less)
	return result
}

func (l *Entries) AllByFrequency() []Entry {
	return l.sorted(func(a, b Entry) bool {
		return a.Frequency < b.Frequency
	})
}

func (l *Entries) AllByProximity(f core.Frequency) []Entry {
	return l.sorted(func(a, b Entry) bool {
		return a.ProximityFactor(f) > b.ProximityFactor(f)
	})
}

func (l *Entries) AllByDistance(f core.Frequency) []Entry {
	return l.sorted(func(a, b Entry) bool {
		deltaA := math.Abs(float64(a.Frequency - f))
		deltaB := math.Abs(float64(b.Frequency - f))
		return deltaA < deltaB
	})
}

func toBandmapEntries(entries []Entry) []core.BandmapEntry {
	result := make([]core.BandmapEntry, len(entries))
	for i, entry := range entries {
		result[i] = entry.BandmapEntry
	}
	return result
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

type Logger struct{}

func (l *Logger) EntryAdded(e core.BandmapEntry) {
	log.Printf("Bandmap entry added: %v", e)
}

func (l *Logger) EntryUpdated(e core.BandmapEntry) {
	log.Printf("Bandmap entry updated: %v", e)
}

func (l *Logger) EntryRemoved(e core.BandmapEntry) {
	log.Printf("Bandmap entry removed: %v", e)
}
