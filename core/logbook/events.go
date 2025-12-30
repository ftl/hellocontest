package logbook

import "github.com/ftl/hellocontest/core"

type LogbookClearedListener interface {
	LogbookCleared()
}

type LogbookClearedFunc func()

func (f LogbookClearedFunc) LogbookCleared() {
	f()
}

// deprecated
type QSOsClearedListener interface {
	QSOsCleared()
}

// deprecated
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

type QSORowSelectedListener interface {
	QSORowSelected(int)
}

type QSORowSelectedListenerFunc func(int)

func (f QSORowSelectedListenerFunc) QSORowSelected(index int) {
	f(index)
}

type ExchangeFieldsChangedListener interface {
	ExchangeFieldsChanged(myExchangeFields []core.ExchangeField, theirExchangeFields []core.ExchangeField)
}

type ExchangeFieldsChangedListenerFunc func(myExchangeFields []core.ExchangeField, theirExchangeFields []core.ExchangeField)

func (f ExchangeFieldsChangedListenerFunc) ExchangeFieldsChanged(myExchangeFields []core.ExchangeField, theirExchangeFields []core.ExchangeField) {
	f(myExchangeFields, theirExchangeFields)
}

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

// deprecated
type QTCsClearedListener interface {
	QTCsCleared()
}

type ScoreChangedListener interface {
	ScoreChanged(core.Score)
}

type ScoreChangedFunc func(core.Score)

func (f ScoreChangedFunc) ScoreChanged(score core.Score) {
	f(score)
}
