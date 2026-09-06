package bandmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestSpotFilter_DefaultState(t *testing.T) {
	filter, _ := setupSpotFilter()

	state := filter.State()

	assert.Equal(t, core.SpotFilterFocused, state.Band.Kind, "band follows the focused VFO")
	assert.Equal(t, core.SpotFilterFocused, state.Mode.Kind, "mode follows the focused VFO")
	assert.Equal(t, core.SortSpotsByFrequency, state.SortBy)
	assert.False(t, state.Descending)
	assert.False(t, state.Folded)
}

func TestSpotFilter_RestoresTheStoredState(t *testing.T) {
	store := newSpotFilterStoreStub()
	stored := core.SpotFilterState{
		Band:       core.FixedSpotFilterBand(core.Band20m),
		Mode:       core.FixedSpotFilterMode(core.ModeCW),
		SortBy:     core.SortSpotsByValue,
		Descending: true,
		Folded:     true,
	}
	store.states[MainSpotFilterID] = stored

	filter := NewSpotFilter(MainSpotFilterID, store)

	assert.Equal(t, stored, filter.State())
}

func TestSpotFilter_StoresEveryChange(t *testing.T) {
	filter, store := setupSpotFilter()

	filter.SetBand(core.FixedSpotFilterBand(core.Band40m))
	filter.SetMode(core.FixedSpotFilterMode(core.ModeSSB))
	filter.SetSort(core.SortSpotsByLastSeen, true)
	filter.SetFolded(true)

	stored, ok := store.states[MainSpotFilterID]
	require.True(t, ok)
	assert.Equal(t, core.FixedSpotFilterBand(core.Band40m), stored.Band)
	assert.Equal(t, core.FixedSpotFilterMode(core.ModeSSB), stored.Mode)
	assert.Equal(t, core.SortSpotsByLastSeen, stored.SortBy)
	assert.True(t, stored.Descending)
	assert.True(t, stored.Folded)
}

func TestSpotFilter_NotifiesAboutRelevantChanges(t *testing.T) {
	filter, _ := setupSpotFilter()
	changes := 0
	filter.OnChanged(func() { changes++ })

	filter.SetBand(core.FixedSpotFilterBand(core.Band40m))
	assert.Equal(t, 1, changes, "a new band changes the content")

	filter.SetMode(core.FixedSpotFilterMode(core.ModeSSB))
	assert.Equal(t, 2, changes, "a new mode changes the content")

	filter.SetSort(core.SortSpotsByValue, true)
	assert.Equal(t, 3, changes, "a new order changes the content")

	filter.SetFolded(true)
	assert.Equal(t, 3, changes, "folding does not change the content")
}

func TestSpotFilter_NotifiesOnlyOnResolvedChanges(t *testing.T) {
	filter, _ := setupSpotFilter()
	filter.SetBand(core.SpotFilterBand{Kind: core.SpotFilterVFO1})
	changes := 0
	filter.OnChanged(func() { changes++ })

	filter.VFOBandChanged(core.VFO1, core.Band40m)
	assert.Equal(t, 1, changes, "the band of VFO1 is the resolved band")

	filter.VFOBandChanged(core.VFO1, core.Band40m)
	assert.Equal(t, 1, changes, "the same band again resolves to the same band")

	filter.VFOBandChanged(core.VFO2, core.Band20m)
	assert.Equal(t, 1, changes, "the band of VFO2 does not matter here")

	filter.FocusedVFOChanged(core.VFO2)
	assert.Equal(t, 1, changes, "the focus does not matter here")
}

