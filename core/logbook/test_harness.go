package logbook

import (
	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core/dxcc"
)

var _ convalCounter = new(testConvalCounter)

type testConvalCounter struct {
	scores          map[string]conval.QSOScore
	workedCallsigns map[string]bool
}

func (t *testConvalCounter) markAsWorked(call callsign.Callsign) {
	if t.workedCallsigns == nil {
		t.workedCallsigns = make(map[string]bool)
	}
	t.workedCallsigns[call.String()] = true
}

func (t *testConvalCounter) worked(call callsign.Callsign) bool {
	if t.workedCallsigns == nil {
		return false
	}
	return t.workedCallsigns[call.String()]
}

func (t *testConvalCounter) Add(qso conval.QSO) conval.QSOScore {
	score := t.scores[qso.TheirCall.String()]
	score.Duplicate = t.worked(qso.TheirCall)
	t.markAsWorked(qso.TheirCall)
	return score
}

func (t *testConvalCounter) Probe(qso conval.QSO) conval.QSOScore {
	score := t.scores[qso.TheirCall.String()]
	score.Duplicate = t.worked(qso.TheirCall)
	return score
}

func (t *testConvalCounter) AddQTC(qtc conval.QTC) conval.QTCScore {
	return conval.QTCScore{
		Value: qtc.Count,
	}
}

func withTestConvalCounter(logbook *Logbook, counter *testConvalCounter) *Logbook {
	logbook.scoreCounter.counterFactory = func() convalCounter {
		return counter
	}
	logbook.scoreCounter.counter = logbook.scoreCounter.counterFactory()
	return logbook
}

var _ DXCCEntities = new(testEntities)

var testEntity = &testEntities{entity: dxcc.Prefix{Prefix: "DL", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}}

type testEntities struct {
	entity dxcc.Prefix
}

func (e *testEntities) Available() bool {
	return true
}

func (e *testEntities) Find(string) (dxcc.Prefix, bool) {
	return e.entity, (e.entity.PrimaryPrefix != "")
}
