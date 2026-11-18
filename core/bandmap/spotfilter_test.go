package bandmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core/model"
)

func TestSpotFilter_DefaultState(t *testing.T) {
	filter, _ := setupSpotFilter()

	state := filter.State()

	assert.Equal(t, model.SpotFilterFocused, state.Band.Kind, "band follows the focused VFO")
	assert.Equal(t, model.SpotFilterFocused, state.Mode.Kind, "mode follows the focused VFO")
	assert.Equal(t, model.SortSpotsByFrequency, state.SortBy)
	assert.False(t, state.Descending)
	assert.False(t, state.Folded)
}

func TestSpotFilter_RestoresTheStoredState(t *testing.T) {
	store := newSpotFilterStoreStub()
	stored := model.SpotFilterState{
		Band:       model.FixedSpotFilterBand(model.Band20m),
		Mode:       model.FixedSpotFilterMode(model.ModeCW),
		SortBy:     model.SortSpotsByValue,
		Descending: true,
		Folded:     true,
	}
	store.states[MainSpotFilterID] = stored

	filter := NewSpotFilter(MainSpotFilterID, store)

	assert.Equal(t, stored, filter.State())
}

func TestSpotFilter_StoresEveryChange(t *testing.T) {
	filter, store := setupSpotFilter()

	filter.SetBand(model.FixedSpotFilterBand(model.Band40m))
	filter.SetMode(model.FixedSpotFilterMode(model.ModeSSB))
	filter.SetSort(model.SortSpotsByLastSeen, true)
	filter.SetFolded(true)

	stored, ok := store.states[MainSpotFilterID]
	require.True(t, ok)
	assert.Equal(t, model.FixedSpotFilterBand(model.Band40m), stored.Band)
	assert.Equal(t, model.FixedSpotFilterMode(model.ModeSSB), stored.Mode)
	assert.Equal(t, model.SortSpotsByLastSeen, stored.SortBy)
	assert.True(t, stored.Descending)
	assert.True(t, stored.Folded)
}

func TestSpotFilter_NotifiesAboutRelevantChanges(t *testing.T) {
	filter, _ := setupSpotFilter()
	changes := 0
	filter.OnChanged(func() { changes++ })

	filter.SetBand(model.FixedSpotFilterBand(model.Band40m))
	assert.Equal(t, 1, changes, "a new band changes the content")

	filter.SetMode(model.FixedSpotFilterMode(model.ModeSSB))
	assert.Equal(t, 2, changes, "a new mode changes the content")

	filter.SetSort(model.SortSpotsByValue, true)
	assert.Equal(t, 3, changes, "a new order changes the content")

	filter.SetFolded(true)
	assert.Equal(t, 3, changes, "folding does not change the content")
}

func TestSpotFilter_NotifiesOnlyOnResolvedChanges(t *testing.T) {
	filter, _ := setupSpotFilter()
	filter.SetBand(model.SpotFilterBand{Kind: model.SpotFilterVFO1})
	changes := 0
	filter.OnChanged(func() { changes++ })

	filter.VFOBandChanged(model.VFO1, model.Band40m)
	assert.Equal(t, 1, changes, "the band of VFO1 is the resolved band")

	filter.VFOBandChanged(model.VFO1, model.Band40m)
	assert.Equal(t, 1, changes, "the same band again resolves to the same band")

	filter.VFOBandChanged(model.VFO2, model.Band20m)
	assert.Equal(t, 1, changes, "the band of VFO2 does not matter here")

	filter.FocusedVFOChanged(model.VFO2)
	assert.Equal(t, 1, changes, "the focus does not matter here")
}

