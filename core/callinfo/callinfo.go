package callinfo

import (
	"fmt"
	"log"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

func NewController(prefixes core.DXCCFinder, callsigns core.CallsignFinder, dupCheck core.DupChecker) core.CallinfoController {
	result := &callinfo{
		prefixes:  prefixes,
		callsigns: callsigns,
		dupCheck:  dupCheck,
	}

	return result
}

type callinfo struct {
	view core.CallinfoView

	prefixes  core.DXCCFinder
	callsigns core.CallsignFinder
	dupCheck  core.DupChecker
}

func (c *callinfo) SetView(view core.CallinfoView) {
	c.view = view
}

func (c *callinfo) Show() {
	if c.view == nil {
		return
	}
	c.view.Show()
}

func (c *callinfo) Hide() {
	if c.view == nil {
		return
	}
	c.view.Hide()
}

func (c *callinfo) ShowCallsign(s string) {
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

func (c *callinfo) showDXCC(callsign string) {
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

func (c *callinfo) showSupercheck(callsign string) {
	matches, err := c.callsigns.Find(callsign)
	if err != nil {
		log.Printf("Callsign search for failed for %s: %v", callsign, err)
		return
	}

	c.view.SetSupercheck(matches)
}
