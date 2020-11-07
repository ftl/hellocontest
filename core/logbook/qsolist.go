package logbook

import (
	"log"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

type QSOsClearedListener interface {
	QSOsCleared()
}

type QSOsClearedListenerFunc func()

func (f QSOsClearedListenerFunc) QSOsCleared() {
	f()
}

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

type QSOSelectedListener interface {
	QSOSelected(core.QSO)
}

type QSOSelectedListenerFunc func(core.QSO)

func (f QSOSelectedListenerFunc) QSOSelected(qso core.QSO) {
	f(qso)
}

type RowSelectedListener interface {
	RowSelected(int)
}

type RowSelectedListenerFunc func(int)

func (f RowSelectedListenerFunc) RowSelected(index int) {
	f(index)
}

// DXCCFinder returns a list of matching prefixes for the given string and indicates if there was a match at all.
type DXCCFinder interface {
	Find(string) (dxcc.Prefix, bool)
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

func (l *QSOList) Clear() {
	l.list = make([]core.QSO, 0)
	l.emitQSOsCleared()
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
	if prefix, found := l.dxccFinder.Find(qso.Callsign.String()); found {
		qso.DXCC = prefix
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

func (l *QSOList) SelectRow(index int) {
	if index < 0 || index >= len(l.list) {
		log.Printf("invalid QSO index %d", index)
		return
	}

	qso := l.list[index]
	l.emitQSOSelected(qso)
	l.emitRowSelected(index)
}

func (l *QSOList) SelectQSO(qso core.QSO) {
	index, ok := l.findIndex(qso.MyNumber)
	if !ok {
		log.Print("qso not found")
		return
	}

	l.emitQSOSelected(qso)
	l.emitRowSelected(index)
}

func (l *QSOList) SelectLastQSO() {
	if len(l.list) == 0 {
		return
	}

	index := len(l.list) - 1
	qso := l.list[index]
	l.emitQSOSelected(qso)
	l.emitRowSelected(index)
}

func (l *QSOList) Find(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	result := make([]core.QSO, 0)
	for _, qso := range l.list {
		if callsign != qso.Callsign {
			continue
		}
		if band != core.NoBand && band != qso.Band {
			continue
		}
		if mode != core.NoMode && mode != qso.Mode {
			continue
		}
		result = append(result, qso)
	}
	return result
}

func (l *QSOList) ForEach(f func(qso *core.QSO)) {
	for i, qso := range l.list {
		f(&qso)
		l.list[i] = qso
	}
}

func (l *QSOList) Notify(listener interface{}) {
	l.listeners = append(l.listeners, listener)
}

func (l *QSOList) emitQSOsCleared() {
	for _, listener := range l.listeners {
		if qsosClearedListener, ok := listener.(QSOsClearedListener); ok {
			qsosClearedListener.QSOsCleared()
		}
	}
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

func (l *QSOList) emitQSOSelected(qso core.QSO) {
	for _, listener := range l.listeners {
		if qsoSelectedListener, ok := listener.(QSOSelectedListener); ok {
			qsoSelectedListener.QSOSelected(qso)
		}
	}
}

func (l *QSOList) emitRowSelected(index int) {
	for _, listener := range l.listeners {
		if rowSelectedListener, ok := listener.(RowSelectedListener); ok {
			rowSelectedListener.RowSelected(index)
		}
	}
}

type nullDXCCFinder struct{}

func (f *nullDXCCFinder) Find(string) (dxcc.Prefix, bool) {
	return dxcc.Prefix{}, false
}
