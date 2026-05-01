package logbook

import (
	"log"

	"github.com/ftl/hellocontest/core"
)

type QTCView interface {
	Show()
}

type QTCList struct {
	enabled bool
	qtcs    []core.QTC

	listeners []any
	view      QTCView
}

func NewQTCList() *QTCList {
	return &QTCList{
		qtcs: make([]core.QTC, 0),
		view: new(nullQTCView),
	}
}

func (l *QTCList) SetView(view QTCView) {
	if view == nil {
		l.view = new(nullQTCView)
	}
	l.view = view
}

func (l *QTCList) Show() {
	l.view.Show()
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
	l.enabled = enabled
	l.emitQTCsEnabled(enabled)
}

func (l *QTCList) QTCsEnabled() bool {
	return l.enabled
}

func (l *QTCList) LogbookCleared() {
	l.qtcs = make([]core.QTC, 0)
	l.emitQTCsCleared()
}

func (l *QTCList) QTCAdded(qtc core.QTC) {
	l.qtcs = append(l.qtcs, qtc)
	l.emitQTCAdded(qtc)
}

func (l *QTCList) SelectRow(index int) {
	if index < 0 || index >= len(l.qtcs) {
		log.Printf("invalid QSO index %d", index)
		return
	}
	qtc := l.qtcs[index]

	l.emitQTCSelected(qtc)
	l.emitQTCRowSelected(index)
}

func (l *QTCList) SelectLastQTC() {
	if len(l.qtcs) == 0 {
		return
	}

	index := len(l.qtcs) - 1
	qtc := l.qtcs[index]

	l.emitQTCSelected(qtc)
	l.emitQTCRowSelected(index)
}

type nullQTCView struct{}

func (*nullQTCView) Show() {}
