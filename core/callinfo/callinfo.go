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
type DupeChecker interface {
	FindWorkedQSOs(callsign.Callsign, core.Band, core.Mode) ([]core.QSO, bool)
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

type Callinfo struct {
	view        View
	asyncRunner core.AsyncRunner
	collector   *Collector
	supercheck  *Supercheck

	predictedExchange   []string
	theirExchangeFields []core.ExchangeField

	frame                     core.CallinfoFrame
	matchOnFrequency          core.AnnotatedCallsign
	matchOnFrequencyAvailable bool
	bestMatch                 core.AnnotatedCallsign
	bestMatchAvailable        bool
	bestMatches               []string
}

func New(entities DXCCFinder, callsigns CallsignFinder, callHistory CallHistoryFinder, dupeChecker DupeChecker, valuer Valuer, asyncRunner core.AsyncRunner) *Callinfo {
	result := &Callinfo{
		view:        new(nullView),
		asyncRunner: asyncRunner,
		collector:   NewCollector(entities, callsigns, callHistory, dupeChecker, valuer),
		supercheck:  NewSupercheck(entities, callsigns, callHistory, dupeChecker, valuer),
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

func (c *Callinfo) GetInfo(call callsign.Callsign, band core.Band, mode core.Mode, currentExchange []string) core.Callinfo {
	return c.collector.GetInfo(call, band, mode, currentExchange)
}

func (c *Callinfo) GetValue(call callsign.Callsign, band core.Band, mode core.Mode) (points, multis int, multiValues map[conval.Property]string) {
	return c.collector.GetValue(call, band, mode)
}

func (c *Callinfo) ShowInfo(call string, band core.Band, mode core.Mode, currentExchange []string) {
	normalizedCall := normalizeInput(call)

	callinfo := c.collector.GetInfoForInput(normalizedCall, band, mode, currentExchange)
	if callinfo.CallValid {
		c.predictedExchange = callinfo.PredictedExchange
	} else {
		// TODO: should the currentExchange be used for the prediction at all?
		c.predictedExchange = currentExchange
	}

	supercheck := c.supercheck.Calculate(normalizedCall, band, mode)
	c.bestMatches = make([]string, 0, len(supercheck))
	c.bestMatch = core.AnnotatedCallsign{}
	c.bestMatchAvailable = false
	for i, match := range supercheck {
		c.bestMatches = append(c.bestMatches, match.Callsign.String())
		if i == 0 {
			c.bestMatch = match
			c.bestMatchAvailable = true
		}
	}
	c.setBestMatch()

	c.frame.NormalizedCallInput = normalizedCall
	c.frame.DXCCEntity = callinfo.DXCCEntity
	c.frame.UserInfo = callinfo.UserText

	c.frame.Points = callinfo.Points
	c.frame.Multis = callinfo.Multis
	c.frame.Value = callinfo.Value

	c.frame.PredictedExchange = callinfo.PredictedExchange
	c.frame.Supercheck = supercheck

	c.view.ShowFrame(c.frame)
}

func (c *Callinfo) setBestMatch() {
	bestMatch, _ := c.findBestMatch()
	c.frame.BestMatchingCallsign = bestMatch
}

func (c *Callinfo) findBestMatch() (core.AnnotatedCallsign, bool) {
	match := c.matchOnFrequency

	if c.bestMatchAvailable {
		match = c.bestMatch
	}

	return match, (match.Callsign.String() != "")
}

func (c *Callinfo) BestMatches() []string {
	return c.bestMatches
}

func (c *Callinfo) BestMatch() string {
	bestMatch, available := c.findBestMatch()
	if !available {
		return ""
	}
	return bestMatch.Callsign.String()
}

func (c *Callinfo) PredictedExchange() []string {
	return c.predictedExchange
}

func (c *Callinfo) EntryOnFrequency(entry core.BandmapEntry, available bool) {
	c.asyncRunner(func() {
		c.matchOnFrequencyAvailable = available

		if available && c.matchOnFrequency.Callsign.String() == entry.Call.String() {
			// go on
		} else if available {
			exactMatch := c.frame.NormalizedCallInput == entry.Call.String()
			c.matchOnFrequency = core.AnnotatedCallsign{
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
		} else {
			c.matchOnFrequency = core.AnnotatedCallsign{}
		}

		c.setBestMatch()
		c.view.ShowFrame(c.frame) // TODO
	})
}

func normalizeInput(input string) string {
	return strings.TrimSpace(strings.ToUpper(input))
}

type nullView struct{}

func (v *nullView) Show()                                           {}
func (v *nullView) Hide()                                           {}
func (v *nullView) SetPredictedExchangeFields([]core.ExchangeField) {}
func (v *nullView) ShowFrame(frame core.CallinfoFrame)              {}
