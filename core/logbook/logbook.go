package logbook

import (
	"fmt"
	"log"
	"math"
	"slices"

	"github.com/ftl/hellocontest/core"
)

type QSOAddedListener interface {
	QSOAdded(core.QSO)
}

type QSOAddedListenerFunc func(core.QSO)

func (f QSOAddedListenerFunc) QSOAdded(qso core.QSO) {
	f(qso)
}

type QTCAddedListener interface {
	QTCAdded(core.QTC)
}

type QTCAddedListenerFunc func(core.QTC)

func (f QTCAddedListenerFunc) QTCAdded(qtc core.QTC) {
	f(qtc)
}

type Writer interface {
	WriteQSO(core.QSO) error
	WriteQTC(core.QTC) error
}

type Logbook struct {
	clock             core.Clock
	writer            Writer
	qsos              []core.QSO
	myLastNumber      int
	qtcs              map[core.QSONumber]core.QTC
	sentQTCsPerSeries []int

	listeners []any
}

// New creates a new empty logbook.
func New(clock core.Clock) *Logbook {
	return &Logbook{
		clock:  clock,
		writer: new(nullWriter),
		qsos:   make([]core.QSO, 0, 1000),
		qtcs:   make(map[core.QSONumber]core.QTC),
	}
}

// Load creates a new log and loads it with the entries from the given reader.
func Load(clock core.Clock, qsos []core.QSO, qtcs []core.QTC) *Logbook {
	result := &Logbook{
		clock:             clock,
		writer:            new(nullWriter),
		qsos:              qsos,
		myLastNumber:      lastNumber(qsos),
		qtcs:              make(map[core.QSONumber]core.QTC, len(qtcs)),
		sentQTCsPerSeries: make([]int, lastSeries(qtcs)),
	}

	for _, qtc := range qtcs {
		if qtc.Timestamp.IsZero() {
			panic(fmt.Errorf("cannot load qtc because its timestamp is unset: %v", qtc))
		}
		result.qtcs[qtc.QSONumber] = qtc
		result.registerQTCSeries(qtc)
	}
	// TODO: setup the lookup table with available QTCs, should probably go into the QTCList

	return result
}

func lastNumber(qsos []core.QSO) int {
	lastNumber := 0
	for _, qso := range qsos {
		lastNumber = max(lastNumber, int(qso.MyNumber))
	}
	return lastNumber
}

func lastSeries(qtcs []core.QTC) int {
	result := 0
	for _, qtc := range qtcs {
		if qtc.Kind != core.SentQTC {
			continue
		}
		result = max(result, qtc.Header.SeriesNumber)
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

func (l *Logbook) emitQSOAdded(qso core.QSO) {
	for _, lis := range l.listeners {
		if listener, ok := lis.(QSOAddedListener); ok {
			listener.QSOAdded(qso)
		}
	}
}

func (l *Logbook) emitQTCAdded(qtc core.QTC) {
	for _, lis := range l.listeners {
		if listener, ok := lis.(QTCAddedListener); ok {
			listener.QTCAdded(qtc)
		}
	}
}

func (l *Logbook) NextNumber() core.QSONumber {
	return core.QSONumber(l.myLastNumber + 1)
}

func (l *Logbook) NextSeriesNumber() int {
	return len(l.sentQTCsPerSeries) + 1
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

func (l *Logbook) LogQSO(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	l.qsos = append(l.qsos, qso)
	l.myLastNumber = int(math.Max(float64(l.myLastNumber), float64(qso.MyNumber)))
	l.writer.WriteQSO(qso)
	l.emitQSOAdded(qso)
	log.Printf("QSO added: %s", qso.String())
}

func (l *Logbook) AllQSOs() []core.QSO {
	return l.qsos
}

func (l *Logbook) LogQTC(qtc core.QTC) {
	if qtc.Timestamp.IsZero() {
		panic("cannot log the given QTC, its timestamp must not be zero")
	}
	if existing, ok := l.qtcs[qtc.QSONumber]; ok {
		panic(fmt.Errorf("QTC for QSO #%d already exists, cannot log another QTC for the same QSO: %v", qtc.QSONumber, existing))
	}

	l.qtcs[qtc.QSONumber] = qtc
	l.writer.WriteQTC(qtc)
	l.registerQTCSeries(qtc)
	l.emitQTCAdded(qtc)
	log.Printf("QTC added: %v", qtc)
}

func (l *Logbook) registerQTCSeries(qtc core.QTC) {
	if qtc.Kind != core.SentQTC {
		return
	}

	index := qtc.Header.SeriesNumber - 1
	switch {
	case len(l.sentQTCsPerSeries) == index: // the first of a new series
		l.sentQTCsPerSeries = append(l.sentQTCsPerSeries, 1)
	case len(l.sentQTCsPerSeries) > index: // the next of an existing series
		l.sentQTCsPerSeries[index]++
		// TODO: check if the series contains more than Header.QTCCount
	default: // this must never happen, the calculation of the next series number is broken
		panic(fmt.Errorf("unknown QTC series number %d, should not be greater than %d", qtc.Header.SeriesNumber, len(l.sentQTCsPerSeries)))
	}
}

func (l *Logbook) AllQTCs() []core.QTC {
	result := make([]core.QTC, 0, len(l.qtcs))
	for _, qtc := range l.qtcs {
		result = append(result, qtc)
	}
	slices.SortStableFunc(result, core.QTCByTimestamp)
	return result
}

func (l *Logbook) WriteAll(writer Writer) error {
	err := l.writeAllQSOs(writer)
	if err != nil {
		return err
	}
	return l.writeAllQTCs(writer)
}

func (l *Logbook) writeAllQSOs(writer Writer) error {
	for _, qso := range l.qsos {
		err := writer.WriteQSO(qso)
		if err != nil {
			return fmt.Errorf("cannot write QSO %v: %w", qso, err)
		}
	}
	return nil
}

func (l *Logbook) writeAllQTCs(writer Writer) error {
	for _, qtc := range l.AllQTCs() {
		err := writer.WriteQTC(qtc)
		if err != nil {
			return fmt.Errorf("cannot write QTC %v: %w", qtc, err)
		}
	}
	return nil
}

func (l *Logbook) ReplayAll() {
	for _, qso := range l.qsos {
		l.emitQSOAdded(qso)
	}
	for _, qtc := range l.AllQTCs() {
		l.emitQTCAdded(qtc)
	}
}

type nullWriter struct{}

func (d *nullWriter) WriteQSO(core.QSO) error {
	return nil
}

func (d *nullWriter) WriteQTC(core.QTC) error {
	return nil
}
