package logbook

import (
	"log"
	"sync"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
)

type QSOScorer interface {
	Clear()
	AddMuted(qso core.QSO) core.QSOScore
	Unmute()
}

// QSOList is the data source for the visible QSO list with all its additional information.
// It is based on the Logbook data but uses several other data sources to enrich the
// QSO information.
// QSOList is thread-safe.
type QSOList struct {
	myExchangeFields    []core.ExchangeField
	theirExchangeFields []core.ExchangeField
	bandRule            conval.BandRule

	dataLock  *sync.RWMutex
	list      []core.QSO
	scorer    QSOScorer
	dupes     dupeIndex     // used to find duplicate QSOs for a given callsign, band and mode, according to the contest rules
	worked    dupeIndex     // used to find worked QSOs for a given callsign
	callsigns *scp.Database // used to find worked callsigns similar to a given string, e.g. for supercheck
	invalid   bool

	listeners []any
}

func NewQSOList(settings core.Settings, scorer QSOScorer) *QSOList {
	contest := settings.Contest()
	return &QSOList{
		myExchangeFields:    contest.MyExchangeFields,
		theirExchangeFields: contest.TheirExchangeFields,
		dataLock:            &sync.RWMutex{},
		list:                make([]core.QSO, 0),
		scorer:              scorer,
		dupes:               make(dupeIndex),
		worked:              make(dupeIndex),
		callsigns:           scp.NewDatabase(),
	}
}

func (l *QSOList) GetExchangeFields() ([]core.ExchangeField, []core.ExchangeField) {
	return l.myExchangeFields, l.theirExchangeFields
}

func (l *QSOList) ContestChanged(contest core.Contest) {
	l.myExchangeFields = contest.MyExchangeFields
	l.theirExchangeFields = contest.TheirExchangeFields
	l.emitExchangeFieldsChanged(l.myExchangeFields, l.theirExchangeFields)

	if contest.Definition != nil {
		l.bandRule = contest.Definition.Scoring.QSOBandRule
	}

	l.invalid = true
}

func (l *QSOList) Valid() bool {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()
	return !l.invalid
}

func (l *QSOList) Clear() {
	l.dataLock.Lock()
	l.clear()
	l.dataLock.Unlock()

	l.emitQSOsCleared()
}

func (l *QSOList) clear() {
	l.list = make([]core.QSO, 0)
	l.dupes = make(dupeIndex)
	l.worked = make(dupeIndex)
	l.callsigns = scp.NewDatabase()
	l.scorer.Clear()
	l.invalid = false
}

func (l *QSOList) Fill(qsos []core.QSO) {
	l.dataLock.Lock()
	l.clear()

	for _, qso := range qsos {
		l.put(qso)
	}
	allQSOs := l.all()

	l.dataLock.Unlock()

	l.scorer.Unmute()
	l.emitQSOsCleared()
	for _, qso := range allQSOs {
		l.emitQSOAdded(qso)
	}
}

func (l *QSOList) Put(qso core.QSO) bool {
	l.dataLock.Lock()

	emitNotifications, updated := l.put(qso)

	l.refreshScore()
	allQSOs := l.all()
	l.dataLock.Unlock()

	emitNotifications(allQSOs)

	return updated
}

func (l *QSOList) put(qso core.QSO) (func([]core.QSO), bool) {
	if len(l.list) == 0 {
		return l.append(qso), false
	}

	lastNumber := l.list[len(l.list)-1].MyNumber
	if qso.MyNumber > lastNumber {
		return l.append(qso), false
	}

	index, found := l.findIndex(qso.MyNumber)
	if !found {
		l.insert(index, qso)
	} else {
		l.update(index, qso)
	}

	return func(qsos []core.QSO) {
		l.scorer.Unmute()
		l.emitQSOsCleared()
		for _, qso := range qsos {
			l.emitQSOAdded(qso)
		}
	}, true
}

func (l *QSOList) findIndex(number core.QSONumber) (int, bool) {
	return findIndex(l.list, number)
}

func (l *QSOList) append(qso core.QSO) func([]core.QSO) {
	l.list = append(l.list, qso)

	return func([]core.QSO) {
		l.scorer.Unmute()
		l.emitQSOAdded(qso)
	}
}

func (l *QSOList) insert(index int, qso core.QSO) {
	l.list = append(l.list[:index+1], l.list[index:]...)
	l.list[index] = qso
}

func (l *QSOList) update(index int, qso core.QSO) {
	l.list[index] = qso
}

