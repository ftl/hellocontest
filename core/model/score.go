package model

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/ftl/conval"
)

type Score struct {
	ScorePerBand    map[Band]BandScore
	GraphPerBand    map[Band]BandGraph
	MultiProperties []conval.Property
}

func NewScore() Score {
	return Score{
		ScorePerBand: make(map[Band]BandScore),
		GraphPerBand: make(map[Band]BandGraph),
	}
}

func (s Score) Copy() Score {
	result := Score{
		ScorePerBand:    make(map[Band]BandScore),
		GraphPerBand:    make(map[Band]BandGraph),
		MultiProperties: slices.Clone(s.MultiProperties),
	}

	for band, bandScore := range s.ScorePerBand {
		result.ScorePerBand[band] = bandScore.Copy()
	}
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

	for i, dataPoint := range g.DataPoints {
		result.DataPoints[i] = dataPoint.Copy()
	}

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
	QSOs              int
	QTCs              int
	Duplicates        int
	Points            int
	Multis            int
	MultisPerProperty map[conval.Property]int
}

func (s BandScore) Copy() BandScore {
	result := s
	result.MultisPerProperty = maps.Clone(s.MultisPerProperty)
	return result
}

func (s *BandScore) addMultisPerProperty(multisPerProperty map[conval.Property]int) {
	if len(multisPerProperty) == 0 {
		return
	}
	// a band score must not share its map with another band score
	if s.MultisPerProperty == nil {
		s.MultisPerProperty = make(map[conval.Property]int, len(multisPerProperty))
	}
	for property, multis := range multisPerProperty {
		s.MultisPerProperty[property] += multis
	}
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
	s.addMultisPerProperty(other.MultisPerProperty)
}

func (s *BandScore) AddQSO(qso QSOScore) {
	s.QSOs += 1
	if qso.Duplicate {
		s.Duplicates += 1
	} else {
		s.Points += qso.Points
		s.Multis += qso.Multis
		s.addMultisPerProperty(qso.MultisPerProperty)
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
	Points            int
	Multis            int
	MultisPerProperty map[conval.Property]int
	Duplicate         bool
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
