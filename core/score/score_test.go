package score

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

var myTestEntity = dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}

func TestNewCounter(t *testing.T) {
	counter := NewCounter(new(testConfig))
	counter.ScorePerBand[core.Band80m] = core.BandScore{SameCountryQSOs: 5}
	assert.Equal(t, 5, counter.ScorePerBand[core.Band80m].SameCountryQSOs)
}

func TestAdd(t *testing.T) {
	counter := NewCounter(new(testConfig))
	counter.SetMyEntity(myTestEntity)
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})

	assert.Equal(t, 2, counter.TotalScore.SameCountryQSOs, "total same country")
	assert.Equal(t, 1, counter.TotalScore.CQZones, "total cq")
	assert.Equal(t, 1, counter.TotalScore.ITUZones, "total itu")
	assert.Equal(t, 1, counter.TotalScore.DXCCEntities, "total dxcc entities")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 2, bandScore.SameCountryQSOs, "band same country")
	assert.Equal(t, 1, bandScore.CQZones, "band cq")
	assert.Equal(t, 1, bandScore.ITUZones, "band itu")
	assert.Equal(t, 1, bandScore.DXCCEntities, "band dxcc entities")
}

func TestAddDuplicate(t *testing.T) {
	counter := NewCounter(new(testConfig))
	counter.SetMyEntity(myTestEntity)
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Duplicate: true, Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})

	assert.Equal(t, 1, counter.TotalScore.SameCountryQSOs, "total same country")
	assert.Equal(t, 1, counter.TotalScore.CQZones, "total cq")
	assert.Equal(t, 1, counter.TotalScore.ITUZones, "total itu")
	assert.Equal(t, 1, counter.TotalScore.DXCCEntities, "total dxcc entities")
	assert.Equal(t, 1, counter.TotalScore.Duplicates, "total duplicates")

	assert.Equal(t, 1, counter.OverallScore.Duplicates, "overall duplicates")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 1, bandScore.SameCountryQSOs, "band same country")
	assert.Equal(t, 1, bandScore.CQZones, "band cq")
	assert.Equal(t, 1, bandScore.ITUZones, "band itu")
	assert.Equal(t, 1, bandScore.DXCCEntities, "band dxcc entities")
	assert.Equal(t, 1, bandScore.Duplicates, "band duplicates")
}

