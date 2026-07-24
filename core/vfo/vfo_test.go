package vfo

import (
	"testing"

	"github.com/ftl/hamradio/bandplan"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestSetBandplanSwapsThePlan(t *testing.T) {
	v := NewVFO(core.VFO1, "VFO 1", bandplan.IARURegion1, nil, nil)
	assert.Equal(t, 1810000.0, float64(v.bandplan[bandplan.Band160m].From))

	v.SetBandplan(bandplan.IARURegion2)
	assert.Equal(t, 1800000.0, float64(v.bandplan[bandplan.Band160m].From))
}

func TestBandNameConversion(t *testing.T) {
	bndpln := bandplan.IARURegion1

	for band, plan := range bndpln {
		assert.Equal(t, band, plan.Name)
	}

	for _, band := range core.Bands {
		plan, ok := bndpln[bandplan.BandName(band)]
		assert.True(t, ok, band)
		assert.Equal(t, string(band), string(plan.Name))
	}

}