func TestSpotFilter_Matches(t *testing.T) {
	tt := []struct {
		desc          string
		band          model.SpotFilterBand
		mode          model.SpotFilterMode
		vfo2Available bool
		entry         model.BandmapEntry
		want          bool
	}{
		{
			desc:  "all bands and all modes",
			band:  model.SpotFilterBand{Kind: model.SpotFilterAll},
			mode:  model.SpotFilterMode{Kind: model.SpotFilterAll},
			entry: model.BandmapEntry{Band: model.Band10m, Mode: model.ModeFM},
			want:  true,
		},
		{
			desc:  "fixed band matches",
			band:  model.FixedSpotFilterBand(model.Band40m),
			mode:  model.SpotFilterMode{Kind: model.SpotFilterAll},
			entry: model.BandmapEntry{Band: model.Band40m, Mode: model.ModeCW},
			want:  true,
		},
		{
			desc:  "fixed band does not match",
			band:  model.FixedSpotFilterBand(model.Band40m),
			mode:  model.SpotFilterMode{Kind: model.SpotFilterAll},
			entry: model.BandmapEntry{Band: model.Band20m, Mode: model.ModeCW},
			want:  false,
		},
		{
			desc:  "fixed mode does not match",
			band:  model.SpotFilterBand{Kind: model.SpotFilterAll},
			mode:  model.FixedSpotFilterMode(model.ModeCW),
			entry: model.BandmapEntry{Band: model.Band40m, Mode: model.ModeSSB},
			want:  false,
		},
		{
			desc:  "band of VFO1",
			band:  model.SpotFilterBand{Kind: model.SpotFilterVFO1},
			mode:  model.SpotFilterMode{Kind: model.SpotFilterAll},
			entry: model.BandmapEntry{Band: model.Band40m, Mode: model.ModeCW},
			want:  true,
		},
		{
			desc:          "band of VFO2",
			band:          model.SpotFilterBand{Kind: model.SpotFilterVFO2},
			mode:          model.SpotFilterMode{Kind: model.SpotFilterAll},
			vfo2Available: true,
			entry:         model.BandmapEntry{Band: model.Band20m, Mode: model.ModeSSB},
			want:          true,
		},
		{
			desc:  "band of VFO2 falls back to VFO1 without a second VFO",
			band:  model.SpotFilterBand{Kind: model.SpotFilterVFO2},
			mode:  model.SpotFilterMode{Kind: model.SpotFilterAll},
			entry: model.BandmapEntry{Band: model.Band40m, Mode: model.ModeCW},
			want:  true,
		},
		{
			desc:          "band of the focused VFO",
			band:          model.SpotFilterBand{Kind: model.SpotFilterFocused},
			mode:          model.SpotFilterMode{Kind: model.SpotFilterFocused},
			vfo2Available: true,
			entry:         model.BandmapEntry{Band: model.Band40m, Mode: model.ModeCW},
			want:          true,
		},
		{
			desc:  "contest band",
			band:  model.SpotFilterBand{Kind: model.SpotFilterContest},
			mode:  model.SpotFilterMode{Kind: model.SpotFilterAll},
			entry: model.BandmapEntry{Band: model.Band20m, Mode: model.ModeCW},
			want:  true,
		},
		{
			desc:  "band outside the contest",
			band:  model.SpotFilterBand{Kind: model.SpotFilterContest},
			mode:  model.SpotFilterMode{Kind: model.SpotFilterAll},
			entry: model.BandmapEntry{Band: model.Band10m, Mode: model.ModeCW},
			want:  false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			filter, _ := setupSpotFilter()
			filter.SetContestBandsAndModes([]model.Band{model.Band40m, model.Band20m}, []model.Mode{model.ModeCW})
			filter.RadioChanged("radio", !tc.vfo2Available)
			filter.VFOBandChanged(model.VFO1, model.Band40m)
			filter.VFOModeChanged(model.VFO1, model.ModeCW)
			filter.VFOBandChanged(model.VFO2, model.Band20m)
			filter.VFOModeChanged(model.VFO2, model.ModeSSB)
			filter.FocusedVFOChanged(model.VFO1)
			filter.SetBand(tc.band)
			filter.SetMode(tc.mode)

			assert.Equal(t, tc.want, filter.Matches(tc.entry))
		})
	}
}

func TestSpotFilter_MatchesEverythingWhileUnresolved(t *testing.T) {
	filter, _ := setupSpotFilter()

	assert.True(t, filter.Matches(model.BandmapEntry{Band: model.Band10m, Mode: model.ModeFM}), "no radio, no restriction")
}

