package logbook

import (
	"log"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

type QSOsClearedListener interface {
	QSOsCleared()
}

type QSOsClearedListenerFunc func()

func (f QSOsClearedListenerFunc) QSOsCleared() {
	f()
}

type QSOFiller interface {
	FillQSO(*core.QSO)
}

type QSOFillerFunc func(*core.QSO)

func (f QSOFillerFunc) FillQSO(qso *core.QSO) {
	f(qso)
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

type QSOList struct {
	allowMultiBand bool
	allowMultiMode bool
	list           []core.QSO
	dupes          dupeIndex
	worked         dupeIndex
	invalid        bool

	listeners []interface{}
}

func NewQSOList(settings core.Settings) *QSOList {
	return &QSOList{
		allowMultiBand: settings.Contest().AllowMultiBand,
		allowMultiMode: settings.Contest().AllowMultiMode,
		list:           make([]core.QSO, 0),
		dupes:          make(dupeIndex),
		worked:         make(dupeIndex),
	}
}

func (l *QSOList) ContestChanged(contest core.Contest) {
	if l.allowMultiBand == contest.AllowMultiBand && l.allowMultiMode == contest.AllowMultiMode {
		return
	}
	l.allowMultiBand = contest.AllowMultiBand
	l.allowMultiMode = contest.AllowMultiMode
	l.invalid = true
}

func (l *QSOList) Valid() bool {
	return !l.invalid
}

func (l *QSOList) Clear() {
	l.list = make([]core.QSO, 0)
	l.dupes = make(dupeIndex)
	l.worked = make(dupeIndex)
	l.invalid = false
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

func (l *QSOList) findIndex(number core.QSONumber) (int, bool) {
	return findIndex(l.list, number)
}

func findIndex(list []core.QSO, number core.QSONumber) (int, bool) {
	low := 0
	high := len(list) - 1

	for low <= high {
		median := (low + high) / 2

		if list[median].MyNumber < number {
			low = median + 1
		} else {
			high = median - 1
		}
	}

	if low == len(list) || list[low].MyNumber != number {
		return low, false
	}

	return low, true
}

func (l *QSOList) append(qso core.QSO) {
	dupeBand, dupeMode := l.dupeBandAndMode(qso.Band, qso.Mode)
	l.dupes.Add(qso.Callsign, dupeBand, dupeMode, qso.MyNumber)
	dupes := l.dupes.Get(qso.Callsign, dupeBand, dupeMode)
	qso.Duplicate = len(dupes) > 1

	l.worked.Add(qso.Callsign, core.NoBand, core.NoMode, qso.MyNumber)

	l.fillQSO(&qso)
	l.list = append(l.list, qso)
	l.emitQSOAdded(qso)
}

func (l *QSOList) insert(index int, qso core.QSO) {
	dupeBand, dupeMode := l.dupeBandAndMode(qso.Band, qso.Mode)
	l.dupes.Add(qso.Callsign, dupeBand, dupeMode, qso.MyNumber)
	dupes := l.dupes.Get(qso.Callsign, dupeBand, dupeMode)

	l.worked.Add(qso.Callsign, core.NoBand, core.NoMode, qso.MyNumber)

	l.fillQSO(&qso)
	l.list = append(l.list[:index+1], l.list[index:]...)
	l.list[index] = qso
	updates := l.updateDuplicateMarkers(dupes)
	l.emitQSOInserted(index, qso)
	for _, update := range updates {
		l.emitQSOUpdated(update.index, update.old, update.new)
	}
}

func (l *QSOList) update(index int, qso core.QSO) {
	old := l.list[index]
	oldDupeBand, oldDupeMode := l.dupeBandAndMode(old.Band, old.Mode)
	l.dupes.Remove(old.Callsign, oldDupeBand, oldDupeMode, old.MyNumber)
	oldDupes := l.dupes.Get(old.Callsign, oldDupeBand, oldDupeMode)
	updates := l.updateDuplicateMarkers(oldDupes)

	dupeBand, dupeMode := l.dupeBandAndMode(qso.Band, qso.Mode)
	l.dupes.Add(qso.Callsign, dupeBand, dupeMode, qso.MyNumber)
	dupes := l.dupes.Get(qso.Callsign, dupeBand, dupeMode)
	qso.Duplicate = len(dupes) > 1

	l.worked.Remove(old.Callsign, core.NoBand, core.NoMode, old.MyNumber)
	l.worked.Add(qso.Callsign, core.NoBand, core.NoMode, qso.MyNumber)

	l.fillQSO(&qso)
	l.list[index] = qso
	updates = append(updates, l.updateDuplicateMarkers(dupes)...)
	l.emitQSOUpdated(index, old, qso)
	for _, update := range updates {
		l.emitQSOUpdated(update.index, update.old, update.new)
	}
}

func (l *QSOList) dupeBandAndMode(band core.Band, mode core.Mode) (core.Band, core.Mode) {
	if !l.allowMultiBand {
		band = core.NoBand
	}
	if !l.allowMultiMode {
		mode = core.NoMode
	}
	return band, mode
}

type qsoUpdate struct {
	index    int
	old, new core.QSO
}

func (l *QSOList) updateDuplicateMarkers(numbers []core.QSONumber) []qsoUpdate {
	result := make([]qsoUpdate, 0, len(numbers))
	if len(numbers) == 0 {
		return result
	}

	first := numbers[0]
	firstIndex := 0
	for i, n := range numbers {
		if n < first {
			first = n
			firstIndex = i
		}
	}
	numbers[len(numbers)-1], numbers[firstIndex] = numbers[firstIndex], numbers[len(numbers)-1]
	numbers = numbers[:len(numbers)-1]

	index, found := l.findIndex(first)
	if found {
		qso := l.list[index]
		if qso.Duplicate {
			update := qsoUpdate{index: index, old: qso}

			qso.Duplicate = false
			l.list[index] = qso

			update.new = qso
			result = append(result, update)
		}
	} else {
		log.Printf("UpdateDuplicateMarkers: cannot find index for FIRST QSO %d", first)
	}

	for _, n := range numbers {
		index, found := l.findIndex(n)
		if found {
			qso := l.list[index]
			if !qso.Duplicate {
				update := qsoUpdate{index: index, old: qso}

				qso.Duplicate = true
				l.list[index] = qso

				update.new = qso
				result = append(result, update)
			}
		} else {
			log.Printf("UpdateDuplicateMarkers: cannot find index for QSO %d", n)
		}
	}
	return result
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

func (l *QSOList) FindDuplicateQSOs(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	band, mode = l.dupeBandAndMode(band, mode)
	numbers := l.dupes.Get(callsign, band, mode)
	return l.GetQSOs(numbers)
}

func (l *QSOList) GetQSOs(numbers []core.QSONumber) []core.QSO {
	result := make([]core.QSO, 0, len(numbers))
	for _, n := range numbers {
		listIndex, found := l.findIndex(n)
		if !found {
			log.Printf("QSO number %d not found", n)
			continue
		}
		qso := l.list[listIndex]
		if len(result) > 0 && n > result[len(result)-1].MyNumber {
			result = append(result, qso)
		} else {
			resultIndex, found := findIndex(result, n)
			if !found {
				result = append(result[:resultIndex+1], result[resultIndex:]...)
			}
			result[resultIndex] = qso
		}
	}
	return result
}

func (l *QSOList) FindWorkedQSOs(callsign callsign.Callsign, band core.Band, mode core.Mode) ([]core.QSO, bool) {
	numbers := l.worked.Get(callsign, core.NoBand, core.NoMode)
	qsos := l.GetQSOs(numbers)
	if len(qsos) == 0 {
		return qsos, false
	}

	duplicate := false
	for _, qso := range qsos {
		switch {
		case qso.Band == band:
			duplicate = qso.Mode == mode || !l.allowMultiMode
		case qso.Mode == mode:
			duplicate = qso.Band == band || !l.allowMultiBand
		}
		if duplicate {
			break
		}
	}
	return qsos, duplicate
}

func (l *QSOList) Notify(listener interface{}) {
	l.listeners = append(l.listeners, listener)
}

func (l *QSOList) fillQSO(qso *core.QSO) {
	for _, listener := range l.listeners {
		if qsoFiller, ok := listener.(QSOFiller); ok {
			qsoFiller.FillQSO(qso)
		}
	}
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
