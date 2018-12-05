package core

import (
	"fmt"
	"time"

	"github.com/ftl/hamradio/callsign"
)

// Log describes the functionality of the log component.
type Log interface {
	SetView(LogView)
	OnRowAdded(RowAddedListener)

	GetNextNumber() QSONumber
	Log(QSO)
	Find(callsign.Callsign) (QSO, bool)
	GetQsosByMyNumber() []QSO
}

// Reader reads log entries.
type Reader interface {
	ReadAll() ([]QSO, error)
}

// Writer writes log entries.
type Writer interface {
	Write(QSO) error
}

// Store allows to read and write log entries.
type Store interface {
	Reader
	Writer
}

// RowAddedListener is notified when a new row is added to the log.
type RowAddedListener func(QSO) error

func (l RowAddedListener) Write(qso QSO) error {
	return l(qso)
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

func (mode *Mode) String() string {
	return string(*mode)
}

// RST represents a signal report using the "Readability/Signalstrength/Tone" system.
type RST string

func (rst *RST) String() string {
	return string(*rst)
}

// QSONumber is the unique number of a QSO in the log, either on my or on their side.
type QSONumber int

func (nr *QSONumber) String() string {
	return fmt.Sprintf("%03d", *nr)
}

// Clock represents a source of the current time.
type Clock interface {
	Now() time.Time
}

// KeyerValues contains the values that can be used as variables in the keyer templates.
type KeyerValues struct {
	MyCall    callsign.Callsign
	TheirCall string
	MyNumber  QSONumber
	MyReport  RST
}

// KeyerValueProvider provides the variable values for the Keyer templates on demand.
type KeyerValueProvider func() KeyerValues

// CWClient defines the interface used by the Keyer to output the CW.
type CWClient interface {
	Send(text string)
}

// Keyer represents the component that sends prepared CW texts using text/templates.
type Keyer interface {
	SetTemplate(index int, pattern string) error
	GetTemplate(index int) string
	GetText(index int) (string, error)
	Send(index int) error
}
