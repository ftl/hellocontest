package settings

import (
	"testing"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
)

type recordingView struct {
	*nullView
	bandplanOptions []string
	bandplan        string
}

func (v *recordingView) Ready() bool { return true }
func (v *recordingView) SetStationBandplanOptions(ids, labels []string) {
	v.bandplanOptions = ids
}
func (v *recordingView) SetStationBandplan(id string) { v.bandplan = id }

func TestShowSettingsSelectsConfiguredBandplan(t *testing.T) {
	station := core.Station{Bandplan: core.BandplanIARURegion3}
	s := New(nil, nil, nil, nil, station, core.Contest{})

	view := &recordingView{nullView: new(nullView)}
	s.SetView(view)

	if view.bandplan != "iaru-region-3" {
		t.Errorf("selected bandplan = %q, want %q", view.bandplan, "iaru-region-3")
	}
}

// mirrors the New()-contest sequence: an old log (no bandplan -> region-1) is
// loaded, then Reset() restores the config default, then the dialog is shown.
func TestNewContestResetRestoresConfiguredBandplan(t *testing.T) {
	config := core.Station{Bandplan: core.BandplanIARURegion3}
	s := New(nil, nil, nil, nil, config, core.Contest{})

	view := &recordingView{nullView: new(nullView)}
	s.SetView(view)

	s.SetStation(core.Station{Bandplan: core.BandplanIARURegion1}) // old log loaded
	s.Reset()
	s.showSettings()

	if view.bandplan != "iaru-region-3" {
		t.Errorf("after reset, selected bandplan = %q, want %q", view.bandplan, "iaru-region-3")
	}
}

type detailedMultisView struct {
	*nullView
	value   bool
	enabled bool
}

func (v *detailedMultisView) Ready() bool { return true }
func (v *detailedMultisView) SetContestDetailedMultis(value bool, enabled bool) {
	v.value = value
	v.enabled = enabled
}

// WAG counts the entities for a german station and the districts for all other stations,
// therefore only one property applies and the option stays disabled
var wagDefinition = &conval.Definition{Scoring: conval.Scoring{MultiRules: []conval.ScoringRule{
	{Property: conval.DXCCEntityProperty, MyCountry: []conval.DXCCEntity{"dl"}},
	{Property: "wag_district", MyCountry: []conval.DXCCEntity{conval.NotCountry, "dl"}},
}}}

// CQ WW counts zones and entities for every station
var cqwwDefinition = &conval.Definition{Scoring: conval.Scoring{MultiRules: []conval.ScoringRule{
	{Property: conval.CQZoneProperty},
	{Property: conval.DXCCEntityProperty},
}}}

type testEntities struct {
	entity dxcc.Prefix
}

func (e *testEntities) Find(string) (dxcc.Prefix, bool) {
	return e.entity, e.entity.PrimaryPrefix != ""
}

func TestShowSettingsOffersTheDetailedMultisOnlyWithSeveralProperties(t *testing.T) {
	tt := []struct {
		desc            string
		definition      *conval.Definition
		detailedMultis  bool
		expectedValue   bool
		expectedEnabled bool
	}{
		{desc: "two properties", definition: cqwwDefinition, detailedMultis: true, expectedValue: true, expectedEnabled: true},
		{desc: "two properties, option off", definition: cqwwDefinition, expectedEnabled: true},
		{desc: "one property for this station", definition: wagDefinition, detailedMultis: true},
		{desc: "no contest definition", detailedMultis: true},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			station := core.Station{Callsign: core.MustParseCallsign("DL1ABC")}
			contest := core.Contest{Definition: tc.definition, DetailedMultis: tc.detailedMultis}
			s := New(nil, nil, nil, nil, station, contest)
			s.SetDXCCEntities(&testEntities{entity: dxcc.Prefix{PrimaryPrefix: "DL", Continent: "EU"}})
			view := &detailedMultisView{nullView: new(nullView)}

			s.SetView(view)

			if view.value != tc.expectedValue {
				t.Errorf("value = %v, want %v", view.value, tc.expectedValue)
			}
			if view.enabled != tc.expectedEnabled {
				t.Errorf("enabled = %v, want %v", view.enabled, tc.expectedEnabled)
			}
		})
	}
}
