package core

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/locator"
)

// QSO contains the details about one radio contact.
type QSO struct {
	Callsign      callsign.Callsign
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
}

func (qso *QSO) String() string {
	return fmt.Sprintf("%s|%-10s|%5.0fkHz|%4s|%-4s|%s|%s|%s|%s|%s|%s|%2d|%2d|%t", qso.Time.Format("15:04"), qso.Callsign.String(), qso.Frequency/1000.0, qso.Band, qso.Mode, qso.MyReport, qso.MyNumber.String(), strings.Join(qso.MyExchange, " "), qso.TheirReport, qso.TheirNumber.String(), strings.Join(qso.TheirExchange, " "), qso.Points, qso.Multis, qso.Duplicate)
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
	SearchPounce Workmode = iota
	Run
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
}

// FilterPlaceholder can be used as placeholder for a missed character in the callsign.
const FilterPlaceholder = "."

// AnnotatedCallsign contains a callsign with additional information retrieved from databases and the logbook.
type AnnotatedCallsign struct {
	Callsign          callsign.Callsign
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

	Comparable interface{}
	Compare    func(interface{}, interface{}) bool
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
	var result string
	for _, match := range m {
		if match.OP != Delete {
			result += match.Value
		}
	}
	return result
}

type Settings interface {
	Station() Station
	Contest() Contest
}

type Station struct {
	Callsign callsign.Callsign
	Operator callsign.Callsign
	Locator  locator.Locator
}

type Contest struct {
	Definition             *conval.Definition
	Name                   string
	ExchangeValues         []string
	GenerateSerialExchange bool
	GenerateReport         bool
	StartTime              time.Time

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
	WPM       int      `json:"wpm"`
	Preset    string   `json:"preset"`
	SPMacros  []string `json:"sp_macros"`
	RunMacros []string `json:"run_macros"`
	SPLabels  []string `json:"sp_labels"`
	RunLabels []string `json:"run_labels"`
}

type KeyerPreset struct {
	Name      string   `json:"name"`
	SPMacros  []string `json:"sp_macros"`
	RunMacros []string `json:"run_macros"`
	SPLabels  []string `json:"sp_labels"`
	RunLabels []string `json:"run_labels"`
}

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

	for band, bandScore := range s.ScorePerBand {
		result.ScorePerBand[band] = bandScore
	}
	for band, bandGraph := range s.GraphPerBand {
		result.GraphPerBand[band] = bandGraph.Copy()
	}

	return result
}

func (s Score) String() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Band QSOs  Dupe Pts     P/Q  Mult Q/M  Result \n")
	fmt.Fprintf(buf, "----------------------------------------------\n")
	for _, band := range Bands {
		if score, ok := s.ScorePerBand[band]; ok {
			fmt.Fprintf(buf, "%4s %s\n", band, score)
		}
	}
	fmt.Fprintf(buf, "----------------------------------------------\n")
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
			binSeconds: graph.binSeconds,
		}

		for i, dataPoint := range graph.DataPoints {
			stackedGraph.DataPoints[i] = dataPoint
			if lastDataPoints != nil {
				stackedGraph.DataPoints[i].QSOs += lastDataPoints[i].QSOs
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
	binSeconds float64
}

func NewBandGraph(band Band, startTime time.Time, duration time.Duration) BandGraph {
	var binCount int
	if startTime.IsZero() || duration == 0 {
		binCount = 1
	} else {
		binCount = 60
	}
	return BandGraph{
		Band:       band,
		DataPoints: make([]BandScore, int(binCount)),
		Max:        BandScore{},

		binSeconds: duration.Seconds() / float64(binCount),
		startTime:  startTime,
	}
}

func (g BandGraph) Copy() BandGraph {
	result := BandGraph{
		Band:       g.Band,
		DataPoints: make([]BandScore, len(g.DataPoints)),
		Max:        g.Max,
		startTime:  g.startTime,
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

type BandScore struct {
	QSOs       int
	Duplicates int
	Points     int
	Multis     int
}

func (s BandScore) String() string {
	return fmt.Sprintf("%5d %4d %7d %4.1f %4d %4.1f %7d", s.QSOs, s.Duplicates, s.Points, s.PointsPerQSO(), s.Multis, s.QSOsPerMulti(), s.Result())
}

func (s *BandScore) Add(other BandScore) {
	s.QSOs += other.QSOs
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

func (s BandScore) Max(other BandScore) BandScore {
	result := s

	if result.QSOs < other.QSOs {
		result.QSOs = other.QSOs
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
	if s.Multis == 0 {
		return s.Points
	}
	return s.Points * s.Multis
}

type QSOScore struct {
	Points    int
	Multis    int
	Duplicate bool
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

	maxSpotTypePriority = 10
)

var spotTypePriorities = map[SpotType]int{
	WorkedSpot:  0,
	ManualSpot:  1,
	SkimmerSpot: 2,
	RBNSpot:     3,
	ClusterSpot: 4,
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
	Call      callsign.Callsign
	Frequency Frequency
	Band      Band
	Mode      Mode
	Time      time.Time
	Source    SpotType
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
	Points      int
	MultiValues map[conval.Property]map[string]bool

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

type Callinfo struct {
	Call callsign.Callsign

	DXCCName          string
	PrimaryPrefix     string
	Continent         string
	ITUZone           int
	CQZone            int
	UserText          string
	PredictedExchange []string
	FilteredExchange  []string
	ExchangeText      string

	Worked        bool // already worked on another band/mode, but does not count as duplicate
	Duplicate     bool // counts as duplicate
	Points        int
	Multis        int
	MultiValues   map[conval.Property]string
	Value         int
	WeightedValue float64
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
	Call      callsign.Callsign
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

type BandmapOrder func(BandmapEntry, BandmapEntry) bool

func Descending(o BandmapOrder) BandmapOrder {
	return func(a, b BandmapEntry) bool {
		return o(b, a)
	}
}

func BandmapByFrequency(a, b BandmapEntry) bool {
	return a.Frequency < b.Frequency
}

func BandmapByDistance(referenceFrequency Frequency) BandmapOrder {
	return func(a, b BandmapEntry) bool {
		deltaA := math.Abs(float64(a.Frequency - referenceFrequency))
		deltaB := math.Abs(float64(b.Frequency - referenceFrequency))
		return deltaA < deltaB
	}
}

func BandmapByDescendingValue(a, b BandmapEntry) bool {
	return a.Info.WeightedValue > b.Info.WeightedValue
}

type BandmapFilter func(entry BandmapEntry) bool

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
		return entry.Info.Continent == continent
	}
}

func FromDXCC(primaryPrefix string) BandmapFilter {
	return func(entry BandmapEntry) bool {
		return entry.Info.PrimaryPrefix == primaryPrefix
	}
}

type BandmapWeights struct {
	AgeSeconds float64
	Spots      float64
	Source     float64
	Quality    float64
}

type VFO interface {
	Notify(any)
	Refresh()
	SetFrequency(Frequency)
	SetBand(Band)
	SetMode(Mode)
}

type VFOFrequencyListener interface {
	VFOFrequencyChanged(Frequency)
}

type VFOBandListener interface {
	VFOBandChanged(Band)
}

type VFOModeListener interface {
	VFOModeChanged(Mode)
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
	CallsignEntered(callsign string)
}

type CallsignLoggedListener interface {
	CallsignLogged(callsign string, frequency Frequency)
}
