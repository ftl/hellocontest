package bandmap

import (
	"log"
	"math"
	"slices"
	"time"

	"github.com/ftl/conval"

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
	FindWorkedQSOs(core.Callsign, core.Band, core.Mode) ([]core.QSO, bool)
}

type Callinfo interface {
	GetInfo(core.Callsign, core.Band, core.Mode, []string) core.Callinfo
	UpdateValue(*core.Callinfo, core.Band, core.Mode) bool
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
	markers   *Markers
	selection *Selection
	filter    *SpotFilter

	clock       core.Clock
	view        View
	dupeChecker DupeChecker
	asyncRunner core.AsyncRunner
	vfo         core.VFO

	activeMode      core.Mode
	activeFrequency core.Frequency
	activeBand      core.Band
	visibleBand     core.Band

	vfoBands       [core.VFOCount]core.Band
	vfoModes       [core.VFOCount]core.Mode
	vfoFrequencies [core.VFOCount]core.Frequency

	updatePeriod time.Duration
	maximumAge   time.Duration
	weights      core.BandmapWeights
	bandRule     conval.BandRule
	qtcsEnabled  bool

	do     chan func()
	closed chan struct{}
}

func NewDefaultBandmap(clock core.Clock, settings core.Settings, dupeChecker DupeChecker, filterStore SpotFilterStore, asyncRunner core.AsyncRunner) *Bandmap {
	return NewBandmap(clock, settings, dupeChecker, filterStore, asyncRunner, DefaultUpdatePeriod, DefaultMaximumAge)
}

func NewBandmap(clock core.Clock, settings core.Settings, dupeChecker DupeChecker, filterStore SpotFilterStore, asyncRunner core.AsyncRunner, updatePeriod time.Duration, maximumAge time.Duration) *Bandmap {
	result := &Bandmap{
		clock:       clock,
		view:        new(nullView),
		dupeChecker: dupeChecker,
		asyncRunner: asyncRunner,

		updatePeriod: updatePeriod,
		maximumAge:   maximumAge,
		weights:      defaultWeights,

		do:     make(chan func(), 1),
		closed: make(chan struct{}),
	}
	result.notifier = &Notifier{
		asyncRunner: result.asyncRunner,
	}
	result.entries = NewEntries(result.notifier, result.countEntryValue)
	result.markers = NewMarkers(result.entries)
	result.entries.SetBands(settings.Contest().Bands())
	result.filter = NewSpotFilter(MainSpotFilterID, filterStore)
	result.filter.SetContestBandsAndModes(settings.Contest().Bands(), settings.Contest().Modes())
	// the filter is only ever changed from inside the bandmap goroutine, therefore the
	// update runs directly. Posting to the do channel would block the goroutine that
	// drains it.
	result.filter.OnChanged(result.update)
	result.selection = NewSelection(result.entries, result.markers, result.notifier, result)

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

	focusedFrequency := m.focusedFrequency()
	// every VFO has its own entry on frequency, because the operator may work a spot on
	// any of them
	for vfo := range core.VFOCount {
		entryOnFrequency, entryOnFrequencyAvailable := m.nextEntryOnFrequency(core.VFOID(vfo))
		m.notifier.emitEntryOnFrequency(core.VFOID(vfo), entryOnFrequency, entryOnFrequencyAvailable)

		markerOnFrequency, markerOnFrequencyAvailable := m.markerOnFrequency(core.VFOID(vfo))
		m.notifier.emitMarkerOnFrequency(core.VFOID(vfo), markerOnFrequency, markerOnFrequencyAvailable)
	}

	m.notifier.emitMarkersChanged(m.markers.All())

	bands := m.entries.Bands(m.activeBand, m.visibleBand)
	m.notifier.emitBandsChanged(bands)
	rows := m.rows()
	index := core.NewFrameIndex(rows)
	frame := core.BandmapFrame{
		Frequency:   focusedFrequency,
		ActiveBand:  m.activeBand,
		VisibleBand: m.visibleBand,
		Mode:        m.activeMode,
		Bands:       bands,
		Rows:        rows,
		Index:       index,
		QTCsEnabled: m.qtcsEnabled,
		Filter:      m.filter.Frame(),
	}

	selectedEntry, selected := m.selection.SelectedEntry()
	if selected && m.filter.Matches(selectedEntry) {
		frame.SelectedEntry = selectedEntry
	}

	nearestEntry, nearestEntryFound := m.nextVisibleEntryBy(core.BandmapByDistance(focusedFrequency), 0, core.Not(core.Or(core.OnFrequency(focusedFrequency), core.IsWorkedSpot)))
	if nearestEntryFound {
		frame.NearestEntry = nearestEntry
	}

	highestValueEntry, highestValueEntryFound := m.nextVisibleEntryBy(core.Descending(core.BandmapByValue), 0, core.Not(core.IsWorkedSpot))
	if highestValueEntryFound {
		frame.HighestValueEntry = highestValueEntry
	}

	m.asyncRunner(func() {
		m.view.ShowFrame(frame)
	})
}

