package score

import (
	"testing"

	"github.com/ftl/hamradio/dxcc"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

var myTestEntity = testEntities{entity: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}

func TestNewController(t *testing.T) {
	counter := NewController(&testSettings{stationCallsign: "DL1AAA"})
	counter.score.ScorePerBand[core.Band80m] = core.BandScore{QSOs: 5}
	assert.Equal(t, 5, counter.score.ScorePerBand[core.Band80m].QSOs)
}

type testSettings struct {
	stationCallsign string
}

func (s *testSettings) Station() core.Station {
	return core.Station{
		Callsign: core.MustParseCallsign(s.stationCallsign),
	}
}

func (s *testSettings) Contest() core.Contest {
	return core.Contest{}
}

type testEntities struct {
	entity dxcc.Prefix
}

func (e *testEntities) Find(string) (dxcc.Prefix, bool) {
	return e.entity, (e.entity.PrimaryPrefix != "")
}
