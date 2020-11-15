package score

import (
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
}

const (
	CQZoneMulti     = "CQ"
	ITUZoneMulti    = "ITU"
	DXCCEntityMulti = "DXCC"
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
		overallMultis:           newMultis(configuration.Multis()),
	}

	for _, prefix := range configuration.SpecificCountryPrefixes() {
		result.specificCountryPrefixes[strings.ToUpper(prefix)] = true
	}

	return result
}

type Counter struct {
	core.Score
	view View

	configuration           Configuration
	myPrefix                dxcc.Prefix
	specificCountryPrefixes map[string]bool
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

func (c *Counter) SetMyPrefix(myPrefix dxcc.Prefix) {
	c.myPrefix = myPrefix
}

func (c *Counter) Clear() {
	c.Score = core.Score{
		ScorePerBand: make(map[core.Band]core.BandScore),
	}
	c.multisPerBand = make(map[core.Band]*multis)
	c.overallMultis = newMultis(c.configuration.Multis())
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Add(qso core.QSO) {
	bandScore := c.ScorePerBand[qso.Band]

	qsoScore := c.qsoScore(1, qso.DXCC)
	c.TotalScore.Add(qsoScore)
	c.OverallScore.Add(qsoScore)
	bandScore.Add(qsoScore)

	overallMultiScore := c.overallMultis.Add(1, qso.DXCC)
	c.OverallScore.Add(overallMultiScore)
	multisPerBand, ok := c.multisPerBand[qso.Band]
	if !ok {
		multisPerBand = newMultis(c.configuration.Multis())
		c.multisPerBand[qso.Band] = multisPerBand
	}
	bandMultiScore := multisPerBand.Add(1, qso.DXCC)
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

	oldOverallMultiScore := c.overallMultis.Add(-1, oldQSO.DXCC)
	overallScore.Add(oldOverallMultiScore)
	oldMultisPerBand, ok := c.multisPerBand[oldQSO.Band]
	if ok {
		oldBandMultiScore := oldMultisPerBand.Add(-1, oldQSO.DXCC)
		oldBandScore.Add(oldBandMultiScore)
		totalScore.Add(oldBandMultiScore)
	}

	newOverallMultiScore := c.overallMultis.Add(1, newQSO.DXCC)
	overallScore.Add(newOverallMultiScore)
	newMultisPerBand, ok := c.multisPerBand[newQSO.Band]
	if !ok {
		newMultisPerBand = newMultis(c.configuration.Multis())
		c.multisPerBand[newQSO.Band] = newMultisPerBand
	}
	newBandMultiScore := newMultisPerBand.Add(1, newQSO.DXCC)
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

func (c *Counter) qsoScore(value int, prefix dxcc.Prefix) core.BandScore {
	var result core.BandScore
	switch {
	case c.isSpecificCountry(prefix):
		result.SpecificCountryQSOs += value
		result.Points += value * c.configuration.SpecificCountryPoints()
	case prefix.PrimaryPrefix == c.myPrefix.PrimaryPrefix:
		result.SameCountryQSOs += value
		result.Points += value * c.configuration.SameCountryPoints()
	case prefix.Continent == c.myPrefix.Continent:
		result.SameContinentQSOs += value
		result.Points += value * c.configuration.SameContinentPoints()
	default:
		result.OtherQSOs += value
		result.Points += value * c.configuration.OtherPoints()
	}

	return result
}

func (c *Counter) isSpecificCountry(prefix dxcc.Prefix) bool {
	return c.specificCountryPrefixes[prefix.PrimaryPrefix]
}

func newMultis(countingMultis []string) *multis {
	return &multis{
		CountingMultis:  countingMultis,
		CQZones:         make(map[dxcc.CQZone]int),
		ITUZones:        make(map[dxcc.ITUZone]int),
		PrimaryPrefixes: make(map[string]int),
	}
}

type multis struct {
	CountingMultis  []string
	CQZones         map[dxcc.CQZone]int
	ITUZones        map[dxcc.ITUZone]int
	PrimaryPrefixes map[string]int
}

func (m *multis) Add(value int, prefix dxcc.Prefix) core.BandScore {
	var result core.BandScore

	oldCQZoneCount := m.CQZones[prefix.CQZone]
	newCQZoneCount := oldCQZoneCount + value
	m.CQZones[prefix.CQZone] = newCQZoneCount
	if oldCQZoneCount == 0 || newCQZoneCount == 0 {
		result.CQZones += value
	}

	oldITUZoneCount := m.ITUZones[prefix.ITUZone]
	newITUZoneCount := oldITUZoneCount + value
	m.ITUZones[prefix.ITUZone] = newITUZoneCount
	if oldITUZoneCount == 0 || newITUZoneCount == 0 {
		result.ITUZones += value
	}

	oldPrimaryPrefixCount := m.PrimaryPrefixes[prefix.PrimaryPrefix]
	newPrimaryPrefixCount := oldPrimaryPrefixCount + value
	m.PrimaryPrefixes[prefix.PrimaryPrefix] = newPrimaryPrefixCount
	if oldPrimaryPrefixCount == 0 || newPrimaryPrefixCount == 0 {
		result.PrimaryPrefixes += value
	}

	for _, token := range m.CountingMultis {
		switch strings.ToUpper(token) {
		case CQZoneMulti:
			result.Multis += result.CQZones
		case ITUZoneMulti:
			result.Multis += result.ITUZones
		case DXCCEntityMulti:
			result.Multis += result.PrimaryPrefixes
		}
	}

	return result
}

type nullView struct{}

func (v *nullView) Show()                      {}
func (v *nullView) Hide()                      {}
func (v *nullView) ShowScore(score core.Score) {}