func (m *Bandmap) rows() []core.BandmapRow {
	state := m.filter.State()
	markers := m.markers.Ordered(m.filter.MatchesMarker, state.SortBy, state.Descending)
	entries := m.entries.Query(m.filter.Order(), m.filter.Matches)

	result := make([]core.BandmapRow, 0, len(markers)+len(entries))
	for _, marker := range markers {
		result = append(result, core.MarkerBandmapRow(marker))
	}
	for _, entry := range entries {
		result = append(result, core.SpotBandmapRow(entry))
	}

	if !blendsMarkersWithSpots(state.SortBy) {
		// a marker has no callsign and no value, therefore it cannot take part in those
		// orders and it keeps a block at the top
		return result
	}

	slices.SortStableFunc(result, func(a, b core.BandmapRow) int {
		return compareRows(a, b, state.SortBy, state.Descending)
	})
	return result
}

func (m *Bandmap) markerOnFrequency(vfo core.VFOID) (core.BandmapMarker, bool) {
	if !validVFO(vfo) {
		return core.BandmapMarker{}, false
	}
	frequency := m.vfoFrequencies[vfo]
	if frequency == 0 {
		return core.BandmapMarker{}, false
	}

	var result core.BandmapMarker
	found := false
	for _, marker := range m.markers.All() {
		if !marker.OnFrequency(frequency) {
			continue
		}
		// the closest marker wins, a marker of the CQ frequency before a generic one
		if found && math.Abs(result.ProximityFactor(frequency)) >= math.Abs(marker.ProximityFactor(frequency)) {
			continue
		}
		result = marker
		found = true
	}
	return result, found
}

func (m *Bandmap) nextEntryOnFrequency(vfo core.VFOID) (core.BandmapEntry, bool) {
	if !validVFO(vfo) {
		return core.BandmapEntry{}, false
	}
	frequency := m.vfoFrequencies[vfo]
	return m.nextEntryBy(core.BandmapByDistance(frequency), 2, m.entryVisibleOn(vfo), core.OnFrequency(frequency))
}

func (m *Bandmap) nextVisibleEntryBy(order core.BandmapOrder, limit int, filter core.BandmapFilter) (core.BandmapEntry, bool) {
	return m.nextEntryBy(order, limit, m.entryVisible, filter)
}

