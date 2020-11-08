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
		totalMultis:   newMultis(),
	}
}

type Counter struct {
	core.Score

	myPrefix  dxcc.Prefix
	listeners []interface{}

	multisPerBand map[core.Band]*multis
	totalMultis   *multis
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
	c.totalMultis = newMultis()
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Add(qso core.QSO) {
	bandScore := c.ScorePerBand[qso.Band]

	qsoScore := c.qsoScore(1, qso.DXCC)
	c.TotalScore.Add(qsoScore)
	bandScore.Add(qsoScore)

	c.TotalScore.Add(c.totalMultis.Add(1, qso.DXCC))
	multisPerBand, ok := c.multisPerBand[qso.Band]
	if !ok {
		multisPerBand = newMultis()
		c.multisPerBand[qso.Band] = multisPerBand
	}
	bandScore.Add(multisPerBand.Add(1, qso.DXCC))

	c.ScorePerBand[qso.Band] = bandScore
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Update(oldQSO, newQSO core.QSO) {
	if oldQSO.DXCC == newQSO.DXCC {
		return
	}
	totalScore := c.TotalScore
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
	oldBandScore.Add(oldQSOScore)

	newQSOScore := c.qsoScore(1, newQSO.DXCC)
	totalScore.Add(newQSOScore)
	newBandScore.Add(newQSOScore)

	oldMultiScore := c.totalMultis.Add(-1, oldQSO.DXCC)
	totalScore.Add(oldMultiScore)
	oldMultisPerBand, ok := c.multisPerBand[oldQSO.Band]
	if ok {
		oldBandScore.Add(oldMultisPerBand.Add(-1, oldQSO.DXCC))
	}

	newMultiScore := c.totalMultis.Add(1, newQSO.DXCC)
	totalScore.Add(newMultiScore)
	newMultisPerBand, ok := c.multisPerBand[newQSO.Band]
	if !ok {
		newMultisPerBand = newMultis()
		c.multisPerBand[newQSO.Band] = newMultisPerBand
	}
	newBandScore.Add(newMultisPerBand.Add(1, newQSO.DXCC))

	c.TotalScore = totalScore
	c.ScorePerBand[oldQSO.Band] = oldBandScore
	c.ScorePerBand[newQSO.Band] = *newBandScore
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) emitScoreUpdated(score core.Score) {
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
