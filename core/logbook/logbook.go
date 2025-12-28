package logbook

import (
	"fmt"
	"log"
	"slices"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

type Writer interface {
	WriteQSO(core.QSO) error
	WriteQTC(core.QTC) error
}

type Logbook struct {
	clock   core.Clock
	writer  Writer
	qsoList *QSOList
	qtcList *QTCList

	qsos              []core.QSO
	myLastNumber      core.QSONumber
	sentQTCs          map[core.QSONumber]core.QTC
	receivedQTCs      []core.QTC
	sentQTCsPerSeries []int

	listeners []any
}

// New creates a new empty logbook.
func New(clock core.Clock, qsoList *QSOList, qtcList *QTCList) *Logbook {
	qsoList.Clear()
	qtcList.Clear()

	return &Logbook{
		clock:        clock,
		writer:       new(nullWriter),
		qsoList:      qsoList,
		qtcList:      qtcList,
		qsos:         make([]core.QSO, 0, 1000),
		sentQTCs:     make(map[core.QSONumber]core.QTC),
		receivedQTCs: make([]core.QTC, 0),
	}
}

// Load creates a new log and loads it with the entries from the given reader.
func Load(clock core.Clock, qsoList *QSOList, qtcList *QTCList, qsos []core.QSO, qtcs []core.QTC) *Logbook {
	result := &Logbook{
		clock:             clock,
		writer:            new(nullWriter),
		qsoList:           qsoList,
		qtcList:           qtcList,
		qsos:              qsos,
		myLastNumber:      lastNumber(qsos),
		sentQTCs:          make(map[core.QSONumber]core.QTC, len(qtcs)),
		sentQTCsPerSeries: make([]int, lastSeries(qtcs)),
		receivedQTCs:      make([]core.QTC, 0),
	}

	for _, qtc := range qtcs {
		if qtc.Timestamp.IsZero() {
			panic(fmt.Errorf("cannot load qtc because its timestamp is unset: %v", qtc))
		}
		if qtc.Kind == core.SentQTC {
			result.sentQTCs[qtc.QSONumber] = qtc
			result.registerQTCSeries(qtc)
		} else {
			result.receivedQTCs = append(result.receivedQTCs, qtc)
		}
	}

	qsoList.Fill(qsos)
	qtcList.Fill(qsos, qtcs)

	return result
}

func lastNumber(qsos []core.QSO) core.QSONumber {
	var lastNumber core.QSONumber = 0
	for _, qso := range qsos {
		lastNumber = max(lastNumber, qso.MyNumber)
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

func (l *Logbook) LastCallsign() callsign.Callsign {
	return l.lastQSO().Callsign
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

func (l *Logbook) Refresh() {
	l.qsoList.Fill(l.qsos)
	l.qtcList.Fill(l.qsoList.All(), l.allQTCs())
}

func (l *Logbook) allQTCs() []core.QTC {
	result := make([]core.QTC, 0, len(l.sentQTCs)+len(l.receivedQTCs))
	for _, qtc := range l.sentQTCs {
		result = append(result, qtc)
	}
	result = append(result, l.receivedQTCs...)
	slices.SortStableFunc(result, core.QTCByTimestamp)
	return result
}

func (l *Logbook) LogQSO(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	l.qsos = append(l.qsos, qso)
	l.myLastNumber = max(l.myLastNumber, qso.MyNumber)
	l.writer.WriteQSO(qso)
	l.qsoList.QSOAdded(qso)
	log.Printf("QSO added: %s", qso.String())
}

func (l *Logbook) LogQTC(qtc core.QTC) {
	if err := qtc.VerifyComplete(); err != nil {
		panic(fmt.Errorf("cannot log the given QTC: %v", err))
	}

	if qtc.Kind == core.SentQTC {
		if existing, ok := l.sentQTCs[qtc.QSONumber]; ok {
			panic(fmt.Errorf("QTC for QSO #%d already exists, cannot log another QTC for the same QSO: %v", qtc.QSONumber, existing))
		}
		l.sentQTCs[qtc.QSONumber] = qtc
		l.registerQTCSeries(qtc)
	} else {
		l.receivedQTCs = append(l.receivedQTCs, qtc)
	}
	l.writer.WriteQTC(qtc)
	l.qtcList.QTCAdded(qtc)
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
	for _, qtc := range l.allQTCs() {
		err := writer.WriteQTC(qtc)
		if err != nil {
			return fmt.Errorf("cannot write QTC %v: %w", qtc, err)
		}
	}
	return nil
}

type nullWriter struct{}

func (d *nullWriter) WriteQSO(core.QSO) error {
	return nil
}

func (d *nullWriter) WriteQTC(core.QTC) error {
	return nil
}