func (m *Bandmap) nextEntryBy(order core.BandmapOrder, limit int, visible core.BandmapFilter, filter core.BandmapFilter) (core.BandmapEntry, bool) {
	entries := m.entries.AllBy(order)
	if limit == 0 || limit > len(entries) {
		limit = len(entries)
	}
	for i := range limit {
		entry := entries[i]
		if !visible(entry) {
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
	// the VFOs register themselves as listeners, see core/app
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
		if contest.Definition != nil {
			m.bandRule = contest.Definition.Scoring.QSOBandRule
		}
		m.qtcsEnabled = contest.EnableQTCs
		m.entries.SetBands(contest.Bands())
		m.filter.SetContestBandsAndModes(contest.Bands(), contest.Modes())
		m.update()
	}
}

func (m *Bandmap) ScoreChanged(_ core.Score) {
	m.do <- func() {
		m.update()
	}
}

func (m *Bandmap) VFOFrequencyChanged(vfo core.VFOID, frequency core.Frequency) {
	m.do <- func() {
		if validVFO(vfo) {
			m.vfoFrequencies[vfo] = frequency
		}
		if vfo == core.VFO1 {
			m.activeFrequency = frequency
		}
		// the frame is not updated with every frequency change, this creates too much load
	}
}

func (m *Bandmap) VFOBandChanged(vfo core.VFOID, band core.Band) {
	m.do <- func() {
		if validVFO(vfo) {
			m.vfoBands[vfo] = band
		}
		m.filter.VFOBandChanged(vfo, band)

		// the active band and the visible band belong to VFO1
		if vfo != core.VFO1 || band == m.activeBand {
			return
		}
		if m.activeBand == m.visibleBand {
			m.visibleBand = band
		}
		m.activeBand = band
		m.update()
	}
}

func (m *Bandmap) VFOModeChanged(vfo core.VFOID, mode core.Mode) {
	m.do <- func() {
		if validVFO(vfo) {
			m.vfoModes[vfo] = mode
		}
		m.filter.VFOModeChanged(vfo, mode)

		// the active mode belongs to VFO1
		if vfo != core.VFO1 || m.activeMode == mode {
			return
		}
		m.activeMode = mode
		m.update()
	}
}

func (m *Bandmap) FocusedVFOChanged(vfo core.VFOID) {
	m.do <- func() {
		m.filter.FocusedVFOChanged(vfo)
		// the navigation and the marks in the frame follow the focused VFO
		m.update()
	}
}

func (m *Bandmap) RadioChanged(name string, singleVFO bool) {
	m.do <- func() {
		m.filter.RadioChanged(name, singleVFO)
	}
}

func (m *Bandmap) SetFilterBand(band core.SpotFilterBand) {
	m.do <- func() {
		m.filter.SetBand(band)
	}
}

func (m *Bandmap) SetFilterMode(mode core.SpotFilterMode) {
	m.do <- func() {
		m.filter.SetMode(mode)
	}
}

func (m *Bandmap) SetFilterSort(column core.SpotSortColumn, descending bool) {
	m.do <- func() {
		m.filter.SetSort(column, descending)
	}
}

func (m *Bandmap) SetFilterFolded(folded bool) {
	m.do <- func() {
		m.filter.SetFolded(folded)
	}
}

func (m *Bandmap) SelectableFilter() core.BandmapFilter {
	return m.filter.Matches
}

func (m *Bandmap) NavigationFilter() core.BandmapFilter {
	return m.entryVisible
}

func (m *Bandmap) MarkerNavigationFilter() core.BandmapMarkerFilter {
	return m.markerVisible
}

func (m *Bandmap) TargetVFO(entry core.BandmapEntry) core.VFOID {
	return m.filter.TargetVFO(entry)
}

func (m *Bandmap) TargetVFOForMarker(band core.Band) core.VFOID {
	return m.filter.TargetVFOForMarker(band)
}

func (m *Bandmap) FocusedVFO() core.VFOID {
	return m.filter.FocusedVFO()
}

func (m *Bandmap) entryVisible(entry core.BandmapEntry) bool {
	// the keyboard navigation tunes the focused VFO, therefore it stays on the band and
	// in the mode of that VFO
	return m.entryVisibleOn(m.filter.FocusedVFO())(entry)
}

func (m *Bandmap) markerVisible(marker core.BandmapMarker) bool {
	if marker.Kind == core.CQMarker {
		// the CQ frequency has its own action, the navigation does not stop there
		return false
	}
	vfo := m.filter.FocusedVFO()
	if !validVFO(vfo) {
		return false
	}
	// a marker has no mode, therefore only the band decides. A band that the rig did not
	// report yet restricts nothing.
	band := m.vfoBands[vfo]
	return band == core.NoBand || marker.Band == band
}

func (m *Bandmap) entryVisibleOn(vfo core.VFOID) core.BandmapFilter {
	return func(entry core.BandmapEntry) bool {
		if !validVFO(vfo) {
			return false
		}
		// a band or a mode that the rig did not report yet restricts nothing, the spot
		// filter uses the same rule
		band := m.vfoBands[vfo]
		mode := m.vfoModes[vfo]
		return (band == core.NoBand || entry.Band == band) &&
			(mode == core.NoMode || entry.Mode == mode)
	}
}

func (m *Bandmap) focusedBand() core.Band {
	return m.vfoBands[m.filter.FocusedVFO()]
}

func (m *Bandmap) focusedMode() core.Mode {
	return m.vfoModes[m.filter.FocusedVFO()]
}

func (m *Bandmap) focusedFrequency() core.Frequency {
	return m.vfoFrequencies[m.filter.FocusedVFO()]
}

func (m *Bandmap) countEntryValue(entry core.BandmapEntry) bool {
	if entry.Source == core.WorkedSpot {
		return false
	}
	// a mode that the rig did not report yet restricts nothing
	return m.activeMode == core.NoMode || entry.Mode == m.activeMode
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

		if !spot.IsWorked() {
			_, worked := m.dupeChecker.FindWorkedQSOs(spot.Call, spot.Band, mode)
			if worked {
				spot.Source = core.WorkedSpot
			}
		}

		m.entries.Add(spot, m.clock.Now(), m.weights)

		if spot.IsWorked() {
			band := spot.Band
			switch m.bandRule {
			case conval.Once:
				band, mode = core.NoBand, core.NoMode
			case conval.OncePerBand:
				mode = core.NoMode
			case conval.OncePerBandAndMode:
				// band, mode = band, mode
			default:
				band, mode = core.NoBand, core.NoMode
			}
			m.entries.MarkAsWorked(spot.Call, band, mode)
		}
	}
}

