package callinfo

import (
	"fmt"
	"log"

	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
)

func NewController() core.CallinfoController {
	result := &callinfo{
		prefixes:   setupDXCC(),
		supercheck: setupSupercheck(),
	}

	return result
}

func setupDXCC() *dxcc.Prefixes {
	localFilename, err := dxcc.LocalFilename()
	if err != nil {
		log.Print(err)
		return nil
	}
	updated, err := dxcc.Update(dxcc.DefaultURL, localFilename)
	if err != nil {
		log.Printf("update of local copy of DXCC prefixes failed: %v", err)
	}
	if updated {
		log.Printf("updated local copy of DXCC prefixes: %v", localFilename)
	}

	result, err := dxcc.LoadLocal(localFilename)
	if err != nil {
		log.Printf("cannot load DXCC prefixes: %v", err)
		return nil
	}
	return result
}

func setupSupercheck() *scp.Database {
	localFilename, err := scp.LocalFilename()
	if err != nil {
		log.Print(err)
		return nil
	}
	updated, err := scp.Update(scp.DefaultURL, localFilename)
	if err != nil {
		log.Printf("update of local copy of Supercheck database failed: %v", err)
	}
	if updated {
		log.Printf("updated local copy of Supercheck database: %v", localFilename)
	}

	result, err := scp.LoadLocal(localFilename)
	if err != nil {
		log.Printf("cannot load Supercheck database: %v", err)
		return nil
	}
	return result
}

type callinfo struct {
	view core.CallinfoView

	prefixes   *dxcc.Prefixes
	supercheck *scp.Database
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
	c.view.SetCallsign(callsign)
	c.showDXCC(callsign)
	c.showSupercheck(callsign)
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
	matches, err := c.supercheck.Find(callsign)
	if err != nil {
		log.Printf("Supercheck failed for %s: %v", callsign, err)
		return
	}

	c.view.SetSupercheck(matches)
}
