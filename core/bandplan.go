package core

import "github.com/ftl/hamradio/bandplan"

// BandplanID identifies a selectable bandplan in its persisted form.
type BandplanID string

func (id BandplanID) String() string {
	return string(id)
}

const (
	BandplanIARURegion1 BandplanID = "iaru-region-1"
	BandplanIARURegion2 BandplanID = "iaru-region-2"
	BandplanIARURegion3 BandplanID = "iaru-region-3"
)

// DefaultBandplanID is the bandplan used when none is selected, preserving the
// historic behavior (IARU Region 1).
const DefaultBandplanID BandplanID = BandplanIARURegion1

// bandplanRegion associates a persisted identifier and a display label with a
// concrete bandplan.
type bandplanRegion struct {
	ID    BandplanID
	Label string
	Plan  bandplan.Bandplan
}

// bandplanRegions lists the selectable bandplans in display order. The first
// entry is the default (see DefaultBandplanID).
var bandplanRegions = []bandplanRegion{
	{BandplanIARURegion1, "IARU Region 1 (Europe, Africa, Middle East, North Asia)", bandplan.IARURegion1},
	{BandplanIARURegion2, "IARU Region 2 (Americas)", bandplan.IARURegion2},
	{BandplanIARURegion3, "IARU Region 3 (Asia-Pacific, Oceania)", bandplan.IARURegion3},
}

// BandplanByID returns the bandplan for the given identifier. An unknown or
// empty identifier yields the default (IARU Region 1).
func BandplanByID(id BandplanID) bandplan.Bandplan {
	for _, r := range bandplanRegions {
		if r.ID == id {
			return r.Plan
		}
	}
	return bandplan.IARURegion1
}

// BandplanIDs returns the selectable bandplan identifiers and their display
// labels in display order, for populating the settings selector.
func BandplanIDs() (ids []BandplanID, labels []string) {
	ids = make([]BandplanID, len(bandplanRegions))
	labels = make([]string, len(bandplanRegions))
	for i, r := range bandplanRegions {
		ids[i] = r.ID
		labels[i] = r.Label
	}
	return ids, labels
}
