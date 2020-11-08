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
	return &Counter{}
}

type Counter struct {
	core.Score
	myPrefix  dxcc.Prefix
	listeners []interface{}
}

func (c *Counter) Notify(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}

func (c *Counter) SetMyPrefix(myPrefix dxcc.Prefix) {
	c.myPrefix = myPrefix
}

func (c *Counter) Clear() {
	c.Score = core.Score{}
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Add(qso core.QSO) {
	c.TotalQSOs++
	c.addCountry(qso.DXCC)
	c.emitScoreUpdated(c.Score)
}

func (c *Counter) Update(oldQSO, newQSO core.QSO) {
	var updated bool
	if oldQSO.DXCC != newQSO.DXCC {
		c.subtractCountry(oldQSO.DXCC)
		c.addCountry(oldQSO.DXCC)
	}
	if updated {
		c.emitScoreUpdated(c.Score)
	}
}

func (c *Counter) emitScoreUpdated(score core.Score) {
	for _, listener := range c.listeners {
		if scoreUpdatedListener, ok := listener.(ScoreUpdatedListener); ok {
			scoreUpdatedListener.ScoreUpdated(score)
		}
	}
}

func (c *Counter) addCountry(prefix dxcc.Prefix) {
	switch {
	case prefix.PrimaryPrefix == c.myPrefix.PrimaryPrefix:
		c.SameCountry++
	case prefix.Continent == c.myPrefix.Continent:
		c.SameContinent++
	default:
		c.DifferentContinent++
	}
}

func (c *Counter) subtractCountry(prefix dxcc.Prefix) {
	switch {
	case prefix.PrimaryPrefix == c.myPrefix.PrimaryPrefix:
		c.SameCountry++
	case prefix.Continent == c.myPrefix.Continent:
		c.SameContinent++
	default:
		c.DifferentContinent++
	}
}
