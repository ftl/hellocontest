package logbook

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/mocked"
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
			list := &QSOList{list: tc.qsos, dxccFinder: new(nullDXCCFinder)}
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
		number       core.QSONumber
		expectedQSOs []core.QSO
	}{
		{"first", toQSOs(2), 1, toQSOs(1, 2)},
		{"middle", toQSOs(1, 3), 2, toQSOs(1, 2, 3)},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			list := &QSOList{list: tc.qsos, dxccFinder: new(nullDXCCFinder)}
			list.Put(core.QSO{MyNumber: tc.number})
			require.Equal(t, len(tc.expectedQSOs), len(list.list), "list has wrong length")
			assert.Equal(t, tc.expectedQSOs, list.list)
		})
	}
}

func TestPut_Update(t *testing.T) {
	tt := []struct {
		name          string
		qsos          []core.QSO
		number        core.QSONumber
		expectedIndex int
	}{
		{"first", toQSOs(1, 2), 1, 0},
		{"middle", toQSOs(1, 2, 3), 2, 1},
		{"last", toQSOs(1, 2, 3), 3, 2},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			list := &QSOList{list: tc.qsos, dxccFinder: new(nullDXCCFinder)}
			expectedQSO := core.QSO{MyNumber: tc.number, TheirNumber: 100}
			list.Put(expectedQSO)
			assert.Equal(t, expectedQSO, list.list[tc.expectedIndex])
		})
	}
}

func TestPut_AddPrefix(t *testing.T) {
	dlPrefix := dxcc.Prefix{Name: "Fed. Rep. of Germany", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}
	dxccFinder := new(mocked.DXCCFinder)
	dxccFinder.On("Find", "DL1ABC").Return(dlPrefix, true)
	dxccFinder.On("Find", "DK9ZZ").Return(dlPrefix, true)
	dxccFinder.On("Find", "K3LR").Return(dxcc.Prefix{}, false)
	list := NewQSOList(dxccFinder)

	list.Put(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1})
	list.Put(core.QSO{Callsign: callsign.MustParse("K3LR"), MyNumber: 3})
	list.Put(core.QSO{Callsign: callsign.MustParse("DK9ZZ"), MyNumber: 2})

	qsos := list.All()

	assert.Equal(t, dlPrefix, qsos[0].DXCC, qsos[0])
	assert.Equal(t, dlPrefix, qsos[1].DXCC, qsos[1])
	assert.Equal(t, dxcc.Prefix{}, qsos[2].DXCC, qsos[2])
}

func TestPut_UpdatePrefix(t *testing.T) {
	dlPrefix := dxcc.Prefix{Name: "Fed. Rep. of Germany", PrimaryPrefix: "DL", Continent: "EU", CQZone: 14, ITUZone: 28}
	kPrefix := dxcc.Prefix{Name: "United States", PrimaryPrefix: "K", Continent: "NA", CQZone: 5, ITUZone: 8}
	dxccFinder := new(mocked.DXCCFinder)
	dxccFinder.On("Find", "DL1ABC").Return(dlPrefix, true)
	dxccFinder.On("Find", "K3LR").Return(kPrefix, true)
	list := NewQSOList(dxccFinder)

	list.Put(core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1})
	list.Put(core.QSO{Callsign: callsign.MustParse("K3LR"), MyNumber: 1})

	qsos := list.All()

	assert.Equal(t, kPrefix, qsos[0].DXCC, qsos[0])
}

func TestSelectQSO(t *testing.T) {
	qso := core.QSO{Callsign: callsign.MustParse("DL1ABC"), MyNumber: 1}
	list := NewQSOList(new(nullDXCCFinder))
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
	list := NewQSOList(new(nullDXCCFinder))
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
	list := NewQSOList(new(nullDXCCFinder))
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
	list := NewQSOList(new(nullDXCCFinder))
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
	list := NewQSOList(new(nullDXCCFinder))
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
	list := NewQSOList(new(nullDXCCFinder))
	qso := core.QSO{MyNumber: 1}
	notified := false
	list.Notify(QSOAddedListenerFunc(func(addedQSO core.QSO) {
		notified = true
		assert.Equal(t, qso, addedQSO)
	}))

	list.Put(qso)

	assert.True(t, notified)
}

func TestQSOInsertedListener(t *testing.T) {
	list := toQSOList(1, 3)
	qso := core.QSO{MyNumber: 2}
	notified := false
	list.Notify(QSOInsertedListenerFunc(func(index int, insertedQSO core.QSO) {
		notified = true
		assert.Equal(t, 1, index)
		assert.Equal(t, qso, insertedQSO)
	}))

	list.Put(qso)

	assert.True(t, notified)
}

func TestQSOUpdatedListener(t *testing.T) {
	list := toQSOList(1, 2, 3)
	qso := core.QSO{MyNumber: 2, TheirNumber: 100}
	notified := false
	list.Notify(QSOUpdatedListenerFunc(func(index int, oldQSO, updatedQSO core.QSO) {
		notified = true
		assert.Equal(t, 1, index)
		assert.Equal(t, core.QSONumber(0), oldQSO.TheirNumber)
		assert.Equal(t, core.QSONumber(100), updatedQSO.TheirNumber)
	}))

	list.Put(qso)

	assert.True(t, notified)
}

func toQSOList(numbers ...int) *QSOList {
	qsos := make([]core.QSO, len(numbers))
	for i, number := range numbers {
		qsos[i] = core.QSO{MyNumber: core.QSONumber(number)}
	}
	return &QSOList{list: qsos, dxccFinder: new(nullDXCCFinder)}
}

func toQSOs(numbers ...int) []core.QSO {
	result := make([]core.QSO, len(numbers))
	for i, number := range numbers {
		result[i] = core.QSO{MyNumber: core.QSONumber(number)}
	}
	return result
}
