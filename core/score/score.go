package score

import (
	"log"
	"regexp"
	"strings"

	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

type ScoreUpdatedListener interface {
	ScoreUpdated(core.Score)
}

type ScoreUpdatedListenerFunc func(core.Score)

func (f ScoreUpdatedListenerFunc) ScoreUpdated(Score core.Score) {
	f(Score)
}

type Configuration interface {
	CountPerBand() bool
	SameCountryPoints() int
	SameContinentPoints() int
	OtherPoints() int
	SpecificCountryPoints() int
	SpecificCountryPrefixes() []string
	Multis() []string
	XchangeMultiPattern() string
}

const (
	CQZoneMulti     = "CQ"
	ITUZoneMulti    = "ITU"
	DXCCEntityMulti = "DXCC"
	XchangeMulti    = "XCHANGE"
)

func NewCounter(configuration Configuration) *Counter {
	result := &Counter{
		Score: core.Score{
			ScorePerBand: make(map[core.Band]core.BandScore),
		},
		view:                    new(nullView),
		configuration:           configuration,
		specificCountryPrefixes: make(map[string]bool),
		multisPerBand:           make(map[core.Band]*multis),
	}

	for _, prefix := range configuration.SpecificCountryPrefixes() {
		result.specificCountryPrefixes[strings.ToUpper(prefix)] = true
	}

	exp, err := regexp.Compile(configuration.XchangeMultiPattern())
	if err != nil {
		log.Printf("Invalid regular expression for Xchange Multis: %v", err)
	} else {
		result.xchangeMultiExpression = exp
		log.Printf("Using pattern %q for Xchange Multis", result.xchangeMultiExpression)
	}
	result.overallMultis = newMultis(configuration.Multis(), result.xchangeMultiExpression)

	return result
}

type Counter struct {
	core.Score
	view View

	configuration           Configuration
	myEntity                dxcc.Prefix
	specificCountryPrefixes map[string]bool
	xchangeMultiExpression  *regexp.Regexp
	listeners               []interface{}

	multisPerBand map[core.Band]*multis
	overallMultis *multis
}

type View interface {
	Show()
	Hide()

	ShowScore(score core.Score)
}

func (c *Counter) SetView(view View) {
	if view == nil {
		c.view = new(nullView)
		return
	}
	c.view = view
	c.view.ShowScore(c.Score)
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

func (c *Counter) SetMyEntity(myEntity dxcc.Prefix) {
	c.myEntity = myEntity
}

func (c *Counter) Clear() {
	c.Score = core.Score{
		ScorePerBand: make(map[core.Band]core.BandScore),
	}
	c.multisPerBand = make(map[core.Band]*multis)
	c.overallMultis = newMultis(c.configuration.Multis(), c.xchangeMultiExpression)
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Add(qso core.QSO) {
	bandScore := c.ScorePerBand[qso.Band]

	qsoScore := c.qsoScore(1, qso.DXCC)
	c.TotalScore.Add(qsoScore)
	c.OverallScore.Add(qsoScore)
	bandScore.Add(qsoScore)

	overallMultiScore := c.overallMultis.Add(1, qso.DXCC, qso.TheirXchange)
	c.OverallScore.Add(overallMultiScore)
	multisPerBand, ok := c.multisPerBand[qso.Band]
	if !ok {
		multisPerBand = newMultis(c.configuration.Multis(), c.xchangeMultiExpression)
		c.multisPerBand[qso.Band] = multisPerBand
	}
	bandMultiScore := multisPerBand.Add(1, qso.DXCC, qso.TheirXchange)
	c.TotalScore.Add(bandMultiScore)
	bandScore.Add(bandMultiScore)

	c.ScorePerBand[qso.Band] = bandScore
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Update(oldQSO, newQSO core.QSO) {
	if oldQSO.DXCC == newQSO.DXCC {
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

	oldQSOScore := c.qsoScore(-1, oldQSO.DXCC)
	totalScore.Add(oldQSOScore)
	overallScore.Add(oldQSOScore)
	oldBandScore.Add(oldQSOScore)

	newQSOScore := c.qsoScore(1, newQSO.DXCC)
	totalScore.Add(newQSOScore)
	overallScore.Add(newQSOScore)
	newBandScore.Add(newQSOScore)

	oldOverallMultiScore := c.overallMultis.Add(-1, oldQSO.DXCC, oldQSO.TheirXchange)
	overallScore.Add(oldOverallMultiScore)
	oldMultisPerBand, ok := c.multisPerBand[oldQSO.Band]
	if ok {
		oldBandMultiScore := oldMultisPerBand.Add(-1, oldQSO.DXCC, oldQSO.TheirXchange)
		oldBandScore.Add(oldBandMultiScore)
		totalScore.Add(oldBandMultiScore)
	}

	newOverallMultiScore := c.overallMultis.Add(1, newQSO.DXCC, newQSO.TheirXchange)
	overallScore.Add(newOverallMultiScore)
	newMultisPerBand, ok := c.multisPerBand[newQSO.Band]
	if !ok {
		newMultisPerBand = newMultis(c.configuration.Multis(), c.xchangeMultiExpression)
		c.multisPerBand[newQSO.Band] = newMultisPerBand
	}
	newBandMultiScore := newMultisPerBand.Add(1, newQSO.DXCC, newQSO.TheirXchange)
	newBandScore.Add(newBandMultiScore)
	totalScore.Add(newBandMultiScore)

	c.TotalScore = totalScore
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
		result.Points += value * c.configuration.SpecificCountryPoints()
	case entity.PrimaryPrefix == c.myEntity.PrimaryPrefix:
		result.SameCountryQSOs += value
		result.Points += value * c.configuration.SameCountryPoints()
	case entity.Continent == c.myEntity.Continent:
		result.SameContinentQSOs += value
		result.Points += value * c.configuration.SameContinentPoints()
	default:
		result.OtherQSOs += value
		result.Points += value * c.configuration.OtherPoints()
	}

	return result
}

func (c *Counter) isSpecificCountry(entity dxcc.Prefix) bool {
	return c.specificCountryPrefixes[entity.PrimaryPrefix]
}

func (c *Counter) Value(entity dxcc.Prefix, band core.Band, _ core.Mode, xchange string) (points, multis int) {
	if c.configuration.CountPerBand() {
		qsoScore := c.qsoScore(1, entity)
		multisPerBand, ok := c.multisPerBand[band]
		if !ok {
			multisPerBand = newMultis(c.configuration.Multis(), c.xchangeMultiExpression)
		}

		return qsoScore.Points, multisPerBand.Value(entity, xchange)
	}
	qsoScore := c.qsoScore(1, entity)
	return qsoScore.Points, c.overallMultis.Value(entity, xchange)
}

func newMultis(countingMultis []string, xchangeMultiExpression *regexp.Regexp) *multis {
	return &multis{
		CountingMultis:         countingMultis,
		XchangeMultiExpression: xchangeMultiExpression,
		CQZones:                make(map[dxcc.CQZone]int),
		ITUZones:               make(map[dxcc.ITUZone]int),
		PrimaryPrefixes:        make(map[string]int),
		XchangeValues:          make(map[string]int),
	}
}

type multis struct {
	CountingMultis         []string
	XchangeMultiExpression *regexp.Regexp
	CQZones                map[dxcc.CQZone]int
	ITUZones               map[dxcc.ITUZone]int
	PrimaryPrefixes        map[string]int
	XchangeValues          map[string]int
}

func (m *multis) Value(entity dxcc.Prefix, xchange string) int {
	var cqZoneValue int
	if m.CQZones[entity.CQZone] == 0 {
		cqZoneValue = 1
	}
	var ituZoneValue int
	if m.ITUZones[entity.ITUZone] == 0 {
		ituZoneValue = 1
	}
	var primaryPrefixValue int
	if m.PrimaryPrefixes[entity.PrimaryPrefix] == 0 {
		primaryPrefixValue = 1
	}
	var xchangeQSOValue int
	xchange = strings.ToUpper(strings.TrimSpace(xchange))
	if m.XchangeMultiExpression != nil {
		matches := m.XchangeMultiExpression.FindStringSubmatch(xchange)
		if len(matches) > 0 {
			xchangeValue := matches[0]
			if len(matches) > 1 {
				xchangeValue = matches[1]
			}
			if m.XchangeValues[xchangeValue] == 0 {
				xchangeQSOValue = 1
			}
		}
	}

	var result int
	for _, token := range m.CountingMultis {
		switch strings.ToUpper(token) {
		case CQZoneMulti:
			result += cqZoneValue
		case ITUZoneMulti:
			result += ituZoneValue
		case DXCCEntityMulti:
			result += primaryPrefixValue
		case XchangeMulti:
			result += xchangeQSOValue
		}
	}
	return result
}

func (m *multis) Add(value int, entity dxcc.Prefix, xchange string) core.BandScore {
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

	oldPrimaryPrefixCount := m.PrimaryPrefixes[entity.PrimaryPrefix]
	newPrimaryPrefixCount := oldPrimaryPrefixCount + value
	m.PrimaryPrefixes[entity.PrimaryPrefix] = newPrimaryPrefixCount
	if oldPrimaryPrefixCount == 0 || newPrimaryPrefixCount == 0 {
		result.PrimaryPrefixes += value
	}

	xchange = strings.ToUpper(strings.TrimSpace(xchange))
	if m.XchangeMultiExpression != nil {
		matches := m.XchangeMultiExpression.FindStringSubmatch(xchange)
		if len(matches) > 0 {
			xchangeValue := matches[0]
			if len(matches) > 1 {
				xchangeValue = matches[1]
			}
			oldXchangeValuesCount := m.XchangeValues[xchangeValue]
			newXchangeValuesCount := oldXchangeValuesCount + value
			m.XchangeValues[xchangeValue] = newXchangeValuesCount
			if oldXchangeValuesCount == 0 || newXchangeValuesCount == 0 {
				result.XchangeValues += value
			}
		}
	}

	for _, token := range m.CountingMultis {
		switch strings.ToUpper(token) {
		case CQZoneMulti:
			result.Multis += result.CQZones
		case ITUZoneMulti:
			result.Multis += result.ITUZones
		case DXCCEntityMulti:
			result.Multis += result.PrimaryPrefixes
		case XchangeMulti:
			result.Multis += result.XchangeValues
		}
	}

	return result
}

type nullView struct{}

func (v *nullView) Show()                      {}
func (v *nullView) Hide()                      {}
func (v *nullView) ShowScore(score core.Score) {}
