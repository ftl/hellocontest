package callinfo

import (
	"log"
	"strings"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core/dxcc"
	"github.com/ftl/hellocontest/core/model"
)

// DXCCFinder returns a list of matching prefixes for the given string and indicates if there was a match at all.
type DXCCFinder interface {
	Find(string) (dxcc.Prefix, bool)
}

// CallsignFinder returns a list of matching callsigns for the given partial string.
type CallsignFinder interface {
	FindStrings(string) ([]string, error)
	Find(string) ([]model.AnnotatedCallsign, error)
}

// CallHistoryFinder returns additional information for a given callsign if a call history file is used.
type CallHistoryFinder interface {
	FindEntry(string) (model.AnnotatedCallsign, bool)
	Find(string) ([]model.AnnotatedCallsign, error)
	AvailableFor(fieldIndex int) bool
}

// DupeChecker can be used to find out if the given callsign was already worked, according to the contest rules.
// It can also find worked callsigns that are similar to a given string, e.g. for supercheck.
type DupeChecker interface {
	FindWorkedQSOs(model.Callsign, model.Band, model.Mode) ([]model.QSO, bool)
	Find(string) ([]model.AnnotatedCallsign, error)
}

// Valuer provides the points and multis of a QSO based on the given information.
type Valuer interface {
	Value(callsign model.Callsign, entity dxcc.Prefix, band model.Band, mode model.Mode, exchange []string) (points, multis int, multiValues map[conval.Property]string)
}

type QTCProvider interface {
	QTCsInLog(model.Callsign) (sent, received int)
}

// View defines the visual part of the call information window.
type View interface {
	ShowFrame(model.VFOID, model.CallinfoFrame)
	SetVFOEnabled(model.VFOID, bool)
}

type CallinfoFrameListener interface {
	CallinfoFrameChanged(model.VFOID, model.CallinfoFrame)
}

type Callinfo struct {
	view        View
	collector   *Collector
	supercheck  *Supercheck
	qtcProvider QTCProvider
	listeners   []any

	theirExchangeFields []model.ExchangeField
	qtcsEnabled         bool
	vfo2Enabled         bool

	frames [model.VFOCount]model.CallinfoFrame
}

func New(entities DXCCFinder, callsigns CallsignFinder, callHistory CallHistoryFinder, dupeChecker DupeChecker, valuer Valuer, qtcProvider QTCProvider) *Callinfo {
	result := &Callinfo{
		view:        new(nullView),
		collector:   NewCollector(entities, callsigns, callHistory, dupeChecker, valuer, qtcProvider),
		supercheck:  NewSupercheck(entities, callsigns, callHistory, dupeChecker, valuer),
		qtcProvider: qtcProvider,
	}

	return result
}

func (c *Callinfo) SetView(view View) {
	if view == nil {
		panic("callinfo.Callinfo.SetView must not be called with nil")
	}
	if _, ok := c.view.(*nullView); !ok {
		panic("callinfo.Callinfo.SetView was already called")
	}

	c.view = view
	c.view.SetVFOEnabled(model.VFO2, c.vfo2Enabled)
}

func (c *Callinfo) StationChanged(station model.Station) {
	c.collector.SetMyLocator(station.Locator)
}

func (c *Callinfo) ContestChanged(contest model.Contest) {
	if contest.Definition == nil {
		log.Printf("there is no contest definition!")
		return
	}
	c.theirExchangeFields = contest.TheirExchangeFields
	c.qtcsEnabled = contest.EnableQTCs
	c.collector.SetTheirExchangeFields(c.theirExchangeFields, contest.TheirReportExchangeField, contest.TheirNumberExchangeField)
	c.supercheck.SetTheirExchangeFields(c.theirExchangeFields)
}

func (c *Callinfo) ScoreChanged(score model.Score) {
	c.collector.ScoreChanged(score)
}

func (c *Callinfo) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Callinfo) emitFrameChanged(vfo model.VFOID) {
	for _, listener := range c.listeners {
		if l, ok := listener.(CallinfoFrameListener); ok {
			l.CallinfoFrameChanged(vfo, c.frames[vfo])
		}
	}
	c.view.ShowFrame(vfo, c.frames[vfo])
}