func TestSpotFilter_Order(t *testing.T) {
	now := time.Now()
	low := model.BandmapEntry{ID: 1, Call: model.MustParseCallsign("dl1abc"), Frequency: 3500000, LastHeard: now.Add(-time.Minute), Info: model.Callinfo{WeightedValue: 1}}
	high := model.BandmapEntry{ID: 2, Call: model.MustParseCallsign("dl2abc"), Frequency: 7000000, LastHeard: now, Info: model.Callinfo{WeightedValue: 9}}

	for _, column := range model.SpotSortColumns {
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
		band          model.SpotFilterBand
		vfo2Available bool
		vfo1Band      model.Band
		vfo2Band      model.Band
		focusedVFO    model.VFOID
		entryBand     model.Band
		want          model.VFOID
	}{
		{
			desc:          "band bound to VFO1",
			band:          model.SpotFilterBand{Kind: model.SpotFilterVFO1},
			vfo2Available: true,
			vfo1Band:      model.Band40m,
			vfo2Band:      model.Band20m,
			focusedVFO:    model.VFO2,
			entryBand:     model.Band40m,
			want:          model.VFO1,
		},
		{
			desc:          "band bound to VFO2",
			band:          model.SpotFilterBand{Kind: model.SpotFilterVFO2},
			vfo2Available: true,
			vfo1Band:      model.Band40m,
			vfo2Band:      model.Band20m,
			focusedVFO:    model.VFO1,
			entryBand:     model.Band20m,
			want:          model.VFO2,
		},
		{
			desc:       "band bound to VFO2 without a second VFO",
			band:       model.SpotFilterBand{Kind: model.SpotFilterVFO2},
			vfo1Band:   model.Band40m,
			focusedVFO: model.VFO1,
			entryBand:  model.Band20m,
			want:       model.VFO1,
		},
		{
			desc:          "band bound to the focused VFO",
			band:          model.SpotFilterBand{Kind: model.SpotFilterFocused},
			vfo2Available: true,
			vfo1Band:      model.Band40m,
			vfo2Band:      model.Band20m,
			focusedVFO:    model.VFO2,
			entryBand:     model.Band20m,
			want:          model.VFO2,
		},
		{
			desc:          "free band, the spot is on the band of VFO2",
			band:          model.SpotFilterBand{Kind: model.SpotFilterAll},
			vfo2Available: true,
			vfo1Band:      model.Band40m,
			vfo2Band:      model.Band20m,
			focusedVFO:    model.VFO1,
			entryBand:     model.Band20m,
			want:          model.VFO2,
		},
		{
			desc:          "free band, the spot is on the band of VFO1",
			band:          model.SpotFilterBand{Kind: model.SpotFilterAll},
			vfo2Available: true,
			vfo1Band:      model.Band40m,
			vfo2Band:      model.Band20m,
			focusedVFO:    model.VFO2,
			entryBand:     model.Band40m,
			want:          model.VFO1,
		},
		{
			desc:          "free band, both VFOs on the band of the spot",
			band:          model.SpotFilterBand{Kind: model.SpotFilterAll},
			vfo2Available: true,
			vfo1Band:      model.Band40m,
			vfo2Band:      model.Band40m,
			focusedVFO:    model.VFO2,
			entryBand:     model.Band40m,
			want:          model.VFO2,
		},
		{
			desc:          "free band, no VFO on the band of the spot",
			band:          model.SpotFilterBand{Kind: model.SpotFilterAll},
			vfo2Available: true,
			vfo1Band:      model.Band40m,
			vfo2Band:      model.Band20m,
			focusedVFO:    model.VFO2,
			entryBand:     model.Band15m,
			want:          model.VFO2,
		},
		{
			desc:          "free band, the spot is on the band of the unavailable VFO2",
			band:          model.SpotFilterBand{Kind: model.SpotFilterAll},
			vfo2Available: false,
			vfo1Band:      model.Band40m,
			vfo2Band:      model.Band20m,
			focusedVFO:    model.VFO1,
			entryBand:     model.Band20m,
			want:          model.VFO1,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			filter, _ := setupSpotFilter()
			filter.RadioChanged("radio", !tc.vfo2Available)
			filter.VFOBandChanged(model.VFO1, tc.vfo1Band)
			filter.VFOBandChanged(model.VFO2, tc.vfo2Band)
			filter.FocusedVFOChanged(tc.focusedVFO)
			filter.SetBand(tc.band)

			assert.Equal(t, tc.want, filter.TargetVFO(model.BandmapEntry{Band: tc.entryBand}))
		})
	}
}

