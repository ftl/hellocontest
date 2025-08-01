package logbook

import (
	"log"
	"sync"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/scp"

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

type ExchangeFieldsChangedListener interface {
	ExchangeFieldsChanged(myExchangeFields []core.ExchangeField, theirExchangeFields []core.ExchangeField)
}

type ExchangeFieldsChangedListenerFunc func(myExchangeFields []core.ExchangeField, theirExchangeFields []core.ExchangeField)

func (f ExchangeFieldsChangedListenerFunc) ExchangeFieldsChanged(myExchangeFields []core.ExchangeField, theirExchangeFields []core.ExchangeField) {
	f(myExchangeFields, theirExchangeFields)
}

type QSOScorer interface {
	Clear()
	AddMuted(qso core.QSO) core.QSOScore
	Unmute()
}

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
	l.invalid = false
}

func (l *QSOList) Fill(qsos []core.QSO) {
	l.dataLock.Lock()

	l.scorer.Clear()
	if len(l.list) > 0 {
		l.clear()
	}

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

func (l *QSOList) Put(qso core.QSO) {
	l.dataLock.Lock()

	emitNotifications := l.put(qso)

	l.refreshScore()
	allQSOs := l.all()
	l.dataLock.Unlock()

	emitNotifications(allQSOs)
}

func (l *QSOList) put(qso core.QSO) func([]core.QSO) {
	if len(l.list) == 0 {
		return l.append(qso)
	}

	lastNumber := l.list[len(l.list)-1].MyNumber
	if qso.MyNumber > lastNumber {
		return l.append(qso)
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
	}
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

func (l *QSOList) append(qso core.QSO) func([]core.QSO) {
	score := l.scorer.AddMuted(qso)
	qso.Points = score.Points
	qso.Multis = score.Multis
	qso.Duplicate = score.Duplicate

	dupeBand, dupeMode := l.dupeBandAndMode(qso.Band, qso.Mode)
	l.dupes.Add(qso.Callsign, dupeBand, dupeMode, qso.MyNumber)
	l.worked.Add(qso.Callsign, core.NoBand, core.NoMode, qso.MyNumber)
	l.callsigns.Add(qso.Callsign.String())

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
	l.emitRowSelected(index)
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
	l.emitRowSelected(index)
}

func (l *QSOList) FindDuplicateQSOs(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	band, mode = l.dupeBandAndMode(band, mode)
	numbers := l.dupes.Get(callsign, band, mode)

	return l.getQSOs(numbers)
}

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

func toAnnotatedCallsigns(matches []scp.Match) []core.AnnotatedCallsign {
	result := make([]core.AnnotatedCallsign, 0, len(matches))

	for _, match := range matches {
		annotatedCallsign, err := toAnnotatedCallsign(match)
		if err != nil {
			log.Print(err)
			continue
		}
		result = append(result, annotatedCallsign)
	}

	return result
}

func toAnnotatedCallsign(match scp.Match) (core.AnnotatedCallsign, error) {
	cs, err := callsign.Parse(match.Key())
	if err != nil {
		return core.AnnotatedCallsign{}, nil
	}
	return core.AnnotatedCallsign{
		Callsign:   cs,
		Assembly:   toMatchingAssembly(match),
		Comparable: match,
		Compare: func(a any, b any) bool {
			aMatch, aOk := a.(scp.Match)
			bMatch, bOk := b.(scp.Match)
			if !aOk || !bOk {
				return false
			}
			return aMatch.LessThan(bMatch)
		},
	}, nil
}

func toMatchingAssembly(match scp.Match) core.MatchingAssembly {
	result := make(core.MatchingAssembly, len(match.Assembly))

	for i, part := range match.Assembly {
		result[i] = core.MatchingPart{OP: core.MatchingOperation(part.OP), Value: part.Value}
	}

	return result
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

func (l *QSOList) emitRowSelected(index int) {
	for _, listener := range l.listeners {
		if rowSelectedListener, ok := listener.(RowSelectedListener); ok {
			rowSelectedListener.RowSelected(index)
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
