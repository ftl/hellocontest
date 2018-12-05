package log

import (
	logger "log"
	"math"
	"sort"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
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

func (l *log) emitRowAdded(qso core.QSO) {
	for _, listener := range l.rowAddedListeners {
		err := listener(qso)
		if err != nil {
			logger.Printf("Error on rowAdded: %T, %v", listener, err)
		}
	}
}

func (l *log) GetNextNumber() core.QSONumber {
	return core.QSONumber(l.myLastNumber + 1)
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

func (l *log) GetQsosByMyNumber() []core.QSO {
	result := make([]core.QSO, len(l.qsos))
	copy(result, l.qsos)
	sort.Slice(result, func(i, j int) bool {
		if l.qsos[i].MyNumber == l.qsos[j].MyNumber {
			return l.qsos[i].LogTimestamp.Before(l.qsos[j].LogTimestamp)
		}
		return l.qsos[i].MyNumber < l.qsos[j].MyNumber
	})
	return result
}

type nullLogView struct{}

func (d *nullLogView) SetLog(core.Log)          {}
func (d *nullLogView) UpdateAllRows([]core.QSO) {}
func (d *nullLogView) RowAdded(core.QSO)        {}
