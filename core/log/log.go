package log

import (
	logger "log"
	"math"
	"sort"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/pkg/errors"
)

// New creates a new empty log.
func New(clock core.Clock) core.Log {
	return &log{
		clock:             clock,
		qsos:              make([]core.QSO, 0, 1000),
		view:              &nullLogView{},
		rowAddedListeners: make([]core.RowAddedListener, 0),
	}
}

// Load creates a new log and loads it with the entries from the given reader.
func Load(clock core.Clock, reader core.Reader) (core.Log, error) {
	logger.Print("Loading QSOs")
	log := &log{
		clock:             clock,
		view:              &nullLogView{},
		rowAddedListeners: make([]core.RowAddedListener, 0),
	}

	var err error
	log.qsos, err = reader.ReadAll()
	if err != nil {
		return nil, err
	}

	log.myLastNumber = lastNumber(log.qsos)
	return log, nil
}

func lastNumber(qsos []core.QSO) int {
	lastNumber := 0
	for _, qso := range qsos {
		lastNumber = int(math.Max(float64(lastNumber), float64(qso.MyNumber)))
	}
	return lastNumber
}

type log struct {
	clock        core.Clock
	qsos         []core.QSO
	myLastNumber int

	view              core.LogView
	rowAddedListeners []core.RowAddedListener
}

func (l *log) SetView(view core.LogView) {
	l.view = view
	l.view.SetLog(l)
	l.view.UpdateAllRows(l.qsos)
}

func (l *log) OnRowAdded(listener core.RowAddedListener) {
	l.rowAddedListeners = append(l.rowAddedListeners, listener)
}

func (l *log) ClearRowAddedListeners() {
	l.rowAddedListeners = make([]core.RowAddedListener, 0)
}

func (l *log) emitRowAdded(qso core.QSO) {
	for _, listener := range l.rowAddedListeners {
		err := listener(qso)
		if err != nil {
			logger.Printf("Error on rowAdded: %T, %v", listener, err)
		}
	}
}

func (l *log) NextNumber() core.QSONumber {
	return core.QSONumber(l.myLastNumber + 1)
}

func (l *log) LastBand() core.Band {
	if len(l.qsos) == 0 {
		return core.NoBand
	}
	return l.qsos[len(l.qsos)-1].Band
}

func (l *log) LastMode() core.Mode {
	if len(l.qsos) == 0 {
		return core.NoMode
	}
	return l.qsos[len(l.qsos)-1].Mode
}

func (l *log) Log(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	l.qsos = append(l.qsos, qso)
	l.myLastNumber = int(math.Max(float64(l.myLastNumber), float64(qso.MyNumber)))
	l.view.RowAdded(qso)
	l.emitRowAdded(qso)
	logger.Printf("QSO added: %s", qso.String())
}

func (l *log) Find(callsign callsign.Callsign) (core.QSO, bool) {
	for i := len(l.qsos) - 1; i >= 0; i-- {
		qso := l.qsos[i]
		if callsign == qso.Callsign {
			return qso, true
		}
	}
	return core.QSO{}, false
}

func (l *log) FindAll(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	result := make([]core.QSO, 0)
	for _, qso := range l.qsos {
		if callsign != qso.Callsign {
			continue
		}
		if band != core.NoBand && band != qso.Band {
			continue
		}
		if mode != core.NoMode && mode != qso.Mode {
			continue
		}
		result = append(result, qso)

	}
	return result
}

func (l *log) QsosOrderedByMyNumber() []core.QSO {
	return byMyNumber(l.qsos)
}

func (l *log) UniqueQsosOrderedByMyNumber() []core.QSO {
	return byMyNumber(unique(l.qsos))
}

func byMyNumber(qsos []core.QSO) []core.QSO {
	result := make([]core.QSO, len(qsos))
	copy(result, qsos)
	sort.Slice(result, func(i, j int) bool {
		if qsos[i].MyNumber == qsos[j].MyNumber {
			return qsos[i].LogTimestamp.Before(qsos[j].LogTimestamp)
		}
		return qsos[i].MyNumber < qsos[j].MyNumber
	})
	return result
}

func unique(qsos []core.QSO) []core.QSO {
	result := make([]core.QSO, 0, len(qsos))
	index := make(map[callsign.Callsign]int)
	for _, qso := range qsos {
		i, ok := index[qso.Callsign]
		if !ok {
			i = len(result)
			result = append(result, qso)
			index[qso.Callsign] = i
			continue
		}
		if qso.LogTimestamp.After(result[i].LogTimestamp) {
			result[i] = qso
		}
	}
	return result
}

func (l *log) WriteAll(writer core.Writer) error {
	for _, qso := range l.qsos {
		err := writer.Write(qso)
		if err != nil {
			return errors.Wrapf(err, "cannot write QSO %v", qso)
		}
	}
	return nil
}

type nullLogView struct{}

func (d *nullLogView) SetLog(core.Log)          {}
func (d *nullLogView) UpdateAllRows([]core.QSO) {}
func (d *nullLogView) RowAdded(core.QSO)        {}
