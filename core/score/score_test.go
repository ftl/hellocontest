package score

import (
	"strings"
	"testing"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/locator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

var myTestEntity = testEntities{entity: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}

func TestNewCounter(t *testing.T) {
	counter := NewCounter(&testSettings{stationCallsign: "DL1AAA"}, &myTestEntity)
	assert.False(t, counter.invalid, "invalid")
	assert.True(t, counter.Valid(), "valid")
	assert.Equal(t, strings.ToLower(myTestEntity.entity.PrimaryPrefix), string(counter.contestSetup.MyCountry), "station entity")

	counter.ScorePerBand[core.Band80m] = core.BandScore{QSOs: 5}
	assert.Equal(t, 5, counter.ScorePerBand[core.Band80m].QSOs)
}

func TestConvalCounter(t *testing.T) {
	counter := NewCounter(&testSettings{stationCallsign: "DL1AAA"}, &myTestEntity)
	counter.StationChanged(core.Station{
		Callsign: callsign.MustParse("DL0ABC"),
		Operator: callsign.MustParse("DL1ABC"),
		Locator:  locator.MustParse("JN01ab"),
	})
	definition, err := conval.IncludedDefinition("CQ-WPX-CW")
	require.NoError(t, err)
	contest := core.Contest{
		Definition:             definition,
		ExchangeValues:         []string{"599", ""},
		GenerateSerialExchange: true,
	}
	contest.UpdateExchangeFields()
	counter.ContestChanged(contest)
	entity, ok := myTestEntity.Find("dl1aaa")
	require.True(t, ok)

	points, multis := counter.Value(callsign.MustParse("dl1aaa"), entity, core.Band80m, core.ModeCW, []string{})

	assert.Equal(t, 1, points, "points")
	assert.Equal(t, 1, multis, "multis")
}

func TestAdd(t *testing.T) {
	counter := NewCounter(&testSettings{stationCallsign: "DL1AAA"}, &myTestEntity)
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})

	assert.Equal(t, 2, counter.Score.Result().QSOs, "total QSOs")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 2, bandScore.QSOs, "band QSOs")
}

type testSettings struct {
	stationCallsign string
}

func (s *testSettings) Station() core.Station {
	return core.Station{
		Callsign: callsign.MustParse(s.stationCallsign),
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
