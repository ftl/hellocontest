package core

import (
	"bytes"
	"fmt"
	"maps"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/constraints"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/latlon"
	"github.com/ftl/hamradio/locator"
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

func (rst RST) String() string {
	return string(rst)
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

// EntryField represents an entry field in the visual part.
type EntryField string

// The entry fields.
const (
	CallsignField EntryField = "callsign"
	BandField     EntryField = "band"
	ModeField     EntryField = "mode"
	OtherField    EntryField = "other"

	myExchangePrefix    string = "myExchange_"
	theirExchangePrefix string = "theirExchange_"
)

func (f EntryField) IsMyExchange() bool {
	return strings.HasPrefix(string(f), myExchangePrefix)
}

func (f EntryField) IsTheirExchange() bool {
	return strings.HasPrefix(string(f), theirExchangePrefix)
}

func IsExchangeField(name string) bool {
	return strings.HasPrefix(name, myExchangePrefix) || strings.HasPrefix(name, theirExchangePrefix)
}

func (f EntryField) ExchangeIndex() int {
	s := string(f)
	var a string
	switch {
	case strings.HasPrefix(s, myExchangePrefix):
		a = s[len(myExchangePrefix):]
	case strings.HasPrefix(s, theirExchangePrefix):
		a = s[len(theirExchangePrefix):]
	default:
		return -1
	}
	result, err := strconv.Atoi(a)
	if err != nil {
		return -1
	}
	return result
}

func (f EntryField) NextExchangeField() EntryField {
	s := string(f)
	var a string
	var prefix string
	switch {
	case strings.HasPrefix(s, myExchangePrefix):
		prefix = myExchangePrefix
		a = s[len(myExchangePrefix):]
	case strings.HasPrefix(s, theirExchangePrefix):
		prefix = theirExchangePrefix
		a = s[len(theirExchangePrefix):]
	default:
		return ""
	}
	i, err := strconv.Atoi(a)
	if err != nil {
		return ""
	}
	return EntryField(prefix + strconv.Itoa(i+1))
}

func MyExchangeField(index int) EntryField {
	return EntryField(fmt.Sprintf("%s%d", myExchangePrefix, index))
}

func TheirExchangeField(index int) EntryField {
	return EntryField(fmt.Sprintf("%s%d", theirExchangePrefix, index))
}

type ExchangeField struct {
	Field            EntryField
	CanContainSerial bool
	CanContainReport bool
	EmptyAllowed     bool
	Properties       conval.ExchangeField

	Short    string
	Name     string
	Hint     string
	ReadOnly bool
}

func DefinitionsToExchangeFields(fieldDefinitions []conval.ExchangeField, exchangeEntryField func(int) EntryField) []ExchangeField {
	result := make([]ExchangeField, 0, len(fieldDefinitions))
	for i, fieldDefinition := range fieldDefinitions {
		short := strings.Join(fieldDefinition.Strings(), "/")
		field := ExchangeField{
			Field:      exchangeEntryField(i + 1),
			Properties: fieldDefinition,
			Short:      short,
		}
		for _, property := range fieldDefinition {
			if property == conval.SerialNumberProperty {
				field.CanContainSerial = true
			}
			if property == conval.RSTProperty {
				field.CanContainReport = true
			}
			if property == conval.EmptyProperty {
				field.EmptyAllowed = true
			}
		}
		result = append(result, field)
	}
	return result
}

// KeyerValues contains the values that can be used as variables in the keyer templates.
type KeyerValues struct {
	TheirCall   string
	MyNumber    QSONumber
	MyReport    RST
	MyXchange   string
	MyExchange  string
	MyExchanges []string
	LastNumber  QSONumber
}

// FilterPlaceholder can be used as placeholder for a missed character in the callsign.
const FilterPlaceholder = "."

// ESMState represents the current state of the ESM state machine.
type ESMState int

const (
	ESMCallsignEmpty ESMState = iota
	ESMCallsignInvalid
	ESMCallsignValid
	ESMExchangeInvalid
	ESMExchangeValid
	ESMUnknown
)

// AnnotatedCallsign contains a callsign with additional information retrieved from databases and the logbook.
type AnnotatedCallsign struct {
	Callsign          Callsign
	Assembly          MatchingAssembly
	Duplicate         bool
	Worked            bool
	ExactMatch        bool
	Points            int
	Multis            int
	PredictedExchange []string
	Name              string
	UserText          string
	OnFrequency       bool

	Comparable any
	Compare    func(any, any) bool
}

func (c AnnotatedCallsign) LessThan(o AnnotatedCallsign) bool {
	if c.ExactMatch && !o.ExactMatch {
		return true
	}

	if c.Compare == nil {
		return false
	}
	if c.Comparable == nil || o.Comparable == nil {
		return false
	}

	return c.Compare(c.Comparable, o.Comparable)
}

type MatchingOperation int

const (
	Matching MatchingOperation = iota
	Insert
	Delete
	Substitute
	FalseFriend
)

type MatchingPart struct {
	OP    MatchingOperation
	Value string
}

type MatchingAssembly []MatchingPart

func (m MatchingAssembly) String() string {
	var result strings.Builder
	for _, match := range m {
		if match.OP != Delete {
			result.WriteString(match.Value)
		}
	}
	return result.String()
}

type Settings interface {
	Station() Station
	Contest() Contest
}

type Station struct {
	Callsign Callsign
	Operator Callsign
	Locator  locator.Locator
	Bandplan BandplanID
}

type Contest struct {
	Definition             *conval.Definition
	Name                   string
	ExchangeValues         []string
	GenerateSerialExchange bool
	GenerateReport         bool
	StartTime              time.Time
	EnableQTCs             bool

	MyExchangeFields         []ExchangeField
	MyReportExchangeField    ExchangeField
	MyNumberExchangeField    ExchangeField
	TheirExchangeFields      []ExchangeField
	TheirReportExchangeField ExchangeField
	TheirNumberExchangeField ExchangeField

	OperationModeSprint   bool
	CallHistoryFilename   string
	CallHistoryFieldNames []string

	QSOsGoal   int
	PointsGoal int
	MultisGoal int
}

func (c Contest) Bands() []Band {
	if c.Definition == nil {
		return nil
	}
	bands := c.Definition.Bands
	if len(bands) == 1 && bands[0] == conval.BandAll {
		bands = conval.AllHFBands
	}

	result := make([]Band, len(bands))
	for i, band := range c.Definition.Bands {
		result[i] = Band(band)
	}
	return result
}

func (c Contest) Modes() []Mode {
	if c.Definition == nil {
		return nil
	}
	result := make([]Mode, 0, len(c.Definition.Modes))
	for _, mode := range c.Definition.Modes {
		switch mode {
		case conval.ModeALL:
			return Modes
		case conval.ModeCW:
			result = append(result, ModeCW)
		case conval.ModeSSB:
			result = append(result, ModeSSB)
		case conval.ModeFM:
			result = append(result, ModeFM)
		case conval.ModeRTTY:
			result = append(result, ModeRTTY)
		case conval.ModeDigital:
			result = append(result, ModeDigital)
		}
	}
	return result
}

func (c Contest) Started(now time.Time) bool {
	if c.StartTime.IsZero() {
		return true
	}
	if c.Definition == nil {
		return true
	}
	if c.Definition.Duration == 0 {
		return true
	}

	return now.After(c.StartTime)
}

func (c Contest) Finished(now time.Time) bool {
	if c.StartTime.IsZero() {
		return false
	}
	if c.Definition == nil {
		return false
	}
	if c.Definition.Duration == 0 {
		return false
	}

	return now.After(c.StartTime.Add(c.Definition.Duration))
}

func (c Contest) Running(now time.Time) bool {
	return c.Started(now) && !c.Finished(now)
}

func (c *Contest) UpdateExchangeFields() {
	c.MyExchangeFields = nil
	c.MyReportExchangeField = ExchangeField{}
	c.MyNumberExchangeField = ExchangeField{}
	c.TheirExchangeFields = nil
	c.TheirReportExchangeField = ExchangeField{}
	c.TheirNumberExchangeField = ExchangeField{}

	if c.Definition == nil {
		return
	}

	fieldDefinitions := c.Definition.ExchangeFields()

	c.MyExchangeFields = DefinitionsToExchangeFields(fieldDefinitions, MyExchangeField)
	for i, field := range c.MyExchangeFields {
		switch {
		case field.Properties.Contains(conval.RSTProperty):
			c.MyReportExchangeField = field
		case field.Properties.Contains(conval.SerialNumberProperty):
			if c.GenerateSerialExchange {
				field.ReadOnly = true
				field.Short = "#"
				field.Hint = "Serial Number"
				c.MyExchangeFields[i] = field
			}
			c.MyNumberExchangeField = field
		}
	}

	c.TheirExchangeFields = DefinitionsToExchangeFields(fieldDefinitions, TheirExchangeField)
	for _, field := range c.TheirExchangeFields {
		switch {
		case field.Properties.Contains(conval.RSTProperty):
			c.TheirReportExchangeField = field
		case field.Properties.Contains(conval.SerialNumberProperty):
			c.TheirNumberExchangeField = field
		}
	}
}

type Radio struct {
	Name    string            `json:"name"`
	Type    RadioType         `json:"type"`
	Address string            `json:"address"`
	Keyer   string            `json:"keyer"`
	Options map[string]string `json:"options"`
}

type RadioType string

const (
	RadioTypeHamlib RadioType = "hamlib"
	RadioTypeTCI    RadioType = "tci"
)

type Keyer struct {
	Name    string    `json:"name"`
	Type    KeyerType `json:"type"`
	Address string    `json:"address"`
}

const RadioKeyer = "radio"

type KeyerType string

const (
	KeyerTypeCWDaemon KeyerType = "cwdaemon"
)

type KeyerSettings struct {
	WPM                   int      `json:"wpm"`
	Preset                string   `json:"preset"`
	SPMacros              []string `json:"sp_macros"`
	RunMacros             []string `json:"run_macros"`
	SPLabels              []string `json:"sp_labels"`
	RunLabels             []string `json:"run_labels"`
	ParrotIntervalSeconds int      `json:"parrot_interval_seconds"`
}

type KeyerPreset struct {
	Name      string   `json:"name"`
	SPMacros  []string `json:"sp_macros"`
	RunMacros []string `json:"run_macros"`
	SPLabels  []string `json:"sp_labels"`
	RunLabels []string `json:"run_labels"`
}

type RemoteServerSettings struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

type Summary struct {
	// comes from the contest definition
	ContestName  string
	CabrilloName string

	// comes from the contest settings
	StartTime   time.Time
	Callsign    Callsign
	MyExchanges string
	QTCsEnabled bool

	// has to be selected by the user in a dialog
	OperatorMode conval.OperatorMode
	Overlay      conval.Overlay
	PowerMode    conval.PowerMode
	Assisted     bool

	// comes from the score counter
	WorkedModes []string
	WorkedBands []string
	Score       Score
	TimeReport  TimeReport
}

func (s Summary) WorkingConditions() string {
	result := []string{}
	if s.OperatorMode != "" {
		result = append(result, fmt.Sprintf("%s operator", string(s.OperatorMode)))
	}
	if s.Overlay != "" {
		result = append(result, fmt.Sprintf("%s overlay", string(s.Overlay)))
	}
	if s.PowerMode != "" {
		result = append(result, fmt.Sprintf("%s power", string(s.PowerMode)))
	}
	if s.Assisted {
		result = append(result, "assisted")
	}
	return strings.Join(result, ", ")
}

func (s Summary) ScoreTable() string {
	var header, separator string
	if s.QTCsEnabled {
		header = "Band QSOs  QTCs  Dupe Pts     P/Q  Mult Q/M  Result \n"
		separator = "----------------------------------------------------\n"
	} else {
		header = "Band QSOs  Dupe Pts     P/Q  Mult Q/M  Result \n"
		separator = "----------------------------------------------\n"
	}

	buf := bytes.NewBufferString("")
	fmt.Fprint(buf, header)
	fmt.Fprint(buf, separator)
	for _, band := range Bands {
		if score, ok := s.Score.ScorePerBand[band]; ok {
			fmt.Fprintf(buf, "%4s %s\n", band, s.bandScoreToString(score))
		}
	}
	fmt.Fprint(buf, separator)
	fmt.Fprintf(buf, "Tot  %s\n", s.bandScoreToString(s.Score.Result()))
	return buf.String()
}

func (s Summary) bandScoreToString(score BandScore) string {
	if s.QTCsEnabled {
		return fmt.Sprintf("%5d %5d %4d %7d %4.1f %4d %4.1f %7d", score.QSOs, score.QTCs, score.Duplicates, score.Points, score.PointsPerQSO(), score.Multis, score.QSOsPerMulti(), score.Result())
	} else {
		return fmt.Sprintf("%5d %4d %7d %4.1f %4d %4.1f %7d", score.QSOs, score.Duplicates, score.Points, score.PointsPerQSO(), score.Multis, score.QSOsPerMulti(), score.Result())
	}
}

type TimeReport = conval.TimeReport

type Score struct {
	ScorePerBand map[Band]BandScore
	GraphPerBand map[Band]BandGraph
}

func NewScore() Score {
	return Score{
		ScorePerBand: make(map[Band]BandScore),
		GraphPerBand: make(map[Band]BandGraph),
	}
}

func (s Score) Copy() Score {
	result := Score{
		ScorePerBand: make(map[Band]BandScore),
		GraphPerBand: make(map[Band]BandGraph),
	}

	maps.Copy(result.ScorePerBand, s.ScorePerBand)
	for band, bandGraph := range s.GraphPerBand {
		result.GraphPerBand[band] = bandGraph.Copy()
	}

	return result
}

func (s Score) String() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Band QSOs  QTCs  Dupe Pts     P/Q  Mult Q/M  Result \n")
	fmt.Fprintf(buf, "----------------------------------------------------\n")
	for _, band := range Bands {
		if score, ok := s.ScorePerBand[band]; ok {
			fmt.Fprintf(buf, "%4s %s\n", band, score)
		}
	}
	fmt.Fprintf(buf, "----------------------------------------------------\n")
	fmt.Fprintf(buf, "Tot  %s\n", s.Result())
	return buf.String()
}

func (s Score) Result() BandScore {
	result := BandScore{}
	for _, score := range s.ScorePerBand {
		result.Add(score)
	}
	return result
}

func (s Score) StackedGraphPerBand() []BandGraph {
	result := make([]BandGraph, 0, len(Bands))
	var lastDataPoints []BandScore
	for _, band := range Bands {
		graph, ok := s.GraphPerBand[band]
		if !ok {
			continue
		}
		stackedGraph := BandGraph{
			Band:       graph.Band,
			DataPoints: make([]BandScore, len(graph.DataPoints)),
			Max:        graph.Max,
			startTime:  graph.startTime,
			duration:   graph.duration,
			binSeconds: graph.binSeconds,
		}

		for i, dataPoint := range graph.DataPoints {
			stackedGraph.DataPoints[i] = dataPoint
			if lastDataPoints != nil {
				stackedGraph.DataPoints[i].QSOs += lastDataPoints[i].QSOs
				stackedGraph.DataPoints[i].QTCs += lastDataPoints[i].QTCs
				stackedGraph.DataPoints[i].Duplicates += lastDataPoints[i].Duplicates
				stackedGraph.DataPoints[i].Points += lastDataPoints[i].Points
				stackedGraph.DataPoints[i].Multis += lastDataPoints[i].Multis
				stackedGraph.Max = stackedGraph.Max.Max(stackedGraph.DataPoints[i])
			}
		}

		result = append(result, stackedGraph)
		lastDataPoints = stackedGraph.DataPoints
	}
	return result
}

type BandGraph struct {
	Band       Band
	DataPoints []BandScore
	Max        BandScore

	startTime  time.Time
	duration   time.Duration
	binSeconds float64
}

func NewBandGraph(band Band, startTime time.Time, duration time.Duration) BandGraph {
	var binCount int
	if startTime.IsZero() || duration == 0 {
		binCount = 1
	} else {
		binCount = BinCountForDuration(duration)
	}
	return BandGraph{
		Band:       band,
		DataPoints: make([]BandScore, int(binCount)),
		Max:        BandScore{},

		binSeconds: duration.Seconds() / float64(binCount),
		startTime:  startTime,
		duration:   duration,
	}
}

func BinCountForDuration(duration time.Duration) int {
	switch {
	case duration <= 0:
		return 1
	case duration <= time.Hour:
		return int(duration / time.Minute)
	case duration <= 6*time.Hour:
		return int(duration / (5 * time.Minute))
	case duration <= 12*time.Hour:
		return int(duration / (10 * time.Minute))
	case duration <= 24*time.Hour:
		return int(duration / (15 * time.Minute))
	case duration <= 36*time.Hour:
		return int(duration / (30 * time.Minute))
	case duration <= 48*time.Hour:
		return int(duration / (60 * time.Minute))
	default:
		return 50
	}
}

func (g BandGraph) Copy() BandGraph {
	result := BandGraph{
		Band:       g.Band,
		DataPoints: make([]BandScore, len(g.DataPoints)),
		Max:        g.Max,
		startTime:  g.startTime,
		duration:   g.duration,
		binSeconds: g.binSeconds,
	}

	copy(result.DataPoints, g.DataPoints)

	return result
}

func (g BandGraph) String() string {
	points := make([]string, len(g.DataPoints))
	multis := make([]string, len(g.DataPoints))
	for i, value := range g.DataPoints {
		points[i] = fmt.Sprintf("%3d", value.Points)
		multis[i] = fmt.Sprintf("%3d", value.Multis)
	}
	return fmt.Sprintf("P: %s\nM: %s\n", strings.Join(points, " | "), strings.Join(multis, " | "))
}

func (g *BandGraph) Add(timestamp time.Time, score QSOScore) {
	bindex := g.Bindex(timestamp)
	if bindex == -1 {
		return
	}

	bandScore := g.DataPoints[bindex]
	bandScore.AddQSO(score)
	g.DataPoints[bindex] = bandScore

	g.Max = g.Max.Max(bandScore)
}

func (g *BandGraph) AddQTC(timestamp time.Time, score QTCScore) {
	bindex := g.Bindex(timestamp)
	if bindex == -1 {
		return
	}

	bandScore := g.DataPoints[bindex]
	bandScore.AddQTC(score)
	g.DataPoints[bindex] = bandScore

	g.Max = g.Max.Max(bandScore)
}

// Bindex returns the index of the bin corresponding to the given timestamp.
func (g *BandGraph) Bindex(timestamp time.Time) int {
	if g.startTime.IsZero() {
		return 0
	}
	if timestamp.IsZero() {
		return -1
	}
	if timestamp.Before(g.startTime) {
		return -1
	}

	binCount := len(g.DataPoints)
	if binCount == 1 {
		return 0
	}
	if g.binSeconds == 0 {
		return -1
	}

	seconds := timestamp.Sub(g.startTime).Seconds()

	result := int(seconds / g.binSeconds)
	if result > binCount-1 {
		return -1
	}

	return result
}

func (g BandGraph) ScaleHourlyGoalToBin(goal int) float64 {
	if g.binSeconds == 0 {
		return float64(goal)
	}
	return (g.binSeconds / 3600.0) * float64(goal)
}

func (g BandGraph) ElapsedTime(timestamp time.Time) time.Duration {
	if g.startTime.IsZero() {
		return 0
	}
	if timestamp.IsZero() {
		return 0
	}
	if timestamp.Before(g.startTime) {
		return 0
	}

	return timestamp.Sub(g.startTime)
}

func (g BandGraph) ElapsedTimePercent(timestamp time.Time) float64 {
	if g.startTime.IsZero() {
		return 0
	}
	if g.duration == 0 {
		return 0
	}
	if timestamp.IsZero() {
		return 0
	}
	if timestamp.Before(g.startTime) {
		return 0
	}

	return float64(g.ElapsedTime(timestamp)) / float64(g.duration)
}

func (g BandGraph) PercentAsDuration(percent float64) time.Duration {
	return time.Duration(float64(g.duration) * percent)
}

type BandScore struct {
	QSOs       int
	QTCs       int
	Duplicates int
	Points     int
	Multis     int
}

func (s BandScore) String() string {
	return fmt.Sprintf("%5d %5d %4d %7d %4.1f %4d %4.1f %7d", s.QSOs, s.QTCs, s.Duplicates, s.Points, s.PointsPerQSO(), s.Multis, s.QSOsPerMulti(), s.Result())
}

func (s *BandScore) Add(other BandScore) {
	s.QSOs += other.QSOs
	s.QTCs += other.QTCs
	s.Duplicates += other.Duplicates
	s.Points += other.Points
	s.Multis += other.Multis
}

func (s *BandScore) AddQSO(qso QSOScore) {
	s.QSOs += 1
	if qso.Duplicate {
		s.Duplicates += 1
	} else {
		s.Points += qso.Points
		s.Multis += qso.Multis
	}
}

func (s *BandScore) AddQTC(qtc QTCScore) {
	s.QTCs += qtc.Value
}

func (s BandScore) Max(other BandScore) BandScore {
	result := s

	if result.QSOs < other.QSOs {
		result.QSOs = other.QSOs
	}
	if result.QTCs < other.QTCs {
		result.QTCs = other.QTCs
	}
	if result.Duplicates < other.Duplicates {
		result.Duplicates = other.Duplicates
	}
	if result.Points < other.Points {
		result.Points = other.Points
	}
	if result.Multis < other.Multis {
		result.Multis = other.Multis
	}

	return result
}

func (s BandScore) PointsPerQSO() float64 {
	if s.QSOs == 0 {
		return 0
	}
	return float64(s.Points) / float64(s.QSOs)
}

func (s BandScore) QSOsPerMulti() float64 {
	if s.Multis == 0 {
		return 0
	}
	return float64(s.QSOs) / float64(s.Multis)
}

func (s BandScore) Result() int {
	points := s.Points + s.QTCs
	if s.Multis == 0 {
		return points
	}
	return points * s.Multis
}

type QSOScore struct {
	Points    int
	Multis    int
	Duplicate bool
}

type QTCScore struct {
	Value int
}

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
	SinceLastQSO time.Duration

	LastHourPoints int
	Last5MinPoints int
	LastHourMultis int
	Last5MinMultis int
}