func TestSpotFilter_Matches(t *testing.T) {
	tt := []struct {
		desc          string
		band          core.SpotFilterBand
		mode          core.SpotFilterMode
		vfo2Available bool
		entry         core.BandmapEntry
		want          bool
	}{
		{
			desc:  "all bands and all modes",
			band:  core.SpotFilterBand{Kind: core.SpotFilterAll},
			mode:  core.SpotFilterMode{Kind: core.SpotFilterAll},
			entry: core.BandmapEntry{Band: core.Band10m, Mode: core.ModeFM},
			want:  true,
		},
		{
			desc:  "fixed band matches",
			band:  core.FixedSpotFilterBand(core.Band40m),
			mode:  core.SpotFilterMode{Kind: core.SpotFilterAll},
			entry: core.BandmapEntry{Band: core.Band40m, Mode: core.ModeCW},
			want:  true,
		},
		{
			desc:  "fixed band does not match",
			band:  core.FixedSpotFilterBand(core.Band40m),
			mode:  core.SpotFilterMode{Kind: core.SpotFilterAll},
			entry: core.BandmapEntry{Band: core.Band20m, Mode: core.ModeCW},
			want:  false,
		},
		{
			desc:  "fixed mode does not match",
			band:  core.SpotFilterBand{Kind: core.SpotFilterAll},
			mode:  core.FixedSpotFilterMode(core.ModeCW),
			entry: core.BandmapEntry{Band: core.Band40m, Mode: core.ModeSSB},
			want:  false,
		},
		{
			desc:  "band of VFO1",
			band:  core.SpotFilterBand{Kind: core.SpotFilterVFO1},
			mode:  core.SpotFilterMode{Kind: core.SpotFilterAll},
			entry: core.BandmapEntry{Band: core.Band40m, Mode: core.ModeCW},
			want:  true,
		},
		{
			desc:          "band of VFO2",
			band:          core.SpotFilterBand{Kind: core.SpotFilterVFO2},
			mode:          core.SpotFilterMode{Kind: core.SpotFilterAll},
			vfo2Available: true,
			entry:         core.BandmapEntry{Band: core.Band20m, Mode: core.ModeSSB},
			want:          true,
		},
		{
			desc:  "band of VFO2 falls back to VFO1 without a second VFO",
			band:  core.SpotFilterBand{Kind: core.SpotFilterVFO2},
			mode:  core.SpotFilterMode{Kind: core.SpotFilterAll},
			entry: core.BandmapEntry{Band: core.Band40m, Mode: core.ModeCW},
			want:  true,
		},
		{
			desc:          "band of the focused VFO",
			band:          core.SpotFilterBand{Kind: core.SpotFilterFocused},
			mode:          core.SpotFilterMode{Kind: core.SpotFilterFocused},
			vfo2Available: true,
			entry:         core.BandmapEntry{Band: core.Band40m, Mode: core.ModeCW},
			want:          true,
		},
		{
			desc:  "contest band",
			band:  core.SpotFilterBand{Kind: core.SpotFilterContest},
			mode:  core.SpotFilterMode{Kind: core.SpotFilterAll},
			entry: core.BandmapEntry{Band: core.Band20m, Mode: core.ModeCW},
			want:  true,
		},
		{
			desc:  "band outside the contest",
			band:  core.SpotFilterBand{Kind: core.SpotFilterContest},
			mode:  core.SpotFilterMode{Kind: core.SpotFilterAll},
			entry: core.BandmapEntry{Band: core.Band10m, Mode: core.ModeCW},
			want:  false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			filter, _ := setupSpotFilter()
			filter.SetContestBandsAndModes([]core.Band{core.Band40m, core.Band20m}, []core.Mode{core.ModeCW})
			filter.RadioChanged("radio", !tc.vfo2Available)
			filter.VFOBandChanged(core.VFO1, core.Band40m)
			filter.VFOModeChanged(core.VFO1, core.ModeCW)
			filter.VFOBandChanged(core.VFO2, core.Band20m)
			filter.VFOModeChanged(core.VFO2, core.ModeSSB)
			filter.FocusedVFOChanged(core.VFO1)
			filter.SetBand(tc.band)
			filter.SetMode(tc.mode)

			assert.Equal(t, tc.want, filter.Matches(tc.entry))
		})
	}
}

func TestSpotFilter_MatchesEverythingWhileUnresolved(t *testing.T) {
	filter, _ := setupSpotFilter()

	assert.True(t, filter.Matches(core.BandmapEntry{Band: core.Band10m, Mode: core.ModeFM}), "no radio, no restriction")
}

