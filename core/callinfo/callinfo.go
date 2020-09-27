package callinfo

import (
	"fmt"
	"log"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

func New(prefixes core.DXCCFinder, callsigns core.CallsignFinder) *Callinfo {
	result := &Callinfo{
		prefixes:  prefixes,
		callsigns: callsigns,
	}

	return result
}

type Callinfo struct {
	view View

	prefixes  core.DXCCFinder
	callsigns core.CallsignFinder
	dupCheck  core.DupChecker
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

func (c *Callinfo) SetDupChecker(dupChecker core.DupChecker) {
	c.dupCheck = dupChecker
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
	var duplicate bool
	cs, err := callsign.Parse(s)
	if err == nil {
		_, duplicate = c.dupCheck.IsDuplicate(cs)
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
		_, duplicate := c.dupCheck.IsDuplicate(cs)
		annotatedMatches[i] = core.AnnotatedCallsign{
			Callsign:  cs,
			Duplicate: duplicate,
		}
	}

	c.view.SetSupercheck(annotatedMatches)
}
