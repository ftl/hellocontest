package core

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
)

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
	DXCC         dxcc.Prefix
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

// Workmode is either search&pounce or run.
type Workmode int

// All work modes.
const (
	SearchPounce Workmode = iota
	Run
)

// EntryField represents an entry field in the visual part.
type EntryField int

// The entry fields.
const (
	CallsignField EntryField = iota
	TheirReportField
	TheirNumberField
	TheirXchangeField
	MyReportField
	MyNumberField
	MyXchangeField
	BandField
	ModeField
	OtherField
)

// KeyerValues contains the values that can be used as variables in the keyer templates.
type KeyerValues struct {
	TheirCall string
	MyNumber  QSONumber
	MyReport  RST
	MyXchange string
}

// AnnotatedCallsign contains a callsign with additional information retrieved from databases and the logbook.
type AnnotatedCallsign struct {
	Callsign   callsign.Callsign
	Duplicate  bool
	Worked     bool
	ExactMatch bool
}

type Score struct {
	ScorePerBand map[Band]BandScore
	TotalScore   BandScore
	OverallScore BandScore
}

func (s Score) String() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Band SpcQ CtyQ ConQ OthQ Pts     P/Q  CQ ITU Cty Xch Mult Q/M  Result \n")
	fmt.Fprintf(buf, "----------------------------------------------------------------------\n")
	for _, band := range Bands {
		if score, ok := s.ScorePerBand[band]; ok {
			fmt.Fprintf(buf, "%4s %s\n", band, score)
		}
	}
	fmt.Fprintf(buf, "----------------------------------------------------------------------\n")
	fmt.Fprintf(buf, "Tot  %s\n", s.TotalScore)
	fmt.Fprintf(buf, "Ovr  %s\n", s.OverallScore)
	return buf.String()
}

type BandScore struct {
	SpecificCountryQSOs int
	SameCountryQSOs     int
	SameContinentQSOs   int
	OtherQSOs           int
	Points              int
	CQZones             int
	ITUZones            int
	PrimaryPrefixes     int
	XchangeValues       int
	Multis              int
}

func (s BandScore) String() string {
	return fmt.Sprintf("%4d %4d %4d %4d %7d %4.1f %2d %3d %3d %3d %4d %4.1f %7d", s.SpecificCountryQSOs, s.SameCountryQSOs, s.SameContinentQSOs, s.OtherQSOs, s.Points, s.PointsPerQSO(), s.CQZones, s.ITUZones, s.PrimaryPrefixes, s.XchangeValues, s.Multis, s.QSOsPerMulti(), s.Result())
}

func (s *BandScore) Add(other BandScore) {
	s.SpecificCountryQSOs += other.SpecificCountryQSOs
	s.SameCountryQSOs += other.SameCountryQSOs
	s.SameContinentQSOs += other.SameContinentQSOs
	s.OtherQSOs += other.OtherQSOs
	s.Points += other.Points
	s.CQZones += other.CQZones
	s.ITUZones += other.ITUZones
	s.PrimaryPrefixes += other.PrimaryPrefixes
	s.XchangeValues += other.XchangeValues
	s.Multis += other.Multis
}

func (s *BandScore) QSOs() int {
	return s.SpecificCountryQSOs + s.SameCountryQSOs + s.SameContinentQSOs + s.OtherQSOs
}

func (s *BandScore) PointsPerQSO() float64 {
	qsos := s.QSOs()
	if qsos == 0 {
		return 0
	}
	return float64(s.Points) / float64(qsos)
}

func (s *BandScore) QSOsPerMulti() float64 {
	qsos := s.QSOs()
	if qsos == 0 {
		return 0
	}
	if s.Multis == 0 {
		return 0
	}
	return float64(qsos) / float64(s.Multis)
}

func (s *BandScore) Result() int {
	return s.Points * s.Multis
}

type MultiplierState int

const (
	NoMultiplier MultiplierState = iota
	NewBandMultiplier
	NewMultiplier
)

// Hour is used as reference to calculate the number of QSOs per hour.
type Hour time.Time

// HourOf returns the given time to the hour.
func HourOf(t time.Time) Hour {
	return Hour(time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		0,
		0,
		0,
		t.Location(),
	))
}

// QSOsPerHour is the rate of QSOs per one hour
type QSOsPerHour int

// QSOsPerHours contains the complete QSO rate statistic mapping each Hour in the contest to the rate of QSOs within this Hour
type QSOsPerHours map[Hour]QSOsPerHour

// QSORate contains all statistics regarding the rate of QSOs in a contest.
type QSORate struct {
	LastHourRate QSOsPerHour
	Last5MinRate QSOsPerHour
	QSOsPerHours QSOsPerHours
}
