package bandmap

import (
	"log"
	"math"
	"slices"
	"time"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core/model"
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

	ShowFrame(frame model.BandmapFrame)
}

type DupeChecker interface {
	FindWorkedQSOs(model.Callsign, model.Band, model.Mode) ([]model.QSO, bool)
}

type Callinfo interface {
	GetInfo(model.Callsign, model.Band, model.Mode, []string) model.Callinfo
	UpdateValue(*model.Callinfo, model.Band, model.Mode) bool
}

var defaultWeights = model.BandmapWeights{
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

	clock       model.Clock
	view        View
	dupeChecker DupeChecker
	asyncRunner model.AsyncRunner
	vfo         model.VFO

	activeMode      model.Mode
	activeFrequency model.Frequency
	activeBand      model.Band
	visibleBand     model.Band

	vfoBands       [model.VFOCount]model.Band
	vfoModes       [model.VFOCount]model.Mode
	vfoFrequencies [model.VFOCount]model.Frequency

	updatePeriod time.Duration
	maximumAge   time.Duration
	weights      model.BandmapWeights
	bandRule     conval.BandRule
	qtcsEnabled  bool

	do     chan func()
	closed chan struct{}
}

func NewDefaultBandmap(clock model.Clock, settings model.Settings, dupeChecker DupeChecker, filterStore SpotFilterStore, asyncRunner model.AsyncRunner) *Bandmap {
	return NewBandmap(clock, settings, dupeChecker, filterStore, asyncRunner, DefaultUpdatePeriod, DefaultMaximumAge)
}

