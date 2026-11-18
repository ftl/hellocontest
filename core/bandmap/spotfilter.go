package bandmap

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/ftl/hellocontest/core/model"
)

const MainSpotFilterID = "main"

type SpotFilterStore interface {
	SpotFilter(id string) (model.SpotFilterState, bool)
	SetSpotFilter(id string, state model.SpotFilterState) error
}

type SpotFilter struct {
	id    string
	store SpotFilterStore
	state model.SpotFilterState

	vfoBands      [model.VFOCount]model.Band
	vfoModes      [model.VFOCount]model.Mode
	focusedVFO    model.VFOID
	vfo2Available bool
	contestBands  []model.Band
	contestModes  []model.Mode

	changed func()
}

func NewSpotFilter(id string, store SpotFilterStore) *SpotFilter {
	result := &SpotFilter{
		id:    id,
		store: store,
		state: defaultSpotFilterState(),
	}
	if store != nil {
		if stored, ok := store.SpotFilter(id); ok {
			result.state = stored
		}
	}
	return result
}

func defaultSpotFilterState() model.SpotFilterState {
	return model.SpotFilterState{
		Band:   model.SpotFilterBand{Kind: model.SpotFilterFocused},
		Mode:   model.SpotFilterMode{Kind: model.SpotFilterFocused},
		SortBy: model.SortSpotsByFrequency,
	}
}

func (f *SpotFilter) OnChanged(changed func()) {
	f.changed = changed
}

func (f *SpotFilter) State() model.SpotFilterState {
	return f.state
}

func (f *SpotFilter) SetBand(band model.SpotFilterBand) {
	if f.state.Band == band {
		return
	}
	f.state.Band = band
	f.storeState()
	f.emitChanged()
}

func (f *SpotFilter) SetMode(mode model.SpotFilterMode) {
	if f.state.Mode == mode {
		return
	}
	f.state.Mode = mode
	f.storeState()
	f.emitChanged()
}

func (f *SpotFilter) SetValue(value model.SpotFilterValue) {
	if f.state.Value == value {
		return
	}
	f.state.Value = value
	f.storeState()
	f.emitChanged()
}

func (f *SpotFilter) SetSort(column model.SpotSortColumn, descending bool) {
	if f.state.SortBy == column && f.state.Descending == descending {
		return
	}
	f.state.SortBy = column
	f.state.Descending = descending
	f.storeState()
	f.emitChanged()
}

func (f *SpotFilter) SetFolded(folded bool) {
	if f.state.Folded == folded {
		return
	}
	// folding changes no resolved value and no order, therefore it notifies nobody
	f.state.Folded = folded
	f.storeState()
}

func (f *SpotFilter) VFOBandChanged(vfo model.VFOID, band model.Band) {
	if !validVFO(vfo) || f.vfoBands[vfo] == band {
		return
	}
	f.emitOnResolvedChange(func() { f.vfoBands[vfo] = band })
}

func (f *SpotFilter) VFOModeChanged(vfo model.VFOID, mode model.Mode) {
	if !validVFO(vfo) || f.vfoModes[vfo] == mode {
		return
	}
	f.emitOnResolvedChange(func() { f.vfoModes[vfo] = mode })
}

func (f *SpotFilter) FocusedVFOChanged(vfo model.VFOID) {
	if !validVFO(vfo) || f.focusedVFO == vfo {
		return
	}
	f.emitOnResolvedChange(func() { f.focusedVFO = vfo })
}

func (f *SpotFilter) RadioChanged(_ string, singleVFO bool) {
	vfo2Available := !singleVFO
	if f.vfo2Available == vfo2Available {
		return
	}
	f.emitOnResolvedChange(func() { f.vfo2Available = vfo2Available })
}