func (r QSORate) SinceLastQSOFormatted() string {
	total := int(r.SinceLastQSO.Truncate(time.Second).Seconds())
	hours := int(total / (60 * 60))
	minutes := int(total/60) % 60
	seconds := int(total % 60)
	switch {
	case total < 60:
		return fmt.Sprintf("%2ds", seconds)
	case total < 60*60:
		return fmt.Sprintf("%02d:%02d", minutes, seconds)
	default:
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
}

type SpotType string

const (
	WorkedSpot  SpotType = "worked"
	ManualSpot  SpotType = "manual"
	SkimmerSpot SpotType = "skimmer"
	RBNSpot     SpotType = "rbn"
	ClusterSpot SpotType = "cluster"
	UnknownSpot SpotType = ""

	maxSpotTypePriority = 10
)

var spotTypePriorities = map[SpotType]int{
	WorkedSpot:  0,
	ManualSpot:  1,
	SkimmerSpot: 2,
	RBNSpot:     3,
	ClusterSpot: 4,
	UnknownSpot: maxSpotTypePriority,
}

func (t SpotType) Priority() int {
	priority, ok := spotTypePriorities[t]
	if !ok {
		return maxSpotTypePriority
	}
	return priority
}

type SpotQuality int

const (
	UnknownSpotQuality SpotQuality = iota
	BustedSpotQuality
	QSYSpotQuality
	ValidSpotQuality
)

const SpotQualityTags = "?BQV"

func (q SpotQuality) Tag() string {
	i := int(q)
	if i > 0 && i < len(SpotQualityTags) {
		return string(SpotQualityTags[q])
	}
	return string(SpotQualityTags[0])
}

type SpotFilter string

const (
	AllSpots              SpotFilter = ""
	OwnContinentSpotsOnly SpotFilter = "continent"
	OwnCountrySpotsOnly   SpotFilter = "country"
)

type SpotSource struct {
	Name            string     `json:"name"`
	Type            SpotType   `json:"type"`
	HostAddress     string     `json:"host_address"`
	Username        string     `json:"username"`
	Password        string     `json:"password,omitempty"`
	Filter          SpotFilter `json:"filter,omitempty"`
	IgnoreTimestamp bool       `json:"ignore_timestamp,omitempty"`
}

type Spot struct {
	Call      Callsign
	Frequency Frequency
	Band      Band
	Mode      Mode
	Time      time.Time
	Source    SpotType
}

func (s Spot) IsWorked() bool {
	return s.Source == WorkedSpot
}

type BandmapFrame struct {
	Frequency         Frequency
	ActiveBand        Band
	VisibleBand       Band
	Mode              Mode
	Bands             []BandSummary
	Entries           []BandmapEntry
	Index             BandmapFrameIndex
	SelectedEntry     BandmapEntry
	NearestEntry      BandmapEntry
	HighestValueEntry BandmapEntry
	QTCsEnabled       bool
	Filter            SpotFilterFrame
}

func (f BandmapFrame) IndexOf(id BandmapEntryID) (int, bool) {
	index, found := f.Index[id]
	return index, found
}

func (f BandmapFrame) EntryByID(id BandmapEntryID) (BandmapEntry, bool) {
	index, found := f.IndexOf(id)
	if !found {
		return BandmapEntry{}, false
	}
	return f.Entries[index], true
}

type BandmapFrameIndex map[BandmapEntryID]int

func NewFrameIndex(entries []BandmapEntry) BandmapFrameIndex {
	result := make(BandmapFrameIndex, len(entries))
	for i, entry := range entries {
		result[entry.ID] = i
	}
	return result
}

type BandSummary struct {
	Band        Band
	SpotCount   int
	Points      int
	MultiValues map[conval.Property]map[string]bool

	MaxSpots  bool
	MaxPoints bool
	MaxMultis bool
	Active    bool
	Visible   bool
}

func (s *BandSummary) AddMultiValues(values map[conval.Property]string) {
	if s.MultiValues == nil {
		s.MultiValues = make(map[conval.Property]map[string]bool)
	}
	for property, value := range values {
		propertyValues, ok := s.MultiValues[property]
		if !ok {
			propertyValues = make(map[string]bool)
		}
		propertyValues[value] = true
		s.MultiValues[property] = propertyValues
	}
}

func (s *BandSummary) Multis() int {
	result := 0
	for _, values := range s.MultiValues {
		result += len(values)
	}
	return result
}

type SpotFilterKind int

const (
	SpotFilterAll SpotFilterKind = iota
	SpotFilterVFO1
	SpotFilterVFO2
	SpotFilterFocused
	SpotFilterContest
	SpotFilterFixed
)

func (k SpotFilterKind) String() string {
	switch k {
	case SpotFilterVFO1:
		return "vfo1"
	case SpotFilterVFO2:
		return "vfo2"
	case SpotFilterFocused:
		return "focused"
	case SpotFilterContest:
		return "contest"
	case SpotFilterFixed:
		return "fixed"
	default:
		return "all"
	}
}

func ParseSpotFilterKind(s string) (SpotFilterKind, bool) {
	switch s {
	case "all":
		return SpotFilterAll, true
	case "vfo1":
		return SpotFilterVFO1, true
	case "vfo2":
		return SpotFilterVFO2, true
	case "focused":
		return SpotFilterFocused, true
	case "contest":
		return SpotFilterContest, true
	case "fixed":
		return SpotFilterFixed, true
	default:
		return SpotFilterAll, false
	}
}

type SpotFilterBand struct {
	Kind SpotFilterKind
	Band Band
}

func FixedSpotFilterBand(band Band) SpotFilterBand {
	return SpotFilterBand{Kind: SpotFilterFixed, Band: band}
}

func (b SpotFilterBand) String() string {
	if b.Kind == SpotFilterFixed {
		return string(b.Band)
	}
	return b.Kind.String()
}

func ParseSpotFilterBand(s string) SpotFilterBand {
	if kind, ok := ParseSpotFilterKind(s); ok && kind != SpotFilterFixed {
		return SpotFilterBand{Kind: kind}
	}
	for _, band := range Bands {
		if string(band) == s {
			return FixedSpotFilterBand(band)
		}
	}
	return SpotFilterBand{Kind: SpotFilterAll}
}

type SpotFilterMode struct {
	Kind SpotFilterKind
	Mode Mode
}

func FixedSpotFilterMode(mode Mode) SpotFilterMode {
	return SpotFilterMode{Kind: SpotFilterFixed, Mode: mode}
}

func (m SpotFilterMode) String() string {
	if m.Kind == SpotFilterFixed {
		return string(m.Mode)
	}
	return m.Kind.String()
}

func ParseSpotFilterMode(s string) SpotFilterMode {
	if kind, ok := ParseSpotFilterKind(s); ok && kind != SpotFilterFixed {
		return SpotFilterMode{Kind: kind}
	}
	for _, mode := range Modes {
		if string(mode) == s {
			return FixedSpotFilterMode(mode)
		}
	}
	return SpotFilterMode{Kind: SpotFilterAll}
}

type SpotSortColumn int

const (
	SortSpotsByFrequency SpotSortColumn = iota
	SortSpotsByCallsign
	SortSpotsByValue
	SortSpotsByLastSeen
)

var SpotSortColumns = []SpotSortColumn{SortSpotsByFrequency, SortSpotsByCallsign, SortSpotsByValue, SortSpotsByLastSeen}

func (c SpotSortColumn) String() string {
	switch c {
	case SortSpotsByCallsign:
		return "callsign"
	case SortSpotsByValue:
		return "value"
	case SortSpotsByLastSeen:
		return "last_seen"
	default:
		return "frequency"
	}
}

func (c SpotSortColumn) Label() string {
	switch c {
	case SortSpotsByCallsign:
		return "Callsign"
	case SortSpotsByValue:
		return "Value"
	case SortSpotsByLastSeen:
		return "Last Seen"
	default:
		return "Frequency"
	}
}

func ParseSpotSortColumn(s string) SpotSortColumn {
	switch s {
	case "callsign":
		return SortSpotsByCallsign
	case "value":
		return SortSpotsByValue
	case "last_seen":
		return SortSpotsByLastSeen
	default:
		return SortSpotsByFrequency
	}
}

type SpotFilterState struct {
	Band       SpotFilterBand
	Mode       SpotFilterMode
	SortBy     SpotSortColumn
	Descending bool
	Folded     bool
}

type SpotFilterFrame struct {
	SpotFilterState

	Bands       []Band
	Modes       []Mode
	Description string
}

type BandMatrixFrame struct {
	Bands         []BandSummary
	VFOBands      [VFOCount]Band
	FocusedVFO    VFOID
	VFO2Available bool
}

type Callinfo struct {
	Input     string
	Call      Callsign
	CallValid bool

	DXCCEntity dxcc.Prefix
	Azimuth    latlon.Degrees
	Distance   latlon.Km

	PredictedExchange []string

	UserText string

	Worked        bool // already worked on another band/mode, but does not count as duplicate
	Duplicate     bool // counts as duplicate
	Points        int
	Multis        int
	MultiValues   map[conval.Property]string
	Value         int
	WeightedValue float64

	SentQTCs     int
	ReceivedQTCs int
}

type CallinfoFrame struct {
	NormalizedCallInput string
	DXCCEntity          dxcc.Prefix
	Azimuth             latlon.Degrees
	Distance            latlon.Km

	CallsignOnFrequency AnnotatedCallsign

	PredictedExchange []string

	Points int
	Multis int
	Value  int

	SentQTCs     int
	ReceivedQTCs int

	UserInfo string

	Supercheck []AnnotatedCallsign
}

func (f CallinfoFrame) BestMatchOnFrequency() AnnotatedCallsign {
	if len(f.Supercheck) > 0 {
		return f.Supercheck[0]
	}
	if f.CallsignOnFrequency.Callsign.String() != "" {
		return f.CallsignOnFrequency
	}
	return AnnotatedCallsign{}
}

func (f CallinfoFrame) GetMatch(i int) string {
	if i < len(f.Supercheck) {
		return f.Supercheck[i].Callsign.String()
	}
	return ""
}

// frequencies within this distance to an entry's frequency will be recognized as "in proximity"
const spotFrequencyProximityThreshold float64 = 2500

// spots within this distance to an entry's frequency will be considered "on frequency"
const spotFrequencyDeltaThreshold float64 = 300

// spots within at least this proximity will be considered "on frequency"
const spotOnFrequencyThreshold float64 = 1.0 - (spotFrequencyDeltaThreshold / spotFrequencyProximityThreshold)

type BandmapEntryID uint64

const NoEntryID BandmapEntryID = 0

type BandmapEntry struct {
	ID        BandmapEntryID
	Label     string
	Call      Callsign
	Frequency Frequency
	Band      Band
	Mode      Mode
	LastHeard time.Time
	Source    SpotType
	SpotCount int
	Quality   SpotQuality

	Info Callinfo
}

// ProximityFactor increases the closer the given frequency is to this entry's frequency.
// 0.0 = not in proximity, 1.0 = exactly on frequency
// the sign indiciates if the entry's frequency is above (>0) or below (<0) the reference frequency
func (e BandmapEntry) ProximityFactor(frequency Frequency) float64 {
	frequencyDelta := math.Abs(float64(e.Frequency - frequency))
	if frequencyDelta > spotFrequencyProximityThreshold {
		return 0.0
	}

	result := 1.0 - (frequencyDelta / spotFrequencyProximityThreshold)
	if e.Frequency < frequency {
		result *= -1.0
	}

	return result
}

// OnFrequency indicates if this entry is on the given frequency, within the defined threshold.
func (e BandmapEntry) OnFrequency(frequency Frequency) bool {
	return math.Abs(e.ProximityFactor(frequency)) >= spotOnFrequencyThreshold
}

type BandmapOrder func(BandmapEntry, BandmapEntry) int

func Compare[T constraints.Ordered](a, b T) int {
	switch {
	case a < b:
		return -1
	case a == b:
		return 0
	case a > b:
		return 1
	default:
		panic("compare")
	}
}

func Descending(o BandmapOrder) BandmapOrder {
	return func(a, b BandmapEntry) int {
		return o(b, a)
	}
}

func BandmapByFrequency(a, b BandmapEntry) int {
	if a.Frequency == b.Frequency {
		return Compare(a.ID, b.ID)
	}
	return Compare(a.Frequency, b.Frequency)
}

func BandmapByDistance(referenceFrequency Frequency) BandmapOrder {
	return func(a, b BandmapEntry) int {
		deltaA := math.Abs(float64(a.Frequency - referenceFrequency))
		deltaB := math.Abs(float64(b.Frequency - referenceFrequency))
		if deltaA == deltaB {
			return Compare(a.ID, b.ID)
		}
		return Compare(deltaA, deltaB)
	}
}

func BandmapByDistanceAndDescendingID(referenceFrequency Frequency) BandmapOrder {
	return func(a, b BandmapEntry) int {
		deltaA := math.Abs(float64(a.Frequency - referenceFrequency))
		deltaB := math.Abs(float64(b.Frequency - referenceFrequency))
		if deltaA == deltaB {
			return Compare(a.ID, b.ID) * -1
		}
		return Compare(deltaA, deltaB)
	}
}

func BandmapByCallsign(a, b BandmapEntry) int {
	if a.Call.String() == b.Call.String() {
		return Compare(a.ID, b.ID)
	}
	return Compare(a.Call.String(), b.Call.String())
}

func BandmapByLastSeen(a, b BandmapEntry) int {
	if a.LastHeard.Equal(b.LastHeard) {
		return Compare(a.ID, b.ID)
	}
	if a.LastHeard.Before(b.LastHeard) {
		return -1
	}
	return 1
}

func BandmapByValue(a, b BandmapEntry) int {
	if a.Info.WeightedValue == b.Info.WeightedValue {
		return Compare(a.ID, b.ID)
	}
	return Compare(a.Info.WeightedValue, b.Info.WeightedValue)
}

type BandmapFilter func(entry BandmapEntry) bool

func And(filters ...BandmapFilter) BandmapFilter {
	return func(entry BandmapEntry) bool {
		for _, filter := range filters {
			if !filter(entry) {
				return false
			}
		}
		return true
	}
}

func Or(filters ...BandmapFilter) BandmapFilter {
	return func(entry BandmapEntry) bool {
		for _, filter := range filters {
			if filter(entry) {
				return true
			}
		}
		return false
	}
}

func Not(filter BandmapFilter) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return !filter(entry)
	}
}