func TestSpotFilter_Order(t *testing.T) {
	now := time.Now()
	low := core.BandmapEntry{ID: 1, Call: core.MustParseCallsign("dl1abc"), Frequency: 3500000, LastHeard: now.Add(-time.Minute), Info: core.Callinfo{WeightedValue: 1}}
	high := core.BandmapEntry{ID: 2, Call: core.MustParseCallsign("dl2abc"), Frequency: 7000000, LastHeard: now, Info: core.Callinfo{WeightedValue: 9}}

	for _, column := range core.SpotSortColumns {
		t.Run(column.String(), func(t *testing.T) {
			filter, _ := setupSpotFilter()

			filter.SetSort(column, false)
			assert.Negative(t, filter.Order()(low, high), "ascending")

			filter.SetSort(column, true)
			assert.Positive(t, filter.Order()(low, high), "descending")
		})
	}
}

func TestSpotFilter_TargetVFO(t *testing.T) {
	tt := []struct {
		desc          string
		band          core.SpotFilterBand
		vfo2Available bool
		vfo1Band      core.Band
		vfo2Band      core.Band
		focusedVFO    core.VFOID
		entryBand     core.Band
		want          core.VFOID
	}{
		{
			desc:          "band bound to VFO1",
			band:          core.SpotFilterBand{Kind: core.SpotFilterVFO1},
			vfo2Available: true,
			vfo1Band:      core.Band40m,
			vfo2Band:      core.Band20m,
			focusedVFO:    core.VFO2,
			entryBand:     core.Band40m,
			want:          core.VFO1,
		},
		{
			desc:          "band bound to VFO2",
			band:          core.SpotFilterBand{Kind: core.SpotFilterVFO2},
			vfo2Available: true,
			vfo1Band:      core.Band40m,
			vfo2Band:      core.Band20m,
			focusedVFO:    core.VFO1,
			entryBand:     core.Band20m,
			want:          core.VFO2,
		},
		{
			desc:       "band bound to VFO2 without a second VFO",
			band:       core.SpotFilterBand{Kind: core.SpotFilterVFO2},
			vfo1Band:   core.Band40m,
			focusedVFO: core.VFO1,
			entryBand:  core.Band20m,
			want:       core.VFO1,
		},
		{
			desc:          "band bound to the focused VFO",
			band:          core.SpotFilterBand{Kind: core.SpotFilterFocused},
			vfo2Available: true,
			vfo1Band:      core.Band40m,
			vfo2Band:      core.Band20m,
			focusedVFO:    core.VFO2,
			entryBand:     core.Band20m,
			want:          core.VFO2,
		},
		{
			desc:          "free band, the spot is on the band of VFO2",
			band:          core.SpotFilterBand{Kind: core.SpotFilterAll},
			vfo2Available: true,
			vfo1Band:      core.Band40m,
			vfo2Band:      core.Band20m,
			focusedVFO:    core.VFO1,
			entryBand:     core.Band20m,
			want:          core.VFO2,
		},
		{
			desc:          "free band, the spot is on the band of VFO1",
			band:          core.SpotFilterBand{Kind: core.SpotFilterAll},
			vfo2Available: true,
			vfo1Band:      core.Band40m,
			vfo2Band:      core.Band20m,
			focusedVFO:    core.VFO2,
			entryBand:     core.Band40m,
			want:          core.VFO1,
		},
		{
			desc:          "free band, both VFOs on the band of the spot",
			band:          core.SpotFilterBand{Kind: core.SpotFilterAll},
			vfo2Available: true,
			vfo1Band:      core.Band40m,
			vfo2Band:      core.Band40m,
			focusedVFO:    core.VFO2,
			entryBand:     core.Band40m,
			want:          core.VFO2,
		},
		{
			desc:          "free band, no VFO on the band of the spot",
			band:          core.SpotFilterBand{Kind: core.SpotFilterAll},
			vfo2Available: true,
			vfo1Band:      core.Band40m,
			vfo2Band:      core.Band20m,
			focusedVFO:    core.VFO2,
			entryBand:     core.Band15m,
			want:          core.VFO2,
		},
		{
			desc:          "free band, the spot is on the band of the unavailable VFO2",
			band:          core.SpotFilterBand{Kind: core.SpotFilterAll},
			vfo2Available: false,
			vfo1Band:      core.Band40m,
			vfo2Band:      core.Band20m,
			focusedVFO:    core.VFO1,
			entryBand:     core.Band20m,
			want:          core.VFO1,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			filter, _ := setupSpotFilter()
			filter.RadioChanged("radio", !tc.vfo2Available)
			filter.VFOBandChanged(core.VFO1, tc.vfo1Band)
			filter.VFOBandChanged(core.VFO2, tc.vfo2Band)
			filter.FocusedVFOChanged(tc.focusedVFO)
			filter.SetBand(tc.band)

			assert.Equal(t, tc.want, filter.TargetVFO(core.BandmapEntry{Band: tc.entryBand}))
		})
	}
}

