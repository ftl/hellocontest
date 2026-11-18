package logbook

import (
	"time"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core/model"
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

var toConvalMode = map[model.Mode]conval.Mode{
	model.ModeCW:      conval.ModeCW,
	model.ModeSSB:     conval.ModeSSB,
	model.ModeFM:      conval.ModeFM,
	model.ModeRTTY:    conval.ModeRTTY,
	model.ModeDigital: conval.ModeDigital,
}

var fromConvalMode = map[conval.Mode]model.Mode{
	conval.ModeCW:      model.ModeCW,
	conval.ModeSSB:     model.ModeSSB,
	conval.ModeFM:      model.ModeFM,
	conval.ModeRTTY:    model.ModeRTTY,
	conval.ModeDigital: model.ModeDigital,
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
