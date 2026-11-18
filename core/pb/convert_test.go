package pb

import (
	"testing"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/model"
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
	if got.Bandplan != model.DefaultBandplanID {
		t.Errorf("bandplan = %q, want default %q", got.Bandplan, model.DefaultBandplanID)
	}
}

func TestContestDetailedMultisRoundtrip(t *testing.T) {
	for _, detailedMultis := range []bool{false, true} {
		got, err := ToContest(ContestToPB(core.Contest{DetailedMultis: detailedMultis}))
		if err != nil {
			t.Fatal(err)
		}
		if got.DetailedMultis != detailedMultis {
			t.Errorf("detailed multis = %v, want %v", got.DetailedMultis, detailedMultis)
		}
	}
}
