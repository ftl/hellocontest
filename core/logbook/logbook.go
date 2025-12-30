package logbook

import (
	"fmt"
	"sync"

	"github.com/ftl/hellocontest/core"
)

type Logbook struct {
	clock core.Clock

	dataLock          *sync.RWMutex
	qsos              []core.QSO
	sentQTCs          map[core.QSONumber]core.QTC // TODO: better store the QTC series?
	receivedQTCs      []core.QTC                  // TODO: better store QTC series?
	sentQTCsPerSeries []int
}

func NewLogbook(clock core.Clock) *Logbook {
	return &Logbook{
		clock:    clock,
		dataLock: new(sync.RWMutex),
	}
}

func (l *Logbook) Load(qsos []core.QSO, qtcs []core.QTC) error {
	return fmt.Errorf("not yet implemented")
}

func (l *Logbook) AddQSO(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	err := l.addQSOLocked(qso)
	if err != nil {
		// TODO: handle otherwise?
		panic(err)
	}
	// TODO: l.emitQSOAdded(qso)
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
		return nil
	}

	// find the right index and insert the QSO
	index, found := l.findQSOIndex(qso.MyNumber)
	if found {
		return fmt.Errorf("Logbook.addQSO: a QSO with number %d already exists, cannot be added again, use Logbook.UpdateQSO instead", qso.MyNumber)
	}
	l.qsos = append(l.qsos[:index+1], l.qsos[index:]...)
	l.qsos[index] = qso

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
	err := l.updateQSOLocked(qso)
	if err != nil {
		// TODO: handle otherwise?
		panic(err)
	}
	// TODO: l.emitQSOUpdated(qso)
}

func (l *Logbook) updateQSOLocked(qso core.QSO) error {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()

	index, found := l.findQSOIndex(qso.MyNumber)
	if !found {
		return fmt.Errorf("Logbook.updateQSO: the QSO with number %d cannot be for update, use Logbook.UpdateQSO instead", qso.MyNumber)
	}
	l.qsos = append(l.qsos[:index+1], l.qsos[index:]...)
	l.qsos[index] = qso

	return nil
}

func (l *Logbook) AllQSOs() []core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	result := make([]core.QSO, len(l.qsos))
	copy(result, l.qsos)
	return result
}

func (l *Logbook) AddQTCSeries(series core.QTCSeries) {

}

func (l *Logbook) AllQTCs() []core.QTC {
	panic("not yet implemented")
}
