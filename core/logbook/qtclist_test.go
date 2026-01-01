package logbook

import (
	"testing"
	"time"

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
	l.LogbookCleared()
	for _, qtc := range initialQTCs {
		l.QTCAdded(qtc)
	}
	assert.Equal(t, initialQTCs, l.qtcs)
	assert.Equal(t, 1, listener.clearEvents)
	assert.Equal(t, 3, listener.addedEvents)

	newQTC := core.QTC{Timestamp: now, QSONumber: 4}
	l.QTCAdded(newQTC)
	assert.Equal(t, 4, len(l.qtcs))
	assert.Equal(t, newQTC, l.qtcs[3])
	assert.Equal(t, 1, listener.clearEvents)
	assert.Equal(t, 4, listener.addedEvents)

	l.LogbookCleared()
	assert.Equal(t, 0, len(l.qtcs))
	assert.Equal(t, 2, listener.clearEvents)
	assert.Equal(t, 4, listener.addedEvents)
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
