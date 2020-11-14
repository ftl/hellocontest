package score

import (
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

func NewCounter() *Counter {
	return &Counter{
		Score: core.Score{
			ScorePerBand: make(map[core.Band]core.BandScore),
		},
		multisPerBand: make(map[core.Band]*multis),
		overallMultis: newMultis(),
	}
}

type Counter struct {
	core.Score
	view View

	myPrefix  dxcc.Prefix
	listeners []interface{}

	multisPerBand map[core.Band]*multis
	overallMultis *multis
}

type View interface {
	Show()
	Hide()
	Visible() bool

	ShowScore(score core.Score)
}

func (c *Counter) SetView(view View) {
	c.view = view
	if c.view.Visible() {
		c.view.ShowScore(c.Score)
	}
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
	c.overallMultis = newMultis()
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Add(qso core.QSO) {
	bandScore := c.ScorePerBand[qso.Band]

	qsoScore := c.qsoScore(1, qso.DXCC)
	c.TotalScore.Add(qsoScore)
	c.OverallScore.Add(qsoScore)
	bandScore.Add(qsoScore)

	c.OverallScore.Add(c.overallMultis.Add(1, qso.DXCC))
	multisPerBand, ok := c.multisPerBand[qso.Band]
	if !ok {
		multisPerBand = newMultis()
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

	overallScore.Add(c.overallMultis.Add(-1, oldQSO.DXCC))
	oldMultisPerBand, ok := c.multisPerBand[oldQSO.Band]
	if ok {
		oldBandMultiScore := oldMultisPerBand.Add(-1, oldQSO.DXCC)
		oldBandScore.Add(oldBandMultiScore)
		totalScore.Add(oldBandMultiScore)
	}

	overallScore.Add(c.overallMultis.Add(1, newQSO.DXCC))
	newMultisPerBand, ok := c.multisPerBand[newQSO.Band]
	if !ok {
		newMultisPerBand = newMultis()
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
	if c.view != nil && c.view.Visible() {
		c.view.ShowScore(score)
	}
	for _, listener := range c.listeners {
		if scoreUpdatedListener, ok := listener.(ScoreUpdatedListener); ok {
			scoreUpdatedListener.ScoreUpdated(score)
		}
	}
}

func (c *Counter) qsoScore(value int, prefix dxcc.Prefix) core.BandScore {
	var result core.BandScore
	switch {
	case prefix.PrimaryPrefix == c.myPrefix.PrimaryPrefix:
		result.SameCountryQSOs += value
	case prefix.Continent == c.myPrefix.Continent:
		result.SameContinentQSOs += value
	default:
		result.OtherQSOs += value
	}

	// TODO add points

	return result
}

func newMultis() *multis {
	return &multis{
		CQZones:         make(map[dxcc.CQZone]int),
		ITUZones:        make(map[dxcc.ITUZone]int),
		PrimaryPrefixes: make(map[string]int),
	}
}

type multis struct {
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

	return result
}
