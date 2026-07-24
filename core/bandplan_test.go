package core

import (
	"testing"

	"github.com/ftl/hamradio/bandplan"
)

func TestBandplanByID(t *testing.T) {
	tt := []struct {
		id   BandplanID
		want bandplan.Bandplan
	}{
		{"iaru-region-1", bandplan.IARURegion1},
		{"iaru-region-2", bandplan.IARURegion2},
		{"iaru-region-3", bandplan.IARURegion3},
		{"", bandplan.IARURegion1},
		{"nonsense", bandplan.IARURegion1},
	}
	for _, tc := range tt {
		t.Run(string(tc.id), func(t *testing.T) {
			got := BandplanByID(tc.id)
			// Bandplan is a map; compare identity via a representative band edge.
			if got[bandplan.Band160m].From != tc.want[bandplan.Band160m].From {
				t.Errorf("BandplanByID(%q) returned the wrong plan", tc.id)
			}
		})
	}
}

func TestBandplanIDsMatchDefault(t *testing.T) {
	ids, labels := BandplanIDs()
	if len(ids) != len(labels) || len(ids) == 0 {
		t.Fatalf("ids/labels length mismatch: %d vs %d", len(ids), len(labels))
	}
	if ids[0] != DefaultBandplanID {
		t.Errorf("first id = %q, want default %q", ids[0], DefaultBandplanID)
	}
}
