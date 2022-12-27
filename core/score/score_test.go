package score

import (
	"fmt"
	"regexp"
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
	assert.Equal(t, myTestEntity.entity, counter.stationEntity, "station entity")

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

func TestAddDuplicate(t *testing.T) {
	counter := NewCounter(&testSettings{stationCallsign: "DL1AAA"}, &myTestEntity)
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Duplicate: true, Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})

	assert.Equal(t, 1, counter.Score.Result().QSOs, "total QSOs")
	assert.Equal(t, 1, counter.Score.Result().Duplicates, "total duplicates")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 1, bandScore.QSOs, "band QSOs")
	assert.Equal(t, 1, bandScore.Duplicates, "band duplicates")
}

func TestUpdateToDuplicate(t *testing.T) {
	anotherQSO := core.QSO{Callsign: callsign.MustParse("DK0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DK", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	oldQSO := core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	newQSO := core.QSO{Duplicate: true, Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	counter := NewCounter(&testSettings{stationCallsign: "DL1AAA"}, &myTestEntity)
	counter.Add(anotherQSO)
	counter.Add(oldQSO)
	counter.Update(oldQSO, newQSO)

	assert.Equal(t, 1, counter.Score.Result().QSOs, "total QSOs")
	assert.Equal(t, 1, counter.Score.Result().Duplicates, "total duplicates")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 1, bandScore.QSOs, "band QSOs")
	assert.Equal(t, 1, bandScore.Duplicates, "band duplicates")
}

func TestUpdateFromDuplicate(t *testing.T) {
	anotherQSO := core.QSO{Callsign: callsign.MustParse("DK0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DK", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	oldQSO := core.QSO{Duplicate: true, Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	newQSO := core.QSO{Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	counter := NewCounter(&testSettings{stationCallsign: "DL1AAA"}, &myTestEntity)
	counter.Add(anotherQSO)
	counter.Add(oldQSO)
	counter.Update(oldQSO, newQSO)

	assert.Equal(t, 2, counter.Score.Result().QSOs, "total QSOs")
	assert.Equal(t, 0, counter.Score.Result().Duplicates, "total duplicates")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 2, bandScore.QSOs, "band QSOs")
	assert.Equal(t, 0, bandScore.Duplicates, "band duplicates")
}

func TestUpdateSameBandAndPrimaryPrefix(t *testing.T) {
	oldQSO := core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	newQSO := core.QSO{Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	counter := NewCounter(&testSettings{stationCallsign: "DL1AAA"}, &myTestEntity)
	counter.Add(oldQSO)
	counter.Update(oldQSO, newQSO)

	assert.Equal(t, 1, counter.Score.Result().QSOs, "total QSOs")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 1, bandScore.QSOs, "band QSOs")
}

func TestMatchXchange(t *testing.T) {
	tt := []struct {
		expression string
		value      string
		multi      string
		match      bool
	}{
		{"", "no expression", "NO EXPRESSION", true},
		{`\d+`, "no digit", "", false},
		{`(\d+)|([A-Za-z]+)`, "digitsorchars", "DIGITSORCHARS", true},
		{`(?P<multi>\d+)|([A-Za-z]+)`, "digitsaremulti", "", false},
		{`(?P<multi>\d+)|([A-Za-z]+)`, "123", "123", true},
		{`(\d+)|(\d*(?P<multi>[A-Za-z])[A-Za-z]*\d*)`, "123", "", false},
		{`(\d+)|(\d*(?P<multi>[A-Za-z])[A-Za-z]*\d*)`, "b36", "B", true},
		{`(\d+[A-Za-z]+)|([A-Za-z]+\d+)|(\d+[A-Za-z]+\d+)`, "nm", "", false},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("%s %s", tc.expression, tc.value), func(t *testing.T) {
			var expression *regexp.Regexp
			if tc.expression != "" {
				expression = regexp.MustCompile(tc.expression)
			}
			m := &multis{XchangeMultiExpression: expression}
			actualMulti, actualMatch := m.matchXchange(tc.value)
			assert.Equal(t, tc.multi, actualMulti)
			assert.Equal(t, tc.match, actualMatch)
		})
	}
}

func TestWPXPrefix(t *testing.T) {
	tt := []struct {
		call     string
		expected string
	}{
		{"DL1ABC", "DL1"},
		{"9A1A", "9A1"},
		{"LY1000A", "LY1000"},
		{"DL/9A1A", "DL0"},
		{"N8BJQ/KH9", "KH9"},
		{"N8BJQ/9", "N8"},
		{"DL1ABC/P", "DL1"},
	}
	for _, tc := range tt {
		t.Run(tc.call, func(t *testing.T) {
			actual := WPXPrefix(callsign.MustParse(tc.call))
			assert.Equal(t, tc.expected, actual)
		})
	}
}

type testSettings struct {
	stationCallsign         string
	countPerBand            bool
	sameCountryPoints       int
	sameContinentPoints     int
	otherPoints             int
	specificCountryPoints   int
	specificCountryPrefixes []string
	multis                  core.Multis
	xchangeMultiPattern     string
}

func (s *testSettings) Station() core.Station {
	return core.Station{
		Callsign: callsign.MustParse(s.stationCallsign),
	}
}

func (s *testSettings) Contest() core.Contest {
	return core.Contest{
		SameCountryPoints:       s.sameCountryPoints,
		SameContinentPoints:     s.sameContinentPoints,
		SpecificCountryPoints:   s.specificCountryPoints,
		SpecificCountryPrefixes: s.specificCountryPrefixes,
		OtherPoints:             s.otherPoints,
		Multis:                  s.multis,
		XchangeMultiPattern:     s.xchangeMultiPattern,
		CountPerBand:            s.countPerBand,
	}
}

type testEntities struct {
	entity dxcc.Prefix
}

func (e *testEntities) Find(string) (dxcc.Prefix, bool) {
	return e.entity, (e.entity.PrimaryPrefix != "")
}