func TestSpotFilter_FrameDescription(t *testing.T) {
	tt := []struct {
		desc string
		band core.SpotFilterBand
		mode core.SpotFilterMode
		sort core.SpotSortColumn
		down bool
		want string
	}{
		{
			desc: "focused VFO",
			band: core.SpotFilterBand{Kind: core.SpotFilterFocused},
			mode: core.SpotFilterMode{Kind: core.SpotFilterFocused},
			sort: core.SortSpotsByFrequency,
			want: "Focused VFO (40m) · Focused VFO (CW) · by Frequency ↑",
		},
		{
			desc: "all bands, fixed mode, descending",
			band: core.SpotFilterBand{Kind: core.SpotFilterAll},
			mode: core.FixedSpotFilterMode(core.ModeSSB),
			sort: core.SortSpotsByValue,
			down: true,
			want: "All bands · SSB · by Value ↓",
		},
		{
			desc: "contest",
			band: core.SpotFilterBand{Kind: core.SpotFilterContest},
			mode: core.SpotFilterMode{Kind: core.SpotFilterContest},
			sort: core.SortSpotsByLastSeen,
			want: "Contest bands · Contest modes · by Last Seen ↑",
		},
		{
			desc: "VFO2",
			band: core.SpotFilterBand{Kind: core.SpotFilterVFO2},
			mode: core.SpotFilterMode{Kind: core.SpotFilterVFO2},
			sort: core.SortSpotsByCallsign,
			want: "VFO2 (20m) · VFO2 (SSB) · by Callsign ↑",
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			filter, _ := setupSpotFilter()
			filter.RadioChanged("radio", false)
			filter.VFOBandChanged(core.VFO1, core.Band40m)
			filter.VFOModeChanged(core.VFO1, core.ModeCW)
			filter.VFOBandChanged(core.VFO2, core.Band20m)
			filter.VFOModeChanged(core.VFO2, core.ModeSSB)
			filter.FocusedVFOChanged(core.VFO1)
			filter.SetBand(tc.band)
			filter.SetMode(tc.mode)
			filter.SetSort(tc.sort, tc.down)

			assert.Equal(t, tc.want, filter.Frame().Description)
		})
	}
}

func TestSpotFilter_FrameOffersAllBandsAndModes(t *testing.T) {
	filter, _ := setupSpotFilter()

	frame := filter.Frame()

	assert.Equal(t, core.Bands, frame.Bands)
	assert.Equal(t, core.Modes, frame.Modes)
}

func setupSpotFilter() (*SpotFilter, *spotFilterStoreStub) {
	store := newSpotFilterStoreStub()
	return NewSpotFilter(MainSpotFilterID, store), store
}

type spotFilterStoreStub struct {
	states map[string]core.SpotFilterState
}

func newSpotFilterStoreStub() *spotFilterStoreStub {
	return &spotFilterStoreStub{
		states: make(map[string]core.SpotFilterState),
	}
}

func (s *spotFilterStoreStub) SpotFilter(id string) (core.SpotFilterState, bool) {
	state, ok := s.states[id]
	return state, ok
}

func (s *spotFilterStoreStub) SetSpotFilter(id string, state core.SpotFilterState) error {
	s.states[id] = state
	return nil
}
