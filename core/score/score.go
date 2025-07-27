package score

import (
	"log"
	"strings"
	"sync"
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
	ComputeMinBreakDuration() time.Duration
}

type convalTimeSheet interface {
	MarkActive(now time.Time)
	TimeReport(minBreakDuration time.Duration) conval.TimeReport
}

var toConvalMode = map[core.Mode]conval.Mode{
	core.ModeCW:      conval.ModeCW,
	core.ModeSSB:     conval.ModeSSB,
	core.ModeFM:      conval.ModeFM,
	core.ModeRTTY:    conval.ModeRTTY,
	core.ModeDigital: conval.ModeDigital,
}

var fromConvalMode = map[conval.Mode]core.Mode{
	conval.ModeCW:      core.ModeCW,
	conval.ModeSSB:     core.ModeSSB,
	conval.ModeFM:      core.ModeFM,
	conval.ModeRTTY:    core.ModeRTTY,
	conval.ModeDigital: core.ModeDigital,
}

// Counter is thread-safe.
type Counter struct {
	score          core.Score
	readScore      core.Score
	scoreLock      *sync.Mutex
	counter        *safeConvalCounter
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

	listeners []any
}

func NewCounter(settings core.Settings, entities DXCCEntities) *Counter {
	result := &Counter{
		score:          core.NewScore(),
		scoreLock:      &sync.Mutex{},
		counter:        newSafeCounter(new(nullCounter), new(nullTimeSheet)),
		view:           new(nullView),
		prefixDatabase: prefixDatabase{entities},
	}

	result.copyScore()
	result.setStation(settings.Station())
	result.setContest(settings.Contest())
	result.resetCounter() // CONVAL WRITE LOCK

	return result
}

func (c *Counter) copyScore() {
	c.readScore = c.score.Copy() // READ
}

func (c *Counter) FillSummary(summary *core.Summary) {
	c.scoreLock.Lock()
	summary.Score = c.readScore
	c.scoreLock.Unlock()

	summary.TimeReport,
		summary.WorkedBands,
		summary.WorkedModes =
		c.counter.SummaryContent(
			c.contestDefinition,
			summary.OperatorMode,
			summary.Overlay,
		) // CONVAL READ LOCK
}

func (c *Counter) Result() int {
	return c.readScore.Result().Result() // READ
}

func (c *Counter) SetView(view View) {
	if view == nil {
		panic("score.Counter.SetView must not be called with nil")
	}
	if _, ok := c.view.(*nullView); !ok {
		panic("score.Counter.SetView was already called")
	}

	c.view = view
	c.view.SetGoals(c.contestPointsGoal, c.contestMultisGoal)
	c.view.ShowScore(c.readScore) // READ
}

func (c *Counter) StationChanged(station core.Station) {
	oldSetup := c.contestSetup
	c.setStation(station)
	c.invalid = (oldSetup.MyCountry != c.contestSetup.MyCountry)

	c.resetCounter() // CONVAL WRITE LOCK
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

	c.resetCounter() // CONVAL WRITE LOCK
}

func (c *Counter) setContest(contest core.Contest) {
	c.contestDefinition = contest.Definition
	c.contestStartTime = contest.StartTime
	c.contestPointsGoal = contest.PointsGoal
	c.contestMultisGoal = contest.MultisGoal
	c.myExchangeFields = toConvalExchangeFields(contest.MyExchangeFields)
	c.theirExchangeFields = toConvalExchangeFields(contest.TheirExchangeFields)
}

func (c *Counter) Valid() bool {
	return !c.invalid && (c.contestSetup.MyCountry != "") && (c.contestSetup.MyContinent != "")
}

func (c *Counter) Show() {
	c.view.Show()
	c.view.SetGoals(c.contestPointsGoal, c.contestMultisGoal)
	c.view.ShowScore(c.readScore) // READ
}

func (c *Counter) Hide() {
	c.view.Hide()
}

func (c *Counter) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Counter) Clear() {
	c.scoreLock.Lock()
	c.score = core.NewScore() // WRITE
	c.copyScore()
	c.scoreLock.Unlock()
	c.resetCounter() // CONVAL WRITE LOCK

	c.invalid = (c.contestSetup.MyCountry == "")
}

func (c *Counter) AddMuted(qso core.QSO) core.QSOScore {
	qsoScore := c.counter.Add(c.toConvalQSO(qso)) // CONVAL WRITE LOCK
	result := core.QSOScore{
		Points:    qsoScore.Points,
		Multis:    qsoScore.Multis,
		Duplicate: qsoScore.Duplicate,
	}

	c.scoreLock.Lock()

	bandScore := c.score.ScorePerBand[qso.Band] // PREPARE WRITE
	bandScore.AddQSO(result)
	c.score.ScorePerBand[qso.Band] = bandScore // WRITE

	if c.contestDefinition != nil {
		graph, ok := c.score.GraphPerBand[qso.Band] // PREPARE WRITE
		if !ok {
			graph = core.NewBandGraph(qso.Band, c.contestStartTime, c.contestDefinition.Duration)
		}
		graph.Add(qso.Time, result)
		c.score.GraphPerBand[graph.Band] = graph // WRITE

		sumGraph, ok := c.score.GraphPerBand[core.NoBand] // PREPARE WRITE
		if !ok {
			sumGraph = core.NewBandGraph(core.NoBand, c.contestStartTime, c.contestDefinition.Duration)
		}
		sumGraph.Add(qso.Time, result)
		c.score.GraphPerBand[core.NoBand] = sumGraph // WRITE
	}

	c.copyScore()
	c.scoreLock.Unlock()

	return result
}

