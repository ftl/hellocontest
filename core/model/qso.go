package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
)

// QSO contains the details about one radio contact.
type QSO struct {
	Callsign      Callsign
	Time          time.Time
	Frequency     Frequency
	Band          Band
	Mode          Mode
	MyReport      RST
	MyNumber      QSONumber
	MyExchange    []string
	TheirReport   RST
	TheirNumber   QSONumber
	TheirExchange []string
	LogTimestamp  time.Time
	DXCC          dxcc.Prefix
	Points        int
	Multis        int
	Duplicate     bool
	Workmode      Workmode
}

func (qso *QSO) String() string {
	return fmt.Sprintf("%s|%-10s|%5.0fkHz|%4s|%-4s|%s|%s|%s|%s|%s|%s|%2d|%2d|%t|%s", qso.Time.Format("15:04"), qso.Callsign.String(), qso.Frequency/1000.0, qso.Band, qso.Mode, qso.MyReport, qso.MyNumber.String(), strings.Join(qso.MyExchange, " "), qso.TheirReport, qso.TheirNumber.String(), strings.Join(qso.TheirExchange, " "), qso.Points, qso.Multis, qso.Duplicate, qso.Workmode.String())
}

type Callsign = callsign.Callsign

var NoCallsign = callsign.NoCallsign

func MustParseCallsign(s string) Callsign {
	return callsign.MustParse(s)
}

func ParseCallsign(s string) (Callsign, error) {
	return callsign.Parse(s)
}

// Frequency in Hz.
type Frequency float64

func (f Frequency) String() string {
	return fmt.Sprintf("%.0fHz", float64(f))
}

func (f Frequency) DistanceTo(other Frequency) Frequency {
	result := f - other
	if result < 0 {
		return -result
	}
	return result
}

func IsFrequencyJump(from Frequency, to Frequency) bool {
	return from.DistanceTo(to) > frequencyJumpThreshold
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

// Bands are all supported HF bands.
var Bands = []Band{Band160m, Band80m, Band60m, Band40m, Band30m, Band20m, Band17m, Band15m, Band12m, Band10m}

func (band *Band) String() string {
	return string(*band)
}

// ParseBand parses a string into a HF band value
func ParseBand(s string) (Band, error) {
	for _, band := range Bands {
		if string(band) == s {
			return band, nil
		}
	}
	return NoBand, fmt.Errorf("%q is not a supported HF band", s)
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

// ParseMode parses a string into a HF Mode value
func ParseMode(s string) (Mode, error) {
	for _, mode := range Modes {
		if string(mode) == s {
			return mode, nil
		}
	}
	return NoMode, fmt.Errorf("%q is not a supported mode", s)
}

// RST represents a signal report using the "Readability/Signalstrength/Tone" system.
type RST string

func (rst RST) String() string {
	return string(rst)
}

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

// QSONumber is the unique number of a QSO in the log, either on my or on their side.
type QSONumber int

func (nr *QSONumber) String() string {
	return fmt.Sprintf("%03d", *nr)
}

// Workmode is either search&pounce or run.
type Workmode int

// All work modes.
const (
	UnknownWorkmode Workmode = iota
	SearchPounce
	Run
)

func (workmode *Workmode) String() string {
	switch *workmode {
	case SearchPounce:
		return "S"
	case Run:
		return "R"
	default:
		return ""
	}
}
