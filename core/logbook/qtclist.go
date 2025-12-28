package logbook

import (
	"log"
	"slices"
	"sync"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

type QTCAddedListener interface {
	QTCAdded(core.QTC)
}

type QTCAddedListenerFunc func(core.QTC)

func (f QTCAddedListenerFunc) QTCAdded(qtc core.QTC) {
	f(qtc)
}

type QTCsEnabledListener interface {
	SetQTCsEnabled(bool)
}

type QTCSelectedListener interface {
	QTCSelected(core.QTC)
}

type QTCSelectedListenerFunc func(core.QTC)

func (f QTCSelectedListenerFunc) QTCSelected(qtc core.QTC) {
	f(qtc)
}

type QTCRowSelectedListener interface {
	QTCRowSelected(int)
}

type QTCRowSelectedListenerFunc func(int)

func (f QTCRowSelectedListenerFunc) QTCRowSelected(index int) {
	f(index)
}

type QTCsClearedListener interface {
	QTCsCleared()
}

type QTCList struct {
	dataLock      *sync.RWMutex
	enabled       bool
	data          []core.QTC
	availableQTCs []core.QTC
	qtcsByCall    map[callsign.Callsign]int

	listeners []any
}

func NewQTCList() *QTCList {
	return &QTCList{
		dataLock:      new(sync.RWMutex),
		data:          make([]core.QTC, 0),
		availableQTCs: make([]core.QTC, 0),
		qtcsByCall:    make(map[callsign.Callsign]int),
	}
}

func (l *QTCList) Notify(listener any) {
	l.listeners = append(l.listeners, listener)
}

func (l *QTCList) emitQTCsEnabled(value bool) {
	for _, lis := range l.listeners {
		if listener, ok := lis.(QTCsEnabledListener); ok {
			listener.SetQTCsEnabled(value)
		}
	}
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

func (l *QTCList) emitQTCSelected(qtc core.QTC) {
	for _, listener := range l.listeners {
		if qsoSelectedListener, ok := listener.(QTCSelectedListener); ok {
			qsoSelectedListener.QTCSelected(qtc)
		}
	}
}

func (l *QTCList) emitQTCRowSelected(index int) {
	for _, listener := range l.listeners {
		if rowSelectedListener, ok := listener.(QTCRowSelectedListener); ok {
			rowSelectedListener.QTCRowSelected(index)
		}
	}
}

func (l *QTCList) SetQTCsEnabled(enabled bool) {
	l.dataLock.Lock()
	l.enabled = enabled
	l.dataLock.Unlock()

	l.emitQTCsEnabled(enabled)
}

func (l *QTCList) QTCsEnabled() bool {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()
	return l.enabled
}

func (l *QTCList) Clear() {
	l.dataLock.Lock()
	l.clear()
	l.dataLock.Unlock()

	l.emitQTCsCleared()
}

func (l *QTCList) clear() {
	l.data = make([]core.QTC, 0)
	l.availableQTCs = make([]core.QTC, 0)
	l.qtcsByCall = make(map[callsign.Callsign]int)
}

func (l *QTCList) Fill(qsos []core.QSO, qtcs []core.QTC) {
	l.dataLock.Lock()
	l.data = make([]core.QTC, 0, len(qtcs))
	l.availableQTCs = make([]core.QTC, 0, len(qsos))
	l.qtcsByCall = make(map[callsign.Callsign]int)

	for _, qso := range qsos {
		l.putQSO(qso)
	}
	for _, qtc := range qtcs {
		l.putQTC(qtc)
	}
	allQTCs := l.all()
	l.dataLock.Unlock()

	l.emitQTCsCleared()
	for _, qtc := range allQTCs {
		l.emitQTCAdded(qtc)
	}
}

func (l *QTCList) PutQTC(qtc core.QTC) {
	l.dataLock.Lock()
	l.putQTC(qtc)
	l.dataLock.Unlock()
	l.emitQTCAdded(qtc)
}

func (l *QTCList) putQTC(qtc core.QTC) {
	l.data = append(l.data, qtc)
	l.removeAvailable(qtc)
	count := l.qtcsByCall[qtc.TheirCallsign]
	count++
	l.qtcsByCall[qtc.TheirCallsign] = count
}

func (l *QTCList) removeAvailable(qtc core.QTC) {
	if qtc.Kind != core.SentQTC {
		return
	}
	index := -1
	for i := range l.availableQTCs {
		if qtc.QSONumber == l.availableQTCs[i].QSONumber {
			index = i
			break
		}
	}
	switch {
	case index < 0:
	case index < len(l.availableQTCs)-1:
		l.availableQTCs = append(l.availableQTCs[:index], l.availableQTCs[index+1:]...)
	case index == len(l.availableQTCs)-1:
		l.availableQTCs = l.availableQTCs[:index]
	}
}

func (l *QTCList) All() []core.QTC {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.all()
}

func (l *QTCList) all() []core.QTC {
	result := make([]core.QTC, len(l.data))
	copy(result, l.data)

	slices.SortStableFunc(result, core.QTCByTimestamp)
	return result
}

func (l *QTCList) SelectRow(index int) {
	l.dataLock.RLock()

	if index < 0 || index >= len(l.data) {
		log.Printf("invalid QSO index %d", index)
		l.dataLock.RUnlock()
		return
	}
	qtc := l.data[index]

	l.dataLock.RUnlock()

	l.emitQTCSelected(qtc)
	l.emitQTCRowSelected(index)
}

func (l *QTCList) SelectLastQTC() {
	l.dataLock.RLock()

	if len(l.data) == 0 {
		l.dataLock.RUnlock()
		return
	}

	index := len(l.data) - 1
	qtc := l.data[index]

	l.dataLock.RUnlock()

	l.emitQTCSelected(qtc)
	l.emitQTCRowSelected(index)
}

func (l *QTCList) PutQSO(qso core.QSO) {
	l.dataLock.Lock()
	l.putQSO(qso)
	l.dataLock.Unlock()
}

func (l *QTCList) putQSO(qso core.QSO) {
	qtc := qtcFromQSO(qso)
	for i, availableQTC := range l.availableQTCs {
		if availableQTC.QSONumber == qso.MyNumber {
			l.availableQTCs[i] = qtc
			return
		}
	}
	l.availableQTCs = append(l.availableQTCs, qtc)
}

func qtcFromQSO(qso core.QSO) core.QTC {
	return core.QTC{
		Kind:        core.SentQTC,
		QSONumber:   qso.MyNumber,
		QTCTime:     core.QTCTimeFromTimestamp(qso.Time),
		QTCCallsign: qso.Callsign,
		QTCNumber:   qso.TheirNumber,
	}
}

// AvailableFor returns the number of QTCs available for the given callsign.
func (l *QTCList) AvailableFor(theirCall callsign.Callsign) int {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	theirCallStr := theirCall.String()
	theirQTCCount := l.qtcsByCall[theirCall]
	theirQSOCount := 0
	for _, qtc := range l.availableQTCs {
		if qtc.QTCCallsign.String() == theirCallStr {
			theirQSOCount++
		}
	}
	return min(core.MaxQTCsPerCall-theirQTCCount, len(l.availableQTCs)-theirQSOCount)
}

func (l *QTCList) PrepareFor(theirCall callsign.Callsign, count int) []core.QTC {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	theirCallStr := theirCall.String()
	theirQTCCount := l.qtcsByCall[theirCall]
	maxLen := max(0, min(core.MaxQTCsPerCall-theirQTCCount, count))

	result := make([]core.QTC, 0, maxLen)
	for _, qtc := range l.availableQTCs {
		if len(result) >= maxLen {
			break
		}
		if qtc.QTCCallsign.String() != theirCallStr {
			qtc.TheirCallsign = theirCall
			result = append(result, qtc)
		}
	}

	return result
}
