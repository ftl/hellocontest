package score

import (
	"log"
	"strings"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

type ScoreUpdatedListener interface {
	ScoreUpdated(core.Score)
}

type ScoreUpdatedListenerFunc func(core.Score)

func (f ScoreUpdatedListenerFunc) ScoreUpdated(score core.Score) {
	f(score)
}

type View interface {
	Show()
	Hide()

	ShowScore(score core.Score)
	SetGoals(points int, multis int)
}

// DXCCEntities returns a list of matching DXCC entities for the given string and indicates if there was a match at all.
type DXCCEntities interface {
	Find(string) (dxcc.Prefix, bool)
}

type convalCounter interface {
	Add(conval.QSO) conval.QSOScore
	Probe(conval.QSO) conval.QSOScore
}

var toConvalMode = map[core.Mode]conval.Mode{
	core.ModeCW:      conval.ModeCW,
	core.ModeSSB:     conval.ModeSSB,
	core.ModeFM:      conval.ModeFM,
	core.ModeRTTY:    conval.ModeRTTY,
	core.ModeDigital: conval.ModeDigital,
}

type Counter struct {
	core.Score
	counter        convalCounter
	view           View
	prefixDatabase prefixDatabase
	invalid        bool

	contestSetup      conval.Setup
	contestDefinition *conval.Definition
	contestStartTime  time.Time
	contestPointsGoal int
	contestMultisGoal int

	myExchangeFields    []conval.ExchangeField
	theirExchangeFields []conval.ExchangeField

	listeners []interface{}
}

func NewCounter(settings core.Settings, entities DXCCEntities) *Counter {
	result := &Counter{
		Score:          core.NewScore(),
		counter:        new(nullCounter),
		view:           new(nullView),
		prefixDatabase: prefixDatabase{entities},
	}

	result.setStation(settings.Station())
	result.setContest(settings.Contest())
	result.resetCounter()

	return result
}

func (c *Counter) Result() int {
	return c.Score.Result().Result()
}

func (c *Counter) SetView(view View) {
	if view == nil {
		c.view = new(nullView)
		return
	}
	c.view = view
	c.view.ShowScore(c.Score)
	c.view.SetGoals(c.contestPointsGoal, c.contestMultisGoal)
}

func (c *Counter) StationChanged(station core.Station) {
	oldSetup := c.contestSetup
	c.setStation(station)
	c.invalid = (oldSetup.MyCountry != c.contestSetup.MyCountry)

	c.resetCounter()
}

func (c *Counter) setStation(station core.Station) {
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
}

func (c *Counter) ContestChanged(contest core.Contest) {
	c.setContest(contest)
	c.view.SetGoals(c.contestPointsGoal, c.contestMultisGoal)
	c.invalid = true

	c.resetCounter()
}

func (c *Counter) setContest(contest core.Contest) {
	c.contestDefinition = contest.Definition
	c.contestStartTime = contest.StartTime
	c.contestPointsGoal = contest.PointsGoal
	c.contestMultisGoal = contest.MultisGoal
	c.myExchangeFields = toConvalExchangeFields(contest.MyExchangeFields)
	c.theirExchangeFields = toConvalExchangeFields(contest.TheirExchangeFields)
}

func (c *Counter) resetCounter() {
	if c.contestDefinition == nil {
		c.counter = new(nullCounter)
		return
	}

	c.counter = conval.NewCounter(*c.contestDefinition, c.contestSetup, c.prefixDatabase)
}

func (c *Counter) Valid() bool {
	return !c.invalid && (c.contestSetup.MyCountry != "") && (c.contestSetup.MyContinent != "")
}

func (c *Counter) Show() {
	c.view.Show()
	c.view.ShowScore(c.Score)
}

func (c *Counter) Hide() {
	c.view.Hide()
}

func (c *Counter) Notify(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}

func (c *Counter) Clear() {
	c.Score = core.NewScore()

	c.resetCounter()

	c.invalid = (c.contestSetup.MyCountry == "")
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Add(qso core.QSO) core.QSOScore {
	qsoScore := c.counter.Add(c.toConvalQSO(qso))
	result := core.QSOScore{
		Points:    qsoScore.Points,
		Multis:    qsoScore.Multis,
		Duplicate: qsoScore.Duplicate,
	}

	bandScore := c.ScorePerBand[qso.Band]
	bandScore.AddQSO(result)
	c.ScorePerBand[qso.Band] = bandScore

	if c.contestDefinition != nil {
		graph, ok := c.GraphPerBand[qso.Band]
		if !ok {
			graph = core.NewBandGraph(qso.Band, c.contestStartTime, c.contestDefinition.Duration)
		}
		graph.Add(qso.Time, result)
		c.GraphPerBand[graph.Band] = graph

		sumGraph, ok := c.GraphPerBand[core.NoBand]
		if !ok {
			sumGraph = core.NewBandGraph(core.NoBand, c.contestStartTime, c.contestDefinition.Duration)
		}
		sumGraph.Add(qso.Time, result)
		c.GraphPerBand[core.NoBand] = sumGraph
	}

	c.emitScoreUpdated(c.Score)

	return result
}

func (c *Counter) emitScoreUpdated(score core.Score) {
	c.view.ShowScore(score)
	for _, listener := range c.listeners {
		if scoreUpdatedListener, ok := listener.(ScoreUpdatedListener); ok {
			scoreUpdatedListener.ScoreUpdated(score)
		}
	}
}

func (c *Counter) Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, mode core.Mode, exchange []string) (points, multis int) {
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

	return qsoScore.Points, qsoScore.Multis
}

func (c *Counter) toConvalQSO(qso core.QSO) conval.QSO {
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

func (c *Counter) toQSOExchange(fields []conval.ExchangeField, values []string) conval.QSOExchange {
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

type nullCounter struct{}

func (c *nullCounter) Add(conval.QSO) conval.QSOScore   { return conval.QSOScore{} }
func (c *nullCounter) Probe(conval.QSO) conval.QSOScore { return conval.QSOScore{} }

type nullView struct{}

func (v *nullView) Show()                {}
func (v *nullView) Hide()                {}
func (v *nullView) ShowScore(core.Score) {}
func (v *nullView) SetGoals(int, int)    {}