func (c *Callinfo) GetInfo(call model.Callsign, band model.Band, mode model.Mode, currentExchange []string) model.Callinfo {
	return c.collector.GetInfo(call, band, mode, currentExchange)
}

func (c *Callinfo) UpdateValue(info *model.Callinfo, band model.Band, mode model.Mode) bool {
	return c.collector.UpdateValue(info, band, mode)
}

func (c *Callinfo) InputChanged(vfo model.VFOID, call string, band model.Band, mode model.Mode, currentExchange []string) {
	normalizedCall := normalizeInput(call)

	callinfo := c.collector.GetInfoForInput(normalizedCall, band, mode, currentExchange)
	supercheck := c.supercheck.Calculate(normalizedCall, band, mode)

	frame := &c.frames[vfo]
	frame.NormalizedCallInput = normalizedCall
	frame.DXCCEntity = callinfo.DXCCEntity
	frame.Azimuth = callinfo.Azimuth
	frame.Distance = callinfo.Distance
	frame.UserInfo = callinfo.UserText

	frame.Points = callinfo.Points
	frame.Multis = callinfo.Multis
	frame.Value = callinfo.Value
	frame.SentQTCs = callinfo.SentQTCs
	frame.ReceivedQTCs = callinfo.ReceivedQTCs

	frame.PredictedExchange = callinfo.PredictedExchange
	frame.Supercheck = supercheck

	c.emitFrameChanged(vfo)
}

func (c *Callinfo) MarkerOnFrequency(vfo model.VFOID, marker model.BandmapMarker, available bool) {
	if int(vfo) < 0 || int(vfo) >= len(c.frames) {
		return
	}
	frame := &c.frames[vfo]

	text := ""
	if available {
		text = marker.Text
	}
	if frame.MarkerOnFrequency == text {
		return
	}
	frame.MarkerOnFrequency = text
	c.emitFrameChanged(vfo)
}

func (c *Callinfo) EntryOnFrequency(vfo model.VFOID, entry model.BandmapEntry, available bool) {
	if int(vfo) < 0 || int(vfo) >= len(c.frames) {
		return
	}
	frame := &c.frames[vfo]
	last := frame.CallsignOnFrequency.Callsign.String()
	if !available {
		frame.CallsignOnFrequency = model.AnnotatedCallsign{}
	} else if frame.CallsignOnFrequency.Callsign.String() == entry.Call.String() {
		// go on
	} else {
		exactMatch := frame.NormalizedCallInput == entry.Call.String()
		frame.CallsignOnFrequency = model.AnnotatedCallsign{
			Callsign:          entry.Call,
			Assembly:          model.MatchingAssembly{{OP: model.Matching, Value: entry.Call.String()}},
			Duplicate:         entry.Info.Duplicate,
			Worked:            entry.Info.Worked,
			ExactMatch:        exactMatch,
			Points:            entry.Info.Points,
			Multis:            entry.Info.Multis,
			PredictedExchange: entry.Info.PredictedExchange,
			OnFrequency:       true,
		}
	}

	if last != frame.CallsignOnFrequency.Callsign.String() {
		c.emitFrameChanged(vfo)
	}
}

// RadioChanged implements model.RadioChangedListener.
func (c *Callinfo) RadioChanged(_ string, singleVFO bool) {
	c.vfo2Enabled = !singleVFO
	c.view.SetVFOEnabled(model.VFO2, c.vfo2Enabled)
}

func normalizeInput(input string) string {
	return strings.TrimSpace(strings.ToUpper(input))
}

type nullView struct{}

func (v *nullView) Show()                                            {}
func (v *nullView) Hide()                                            {}
func (v *nullView) SetPredictedExchangeFields([]model.ExchangeField) {}
func (v *nullView) ShowFrame(model.VFOID, model.CallinfoFrame)       {}
func (v *nullView) SetVFOEnabled(model.VFOID, bool)                  {}
