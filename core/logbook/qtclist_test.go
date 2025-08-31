package logbook

import (
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestQTCList_basicSetup(t *testing.T) {
	l := NewQTCList()
	now := time.Now()
	listener := new(qtcListListener)
	l.Notify(listener)

	initialQTCs := []core.QTC{
		{Timestamp: now.Add(-4 * time.Minute), QSONumber: 1},
		{Timestamp: now.Add(-2 * time.Minute), QSONumber: 2},
		{Timestamp: now.Add(-1 * time.Minute), QSONumber: 3},
	}
	l.Fill(initialQTCs)
	assert.Equal(t, initialQTCs, l.All())
	assert.Equal(t, 1, listener.clearEvents)
	assert.Equal(t, 3, listener.addedEvents)

	newQTC := core.QTC{Timestamp: now, QSONumber: 4}
	l.QTCAdded(newQTC)
	assert.Equal(t, 4, len(l.All()))
	assert.Equal(t, newQTC, l.All()[3])
	assert.Equal(t, 1, listener.clearEvents)
	assert.Equal(t, 4, listener.addedEvents)

	l.Clear()
	assert.Equal(t, 0, len(l.All()))
	assert.Equal(t, 2, listener.clearEvents)
	assert.Equal(t, 4, listener.addedEvents)
}

func TestQTCList_AvailableFor(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	dl2abc := callsign.MustParse("DL2ABC")

	l := NewQTCList()
	assert.Equal(t, 0, l.AvailableFor(dl1abc))
	l.QSOAdded(core.QSO{Callsign: dl1abc, MyNumber: 1})
	assert.Equal(t, 0, l.AvailableFor(dl1abc))

	for i := range 20 {
		l.QSOAdded(core.QSO{Callsign: dl2abc, MyNumber: core.QSONumber(i + 2)})
		assert.Equal(t, min(i+1, core.MaxQTCsPerCall), l.AvailableFor(dl1abc))
	}

	for i := range core.MaxQTCsPerCall {
		l.QTCAdded(core.QTC{TheirCallsign: dl1abc, QSONumber: core.QSONumber(i + 2), Kind: core.SentQTC})
		assert.Equal(t, core.MaxQTCsPerCall-i-1, l.AvailableFor(dl1abc))
	}
}

// helpers

type qtcListListener struct {
	clearEvents int
	addedEvents int
}

func (l *qtcListListener) QTCsCleared() {
	l.clearEvents++
}

func (l *qtcListListener) QTCAdded(core.QTC) {
	l.addedEvents++
}
