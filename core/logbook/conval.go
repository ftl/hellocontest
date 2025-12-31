package logbook

import (
	"time"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core"
)

type convalCounter interface {
	Add(conval.QSO) conval.QSOScore
	Probe(conval.QSO) conval.QSOScore
	AddQTC(conval.QTC) conval.QTCScore
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

var _ convalCounter = new(nullCounter)

type nullCounter struct{}

func (c *nullCounter) Add(conval.QSO) conval.QSOScore    { return conval.QSOScore{} }
func (c *nullCounter) Probe(conval.QSO) conval.QSOScore  { return conval.QSOScore{} }
func (c *nullCounter) AddQTC(conval.QTC) conval.QTCScore { return conval.QTCScore{} }

var _ convalTimeSheet = new(nullTimeSheet)

type nullTimeSheet struct{}

func (t *nullTimeSheet) MarkActive(time.Time) {}
func (t *nullTimeSheet) TimeReport(minBreakDuration time.Duration) conval.TimeReport {
	return conval.TimeReport{}
}
