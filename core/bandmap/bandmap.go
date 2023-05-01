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
	Show()
	Hide()

	ShowFrame(frame core.BandmapFrame)
}

type DupeChecker interface {
	FindWorkedQSOs(callsign.Callsign, core.Band, core.Mode) ([]core.QSO, bool)
}

type Callinfo interface {
	GetInfo(callsign.Callsign, core.Band, core.Mode, []string) core.Callinfo
	GetValue(callsign.Callsign, core.Band, core.Mode, []string) (int, int)
}

type Bandmap struct {
	entries      *Entries
	selectedMode core.Mode

	clock            core.Clock
	view             View
	dupeChecker      DupeChecker
	currentFrequency core.Frequency

	updatePeriod time.Duration
	maximumAge   time.Duration

	spots  chan core.Spot
	do     chan func()
	closed chan struct{}
}

func NewDefaultBandmap(clock core.Clock, dupeChecker DupeChecker) *Bandmap {
	return NewBandmap(clock, dupeChecker, DefaultUpdatePeriod, DefaultMaximumAge)
}

func NewBandmap(clock core.Clock, dupeChecker DupeChecker, updatePeriod time.Duration, maximumAge time.Duration) *Bandmap {
	result := &Bandmap{
		entries: NewEntries(),

		clock:       clock,
		view:        new(nullView),
		dupeChecker: dupeChecker,

		updatePeriod: updatePeriod,
		maximumAge:   maximumAge,

		spots:  make(chan core.Spot),
		do:     make(chan func()),
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

	entries := m.entries.All()
	frame := core.BandmapFrame{
		VFO:     m.currentFrequency,
		Entries: entries,
	}
	m.view.ShowFrame(frame)
}

func (m *Bandmap) Close() {
	select {
	case <-m.closed:
		return
	default:
		close(m.closed)
	}
}

func (m *Bandmap) SetView(view View) {
	if view == nil {
		m.view = new(nullView)
		return
	}

	m.view = view
	m.Notify(view)
	m.do <- m.update
}

func (m *Bandmap) SetCallinfo(callinfo Callinfo) {
	m.entries.SetCallinfo(callinfo)
	m.do <- m.update
}

func (m *Bandmap) Show() {
	m.view.Show()
	m.do <- m.update
}

func (m *Bandmap) Hide() {
	m.view.Hide()
}

func (m *Bandmap) SetMode(mode core.Mode) {
	m.selectedMode = mode
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
	mode := spot.Mode
	if mode == core.NoMode {
		mode = m.selectedMode
	}
	_, worked := m.dupeChecker.FindWorkedQSOs(spot.Call, spot.Band, mode)
	if worked {
		spot.Source = core.WorkedSpot
	}
	m.spots <- spot
}

func (m *Bandmap) AllBy(order core.BandmapOrder) []core.BandmapEntry {
	result := make(chan []core.BandmapEntry)
	m.do <- func() {
		result <- m.entries.AllBy(order)
	}
	return <-result
}

const (
	// spots within this distance to an entry's frequency will be added to the entry
	spotFrequencyDeltaThreshold float64 = 500
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
	entries  []*Entry
	order    core.BandmapOrder
	callinfo Callinfo

	listeners []any
}

func NewEntries() *Entries {
	result := &Entries{
		order:    core.BandmapByFrequency,
		callinfo: new(nullCallinfo),
	}
	result.Clear()
	return result
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
	if newEntry.Call.String() != "" {
		newEntry.Info = l.callinfo.GetInfo(newEntry.Call, newEntry.Band, newEntry.Mode, []string{})
		log.Printf("got callinfo for %s on band %s and mode %s : %+v", newEntry.Call, newEntry.Band, newEntry.Mode, newEntry.Info)
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

func (l *Entries) CleanOut(maximumAge time.Duration, now time.Time) {
	deadline := now.Add(-maximumAge)
	removedEntries := make([]Entry, 0, len(l.entries))
	l.entries = filterSlice(l.entries, func(e *Entry) bool {
		matches := e.RemoveSpotsBefore(deadline)
		if !matches {
			removedEntries = append(removedEntries, *e)
		}
		return matches
	})
	for i, e := range l.entries {
		e.Index = i
		l.entries[i] = e
	}

	for i, e := range removedEntries {
		e.Index -= i
		l.emitEntryRemoved(e)
	}
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

type nullView struct{}

func (v *nullView) Show()                       {}
func (v *nullView) Hide()                       {}
func (v *nullView) ShowFrame(core.BandmapFrame) {}

type nullCallinfo struct{}

func (n *nullCallinfo) GetInfo(callsign.Callsign, core.Band, core.Mode, []string) core.Callinfo {
	return core.Callinfo{}
}
func (n *nullCallinfo) GetValue(callsign.Callsign, core.Band, core.Mode, []string) (int, int) {
	return 0, 0
}