func IsWorkedSpot(entry BandmapEntry) bool {
	return entry.Source == WorkedSpot
}

func OnFrequency(frequency Frequency) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return entry.OnFrequency(frequency)
	}
}

func OnBand(band Band) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return entry.Band == band
	}
}

func InMode(mode Mode) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return entry.Mode == mode
	}
}

func FromSource(source SpotType) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return entry.Source == source
	}
}

func WithQuality(quality SpotQuality) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return entry.Quality == quality
	}
}

func HeardAfter(deadline time.Time) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return entry.LastHeard.After(deadline)
	}
}

func FromContinent(continent string) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return entry.Info.DXCCEntity.Continent == continent
	}
}

func FromDXCC(primaryPrefix string) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return entry.Info.DXCCEntity.PrimaryPrefix == primaryPrefix
	}
}

type BandmapWeights struct {
	AgeSeconds float64
	Spots      float64
	Source     float64
	Quality    float64
}

type VFOID int

const (
	VFO1 VFOID = iota
	VFO2

	VFOCount
)

type VFO interface {
	Name() string
	Notify(any)
	Refresh()
	SetFrequency(Frequency)
	ShiftFrequency(Frequency)
	SetBand(Band)
	SetMode(Mode)
	IncrementalTuningActive(kind IncrementalTuningKind) bool
	SetIncrementalTuningActive(kind IncrementalTuningKind, active bool)
	ShiftIncrementalTuning(IncrementalTuningKind, Frequency)
	ToggleAvailableIncrementalTuning()
	ShiftAvailableIncrementalTuning(Frequency)
}