func (l *QSOList) refreshScore() {
	l.scorer.Clear()
	l.dupes = make(dupeIndex)
	l.worked = make(dupeIndex)
	l.callsigns = scp.NewDatabase()
	for i, qso := range l.list {
		score := l.scorer.AddMuted(qso)
		qso.Points = score.Points
		qso.Multis = score.Multis
		qso.Duplicate = score.Duplicate

		dupeBand, dupeMode := l.dupeBandAndMode(qso.Band, qso.Mode)
		l.dupes.Add(qso.Callsign, dupeBand, dupeMode, qso.MyNumber)
		l.worked.Add(qso.Callsign, core.NoBand, core.NoMode, qso.MyNumber)
		l.callsigns.Add(qso.Callsign.String())

		l.list[i] = qso
	}
}

func (l *QSOList) dupeBandAndMode(band core.Band, mode core.Mode) (core.Band, core.Mode) {
	switch l.bandRule {
	case conval.Once:
		return core.NoBand, core.NoMode
	case conval.OncePerBand:
		return band, core.NoMode
	case conval.OncePerBandAndMode:
		return band, mode
	default:
		return core.NoBand, core.NoMode
	}
}

func (l *QSOList) All() []core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.all()
}

func (l *QSOList) all() []core.QSO {
	result := make([]core.QSO, len(l.list))
	copy(result, l.list)
	return result
}

func (l *QSOList) SelectRow(index int) {
	l.dataLock.RLock()

	if index < 0 || index >= len(l.list) {
		log.Printf("invalid QSO index %d", index)
		l.dataLock.RUnlock()
		return
	}
	qso := l.list[index]

	l.dataLock.RUnlock()

	l.emitQSOSelected(qso)
	l.emitQSORowSelected(index)
}

func (l *QSOList) SelectLastQSO() {
	l.dataLock.RLock()

	if len(l.list) == 0 {
		l.dataLock.RUnlock()
		return
	}

	index := len(l.list) - 1
	qso := l.list[index]

	l.dataLock.RUnlock()

	l.emitQSOSelected(qso)
	l.emitQSORowSelected(index)
}

// deprecated
func (l *QSOList) FindDuplicateQSOs(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	band, mode = l.dupeBandAndMode(band, mode)
	numbers := l.dupes.Get(callsign, band, mode)

	return l.getQSOs(numbers)
}

// deprecated
func (l *QSOList) GetQSOs(numbers []core.QSONumber) []core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.getQSOs(numbers)
}

func (l *QSOList) getQSOs(numbers []core.QSONumber) []core.QSO {
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

// deprecated
func (l *QSOList) FindWorkedQSOs(callsign callsign.Callsign, band core.Band, mode core.Mode) ([]core.QSO, bool) {
	l.dataLock.RLock()

	numbers := l.worked.Get(callsign, core.NoBand, core.NoMode)
	qsos := l.getQSOs(numbers)

	l.dataLock.RUnlock()

	if len(qsos) == 0 {
		return qsos, false
	}

	duplicate := false
	for _, qso := range qsos {
		switch l.bandRule {
		case conval.Once:
			duplicate = true
		case conval.OncePerBand:
			duplicate = (qso.Band == band)
		case conval.OncePerBandAndMode:
			duplicate = (qso.Band == band) && (qso.Mode == mode)
		default:
			duplicate = false
		}
		if duplicate {
			break
		}
	}
	return qsos, duplicate
}

// deprecated
func (l *QSOList) Find(s string) ([]core.AnnotatedCallsign, error) {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	if l.callsigns == nil {
		return nil, nil
	}
	matches, err := l.callsigns.Find(s)
	if err != nil {
		return nil, err
	}

	return toAnnotatedCallsigns(matches), nil
}

func (l *QSOList) Notify(listener any) {
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

func (l *QSOList) emitQSOSelected(qso core.QSO) {
	for _, listener := range l.listeners {
		if qsoSelectedListener, ok := listener.(QSOSelectedListener); ok {
			qsoSelectedListener.QSOSelected(qso)
		}
	}
}

func (l *QSOList) emitQSORowSelected(index int) {
	for _, listener := range l.listeners {
		if rowSelectedListener, ok := listener.(QSORowSelectedListener); ok {
			rowSelectedListener.QSORowSelected(index)
		}
	}
}

func (l *QSOList) emitExchangeFieldsChanged(myExchangeFields []core.ExchangeField, theirExchangeFields []core.ExchangeField) {
	for _, listener := range l.listeners {
		if exchangeFieldsChangedListener, ok := listener.(ExchangeFieldsChangedListener); ok {
			exchangeFieldsChangedListener.ExchangeFieldsChanged(myExchangeFields, theirExchangeFields)
		}
	}
}
