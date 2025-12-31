package logbook

import (
	"log"
	"strings"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
)

type scoreCounter struct {
	score          core.Score
	counter        convalCounter
	timeSheet      convalTimeSheet
	activeBands    map[core.Band]bool
	activeModes    map[core.Mode]bool
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
		counter:        new(nullCounter),
		timeSheet:      new(nullTimeSheet),
		activeBands:    make(map[core.Band]bool),
		activeModes:    make(map[core.Mode]bool),
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
	c.score = core.NewScore()
	c.counter = c.counterFactory()
	c.timeSheet = c.timeSheetFactory()
	c.activeBands = make(map[core.Band]bool)
	c.activeModes = make(map[core.Mode]bool)
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

	c.timeSheet.MarkActive(qso.Time)
	c.activeBands[qso.Band] = true
	c.activeModes[qso.Mode] = true

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

func (c *scoreCounter) Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, mode core.Mode, exchange []string) (points, multis int, multiValues map[conval.Property]string) {
	continent, country, _, _ := toConvalDXCCEntity(entity)
	convalQSO := conval.QSO{
		TheirCall:      callsign,
		TheirContinent: continent,
		TheirCountry:   country,
		Band:           conval.ContestBand(band),
		Mode:           toConvalMode[mode],
		TheirExchange:  c.toQSOExchange(c.theirExchangeFields, exchange),
	}
	qsoScore := c.counter.Probe(convalQSO)

	return qsoScore.Points, qsoScore.Multis, qsoScore.MultiValues
}

func (c *scoreCounter) Score() core.Score {
	return c.score
}

func (c *scoreCounter) FillSummary(summary *core.Summary) {
	breakDuration := c.computeMinBreakDuration(c.contestDefinition, summary.OperatorMode, summary.Overlay)
	timeReport := c.timeSheet.TimeReport(breakDuration)
	bands := make([]string, 0, len(c.activeBands))
	for _, band := range core.Bands {
		if c.activeBands[band] {
			bands = append(bands, string(band))
		}
	}
	modes := make([]string, 0, len(c.activeModes))
	for _, mode := range core.Modes {
		if c.activeModes[mode] {
			modes = append(modes, string(mode))
		}
	}

	summary.Score = c.score
	summary.TimeReport = timeReport
	summary.WorkedBands = bands
	summary.WorkedModes = modes
}

func (c *scoreCounter) computeMinBreakDuration(definition *conval.Definition, operatorMode conval.OperatorMode, overlay conval.Overlay) time.Duration {
	if len(definition.Breaks) == 1 {
		return definition.Breaks[0].Duration
	}

	for _, b := range definition.Breaks {
		if (b.Constraint.OperatorMode == operatorMode) &&
			(b.Constraint.Overlay == overlay) {
			return b.Duration
		}
	}

	return conval.DefaultBreakDuration
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
