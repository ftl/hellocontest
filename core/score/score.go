package score

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode"

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

func NewCounter(settings core.Settings, entities DXCCEntities) *Counter {
	result := &Counter{
		Score: core.Score{
			ScorePerBand: make(map[core.Band]core.BandScore),
		},
		counter:        new(nullCounter),
		view:           new(nullView),
		entities:       entities,
		prefixDatabase: prefixDatabase{entities},

		specificCountryPrefixes: make(map[string]bool),
		multisPerBand:           make(map[core.Band]*multis),
	}

	result.setStation(settings.Station())
	result.setContest(settings.Contest())
	result.resetCounter()

	return result
}

type Counter struct {
	core.Score
	counter        convalCounter
	view           View
	entities       DXCCEntities
	prefixDatabase prefixDatabase
	invalid        bool

	contestSetup        conval.Setup
	contestDefinition   *conval.Definition
	myExchangeFields    []conval.ExchangeField
	theirExchangeFields []conval.ExchangeField

	stationEntity           dxcc.Prefix
	countPerBand            bool
	sameCountryPoints       int
	sameContinentPoints     int
	specificCountryPoints   int
	specificCountryPrefixes map[string]bool
	otherPoints             int
	multis                  core.Multis
	xchangeMultiExpression  *regexp.Regexp

	listeners []interface{}

	multisPerBand map[core.Band]*multis
	overallMultis *multis
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
}

func (c *Counter) StationChanged(station core.Station) {
	oldEntity := c.stationEntity
	c.setStation(station)
	c.invalid = (oldEntity != c.stationEntity)

	c.resetCounter()
}

func (c *Counter) setStation(station core.Station) {
	entity, found := c.entities.Find(station.Callsign.String())
	if !found {
		log.Printf("No DXCC entity found for the station callsign %s", station.Callsign)
		c.stationEntity = dxcc.Prefix{}
		return
	}
	c.stationEntity = entity
	log.Printf("Using %v as station entity", c.stationEntity)

	// new stuff from here
	continent, country := toConvalDXCCEntity(c.stationEntity)
	c.contestSetup = conval.Setup{
		MyCall:      station.Callsign,
		MyContinent: continent,
		MyCountry:   country,
		GridLocator: station.Locator,
	}
}

func (c *Counter) ContestChanged(contest core.Contest) {
	c.setContest(contest)
	c.invalid = true

	c.resetCounter()
}

func (c *Counter) setContest(contest core.Contest) {
	c.sameCountryPoints = contest.SameCountryPoints
	c.sameContinentPoints = contest.SameContinentPoints
	c.specificCountryPoints = contest.SpecificCountryPoints
	c.otherPoints = contest.OtherPoints
	c.multis = contest.Multis
	c.countPerBand = contest.CountPerBand

	for _, prefix := range contest.SpecificCountryPrefixes {
		c.specificCountryPrefixes[strings.ToUpper(prefix)] = true
	}

	exp, err := regexp.Compile(contest.XchangeMultiPattern)
	if err != nil {
		log.Printf("Invalid regular expression for Xchange Multis: %v", err)
	} else {
		c.xchangeMultiExpression = exp
		log.Printf("Using pattern %q for Xchange Multis", c.xchangeMultiExpression)
	}
	c.overallMultis = newMultis(contest.Multis, c.xchangeMultiExpression)

	// new stuff from here
	c.contestDefinition = contest.Definition
	c.myExchangeFields = toConvalExchangeFields(contest.MyExchangeFields)
	c.theirExchangeFields = toConvalExchangeFields(contest.TheirExchangeFields)
}

func (c *Counter) resetCounter() {
	if c.contestDefinition == nil {
		c.counter = new(nullCounter)
		return
	}

	c.counter = conval.NewCounter(*c.contestDefinition, c.contestSetup)
}

