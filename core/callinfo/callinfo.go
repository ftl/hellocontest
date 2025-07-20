package callinfo

import (
	"log"
	"strings"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
)

// DXCCFinder returns a list of matching prefixes for the given string and indicates if there was a match at all.
type DXCCFinder interface {
	Find(string) (dxcc.Prefix, bool)
}

// CallsignFinder returns a list of matching callsigns for the given partial string.
type CallsignFinder interface {
	FindStrings(string) ([]string, error)
	Find(string) ([]core.AnnotatedCallsign, error)
}

// CallHistoryFinder returns additional information for a given callsign if a call history file is used.
type CallHistoryFinder interface {
	FindEntry(string) (core.AnnotatedCallsign, bool)
	Find(string) ([]core.AnnotatedCallsign, error)
}

// DupeChecker can be used to find out if the given callsign was already worked, according to the contest rules.
// It can also find worked callsigns that are similar to a given string, e.g. for supercheck.
type DupeChecker interface {
	FindWorkedQSOs(callsign.Callsign, core.Band, core.Mode) ([]core.QSO, bool)
	Find(string) ([]core.AnnotatedCallsign, error)
}

// Valuer provides the points and multis of a QSO based on the given information.
type Valuer interface {
	Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, mode core.Mode, exchange []string) (points, multis int, multiValues map[conval.Property]string)
}

// View defines the visual part of the call information window.
type View interface {
	SetPredictedExchangeFields(fields []core.ExchangeField)
	ShowFrame(core.CallinfoFrame)
}

type CallinfoFrameListener interface {
	CallinfoFrameChanged(core.CallinfoFrame)
}

type Callinfo struct {
	view       View
	collector  *Collector
	supercheck *Supercheck
	listeners  []any

	theirExchangeFields []core.ExchangeField

	frame core.CallinfoFrame
}

func New(entities DXCCFinder, callsigns CallsignFinder, callHistory CallHistoryFinder, dupeChecker DupeChecker, valuer Valuer) *Callinfo {
	result := &Callinfo{
		view:       new(nullView),
		collector:  NewCollector(entities, callsigns, callHistory, dupeChecker, valuer),
		supercheck: NewSupercheck(entities, callsigns, callHistory, dupeChecker, valuer),
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
	c.view.SetPredictedExchangeFields(c.theirExchangeFields)
}

func (c *Callinfo) StationChanged(station core.Station) {
	c.collector.SetMyLocator(station.Locator)
}

func (c *Callinfo) ContestChanged(contest core.Contest) {
	if contest.Definition == nil {
		log.Printf("there is no contest definition!")
		return
	}
	c.theirExchangeFields = contest.TheirExchangeFields
	c.collector.SetTheirExchangeFields(c.theirExchangeFields, contest.TheirReportExchangeField, contest.TheirNumberExchangeField)
	c.supercheck.SetTheirExchangeFields(c.theirExchangeFields)
	c.view.SetPredictedExchangeFields(c.theirExchangeFields)
}

func (c *Callinfo) ScoreUpdated(score core.Score) {
	c.collector.ScoreUpdated(score)
}

func (c *Callinfo) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Callinfo) emitFrameChanged() {
	for _, listener := range c.listeners {
		if l, ok := listener.(CallinfoFrameListener); ok {
			l.CallinfoFrameChanged(c.frame)
		}
	}
	c.view.ShowFrame(c.frame)
}

func (c *Callinfo) GetInfo(call callsign.Callsign, band core.Band, mode core.Mode, currentExchange []string) core.Callinfo {
	return c.collector.GetInfo(call, band, mode, currentExchange)
}

func (c *Callinfo) UpdateValue(info *core.Callinfo, band core.Band, mode core.Mode) bool {
	return c.collector.UpdateValue(info, band, mode)
}

func (c *Callinfo) InputChanged(call string, band core.Band, mode core.Mode, currentExchange []string) {
	normalizedCall := normalizeInput(call)

	callinfo := c.collector.GetInfoForInput(normalizedCall, band, mode, currentExchange)
	supercheck := c.supercheck.Calculate(normalizedCall, band, mode)

	c.frame.NormalizedCallInput = normalizedCall
	c.frame.DXCCEntity = callinfo.DXCCEntity
	c.frame.Azimuth = callinfo.Azimuth
	c.frame.Distance = callinfo.Distance
	c.frame.UserInfo = callinfo.UserText

	c.frame.Points = callinfo.Points
	c.frame.Multis = callinfo.Multis
	c.frame.Value = callinfo.Value

	c.frame.PredictedExchange = callinfo.PredictedExchange
	c.frame.Supercheck = supercheck

	c.emitFrameChanged()
}

func (c *Callinfo) EntryOnFrequency(entry core.BandmapEntry, available bool) {
	last := c.frame.CallsignOnFrequency.Callsign.String()
	if !available {
		c.frame.CallsignOnFrequency = core.AnnotatedCallsign{}
	} else if c.frame.CallsignOnFrequency.Callsign.String() == entry.Call.String() {
		// go on
	} else {
		exactMatch := c.frame.NormalizedCallInput == entry.Call.String()
		c.frame.CallsignOnFrequency = core.AnnotatedCallsign{
			Callsign:          entry.Call,
			Assembly:          core.MatchingAssembly{{OP: core.Matching, Value: entry.Call.String()}},
			Duplicate:         entry.Info.Duplicate,
			Worked:            entry.Info.Worked,
			ExactMatch:        exactMatch,
			Points:            entry.Info.Points,
			Multis:            entry.Info.Multis,
			PredictedExchange: entry.Info.PredictedExchange,
			OnFrequency:       true,
		}
	}

	if last != c.frame.CallsignOnFrequency.Callsign.String() {
		c.emitFrameChanged()
	}
}

func normalizeInput(input string) string {
	return strings.TrimSpace(strings.ToUpper(input))
}

type nullView struct{}

func (v *nullView) Show()                                           {}
func (v *nullView) Hide()                                           {}
func (v *nullView) SetPredictedExchangeFields([]core.ExchangeField) {}
func (v *nullView) ShowFrame(frame core.CallinfoFrame)              {}