func TestUpdateToDuplicate(t *testing.T) {
	anotherQSO := core.QSO{Callsign: callsign.MustParse("DK0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DK", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	oldQSO := core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	newQSO := core.QSO{Duplicate: true, Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	counter := NewCounter(new(testConfig))
	counter.SetMyEntity(myTestEntity)
	counter.Add(anotherQSO)
	counter.Add(oldQSO)
	counter.Update(oldQSO, newQSO)

	assert.Equal(t, 1, counter.TotalScore.SameCountryQSOs, "total same country")
	assert.Equal(t, 1, counter.TotalScore.CQZones, "total cq")
	assert.Equal(t, 1, counter.TotalScore.ITUZones, "total itu")
	assert.Equal(t, 1, counter.TotalScore.DXCCEntities, "total dxcc entities")
	assert.Equal(t, 1, counter.TotalScore.Duplicates, "total duplicates")

	assert.Equal(t, 1, counter.OverallScore.Duplicates, "overall duplicates")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 1, bandScore.SameCountryQSOs, "band same country")
	assert.Equal(t, 1, bandScore.CQZones, "band cq")
	assert.Equal(t, 1, bandScore.ITUZones, "band itu")
	assert.Equal(t, 1, bandScore.DXCCEntities, "band dxcc entities")
	assert.Equal(t, 1, bandScore.Duplicates, "band duplicates")
}

func TestUpdateFromDuplicate(t *testing.T) {
	anotherQSO := core.QSO{Callsign: callsign.MustParse("DK0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DK", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	oldQSO := core.QSO{Duplicate: true, Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	newQSO := core.QSO{Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	counter := NewCounter(new(testConfig))
	counter.SetMyEntity(myTestEntity)
	counter.Add(anotherQSO)
	counter.Add(oldQSO)
	counter.Update(oldQSO, newQSO)

	assert.Equal(t, 2, counter.TotalScore.SameCountryQSOs, "total same country")
	assert.Equal(t, 1, counter.TotalScore.CQZones, "total cq")
	assert.Equal(t, 1, counter.TotalScore.ITUZones, "total itu")
	assert.Equal(t, 1, counter.TotalScore.DXCCEntities, "total dxcc entities")
	assert.Equal(t, 0, counter.TotalScore.Duplicates, "total duplicates")

	assert.Equal(t, 0, counter.OverallScore.Duplicates, "overall duplicates")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 2, bandScore.SameCountryQSOs, "band same country")
	assert.Equal(t, 1, bandScore.CQZones, "band cq")
	assert.Equal(t, 1, bandScore.ITUZones, "band itu")
	assert.Equal(t, 1, bandScore.DXCCEntities, "band dxcc entities")
	assert.Equal(t, 0, bandScore.Duplicates, "band duplicates")
}

func TestUpdateSameBandAndPrimaryPrefix(t *testing.T) {
	oldQSO := core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	newQSO := core.QSO{Callsign: callsign.MustParse("DF0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DF", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}
	counter := NewCounter(new(testConfig))
	counter.SetMyEntity(myTestEntity)
	counter.Add(oldQSO)
	counter.Update(oldQSO, newQSO)

	assert.Equal(t, 1, counter.TotalScore.SameCountryQSOs, "total same country")
	assert.Equal(t, 1, counter.TotalScore.CQZones, "total cq")
	assert.Equal(t, 1, counter.TotalScore.ITUZones, "total itu")
	assert.Equal(t, 1, counter.TotalScore.DXCCEntities, "total dxcc entities")

	assert.Equal(t, 1, len(counter.ScorePerBand))
	bandScore := counter.ScorePerBand[core.Band80m]
	assert.Equal(t, 1, bandScore.SameCountryQSOs, "band same country")
	assert.Equal(t, 1, bandScore.CQZones, "band cq")
	assert.Equal(t, 1, bandScore.ITUZones, "band itu")
	assert.Equal(t, 1, bandScore.DXCCEntities, "band dxcc entities")
}

func TestCalculatePoints(t *testing.T) {
	counter := NewCounter(&testConfig{
		sameCountryPoints:   1,
		sameContinentPoints: 3,
		otherPoints:         5,
	})
	counter.SetMyEntity(myTestEntity)
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("EA0ABC"), Band: core.Band40m, DXCC: dxcc.Prefix{Prefix: "EA", PrimaryPrefix: "EA", Continent: "EU", CQZone: 14, ITUZone: 37}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("K0ABC"), Band: core.Band20m, DXCC: dxcc.Prefix{Prefix: "K", PrimaryPrefix: "K", Continent: "NA", CQZone: 5, ITUZone: 8}})

	assert.Equal(t, 1, counter.ScorePerBand[core.Band80m].Points, "same country")
	assert.Equal(t, 3, counter.ScorePerBand[core.Band40m].Points, "same continent")
	assert.Equal(t, 5, counter.ScorePerBand[core.Band20m].Points, "other")
	assert.Equal(t, 9, counter.TotalScore.Points, "total")
	assert.Equal(t, 9, counter.OverallScore.Points, "overall")
}

func TestCalculatePointsForSpecificCountry(t *testing.T) {
	counter := NewCounter(&testConfig{
		specificCountryPoints:   1,
		specificCountryPrefixes: []string{"DL", "EA"},
	})
	counter.SetMyEntity(myTestEntity)
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("EA0ABC"), Band: core.Band40m, DXCC: dxcc.Prefix{Prefix: "EA", PrimaryPrefix: "EA", Continent: "EU", CQZone: 14, ITUZone: 37}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("K0ABC"), Band: core.Band20m, DXCC: dxcc.Prefix{Prefix: "K", PrimaryPrefix: "K", Continent: "NA", CQZone: 5, ITUZone: 8}})

	assert.Equal(t, 1, counter.ScorePerBand[core.Band80m].Points, "DL")
	assert.Equal(t, 1, counter.ScorePerBand[core.Band40m].Points, "EA")
	assert.Equal(t, 0, counter.ScorePerBand[core.Band20m].Points, "other")
	assert.Equal(t, 2, counter.TotalScore.Points, "total")
	assert.Equal(t, 2, counter.OverallScore.Points, "overall")
}

func TestCalculateMultipliers(t *testing.T) {
	counter := NewCounter(&testConfig{
		multis: []string{"CQ"},
	})
	counter.SetMyEntity(myTestEntity)
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("EA0ABC"), Band: core.Band40m, DXCC: dxcc.Prefix{Prefix: "EA", PrimaryPrefix: "EA", Continent: "EU", CQZone: 14, ITUZone: 37}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("K0ABC"), Band: core.Band20m, DXCC: dxcc.Prefix{Prefix: "K", PrimaryPrefix: "K", Continent: "NA", CQZone: 5, ITUZone: 8}})

	assert.Equal(t, 1, counter.ScorePerBand[core.Band80m].Multis, "same country")
	assert.Equal(t, 1, counter.ScorePerBand[core.Band40m].Multis, "same continent")
	assert.Equal(t, 1, counter.ScorePerBand[core.Band20m].Multis, "other")
	assert.Equal(t, 3, counter.TotalScore.Multis, "total")
	assert.Equal(t, 2, counter.OverallScore.Multis, "overall")
}