func (c *Counter) Valid() bool {
	return !c.invalid && (c.stationEntity.PrimaryPrefix != "") && (c.stationEntity.Continent != "")
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
	c.Score = core.Score{
		ScorePerBand: make(map[core.Band]core.BandScore),
	}

	c.resetCounter()

	c.invalid = c.stationEntity.Name == ""
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Add(qso core.QSO) {
	bandScore := c.ScorePerBand[qso.Band]

	if qso.Duplicate {
		bandScore.Duplicates++
		c.ScorePerBand[qso.Band] = bandScore
		return
	}

	qsoScore := c.counter.Add(c.toConvalQSO(qso))
	bandScore.Add(core.BandScore{
		QSOs:   1,
		Points: qsoScore.Points,
		Multis: qsoScore.Multis,
	})
	c.ScorePerBand[qso.Band] = bandScore

	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Update(oldQSO, newQSO core.QSO) {
	// TODO: implement new stuff from here - update not supported by conval, need to clear and replay all

	if (oldQSO.DXCC == newQSO.DXCC) && (fmt.Sprintf("%v", oldQSO.TheirExchange) == fmt.Sprintf("%v", newQSO.TheirExchange)) && (oldQSO.Duplicate == newQSO.Duplicate) {
		return
	}
	oldBandScore := c.ScorePerBand[oldQSO.Band]
	var newBandScore *core.BandScore
	if oldQSO.Band == newQSO.Band {
		newBandScore = &oldBandScore
	} else {
		s := c.ScorePerBand[newQSO.Band]
		newBandScore = &s
	}

	if oldQSO.Duplicate {
		oldBandScore.Duplicates--
	}

	if newQSO.Duplicate {
		newBandScore.Duplicates++
	}

	if !oldQSO.Duplicate {
		oldQSOScore := c.qsoScore(-1, oldQSO.DXCC)
		oldBandScore.Add(oldQSOScore)
	}

	if !newQSO.Duplicate {
		newQSOScore := c.qsoScore(1, newQSO.DXCC)
		newBandScore.Add(newQSOScore)
	}

	if !oldQSO.Duplicate {
		oldMultisPerBand, ok := c.multisPerBand[oldQSO.Band]
		if ok {
			oldBandMultiScore := oldMultisPerBand.Add(-1, oldQSO.Callsign, oldQSO.DXCC, "") // oldQSO.TheirXchange) // TODO use the new exhange fields
			oldBandScore.Add(oldBandMultiScore)
		}
	}

	if !newQSO.Duplicate {
		newMultisPerBand, ok := c.multisPerBand[newQSO.Band]
		if !ok {
			newMultisPerBand = newMultis(c.multis, c.xchangeMultiExpression)
			c.multisPerBand[newQSO.Band] = newMultisPerBand
		}
		newBandMultiScore := newMultisPerBand.Add(1, newQSO.Callsign, newQSO.DXCC, "") // newQSO.TheirXchange) // TODO use the new exhange fields
		newBandScore.Add(newBandMultiScore)
	}

	c.ScorePerBand[oldQSO.Band] = oldBandScore
	c.ScorePerBand[newQSO.Band] = *newBandScore

	c.emitScoreUpdated(c.Score)
}

func (c *Counter) emitScoreUpdated(score core.Score) {
	c.view.ShowScore(score)
	for _, listener := range c.listeners {
		if scoreUpdatedListener, ok := listener.(ScoreUpdatedListener); ok {
			scoreUpdatedListener.ScoreUpdated(score)
		}
	}
}

func (c *Counter) qsoScore(value int, entity dxcc.Prefix) core.BandScore {
	var result core.BandScore
	switch {
	case c.isSpecificCountry(entity):
		result.QSOs += value
		result.Points += value * c.specificCountryPoints
	case entity.PrimaryPrefix == c.stationEntity.PrimaryPrefix:
		result.QSOs += value
		result.Points += value * c.sameCountryPoints
	case entity.Continent == c.stationEntity.Continent:
		result.QSOs += value
		result.Points += value * c.sameContinentPoints
	default:
		result.QSOs += value
		result.Points += value * c.otherPoints
	}

	return result
}

func (c *Counter) isSpecificCountry(entity dxcc.Prefix) bool {
	return c.specificCountryPrefixes[entity.PrimaryPrefix]
}

func (c *Counter) Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, mode core.Mode, exchange []string) (points, multis int) {
	continent, country := toConvalDXCCEntity(entity)
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
	continent, country, ok := c.prefixDatabase.Find(qso.Callsign.String())
	if ok {
		result.TheirContinent = continent
		result.TheirCountry = country
	}
	return result
}

func (c *Counter) toQSOExchange(fields []conval.ExchangeField, values []string) conval.QSOExchange {
	return conval.ParseExchange(fields, values, c.prefixDatabase)
}

func toConvalExchangeFields(fields []core.ExchangeField) []conval.ExchangeField {
	result := make([]conval.ExchangeField, len(fields))
	for i, field := range fields {
		result[i] = field.Properties
	}
	return result
}

func newMultis(countingMultis core.Multis, xchangeMultiExpression *regexp.Regexp) *multis {
	return &multis{
		CountingMultis:         countingMultis,
		XchangeMultiExpression: xchangeMultiExpression,
		CQZones:                make(map[dxcc.CQZone]int),
		ITUZones:               make(map[dxcc.ITUZone]int),
		DXCCEntities:           make(map[string]int),
		WPXPrefixes:            make(map[string]int),
		XchangeValues:          make(map[string]int),
	}
}

type multis struct {
	CountingMultis         core.Multis
	XchangeMultiExpression *regexp.Regexp
	CQZones                map[dxcc.CQZone]int
	ITUZones               map[dxcc.ITUZone]int
	DXCCEntities           map[string]int
	WPXPrefixes            map[string]int
	XchangeValues          map[string]int
}

func (m *multis) Value(callsign callsign.Callsign, entity dxcc.Prefix, xchange string) int {
	var dxccEntitiesValue int
	if m.DXCCEntities[entity.PrimaryPrefix] == 0 {
		dxccEntitiesValue = 1
	}
	var wpxPrefixesValue int
	wpxPrefix := WPXPrefix(callsign)
	if (m.WPXPrefixes[wpxPrefix] == 0) && (wpxPrefix != "") {
		wpxPrefixesValue = 1
	}
	var xchangeValue int
	xchangeMulti, xchangeMatch := m.matchXchange(xchange)
	if xchangeMatch && m.XchangeValues[xchangeMulti] == 0 {
		xchangeValue = 1
	}

	var result int
	if m.CountingMultis.DXCC {
		result += dxccEntitiesValue
	}
	if m.CountingMultis.WPX {
		result += wpxPrefixesValue
	}
	if m.CountingMultis.Xchange {
		result += xchangeValue
	}
	return result
}

func (m *multis) Add(value int, callsign callsign.Callsign, entity dxcc.Prefix, xchange string) core.BandScore {
	var result core.BandScore

	oldCQZoneCount := m.CQZones[entity.CQZone]
	newCQZoneCount := oldCQZoneCount + value
	m.CQZones[entity.CQZone] = newCQZoneCount
	if oldCQZoneCount == 0 || newCQZoneCount == 0 {
	}

	oldITUZoneCount := m.ITUZones[entity.ITUZone]
	newITUZoneCount := oldITUZoneCount + value
	m.ITUZones[entity.ITUZone] = newITUZoneCount
	if oldITUZoneCount == 0 || newITUZoneCount == 0 {
	}

	oldDXCCEntitiesCount := m.DXCCEntities[entity.PrimaryPrefix]
	newDXCCEntitiesCount := oldDXCCEntitiesCount + value
	m.DXCCEntities[entity.PrimaryPrefix] = newDXCCEntitiesCount
	if oldDXCCEntitiesCount == 0 || newDXCCEntitiesCount == 0 {
	}

	wpxPrefix := WPXPrefix(callsign)
	if wpxPrefix != "" {
		oldWPXPrefixesCount := m.WPXPrefixes[wpxPrefix]
		newWPXPrefixesCount := oldWPXPrefixesCount + value
		m.WPXPrefixes[wpxPrefix] = newWPXPrefixesCount
		if oldWPXPrefixesCount == 0 || newWPXPrefixesCount == 0 {
		}
	}

	xchangeMulti, xchangeMatch := m.matchXchange(xchange)
	if xchangeMatch {
		oldXchangeValuesCount := m.XchangeValues[xchangeMulti]
		newXchangeValuesCount := oldXchangeValuesCount + value
		m.XchangeValues[xchangeMulti] = newXchangeValuesCount
		if oldXchangeValuesCount == 0 || newXchangeValuesCount == 0 {
		}
	}

	if m.CountingMultis.DXCC {
	}
	if m.CountingMultis.WPX {
	}
	if m.CountingMultis.Xchange {
	}

	return result
}

func (m *multis) matchXchange(xchange string) (string, bool) {
	return MatchXchange(m.XchangeMultiExpression, xchange)
}

func MatchXchange(exp *regexp.Regexp, xchange string) (string, bool) {
	xchange = strings.ToUpper(strings.TrimSpace(xchange))
	if exp == nil {
		return xchange, true
	}

	matches := exp.FindStringSubmatch(xchange)
	if len(matches) == 0 {
		return "", false
	}

	multiIndex := exp.SubexpIndex("multi")
	var multi string
	if multiIndex == -1 {
		multi = matches[0]
	} else {
		multi = matches[multiIndex]
	}
	return multi, (multi != "")
}

var parseWPXPrefixExpression = regexp.MustCompile(`^[A-Z0-9]?[A-Z][0-9]*`)

func WPXPrefix(callsign callsign.Callsign) string {
	var p string
	if p == "" && callsign.Prefix != "" {
		p = parseWPXPrefixExpression.FindString(callsign.Prefix)
	}
	if p == "" && callsign.Suffix != "" {
		p = parseWPXPrefixExpression.FindString(callsign.Suffix)
	}
	if p == "" {
		p = parseWPXPrefixExpression.FindString(callsign.BaseCall)
	}
	if p == "" {
		return ""
	}
	runes := []rune(p)
	if !unicode.IsDigit(runes[len(runes)-1]) {
		p += "0"
	}
	return p
}

type prefixDatabase struct {
	prefixes DXCCEntities
}

func (d prefixDatabase) Find(s string) (conval.Continent, conval.DXCCEntity, bool) {
	entity, found := d.prefixes.Find(s)
	if !found {
		return "", "", false
	}

	continent, country := toConvalDXCCEntity(entity)
	return continent, country, true
}

func toConvalDXCCEntity(entity dxcc.Prefix) (conval.Continent, conval.DXCCEntity) {
	return conval.Continent(strings.ToLower(entity.Continent)), conval.DXCCEntity(strings.ToLower(entity.PrimaryPrefix))
}

type nullCounter struct{}

func (c *nullCounter) Add(conval.QSO) conval.QSOScore   { return conval.QSOScore{} }
func (c *nullCounter) Probe(conval.QSO) conval.QSOScore { return conval.QSOScore{} }

type nullView struct{}

func (v *nullView) Show()                      {}
func (v *nullView) Hide()                      {}
func (v *nullView) ShowScore(score core.Score) {}
