package logbook

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

func TestFindIndex(t *testing.T) {
	list := toQSOList(1, 2, 3, 4, 6)
	tt := []struct {
		name          string
		number        core.QSONumber
		exists        bool
		expectedIndex int
	}{
		{"first", 1, true, 0},
		{"last", 6, true, 4},
		{"gap", 5, false, 4},
		{"next", 7, false, 5},
		{"future", 100, false, 5},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actualIndex, found := list.findIndex(tc.number)
			if tc.exists {
				assert.True(t, found)
				assert.Equal(t, tc.expectedIndex, actualIndex)
			} else {
				assert.False(t, found)
			}
		})
	}
}

func TestPut_Append(t *testing.T) {
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
			list := NewQSOList(new(testSettings), new(testScorer))
			list.list = tc.qsos
			list.Put(core.QSO{MyNumber: tc.number})
			require.True(t, len(list.list) > 0, "list must not be empty")
			assert.Equal(t, tc.number, list.list[len(list.list)-1].MyNumber)
		})
	}
}

func TestPut_Insert(t *testing.T) {
	tt := []struct {
		name         string
		qsos         []core.QSO
		number       int
		expectedQSOs []core.QSO
	}{
		{"first", toQSOs(2), 1, toQSOs(1, 2)},
		{"middle", toQSOs(1, 3), 2, toQSOs(1, 2, 3)},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			list := NewQSOList(new(testSettings), new(testScorer))
			list.list = tc.qsos
			list.Put(toQSO(tc.number))
			require.Equal(t, len(tc.expectedQSOs), len(list.list), "list has wrong length")
			assert.Equal(t, tc.expectedQSOs, list.list)
		})
	}
}

func TestPut_Update(t *testing.T) {
	tt := []struct {
		name          string
		qsos          []core.QSO
		number        int
		expectedIndex int
	}{
		{"first", toQSOs(1, 2), 1, 0},
		{"middle", toQSOs(1, 2, 3), 2, 1},
		{"last", toQSOs(1, 2, 3), 3, 2},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			list := NewQSOList(new(testSettings), new(testScorer))
			list.list = tc.qsos
			expectedQSO := toQSO(tc.number)
			expectedQSO.TheirNumber = 100
			list.Put(expectedQSO)
			assert.Equal(t, expectedQSO, list.list[tc.expectedIndex])
		})
	}
}

func TestPut_Add_ScoreQSO(t *testing.T) {
	list := NewQSOList(new(testSettings), &testScorer{
		scores: map[string]core.QSOScore{
			"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
			"K3LR":   {Points: 6, Multis: 7, Duplicate: false},
			"DK9ZZ":  {Points: 3, Multis: 4, Duplicate: false},
		},
	})

	list.Put(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1})
	list.Put(core.QSO{Callsign: callsign.MustParse("K3LR"), MyNumber: 3})
	list.Put(core.QSO{Callsign: callsign.MustParse("DK9ZZ"), MyNumber: 2})

	qsos := list.All()

	assert.Equal(t, 1, qsos[0].Points)
	assert.Equal(t, 2, qsos[0].Multis)
	assert.Equal(t, 3, qsos[1].Points)
	assert.Equal(t, 4, qsos[1].Multis)
	assert.Equal(t, 6, qsos[2].Points)
	assert.Equal(t, 7, qsos[2].Multis)
}

func TestPut_Update_ScoreQSO(t *testing.T) {
	list := NewQSOList(new(testSettings), &testScorer{
		scores: map[string]core.QSOScore{
			"DL1ABC": {Points: 1, Multis: 2, Duplicate: false},
			"K3LR":   {Points: 6, Multis: 7, Duplicate: false},
		},
	})

	list.Put(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1})
	list.Put(core.QSO{Callsign: callsign.MustParse("K3LR"), MyNumber: 1})

	qsos := list.All()

	assert.Equal(t, 6, qsos[0].Points)
	assert.Equal(t, 7, qsos[0].Multis)
}

