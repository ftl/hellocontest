package model

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const MaxQTCsPerCall = 10 // TODO: this should be configurable through the contest definition

// QTC represents one QTC. If the QTC is not yet logged, it contains no timestamp, frequency,
// and mode information
type QTC struct {
	// Basic Information
	Kind          QTCKind
	QSONumber     QSONumber // used only for SentQTC, references the QSO in my log
	TheirCallsign Callsign
	Header        QTCHeader

	// Log Information, must be set, when actually sending the QTC
	Timestamp time.Time
	Frequency Frequency
	Band      Band
	Mode      Mode
	Confirmed bool // this field is not persistent, it is only used during the QTC workflow to track which QTCs have been confirmed

	// Data from referenced QSO, either from my log, or from their log
	QTCTime     QTCTime
	QTCCallsign Callsign
	QTCNumber   QSONumber
}

func (q QTC) String() string {
	return fmt.Sprintf("%s|%-10s|%5.0fkHz|%4s|%-4s|%d|%s|%s|%d", q.Timestamp.Format("15:04"), q.TheirCallsign.String(), q.Frequency/1000.0, q.Band, q.Mode, q.Kind, q.QTCTime.String(), q.QTCCallsign, q.QTCNumber)
}

func (q QTC) SeriesKey() string {
	return fmt.Sprintf("%s|%d|%d", q.TheirCallsign, q.Header.SeriesNumber, q.Header.QTCCount)
}

func (q QTC) WasTransmitted() bool {
	return !q.Timestamp.IsZero()
}

func (q QTC) VerifyComplete() error {
	switch {
	case q.Timestamp.IsZero():
		return fmt.Errorf("the timestamp must not be zero")
	case q.Band == NoBand:
		return fmt.Errorf("the band must not be empty")
	case q.Mode == NoMode:
		return fmt.Errorf("the mode must not be empty")
	case q.Frequency == 0:
		return fmt.Errorf("the frequency must not be zero")
	case (q.Kind == SentQTC) && (q.QSONumber <= 0):
		return fmt.Errorf("the QSO number must be greater than zero")
	default:
		return nil
	}
}

type QTCMode int

const (
	ReceiveQTC QTCMode = iota
	ProvideQTC
)

type QTCKind int

const (
	ReceivedQTC QTCKind = iota
	SentQTC
)

type QTCHeader struct {
	SeriesNumber int
	QTCCount     int
}

var qtcHeaderFormat = regexp.MustCompile(`([[:digit:]]+)/((1)?[[:digit:]]$)`)

func ParseQTCHeader(s string) (QTCHeader, error) {
	result := QTCHeader{}

	fields := qtcHeaderFormat.FindStringSubmatch(s)
	if len(fields) < 3 {
		return QTCHeader{}, fmt.Errorf("%q is not a valid QTC header", s)
	}

	result.SeriesNumber, _ = strconv.Atoi(fields[1])
	result.QTCCount, _ = strconv.Atoi(fields[2])

	return result, nil
}

func (h QTCHeader) String() string {
	return fmt.Sprintf("%d/%d", h.SeriesNumber, h.QTCCount)
}

func QTCByTimestamp(a, b QTC) int {
	switch {
	case a.Timestamp.Before(b.Timestamp):
		return -1
	case a.Timestamp.After(b.Timestamp):
		return 1
	default:
		return 0
	}
}

type QTCTime struct {
	Hour   int
	Minute int
}

var ZeroQTCTime = QTCTime{0, 0}

// ParseQTCTime parses the given string as QTC time (hhmm). The additional reference time
// is used, if only the minutes are given in the string. In that case, the hours are taken
// from the reference.
func ParseQTCTime(s string, reference QTCTime) (QTCTime, error) {
	l := len(s)
	if l < 1 || l > 4 {
		return ZeroQTCTime, fmt.Errorf("cannot parse QTC time: %q is invalid", s)
	}

	result := QTCTime{}
	var err error

	switch {
	case l < 3:
		result.Hour = reference.Hour
	default:
		result.Hour, err = strconv.Atoi(string(s[0 : l-2]))
	}
	if err != nil {
		return ZeroQTCTime, fmt.Errorf("cannot parse QTC time: %q is not a number", s)
	}
	if result.Hour < 0 || result.Hour > 23 {
		return ZeroQTCTime, fmt.Errorf("cannot parse QTC time: %d is not valid for the hour section", result.Hour)
	}

	result.Minute, err = strconv.Atoi(s[max(l-2, 0):l])
	if err != nil {
		return ZeroQTCTime, fmt.Errorf("cannot parse QTC time: %q is not a number", s)
	}
	if result.Minute < 0 || result.Minute > 59 {
		return ZeroQTCTime, fmt.Errorf("cannot parse QTC time: %d is not valid for the minute section", result.Minute)
	}

	return result, nil
}