func (m *Bandmap) MacroSent(vfo core.VFOID, workmode core.Workmode, index int) {
	m.do <- func() {
		// only VFO1 works in the running workmode, and only the CQ macro marks a
		// running frequency
		if vfo != core.VFO1 || workmode != core.Run || index != core.CQMacroIndex {
			return
		}
		m.setCQMarker()
	}
}

func (m *Bandmap) setCQMarker() {
	frequency := m.vfoFrequencies[core.VFO1]
	if frequency == 0 {
		// the rig did not report a frequency yet
		return
	}
	m.markers.SetCQ(frequency, m.vfoBands[core.VFO1], m.clock.Now())
	m.update()
}

func (m *Bandmap) SelectMarker(id core.BandmapEntryID) {
	m.do <- func() {
		marker, found := m.markers.Get(id)
		if !found {
			return
		}
		m.selectMarker(marker)
	}
}

func (m *Bandmap) GotoNumberedMarker(number int) {
	m.do <- func() {
		marker, found := m.markers.Numbered(number)
		if !found {
			return
		}
		m.selectMarker(marker)
	}
}

func (m *Bandmap) GotoCQMarker() {
	m.do <- func() {
		marker, found := m.markers.CQ()
		if !found {
			return
		}
		// the CQ frequency belongs to VFO1 and to the running workmode, therefore this
		// is no ordinary marker selection
		m.notifier.emitCQMarkerSelected(marker)
	}
}

func (m *Bandmap) selectMarker(marker core.BandmapMarker) {
	m.selection.SelectMarker(marker)
	m.update()
}

func (m *Bandmap) MarkWithNextNumber(frequency core.Frequency, band core.Band) {
	m.do <- func() {
		m.markers.MarkWithNextNumber(frequency, band, m.clock.Now())
		m.update()
	}
}

