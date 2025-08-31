package logbook

import (
	"slices"
	"sync"

	"github.com/ftl/hellocontest/core"
)

type QTCsClearedListener interface {
	QTCsCleared()
}

type QTCList struct {
	dataLock *sync.RWMutex
	data     map[core.QSONumber]core.QTC

	listeners []any
}

func NewQTCList() *QTCList {
	return &QTCList{
		dataLock: new(sync.RWMutex),
		data:     make(map[core.QSONumber]core.QTC),
	}
}

func (l *QTCList) Notify(listener any) {
	l.listeners = append(l.listeners, listener)
}

func (l *QTCList) emitQTCsCleared() {
	for _, lis := range l.listeners {
		if listener, ok := lis.(QTCsClearedListener); ok {
			listener.QTCsCleared()
		}
	}
}

func (l *QTCList) emitQTCAdded(qtc core.QTC) {
	for _, lis := range l.listeners {
		if listener, ok := lis.(QTCAddedListener); ok {
			listener.QTCAdded(qtc)
		}
	}
}

func (l *QTCList) Clear() {
	l.dataLock.Lock()
	l.clear()
	l.dataLock.Unlock()

	l.emitQTCsCleared()
}

func (l *QTCList) clear() {
	l.data = make(map[core.QSONumber]core.QTC)
}

func (l *QTCList) Fill(qtcs []core.QTC) {
	l.dataLock.Lock()
	if len(l.data) > 0 {
		l.clear()
	}
	for _, qtc := range qtcs {
		l.put(qtc)
	}
	allQTCs := l.all()
	l.dataLock.Unlock()

	l.emitQTCsCleared()
	for _, qtc := range allQTCs {
		l.emitQTCAdded(qtc)
	}
}

func (l *QTCList) QTCAdded(qtc core.QTC) {
	l.dataLock.Lock()
	l.put(qtc)
	l.dataLock.Unlock()
	l.emitQTCAdded(qtc)
}

func (l *QTCList) put(qtc core.QTC) {
	l.data[qtc.QSONumber] = qtc
}

func (l *QTCList) All() []core.QTC {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.all()
}

func (l *QTCList) all() []core.QTC {
	result := make([]core.QTC, 0, len(l.data))
	for _, qtc := range l.data {
		result = append(result, qtc)
	}

	slices.SortStableFunc(result, core.QTCByTimestamp)
	return result
}

func (l *QTCList) QSOAdded(qso core.QSO) {
	// TODO: add an available QTC for the new QSO if appropriate
}
