package qso

import (
	"log"
	"math"

	"github.com/pkg/errors"

	"github.com/ftl/hellocontest/core"
)

// New creates a new empty logbook.
func New(clock core.Clock) *Logbook {
	return &Logbook{
		clock:     clock,
		writer:    new(nullWriter),
		qsos:      make([]core.QSO, 0, 1000),
		listeners: make([]any, 0),
	}
}

// LoadLogbook creates a new log and loads it with the entries from the given reader.
func LoadLogbook(clock core.Clock, qsos []core.QSO) *Logbook {
	return &Logbook{
		clock:        clock,
		writer:       new(nullWriter),
		qsos:         qsos,
		myLastNumber: lastNumber(qsos),
		listeners:    make([]any, 0),
	}
}

func lastNumber(qsos []core.QSO) int {
	lastNumber := 0
	for _, qso := range qsos {
		lastNumber = int(math.Max(float64(lastNumber), float64(qso.MyNumber)))
	}
	return lastNumber
}

type Logbook struct {
	clock        core.Clock
	writer       Writer
	qsos         []core.QSO
	myLastNumber int

	listeners []any
}

// Writer writes log entries.
type Writer interface {
	WriteQSO(core.QSO) error
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

func (l *Logbook) emitQSOAdded(qso core.QSO) {
	for _, lis := range l.listeners {
		if listener, ok := lis.(QSOAddedListener); ok {
			listener.QSOAdded(qso)
		}
	}
}

func (l *Logbook) ReplayAll() {
	for _, qso := range l.qsos {
		l.emitQSOAdded(qso)
	}
}

func (l *Logbook) NextNumber() core.QSONumber {
	return core.QSONumber(l.myLastNumber + 1)
}

func (l *Logbook) lastQSO() core.QSO {
	if len(l.qsos) == 0 {
		return core.QSO{}
	}
	return l.qsos[len(l.qsos)-1]
}

func (l *Logbook) LastBand() core.Band {
	return l.lastQSO().Band
}

func (l *Logbook) LastMode() core.Mode {
	return l.lastQSO().Mode
}

func (l *Logbook) LastExchange() []string {
	return l.lastQSO().MyExchange
}

func (l *Logbook) Log(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	l.qsos = append(l.qsos, qso)
	l.myLastNumber = int(math.Max(float64(l.myLastNumber), float64(qso.MyNumber)))
	l.writer.WriteQSO(qso)
	l.emitQSOAdded(qso)
	log.Printf("QSO added: %s", qso.String())
}

func (l *Logbook) All() []core.QSO {
	return l.qsos
}

func (l *Logbook) WriteAll(writer Writer) error {
	for _, qso := range l.qsos {
		err := writer.WriteQSO(qso)
		if err != nil {
			return errors.Wrapf(err, "cannot write QSO %v", qso)
		}
	}
	return nil
}

type nullWriter struct{}

func (d *nullWriter) WriteQSO(core.QSO) error {
	return nil
}
