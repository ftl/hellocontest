package settings

import (
	"testing"

	"github.com/ftl/hellocontest/core"
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
