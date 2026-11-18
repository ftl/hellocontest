package logbook

import (
	"log"
	"strings"
	"time"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core/dxcc"
	"github.com/ftl/hellocontest/core/model"
)

type scoreCounter struct {
	score          model.Score
	counter        convalCounter
	timeSheet      convalTimeSheet
	activeBands    map[model.Band]bool
	activeModes    map[model.Mode]bool
	prefixDatabase prefixDatabase
	valid          bool

	multiProperties []conval.Property

	contestSetup        conval.Setup
	contestDefinition   *conval.Definition
	contestStartTime    time.Time
	myExchangeFields    []conval.ExchangeField
	theirExchangeFields []conval.ExchangeField

	counterFactory   func() convalCounter
	timeSheetFactory func() convalTimeSheet
}

func newScoreCounter(settings model.Settings, entities DXCCEntities) *scoreCounter {
	result := &scoreCounter{
		score:          model.NewScore(),
		counter:        new(nullCounter),
		timeSheet:      new(nullTimeSheet),
		activeBands:    make(map[model.Band]bool),
		activeModes:    make(map[model.Mode]bool),
		prefixDatabase: prefixDatabase{entities},
	}
	result.counterFactory = result.newConvalCounter
	result.timeSheetFactory = result.newConvalTimeSheet

	result.StationChanged(settings.Station())
	result.ContestChanged(settings.Contest())

	return result
}

func (c *scoreCounter) StationChanged(station model.Station) {
	oldSetup := c.contestSetup

	entity, found := c.prefixDatabase.prefixes.Find(station.Callsign.String())
	if !found {
		log.Printf("No DXCC entity found for the station callsign %s", station.Callsign)
		return
	}

	c.contestSetup = model.ContestSetup(station, entity)
	log.Printf("Using %+v as station setup", c.contestSetup)

	// the rules decide per station which multiplier properties count
	c.updateMultiProperties()

	c.valid = (oldSetup.MyCountry == c.contestSetup.MyCountry)
	if !c.valid {
		c.clear()
	}
}

func (c *scoreCounter) updateMultiProperties() {
	if c.contestDefinition == nil {
		c.multiProperties = nil
	} else {
		c.multiProperties = c.contestDefinition.MultiProperties(c.contestSetup)
	}
	c.score.MultiProperties = c.multiProperties
}

func (c *scoreCounter) ContestChanged(contest model.Contest) {
	c.contestDefinition = contest.Definition
	c.contestStartTime = contest.StartTime
	c.myExchangeFields = toConvalExchangeFields(contest.MyExchangeFields)
	c.theirExchangeFields = toConvalExchangeFields(contest.TheirExchangeFields)
	c.updateMultiProperties()

	c.valid = false
	c.clear()
}

func (c *scoreCounter) Valid() bool {
	return c.valid && (c.contestSetup.MyCountry != "") && (c.contestSetup.MyContinent != "")
}

func (c *scoreCounter) Clear() {
	c.clear()
	c.valid = (c.contestSetup.MyCountry != "") // RESET the valid flag
}