func TestDuplicateMarkers(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	dl2abc := callsign.MustParse("DL2ABC")
	list := NewQSOList(new(testSettings), new(testScorer))

	list.Put(core.QSO{Callsign: dl1abc, MyNumber: 1})
	assert.False(t, list.list[0].Duplicate, "first qso")

	list.Put(core.QSO{Callsign: dl1abc, MyNumber: 3})
	assert.False(t, list.list[0].Duplicate, "first qso, duplicate")
	assert.True(t, list.list[1].Duplicate, "duplicate of first qso")

	list.Put(core.QSO{Callsign: dl2abc, MyNumber: 1})
	assert.False(t, list.list[0].Duplicate, "first qso, edited")
	assert.False(t, list.list[1].Duplicate, "second qso, after edit")

	list.Put(core.QSO{Callsign: dl1abc, MyNumber: 2})
	assert.False(t, list.list[0].Duplicate, "first qso, after insert")
	assert.False(t, list.list[1].Duplicate, "inserted qso")
	assert.True(t, list.list[2].Duplicate, "second qso, after insert")
}

func TestFindDuplicateQSOs(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	dl2abc := callsign.MustParse("DL2ABC")
	list := NewQSOList(new(testSettings), new(testScorer))
	list.Put(core.QSO{Callsign: dl1abc, MyNumber: 1})
	list.Put(core.QSO{Callsign: dl2abc, MyNumber: 3})
	list.Put(core.QSO{Callsign: dl1abc, MyNumber: 2})

	dupes := list.FindDuplicateQSOs(dl1abc, core.NoBand, core.NoMode)
	assert.Equal(t, []core.QSO{
		{Callsign: dl1abc, MyNumber: 1},
		{Callsign: dl1abc, MyNumber: 2, Duplicate: true},
	}, dupes)
}

func TestFindWorkedQSOs(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	dl2abc := callsign.MustParse("DL2ABC")
	list := NewQSOList(new(testSettings), new(testScorer))
	list.bandRule = conval.Once
	list.Put(core.QSO{Callsign: dl1abc, MyNumber: 1})
	list.Put(core.QSO{Callsign: dl2abc, MyNumber: 3})
	list.Put(core.QSO{Callsign: dl1abc, MyNumber: 2})
	list.Put(core.QSO{Callsign: dl2abc, MyNumber: 2})

	workedDL1ABC, dupeDL1ABC := list.FindWorkedQSOs(dl1abc, core.NoBand, core.NoMode)
	assert.Equal(t, []core.QSO{
		{Callsign: dl1abc, MyNumber: 1},
	}, workedDL1ABC, "dl1abc worked")
	assert.True(t, dupeDL1ABC, "dl1abc dupe")

	workedDL2ABC, dupeDL2ABC := list.FindWorkedQSOs(dl2abc, core.NoBand, core.NoMode)
	assert.Equal(t, []core.QSO{
		{Callsign: dl2abc, MyNumber: 2},
		{Callsign: dl2abc, MyNumber: 3, Duplicate: true},
	}, workedDL2ABC, "dl2abc worked")
	assert.True(t, dupeDL2ABC, "dl2abc dupe")
}

func TestSelectQSO(t *testing.T) {
	qso := core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1}
	list := NewQSOList(new(testSettings), new(testScorer))
	list.Put(qso)
	list.Put(core.QSO{Callsign: callsign.MustParse("K3LR"), MyNumber: 2})
	qsoNotified := false
	indexNotified := false
	list.Notify(QSOSelectedListenerFunc(func(selectedQSO core.QSO) {
		qsoNotified = true
		assert.Equal(t, qso, selectedQSO)
	}))
	list.Notify(RowSelectedListenerFunc(func(index int) {
		indexNotified = true
		assert.Equal(t, 0, index)
	}))

	list.SelectQSO(qso)

	assert.True(t, qsoNotified, "qsoNotified")
	assert.True(t, indexNotified, "indexNotified")
}

func TestSelectRow(t *testing.T) {
	qso := core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1}
	list := NewQSOList(new(testSettings), new(testScorer))
	list.Put(qso)
	list.Put(core.QSO{Callsign: callsign.MustParse("K3LR"), MyNumber: 2})
	qsoNotified := false
	indexNotified := false
	list.Notify(QSOSelectedListenerFunc(func(selectedQSO core.QSO) {
		qsoNotified = true
		assert.Equal(t, qso, selectedQSO)
	}))
	list.Notify(RowSelectedListenerFunc(func(index int) {
		indexNotified = true
		assert.Equal(t, 0, index)
	}))

	list.SelectRow(0)

	assert.True(t, qsoNotified, "qsoNotified")
	assert.True(t, indexNotified, "indexNotified")
}

