package tci

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestDiffMarkers(t *testing.T) {
	cq := core.BandmapMarker{Kind: core.CQMarker, Text: "CQ", Frequency: 3560000, Band: core.Band80m}
	one := core.BandmapMarker{Kind: core.NumberedMarker, Text: "1", Number: 1, Frequency: 3540000, Band: core.Band80m}
	movedOne := core.BandmapMarker{Kind: core.NumberedMarker, Text: "1", Number: 1, Frequency: 3545000, Band: core.Band80m}
	onTwenty := core.BandmapMarker{Kind: core.TextMarker, Text: "beacon", Frequency: 14100000, Band: core.Band20m}

	tt := []struct {
		desc        string
		bands       []core.Band
		sent        map[string]core.BandmapMarker
		markers     []core.BandmapMarker
		wantAdded   []core.BandmapMarker
		wantRemoved []string
	}{
		{
			desc:      "the band of the second TRX",
			bands:     []core.Band{core.Band80m, core.Band20m},
			sent:      map[string]core.BandmapMarker{},
			markers:   []core.BandmapMarker{one, onTwenty},
			wantAdded: []core.BandmapMarker{one, onTwenty},
		},
		{
			desc:      "a new marker",
			sent:      map[string]core.BandmapMarker{},
			markers:   []core.BandmapMarker{cq, one},
			wantAdded: []core.BandmapMarker{cq, one},
		},
		{
			desc:      "an unchanged marker",
			sent:      map[string]core.BandmapMarker{"CQ": cq, "1": one},
			markers:   []core.BandmapMarker{cq, one},
			wantAdded: []core.BandmapMarker{},
		},
		{
			desc:      "a marker on a new frequency",
			sent:      map[string]core.BandmapMarker{"1": one},
			markers:   []core.BandmapMarker{movedOne},
			wantAdded: []core.BandmapMarker{movedOne},
		},
		{
			desc:        "a marker that is gone",
			sent:        map[string]core.BandmapMarker{"CQ": cq, "1": one},
			markers:     []core.BandmapMarker{one},
			wantAdded:   []core.BandmapMarker{},
			wantRemoved: []string{"CQ"},
		},
		{
			desc:        "a marker on another band",
			sent:        map[string]core.BandmapMarker{"beacon": onTwenty},
			markers:     []core.BandmapMarker{one, onTwenty},
			wantAdded:   []core.BandmapMarker{one},
			wantRemoved: []string{"beacon"},
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			bands := tc.bands
			if bands == nil {
				bands = []core.Band{core.Band80m}
			}

			_, added, removed := diffMarkers(tc.sent, tc.markers, bands)

			assert.Equal(t, tc.wantAdded, added, "added markers")
			if len(tc.wantRemoved) == 0 {
				assert.Empty(t, removed, "removed markers")
			} else {
				assert.Equal(t, tc.wantRemoved, removed, "removed markers")
			}
		})
	}
}

func TestDiffMarkers_CurrentMarkersAreTheNextSentMarkers(t *testing.T) {
	one := core.BandmapMarker{Kind: core.NumberedMarker, Text: "1", Number: 1, Frequency: 3540000, Band: core.Band80m}
	onTwenty := core.BandmapMarker{Kind: core.TextMarker, Text: "beacon", Frequency: 14100000, Band: core.Band20m}

	current, _, _ := diffMarkers(map[string]core.BandmapMarker{}, []core.BandmapMarker{one, onTwenty}, []core.Band{core.Band80m})

	assert.Equal(t, map[string]core.BandmapMarker{"1": one}, current, "only the markers on the band of the TRX")
}
