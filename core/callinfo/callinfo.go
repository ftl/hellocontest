package callinfo

import (
	"fmt"
	"log"

	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

func NewController() core.CallinfoController {
	result := new(callinfo)

	localFilename, err := dxcc.LocalFilename()
	if err != nil {
		log.Fatal(err)
	}
	updated, err := dxcc.Update(dxcc.DefaultURL, localFilename)
	if err != nil {
		log.Printf("update of local copy failed: %v\n", err)
	}
	if updated {
		log.Printf("updated local copy: %v\n", localFilename)
	}

	prefixes, err := dxcc.LoadLocal(localFilename)
	if err != nil {
		log.Printf("cannot load prefixes: %v", err)
	} else {
		result.prefixes = prefixes
	}

	return result
}

type callinfo struct {
	view core.CallinfoView

	prefixes *dxcc.Prefixes
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

func (c *callinfo) ShowCallsign(callsign string) {
	log.Printf("Callinfo for %s", callsign)
	c.view.SetCallsign(callsign)
	c.showDXCC(callsign)
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