func (c *scoreCounter) clear() {
	c.score = model.NewScore()
	c.score.MultiProperties = c.multiProperties
	c.counter = c.counterFactory()
	c.timeSheet = c.timeSheetFactory()
	c.activeBands = make(map[model.Band]bool)
	c.activeModes = make(map[model.Mode]bool)
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

// Add the given QSO and return the QSO with updated fields for points, multis, and duplicate.
func (c *scoreCounter) AddQSO(qso model.QSO) model.QSO {
	convalScore := c.counter.Add(c.toConvalQSO(qso))
	qsoScore := model.QSOScore{
		Points:            convalScore.Points,
		Multis:            convalScore.Multis,
		MultisPerProperty: convalScore.MultisPerProperty,
		Duplicate:         convalScore.Duplicate,
	}
	qso.Points = qsoScore.Points
	qso.Multis = qsoScore.Multis
	qso.Duplicate = qsoScore.Duplicate

	c.timeSheet.MarkActive(qso.Time)
	c.activeBands[qso.Band] = true
	c.activeModes[qso.Mode] = true

	bandScore := c.score.ScorePerBand[qso.Band]
	bandScore.AddQSO(qsoScore)
	c.score.ScorePerBand[qso.Band] = bandScore

	if c.contestDefinition != nil {
		graph, ok := c.score.GraphPerBand[qso.Band]
		if !ok {
			graph = model.NewBandGraph(qso.Band, c.contestStartTime, c.contestDefinition.Duration)
		}
		graph.Add(qso.Time, qsoScore)
		c.score.GraphPerBand[graph.Band] = graph

		sumGraph, ok := c.score.GraphPerBand[model.NoBand]
		if !ok {
			sumGraph = model.NewBandGraph(model.NoBand, c.contestStartTime, c.contestDefinition.Duration)
		}
		sumGraph.Add(qso.Time, qsoScore)
		c.score.GraphPerBand[model.NoBand] = sumGraph
	}

	return qso
}

func (c *scoreCounter) AddQTCSeries(series model.QTCSeries) {
	if len(series.QTCs) == 0 {
		return
	}
	firstQTC := series.QTCs[0]

	convalQTC := c.toConvalQTC(series)
	convalScore := c.counter.AddQTC(convalQTC)
	qtcScore := model.QTCScore{
		Value: convalScore.Value,
	}

	c.timeSheet.MarkActive(firstQTC.Timestamp)
	c.activeBands[firstQTC.Band] = true
	c.activeModes[firstQTC.Mode] = true

	bandScore := c.score.ScorePerBand[firstQTC.Band]
	bandScore.AddQTC(qtcScore)
	c.score.ScorePerBand[firstQTC.Band] = bandScore

	if c.contestDefinition != nil {
		graph, ok := c.score.GraphPerBand[firstQTC.Band]
		if !ok {
			graph = model.NewBandGraph(firstQTC.Band, c.contestStartTime, c.contestDefinition.Duration)
		}
		graph.AddQTC(firstQTC.Timestamp, qtcScore)
		c.score.GraphPerBand[graph.Band] = graph

		sumGraph, ok := c.score.GraphPerBand[model.NoBand]
		if !ok {
			sumGraph = model.NewBandGraph(model.NoBand, c.contestStartTime, c.contestDefinition.Duration)
		}
		sumGraph.AddQTC(firstQTC.Timestamp, qtcScore)
		c.score.GraphPerBand[model.NoBand] = sumGraph
	}
}

func (c *scoreCounter) Value(callsign model.Callsign, entity dxcc.Prefix, band model.Band, mode model.Mode, exchange []string) (points, multis int, multiValues map[conval.Property]string) {
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

func (c *scoreCounter) Score() model.Score {
	return c.score
}

func (c *scoreCounter) FillSummary(summary *model.Summary) {
	breakDuration := c.computeMinBreakDuration(c.contestDefinition, summary.OperatorMode, summary.Overlay)
	timeReport := c.timeSheet.TimeReport(breakDuration)
	bands := make([]string, 0, len(c.activeBands))
	for _, band := range model.Bands {
		if c.activeBands[band] {
			bands = append(bands, string(band))
		}
	}
	modes := make([]string, 0, len(c.activeModes))
	for _, mode := range model.Modes {
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

func (c *scoreCounter) toConvalQSO(qso model.QSO) conval.QSO {
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

func toConvalExchangeFields(fields []model.ExchangeField) []conval.ExchangeField {
	result := make([]conval.ExchangeField, len(fields))
	for i, field := range fields {
		result[i] = field.Properties
	}
	return result
}

func (c *scoreCounter) toConvalQTC(series model.QTCSeries) conval.QTC {
	if len(series.QTCs) == 0 {
		return conval.QTC{}
	}
	band := series.QTCs[0].Band
	mode := series.QTCs[0].Mode
	result := conval.QTC{
		Kind:      "",
		Header:    series.Header.String(),
		TheirCall: series.TheirCallsign,
		Band:      conval.ContestBand(band),
		Mode:      toConvalMode[mode],
		Count:     len(series.QTCs),
	}
	continent, country, _, _, ok := c.prefixDatabase.Find(series.TheirCallsign.String())
	if ok {
		result.TheirContinent = continent
		result.TheirCountry = country
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