func (m *Bandmap) MarkWithNumber(number int, frequency core.Frequency, band core.Band) {
	m.do <- func() {
		m.markers.MarkWithNumber(number, frequency, band, m.clock.Now())
		m.update()
	}
}

func (m *Bandmap) MarkWithText(text string, frequency core.Frequency, band core.Band) {
	m.do <- func() {
		m.markers.MarkWithText(text, frequency, band, m.clock.Now())
		m.update()
	}
}

func (m *Bandmap) AddNumberedMarker(number int, frequency core.Frequency, band core.Band) {
	m.do <- func() {
		m.markers.AddNumbered(number, frequency, band, m.clock.Now())
		m.update()
	}
}

func (m *Bandmap) AddTextMarker(text string, frequency core.Frequency, band core.Band) {
	m.do <- func() {
		m.markers.AddText(text, frequency, band, m.clock.Now())
		m.update()
	}
}

func (m *Bandmap) DeleteMarkerOnFrequency() {
	m.do <- func() {
		_, found := m.markers.RemoveOnFrequency(m.focusedFrequency())
		if !found {
			return
		}
		m.update()
	}
}

func (m *Bandmap) RemoveMarker(id core.BandmapEntryID) {
	m.do <- func() {
		m.markers.Remove(id)
		m.update()
	}
}

func (m *Bandmap) ClearMarkers() {
	m.do <- func() {
		m.markers.Clear()
		m.update()
	}
}

func (m *Bandmap) QTCAdded(qtc core.QTC) {
	m.do <- func() {
		m.entries.RefreshCallinfo(qtc.TheirCallsign, m.clock.Now(), m.weights)
	}
}

// Remove takes a spot back, for a source that knows that the station is gone.
func (m *Bandmap) Remove(spot core.Spot) {
	m.do <- func() {
		if !m.entries.Remove(spot) {
			return
		}
		m.update()
	}
}

func (m *Bandmap) SelectEntry(id core.BandmapEntryID) {
	m.do <- func() {
		m.selection.SelectEntry(id)
		m.update()
	}
}

func (m *Bandmap) SelectByCallsign(call core.Callsign) {
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
		m.selection.SelectNearest(m.focusedFrequency())
		m.update()
	}
}

func (m *Bandmap) GotoNextEntryUp() {
	m.do <- func() {
		m.selection.SelectNextUp(m.focusedFrequency())
		m.update()
	}
}

func (m *Bandmap) GotoNextEntryDown() {
	m.do <- func() {
		m.selection.SelectNextDown(m.focusedFrequency())
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

func (n *nullVFO) Name() string                                                          { return "" }
func (n *nullVFO) Notify(any)                                                            {}
func (n *nullVFO) Active() bool                                                          { return false }
func (n *nullVFO) Refresh()                                                              {}
func (n *nullVFO) SetFrequency(core.Frequency)                                           {}
func (n *nullVFO) ShiftFrequency(core.Frequency)                                         {}
func (n *nullVFO) SetBand(core.Band)                                                     {}
func (n *nullVFO) SetMode(core.Mode)                                                     {}
func (n *nullVFO) IncrementalTuningActive(core.IncrementalTuningKind) bool               { return false }
func (n *nullVFO) SetIncrementalTuningActive(core.IncrementalTuningKind, bool)           {}
func (n *nullVFO) SetIncrementalTuning(core.IncrementalTuningKind, bool, core.Frequency) {}
func (n *nullVFO) ShiftIncrementalTuning(core.IncrementalTuningKind, core.Frequency)     {}
func (n *nullVFO) ToggleAvailableIncrementalTuning()                                     {}
func (n *nullVFO) ShiftAvailableIncrementalTuning(core.Frequency)                        {}

type nullCallinfo struct{}

func (n *nullCallinfo) GetInfo(core.Callsign, core.Band, core.Mode, []string) core.Callinfo {
	return core.Callinfo{}
}
func (n *nullCallinfo) UpdateValue(*core.Callinfo, core.Band, core.Mode) bool {
	return false
}
