package logbook

import (
	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core/dxcc"
)

var _ convalCounter = new(testConvalCounter)

type testConvalCounter struct {
	scores map[string]conval.QSOScore
}

func (t *testConvalCounter) Add(qso conval.QSO) conval.QSOScore {
	return t.scores[qso.TheirCall.String()]
}

func (t *testConvalCounter) Probe(qso conval.QSO) conval.QSOScore {
	return t.scores[qso.TheirCall.String()]
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

func (e *testEntities) Find(string) (dxcc.Prefix, bool) {
	return e.entity, (e.entity.PrimaryPrefix != "")
}
