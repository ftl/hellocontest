package core

import (
	"fmt"
	logger "log"
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

	GetNextNumber() QSONumber
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
	Mode         Mode
	MyReport     RST
	MyNumber     QSONumber
	TheirReport  RST
	TheirNumber  QSONumber
	LogTimestamp time.Time
}

func (qso *QSO) String() string {
	return fmt.Sprintf("%s|%-10s|%4s|%-4s|%s|%s|%s|%s", qso.Time.Format("15:04"), qso.Callsign.String(), qso.Band, qso.Mode, qso.MyReport, qso.MyNumber.String(), qso.TheirReport, qso.TheirNumber.String())
}

// Band represents an amateur radio band.
type Band string

// All HF bands.
const (
	NoBand   Band = ""
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

// Bands are all HF bands.
var Bands = []Band{Band160m, Band80m, Band60m, Band40m, Band30m, Band20m, Band17m, Band15m, Band12m, Band10m}

// ParseBand parses a string into a HF band value
func ParseBand(s string) (Band, error) {
	for _, band := range Bands {
		if string(band) == s {
			return band, nil
		}
	}
	return Band160m, fmt.Errorf("%q is not a supported HF band", s)
}

func (band *Band) String() string {
	return string(*band)
}

// Mode represents the mode.
type Mode string

// All relevant modes.
const (
	NoMode   Mode = ""
	ModeCW   Mode = "CW"
	ModeSSB  Mode = "SSB"
	ModeRTTY Mode = "RTTY"
)

// Modes are all relevant modes.
var Modes = []Mode{ModeCW, ModeSSB, ModeRTTY}

// ParseMode parses a string into a HF Mode value
func ParseMode(s string) (Mode, error) {
	for _, mode := range Modes {
		if string(mode) == s {
			return mode, nil
		}
	}
	return ModeCW, fmt.Errorf("%q is not a supported mode", s)
}

func (mode *Mode) String() string {
	return string(*mode)
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

// QSONumber is the unique number of a QSO in the log, either on my or on their side.
type QSONumber int

func (nr *QSONumber) String() string {
	return fmt.Sprintf("%03d", *nr)
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

func (l *log) GetNextNumber() QSONumber {
	return QSONumber(l.myLastNumber + 1)
}

func (l *log) Log(qso QSO) {
	qso.LogTimestamp = l.clock.Now()
	l.qsos = append(l.qsos, qso)
	l.myLastNumber = int(math.Max(float64(l.myLastNumber), float64(qso.MyNumber)))
	l.view.RowAdded(qso)
	logger.Printf("QSO added: %s", qso.String())
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
