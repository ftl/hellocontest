package callinfo

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

func New(entities DXCCFinder, callsigns CallsignFinder, callHistory CallHistoryFinder, dupeChecker DupeChecker, valuer Valuer, exchangeFilter ExchangeFilter) *Callinfo {
	result := &Callinfo{
		view:           new(nullView),
		entities:       entities,
		callsigns:      callsigns,
		callHistory:    callHistory,
		dupeChecker:    dupeChecker,
		valuer:         valuer,
		exchangeFilter: exchangeFilter,
	}

	return result
}

type Callinfo struct {
	view View

	entities       DXCCFinder
	callsigns      CallsignFinder
	callHistory    CallHistoryFinder
	dupeChecker    DupeChecker
	valuer         Valuer
	exchangeFilter ExchangeFilter

	lastCallsign      string
	lastBand          core.Band
	lastMode          core.Mode
	lastExchange      []string
	predictedExchange []string
}

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
	Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, mode core.Mode, exchange []string) (points, multis int)
}

// ExchangeFilter clears the exchange values that cannot be predicted (RST, serial).
type ExchangeFilter interface {
	FilterExchange([]string) []string
}

// View defines the visual part of the call information window.
type View interface {
	Show()
	Hide()

	SetCallsign(callsign string, worked, duplicate bool)
	SetDXCC(string, string, int, int, bool)
	SetValue(points, multis int, xchange string)
	SetSupercheck(callsigns []core.AnnotatedCallsign)
}

func (c *Callinfo) SetView(view View) {
	if view == nil {
		c.view = new(nullView)
		return
	}
	c.view = view
}

func (c *Callinfo) Show() {
	c.view.Show()
	c.ShowInfo(c.lastCallsign, c.lastBand, c.lastMode, c.lastExchange)
}

func (c *Callinfo) Hide() {
	c.view.Hide()
}

func (c *Callinfo) PredictedExchange() []string {
	return c.predictedExchange
}

func (c *Callinfo) ShowInfo(call string, band core.Band, mode core.Mode, exchange []string) {
	c.lastCallsign = call
	c.lastBand = band
	c.lastMode = mode
	c.lastExchange = exchange
	worked := false
	duplicate := false
	cs, err := callsign.Parse(call)
	if err == nil {
		var qsos []core.QSO
		qsos, duplicate = c.dupeChecker.FindWorkedQSOs(cs, band, mode)
		worked = len(qsos) > 0

		entry, found := c.callHistory.FindEntry(call)
		var historicExchange []string
		if found {
			historicExchange = entry.PredictedExchange
		}

		exchange = c.predictExchange(call, qsos, historicExchange)
	}
	c.predictedExchange = exchange

	c.view.SetCallsign(call, worked, duplicate)
	c.showDXCCAndValue(call, band, mode, exchange)
	c.showSupercheck(call)
}

func (c *Callinfo) showDXCCAndValue(call string, band core.Band, mode core.Mode, exchange []string) {
	if c.entities == nil {
		c.view.SetDXCC("", "", 0, 0, false)
		c.view.SetValue(0, 0, "")
		return
	}
	entity, found := c.entities.Find(call)
	if !found {
		c.view.SetDXCC("", "", 0, 0, false)
		c.view.SetValue(0, 0, "")
		return
	}
	parsedCall, err := callsign.Parse(call)
	if err != nil {
		parsedCall = callsign.Callsign{}
	}

	dxccName := fmt.Sprintf("%s (%s)", entity.Name, entity.PrimaryPrefix)
	c.view.SetDXCC(dxccName, entity.Continent, int(entity.ITUZone), int(entity.CQZone), !entity.NotARRLCompliant)
	points, multis := c.valuer.Value(parsedCall, entity, band, mode, exchange)

	exchange = c.exchangeFilter.FilterExchange(exchange)
	exchangeText := strings.Join(exchange, " ")

	c.view.SetValue(points, multis, exchangeText)
}

func (c *Callinfo) showSupercheck(s string) {
	normalizedInput := strings.TrimSpace(strings.ToUpper(s))
	scpMatches, err := c.callsigns.Find(s)
	if err != nil {
		log.Printf("Callsign search for failed for %s: %v", s, err)
		return
	}
	historicMatches, _ := c.callHistory.Find(s)

	annotatedCallsigns := make(map[callsign.Callsign]core.AnnotatedCallsign, len(scpMatches)+len(historicMatches))
	for _, match := range scpMatches {
		annotatedCallsigns[match.Callsign] = match
	}
	for _, match := range historicMatches {
		var annotatedCallsign core.AnnotatedCallsign
		storedCallsign, found := annotatedCallsigns[match.Callsign]
		if found {
			annotatedCallsign = storedCallsign
		} else {
			annotatedCallsign = match
		}
		annotatedCallsign.PredictedExchange = match.PredictedExchange
		annotatedCallsigns[annotatedCallsign.Callsign] = annotatedCallsign
	}

	result := make([]core.AnnotatedCallsign, 0, len(annotatedCallsigns))
	for _, annotatedCallsign := range annotatedCallsigns {
		matchString := annotatedCallsign.Callsign.String()
		exactMatch := (matchString == normalizedInput)

		qsos, duplicate := c.dupeChecker.FindWorkedQSOs(annotatedCallsign.Callsign, c.lastBand, c.lastMode)
		predictedExchange := c.predictExchange(matchString, qsos, annotatedCallsign.PredictedExchange)

		entity, entityFound := c.entities.Find(matchString)

		var points, multis int
		if entityFound {
			points, multis = c.valuer.Value(annotatedCallsign.Callsign, entity, c.lastBand, c.lastMode, predictedExchange)
		}

		annotatedCallsign.Duplicate = duplicate
		annotatedCallsign.Worked = len(qsos) > 0
		annotatedCallsign.ExactMatch = exactMatch
		annotatedCallsign.Points = points
		annotatedCallsign.Multis = multis
		annotatedCallsign.PredictedExchange = predictedExchange

		result = append(result, annotatedCallsign)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LessThan(result[j])
	})

	c.view.SetSupercheck(result)
}

func (c *Callinfo) predictExchange(call string, qsos []core.QSO, historicExchange []string) []string {
	if len(qsos) == 0 {
		return historicExchange
	}

	result := make([]string, len(historicExchange))
	for i := range result {
		for _, qso := range qsos {
			if result[i] == "" {
				result[i] = qso.TheirExchange[i]
			} else if result[i] != qso.TheirExchange[i] {
				result[i] = ""
				break
			}
		}
		if result[i] == "" {
			result[i] = historicExchange[i]
		}
	}

	return result
}

type nullView struct{}

func (v *nullView) Show()                                               {}
func (v *nullView) Hide()                                               {}
func (v *nullView) SetCallsign(callsign string, worked, duplicate bool) {}
func (v *nullView) SetDXCC(string, string, int, int, bool)              {}
func (v *nullView) SetValue(int, int, string)                           {}
func (v *nullView) SetSupercheck(callsigns []core.AnnotatedCallsign)    {}
