package logbook

import (
	"sync"
	"time"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core"
)

type convalCounter interface {
	Add(conval.QSO) conval.QSOScore
	Probe(conval.QSO) conval.QSOScore
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

type safeConvalCounter struct {
	counter   convalCounter
	timeSheet convalTimeSheet

	bands map[conval.ContestBand]bool
	modes map[conval.Mode]bool

	lock *sync.RWMutex
}

func newSafeConvalCounter(counter convalCounter, timeSheet convalTimeSheet) *safeConvalCounter {
	return &safeConvalCounter{
		counter:   counter,
		timeSheet: timeSheet,
		bands:     make(map[conval.ContestBand]bool),
		modes:     make(map[conval.Mode]bool),
		lock:      &sync.RWMutex{},
	}
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

func (c *safeConvalCounter) SummaryContent(definition *conval.Definition, operatorMode conval.OperatorMode, overlay conval.Overlay) (core.TimeReport, []string, []string) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	timeReport := c.timeSheet.TimeReport(c.computeMinBreakDuration(definition, operatorMode, overlay))
	bands := make([]string, 0, len(c.bands))
	for _, band := range core.Bands {
		if c.bands[conval.ContestBand(band)] {
			bands = append(bands, string(band))
		}
	}
	modes := make([]string, 0, len(c.modes))
	for _, mode := range core.Modes {
		if c.modes[toConvalMode[mode]] {
			modes = append(modes, string(mode))
		}
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

var _ convalCounter = new(nullCounter)

type nullCounter struct{}

func (c *nullCounter) Add(conval.QSO) conval.QSOScore   { return conval.QSOScore{} }
func (c *nullCounter) Probe(conval.QSO) conval.QSOScore { return conval.QSOScore{} }

var _ convalTimeSheet = new(nullTimeSheet)

type nullTimeSheet struct{}

func (t *nullTimeSheet) MarkActive(time.Time) {}
func (t *nullTimeSheet) TimeReport(minBreakDuration time.Duration) conval.TimeReport {
	return conval.TimeReport{}
}
