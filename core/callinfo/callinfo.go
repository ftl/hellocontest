package callinfo

import (
	"fmt"
	"log"
	"strings"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

func New(entities DXCCFinder, callsigns CallsignFinder, dupeChecker DupeChecker, valuer Valuer) *Callinfo {
	result := &Callinfo{
		view:        new(nullView),
		entities:    entities,
		callsigns:   callsigns,
		dupeChecker: dupeChecker,
		valuer:      valuer,
	}

	return result
}

type Callinfo struct {
	view View

	entities    DXCCFinder
	callsigns   CallsignFinder
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
	Find(string) ([]string, error)
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
	SetValue(points, multis int)
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
	}
	c.view.SetCallsign(call, worked, duplicate)
	c.showDXCCAndValue(call, band, mode, xchange)
	c.showSupercheck(call)
}

func (c *Callinfo) showDXCCAndValue(call string, band core.Band, mode core.Mode, xchange string) {
	if c.entities == nil {
		c.view.SetDXCC("", "", 0, 0, false)
		c.view.SetValue(0, 0)
		return
	}
	entity, found := c.entities.Find(call)
	if !found {
		c.view.SetDXCC("", "", 0, 0, false)
		c.view.SetValue(0, 0)
		return
	}
	parsedCall, err := callsign.Parse(call)
	if err != nil {
		parsedCall = callsign.Callsign{}
	}

	dxccName := fmt.Sprintf("%s (%s)", entity.Name, entity.PrimaryPrefix)
	c.view.SetDXCC(dxccName, entity.Continent, int(entity.ITUZone), int(entity.CQZone), !entity.NotARRLCompliant)
	points, multis := c.valuer.Value(parsedCall, entity, band, mode, xchange)
	c.view.SetValue(points, multis)
}

func (c *Callinfo) showSupercheck(s string) {
	normalizedInput := strings.TrimSpace(strings.ToUpper(s))
	matches, err := c.callsigns.Find(s)
	if err != nil {
		log.Printf("Callsign search for failed for %s: %v", s, err)
		return
	}

	annotatedMatches := make([]core.AnnotatedCallsign, len(matches))
	for i, match := range matches {
		cs, err := callsign.Parse(match)
		if err != nil {
			log.Printf("Supercheck match %s is not a valid callsign: %v", match, err)
			continue
		}
		entity, entityFound := c.entities.Find(match)
		var points, multis int
		if entityFound {
			points, multis = c.valuer.Value(cs, entity, c.lastBand, c.lastMode, "") // TODO predict exchange
		}
		qsos, duplicate := c.dupeChecker.FindWorkedQSOs(cs, c.lastBand, c.lastMode)
		exactMatch := (match == normalizedInput)
		annotatedMatches[i] = core.AnnotatedCallsign{
			Callsign:   cs,
			Duplicate:  duplicate,
			Worked:     len(qsos) > 0,
			ExactMatch: exactMatch,
			Points:     points,
			Multis:     multis,
		}
	}

	c.view.SetSupercheck(annotatedMatches)
}

type nullView struct{}

func (v *nullView) Show()                                               {}
func (v *nullView) Hide()                                               {}
func (v *nullView) SetCallsign(callsign string, worked, duplicate bool) {}
func (v *nullView) SetDXCC(string, string, int, int, bool)              {}
func (v *nullView) SetValue(points, multis int)                         {}
func (v *nullView) SetSupercheck(callsigns []core.AnnotatedCallsign)    {}
