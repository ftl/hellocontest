package logbook

import (
	"log"
	"math"
	"sort"

	"github.com/pkg/errors"

	"github.com/ftl/hellocontest/core"
)

// New creates a new empty logbook.
func New(clock core.Clock) *Logbook {
	return &Logbook{
		clock:             clock,
		writer:            new(nullWriter),
		qsos:              make([]core.QSO, 0, 1000),
		rowAddedListeners: make([]RowAddedListener, 0),
	}
}

// Load creates a new log and loads it with the entries from the given reader.
func Load(clock core.Clock, reader Reader) (*Logbook, error) {
	log.Print("Loading QSOs")
	logbook := &Logbook{
		clock:             clock,
		writer:            new(nullWriter),
		rowAddedListeners: make([]RowAddedListener, 0),
	}

	var err error
	logbook.qsos, err = reader.ReadAll()
	if err != nil {
		return nil, err
	}

	logbook.myLastNumber = lastNumber(logbook.qsos)
	return logbook, nil
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

	rowAddedListeners []RowAddedListener
}

// Reader reads log entries.
type Reader interface {
	ReadAll() ([]core.QSO, error)
}

// Writer writes log entries.
type Writer interface {
	Write(core.QSO) error
}

// Store allows to read and write log entries.
type Store interface {
	Reader
	Writer
	Clear() error
}

// RowAddedListener is notified when a new row is added to the log.
type RowAddedListener func(core.QSO)

func (l *Logbook) SetWriter(writer Writer) {
	if writer == nil {
		l.writer = new(nullWriter)
	}
	l.writer = writer
}

func (l *Logbook) OnRowAdded(listener RowAddedListener) {
	l.rowAddedListeners = append(l.rowAddedListeners, listener)
}

func (l *Logbook) ClearRowAddedListeners() {
	l.rowAddedListeners = make([]RowAddedListener, 0)
}

func (l *Logbook) emitRowAdded(qso core.QSO) {
	for _, listener := range l.rowAddedListeners {
		listener(qso)
	}
}

func (l *Logbook) ReplayAll() {
	for _, qso := range l.qsos {
		l.emitRowAdded(qso)
	}
}

func (l *Logbook) NextNumber() core.QSONumber {
	return core.QSONumber(l.myLastNumber + 1)
}

func (l *Logbook) Log(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	l.qsos = append(l.qsos, qso)
	l.myLastNumber = int(math.Max(float64(l.myLastNumber), float64(qso.MyNumber)))
	l.writer.Write(qso)
	l.emitRowAdded(qso)
	log.Printf("QSO added: %s", qso.String())
}

func (l *Logbook) QsosOrderedByMyNumber() []core.QSO { // TODO use QSOList instead
	return byMyNumber(l.qsos)
}

func (l *Logbook) UniqueQsosOrderedByMyNumber() []core.QSO { // TODO use QSOList instead
	return byMyNumber(unique(l.qsos))
}

func byMyNumber(qsos []core.QSO) []core.QSO {
	result := make([]core.QSO, len(qsos))
	copy(result, qsos)
	sort.Slice(result, func(i, j int) bool {
		if result[i].MyNumber == result[j].MyNumber {
			return result[i].LogTimestamp.Before(result[j].LogTimestamp)
		}
		return result[i].MyNumber < result[j].MyNumber
	})
	return result
}

func unique(qsos []core.QSO) []core.QSO {
	index := make(map[core.QSONumber]core.QSO)
	for _, qso := range qsos {
		former, ok := index[qso.MyNumber]
		if !ok || qso.LogTimestamp.After(former.LogTimestamp) {
			index[qso.MyNumber] = qso
		}
	}

	result := make([]core.QSO, len(index))
	i := 0
	for _, qso := range index {
		result[i] = qso
		i++
	}
	return result
}

func (l *Logbook) WriteAll(writer Writer) error {
	for _, qso := range l.qsos {
		err := writer.Write(qso)
		if err != nil {
			return errors.Wrapf(err, "cannot write QSO %v", qso)
		}
	}
	return nil
}

type nullWriter struct{}

func (d *nullWriter) Write(core.QSO) error {
	return nil
}