func TestSpotFilter_TargetVFOForBand_FollowsTheRulesOfASpot(t *testing.T) {
	filter, _ := setupSpotFilter()
	filter.RadioChanged("radio", false)
	filter.VFOBandChanged(model.VFO1, model.Band80m)
	filter.VFOBandChanged(model.VFO2, model.Band20m)
	filter.FocusedVFOChanged(model.VFO1)
	filter.SetBand(model.SpotFilterBand{Kind: model.SpotFilterAll})

	for _, band := range []model.Band{model.Band80m, model.Band20m, model.Band15m} {
		t.Run(string(band), func(t *testing.T) {
			assert.Equal(t, filter.TargetVFO(model.BandmapEntry{Band: band}), filter.TargetVFOForBand(band))
		})
	}
}

func TestSpotFilter_TargetVFOForMarker(t *testing.T) {
	tt := []struct {
		desc          string
		band          model.SpotFilterBand
		vfo2Available bool
		markerBand    model.Band
		want          model.VFOID
	}{
		{
			desc:          "the other VFO is on the band of the marker",
			band:          model.SpotFilterBand{Kind: model.SpotFilterFocused},
			vfo2Available: true,
			markerBand:    model.Band20m,
			want:          model.VFO2,
		},
		{
			desc:          "the focused VFO is on the band of the marker",
			band:          model.SpotFilterBand{Kind: model.SpotFilterFocused},
			vfo2Available: true,
			markerBand:    model.Band80m,
			want:          model.VFO1,
		},
		{
			desc:          "no VFO is on the band of the marker",
			band:          model.SpotFilterBand{Kind: model.SpotFilterAll},
			vfo2Available: true,
			markerBand:    model.Band15m,
			want:          model.VFO1,
		},
		{
			desc:          "the band selection binds the list to VFO1",
			band:          model.SpotFilterBand{Kind: model.SpotFilterVFO1},
			vfo2Available: true,
			markerBand:    model.Band20m,
			want:          model.VFO2,
		},
		{
			desc:       "without a second VFO",
			band:       model.SpotFilterBand{Kind: model.SpotFilterAll},
			markerBand: model.Band20m,
			want:       model.VFO1,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			filter, _ := setupSpotFilter()
			filter.RadioChanged("radio", !tc.vfo2Available)
			filter.VFOBandChanged(model.VFO1, model.Band80m)
			filter.VFOBandChanged(model.VFO2, model.Band20m)
			filter.FocusedVFOChanged(model.VFO1)
			filter.SetBand(tc.band)

			assert.Equal(t, tc.want, filter.TargetVFOForMarker(tc.markerBand),
				"a VFO that is already on the band of the marker goes there, no VFO changes its band")
		})
	}
}