func (c *Counter) Unmute() {
	c.emitScoreUpdated(c.readScore)
}

func (c *Counter) emitScoreUpdated(score core.Score) {
	c.view.ShowScore(score)
	for _, listener := range c.listeners {
		if scoreUpdatedListener, ok := listener.(ScoreUpdatedListener); ok {
			scoreUpdatedListener.ScoreUpdated(score)
		}
	}
}

func (c *Counter) Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, mode core.Mode, exchange []string) (points, multis int, multiValues map[conval.Property]string) {
	continent, country, _, _ := toConvalDXCCEntity(entity)
	convalQSO := conval.QSO{
		TheirCall:      callsign,
		TheirContinent: continent,
		TheirCountry:   country,
		Band:           conval.ContestBand(band),
		Mode:           toConvalMode[mode],
		TheirExchange:  c.toQSOExchange(c.theirExchangeFields, exchange),
	}
	qsoScore := c.counter.Probe(convalQSO) // CONVAL READ LOCK

	return qsoScore.Points, qsoScore.Multis, qsoScore.MultiValues
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

func (c *Counter) resetCounter() {
	var counter convalCounter
	var timeSheet convalTimeSheet
	if c.contestDefinition == nil {
		counter = new(nullCounter)
		timeSheet = new(nullTimeSheet)
	} else {
		counter = conval.NewCounter(*c.contestDefinition, c.contestSetup, c.prefixDatabase)
		timeSheet = conval.NewTimeSheet(c.contestStartTime, c.contestDefinition.Duration)
	}
	c.counter.Reset(counter, timeSheet) // CONVAL WRITE LOCK
}

type safeConvalCounter struct {
	counter   convalCounter
	timeSheet convalTimeSheet

	bands map[conval.ContestBand]bool
	modes map[conval.Mode]bool

	lock *sync.RWMutex
}

func newSafeCounter(counter convalCounter, timeSheet convalTimeSheet) *safeConvalCounter {
	return &safeConvalCounter{
		counter:   counter,
		timeSheet: timeSheet,
		bands:     make(map[conval.ContestBand]bool),
		modes:     make(map[conval.Mode]bool),
		lock:      &sync.RWMutex{},
	}
}

func (c *safeConvalCounter) Reset(counter convalCounter, timeSheet convalTimeSheet) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if counter == nil {
		c.counter = new(nullCounter)
	} else {
		c.counter = counter
	}
	if timeSheet == nil {
		c.timeSheet = new(nullTimeSheet)
	} else {
		c.timeSheet = timeSheet
	}

	c.bands = make(map[conval.ContestBand]bool)
	c.modes = make(map[conval.Mode]bool)
}

func (c *safeConvalCounter) Add(qso conval.QSO) conval.QSOScore {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.timeSheet.MarkActive(qso.Timestamp)
	c.bands[qso.Band] = true
	c.modes[qso.Mode] = true

	return c.counter.Add(qso)
}

func (c *safeConvalCounter) Probe(qso conval.QSO) conval.QSOScore {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.counter.Probe(qso)
}

func (c *safeConvalCounter) SummaryContent(definition *conval.Definition, operatorMode conval.OperatorMode, overlay conval.Overlay) (core.TimeReport, []core.Band, []core.Mode) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	timeReport := c.timeSheet.TimeReport(c.computeMinBreakDuration(definition, operatorMode, overlay))
	bands := make([]core.Band, 0, len(c.bands))
	for band := range c.bands {
		bands = append(bands, core.Band(band))
	}
	modes := make([]core.Mode, 0, len(c.modes))
	for mode := range c.modes {
		modes = append(modes, fromConvalMode[mode])
	}

	return timeReport, bands, modes
}

func (c *safeConvalCounter) computeMinBreakDuration(definition *conval.Definition, operatorMode conval.OperatorMode, overlay conval.Overlay) time.Duration {
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

func (c *nullCounter) Add(conval.QSO) conval.QSOScore         { return conval.QSOScore{} }
func (c *nullCounter) Probe(conval.QSO) conval.QSOScore       { return conval.QSOScore{} }
func (c *nullCounter) ComputeMinBreakDuration() time.Duration { return 0 }

type nullTimeSheet struct{}

func (t *nullTimeSheet) MarkActive(time.Time) {}
func (t *nullTimeSheet) TimeReport(minBreakDuration time.Duration) conval.TimeReport {
	return conval.TimeReport{}
}

type nullView struct{}

func (v *nullView) Show()                {}
func (v *nullView) Hide()                {}
func (v *nullView) ShowScore(core.Score) {}
func (v *nullView) SetGoals(int, int)    {}
