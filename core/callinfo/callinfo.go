package callinfo

import (
	"fmt"
	"log"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

func New(prefixes DXCCFinder, callsigns CallsignFinder) *Callinfo {
	result := &Callinfo{
		prefixes:  prefixes,
		callsigns: callsigns,
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
	Find(string) ([]dxcc.Prefix, bool)
}

// CallsignFinder returns a list of matching callsigns for the given partial string.
type CallsignFinder interface {
	Find(string) ([]string, error)
}

// DupeChecker can be used to find out if the given callsign was already worked, according to the contest rules.
type DupeChecker interface {
	IsDuplicate(callsign callsign.Callsign) (core.QSO, bool)
}

// View defines the visual part of the call information window.
type View interface {
	Show()
	Hide()

	SetCallsign(string)
	SetDuplicateMarker(bool)
	SetDXCC(string, string, int, int, bool)
	SetSupercheck(callsigns []core.AnnotatedCallsign)
}

func (c *Callinfo) SetView(view View) {
	c.view = view
}

func (c *Callinfo) SetDupeChecker(dupeChecker DupeChecker) {
	c.dupeChecker = dupeChecker
}

func (c *Callinfo) Show() {
	if c.view == nil {
		return
	}
	c.view.Show()
}

func (c *Callinfo) Hide() {
	if c.view == nil {
		return
	}
	c.view.Hide()
}

func (c *Callinfo) ShowCallsign(s string) {
	if c.view == nil {
		return
	}
	var duplicate bool
	cs, err := callsign.Parse(s)
	if err == nil {
		_, duplicate = c.dupeChecker.IsDuplicate(cs)
	}

	c.view.SetDuplicateMarker(duplicate)
	c.view.SetCallsign(s)
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
	if len(prefix) != 1 {
		c.view.SetDXCC("", "", 0, 0, false)
		return
	}
	dxccName := fmt.Sprintf("%s (%s)", prefix[0].Name, prefix[0].PrimaryPrefix)
	c.view.SetDXCC(dxccName, prefix[0].Continent, int(prefix[0].ITUZone), int(prefix[0].CQZone), !prefix[0].NotARRLCompliant)
}

func (c *Callinfo) showSupercheck(s string) {
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
		_, duplicate := c.dupeChecker.IsDuplicate(cs)
		annotatedMatches[i] = core.AnnotatedCallsign{
			Callsign:  cs,
			Duplicate: duplicate,
		}
	}

	c.view.SetSupercheck(annotatedMatches)
}
