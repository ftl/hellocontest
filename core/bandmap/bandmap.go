package bandmap

import (
	"log"
	"math"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

const (
	// DefaultUpdatePeriod: the bandmap is updated with this period
	DefaultUpdatePeriod time.Duration = 250 * time.Millisecond
	// DefaultMaximumAge of entries in the bandmap
	// entries that were not seen within this period are removed from the bandmap
	DefaultMaximumAge time.Duration = 10 * time.Minute
)

type View interface {
	Show()
	Hide()

	ShowFrame(frame core.BandmapFrame)
	RevealEntry(entry core.BandmapEntry)
}

type DupeChecker interface {
	FindWorkedQSOs(callsign.Callsign, core.Band, core.Mode) ([]core.QSO, bool)
}

type Callinfo interface {
	GetInfo(callsign.Callsign, core.Band, core.Mode, []string) core.Callinfo
	GetValue(callsign.Callsign, core.Band, core.Mode, []string) (int, int, map[conval.Property]string)
}

var defaultWeights = core.BandmapWeights{
	TotalPoints: 1,
	TotalMultis: 1,
	AgeSeconds:  -0.001,
	Spots:       0.001,
	Source:      0,
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

	spots  chan core.Spot
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

		spots:  make(chan core.Spot),
		do:     make(chan func()),
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
		case spot := <-m.spots:
			m.entries.Add(spot, m.clock.Now(), m.weights)
		case command := <-m.do:
			command()
		case <-updateTicker.C:
			m.update()
		}
	}
}

func (m *Bandmap) update() {
	m.entries.CleanOut(m.maximumAge, m.clock.Now(), m.weights)

	nearestEntry, nearestEntryFound := m.nextVisibleEntry(m.activeFrequency, func(entry core.BandmapEntry) bool {
		return entry.Frequency != m.activeFrequency && entry.Source != core.WorkedSpot
	})

	bands := m.entries.Bands(m.activeBand, m.visibleBand)
	entries := m.entries.All()
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
	m.Notify(view)
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

func (m *Bandmap) ContestChanged(contest core.Contest) {
	m.do <- func() {
		m.entries.SetBands(contest.Bands())
		m.update()
	}
}

func (m *Bandmap) ScoreUpdated(score core.Score) {
	m.do <- func() {
		totalScore := score.Result()
		m.weights.TotalPoints = float64(totalScore.Points)
		m.weights.TotalMultis = float64(totalScore.Multis)
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
		if m.activeBand == m.visibleBand {
			m.visibleBand = band
		}
		m.activeBand = band
		m.update()
	}
}

func (m *Bandmap) VFOModeChanged(mode core.Mode) {
	m.do <- func() {
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

func (m *Bandmap) RemainingLifetime(index int) float64 {
	result := make(chan float64)
	m.do <- func() {
		m.entries.DoOnEntry(index, func(entry core.BandmapEntry) {
			result <- m.remainingLifetime(entry)
		})
	}
	return <-result
}

func (m *Bandmap) remainingLifetime(entry core.BandmapEntry) float64 {
	age := m.clock.Now().UTC().UnixMilli() - entry.LastHeard.UTC().UnixMilli()
	result := 1 - (float64(age) / float64(m.maximumAge.Milliseconds()))
	result = math.Max(0, math.Min(1, result))
	return result
}

func (m *Bandmap) EntryVisible(index int) bool {
	result := make(chan bool)
	m.do <- func() {
		m.entries.DoOnEntry(index, func(entry core.BandmapEntry) {
			result <- m.entryVisible(entry)
		})
	}
	return <-result
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

func (m *Bandmap) SelectEntry(index int) {
	m.do <- func() {
		m.entries.Select(index)
	}
}

func (m *Bandmap) SelectByCallsign(call callsign.Callsign) bool {
	result := make(chan bool, 1)
	callStr := call.String()
	m.do <- func() {
		foundIndex := -1
		m.entries.ForEach(func(entry core.BandmapEntry) bool {
			if entry.Call.String() == callStr && entry.Band == m.visibleBand {
				foundIndex = entry.Index
				return true
			}
			return false
		})
		m.entries.Select(foundIndex)
		result <- (foundIndex > -1)
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
		entry, found := m.nextVisibleEntry(m.activeFrequency, f)
		if found {
			m.entries.Select(entry.Index)
		}
	}
}

func (m *Bandmap) findAndSelectNextVisibleEntryBy(order core.BandmapOrder, f func(entry core.BandmapEntry) bool) {
	m.do <- func() {
		entry, found := m.nextVisibleEntryBy(order, f)
		if found {
			m.entries.Select(entry.Index)
		}
	}
}

func (m *Bandmap) nextVisibleEntry(frequency core.Frequency, f func(core.BandmapEntry) bool) (core.BandmapEntry, bool) {
	return m.nextVisibleEntryBy(core.BandmapByDistance(frequency), f)
}

func (m *Bandmap) nextVisibleEntryBy(order core.BandmapOrder, f func(core.BandmapEntry) bool) (core.BandmapEntry, bool) {
	entries := m.entries.AllBy(order)
	for i := 0; i < len(entries); i++ {
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
