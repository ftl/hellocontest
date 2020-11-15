package callinfo

import (
	"fmt"
	"log"
	"strings"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

func New(prefixes DXCCFinder, callsigns CallsignFinder, dupeChecker DupeChecker) *Callinfo {
	result := &Callinfo{
		view:        new(nullView),
		prefixes:    prefixes,
		callsigns:   callsigns,
		dupeChecker: dupeChecker,
	}

	return result
}

type Callinfo struct {
	view View

	prefixes    DXCCFinder
	callsigns   CallsignFinder
	dupeChecker DupeChecker
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
	IsWorked(callsign callsign.Callsign) ([]core.QSO, bool)
}

// View defines the visual part of the call information window.
type View interface {
	Show()
	Hide()

	SetCallsign(callsign string, worked, duplicate bool)
	SetDXCC(string, string, int, int, bool)
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
}

func (c *Callinfo) Hide() {
	c.view.Hide()
}

func (c *Callinfo) ShowCallsign(s string) {
	worked := false
	duplicate := false
	cs, err := callsign.Parse(s)
	if err == nil {
		var qsos []core.QSO
		qsos, duplicate = c.dupeChecker.IsWorked(cs)
		worked = len(qsos) > 0
	}
	c.view.SetCallsign(s, worked, duplicate)
	c.showDXCC(s)
	c.showSupercheck(s)
}

func (c *Callinfo) showDXCC(callsign string) {
	if c.prefixes == nil {
		c.view.SetDXCC("", "", 0, 0, false)
		return
	}
	prefix, found := c.prefixes.Find(callsign)
	if !found {
		c.view.SetDXCC("", "", 0, 0, false)
		return
	}
	dxccName := fmt.Sprintf("%s (%s)", prefix.Name, prefix.PrimaryPrefix)
	c.view.SetDXCC(dxccName, prefix.Continent, int(prefix.ITUZone), int(prefix.CQZone), !prefix.NotARRLCompliant)
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
		qsos, duplicate := c.dupeChecker.IsWorked(cs)
		exactMatch := (match == normalizedInput)
		annotatedMatches[i] = core.AnnotatedCallsign{
			Callsign:   cs,
			Duplicate:  duplicate,
			Worked:     len(qsos) > 0,
			ExactMatch: exactMatch,
		}
	}

	c.view.SetSupercheck(annotatedMatches)
}

type nullView struct{}

func (v *nullView) Show()                                               {}
func (v *nullView) Hide()                                               {}
func (v *nullView) SetCallsign(callsign string, worked, duplicate bool) {}
func (v *nullView) SetDXCC(string, string, int, int, bool)              {}
func (v *nullView) SetSupercheck(callsigns []core.AnnotatedCallsign)    {}
