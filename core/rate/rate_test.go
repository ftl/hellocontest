package rate

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestHourOf(t *testing.T) {
	assert.Equal(t, core.Hour(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)), core.HourOf(time.Date(2009, time.November, 10, 23, 24, 25, 26, time.UTC)))
	assert.Equal(t, core.Hour(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.Local)), core.HourOf(time.Date(2009, time.November, 10, 23, 24, 25, 26, time.Local)))
}

func TestNewQSOListIsEmpty(t *testing.T) {
	fifo := new(qsoList)
	assert.True(t, fifo.Empty())
	assert.Equal(t, 0, fifo.Length())
}

func TestQSOList_AddFirstQSO(t *testing.T) {
	qso := core.QSO{Time: time.Now()}
	list := new(qsoList)

	list.Add(qso)

	assert.Same(t, list.first, list.last)
	assert.NotNil(t, list.first)
	assert.Nil(t, list.first.Previous)
	assert.Nil(t, list.first.Next)
	assert.False(t, list.Empty())
	assert.Equal(t, 1, list.Length())
}

func TestQSOList_AddTwoQSOs(t *testing.T) {
	now := time.Now()
	qso1 := core.QSO{Time: now.Add(1 * time.Minute)}
	qso2 := core.QSO{Time: now}
	list := new(qsoList)

	list.Add(qso1)
	list.Add(qso2)

	assert.NotSame(t, list.first, list.last)
	assert.NotNil(t, list.first)
	assert.NotNil(t, list.last)
	assert.Nil(t, list.first.Previous)
	assert.NotNil(t, list.first.Next)
	assert.NotNil(t, list.last.Previous)
	assert.Nil(t, list.last.Next)
	assert.False(t, list.Empty())
	assert.Equal(t, 2, list.Length())

	assert.Equal(t, qso2, list.first.QSO)
	assert.Equal(t, qso1, list.last.QSO)
}

func TestQSOList_OrderByQSOTime(t *testing.T) {
	now := time.Now()
	list := new(qsoList)
	for i := 0; i < 10; i++ {
		offset := 0
		if i%2 == 0 {
			offset = 10
		}
		list.Add(core.QSO{MyNumber: core.QSONumber(i + 1), Time: now.Add(time.Duration(-(10-i)-offset) * time.Minute)})
	}

	for e := list.first; e != nil; e = e.Next {
		assert.Truef(t, e.Previous == nil || !e.Previous.QSO.Time.After(e.QSO.Time), "%d: %s", e.QSO.MyNumber, e.QSO.Time)
	}
}

func TestQSOList_RemoveBefore(t *testing.T) {
	now := time.Now()
	list := new(qsoList)
	for i := 0; i < 10; i++ {
		offset := 0
		if i%2 == 0 {
			offset = 10
		}
		list.Add(core.QSO{MyNumber: core.QSONumber(i + 1), Time: now.Add(time.Duration(-(10-i)-offset) * time.Minute)})
	}

	endOfLife := now.Add(-10 * time.Minute)
	list.RemoveBefore(endOfLife)

	assert.Equal(t, 5, list.Length())
	for e := list.first; e != nil; e = e.Next {
		assert.True(t, e.QSO.Time.After(endOfLife))
	}
}

func TestQSOList_RemoveQSO(t *testing.T) {
	now := time.Now()
	list := new(qsoList)
	for i := 0; i < 10; i++ {
		list.Add(core.QSO{MyNumber: core.QSONumber(i + 1 + (i % 2)), Time: now.Add(time.Duration(-(10 - i)) * time.Minute)})
	}

	list.RemoveQSO(core.QSO{MyNumber: 3})
	list.RemoveQSO(core.QSO{MyNumber: 5})

	assert.Equal(t, 6, list.Length())
	for e := list.first; e != nil; e = e.Next {
		assert.NotEqual(t, 3, e.QSO.MyNumber)
		assert.NotEqual(t, 5, e.QSO.MyNumber)
	}
}

func TestQSOList_LengthAfter(t *testing.T) {
	now := time.Now()
	list := new(qsoList)
	for i := 0; i < 10; i++ {
		list.Add(core.QSO{MyNumber: core.QSONumber(i + 1 + (i % 2)), Time: now.Add(time.Duration(-(10 - i)) * time.Minute)})
	}

	assert.Equal(t, 4, list.LengthAfter(now.Add(-5*time.Minute)))
}

func printList(list *qsoList) {
	list.forward(func(e *qsoListEntry) bool {
		var n core.QSONumber
		var p core.QSONumber
		if e.Next != nil {
			n = e.Next.QSO.MyNumber
		}
		if e.Previous != nil {
			p = e.Previous.QSO.MyNumber
		}
		log.Printf("entry %d (%d | %d): %s", e.QSO.MyNumber, p, n, e.QSO.Time)
		return true
	})
}
