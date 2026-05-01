package logbook

import (
	"log"

	"github.com/ftl/hellocontest/core"
)

type QSOScorer interface {
	Clear()
	AddMuted(qso core.QSO) core.QSOScore
	Unmute()
}

type QSOView interface {
	Show()
}

// QSOList is the data source for the visible QSO list with all its additional information.
// It is based on the Logbook data but uses several other data sources to enrich the
// QSO information.
// QSOList is thread-safe.
type QSOList struct {
	myExchangeFields    []core.ExchangeField
	theirExchangeFields []core.ExchangeField

	list    []core.QSO
	invalid bool

	listeners []any
	view      QSOView
}

func NewQSOList(settings core.Settings) *QSOList {
	contest := settings.Contest()
	return &QSOList{
		myExchangeFields:    contest.MyExchangeFields,
		theirExchangeFields: contest.TheirExchangeFields,
		list:                make([]core.QSO, 0),
		view:                new(nullQSOView),
	}
}

func (l *QSOList) SetView(view QSOView) {
	if view == nil {
		l.view = new(nullQSOView)
	}
	l.view = view
}

func (l *QSOList) Show() {
	l.view.Show()
}

func (l *QSOList) ContestChanged(contest core.Contest) {
	l.myExchangeFields = contest.MyExchangeFields
	l.theirExchangeFields = contest.TheirExchangeFields
	l.emitExchangeFieldsChanged(l.myExchangeFields, l.theirExchangeFields)
	l.invalid = true
}

func (l *QSOList) LogbookCleared() {
	l.list = make([]core.QSO, 0)
	l.invalid = false

	l.emitQSOsCleared()
}

func (l *QSOList) QSOAdded(qso core.QSO) {
	l.list = append(l.list, qso)

	l.emitQSOAdded(qso)
}

func (l *QSOList) Valid() bool {
	return !l.invalid
}

func (l *QSOList) GetExchangeFields() ([]core.ExchangeField, []core.ExchangeField) {
	return l.myExchangeFields, l.theirExchangeFields
}

func (l *QSOList) SelectRow(index int) {
	if index < 0 || index >= len(l.list) {
		log.Printf("invalid QSO index %d", index)
		return
	}
	qso := l.list[index]

	l.emitQSOSelected(qso)
	l.emitQSORowSelected(index)
}

func (l *QSOList) SelectLastQSO() {
	if len(l.list) == 0 {
		return
	}

	index := len(l.list) - 1
	qso := l.list[index]

	l.emitQSOSelected(qso)
	l.emitQSORowSelected(index)
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

type nullQSOView struct{}

func (*nullQSOView) Show() {}
