package logbook

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestQSOList_QSOAdded(t *testing.T) {
	tt := []struct {
		name   string
		qsos   []core.QSO
		number core.QSONumber
	}{
		{"empty", []core.QSO{}, 1},
		{"empty with high number", []core.QSO{}, 100},
		{"next", toQSOs(1, 2, 3), 4},
		{"future", toQSOs(1, 2, 3), 400},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			list := NewQSOList(new(testSettings))
			list.list = tc.qsos
			list.QSOAdded(core.QSO{MyNumber: tc.number})
			require.True(t, len(list.list) > 0, "list must not be empty")
			assert.Equal(t, tc.number, list.list[len(list.list)-1].MyNumber)
		})
	}
}

func TestQSOList_SelectRow(t *testing.T) {
	qso := core.QSO{Callsign: core.MustParseCallsign("DL1ABC"), MyNumber: 1}
	list := NewQSOList(new(testSettings))
	list.QSOAdded(qso)
	list.QSOAdded(core.QSO{Callsign: core.MustParseCallsign("K3LR"), MyNumber: 2})
	qsoNotified := false
	indexNotified := false
	list.Notify(QSOSelectedListenerFunc(func(selectedQSO core.QSO) {
		qsoNotified = true
		assert.Equal(t, qso, selectedQSO)
	}))
	list.Notify(QSORowSelectedListenerFunc(func(index int) {
		indexNotified = true
		assert.Equal(t, 0, index)
	}))

	list.SelectRow(0)

	assert.True(t, qsoNotified, "qsoNotified")
	assert.True(t, indexNotified, "indexNotified")
}

func TestQSOList_SelectLastQSO(t *testing.T) {
	qso := core.QSO{Callsign: core.MustParseCallsign("DL1ABC"), MyNumber: 1}
	lastQSO := core.QSO{Callsign: core.MustParseCallsign("K3LR"), MyNumber: 2}
	list := NewQSOList(new(testSettings))
	list.QSOAdded(qso)
	list.QSOAdded(lastQSO)
	qsoNotified := false
	indexNotified := false
	list.Notify(QSOSelectedListenerFunc(func(selectedQSO core.QSO) {
		qsoNotified = true
		assert.Equal(t, lastQSO, selectedQSO)
	}))
	list.Notify(QSORowSelectedListenerFunc(func(index int) {
		indexNotified = true
		assert.Equal(t, 1, index)
	}))

	list.SelectLastQSO()

	assert.True(t, qsoNotified, "qsoNotified")
	assert.True(t, indexNotified, "indexNotified")
}

func TestQSOList_QSOAddedListener(t *testing.T) {
	list := NewQSOList(new(testSettings))
	qso := core.QSO{MyNumber: 1}
	notified := false
	list.Notify(QSOAddedListenerFunc(func(addedQSO core.QSO) {
		notified = true
		assert.Equal(t, qso, addedQSO)
	}))

	list.QSOAdded(qso)

	assert.True(t, notified)
}

func toQSOs(numbers ...int) []core.QSO {
	result := make([]core.QSO, len(numbers))
	for i, number := range numbers {
		result[i] = toQSO(number)
	}
	return result
}

func toQSO(number int) core.QSO {
	return core.QSO{Callsign: core.MustParseCallsign(fmt.Sprintf("DL%dNN", number)), MyNumber: core.QSONumber(number)}
}

type testSettings struct{}

func (c *testSettings) Station() core.Station {
	return core.Station{}
}

func (c *testSettings) Contest() core.Contest {
	return core.Contest{}
}

type testScorer struct {
	scores map[string]core.QSOScore
	worked []string
}

func (s *testScorer) Clear() {
	s.worked = make([]string, 1)
}

func (s *testScorer) AddMuted(qso core.QSO) core.QSOScore {
	if s.scores == nil {
		s.scores = make(map[string]core.QSOScore)
	}
	if s.worked == nil {
		s.worked = make([]string, 0)
	}

	callsign := qso.Callsign.String()
	result := s.scores[callsign]
	for _, w := range s.worked {
		if w == callsign {
			result.Duplicate = true
			break
		}
	}
	s.worked = append(s.worked, callsign)
	return result
}

func (s *testScorer) Unmute() {}
