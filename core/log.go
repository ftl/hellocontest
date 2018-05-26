package core

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ftl/hamradio/callsign"
)

// Log describes the functionality of the log component.
type Log interface {
	SetView(LogView)

	GetNextNumber() int
	Log(QSO)
	Find(callsign.Callsign) (QSO, bool)
	GetQsosByMyNumber() []QSO
}

// LogView represents the visual part of the log.
type LogView interface {
	SetLog(Log)

	UpdateAllRows([]QSO)
	RowAdded(QSO)
}

// QSO contains the details about one radio contact.
type QSO struct {
	Callsign     callsign.Callsign
	Time         time.Time
	Band         Band
	MyReport     RST
	MyNumber     int
	TheirReport  RST
	TheirNumber  int
	LogTimestamp time.Time
}

func (qso *QSO) String() string {
	return fmt.Sprintf("%s|%s|%s|%s|%03d|%s|%03d", qso.Time.Format("15:04"), qso.Callsign.String(), qso.Band, qso.MyReport, qso.MyNumber, qso.TheirReport, qso.TheirNumber)
}

// Band represents an amateur radio band.
type Band string

// All HF bands.
const (
	Band160m Band = "160m"
	Band80m  Band = "80m"
	Band60m  Band = "60m"
	Band40m  Band = "40m"
	Band30m  Band = "30m"
	Band20m  Band = "20m"
	Band17m  Band = "17m"
	Band15m  Band = "15m"
	Band12m  Band = "12m"
	Band10m  Band = "10m"
)

func (band *Band) String() string {
	return string(*band)
}

// RST represents a signal report using the "Readability/Signalstrength/Tone" system.
type RST string

var parseRSTExpression = regexp.MustCompile("\\b[1-5]([1-9]([1-9])?)?\\b")

// ParseRST parses the given string for a report and returns the parsed RST value.
func ParseRST(s string) (RST, error) {
	normalized := strings.TrimSpace(s)
	length := len(normalized)
	if length == 0 {
		return RST(""), fmt.Errorf("The report in RST notation must not be empty")
	}
	if length > 3 {
		return RST(""), fmt.Errorf("%q is not a valid report in RST notation", s)
	}
	if !parseRSTExpression.MatchString(normalized) {
		return RST(""), fmt.Errorf("%q is not a valid report in RST notation", s)
	}
	return RST(normalized), nil
}

func (rst *RST) String() string {
	return string(*rst)
}

// NewLog creates a new empty log.
func NewLog(clock Clock) Log {
	return &log{
		clock: clock,
		qsos:  make([]QSO, 0, 1000),
		view:  &nullLogView{},
	}
}

type log struct {
	clock        Clock
	qsos         []QSO
	myLastNumber int

	view LogView
}

func (l *log) SetView(view LogView) {
	l.view = view
	l.view.SetLog(l)
	l.view.UpdateAllRows(l.qsos)
}

func (l *log) GetNextNumber() int {
	return l.myLastNumber + 1
}

func (l *log) Log(qso QSO) {
	qso.LogTimestamp = l.clock.Now()
	l.qsos = append(l.qsos, qso)
	l.myLastNumber = int(math.Max(float64(l.myLastNumber), float64(qso.MyNumber)))
	l.view.RowAdded(qso)
}

func (l *log) Find(callsign callsign.Callsign) (QSO, bool) {
	for i := len(l.qsos) - 1; i >= 0; i-- {
		qso := l.qsos[i]
		if callsign == qso.Callsign {
			return qso, true
		}
	}
	return QSO{}, false
}

func (l *log) GetQsosByMyNumber() []QSO {
	result := make([]QSO, len(l.qsos))
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

func (d *nullLogView) SetLog(Log)          {}
func (d *nullLogView) UpdateAllRows([]QSO) {}
func (d *nullLogView) RowAdded(QSO)        {}
