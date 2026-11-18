package logbook

import "github.com/ftl/hellocontest/core/model"

type LogbookClearedListener interface {
	LogbookCleared()
}

type LogbookClearedFunc func()

func (f LogbookClearedFunc) LogbookCleared() {
	f()
}

type LogbookLoadedListener interface {
	LogbookLoaded()
}

type LogbookLoadedFunc func()

func (f LogbookLoadedFunc) LogbookLoaded() {
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
	QSOAdded(model.QSO)
}

type QSOAddedListenerFunc func(model.QSO)

func (f QSOAddedListenerFunc) QSOAdded(qso model.QSO) {
	f(qso)
}

type QSOSelectedListener interface {
	QSOSelected(model.QSO)
}

type QSOSelectedListenerFunc func(model.QSO)

func (f QSOSelectedListenerFunc) QSOSelected(qso model.QSO) {
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
	ExchangeFieldsChanged(myExchangeFields []model.ExchangeField, theirExchangeFields []model.ExchangeField)
}

type ExchangeFieldsChangedListenerFunc func(myExchangeFields []model.ExchangeField, theirExchangeFields []model.ExchangeField)

func (f ExchangeFieldsChangedListenerFunc) ExchangeFieldsChanged(myExchangeFields []model.ExchangeField, theirExchangeFields []model.ExchangeField) {
	f(myExchangeFields, theirExchangeFields)
}

type QTCAddedListener interface {
	QTCAdded(model.QTC)
}

type QTCAddedListenerFunc func(model.QTC)

func (f QTCAddedListenerFunc) QTCAdded(qtc model.QTC) {
	f(qtc)
}

type QTCsEnabledListener interface {
	SetQTCsEnabled(bool)
}

type QTCSelectedListener interface {
	QTCSelected(model.QTC)
}

type QTCSelectedListenerFunc func(model.QTC)

func (f QTCSelectedListenerFunc) QTCSelected(qtc model.QTC) {
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
	ScoreChanged(model.Score)
}

type ScoreChangedFunc func(model.Score)

func (f ScoreChangedFunc) ScoreChanged(score model.Score) {
	f(score)
}
