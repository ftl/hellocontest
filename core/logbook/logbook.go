package logbook

import (
	"fmt"
	"sync"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
)

type DXCCEntities interface {
	Find(string) (dxcc.Prefix, bool)
}

type Logbook struct {
	clock core.Clock

	listeners []any

	dataLock          *sync.RWMutex
	qsos              []core.QSO
	sentQTCs          map[core.QSONumber]core.QTC // TODO: better store the QTC series?
	receivedQTCs      []core.QTC                  // TODO: better store QTC series?
	sentQTCsPerSeries []int

	scoreCounter *scoreCounter
}

func NewLogbook(clock core.Clock, settings core.Settings, entities DXCCEntities) *Logbook {
	result := &Logbook{
		clock:        clock,
		dataLock:     new(sync.RWMutex),
		scoreCounter: newScoreCounter(settings, entities),
	}

	result.clear(1000, 0)

	return result
}

// Settings

func (l *Logbook) StationChanged(station core.Station) {
	l.scoreCounter.StationChanged(station)
	if !l.scoreCounter.Valid() {
		return
	}
	// TODO: refresh derived data if l.scoreCounter.Valid()
}

func (l *Logbook) ContestChanged(contest core.Contest) {
	l.scoreCounter.ContestChanged(contest)
	if !l.scoreCounter.Valid() {
		return
	}
	// TODO: refresh derived data if l.scoreCounter.Valid()
}

// Loading

func (l *Logbook) Load(qsos []core.QSO, qtcs []core.QTC) error {
	loadedQSOs, loadedQTCs, score, err := l.loadLocked(qsos, qtcs)
	if err != nil {
		return err
	}

	l.emitNotificationsAfterRefresh(loadedQSOs, loadedQTCs, score)

	return nil
}

func (l *Logbook) loadLocked(qsos []core.QSO, qtcs []core.QTC) ([]core.QSO, []core.QTC, core.Score, error) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()

	l.clear(len(qsos)*2, len(qtcs)*2)
	err := l.putQSOs(qsos)
	if err != nil {
		return nil, nil, core.Score{}, err
	}

	// TODO: load qtcs

	// update first, because the QSO score might change during the update
	score := l.refreshDerivedData()

	// copy QSOs and QTCs for further processing outside the dataLock
	loadedQSOs := l.allQSOs()
	loadedQTCs := l.allQTCs()

	return loadedQSOs, loadedQTCs, score, nil
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
	qso, score, err := l.addQSOLocked(qso)
	if err != nil {
		// this must never happen
		panic(err)
	}
	l.emitQSOAdded(qso)
	l.emitScoreChanged(score)
}

func (l *Logbook) addQSOLocked(qso core.QSO) (core.QSO, core.Score, error) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()
	if qso.MyNumber <= 0 {
		return core.QSO{}, core.Score{}, fmt.Errorf("Logbook.addQSO: invalid QSO number %d", qso.MyNumber)
	}

	lastNumber := l.lastQSONumber()
	if (lastNumber <= 0) || (qso.MyNumber > lastNumber) {
		qso = l.scoreCounter.AddQSO(qso)
		l.qsos = append(l.qsos, qso)
		score := l.scoreCounter.Score()

		return qso, score, nil
	}

	// find the right index and insert the QSO
	index, found := l.findQSOIndex(qso.MyNumber)
	if found {
		return core.QSO{}, core.Score{}, fmt.Errorf("Logbook.addQSO: a QSO with number %d already exists, cannot be added again, use Logbook.UpdateQSO instead", qso.MyNumber)
	}
	qso = l.scoreCounter.AddQSO(qso)
	l.qsos = append(l.qsos[:index+1], l.qsos[index:]...)
	l.qsos[index] = qso
	score := l.scoreCounter.Score()

	return qso, score, nil
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
	updatedQSOs, updatedQTCs, score, err := l.updateQSOLocked(qso)
	if err != nil {
		// this must never happen
		panic(err)
	}

	l.emitNotificationsAfterRefresh(updatedQSOs, updatedQTCs, score)
}

func (l *Logbook) updateQSOLocked(qso core.QSO) ([]core.QSO, []core.QTC, core.Score, error) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()
	if qso.MyNumber <= 0 {
		return nil, nil, core.Score{}, fmt.Errorf("Logbook.updateQSO: invalid QSO number %d", qso.MyNumber)
	}

	index, found := l.findQSOIndex(qso.MyNumber)
	if !found {
		return nil, nil, core.Score{}, fmt.Errorf("Logbook.updateQSO: the QSO with number %d cannot be found for update, use Logbook.UpdateQSO instead", qso.MyNumber)
	}
	l.qsos[index] = qso

	// update derived data first, because the QSO score might change during the update
	score := l.refreshDerivedData()

	// copy QSOs and QTCs for further processing outside the dataLock
	updatedQSOs := l.allQSOs()
	updatedQTCs := l.allQTCs()

	return updatedQSOs, updatedQTCs, score, nil
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

// Derived Data

func (l *Logbook) Score() core.Score {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.scoreCounter.Score()
}

func (l *Logbook) refreshDerivedData() core.Score {
	l.scoreCounter.Clear()
	for i, qso := range l.qsos {
		l.qsos[i] = l.scoreCounter.AddQSO(qso)
	}
	score := l.scoreCounter.Score()

	// TODO: update dupes, worked, callsigns

	return score
}

// Notifications

func (l *Logbook) Notify(listener any) {
	l.listeners = append(l.listeners, listener)
}

func (l *Logbook) emitNotificationsAfterRefresh(qsos []core.QSO, qtcs []core.QTC, score core.Score) {
	l.emitLogbookCleared()
	for _, qso := range qsos {
		l.emitQSOAdded(qso)
	}
	for _, qtc := range qtcs {
		l.emitQTCAdded(qtc)
	}
	l.emitScoreChanged(score)
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

func (l *Logbook) emitScoreChanged(score core.Score) {
	for _, rawListener := range l.listeners {
		if listener, ok := rawListener.(ScoreChangedListener); ok {
			listener.ScoreChanged(score)
		}
	}
}
