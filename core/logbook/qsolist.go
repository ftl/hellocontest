package logbook

import (
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hellocontest/core"
)

type QSOAddedListener interface {
	QSOAdded(core.QSO)
}

type QSOAddedListenerFunc func(core.QSO)

func (f QSOAddedListenerFunc) QSOAdded(qso core.QSO) {
	f(qso)
}

type QSOInsertedListener interface {
	QSOInserted(int, core.QSO)
}

type QSOInsertedListenerFunc func(int, core.QSO)

func (f QSOInsertedListenerFunc) QSOInserted(index int, qso core.QSO) {
	f(index, qso)
}

type QSOUpdatedListener interface {
	QSOUpdated(int, core.QSO, core.QSO)
}

type QSOUpdatedListenerFunc func(int, core.QSO, core.QSO)

func (f QSOUpdatedListenerFunc) QSOUpdated(index int, old, new core.QSO) {
	f(index, old, new)
}

// DXCCFinder returns a list of matching prefixes for the given string and indicates if there was a match at all.
type DXCCFinder interface {
	Find(string) ([]dxcc.Prefix, bool)
}

type QSOList struct {
	list       []core.QSO
	listeners  []interface{}
	dxccFinder DXCCFinder
}

func NewQSOList(dxccFinder DXCCFinder) *QSOList {
	return &QSOList{
		list:       make([]core.QSO, 0),
		dxccFinder: dxccFinder,
	}
}

func (l *QSOList) Put(qso core.QSO) {
	if len(l.list) == 0 {
		l.append(qso)
		return
	}
	lastNumber := l.list[len(l.list)-1].MyNumber
	if qso.MyNumber > lastNumber {
		l.append(qso)
		return
	}
	index, found := l.findIndex(qso.MyNumber)
	if !found {
		l.insert(index, qso)
		return
	}
	l.update(index, qso)
}

func (l *QSOList) findIndex(myNumber core.QSONumber) (int, bool) {
	low := 0
	high := len(l.list) - 1

	for low <= high {
		median := (low + high) / 2

		if l.list[median].MyNumber < myNumber {
			low = median + 1
		} else {
			high = median - 1
		}
	}

	if low == len(l.list) || l.list[low].MyNumber != myNumber {
		return low, false
	}

	return low, true
}

func (l *QSOList) append(qso core.QSO) {
	l.setDXCC(&qso)
	l.list = append(l.list, qso)
	l.emitQSOAdded(qso)
}

func (l *QSOList) setDXCC(qso *core.QSO) {
	prefixes, found := l.dxccFinder.Find(qso.Callsign.String())
	if found {
		qso.DXCC = prefixes[0]
	}
}

func (l *QSOList) insert(index int, qso core.QSO) {
	l.setDXCC(&qso)
	l.list = append(l.list[:index+1], l.list[index:]...)
	l.list[index] = qso
	l.emitQSOInserted(index, qso)
}

func (l *QSOList) update(index int, qso core.QSO) {
	l.setDXCC(&qso)
	old := l.list[index]
	l.list[index] = qso
	l.emitQSOUpdated(index, old, qso)
}

func (l *QSOList) All() []core.QSO {
	return l.list
}

func (l *QSOList) Notify(listener interface{}) {
	l.listeners = append(l.listeners, listener)
}

func (l *QSOList) emitQSOAdded(qso core.QSO) {
	for _, listener := range l.listeners {
		if qsoAddedListener, ok := listener.(QSOAddedListener); ok {
			qsoAddedListener.QSOAdded(qso)
		}
	}
}

func (l *QSOList) emitQSOInserted(index int, qso core.QSO) {
	for _, listener := range l.listeners {
		if qsoInsertedListener, ok := listener.(QSOInsertedListener); ok {
			qsoInsertedListener.QSOInserted(index, qso)
		}
	}
}

func (l *QSOList) emitQSOUpdated(index int, old, new core.QSO) {
	for _, listener := range l.listeners {
		if qsoUpdatedListener, ok := listener.(QSOUpdatedListener); ok {
			qsoUpdatedListener.QSOUpdated(index, old, new)
		}
	}
}

type nullDXCCFinder struct{}

func (f *nullDXCCFinder) Find(string) ([]dxcc.Prefix, bool) {
	return []dxcc.Prefix{}, false
}