func QTCTimeFromTimestamp(ts time.Time) QTCTime {
	return QTCTime{
		Hour:   ts.Hour(),
		Minute: ts.Minute(),
	}
}

func (t QTCTime) String() string {
	return fmt.Sprintf("%02d%02d", t.Hour, t.Minute)
}

func (t QTCTime) ShortString() string {
	return fmt.Sprintf("%02d", t.Minute)
}

const NoQTCIndex int = -1

type QTCSeries struct {
	TheirCallsign Callsign
	Header        QTCHeader
	QTCs          []QTC
}

func NewQTCSeries(seriesNumber int, qtcs []QTC) (QTCSeries, error) {
	if len(qtcs) > MaxQTCsPerCall {
		return QTCSeries{}, fmt.Errorf("a QTC series must not have more than %d QTCs", MaxQTCsPerCall)
	}
	if len(qtcs) < 1 {
		return QTCSeries{}, fmt.Errorf("a QTC series must have at least one QTC")
	}

	result := QTCSeries{
		Header: QTCHeader{SeriesNumber: seriesNumber, QTCCount: len(qtcs)},
	}
	for i := range qtcs {
		qtcs[i].Header = result.Header
	}
	result.QTCs = qtcs
	result.TheirCallsign = result.theirCallsign()
	if result.TheirCallsign == NoCallsign {
		return QTCSeries{}, fmt.Errorf("a QTC series must have a consistent value for 'their callsign'")
	}
	return result, nil
}

func NewReceivingQTCSeries(theirCallsign Callsign) QTCSeries {
	result := QTCSeries{
		TheirCallsign: theirCallsign,
	}
	return result
}

func (s *QTCSeries) theirCallsign() Callsign {
	if len(s.QTCs) == 0 {
		return NoCallsign
	}

	theirCallsign := s.QTCs[0].TheirCallsign
	theirCallsignString := theirCallsign.String()
	for _, qtc := range s.QTCs {
		if qtc.TheirCallsign.String() != theirCallsignString {
			return NoCallsign
		}
	}
	return theirCallsign
}

func (s *QTCSeries) Key() string {
	return fmt.Sprintf("%s|%d|%d", s.theirCallsign(), s.Header.SeriesNumber, s.Header.QTCCount)
}

func (s *QTCSeries) IsPrepared() bool {
	return s.TheirCallsign != NoCallsign && len(s.QTCs) > 0
}

func (s *QTCSeries) IsComplete() bool {
	return s.IsPrepared() && s.Header.QTCCount > 0 && len(s.QTCs) == s.Header.QTCCount
}

func (s *QTCSeries) IsValidQTCIndex(index int) bool {
	return index >= 0 && index < s.Header.QTCCount
}

func (s *QTCSeries) IsLastQTCIndex(index int) bool {
	return index >= 0 && index == s.Header.QTCCount-1
}

func (s *QTCSeries) SetData(index int, qtc QTC) {
	if index >= 0 && index < len(s.QTCs) {
		s.QTCs[index] = qtc
	} else {
		s.QTCs = append(s.QTCs, qtc)
	}
}

// QTCWorkflowPhase represents the phases in the QTC exchange workflow
type QTCWorkflowPhase int

const (
	QTCNone QTCWorkflowPhase = iota
	QTCStart
	QTCExchangeHeader
	QTCExchangeData
	QTCFinish
)

// QTCField represents the currently active entry field in the QTC exchange workflow
type QTCField string

const (
	QTCNoneField      QTCField = ""
	QTCHeaderField    QTCField = "header"
	QTCTimestampField QTCField = "timestamp"
	QTCCallsignField  QTCField = "callsign"
	QTCExchangeField  QTCField = "exchange"
)

// QSODataState represents the current state of the entered QSO data
type QSODataState int

const (
	QSODataEmpty QSODataState = iota
	QSODataInvalid
	QSODataValid
)