func TestSelectLastQSO(t *testing.T) {
	qso := core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1}
	lastQSO := core.QSO{Callsign: callsign.MustParse("K3LR"), MyNumber: 2}
	list := NewQSOList(new(testSettings), new(testScorer))
	list.Put(qso)
	list.Put(lastQSO)
	qsoNotified := false
	indexNotified := false
	list.Notify(QSOSelectedListenerFunc(func(selectedQSO core.QSO) {
		qsoNotified = true
		assert.Equal(t, lastQSO, selectedQSO)
	}))
	list.Notify(RowSelectedListenerFunc(func(index int) {
		indexNotified = true
		assert.Equal(t, 1, index)
	}))

	list.SelectLastQSO()

	assert.True(t, qsoNotified, "qsoNotified")
	assert.True(t, indexNotified, "indexNotified")
}

func TestFind(t *testing.T) {
	list := NewQSOList(new(testSettings), new(testScorer))
	aa1zzz := callsign.MustParse("AA1ZZZ")
	list.Put(core.QSO{Callsign: aa1zzz, Band: core.Band10m, Mode: core.ModeCW, MyNumber: core.QSONumber(1)})
	list.Put(core.QSO{Callsign: aa1zzz, Band: core.Band10m, Mode: core.ModeSSB, MyNumber: core.QSONumber(2)})
	list.Put(core.QSO{Callsign: aa1zzz, Band: core.Band20m, Mode: core.ModeCW, MyNumber: core.QSONumber(3)})
	list.Put(core.QSO{Callsign: aa1zzz, Band: core.Band20m, Mode: core.ModeSSB, MyNumber: core.QSONumber(4)})
	list.Put(core.QSO{Callsign: aa1zzz, Band: core.Band20m, Mode: core.ModeRTTY, MyNumber: core.QSONumber(5)})

	testCases := []struct {
		band        core.Band
		mode        core.Mode
		expectedLen int
	}{
		{core.NoBand, core.NoMode, 5},
		{core.Band10m, core.NoMode, 2},
		{core.Band20m, core.NoMode, 3},
		{core.Band10m, core.ModeCW, 1},
		{core.Band10m, core.ModeRTTY, 0},
		{core.NoBand, core.ModeCW, 2},
	}
	for _, tC := range testCases {
		t.Run(fmt.Sprintf("%v, %v", tC.band, tC.mode), func(t *testing.T) {
			qsos := list.Find(aa1zzz, tC.band, tC.mode)
			assert.Equal(t, tC.expectedLen, len(qsos))
		})
	}
}

func TestDoNotFindEditedCallsign(t *testing.T) {
	list := NewQSOList(new(testSettings), new(testScorer))
	aa1zzz := callsign.MustParse("AA1ZZZ")
	a1bc := callsign.MustParse("A1BC")
	list.Put(core.QSO{Callsign: aa1zzz, MyNumber: core.QSONumber(5)})
	list.Put(core.QSO{Callsign: a1bc, MyNumber: core.QSONumber(5)})

	assert.Empty(t, list.Find(aa1zzz, core.NoBand, core.NoMode))
	newQSOs := list.Find(a1bc, core.NoBand, core.NoMode)
	require.Equal(t, 1, len(newQSOs))
	assert.Equal(t, core.QSONumber(5), newQSOs[0].MyNumber)
	assert.Equal(t, a1bc, newQSOs[0].Callsign)
}

func TestQSOAddedListener(t *testing.T) {
	list := NewQSOList(new(testSettings), new(testScorer))
	qso := core.QSO{MyNumber: 1}
	notified := false
	list.Notify(QSOAddedListenerFunc(func(addedQSO core.QSO) {
		notified = true
		assert.Equal(t, qso, addedQSO)
	}))

	list.Put(qso)

	assert.True(t, notified)
}

func toQSOList(numbers ...int) *QSOList {
	qsos := make([]core.QSO, len(numbers))
	for i, number := range numbers {
		qsos[i] = toQSO(number)
	}
	result := NewQSOList(new(testSettings), new(testScorer))
	result.list = qsos
	return result
}

func toQSOs(numbers ...int) []core.QSO {
	result := make([]core.QSO, len(numbers))
	for i, number := range numbers {
		result[i] = toQSO(number)
	}
	return result
}

func toQSO(number int) core.QSO {
	return core.QSO{Callsign: callsign.MustParse(fmt.Sprintf("DL%dNN", number)), MyNumber: core.QSONumber(number)}
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

func (s *testScorer) Add(qso core.QSO) core.QSOScore {
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
