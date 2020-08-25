package logbook

import (
	logger "log"
	"math"
	"sort"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/pkg/errors"
)

// New creates a new empty logbook.
func New(clock core.Clock) core.Logbook {
	return &logbook{
		clock:             clock,
		qsos:              make([]core.QSO, 0, 1000),
		view:              &nullLogbookView{},
		rowAddedListeners: make([]core.RowAddedListener, 0),
	}
}

// Load creates a new log and loads it with the entries from the given reader.
func Load(clock core.Clock, reader core.Reader) (core.Logbook, error) {
	logger.Print("Loading QSOs")
	logbook := &logbook{
		clock:             clock,
		view:              &nullLogbookView{},
		rowAddedListeners: make([]core.RowAddedListener, 0),
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

type logbook struct {
	clock           core.Clock
	qsos            []core.QSO
	myLastNumber    int
	ignoreSelection bool

	view                 core.LogbookView
	rowAddedListeners    []core.RowAddedListener
	rowSelectedListeners []core.RowSelectedListener
}

func (l *logbook) SetView(view core.LogbookView) {
	l.ignoreSelection = true
	defer func() { l.ignoreSelection = false }()

	l.view = view
	l.view.SetLogbook(l)
	l.view.UpdateAllRows(l.qsos)
}

func (l *logbook) OnRowAdded(listener core.RowAddedListener) {
	l.rowAddedListeners = append(l.rowAddedListeners, listener)
}

func (l *logbook) ClearRowAddedListeners() {
	l.rowAddedListeners = make([]core.RowAddedListener, 0)
}

func (l *logbook) emitRowAdded(qso core.QSO) {
	for _, listener := range l.rowAddedListeners {
		err := listener(qso)
		if err != nil {
			logger.Printf("Error on rowAdded: %T, %v", listener, err)
		}
	}
}

func (l *logbook) OnRowSelected(listener core.RowSelectedListener) {
	l.rowSelectedListeners = append(l.rowSelectedListeners, listener)
}

func (l *logbook) ClearRowSelectedListeners() {
	l.rowSelectedListeners = make([]core.RowSelectedListener, 0)
}

func (l *logbook) emitRowSelected(qso core.QSO) {
	for _, listener := range l.rowSelectedListeners {
		listener(qso)
	}
}

func (l *logbook) Select(i int) {
	if i < 0 || i >= len(l.qsos) {
		logger.Printf("invalid QSO index %d", i)
		return
	}
	if l.ignoreSelection {
		return
	}
	qso := l.qsos[i]
	l.emitRowSelected(qso)
}

func (l *logbook) NextNumber() core.QSONumber {
	return core.QSONumber(l.myLastNumber + 1)
}

func (l *logbook) LastBand() core.Band {
	if len(l.qsos) == 0 {
		return core.NoBand
	}
	return l.qsos[len(l.qsos)-1].Band
}

func (l *logbook) LastMode() core.Mode {
	if len(l.qsos) == 0 {
		return core.NoMode
	}
	return l.qsos[len(l.qsos)-1].Mode
}

func (l *logbook) Log(qso core.QSO) {
	l.ignoreSelection = true
	defer func() { l.ignoreSelection = false }()

	qso.LogTimestamp = l.clock.Now()
	l.qsos = append(l.qsos, qso)
	l.myLastNumber = int(math.Max(float64(l.myLastNumber), float64(qso.MyNumber)))
	l.view.RowAdded(qso)
	l.emitRowAdded(qso)
	logger.Printf("QSO added: %s", qso.String())
}

func (l *logbook) Find(callsign callsign.Callsign) (core.QSO, bool) {
	checkedNumbers := make(map[core.QSONumber]bool)
	for i := len(l.qsos) - 1; i >= 0; i-- {
		qso := l.qsos[i]
		if checkedNumbers[qso.MyNumber] {
			continue
		}
		checkedNumbers[qso.MyNumber] = true

		if callsign == qso.Callsign {
			return qso, true
		}
	}
	return core.QSO{}, false
}

func (l *logbook) FindAll(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	checkedNumbers := make(map[core.QSONumber]bool)
	result := make([]core.QSO, 0)
	for i := len(l.qsos) - 1; i >= 0; i-- {
		qso := l.qsos[i]
		if checkedNumbers[qso.MyNumber] {
			continue
		}
		checkedNumbers[qso.MyNumber] = true

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

func (l *logbook) QsosOrderedByMyNumber() []core.QSO {
	return byMyNumber(l.qsos)
}

func (l *logbook) UniqueQsosOrderedByMyNumber() []core.QSO {
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

func (l *logbook) WriteAll(writer core.Writer) error {
	for _, qso := range l.qsos {
		err := writer.Write(qso)
		if err != nil {
			return errors.Wrapf(err, "cannot write QSO %v", qso)
		}
	}
	return nil
}

type nullLogbookView struct{}

func (d *nullLogbookView) SetLogbook(core.Logbook)  {}
func (d *nullLogbookView) UpdateAllRows([]core.QSO) {}
func (d *nullLogbookView) RowAdded(core.QSO)        {}