func TestCalculateMultipliersForDistinctXchangeValues(t *testing.T) {
	counter := NewCounter(&testConfig{
		multis:              []string{"Xchange"},
		xchangeMultiPattern: `(\d+)|(\d*(?P<multi>[A-Za-z])[A-Za-z]*\d*)`,
	})
	counter.SetMyEntity(myTestEntity)
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, TheirXchange: "B36", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("DF0ABC"), Band: core.Band40m, TheirXchange: "B05", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("K0ABC"), Band: core.Band20m, TheirXchange: "001", DXCC: dxcc.Prefix{Prefix: "K", PrimaryPrefix: "K", Continent: "NA", CQZone: 5, ITUZone: 8}})

	assert.Equal(t, 1, counter.ScorePerBand[core.Band80m].Multis, "80m")
	assert.Equal(t, 1, counter.ScorePerBand[core.Band40m].Multis, "40m")
	assert.Equal(t, 0, counter.ScorePerBand[core.Band20m].Multis, "20m")
	assert.Equal(t, 2, counter.TotalScore.Multis, "total")
	assert.Equal(t, 1, counter.OverallScore.Multis, "overall")
}

func TestCalculateMutlipliersForWPXPrefixes(t *testing.T) {
	counter := NewCounter(&testConfig{
		multis: []string{"WPX"},
	})
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC"), Band: core.Band80m, TheirXchange: "B36", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("PA/DL0ABC"), Band: core.Band40m, TheirXchange: "B36", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("DL0ABC/P"), Band: core.Band20m, TheirXchange: "B36", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("9A1A"), Band: core.Band20m, TheirXchange: "001", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("LY1000A"), Band: core.Band15m, TheirXchange: "001", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("N8BJQ/KH9"), Band: core.Band10m, TheirXchange: "001", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("NG2M/KH9"), Band: core.Band10m, TheirXchange: "001", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})
	counter.Add(core.QSO{Callsign: callsign.MustParse("N8BJQ/9"), Band: core.Band10m, TheirXchange: "001", DXCC: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}})

	assert.Equal(t, 1, counter.ScorePerBand[core.Band80m].Multis, "80m")
	assert.Equal(t, 1, counter.ScorePerBand[core.Band40m].Multis, "40m")
	assert.Equal(t, 2, counter.ScorePerBand[core.Band20m].Multis, "20m")
	assert.Equal(t, 1, counter.ScorePerBand[core.Band15m].Multis, "15m")
	assert.Equal(t, 2, counter.ScorePerBand[core.Band10m].Multis, "10m")
	assert.Equal(t, 7, counter.TotalScore.Multis, "total")
	assert.Equal(t, 6, counter.OverallScore.Multis, "overall")
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

type testConfig struct {
	countPerBand            bool
	sameCountryPoints       int
	sameContinentPoints     int
	otherPoints             int
	specificCountryPoints   int
	specificCountryPrefixes []string
	multis                  []string
	xchangeMultiPattern     string
}

func (c *testConfig) CountPerBand() bool {
	return c.countPerBand
}

func (c *testConfig) SameCountryPoints() int {
	return c.sameCountryPoints
}

func (c *testConfig) SameContinentPoints() int {
	return c.sameContinentPoints
}

func (c *testConfig) OtherPoints() int {
	return c.otherPoints
}

func (c *testConfig) SpecificCountryPoints() int {
	return c.specificCountryPoints
}

func (c *testConfig) SpecificCountryPrefixes() []string {
	return c.specificCountryPrefixes
}

func (c *testConfig) Multis() []string {
	return c.multis
}

func (c *testConfig) XchangeMultiPattern() string {
	return c.xchangeMultiPattern
}
