package callinfo

import (
	"fmt"
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

// ExchangeFilter clears the exchange values that cannot be predicted (RST, serial).
type ExchangeFilter interface {
	FilterExchange([]string) []string
}

// View defines the visual part of the call information window.
type View interface {
	SetBestMatchingCallsign(callsign core.AnnotatedCallsign)
	SetDXCC(string, string, int, int)
	SetValue(points, multis, value int)
	SetPredictedExchange(index int, text string)
	SetPredictedExchangeFields(fields []core.ExchangeField)
	SetUserInfo(string)
	SetSupercheck(callsigns []core.AnnotatedCallsign)
}

type Callinfo struct {
	view        View
	asyncRunner core.AsyncRunner
	collector   *Collector
	supercheck  *Supercheck

	lastCallsign        string
	lastBand            core.Band
	lastMode            core.Mode
	lastExchange        []string
	predictedExchange   []string
	theirExchangeFields []core.ExchangeField

	matchOnFrequency          core.AnnotatedCallsign
	matchOnFrequencyAvailable bool
	bestMatch                 core.AnnotatedCallsign
	bestMatchAvailable        bool
	bestMatches               []string
}

func New(entities DXCCFinder, callsigns CallsignFinder, callHistory CallHistoryFinder, dupeChecker DupeChecker, valuer Valuer, exchangeFilter ExchangeFilter, asyncRunner core.AsyncRunner) *Callinfo {
	result := &Callinfo{
		view:        new(nullView),
		asyncRunner: asyncRunner,
		collector:   NewCollector(entities, callsigns, callHistory, dupeChecker, valuer, exchangeFilter),
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

func (c *Callinfo) Refresh() {
	c.ShowInfo(c.lastCallsign, c.lastBand, c.lastMode, c.lastExchange)
}

func (c *Callinfo) ContestChanged(contest core.Contest) {
	if contest.Definition == nil {
		log.Printf("there is no contest definition!")
		return
	}
	c.theirExchangeFields = contest.TheirExchangeFields
	c.collector.SetTheirExchangeFields(c.theirExchangeFields)
	c.supercheck.SetTheirExchangeFields(c.theirExchangeFields)
	c.view.SetPredictedExchangeFields(c.theirExchangeFields)
}

func (c *Callinfo) ScoreUpdated(score core.Score) {
	c.collector.ScoreUpdated(score)
}

func (c *Callinfo) EntryOnFrequency(entry core.BandmapEntry, available bool) {
	c.asyncRunner(func() {
		c.matchOnFrequencyAvailable = available

		if available && c.matchOnFrequency.Callsign.String() == entry.Call.String() {
			// go on
		} else if available {
			normalizedCall := strings.TrimSpace(strings.ToUpper(c.lastCallsign))
			exactMatch := normalizedCall == entry.Call.String()
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

		c.showBestMatch()
	})
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

func (c *Callinfo) GetInfo(call callsign.Callsign, band core.Band, mode core.Mode, currentExchange []string) core.Callinfo {
	return c.collector.GetInfo(call, band, mode, currentExchange)
}

func (c *Callinfo) ShowInfo(call string, band core.Band, mode core.Mode, currentExchange []string) {
	c.lastCallsign = call
	c.lastBand = band
	c.lastMode = mode
	c.lastExchange = currentExchange

	callinfo := c.collector.GetInfoForInput(call, band, mode, currentExchange)
	if callinfo.CallValid {
		c.predictedExchange = callinfo.PredictedExchange
	} else {
		// TODO: should the currentExchange be used for the prediction at all?
		c.predictedExchange = currentExchange
	}

	supercheck := c.supercheck.Calculate(call, c.lastBand, c.lastMode)

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

	c.showDXCCEntity(callinfo)
	c.showBestMatch()
	c.view.SetUserInfo(callinfo.UserText)
	c.view.SetValue(callinfo.Points, callinfo.Multis, callinfo.Value)
	for i := range c.theirExchangeFields {
		text := ""
		if i < len(callinfo.FilteredExchange) {
			text = callinfo.FilteredExchange[i]
		}
		c.view.SetPredictedExchange(i+1, text)
	}
	c.view.SetSupercheck(supercheck)
}

func (c *Callinfo) GetValue(call callsign.Callsign, band core.Band, mode core.Mode) (points, multis int, multiValues map[conval.Property]string) {
	return c.collector.GetValue(call, band, mode)
}

func (c *Callinfo) showDXCCEntity(callinfo core.Callinfo) {
	var dxccName string
	if callinfo.DXCCEntity.PrimaryPrefix != "" {
		dxccName = fmt.Sprintf("%s (%s)", callinfo.DXCCEntity.Name, callinfo.DXCCEntity.PrimaryPrefix)
	}
	c.view.SetDXCC(dxccName, callinfo.DXCCEntity.Continent, int(callinfo.DXCCEntity.ITUZone), int(callinfo.DXCCEntity.CQZone))
}

func (c *Callinfo) findBestMatch() (core.AnnotatedCallsign, bool) {
	match := c.matchOnFrequency

	if c.bestMatchAvailable {
		match = c.bestMatch
	}

	return match, (match.Callsign.String() != "")
}

func (c *Callinfo) showBestMatch() {
	bestMatch, _ := c.findBestMatch()
	c.view.SetBestMatchingCallsign(bestMatch)
}

func normalizeInput(input string) string {
	return strings.TrimSpace(strings.ToUpper(input))
}

type nullView struct{}

func (v *nullView) Show()                                                   {}
func (v *nullView) Hide()                                                   {}
func (v *nullView) SetBestMatchingCallsign(callsign core.AnnotatedCallsign) {}
func (v *nullView) SetDXCC(string, string, int, int)                        {}
func (v *nullView) SetValue(int, int, int)                                  {}
func (v *nullView) SetPredictedExchange(int, string)                        {}
func (v *nullView) SetPredictedExchangeFields(fields []core.ExchangeField)  {}
func (v *nullView) SetUserInfo(string)                                      {}
func (v *nullView) SetSupercheck(callsigns []core.AnnotatedCallsign)        {}