func (f *SpotFilter) SetContestBandsAndModes(bands []model.Band, modes []model.Mode) {
	if slices.Equal(f.contestBands, bands) && slices.Equal(f.contestModes, modes) {
		return
	}
	f.emitOnResolvedChange(func() {
		f.contestBands = bands
		f.contestModes = modes
	})
}

func (f *SpotFilter) Matches(entry model.BandmapEntry) bool {
	return f.bandMatches(entry.Band) && f.modeMatches(entry.Mode) && f.state.Value.Matches(entry)
}

func (f *SpotFilter) MatchesMarker(marker model.BandmapMarker) bool {
	// a marker has no mode, therefore only the band decides
	return f.bandMatches(marker.Band)
}

func (f *SpotFilter) Order() model.BandmapOrder {
	var order model.BandmapOrder
	switch f.state.SortBy {
	case model.SortSpotsByCallsign:
		order = model.BandmapByCallsign
	case model.SortSpotsByValue:
		order = model.BandmapByValue
	case model.SortSpotsByLastSeen:
		order = model.BandmapByLastSeen
	default:
		order = model.BandmapByFrequency
	}
	if f.state.Descending {
		return model.Descending(order)
	}
	return order
}

func (f *SpotFilter) TargetVFO(entry model.BandmapEntry) model.VFOID {
	return f.TargetVFOForBand(entry.Band)
}

func (f *SpotFilter) TargetVFOForMarker(band model.Band) model.VFOID {
	// a marker names a frequency to go to: if a VFO is already on that band, that VFO goes
	// there and no VFO changes its band
	if f.vfoOnBand(f.FocusedVFO(), band) {
		return f.FocusedVFO()
	}
	for vfo := range model.VFOCount {
		if f.vfoOnBand(model.VFOID(vfo), band) {
			return model.VFOID(vfo)
		}
	}
	return f.TargetVFOForBand(band)
}

func (f *SpotFilter) TargetVFOForBand(band model.Band) model.VFOID {
	// a band selection that binds the table to one VFO names that VFO
	switch f.state.Band.Kind {
	case model.SpotFilterVFO1:
		return model.VFO1
	case model.SpotFilterVFO2:
		return f.availableVFO(model.VFO2)
	case model.SpotFilterFocused:
		return f.FocusedVFO()
	}

	// a free band selection prefers the VFO that already is on that band
	if f.vfoOnBand(f.FocusedVFO(), band) {
		return f.FocusedVFO()
	}
	for vfo := range model.VFOCount {
		if f.vfoOnBand(model.VFOID(vfo), band) {
			return model.VFOID(vfo)
		}
	}
	return f.FocusedVFO()
}

func (f *SpotFilter) FocusedVFO() model.VFOID {
	return f.availableVFO(f.focusedVFO)
}

func (f *SpotFilter) Frame() model.SpotFilterFrame {
	return model.SpotFilterFrame{
		SpotFilterState: f.state,
		Bands:           model.Bands,
		Modes:           model.Modes,
		Description:     f.description(),
	}
}

func (f *SpotFilter) bandMatches(band model.Band) bool {
	switch f.state.Band.Kind {
	case model.SpotFilterFixed:
		return band == f.state.Band.Band
	case model.SpotFilterContest:
		return len(f.contestBands) == 0 || slices.Contains(f.contestBands, band)
	case model.SpotFilterVFO1, model.SpotFilterVFO2, model.SpotFilterFocused:
		selected := f.selectedBand()
		return selected == model.NoBand || band == selected
	default:
		return true
	}
}

func (f *SpotFilter) modeMatches(mode model.Mode) bool {
	switch f.state.Mode.Kind {
	case model.SpotFilterFixed:
		return mode == f.state.Mode.Mode
	case model.SpotFilterContest:
		return len(f.contestModes) == 0 || slices.Contains(f.contestModes, mode)
	case model.SpotFilterVFO1, model.SpotFilterVFO2, model.SpotFilterFocused:
		selected := f.selectedMode()
		return selected == model.NoMode || mode == selected
	default:
		return true
	}
}

