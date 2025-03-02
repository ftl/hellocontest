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
	notifier  *Notifier
	entries   *Entries
	selection *Selection

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
	result.notifier = &Notifier{}
	result.entries = NewEntries(result.notifier, result.countEntryValue)
	result.entries.SetBands(settings.Contest().Bands())
	result.selection = NewSelection(result.entries, result.notifier, result.entryVisible)

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

	entryOnFrequency, entryOnFrequencyAvailable := m.nextVisibleEntryBy(core.BandmapByDistance(m.activeFrequency), 2, core.OnFrequency(m.activeFrequency))
	m.notifier.emitEntryOnFrequency(entryOnFrequency, entryOnFrequencyAvailable)

	bands := m.entries.Bands(m.activeBand, m.visibleBand)
	entries := m.entries.Query(nil, m.entryVisible)
	index := core.NewFrameIndex(entries)
	frame := core.BandmapFrame{
		Frequency:   m.activeFrequency,
		ActiveBand:  m.activeBand,
		VisibleBand: m.visibleBand,
		Mode:        m.activeMode,
		Bands:       bands,
		Entries:     entries,
		Index:       index,
	}

	selectedEntry, selected := m.selection.SelectedEntry()
	if selected && m.entryVisible(selectedEntry) {
		frame.SelectedEntry = selectedEntry
	}

	nearestEntry, nearestEntryFound := m.nextVisibleEntryBy(core.BandmapByDistance(m.activeFrequency), 0, core.Not(core.Or(core.OnFrequency(m.activeFrequency), core.IsWorkedSpot)))
	if nearestEntryFound {
		frame.NearestEntry = nearestEntry
	}

	highestValueEntry, highestValueEntryFound := m.nextVisibleEntryBy(core.Descending(core.BandmapByValue), 0, core.Not(core.IsWorkedSpot))
	if highestValueEntryFound {
		frame.HighestValueEntry = highestValueEntry
	}

	m.view.ShowFrame(frame)
}

func (m *Bandmap) nextVisibleEntryBy(order core.BandmapOrder, limit int, filter core.BandmapFilter) (core.BandmapEntry, bool) {
	entries := m.entries.AllBy(order)
	if limit == 0 || limit > len(entries) {
		limit = len(entries)
	}
	for i := range limit {
		entry := entries[i]
		if !m.entryVisible(entry) {
			continue
		}
		if !filter(entry) {
			continue
		}

		return entry, true
	}
	return core.BandmapEntry{}, false
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
		panic("bandmap.Bandmap.SetView must not be called with nil")
	}
	if _, ok := m.view.(*nullView); !ok {
		panic("bandmap.Bandmap.SetView was already called")
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
		m.notifier.Notify(listener)
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

func (m *Bandmap) SelectEntry(id core.BandmapEntryID) {
	m.do <- func() {
		m.selection.SelectEntry(id)
		m.update()
	}
}

func (m *Bandmap) SelectByCallsign(call callsign.Callsign) {
	m.do <- func() {
		m.selection.SelectByCallsign(call)
		m.update()
	}
}

func (m *Bandmap) GotoHighestValueEntry() {
	m.do <- func() {
		m.selection.SelectHighestValue()
		m.update()
	}
}

func (m *Bandmap) GotoNearestEntry() {
	m.do <- func() {
		m.selection.SelectNearest(m.activeFrequency)
		m.update()
	}
}

func (m *Bandmap) GotoNextEntryUp() {
	m.do <- func() {
		m.selection.SelectNextUp(m.activeFrequency)
		m.update()
	}
}

func (m *Bandmap) GotoNextEntryDown() {
	m.do <- func() {
		m.selection.SelectNextDown(m.activeFrequency)
		m.update()
	}
}

/**********
 * HELPERS
 **********/

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