type CurrentVFOListener interface {
	CurrentVFOChanged(VFOID)
}

type TXVFOListener interface {
	TXVFOChanged(VFOID)
}

type FocusedVFOListener interface {
	FocusedVFOChanged(VFOID)
}

type FocusedVFOListenerFunc func(VFOID)

func (f FocusedVFOListenerFunc) FocusedVFOChanged(vfo VFOID) {
	f(vfo)
}

type VFOFrequencyListener interface {
	VFOFrequencyChanged(VFOID, Frequency)
}

type VFOBandListener interface {
	VFOBandChanged(VFOID, Band)
}

type VFOModeListener interface {
	VFOModeChanged(VFOID, Mode)
}

type IncrementalTuningKind int

const (
	RIT IncrementalTuningKind = iota
	XIT
)

func (k IncrementalTuningKind) String() string {
	switch k {
	case RIT:
		return "RIT"
	case XIT:
		return "XIT"
	default:
		return "unknown"
	}
}

func (k IncrementalTuningKind) Workmode() Workmode {
	if k == RIT {
		return Run
	}
	return SearchPounce
}

type VFOIncrementalTuningListener interface {
	VFOIncrementalTuningChanged(vfo VFOID, kind IncrementalTuningKind, active bool, offset Frequency)
}

type VFOIncrementalTuningActiveListener interface {
	VFOIncrementalTuningActiveChanged(vfo VFOID, kind IncrementalTuningKind, active bool)
}

