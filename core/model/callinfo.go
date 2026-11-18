package model

import (
	"github.com/ftl/conval"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/latlon"
)

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
	MarkerOnFrequency   string

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
