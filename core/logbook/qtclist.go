package logbook

import (
	"fmt"
	"slices"
	"sync"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

type QTCsClearedListener interface {
	QTCsCleared()
}

type QTCList struct {
	dataLock      *sync.RWMutex
	data          map[core.QSONumber]core.QTC
	availableQTCs []core.QTC
	qtcsByCall    map[callsign.Callsign]int

	listeners []any
}

func NewQTCList() *QTCList {
	return &QTCList{
		dataLock:      new(sync.RWMutex),
		data:          make(map[core.QSONumber]core.QTC),
		availableQTCs: []core.QTC{},
		qtcsByCall:    make(map[callsign.Callsign]int),
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
	l.availableQTCs = []core.QTC{}
	l.qtcsByCall = make(map[callsign.Callsign]int)
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
	result := make([]core.QTC, 0, len(l.data))
	for _, qtc := range l.data {
		result = append(result, qtc)
	}

	slices.SortStableFunc(result, core.QTCByTimestamp)
	return result
}

func (l *QTCList) QSOAdded(qso core.QSO) {
	qtc := qtcFromQSO(qso)

	l.dataLock.Lock()
	l.availableQTCs = append(l.availableQTCs, qtc)
	l.dataLock.Unlock()
}

func qtcFromQSO(qso core.QSO) core.QTC {
	return core.QTC{
		Frequency:   qso.Frequency,
		Band:        qso.Band,
		Mode:        qso.Mode,
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

func (l *QTCList) PrepareFor(theirCall callsign.Callsign, count int) (core.QTCSeries, error) {
	return core.QTCSeries{}, fmt.Errorf("not yet implemented")
}