func NewBandmap(clock model.Clock, settings model.Settings, dupeChecker DupeChecker, filterStore SpotFilterStore, asyncRunner model.AsyncRunner, updatePeriod time.Duration, maximumAge time.Duration) *Bandmap {
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
	for vfo := range model.VFOCount {
		entryOnFrequency, entryOnFrequencyAvailable := m.nextEntryOnFrequency(model.VFOID(vfo))
		m.notifier.emitEntryOnFrequency(model.VFOID(vfo), entryOnFrequency, entryOnFrequencyAvailable)

		markerOnFrequency, markerOnFrequencyAvailable := m.markerOnFrequency(model.VFOID(vfo))
		m.notifier.emitMarkerOnFrequency(model.VFOID(vfo), markerOnFrequency, markerOnFrequencyAvailable)
	}

	m.notifier.emitMarkersChanged(m.markers.All())

	bands := m.entries.Bands(m.activeBand, m.visibleBand)
	m.notifier.emitBandsChanged(bands)
	rows := m.rows()
	index := model.NewFrameIndex(rows)
	frame := model.BandmapFrame{
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

	nearestEntry, nearestEntryFound := m.nextVisibleEntryBy(model.BandmapByDistance(focusedFrequency), 0, model.Not(model.Or(model.OnFrequency(focusedFrequency), model.IsWorkedSpot)))
	if nearestEntryFound {
		frame.NearestEntry = nearestEntry
	}

	highestValueEntry, highestValueEntryFound := m.nextVisibleEntryBy(model.Descending(model.BandmapByValue), 0, model.Not(model.IsWorkedSpot))
	if highestValueEntryFound {
		frame.HighestValueEntry = highestValueEntry
	}

	m.asyncRunner(func() {
		m.view.ShowFrame(frame)
	})
}

func (m *Bandmap) rows() []model.BandmapRow {
	state := m.filter.State()
	markers := m.markers.Ordered(m.filter.MatchesMarker, state.SortBy, state.Descending)
	entries := m.entries.Query(m.filter.Order(), m.filter.Matches)

	result := make([]model.BandmapRow, 0, len(markers)+len(entries))
	for _, marker := range markers {
		result = append(result, model.MarkerBandmapRow(marker))
	}
	for _, entry := range entries {
		result = append(result, model.SpotBandmapRow(entry))
	}

	if !blendsMarkersWithSpots(state.SortBy) {
		// a marker has no callsign and no value, therefore it cannot take part in those
		// orders and it keeps a block at the top
		return result
	}

	slices.SortStableFunc(result, func(a, b model.BandmapRow) int {
		return compareRows(a, b, state.SortBy, state.Descending)
	})
	return result
}

func (m *Bandmap) markerOnFrequency(vfo model.VFOID) (model.BandmapMarker, bool) {
	if !validVFO(vfo) {
		return model.BandmapMarker{}, false
	}
	frequency := m.vfoFrequencies[vfo]
	if frequency == 0 {
		return model.BandmapMarker{}, false
	}

	var result model.BandmapMarker
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

func (m *Bandmap) nextEntryOnFrequency(vfo model.VFOID) (model.BandmapEntry, bool) {
	if !validVFO(vfo) {
		return model.BandmapEntry{}, false
	}
	frequency := m.vfoFrequencies[vfo]
	return m.nextEntryBy(model.BandmapByDistance(frequency), 2, m.entryVisibleOn(vfo), model.OnFrequency(frequency))
}

func (m *Bandmap) nextVisibleEntryBy(order model.BandmapOrder, limit int, filter model.BandmapFilter) (model.BandmapEntry, bool) {
	return m.nextEntryBy(order, limit, m.entryVisible, filter)
}

func (m *Bandmap) nextEntryBy(order model.BandmapOrder, limit int, visible model.BandmapFilter, filter model.BandmapFilter) (model.BandmapEntry, bool) {
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
	return model.BandmapEntry{}, false
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

func (m *Bandmap) SetVFO(vfo model.VFO) {
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

func (m *Bandmap) ContestChanged(contest model.Contest) {
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

func (m *Bandmap) ScoreChanged(_ model.Score) {
	m.do <- func() {
		m.update()
	}
}

func (m *Bandmap) VFOFrequencyChanged(vfo model.VFOID, frequency model.Frequency) {
	m.do <- func() {
		if validVFO(vfo) {
			m.vfoFrequencies[vfo] = frequency
		}
		if vfo == model.VFO1 {
			m.activeFrequency = frequency
		}
		// the frame is not updated with every frequency change, this creates too much load
	}
}

func (m *Bandmap) VFOBandChanged(vfo model.VFOID, band model.Band) {
	m.do <- func() {
		if validVFO(vfo) {
			m.vfoBands[vfo] = band
		}
		m.filter.VFOBandChanged(vfo, band)

		// the active band and the visible band belong to VFO1
		if vfo != model.VFO1 || band == m.activeBand {
			return
		}
		if m.activeBand == m.visibleBand {
			m.visibleBand = band
		}
		m.activeBand = band
		m.update()
	}
}

func (m *Bandmap) VFOModeChanged(vfo model.VFOID, mode model.Mode) {
	m.do <- func() {
		if validVFO(vfo) {
			m.vfoModes[vfo] = mode
		}
		m.filter.VFOModeChanged(vfo, mode)

		// the active mode belongs to VFO1
		if vfo != model.VFO1 || m.activeMode == mode {
			return
		}
		m.activeMode = mode
		m.update()
	}
}

func (m *Bandmap) FocusedVFOChanged(vfo model.VFOID) {
	m.do <- func() {
		m.filter.FocusedVFOChanged(vfo)
		// the operator works with another VFO now, therefore the navigation starts at the
		// frequency of that VFO again
		m.selection.clear()
		// the navigation and the marks in the frame follow the focused VFO
		m.update()
	}
}

func (m *Bandmap) VFOTuned(vfo model.VFOID, frequency model.Frequency) {
	m.do <- func() {
		if vfo != m.filter.FocusedVFO() {
			return
		}
		if !model.IsFrequencyJump(m.vfoFrequencies[vfo], frequency) {
			return
		}
		// the operator left the frequency of the selected row, therefore the navigation
		// starts at the new frequency again
		m.selection.clear()
		m.update()
	}
}

func (m *Bandmap) RadioChanged(name string, singleVFO bool) {
	m.do <- func() {
		m.filter.RadioChanged(name, singleVFO)
	}
}

func (m *Bandmap) SetFilterBand(band model.SpotFilterBand) {
	m.do <- func() {
		m.filter.SetBand(band)
	}
}

func (m *Bandmap) SetFilterMode(mode model.SpotFilterMode) {
	m.do <- func() {
		m.filter.SetMode(mode)
	}
}

func (m *Bandmap) SetFilterValue(value model.SpotFilterValue) {
	m.do <- func() {
		m.filter.SetValue(value)
	}
}

func (m *Bandmap) SetFilterSort(column model.SpotSortColumn, descending bool) {
	m.do <- func() {
		m.filter.SetSort(column, descending)
	}
}

func (m *Bandmap) SetFilterFolded(folded bool) {
	m.do <- func() {
		m.filter.SetFolded(folded)
	}
}

func (m *Bandmap) VisibleRows() []model.BandmapRow {
	return m.rows()
}

func (m *Bandmap) SelectableFilter() model.BandmapFilter {
	return m.filter.Matches
}

func (m *Bandmap) NavigationFilter() model.BandmapFilter {
	return m.entryVisible
}

func (m *Bandmap) MarkerNavigationFilter() model.BandmapMarkerFilter {
	return m.markerVisible
}

func (m *Bandmap) TargetVFO(entry model.BandmapEntry) model.VFOID {
	return m.filter.TargetVFO(entry)
}

func (m *Bandmap) TargetVFOForMarker(band model.Band) model.VFOID {
	return m.filter.TargetVFOForMarker(band)
}

func (m *Bandmap) FocusedVFO() model.VFOID {
	return m.filter.FocusedVFO()
}

func (m *Bandmap) entryVisible(entry model.BandmapEntry) bool {
	// the keyboard navigation tunes the focused VFO, therefore it stays on the band and
	// in the mode of that VFO
	return m.entryVisibleOn(m.filter.FocusedVFO())(entry)
}

func (m *Bandmap) markerVisible(marker model.BandmapMarker) bool {
	if marker.Kind == model.CQMarker {
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
	return band == model.NoBand || marker.Band == band
}

func (m *Bandmap) entryVisibleOn(vfo model.VFOID) model.BandmapFilter {
	return func(entry model.BandmapEntry) bool {
		if !validVFO(vfo) {
			return false
		}
		// a band or a mode that the rig did not report yet restricts nothing, the spot
		// filter uses the same rule
		band := m.vfoBands[vfo]
		mode := m.vfoModes[vfo]
		return (band == model.NoBand || entry.Band == band) &&
			(mode == model.NoMode || entry.Mode == mode)
	}
}

func (m *Bandmap) focusedBand() model.Band {
	return m.vfoBands[m.filter.FocusedVFO()]
}

func (m *Bandmap) focusedMode() model.Mode {
	return m.vfoModes[m.filter.FocusedVFO()]
}

func (m *Bandmap) focusedFrequency() model.Frequency {
	return m.vfoFrequencies[m.filter.FocusedVFO()]
}

func (m *Bandmap) countEntryValue(entry model.BandmapEntry) bool {
	if entry.Source == model.WorkedSpot {
		return false
	}
	// a mode that the rig did not report yet restricts nothing
	return m.activeMode == model.NoMode || entry.Mode == m.activeMode
}

func (m *Bandmap) Notify(listener any) {
	m.do <- func() {
		m.notifier.Notify(listener)
	}
}

func (m *Bandmap) Add(spot model.Spot) {
	m.do <- func() {
		mode := spot.Mode
		if mode == model.NoMode {
			mode = m.activeMode
		}

		if !spot.IsWorked() {
			_, worked := m.dupeChecker.FindWorkedQSOs(spot.Call, spot.Band, mode)
			if worked {
				spot.Source = model.WorkedSpot
			}
		}

		m.entries.Add(spot, m.clock.Now(), m.weights)

		if spot.IsWorked() {
			band := spot.Band
			switch m.bandRule {
			case conval.Once:
				band, mode = model.NoBand, model.NoMode
			case conval.OncePerBand:
				mode = model.NoMode
			case conval.OncePerBandAndMode:
				// band, mode = band, mode
			default:
				band, mode = model.NoBand, model.NoMode
			}
			m.entries.MarkAsWorked(spot.Call, band, mode)
		}
	}
}

func (m *Bandmap) MacroSent(vfo model.VFOID, workmode model.Workmode, index int) {
	m.do <- func() {
		// only VFO1 works in the running workmode, and only the CQ macro marks a
		// running frequency
		if vfo != model.VFO1 || workmode != model.Run || index != model.CQMacroIndex {
			return
		}
		m.setCQMarker()
	}
}

func (m *Bandmap) setCQMarker() {
	frequency := m.vfoFrequencies[model.VFO1]
	if frequency == 0 {
		// the rig did not report a frequency yet
		return
	}
	m.markers.SetCQ(frequency, m.vfoBands[model.VFO1], m.clock.Now())
	m.update()
}

func (m *Bandmap) SelectMarker(id model.BandmapEntryID) {
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

func (m *Bandmap) selectMarker(marker model.BandmapMarker) {
	m.selection.SelectMarker(marker)
	m.update()
}

func (m *Bandmap) MarkWithNextNumber(frequency model.Frequency, band model.Band, input model.MarkerInput) {
	m.do <- func() {
		marker := m.markers.MarkWithNextNumber(frequency, band, m.clock.Now())
		m.markers.SetInput(marker.ID, input)
		m.update()
	}
}

func (m *Bandmap) MarkWithNumber(number int, frequency model.Frequency, band model.Band, input model.MarkerInput) {
	m.do <- func() {
		marker := m.markers.MarkWithNumber(number, frequency, band, m.clock.Now())
		m.markers.SetInput(marker.ID, input)
		m.update()
	}
}

func (m *Bandmap) MarkWithText(text string, frequency model.Frequency, band model.Band) {
	m.do <- func() {
		m.markers.MarkWithText(text, frequency, band, m.clock.Now())
		m.update()
	}
}

func (m *Bandmap) AddNumberedMarker(number int, frequency model.Frequency, band model.Band) {
	m.do <- func() {
		m.markers.AddNumbered(number, frequency, band, m.clock.Now())
		m.update()
	}
}

func (m *Bandmap) AddTextMarker(text string, frequency model.Frequency, band model.Band) {
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

func (m *Bandmap) RemoveMarker(id model.BandmapEntryID) {
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

func (m *Bandmap) QTCAdded(qtc model.QTC) {
	m.do <- func() {
		m.entries.RefreshCallinfo(qtc.TheirCallsign, m.clock.Now(), m.weights)
	}
}

func (m *Bandmap) Remove(spot model.Spot) {
	m.do <- func() {
		if !m.entries.Remove(spot) {
			return
		}
		m.update()
	}
}

func (m *Bandmap) IgnoreEntry(id model.BandmapEntryID) {
	m.do <- func() {
		m.setEntryIgnored(id, true)
	}
}

func (m *Bandmap) RestoreEntry(id model.BandmapEntryID) {
	m.do <- func() {
		m.setEntryIgnored(id, false)
	}
}

func (m *Bandmap) IgnoreEntryOnFrequency() {
	m.do <- func() {
		m.setEntryOnFrequencyIgnored(true)
	}
}

func (m *Bandmap) RestoreEntryOnFrequency() {
	m.do <- func() {
		m.setEntryOnFrequencyIgnored(false)
	}
}

func (m *Bandmap) setEntryOnFrequencyIgnored(ignored bool) {
	// the operator names the spot with the frequency of the focused VFO, with the same
	// rule as for the deletion of a marker
	entry, found := m.nextEntryOnFrequency(m.filter.FocusedVFO())
	if !found {
		return
	}
	m.setEntryIgnored(entry.ID, ignored)
}

func (m *Bandmap) setEntryIgnored(id model.BandmapEntryID, ignored bool) {
	if !m.entries.SetIgnored(id, ignored) {
		return
	}
	// the value of the entry follows the ignore flag, the update recalculates it
	m.update()
}

func (m *Bandmap) SelectEntry(id model.BandmapEntryID) {
	m.do <- func() {
		m.selection.SelectEntry(id)
		m.update()
	}
}

func (m *Bandmap) SelectByCallsign(call model.Callsign) {
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

func (m *Bandmap) GotoNearestEntry(frequency model.Frequency) {
	m.do <- func() {
		m.selection.SelectNearest(frequency)
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

func (l *Logger) EntryAdded(e model.BandmapEntry) {
	log.Printf("Bandmap entry added: %v", e)
}

func (l *Logger) EntryUpdated(e model.BandmapEntry) {
	log.Printf("Bandmap entry updated: %v", e)
}

func (l *Logger) EntryRemoved(e model.BandmapEntry) {
	log.Printf("Bandmap entry removed: %v", e)
}

type nullView struct{}

func (v *nullView) Show()                          {}
func (v *nullView) Hide()                          {}
func (v *nullView) ShowFrame(model.BandmapFrame)   {}
func (v *nullView) RevealEntry(model.BandmapEntry) {}

type nullVFO struct{}

func (n *nullVFO) Name() string                                                            { return "" }
func (n *nullVFO) Notify(any)                                                              {}
func (n *nullVFO) Active() bool                                                            { return false }
func (n *nullVFO) Refresh()                                                                {}
func (n *nullVFO) SetFrequency(model.Frequency)                                            {}
func (n *nullVFO) ShiftFrequency(model.Frequency)                                          {}
func (n *nullVFO) SetBand(model.Band)                                                      {}
func (n *nullVFO) SetMode(model.Mode)                                                      {}
func (n *nullVFO) IncrementalTuningActive(model.IncrementalTuningKind) bool                { return false }
func (n *nullVFO) SetIncrementalTuningActive(model.IncrementalTuningKind, bool)            {}
func (n *nullVFO) SetIncrementalTuning(model.IncrementalTuningKind, bool, model.Frequency) {}
func (n *nullVFO) ShiftIncrementalTuning(model.IncrementalTuningKind, model.Frequency)     {}
func (n *nullVFO) ToggleAvailableIncrementalTuning()                                       {}
func (n *nullVFO) ShiftAvailableIncrementalTuning(model.Frequency)                         {}

type nullCallinfo struct{}

func (n *nullCallinfo) GetInfo(model.Callsign, model.Band, model.Mode, []string) model.Callinfo {
	return model.Callinfo{}
}
func (n *nullCallinfo) UpdateValue(*model.Callinfo, model.Band, model.Mode) bool {
	return false
}
