package score

import (
	"log"
	"regexp"
	"strings"
	"unicode"

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

func NewCounter(settings core.Settings, entities DXCCEntities) *Counter {
	result := &Counter{
		Score: core.Score{
			ScorePerBand: make(map[core.Band]core.BandScore),
		},
		entities: entities,
		view:     new(nullView),

		specificCountryPrefixes: make(map[string]bool),
		multisPerBand:           make(map[core.Band]*multis),
	}

	result.setStation(settings.Station())
	result.setContest(settings.Contest())

	return result
}

type Counter struct {
	core.Score
	view     View
	entities DXCCEntities
	invalid  bool

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
	if c.countPerBand {
		return c.TotalScore.Result()
	} else {
		return c.OverallScore.Result()
	}
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
}

func (c *Counter) ContestChanged(contest core.Contest) {
	c.setContest(contest)
	c.invalid = true
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
	c.multisPerBand = make(map[core.Band]*multis)
	c.overallMultis = newMultis(c.multis, c.xchangeMultiExpression)
	c.invalid = c.stationEntity.Name == ""
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Add(qso core.QSO) {
	bandScore := c.ScorePerBand[qso.Band]

	if qso.Duplicate {
		bandScore.Duplicates++
		c.ScorePerBand[qso.Band] = bandScore
		c.OverallScore.Duplicates++
		c.TotalScore.Duplicates++
		return
	}

	qsoScore := c.qsoScore(1, qso.DXCC)
	c.TotalScore.Add(qsoScore)
	c.OverallScore.Add(qsoScore)
	bandScore.Add(qsoScore)

	overallMultiScore := c.overallMultis.Add(1, qso.Callsign, qso.DXCC, qso.TheirXchange)
	c.OverallScore.Add(overallMultiScore)
	multisPerBand, ok := c.multisPerBand[qso.Band]
	if !ok {
		multisPerBand = newMultis(c.multis, c.xchangeMultiExpression)
		c.multisPerBand[qso.Band] = multisPerBand
	}
	bandMultiScore := multisPerBand.Add(1, qso.Callsign, qso.DXCC, qso.TheirXchange)
	c.TotalScore.Add(bandMultiScore)
	bandScore.Add(bandMultiScore)

	c.ScorePerBand[qso.Band] = bandScore
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Update(oldQSO, newQSO core.QSO) {
	if (oldQSO.DXCC == newQSO.DXCC) && (oldQSO.TheirXchange == newQSO.TheirXchange) && (oldQSO.Duplicate == newQSO.Duplicate) {
		return
	}
	totalScore := c.TotalScore
	overallScore := c.OverallScore
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
		overallScore.Duplicates--
		totalScore.Duplicates--
	}

	if newQSO.Duplicate {
		newBandScore.Duplicates++
		overallScore.Duplicates++
		totalScore.Duplicates++
	}

	if !oldQSO.Duplicate {
		oldQSOScore := c.qsoScore(-1, oldQSO.DXCC)
		totalScore.Add(oldQSOScore)
		overallScore.Add(oldQSOScore)
		oldBandScore.Add(oldQSOScore)
	}

	if !newQSO.Duplicate {
		newQSOScore := c.qsoScore(1, newQSO.DXCC)
		totalScore.Add(newQSOScore)
		overallScore.Add(newQSOScore)
		newBandScore.Add(newQSOScore)
	}

	if !oldQSO.Duplicate {
		oldOverallMultiScore := c.overallMultis.Add(-1, oldQSO.Callsign, oldQSO.DXCC, oldQSO.TheirXchange)
		overallScore.Add(oldOverallMultiScore)
		oldMultisPerBand, ok := c.multisPerBand[oldQSO.Band]
		if ok {
			oldBandMultiScore := oldMultisPerBand.Add(-1, oldQSO.Callsign, oldQSO.DXCC, oldQSO.TheirXchange)
			oldBandScore.Add(oldBandMultiScore)
			totalScore.Add(oldBandMultiScore)
		}
	}

	if !newQSO.Duplicate {
		newOverallMultiScore := c.overallMultis.Add(1, newQSO.Callsign, newQSO.DXCC, newQSO.TheirXchange)
		overallScore.Add(newOverallMultiScore)
		newMultisPerBand, ok := c.multisPerBand[newQSO.Band]
		if !ok {
			newMultisPerBand = newMultis(c.multis, c.xchangeMultiExpression)
			c.multisPerBand[newQSO.Band] = newMultisPerBand
		}
		newBandMultiScore := newMultisPerBand.Add(1, newQSO.Callsign, newQSO.DXCC, newQSO.TheirXchange)
		newBandScore.Add(newBandMultiScore)
		totalScore.Add(newBandMultiScore)
	}

	c.TotalScore = totalScore
	c.OverallScore = overallScore
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
		result.SpecificCountryQSOs += value
		result.Points += value * c.specificCountryPoints
	case entity.PrimaryPrefix == c.stationEntity.PrimaryPrefix:
		result.SameCountryQSOs += value
		result.Points += value * c.sameCountryPoints
	case entity.Continent == c.stationEntity.Continent:
		result.SameContinentQSOs += value
		result.Points += value * c.sameContinentPoints
	default:
		result.OtherQSOs += value
		result.Points += value * c.otherPoints
	}

	return result
}

func (c *Counter) isSpecificCountry(entity dxcc.Prefix) bool {
	return c.specificCountryPrefixes[entity.PrimaryPrefix]
}

func (c *Counter) Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, _ core.Mode, xchange string) (points, multis int) {
	if c.countPerBand {
		qsoScore := c.qsoScore(1, entity)
		multisPerBand, ok := c.multisPerBand[band]
		if !ok {
			multisPerBand = newMultis(c.multis, c.xchangeMultiExpression)
		}

		return qsoScore.Points, multisPerBand.Value(callsign, entity, xchange)
	}
	qsoScore := c.qsoScore(1, entity)
	return qsoScore.Points, c.overallMultis.Value(callsign, entity, xchange)
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
		result.CQZones += value
	}

	oldITUZoneCount := m.ITUZones[entity.ITUZone]
	newITUZoneCount := oldITUZoneCount + value
	m.ITUZones[entity.ITUZone] = newITUZoneCount
	if oldITUZoneCount == 0 || newITUZoneCount == 0 {
		result.ITUZones += value
	}

	oldDXCCEntitiesCount := m.DXCCEntities[entity.PrimaryPrefix]
	newDXCCEntitiesCount := oldDXCCEntitiesCount + value
	m.DXCCEntities[entity.PrimaryPrefix] = newDXCCEntitiesCount
	if oldDXCCEntitiesCount == 0 || newDXCCEntitiesCount == 0 {
		result.DXCCEntities += value
	}

	wpxPrefix := WPXPrefix(callsign)
	if wpxPrefix != "" {
		oldWPXPrefixesCount := m.WPXPrefixes[wpxPrefix]
		newWPXPrefixesCount := oldWPXPrefixesCount + value
		m.WPXPrefixes[wpxPrefix] = newWPXPrefixesCount
		if oldWPXPrefixesCount == 0 || newWPXPrefixesCount == 0 {
			result.WPXPrefixes += value
		}
	}

	xchangeMulti, xchangeMatch := m.matchXchange(xchange)
	if xchangeMatch {
		oldXchangeValuesCount := m.XchangeValues[xchangeMulti]
		newXchangeValuesCount := oldXchangeValuesCount + value
		m.XchangeValues[xchangeMulti] = newXchangeValuesCount
		if oldXchangeValuesCount == 0 || newXchangeValuesCount == 0 {
			result.XchangeValues += value
		}
	}

	if m.CountingMultis.DXCC {
		result.Multis += result.DXCCEntities
	}
	if m.CountingMultis.WPX {
		result.Multis += result.WPXPrefixes
	}
	if m.CountingMultis.Xchange {
		result.Multis += result.XchangeValues
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

type nullView struct{}

func (v *nullView) Show()                      {}
func (v *nullView) Hide()                      {}
func (v *nullView) ShowScore(score core.Score) {}
