package multis

import (
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

type MultisUpdatedListener interface {
	MultisUpdated(core.Multis)
}

type MultisUpdatedListenerFunc func(core.Multis)

func (f MultisUpdatedListenerFunc) MultisUpdated(multis core.Multis) {
	f(multis)
}

func NewCounter() *Counter {
	return &Counter{}
}

type Counter struct {
	core.Multis
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
	c.Multis = core.Multis{}
	c.emitMultisUpdated(c.Multis)
}

func (c *Counter) Add(qso core.QSO) {
	c.TotalCount++
	c.addCountry(qso.DXCC)
	c.emitMultisUpdated(c.Multis)
}

func (c *Counter) Update(oldQSO, newQSO core.QSO) {
	var updated bool
	if oldQSO.DXCC != newQSO.DXCC {
		c.subtractCountry(oldQSO.DXCC)
		c.addCountry(oldQSO.DXCC)
	}
	if updated {
		c.emitMultisUpdated(c.Multis)
	}
}

func (c *Counter) emitMultisUpdated(multis core.Multis) {
	for _, listener := range c.listeners {
		if multisUpdatedListener, ok := listener.(MultisUpdatedListener); ok {
			multisUpdatedListener.MultisUpdated(multis)
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
