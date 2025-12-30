package logbook

import (
	"fmt"
	"sync"

	"github.com/ftl/hellocontest/core"
)

type Logbook struct {
	clock core.Clock

	listeners []any

	dataLock          *sync.RWMutex
	qsos              []core.QSO
	sentQTCs          map[core.QSONumber]core.QTC // TODO: better store the QTC series?
	receivedQTCs      []core.QTC                  // TODO: better store QTC series?
	sentQTCsPerSeries []int
}

func NewLogbook(clock core.Clock) *Logbook {
	result := &Logbook{
		clock:    clock,
		dataLock: new(sync.RWMutex),
	}

	result.clear(1000, 0)

	return result
}

// Loading

func (l *Logbook) Load(qsos []core.QSO, qtcs []core.QTC) error {
	loadedQSOs, loadedQTCs, err := l.loadLocked(qsos, qtcs)
	if err != nil {
		return err
	}

	l.emitNotificationsAfterRefresh(loadedQSOs, loadedQTCs)

	return nil
}

func (l *Logbook) loadLocked(qsos []core.QSO, qtcs []core.QTC) ([]core.QSO, []core.QTC, error) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()

	l.clear(len(qsos)*2, len(qtcs)*2)
	err := l.putQSOs(qsos)
	if err != nil {
		return nil, nil, err
	}

	// TODO:
	// load qtcs
	// update derived data

	loadedQSOs := l.allQSOs()

	return loadedQSOs, nil, nil
}

func (l *Logbook) clear(capQSOs int, capQTCs int) {
	l.qsos = make([]core.QSO, 0, capQSOs)
	l.sentQTCs = make(map[core.QSONumber]core.QTC, capQTCs)
	l.receivedQTCs = make([]core.QTC, 0, capQTCs)
	l.sentQTCsPerSeries = make([]int, 0, capQTCs)
}

func (l *Logbook) putQSOs(qsos []core.QSO) error {
	for _, qso := range qsos {
		err := l.putQSO(qso)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Logbook) putQSO(qso core.QSO) error {
	if qso.MyNumber <= 0 {
		return fmt.Errorf("Logbook.putQSO: invalid QSO number %d", qso.MyNumber)
	}

	lastNumber := l.lastQSONumber()
	if (lastNumber <= 0) || (qso.MyNumber > lastNumber) {
		l.qsos = append(l.qsos, qso)
		return nil
	}

	index, found := l.findQSOIndex(qso.MyNumber)
	if !found {
		l.qsos = append(l.qsos[:index+1], l.qsos[index:]...)
	}
	l.qsos[index] = qso
	return nil
}

// QSOs

func (l *Logbook) AddQSO(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	err := l.addQSOLocked(qso)
	if err != nil {
		// TODO: handle otherwise?
		panic(err)
	}
	l.emitQSOAdded(qso)
}

func (l *Logbook) addQSOLocked(qso core.QSO) error {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()
	if qso.MyNumber <= 0 {
		return fmt.Errorf("Logbook.addQSO: invalid QSO number %d", qso.MyNumber)
	}

	lastNumber := l.lastQSONumber()
	if (lastNumber <= 0) || (qso.MyNumber > lastNumber) {
		l.qsos = append(l.qsos, qso)

		// TODO: update derived data

		return nil
	}

	// find the right index and insert the QSO
	index, found := l.findQSOIndex(qso.MyNumber)
	if found {
		return fmt.Errorf("Logbook.addQSO: a QSO with number %d already exists, cannot be added again, use Logbook.UpdateQSO instead", qso.MyNumber)
	}
	l.qsos = append(l.qsos[:index+1], l.qsos[index:]...)
	l.qsos[index] = qso

	// TODO: update derived data

	return nil
}

func (l *Logbook) lastQSONumber() core.QSONumber {
	lastIndex := len(l.qsos) - 1
	if lastIndex < 0 {
		return 0
	}
	return l.qsos[lastIndex].MyNumber
}

func (l *Logbook) findQSOIndex(number core.QSONumber) (int, bool) {
	low := 0
	high := len(l.qsos) - 1

	for low <= high {
		median := (low + high) / 2

		if l.qsos[median].MyNumber < number {
			low = median + 1
		} else {
			high = median - 1
		}
	}

	if low == len(l.qsos) || l.qsos[low].MyNumber != number {
		return low, false
	}

	return low, true
}

func (l *Logbook) UpdateQSO(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	updatedQSOs, updatedQTCs, err := l.updateQSOLocked(qso)
	if err != nil {
		// TODO: handle otherwise?
		panic(err)
	}

	l.emitNotificationsAfterRefresh(updatedQSOs, updatedQTCs)
}

func (l *Logbook) updateQSOLocked(qso core.QSO) ([]core.QSO, []core.QTC, error) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()
	if qso.MyNumber <= 0 {
		return nil, nil, fmt.Errorf("Logbook.updateQSO: invalid QSO number %d", qso.MyNumber)
	}

	index, found := l.findQSOIndex(qso.MyNumber)
	if !found {
		return nil, nil, fmt.Errorf("Logbook.updateQSO: the QSO with number %d cannot be found for update, use Logbook.UpdateQSO instead", qso.MyNumber)
	}
	l.qsos[index] = qso

	// TODO: update derived data

	updatedQSOs := l.allQSOs()
	updatedQTCs := l.allQTCs()

	return updatedQSOs, updatedQTCs, nil
}

func (l *Logbook) AllQSOs() []core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.allQSOs()
}

func (l *Logbook) allQSOs() []core.QSO {
	result := make([]core.QSO, len(l.qsos))
	copy(result, l.qsos)
	return result
}

// QTCs

func (l *Logbook) AddQTCSeries(series core.QTCSeries) {
	panic("not yet implemented")
}

func (l *Logbook) AllQTCs() []core.QTC {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.allQTCs()
}

func (l *Logbook) allQTCs() []core.QTC {
	// TODO: implement
	return nil
}

// Notifications

func (l *Logbook) Notify(listener any) {
	l.listeners = append(l.listeners, listener)
}

func (l *Logbook) emitNotificationsAfterRefresh(qsos []core.QSO, qtcs []core.QTC) {
	l.emitLogbookCleared()
	for _, qso := range qsos {
		l.emitQSOAdded(qso)
	}
	for _, qtc := range qtcs {
		l.emitQTCAdded(qtc)
	}
}

func (l *Logbook) emitLogbookCleared() {
	for _, rawListener := range l.listeners {
		if listener, ok := rawListener.(LogbookClearedListener); ok {
			listener.LogbookCleared()
		}
	}
}

func (l *Logbook) emitQSOAdded(qso core.QSO) {
	for _, rawListener := range l.listeners {
		if listener, ok := rawListener.(QSOAddedListener); ok {
			listener.QSOAdded(qso)
		}
	}
}

func (l *Logbook) emitQTCAdded(qtc core.QTC) {
	for _, rawListener := range l.listeners {
		if listener, ok := rawListener.(QTCAddedListener); ok {
			listener.QTCAdded(qtc)
		}
	}
}
