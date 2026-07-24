package pb

import (
	"testing"

	"github.com/ftl/hellocontest/core"
)

func TestStationBandplanRoundtrip(t *testing.T) {
	station := core.Station{Bandplan: "iaru-region-2"}
	got, err := ToStation(StationToPB(station))
	if err != nil {
		t.Fatal(err)
	}
	if got.Bandplan != "iaru-region-2" {
		t.Errorf("bandplan = %q, want %q", got.Bandplan, "iaru-region-2")
	}
}

func TestStationBandplanEmptyDefaultsToRegion1(t *testing.T) {
	got, err := ToStation(StationToPB(core.Station{}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Bandplan != core.DefaultBandplanID {
		t.Errorf("bandplan = %q, want default %q", got.Bandplan, core.DefaultBandplanID)
	}
}
