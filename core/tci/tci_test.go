package tci

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core/model"
)

func TestDiffMarkers(t *testing.T) {
	cq := model.BandmapMarker{Kind: model.CQMarker, Text: "CQ", Frequency: 3560000, Band: model.Band80m}
	one := model.BandmapMarker{Kind: model.NumberedMarker, Text: "1", Number: 1, Frequency: 3540000, Band: model.Band80m}
	movedOne := model.BandmapMarker{Kind: model.NumberedMarker, Text: "1", Number: 1, Frequency: 3545000, Band: model.Band80m}
	onTwenty := model.BandmapMarker{Kind: model.TextMarker, Text: "beacon", Frequency: 14100000, Band: model.Band20m}

	tt := []struct {
		desc        string
		bands       []model.Band
		sent        map[string]model.BandmapMarker
		markers     []model.BandmapMarker
		wantAdded   []model.BandmapMarker
		wantRemoved []string
	}{
		{
			desc:      "the band of the second TRX",
			bands:     []model.Band{model.Band80m, model.Band20m},
			sent:      map[string]model.BandmapMarker{},
			markers:   []model.BandmapMarker{one, onTwenty},
			wantAdded: []model.BandmapMarker{one, onTwenty},
		},
		{
			desc:      "a new marker",
			sent:      map[string]model.BandmapMarker{},
			markers:   []model.BandmapMarker{cq, one},
			wantAdded: []model.BandmapMarker{cq, one},
		},
		{
			desc:      "an unchanged marker",
			sent:      map[string]model.BandmapMarker{"CQ": cq, "1": one},
			markers:   []model.BandmapMarker{cq, one},
			wantAdded: []model.BandmapMarker{},
		},
		{
			desc:      "a marker on a new frequency",
			sent:      map[string]model.BandmapMarker{"1": one},
			markers:   []model.BandmapMarker{movedOne},
			wantAdded: []model.BandmapMarker{movedOne},
		},
		{
			desc:        "a marker that is gone",
			sent:        map[string]model.BandmapMarker{"CQ": cq, "1": one},
			markers:     []model.BandmapMarker{one},
			wantAdded:   []model.BandmapMarker{},
			wantRemoved: []string{"CQ"},
		},
		{
			desc:        "a marker on another band",
			sent:        map[string]model.BandmapMarker{"beacon": onTwenty},
			markers:     []model.BandmapMarker{one, onTwenty},
			wantAdded:   []model.BandmapMarker{one},
			wantRemoved: []string{"beacon"},
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			bands := tc.bands
			if bands == nil {
				bands = []model.Band{model.Band80m}
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
	one := model.BandmapMarker{Kind: model.NumberedMarker, Text: "1", Number: 1, Frequency: 3540000, Band: model.Band80m}
	onTwenty := model.BandmapMarker{Kind: model.TextMarker, Text: "beacon", Frequency: 14100000, Band: model.Band20m}

	current, _, _ := diffMarkers(map[string]model.BandmapMarker{}, []model.BandmapMarker{one, onTwenty}, []model.Band{model.Band80m})

	assert.Equal(t, map[string]model.BandmapMarker{"1": one}, current, "only the markers on the band of the TRX")
}

func TestEmitUsesTheAsyncRunner(t *testing.T) {
	testCases := []struct {
		desc string
		emit func(c *Client)
	}{
		{
			desc: "connection changed",
			emit: func(c *Client) { c.emitConnectionChanged(true) },
		},
		{
			desc: "frequency changed",
			emit: func(c *Client) { c.emitFrequencyChanged(model.VFO1, 3500000) },
		},
		{
			desc: "band changed",
			emit: func(c *Client) { c.emitBandChanged(model.VFO1, model.Band80m) },
		},
		{
			desc: "mode changed",
			emit: func(c *Client) { c.emitModeChanged(model.VFO1, model.ModeCW) },
		},
		{
			desc: "incremental tuning changed",
			emit: func(c *Client) { c.emitIncrementalTuningChanged(model.VFO1, model.RIT, true, 100) },
		},
		{
			desc: "PTT changed",
			emit: func(c *Client) { c.emitPTTChanged(model.VFO1, true) },
		},
		{
			desc: "TX VFO changed",
			emit: func(c *Client) { c.emitTXVFOChanged(model.VFO1) },
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			runCount := 0
			c := &Client{asyncRunner: func(f func()) {
				runCount++
				f()
			}}

			tc.emit(c)

			assert.Equal(t, 1, runCount, "the listeners must be notified through the async runner")
		})
	}
}
