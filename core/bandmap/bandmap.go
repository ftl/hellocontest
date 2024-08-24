package bandmap

import (
	"log"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

const (
	// DefaultUpdatePeriod: the bandmap is updated with this period
	DefaultUpdatePeriod time.Duration = 2 * time.Second
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
	GetValue(callsign.Callsign, core.Band, core.Mode, []string) (int, int, map[conval.Property]string)
}

var defaultWeights = core.BandmapWeights{
	AgeSeconds: -0.001,
	Spots:      0.001,
	Source:     0,
	Quality:    0.01,
}

type Bandmap struct {
	entries *Entries

	clock       core.Clock
	view        View
	dupeChecker DupeChecker

	vfo             core.VFO
	activeFrequency core.Frequency
	activeBand      core.Band
	visibleBand     core.Band
	activeMode      core.Mode

	updatePeriod time.Duration
	maximumAge   time.Duration
	weights      core.BandmapWeights

	do     chan func()
	closed chan struct{}
}

func NewDefaultBandmap(clock core.Clock, settings core.Settings, dupeChecker DupeChecker) *Bandmap {
	return NewBandmap(clock, settings, dupeChecker, DefaultUpdatePeriod, DefaultMaximumAge)
}

func NewBandmap(clock core.Clock, settings core.Settings, dupeChecker DupeChecker, updatePeriod time.Duration, maximumAge time.Duration) *Bandmap {
	result := &Bandmap{
		clock:       clock,
		view:        new(nullView),
		dupeChecker: dupeChecker,

		updatePeriod: updatePeriod,
		maximumAge:   maximumAge,
		weights:      defaultWeights,

		do:     make(chan func(), 1),
		closed: make(chan struct{}),
	}
	result.entries = NewEntries(result.countEntryValue)
	result.entries.SetBands(settings.Contest().Bands())

	go result.run()

	return result
}

func (m *Bandmap) run() {
	updateTicker := time.NewTicker(m.updatePeriod)
	defer updateTicker.Stop()
	for {
		select {
		case <-m.closed:
			return
		case command := <-m.do:
			command()
		case <-updateTicker.C:
			m.update()
		}
	}
}

func (m *Bandmap) update() {
	m.entries.CleanOut(m.maximumAge, m.clock.Now(), m.weights)

	entryOnFrequency, entryOnFrequencyAvailable := m.nextVisibleEntry(m.activeFrequency, 2, func(entry core.BandmapEntry) bool {
		return entry.OnFrequency(m.activeFrequency)
	})
	m.entries.emitEntryOnFrequency(entryOnFrequency, entryOnFrequencyAvailable)

	nearestEntry, nearestEntryFound := m.nextVisibleEntry(m.activeFrequency, 0, func(entry core.BandmapEntry) bool {
		return entry.Frequency != m.activeFrequency && entry.Source != core.WorkedSpot
	})

	bands := m.entries.Bands(m.activeBand, m.visibleBand)
	entries := m.entries.Query(nil, m.entryVisible)
	frame := core.BandmapFrame{
		Frequency:          m.activeFrequency,
		ActiveBand:         m.activeBand,
		VisibleBand:        m.visibleBand,
		Mode:               m.activeMode,
		Bands:              bands,
		Entries:            entries,
		NearestEntry:       nearestEntry,
		RevealNearestEntry: nearestEntryFound,
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
	m.do <- m.update
}

func (m *Bandmap) SetVFO(vfo core.VFO) {
	if vfo == nil {
		m.vfo = new(nullVFO)
	} else {
		m.vfo = vfo
	}
	vfo.Notify(m)
}

func (m *Bandmap) SetCallinfo(callinfo Callinfo) {
	m.do <- func() {
		m.entries.SetCallinfo(callinfo)
		m.update()
	}
}

func (m *Bandmap) Show() {
	m.view.Show()
	m.do <- m.update
}

func (m *Bandmap) Hide() {
	m.view.Hide()
}

func (m *Bandmap) ContestChanged(contest core.Contest) {
	m.do <- func() {
		m.entries.SetBands(contest.Bands())
		m.update()
	}
}

func (m *Bandmap) ScoreUpdated(_ core.Score) {
	m.do <- func() {
		m.update()
	}
}

func (m *Bandmap) VFOFrequencyChanged(frequency core.Frequency) {
	m.do <- func() {
		m.activeFrequency = frequency
	}
}

func (m *Bandmap) VFOBandChanged(band core.Band) {
	m.do <- func() {
		if band == m.activeBand {
			return
		}

		if m.activeBand == m.visibleBand {
			m.visibleBand = band
		}
		m.activeBand = band
		m.update()
	}
}

func (m *Bandmap) VFOModeChanged(mode core.Mode) {
	m.do <- func() {
		if m.activeMode == mode {
			return
		}

		m.activeMode = mode
		m.update()
	}
}

func (m *Bandmap) SetVisibleBand(band core.Band) {
	m.do <- func() {
		m.visibleBand = band
		m.update()
	}
}

func (m *Bandmap) SetActiveBand(band core.Band) {
	m.vfo.SetBand(band)
}

func (m *Bandmap) entryVisible(entry core.BandmapEntry) bool {
	return (entry.Band == m.visibleBand) && (entry.Mode == m.activeMode)
}

func (m *Bandmap) countEntryValue(entry core.BandmapEntry) bool {
	return (entry.Mode == m.activeMode) && (entry.Source != core.WorkedSpot)
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
	m.do <- func() {
		mode := spot.Mode
		if mode == core.NoMode {
			mode = m.activeMode
		}

		_, worked := m.dupeChecker.FindWorkedQSOs(spot.Call, spot.Band, mode)

		if worked {
			spot.Source = core.WorkedSpot
		}

		m.entries.Add(spot, m.clock.Now(), m.weights)
	}
}

func (m *Bandmap) AllBy(order core.BandmapOrder) []core.BandmapEntry {
	result := make(chan []core.BandmapEntry)
	m.do <- func() {
		result <- m.entries.AllBy(order)
	}
	return <-result
}

func (m *Bandmap) Query(order core.BandmapOrder, filters ...core.BandmapFilter) []core.BandmapEntry {
	result := make(chan []core.BandmapEntry)
	m.do <- func() {
		result <- m.entries.Query(order, filters...)
	}
	return <-result
}

func (m *Bandmap) SelectEntry(id core.BandmapEntryID) {
	m.do <- func() {
		m.entries.Select(id)
	}
}

func (m *Bandmap) SelectByCallsign(call callsign.Callsign) bool {
	result := make(chan bool, 1)
	callStr := call.String()
	m.do <- func() {
		foundEntryID := core.NoEntryID
		m.entries.ForEach(func(entry core.BandmapEntry) bool {
			if entry.Call.String() == callStr && entry.Band == m.visibleBand {
				foundEntryID = entry.ID
				return true
			}
			return false
		})
		m.entries.Select(foundEntryID)
		result <- (foundEntryID > core.NoEntryID)
	}
	return <-result
}

func (m *Bandmap) GotoHighestValueEntry() {
	m.findAndSelectNextVisibleEntryBy(core.BandmapByDescendingValue, func(entry core.BandmapEntry) bool {
		return entry.Frequency != m.activeFrequency && entry.Source != core.WorkedSpot
	})
}

func (m *Bandmap) GotoNearestEntry() {
	m.findAndSelectNextVisibleEntry(func(entry core.BandmapEntry) bool {
		return entry.Frequency != m.activeFrequency && entry.Source != core.WorkedSpot
	})
}

func (m *Bandmap) GotoNextEntryUp() {
	m.findAndSelectNextVisibleEntry(func(entry core.BandmapEntry) bool {
		return entry.Frequency > m.activeFrequency && entry.Source != core.WorkedSpot
	})
}

func (m *Bandmap) GotoNextEntryDown() {
	m.findAndSelectNextVisibleEntry(func(entry core.BandmapEntry) bool {
		return entry.Frequency < m.activeFrequency && entry.Source != core.WorkedSpot
	})
}

func (m *Bandmap) findAndSelectNextVisibleEntry(f func(entry core.BandmapEntry) bool) {
	m.do <- func() {
		entry, found := m.nextVisibleEntry(m.activeFrequency, 0, f)
		if found {
			m.entries.Select(entry.ID)
		}
	}
}

func (m *Bandmap) findAndSelectNextVisibleEntryBy(order core.BandmapOrder, f func(entry core.BandmapEntry) bool) {
	m.do <- func() {
		entry, found := m.nextVisibleEntryBy(order, 0, f)
		if found {
			m.entries.Select(entry.ID)
		}
	}
}

func (m *Bandmap) nextVisibleEntry(frequency core.Frequency, limit int, f func(core.BandmapEntry) bool) (core.BandmapEntry, bool) {
	return m.nextVisibleEntryBy(core.BandmapByDistance(frequency), limit, f)
}

func (m *Bandmap) nextVisibleEntryBy(order core.BandmapOrder, limit int, f func(core.BandmapEntry) bool) (core.BandmapEntry, bool) {
	entries := m.entries.AllBy(order)
	if limit == 0 || limit > len(entries) {
		limit = len(entries)
	}
	for i := 0; i < limit; i++ {
		entry := entries[i]
		if !m.entryVisible(entry) {
			continue
		}
		if !f(entry) {
			continue
		}

		return entry, true
	}
	return core.BandmapEntry{}, false
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

func (v *nullView) Show()                         {}
func (v *nullView) Hide()                         {}
func (v *nullView) ShowFrame(core.BandmapFrame)   {}
func (v *nullView) RevealEntry(core.BandmapEntry) {}

type nullVFO struct{}

func (n *nullVFO) Notify(any)                  {}
func (n *nullVFO) Active() bool                { return false }
func (n *nullVFO) Refresh()                    {}
func (n *nullVFO) SetFrequency(core.Frequency) {}
func (n *nullVFO) SetBand(core.Band)           {}
func (n *nullVFO) SetMode(core.Mode)           {}

type nullCallinfo struct{}

func (n *nullCallinfo) GetInfo(callsign.Callsign, core.Band, core.Mode, []string) core.Callinfo {
	return core.Callinfo{}
}
func (n *nullCallinfo) GetValue(callsign.Callsign, core.Band, core.Mode, []string) (int, int, map[conval.Property]string) {
	return 0, 0, nil
}
