package bandmap

import (
	"math"
	"time"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

type View interface {
	Show(frame BandmapFrame)
}

type BandmapFrame struct {
	LowerBound core.Frequency
	UpperBound core.Frequency
	VFO        core.Frequency
	Entries    []Entry
}

type Bandmap struct {
	view View
}

func NewBandmap() *Bandmap {
	result := &Bandmap{}

	// TODO: start some kind of periodic update

	return result
}

func (m *Bandmap) SetView(v View) {
	m.view = v
	m.update()
}

func (m *Bandmap) Add(spot Spot) {

}

func (m *Bandmap) Clear() {

}

func (m *Bandmap) update() {
	// TODO: calculate the current frame
	frame := BandmapFrame{}

	if m.view != nil {
		m.view.Show(frame)
	}
}

type SpotSource string

const (
	ManualSpot  SpotSource = "manual"
	SkimmerSpot SpotSource = "skimmer"
	ClusterSpot SpotSource = "cluster"
	RBNSpot     SpotSource = "rbn"
)

type Spot struct {
	Call      callsign.Callsign
	Frequency core.Frequency
	Mode      core.Mode
	Time      time.Time
	Source    SpotSource
}

const (
	SpotFrequencyDeltaThreshold     float64 = 125  // spots within this distance to an entry's frequency will be added to the entry
	SpotFrequencyProximityThreshold float64 = 1000 // frequencies within this distance to an entry's frequency will be recognized as "in proximity"
)

type Entry struct {
	Call      callsign.Callsign
	Frequency core.Frequency
	LastHeard time.Time

	spots []Spot
}

func (e *Entry) Len() int {
	return len(e.spots)
}

func (e *Entry) Add(spot Spot) bool {
	if spot.Call != e.Call {
		return false
	}

	frequencyDelta := math.Abs(float64(e.Frequency - spot.Frequency))
	if frequencyDelta > SpotFrequencyDeltaThreshold {
		return false
	}

	e.spots = append(e.spots, spot)
	e.updateFrequency()
	if e.LastHeard.Before(spot.Time) {
		e.LastHeard = spot.Time
	}

	return true
}

func (e *Entry) RemoveSpotsBefore(timestamp time.Time) bool {
	k := 0
	for i, s := range e.spots {
		if !s.Time.Before(timestamp) {
			if i != k {
				e.spots[k] = s
			}
			k++
		}
	}
	e.spots = e.spots[:k]

	e.update()

	return len(e.spots) > 0
}

// ProximityFactor increases the closer the given frequency is to this entry's frequency.
// 0.0 = not in proximity, 1.0 = exactly on frequency
func (e *Entry) ProximityFactor(f core.Frequency) float64 {
	frequencyDelta := math.Abs(float64(e.Frequency - f))
	if frequencyDelta > SpotFrequencyProximityThreshold {
		return 0.0
	}

	return 1.0 - (frequencyDelta / SpotFrequencyProximityThreshold)
}

func (e *Entry) update() {
	e.updateFrequency()

	lastHeard := time.Time{}
	for _, s := range e.spots {
		if lastHeard.Before(s.Time) {
			lastHeard = s.Time
		}
	}
	e.LastHeard = lastHeard
}

func (e *Entry) updateFrequency() {
	if len(e.spots) == 0 {
		e.Frequency = 0
		return
	}

	var sum core.Frequency
	for _, s := range e.spots {
		sum += s.Frequency
	}
	e.Frequency = core.Frequency(math.RoundToEven(float64(sum)/float64(len(e.spots)*10))) * 10.0
}

type SpotList struct {
}

func NewSpotList() *SpotList {
	return &SpotList{}
}

func (l *SpotList) Add(spot Spot) {

}

func (l *SpotList) Update() {

}

func (l *SpotList) AllByFrequency() []Entry {
	return nil
}
