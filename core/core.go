package core

import (
	"fmt"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/locator"
)

// Log describes the functionality of the log component.
type Logbook interface {
	SetView(LogbookView)
	OnRowAdded(RowAddedListener)
	ClearRowAddedListeners()
	OnRowSelected(RowSelectedListener)
	ClearRowSelectedListeners()

	NextNumber() QSONumber
	LastBand() Band
	LastMode() Mode
	Select(int)
	Log(QSO)
	Find(callsign.Callsign) (QSO, bool)
	FindAll(callsign.Callsign, Band, Mode) []QSO
	QsosOrderedByMyNumber() []QSO
	UniqueQsosOrderedByMyNumber() []QSO
	WriteAll(Writer) error
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
	Clear() error
}

// RowAddedListener is notified when a new row is added to the log.
type RowAddedListener func(QSO) error

// RowSelectedListener is notified when a row is selected in the log view.
type RowSelectedListener func(QSO)

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
	MyXchange    string
	TheirReport  RST
	TheirNumber  QSONumber
	TheirXchange string
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
	NoMode      Mode = ""
	ModeCW      Mode = "CW"
	ModeSSB     Mode = "SSB"
	ModeFM      Mode = "FM"
	ModeRTTY    Mode = "RTTY"
	ModeDigital Mode = "DIGI"
)

// Modes are all relevant modes.
var Modes = []Mode{ModeCW, ModeSSB, ModeFM, ModeRTTY, ModeDigital}

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

// Quitter allows to quit the application. This interfaces is used to call the actual application framework to quit.
type Quitter interface {
	Quit()
}

// Configuration provides read access to the configuration data.
type Configuration interface {
	MyCall() callsign.Callsign
	MyLocator() locator.Locator

	EnterTheirNumber() bool
	EnterTheirXchange() bool
	CabrilloQSOTemplate() string
	AllowMultiBand() bool
	AllowMultiMode() bool

	KeyerHost() string
	KeyerPort() int
	KeyerSPPatterns() []string
	KeyerRunPatterns() []string
}

// KeyerValues contains the values that can be used as variables in the keyer templates.
type KeyerValues struct {
	TheirCall string
	MyNumber  QSONumber
	MyReport  RST
	MyXchange string
}

// KeyerValueProvider provides the variable values for the Keyer templates on demand.
type KeyerValueProvider func() KeyerValues

// CWClient defines the interface used by the Keyer to output the CW.
type CWClient interface {
	Connect() error
	Disconnect()
	IsConnected() bool
	Speed(int)
	Send(text string)
	Abort()
}