func TestSpotFilter_FrameDescription(t *testing.T) {
	tt := []struct {
		desc string
		band model.SpotFilterBand
		mode model.SpotFilterMode
		sort model.SpotSortColumn
		down bool
		want string
	}{
		{
			desc: "focused VFO",
			band: model.SpotFilterBand{Kind: model.SpotFilterFocused},
			mode: model.SpotFilterMode{Kind: model.SpotFilterFocused},
			sort: model.SortSpotsByFrequency,
			want: "Focused VFO (40m) · Focused VFO (CW) · by Frequency ↑",
		},
		{
			desc: "all bands, fixed mode, descending",
			band: model.SpotFilterBand{Kind: model.SpotFilterAll},
			mode: model.FixedSpotFilterMode(model.ModeSSB),
			sort: model.SortSpotsByValue,
			down: true,
			want: "All bands · SSB · by Value ↓",
		},
		{
			desc: "contest",
			band: model.SpotFilterBand{Kind: model.SpotFilterContest},
			mode: model.SpotFilterMode{Kind: model.SpotFilterContest},
			sort: model.SortSpotsByLastSeen,
			want: "Contest bands · Contest modes · by Last Seen ↑",
		},
		{
			desc: "VFO2",
			band: model.SpotFilterBand{Kind: model.SpotFilterVFO2},
			mode: model.SpotFilterMode{Kind: model.SpotFilterVFO2},
			sort: model.SortSpotsByCallsign,
			want: "VFO2 (20m) · VFO2 (SSB) · by Callsign ↑",
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			filter, _ := setupSpotFilter()
			filter.RadioChanged("radio", false)
			filter.VFOBandChanged(model.VFO1, model.Band40m)
			filter.VFOModeChanged(model.VFO1, model.ModeCW)
			filter.VFOBandChanged(model.VFO2, model.Band20m)
			filter.VFOModeChanged(model.VFO2, model.ModeSSB)
			filter.FocusedVFOChanged(model.VFO1)
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

	assert.Equal(t, model.Bands, frame.Bands)
	assert.Equal(t, model.Modes, frame.Modes)
}

func setupSpotFilter() (*SpotFilter, *spotFilterStoreStub) {
	store := newSpotFilterStoreStub()
	return NewSpotFilter(MainSpotFilterID, store), store
}

type spotFilterStoreStub struct {
	states map[string]model.SpotFilterState
}

func newSpotFilterStoreStub() *spotFilterStoreStub {
	return &spotFilterStoreStub{
		states: make(map[string]model.SpotFilterState),
	}
}

func (s *spotFilterStoreStub) SpotFilter(id string) (model.SpotFilterState, bool) {
	state, ok := s.states[id]
	return state, ok
}

func (s *spotFilterStoreStub) SetSpotFilter(id string, state model.SpotFilterState) error {
	s.states[id] = state
	return nil
}

func TestSpotFilter_Value(t *testing.T) {
	scoringEntry := model.BandmapEntry{Band: model.Band80m, Mode: model.ModeCW}
	scoringEntry.Info.Value = 40

	multiEntry := model.BandmapEntry{Band: model.Band80m, Mode: model.ModeCW}
	multiEntry.Info.Value = 25
	multiEntry.Info.Multis = 1

	worthlessEntry := model.BandmapEntry{Band: model.Band80m, Mode: model.ModeCW}

	duplicateEntry := model.BandmapEntry{Band: model.Band80m, Mode: model.ModeCW}
	duplicateEntry.Info.Value = 40
	duplicateEntry.Info.Multis = 1
	duplicateEntry.Info.Duplicate = true

	tt := []struct {
		desc          string
		value         model.SpotFilterValue
		wantScoring   bool
		wantMulti     bool
		wantWorthless bool
		wantDuplicate bool
	}{
		{desc: "all", value: model.SpotFilterAnyValue, wantScoring: true, wantMulti: true, wantWorthless: true, wantDuplicate: true},
		{desc: "scoring", value: model.SpotFilterScoring, wantScoring: true, wantMulti: true},
		{desc: "multi", value: model.SpotFilterMulti, wantMulti: true},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			filter := NewSpotFilter(MainSpotFilterID, nil)
			filter.SetBand(model.SpotFilterBand{Kind: model.SpotFilterAll})
			filter.SetMode(model.SpotFilterMode{Kind: model.SpotFilterAll})

			filter.SetValue(tc.value)

			assert.Equal(t, tc.wantScoring, filter.Matches(scoringEntry), "the spot with points")
			assert.Equal(t, tc.wantMulti, filter.Matches(multiEntry), "the multiplier")
			assert.Equal(t, tc.wantWorthless, filter.Matches(worthlessEntry), "the spot without value")
			assert.Equal(t, tc.wantDuplicate, filter.Matches(duplicateEntry), "the spot that was worked already")
		})
	}
}

func TestSpotFilter_ValueInTheDescription(t *testing.T) {
	filter := NewSpotFilter(MainSpotFilterID, nil)

	assert.NotContains(t, filter.Frame().Description, "Multi", "the description names no value filter when it shows all spots")

	filter.SetValue(model.SpotFilterMulti)

	assert.Contains(t, filter.Frame().Description, "Multi > 0")
}
