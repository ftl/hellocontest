package model

import (
	"fmt"
	"math"
	"time"

	"github.com/ftl/conval"
	"golang.org/x/exp/constraints"
)

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

type SpotProtocol string

const (
	// ClusterixProtocol is the DX cluster telnet protocol, the default of every spot source
	ClusterixProtocol SpotProtocol = ""
	// SDRainerProtocol is the channel stream of SDRainer, over gRPC
	SDRainerProtocol SpotProtocol = "sdrainer"
)

type SpotSource struct {
	Name            string       `json:"name"`
	Type            SpotType     `json:"type"`
	Protocol        SpotProtocol `json:"protocol,omitempty"`
	HostAddress     string       `json:"host_address"`
	Username        string       `json:"username"`
	Password        string       `json:"password,omitempty"`
	Filter          SpotFilter   `json:"filter,omitempty"`
	IgnoreTimestamp bool         `json:"ignore_timestamp,omitempty"`
}

type Spot struct {
	Call      Callsign
	Frequency Frequency
	Band      Band
	Mode      Mode
	Time      time.Time
	Source    SpotType
	// Quality is the judgement of the source about its own spot. A source that reports no
	// quality leaves this field at UnknownSpotQuality, and the bandmap judges by itself.
	Quality SpotQuality
	// SNR is the signal to noise ratio in dB that the source measured. A signal at the noise
	// floor is no spot, therefore 0 means that the source reported no value.
	SNR float64
}

func (s Spot) IsWorked() bool {
	return s.Source == WorkedSpot
}

type BandmapMarkerKind int

const (
	CQMarker BandmapMarkerKind = iota
	NumberedMarker
	TextMarker
)

type BandmapMarker struct {
	ID        BandmapEntryID
	Kind      BandmapMarkerKind
	Text      string
	Number    int
	Frequency Frequency
	Band      Band
	CreatedAt time.Time
	Input     MarkerInput
}

type MarkerInput struct {
	Call          string
	TheirExchange []string
}

func (i MarkerInput) IsEmpty() bool {
	if i.Call != "" {
		return false
	}
	for _, value := range i.TheirExchange {
		if value != "" {
			return false
		}
	}
	return true
}

func (m BandmapMarker) Label() string {
	if m.Input.Call == "" {
		return m.Text
	}
	// the operator sees which marker holds which station, the text stays the sort key
	return fmt.Sprintf("%s: %s", m.Text, m.Input.Call)
}

func (m BandmapMarker) ProximityFactor(frequency Frequency) float64 {
	return proximityFactor(m.Frequency, frequency)
}

func (m BandmapMarker) OnFrequency(frequency Frequency) bool {
	return math.Abs(m.ProximityFactor(frequency)) >= spotOnFrequencyThreshold
}

type BandmapRowKind int

const (
	SpotRow BandmapRowKind = iota
	MarkerRow
)

type BandmapRow struct {
	Kind   BandmapRowKind
	Entry  BandmapEntry
	Marker BandmapMarker
}

func SpotBandmapRow(entry BandmapEntry) BandmapRow {
	return BandmapRow{Kind: SpotRow, Entry: entry}
}

func MarkerBandmapRow(marker BandmapMarker) BandmapRow {
	return BandmapRow{Kind: MarkerRow, Marker: marker}
}

func (r BandmapRow) ID() BandmapEntryID {
	if r.Kind == MarkerRow {
		return r.Marker.ID
	}
	return r.Entry.ID
}

func (r BandmapRow) Frequency() Frequency {
	if r.Kind == MarkerRow {
		return r.Marker.Frequency
	}
	return r.Entry.Frequency
}

func (r BandmapRow) LastSeen() time.Time {
	if r.Kind == MarkerRow {
		// the creation time is the last seen value of a marker
		return r.Marker.CreatedAt
	}
	return r.Entry.LastHeard
}

type BandmapFrame struct {
	Frequency         Frequency
	ActiveBand        Band
	VisibleBand       Band
	Mode              Mode
	Bands             []BandSummary
	Rows              []BandmapRow
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

func (f BandmapFrame) RowByID(id BandmapEntryID) (BandmapRow, bool) {
	index, found := f.Index[id]
	if !found {
		return BandmapRow{}, false
	}
	return f.Rows[index], true
}

func (f BandmapFrame) EntryByID(id BandmapEntryID) (BandmapEntry, bool) {
	row, found := f.RowByID(id)
	if !found || row.Kind != SpotRow {
		return BandmapEntry{}, false
	}
	return row.Entry, true
}

type BandmapFrameIndex map[BandmapEntryID]int

func NewFrameIndex(rows []BandmapRow) BandmapFrameIndex {
	result := make(BandmapFrameIndex, len(rows))
	for i, row := range rows {
		result[row.ID()] = i
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

// SpotFilterValue selects the spots by their predicted contribution to the score.
type SpotFilterValue int

const (
	SpotFilterAnyValue SpotFilterValue = iota
	SpotFilterScoring
	SpotFilterMulti
)

var SpotFilterValues = []SpotFilterValue{SpotFilterAnyValue, SpotFilterScoring, SpotFilterMulti}

func (v SpotFilterValue) String() string {
	switch v {
	case SpotFilterScoring:
		return "scoring"
	case SpotFilterMulti:
		return "multi"
	default:
		return "all"
	}
}

func (v SpotFilterValue) Label() string {
	switch v {
	case SpotFilterScoring:
		return "Score > 0"
	case SpotFilterMulti:
		return "Multi > 0"
	default:
		return "All"
	}
}

func ParseSpotFilterValue(s string) SpotFilterValue {
	switch s {
	case "scoring":
		return SpotFilterScoring
	case "multi":
		return SpotFilterMulti
	default:
		return SpotFilterAnyValue
	}
}

// Matches indicates if the entry contributes as much to the score as this filter requires.
// A duplicate contributes nothing, although the contest rules still name points for the QSO.
func (v SpotFilterValue) Matches(entry BandmapEntry) bool {
	switch v {
	case SpotFilterScoring:
		return !entry.Info.Duplicate && entry.Info.Value > 0
	case SpotFilterMulti:
		return !entry.Info.Duplicate && entry.Info.Multis > 0
	default:
		return true
	}
}

type SpotFilterState struct {
	Band       SpotFilterBand
	Mode       SpotFilterMode
	Value      SpotFilterValue
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
	SNR       float64
	Ignored   bool

	Info Callinfo
}

// ProximityFactor increases the closer the given frequency is to this entry's frequency.
// 0.0 = not in proximity, 1.0 = exactly on frequency
// the sign indiciates if the entry's frequency is above (>0) or below (<0) the reference frequency
func (e BandmapEntry) ProximityFactor(frequency Frequency) float64 {
	return proximityFactor(e.Frequency, frequency)
}

func proximityFactor(ownFrequency Frequency, frequency Frequency) float64 {
	frequencyDelta := math.Abs(float64(ownFrequency - frequency))
	if frequencyDelta > spotFrequencyProximityThreshold {
		return 0.0
	}

	result := 1.0 - (frequencyDelta / spotFrequencyProximityThreshold)
	if ownFrequency < frequency {
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

type BandmapMarkerFilter func(marker BandmapMarker) bool

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
