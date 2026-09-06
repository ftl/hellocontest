package bandmap

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/ftl/hellocontest/core"
)

const MainSpotFilterID = "main"

type SpotFilterStore interface {
	SpotFilter(id string) (core.SpotFilterState, bool)
	SetSpotFilter(id string, state core.SpotFilterState) error
}

type SpotFilter struct {
	id    string
	store SpotFilterStore
	state core.SpotFilterState

	vfoBands      [core.VFOCount]core.Band
	vfoModes      [core.VFOCount]core.Mode
	focusedVFO    core.VFOID
	vfo2Available bool
	contestBands  []core.Band
	contestModes  []core.Mode

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

func defaultSpotFilterState() core.SpotFilterState {
	return core.SpotFilterState{
		Band:   core.SpotFilterBand{Kind: core.SpotFilterFocused},
		Mode:   core.SpotFilterMode{Kind: core.SpotFilterFocused},
		SortBy: core.SortSpotsByFrequency,
	}
}

func (f *SpotFilter) OnChanged(changed func()) {
	f.changed = changed
}

func (f *SpotFilter) State() core.SpotFilterState {
	return f.state
}

func (f *SpotFilter) SetBand(band core.SpotFilterBand) {
	if f.state.Band == band {
		return
	}
	f.state.Band = band
	f.storeState()
	f.emitChanged()
}

func (f *SpotFilter) SetMode(mode core.SpotFilterMode) {
	if f.state.Mode == mode {
		return
	}
	f.state.Mode = mode
	f.storeState()
	f.emitChanged()
}

func (f *SpotFilter) SetSort(column core.SpotSortColumn, descending bool) {
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

func (f *SpotFilter) VFOBandChanged(vfo core.VFOID, band core.Band) {
	if !validVFO(vfo) || f.vfoBands[vfo] == band {
		return
	}
	f.emitOnResolvedChange(func() { f.vfoBands[vfo] = band })
}

func (f *SpotFilter) VFOModeChanged(vfo core.VFOID, mode core.Mode) {
	if !validVFO(vfo) || f.vfoModes[vfo] == mode {
		return
	}
	f.emitOnResolvedChange(func() { f.vfoModes[vfo] = mode })
}

func (f *SpotFilter) FocusedVFOChanged(vfo core.VFOID) {
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

func (f *SpotFilter) SetContestBandsAndModes(bands []core.Band, modes []core.Mode) {
	if slices.Equal(f.contestBands, bands) && slices.Equal(f.contestModes, modes) {
		return
	}
	f.emitOnResolvedChange(func() {
		f.contestBands = bands
		f.contestModes = modes
	})
}

func (f *SpotFilter) Matches(entry core.BandmapEntry) bool {
	return f.bandMatches(entry.Band) && f.modeMatches(entry.Mode)
}

func (f *SpotFilter) Order() core.BandmapOrder {
	var order core.BandmapOrder
	switch f.state.SortBy {
	case core.SortSpotsByCallsign:
		order = core.BandmapByCallsign
	case core.SortSpotsByValue:
		order = core.BandmapByValue
	case core.SortSpotsByLastSeen:
		order = core.BandmapByLastSeen
	default:
		order = core.BandmapByFrequency
	}
	if f.state.Descending {
		return core.Descending(order)
	}
	return order
}

func (f *SpotFilter) TargetVFO(entry core.BandmapEntry) core.VFOID {
	// a band selection that binds the table to one VFO names that VFO
	switch f.state.Band.Kind {
	case core.SpotFilterVFO1:
		return core.VFO1
	case core.SpotFilterVFO2:
		return f.availableVFO(core.VFO2)
	case core.SpotFilterFocused:
		return f.FocusedVFO()
	}

	// a free band selection prefers the VFO that already is on the band of the spot
	if f.vfoOnBand(f.FocusedVFO(), entry.Band) {
		return f.FocusedVFO()
	}
	for vfo := range core.VFOCount {
		if f.vfoOnBand(core.VFOID(vfo), entry.Band) {
			return core.VFOID(vfo)
		}
	}
	return f.FocusedVFO()
}

func (f *SpotFilter) FocusedVFO() core.VFOID {
	return f.availableVFO(f.focusedVFO)
}

func (f *SpotFilter) Frame() core.SpotFilterFrame {
	return core.SpotFilterFrame{
		SpotFilterState: f.state,
		Bands:           core.Bands,
		Modes:           core.Modes,
		Description:     f.description(),
	}
}

func (f *SpotFilter) bandMatches(band core.Band) bool {
	switch f.state.Band.Kind {
	case core.SpotFilterFixed:
		return band == f.state.Band.Band
	case core.SpotFilterContest:
		return len(f.contestBands) == 0 || slices.Contains(f.contestBands, band)
	case core.SpotFilterVFO1, core.SpotFilterVFO2, core.SpotFilterFocused:
		selected := f.selectedBand()
		return selected == core.NoBand || band == selected
	default:
		return true
	}
}

func (f *SpotFilter) modeMatches(mode core.Mode) bool {
	switch f.state.Mode.Kind {
	case core.SpotFilterFixed:
		return mode == f.state.Mode.Mode
	case core.SpotFilterContest:
		return len(f.contestModes) == 0 || slices.Contains(f.contestModes, mode)
	case core.SpotFilterVFO1, core.SpotFilterVFO2, core.SpotFilterFocused:
		selected := f.selectedMode()
		return selected == core.NoMode || mode == selected
	default:
		return true
	}
}

func (f *SpotFilter) selectedBand() core.Band {
	vfo, ok := f.boundVFO(f.state.Band.Kind)
	if !ok {
		return core.NoBand
	}
	return f.vfoBands[vfo]
}

func (f *SpotFilter) selectedMode() core.Mode {
	vfo, ok := f.boundVFO(f.state.Mode.Kind)
	if !ok {
		return core.NoMode
	}
	return f.vfoModes[vfo]
}

func (f *SpotFilter) boundVFO(kind core.SpotFilterKind) (core.VFOID, bool) {
	switch kind {
	case core.SpotFilterVFO1:
		return core.VFO1, true
	case core.SpotFilterVFO2:
		return f.availableVFO(core.VFO2), true
	case core.SpotFilterFocused:
		return f.FocusedVFO(), true
	default:
		return core.VFO1, false
	}
}

func (f *SpotFilter) vfoOnBand(vfo core.VFOID, band core.Band) bool {
	if !validVFO(vfo) || band == core.NoBand {
		return false
	}
	if vfo == core.VFO2 && !f.vfo2Available {
		return false
	}
	return f.vfoBands[vfo] == band
}

func (f *SpotFilter) availableVFO(vfo core.VFOID) core.VFOID {
	if vfo == core.VFO2 && !f.vfo2Available {
		return core.VFO1
	}
	if !validVFO(vfo) {
		return core.VFO1
	}
	return vfo
}

func (f *SpotFilter) description() string {
	direction := "↑"
	if f.state.Descending {
		direction = "↓"
	}
	return strings.Join([]string{
		f.bandDescription(),
		f.modeDescription(),
		fmt.Sprintf("by %s %s", f.state.SortBy.Label(), direction),
	}, " · ")
}

func (f *SpotFilter) bandDescription() string {
	switch f.state.Band.Kind {
	case core.SpotFilterFixed:
		return string(f.state.Band.Band)
	case core.SpotFilterContest:
		return "Contest bands"
	case core.SpotFilterVFO1, core.SpotFilterVFO2, core.SpotFilterFocused:
		return boundDescription(f.state.Band.Kind, string(f.selectedBand()))
	default:
		return "All bands"
	}
}

func (f *SpotFilter) modeDescription() string {
	switch f.state.Mode.Kind {
	case core.SpotFilterFixed:
		return string(f.state.Mode.Mode)
	case core.SpotFilterContest:
		return "Contest modes"
	case core.SpotFilterVFO1, core.SpotFilterVFO2, core.SpotFilterFocused:
		return boundDescription(f.state.Mode.Kind, string(f.selectedMode()))
	default:
		return "All modes"
	}
}

func boundDescription(kind core.SpotFilterKind, value string) string {
	name := "Focused VFO"
	switch kind {
	case core.SpotFilterVFO1:
		name = "VFO1"
	case core.SpotFilterVFO2:
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

func validVFO(vfo core.VFOID) bool {
	return vfo >= 0 && vfo < core.VFOCount
}