func (f *SpotFilter) selectedBand() model.Band {
	vfo, ok := f.boundVFO(f.state.Band.Kind)
	if !ok {
		return model.NoBand
	}
	return f.vfoBands[vfo]
}

func (f *SpotFilter) selectedMode() model.Mode {
	vfo, ok := f.boundVFO(f.state.Mode.Kind)
	if !ok {
		return model.NoMode
	}
	return f.vfoModes[vfo]
}

func (f *SpotFilter) boundVFO(kind model.SpotFilterKind) (model.VFOID, bool) {
	switch kind {
	case model.SpotFilterVFO1:
		return model.VFO1, true
	case model.SpotFilterVFO2:
		return f.availableVFO(model.VFO2), true
	case model.SpotFilterFocused:
		return f.FocusedVFO(), true
	default:
		return model.VFO1, false
	}
}

func (f *SpotFilter) vfoOnBand(vfo model.VFOID, band model.Band) bool {
	if !validVFO(vfo) || band == model.NoBand {
		return false
	}
	if vfo == model.VFO2 && !f.vfo2Available {
		return false
	}
	return f.vfoBands[vfo] == band
}

func (f *SpotFilter) availableVFO(vfo model.VFOID) model.VFOID {
	if vfo == model.VFO2 && !f.vfo2Available {
		return model.VFO1
	}
	if !validVFO(vfo) {
		return model.VFO1
	}
	return vfo
}

func (f *SpotFilter) description() string {
	direction := "↑"
	if f.state.Descending {
		direction = "↓"
	}
	parts := []string{
		f.bandDescription(),
		f.modeDescription(),
	}
	if f.state.Value != model.SpotFilterAnyValue {
		parts = append(parts, f.state.Value.Label())
	}
	parts = append(parts, fmt.Sprintf("by %s %s", f.state.SortBy.Label(), direction))
	return strings.Join(parts, " · ")
}

func (f *SpotFilter) bandDescription() string {
	switch f.state.Band.Kind {
	case model.SpotFilterFixed:
		return string(f.state.Band.Band)
	case model.SpotFilterContest:
		return "Contest bands"
	case model.SpotFilterVFO1, model.SpotFilterVFO2, model.SpotFilterFocused:
		return boundDescription(f.state.Band.Kind, string(f.selectedBand()))
	default:
		return "All bands"
	}
}

func (f *SpotFilter) modeDescription() string {
	switch f.state.Mode.Kind {
	case model.SpotFilterFixed:
		return string(f.state.Mode.Mode)
	case model.SpotFilterContest:
		return "Contest modes"
	case model.SpotFilterVFO1, model.SpotFilterVFO2, model.SpotFilterFocused:
		return boundDescription(f.state.Mode.Kind, string(f.selectedMode()))
	default:
		return "All modes"
	}
}

func boundDescription(kind model.SpotFilterKind, value string) string {
	name := "Focused VFO"
	switch kind {
	case model.SpotFilterVFO1:
		name = "VFO1"
	case model.SpotFilterVFO2:
		name = "VFO2"
	}
	if value == "" {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, value)
}

func (f *SpotFilter) emitOnResolvedChange(update func()) {
	oldBand, oldMode := f.selectedBand(), f.selectedMode()
	update()
	if oldBand == f.selectedBand() && oldMode == f.selectedMode() {
		return
	}
	f.emitChanged()
}

func (f *SpotFilter) emitChanged() {
	if f.changed == nil {
		return
	}
	f.changed()
}

func (f *SpotFilter) storeState() {
	if f.store == nil {
		return
	}
	if err := f.store.SetSpotFilter(f.id, f.state); err != nil {
		log.Printf("cannot store the spot filter %s: %v", f.id, err)
	}
}

func validVFO(vfo model.VFOID) bool {
	return vfo >= 0 && vfo < model.VFOCount
}
