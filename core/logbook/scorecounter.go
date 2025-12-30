package logbook

import (
	"log"
	"strings"
	"time"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
)

type scoreCounter struct {
	score          core.Score
	counter        convalCounter
	prefixDatabase prefixDatabase
	invalid        bool

	contestSetup        conval.Setup
	contestDefinition   *conval.Definition
	contestStartTime    time.Time
	myExchangeFields    []conval.ExchangeField
	theirExchangeFields []conval.ExchangeField

	counterFactory   func() convalCounter
	timeSheetFactory func() convalTimeSheet
}

func newScoreCounter(settings core.Settings, entities DXCCEntities) *scoreCounter {
	result := &scoreCounter{
		score:          core.NewScore(),
		counter:        newSafeConvalCounter(new(nullCounter), new(nullTimeSheet)),
		prefixDatabase: prefixDatabase{entities},
	}
	result.counterFactory = result.newConvalCounter
	result.timeSheetFactory = result.newConvalTimeSheet

	result.StationChanged(settings.Station())
	result.ContestChanged(settings.Contest())

	return result
}

func (c *scoreCounter) StationChanged(station core.Station) {
	oldSetup := c.contestSetup

	continent, country, _, _, found := c.prefixDatabase.Find(station.Callsign.String())
	if !found {
		log.Printf("No DXCC entity found for the station callsign %s", station.Callsign)
		return
	}

	c.contestSetup = conval.Setup{
		MyCall:      station.Callsign,
		MyContinent: continent,
		MyCountry:   country,
		GridLocator: station.Locator,
	}
	log.Printf("Using %+v as station setup", c.contestSetup)

	c.invalid = (oldSetup.MyCountry != c.contestSetup.MyCountry)
	c.clear()
}

func (c *scoreCounter) ContestChanged(contest core.Contest) {
	c.contestDefinition = contest.Definition
	c.contestStartTime = contest.StartTime
	c.myExchangeFields = toConvalExchangeFields(contest.MyExchangeFields)
	c.theirExchangeFields = toConvalExchangeFields(contest.TheirExchangeFields)

	c.invalid = true
	c.clear()
}

func (c *scoreCounter) clear() {
	counter := c.counterFactory()
	timeSheet := c.timeSheetFactory()
	c.counter = newSafeConvalCounter(counter, timeSheet)
	c.score = core.NewScore()
}

func (c *scoreCounter) newConvalTimeSheet() convalTimeSheet {
	if c.contestDefinition == nil {
		return new(nullTimeSheet)
	}
	return conval.NewTimeSheet(c.contestStartTime, c.contestDefinition.Duration)
}

func (c *scoreCounter) newConvalCounter() convalCounter {
	if c.contestDefinition == nil {
		return new(nullCounter)
	}
	return conval.NewCounter(*c.contestDefinition, c.contestSetup, c.prefixDatabase)
}

func (c *scoreCounter) Valid() bool {
	return !c.invalid && (c.contestSetup.MyCountry != "") && (c.contestSetup.MyContinent != "")
}

func (c *scoreCounter) Clear() {
	c.clear()
	c.invalid = (c.contestSetup.MyCountry == "")
}

// Fill the score counter with the given QSOs and QTCs, update the QSOs fields for points, multis and duplicate.
func (c *scoreCounter) Fill(qsos []core.QSO, qtcs []core.QTC) []core.QSO {
	c.clear()
	for i, qso := range qsos {
		qsoScore := c.addQSO(qso)
		qso.Points = qsoScore.Points
		qso.Multis = qsoScore.Multis
		qso.Duplicate = qsoScore.Duplicate
		qsos[i] = qso
	}

	// TODO: fill score counter with qtcs

	return qsos
}

// Add the given QSO and return the QSO with updated fields for points, multis, and duplicate.
func (c *scoreCounter) AddQSO(qso core.QSO) core.QSO {
	qsoScore := c.addQSO(qso)
	qso.Points = qsoScore.Points
	qso.Multis = qsoScore.Multis
	qso.Duplicate = qsoScore.Duplicate

	return qso
}

func (c *scoreCounter) addQSO(qso core.QSO) core.QSOScore {
	qsoScore := c.counter.Add(c.toConvalQSO(qso))
	result := core.QSOScore{
		Points:    qsoScore.Points,
		Multis:    qsoScore.Multis,
		Duplicate: qsoScore.Duplicate,
	}

	bandScore := c.score.ScorePerBand[qso.Band]
	bandScore.AddQSO(result)
	c.score.ScorePerBand[qso.Band] = bandScore

	if c.contestDefinition != nil {
		graph, ok := c.score.GraphPerBand[qso.Band]
		if !ok {
			graph = core.NewBandGraph(qso.Band, c.contestStartTime, c.contestDefinition.Duration)
		}
		graph.Add(qso.Time, result)
		c.score.GraphPerBand[graph.Band] = graph

		sumGraph, ok := c.score.GraphPerBand[core.NoBand]
		if !ok {
			sumGraph = core.NewBandGraph(core.NoBand, c.contestStartTime, c.contestDefinition.Duration)
		}
		sumGraph.Add(qso.Time, result)
		c.score.GraphPerBand[core.NoBand] = sumGraph
	}

	return result
}

func (c *scoreCounter) Score() core.Score {
	return c.score
}

func (c *scoreCounter) toConvalQSO(qso core.QSO) conval.QSO {
	result := conval.QSO{
		TheirCall:     qso.Callsign,
		Timestamp:     qso.Time,
		Band:          conval.ContestBand(qso.Band),
		Mode:          toConvalMode[qso.Mode],
		MyExchange:    c.toQSOExchange(c.myExchangeFields, qso.MyExchange),
		TheirExchange: c.toQSOExchange(c.theirExchangeFields, qso.TheirExchange),
	}
	continent, country, _, _, ok := c.prefixDatabase.Find(qso.Callsign.String())
	if ok {
		result.TheirContinent = continent
		result.TheirCountry = country
	}
	return result
}

func (c *scoreCounter) toQSOExchange(fields []conval.ExchangeField, values []string) conval.QSOExchange {
	return conval.ParseExchange(fields, values, c.prefixDatabase, c.contestDefinition)
}

func toConvalExchangeFields(fields []core.ExchangeField) []conval.ExchangeField {
	result := make([]conval.ExchangeField, len(fields))
	for i, field := range fields {
		result[i] = field.Properties
	}
	return result
}

type prefixDatabase struct {
	prefixes DXCCEntities
}

func (d prefixDatabase) Find(s string) (conval.Continent, conval.DXCCEntity, conval.CQZone, conval.ITUZone, bool) {
	entity, found := d.prefixes.Find(s)
	if !found {
		return "", "", 0, 0, false
	}

	continent, country, cqZone, ituZone := toConvalDXCCEntity(entity)
	return continent, country, cqZone, ituZone, true
}

func toConvalDXCCEntity(entity dxcc.Prefix) (conval.Continent, conval.DXCCEntity, conval.CQZone, conval.ITUZone) {
	return conval.Continent(strings.ToLower(entity.Continent)),
		conval.DXCCEntity(strings.ToLower(entity.PrimaryPrefix)),
		conval.CQZone(entity.CQZone),
		conval.ITUZone(entity.ITUZone)
}
