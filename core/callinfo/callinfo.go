package callinfo

import (
	"fmt"
	"log"
	"strings"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

func New(entities DXCCFinder, callsigns CallsignFinder, callHistory CallHistoryFinder, dupeChecker DupeChecker, valuer Valuer) *Callinfo {
	result := &Callinfo{
		view:        new(nullView),
		entities:    entities,
		callsigns:   callsigns,
		callHistory: callHistory,
		dupeChecker: dupeChecker,
		valuer:      valuer,
	}

	return result
}

type Callinfo struct {
	view View

	entities    DXCCFinder
	callsigns   CallsignFinder
	callHistory CallHistoryFinder
	dupeChecker DupeChecker
	valuer      Valuer

	lastCallsign string
	lastBand     core.Band
	lastMode     core.Mode
	lastXchange  string
}

// DXCCFinder returns a list of matching prefixes for the given string and indicates if there was a match at all.
type DXCCFinder interface {
	Find(string) (dxcc.Prefix, bool)
}

// CallsignFinder returns a list of matching callsigns for the given partial string.
type CallsignFinder interface {
	FindStrings(string) ([]string, error)
	FindAnnotated(string) ([]core.AnnotatedMatch, error)
}

// CallHistoryFinder returns additional information for a given callsign if a call history file is used.
type CallHistoryFinder interface {
	FindEntry(string) (core.CallHistoryEntry, bool)
}

// DupeChecker can be used to find out if the given callsign was already worked, according to the contest rules.
type DupeChecker interface {
	FindWorkedQSOs(callsign.Callsign, core.Band, core.Mode) ([]core.QSO, bool)
}

// Valuer provides the points and multis of a QSO based on the given information.
type Valuer interface {
	Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, mode core.Mode, xchange string) (points, multis int)
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
	c.ShowInfo(c.lastCallsign, c.lastBand, c.lastMode, c.lastXchange)
}

func (c *Callinfo) Hide() {
	c.view.Hide()
}

func (c *Callinfo) ShowInfo(call string, band core.Band, mode core.Mode, xchange string) {
	c.lastCallsign = call
	c.lastBand = band
	c.lastMode = mode
	c.lastXchange = xchange
	worked := false
	duplicate := false
	cs, err := callsign.Parse(call)
	if err == nil {
		var qsos []core.QSO
		qsos, duplicate = c.dupeChecker.FindWorkedQSOs(cs, band, mode)
		worked = len(qsos) > 0
		if xchange == "" {
			xchange = c.predictXchange(call, qsos, true)
		}
	}
	c.view.SetCallsign(call, worked, duplicate)
	c.showDXCCAndValue(call, band, mode, xchange)
	c.showSupercheck(call)
}

func (c *Callinfo) showDXCCAndValue(call string, band core.Band, mode core.Mode, xchange string) {
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
	points, multis := c.valuer.Value(parsedCall, entity, band, mode, xchange)
	c.view.SetValue(points, multis, xchange)
}

func (c *Callinfo) showSupercheck(s string) {
	normalizedInput := strings.TrimSpace(strings.ToUpper(s))
	matches, err := c.callsigns.FindAnnotated(s)
	if err != nil {
		log.Printf("Callsign search for failed for %s: %v", s, err)
		return
	}

	annotatedMatches := make([]core.AnnotatedCallsign, len(matches))
	for i, match := range matches {
		matchString := match.String()
		cs, err := callsign.Parse(matchString)
		if err != nil {
			log.Printf("Supercheck match %s is not a valid callsign: %v", matchString, err)
			continue
		}
		exactMatch := (matchString == normalizedInput)

		qsos, duplicate := c.dupeChecker.FindWorkedQSOs(cs, c.lastBand, c.lastMode)
		predictedXchange := c.predictXchange(matchString, qsos, false)

		entity, entityFound := c.entities.Find(matchString)

		var points, multis int
		if entityFound {
			points, multis = c.valuer.Value(cs, entity, c.lastBand, c.lastMode, predictedXchange)
		}

		annotatedMatches[i] = core.AnnotatedCallsign{
			Callsign:         cs,
			Match:            match,
			Duplicate:        duplicate,
			Worked:           len(qsos) > 0,
			ExactMatch:       exactMatch,
			Points:           points,
			Multis:           multis,
			PredictedXchange: predictedXchange,
		}
	}

	c.view.SetSupercheck(annotatedMatches)
}

func (c *Callinfo) predictXchange(call string, qsos []core.QSO, exactMatch bool) string {
	log.Printf("predicting Xchange for %s", call)
	result := ""

	// TODO do not use the callHistory here, merge the callHistory result with the SCP result and provide the CallHistoryEntry here
	if exactMatch {
		entry, found := c.callHistory.FindEntry(call)
		if found {
			result = entry.PredictedXchange
		}
	}

	if len(qsos) == 0 {
		return result
	}

	var lastXchange string
	for _, qso := range qsos {
		if lastXchange == "" {
			lastXchange = qso.TheirXchange
		} else if lastXchange != qso.TheirXchange {
			return ""
		}
	}

	if lastXchange != "" {
		return lastXchange
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
