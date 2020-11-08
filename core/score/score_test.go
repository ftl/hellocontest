package score

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hellocontest/core"
)

func TestNewCounter(t *testing.T) {
	counter := NewCounter()
	counter.ScorePerBand[core.Band80m] = core.BandScore{SameCountryQSOs: 5}
	assert.Equal(t, 5, counter.ScorePerBand[core.Band80m].SameCountryQSOs)
}

func TestAdd(t *testing.T) {
	counter := NewCounter()
	counter.SetMyPrefix(dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", CQZone: 14, ITUZone: 28})
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", CQZone: 14, ITUZone: 28}})

	assert.Equal(t, 2, counter.TotalScore.SameCountryQSOs, "total same country")
	assert.Equal(t, 1, counter.TotalScore.CQZones, "total cq")
	assert.Equal(t, 1, counter.TotalScore.ITUZones, "total itu")
	assert.Equal(t, 1, counter.TotalScore.PrimaryPrefixes, "total prefixes")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 2, bandScore.SameCountryQSOs, "band same country")
	assert.Equal(t, 1, bandScore.CQZones, "band cq")
	assert.Equal(t, 1, bandScore.ITUZones, "band itu")
	assert.Equal(t, 1, bandScore.PrimaryPrefixes, "band prefixes")
}

func TestUpdateSameBandAndPrimaryPrefix(t *testing.T) {
	oldQSO := core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", CQZone: 14, ITUZone: 28}}
	newQSO := core.QSO{Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", CQZone: 14, ITUZone: 28}}
	counter := NewCounter()
	counter.SetMyPrefix(dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", CQZone: 14, ITUZone: 28})
	counter.Add(oldQSO)
	counter.Update(oldQSO, newQSO)

	assert.Equal(t, 1, counter.TotalScore.SameCountryQSOs, "total same country")
	assert.Equal(t, 1, counter.TotalScore.CQZones, "total cq")
	assert.Equal(t, 1, counter.TotalScore.ITUZones, "total itu")
	assert.Equal(t, 1, counter.TotalScore.PrimaryPrefixes, "total prefixes")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 1, bandScore.SameCountryQSOs, "band same country")
	assert.Equal(t, 1, bandScore.CQZones, "band cq")
	assert.Equal(t, 1, bandScore.ITUZones, "band itu")
	assert.Equal(t, 1, bandScore.PrimaryPrefixes, "band prefixes")
}