type VFOIncrementalTuningVisibilityListener interface {
	VFOIncrementalTuningVisibilityChanged(vfo VFOID, kind IncrementalTuningKind, visible bool)
}

type VFOPTTListener interface {
	VFOPTTChanged(VFOID, bool)
}

type RadioChangedListener interface {
	RadioChanged(name string, singleVFO bool)
}

type ConnectionChangedListener interface {
	ConnectionChanged(bool)
}

type ConnectionChangedFunc func(bool)

func (f ConnectionChangedFunc) ConnectionChanged(connected bool) {
	f(connected)
}

type Service int

const (
	NoService Service = iota
	RadioService
	KeyerService
	DXCCService
	SCPService
	CallHistoryService
	MapService
	RemoteService
)

type ServiceStatusListener interface {
	StatusChanged(service Service, avialable bool)
}

type ServiceStatusListenerFunc func(Service, bool)

func (f ServiceStatusListenerFunc) StatusChanged(service Service, available bool) {
	f(service, available)
}

type AsyncRunner func(func())

type CallsignEnteredListener interface {
	CallsignEntered(vfo VFOID, callsign string)
}

type CallsignLoggedListener interface {
	CallsignLogged(callsign string, frequency Frequency)
}

type TransmissionStartedListener interface {
	TransmissionStarted(vfo VFOID)
}

func FormatTimestamp(ts time.Time) string {
	return ts.UTC().Format("2006-01-02 15:04Z")
}

func FormatDuration(d time.Duration) string {
	return fmt.Sprintf("%02dh%02d", int(d.Hours()), int(d.Minutes())-int(d.Hours())*60)
}

func Emit[L any](listeners []any, notify func(listener L)) {
	for i := range listeners {
		listener, ok := listeners[i].(L)
		if ok {
			notify(listener)
		}
	}
}
