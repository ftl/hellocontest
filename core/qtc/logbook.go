package qtc

import (
	"slices"

	"github.com/pkg/errors"

	"github.com/ftl/hellocontest/core"
)

type QTCAddedListener interface {
	QTCAdded(core.QTC)
}

type Writer interface {
	WriteQTC(core.QTC) error
}

type Logbook struct {
	qtcs map[core.QSONumber]core.QTC

	writer    Writer
	listeners []any
}

func NewLogbook() *Logbook {
	return &Logbook{
		qtcs:   make(map[core.QSONumber]core.QTC),
		writer: &nullWriter{},
	}
}

func LoadLogbook(qtcs []core.QTC) *Logbook {
	result := NewLogbook()
	for _, qtc := range qtcs {
		result.Log(qtc)
	}
	return result
}

func (l *Logbook) SetWriter(writer Writer) {
	if writer == nil {
		l.writer = new(nullWriter)
	}
	l.writer = writer
}

func (l *Logbook) Notify(listener any) {
	l.listeners = append(l.listeners, listener)
}

func (l *Logbook) emitQTCAdded(qtc core.QTC) {
	for _, lis := range l.listeners {
		if listener, ok := lis.(QTCAddedListener); ok {
			listener.QTCAdded(qtc)
		}
	}
}

func (l *Logbook) Log(qtc core.QTC) {
	if qtc.Timestamp.IsZero() {
		panic("cannot log the given QTC, its timestamp must not be zero")
	}

	l.qtcs[qtc.QSONumber] = qtc
	l.writer.WriteQTC(qtc)
	l.emitQTCAdded(qtc)
}

func (l *Logbook) All() []core.QTC {
	result := make([]core.QTC, 0, len(l.qtcs))
	for _, qtc := range l.qtcs {
		result = append(result, qtc)
	}
	slices.SortStableFunc(result, core.QTCByTimestamp)
	return result
}

func (l *Logbook) ReplayAll() {
	for _, qtc := range l.All() {
		l.emitQTCAdded(qtc)
	}
}

func (l *Logbook) WriteAll(writer Writer) error {
	for _, qtc := range l.All() {
		err := writer.WriteQTC(qtc)
		if err != nil {
			return errors.Wrapf(err, "cannot write QTC %v", qtc)
		}
	}
	return nil
}

type nullWriter struct{}

func (d *nullWriter) WriteQTC(core.QTC) error {
	return nil
}
