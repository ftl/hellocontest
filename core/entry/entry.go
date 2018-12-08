package entry

import (
	"fmt"
	logger "log"
	"strconv"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/parse"
)

// NewController returns a new EntryController.
func NewController(clock core.Clock, log core.Log) core.EntryController {
	return &controller{
		clock: clock,
		log:   log,
	}
}

type controller struct {
	clock        core.Clock
	log          core.Log
	view         core.EntryView
	activeField  core.EntryField
	selectedBand core.Band
	selectedMode core.Mode
}

func (c *controller) SetView(view core.EntryView) {
	c.view = view
	c.view.SetEntryController(c)
	c.Reset()
}

func (c *controller) GotoNextField() core.EntryField {
	switch c.activeField {
	case core.CallsignField:
		c.leaveCallsignField()
	}

	transitions := map[core.EntryField]core.EntryField{
		core.CallsignField:     core.TheirReportField,
		core.TheirReportField:  core.TheirNumberField,
		core.TheirNumberField:  core.TheirXchangeField,
		core.TheirXchangeField: core.CallsignField,
		core.MyReportField:     core.CallsignField,
		core.MyNumberField:     core.CallsignField,
	}
	c.activeField = transitions[c.activeField]
	c.view.SetActiveField(c.activeField)
	return c.activeField
}

func (c *controller) leaveCallsignField() {
	callsign, err := callsign.Parse(c.view.GetCallsign())
	if err != nil {
		fmt.Println(err)
		return
	}

	qso, found := c.log.Find(callsign)
	if !found {
		c.view.SetDuplicateMarker(false)
		return
	}

	c.view.SetBand(string(qso.Band))
	c.view.SetMode(string(qso.Mode))
	c.view.SetTheirReport(string(qso.TheirReport))
	c.view.SetTheirNumber(qso.TheirNumber.String())
	c.view.SetTheirXchange(qso.TheirXchange)
	c.view.SetMyReport(string(qso.MyReport))
	c.view.SetMyNumber(qso.MyNumber.String())
	c.view.SetMyXchange(qso.MyXchange)
	c.view.SetDuplicateMarker(true)
}

func (c *controller) GetActiveField() core.EntryField {
	return c.activeField
}

func (c *controller) SetActiveField(field core.EntryField) {
	c.activeField = field
}

func (c *controller) BandSelected(s string) {
	if band, err := parse.Band(s); err == nil {
		logger.Printf("Band selected: %v", band)
		c.selectedBand = band
	}
}

func (c *controller) ModeSelected(s string) {
	if mode, err := parse.Mode(s); err == nil {
		logger.Printf("Mode selected: %v", mode)
		c.selectedMode = mode

		if c.selectedMode == core.ModeSSB {
			c.view.SetTheirReport("59")
			c.view.SetMyReport("59")
		} else {
			c.view.SetMyReport("599")
			c.view.SetTheirReport("599")
		}
	}
}

func (c *controller) Log() {
	var err error
	qso := core.QSO{}
	qso.Callsign, err = callsign.Parse(c.view.GetCallsign())
	if err != nil {
		c.showErrorOnField(err, core.CallsignField)
		return
	}
	qso.Time = c.clock.Now()

	qso.Band, err = parse.Band(c.view.GetBand())
	if err != nil {
		c.view.ShowError(err)
		return
	}

	qso.Mode, err = parse.Mode(c.view.GetMode())
	if err != nil {
		c.view.ShowError(err)
		return
	}

	qso.TheirReport, err = parse.RST(c.view.GetTheirReport())
	if err != nil {
		c.showErrorOnField(err, core.TheirReportField)
		return
	}

	theirNumber, err := strconv.Atoi(c.view.GetTheirNumber())
	if err != nil {
		c.showErrorOnField(err, core.TheirNumberField)
		return
	}
	qso.TheirNumber = core.QSONumber(theirNumber)

	qso.TheirXchange = c.view.GetTheirXchange()

	qso.MyReport, err = parse.RST(c.view.GetMyReport())
	if err != nil {
		c.showErrorOnField(err, core.MyReportField)
		return
	}

	myNumber, err := strconv.Atoi(c.view.GetMyNumber())
	if err != nil {
		c.showErrorOnField(err, core.MyNumberField)
		return
	}
	qso.MyNumber = core.QSONumber(myNumber)

	qso.MyXchange = c.view.GetMyXchange()

	duplicateQso, duplicate := c.log.Find(qso.Callsign)
	if duplicate && duplicateQso.MyNumber != qso.MyNumber {
		c.showErrorOnField(fmt.Errorf("%s was worked before in QSO #%s", qso.Callsign, duplicateQso.MyNumber.String()), core.CallsignField)
		return
	}

	c.log.Log(qso)
	c.Reset()
}

func (c *controller) showErrorOnField(err error, field core.EntryField) {
	c.activeField = field
	c.view.SetActiveField(c.activeField)
	c.view.ShowError(err)
}

func (c *controller) Reset() {
	nextNumber := c.log.GetNextNumber()
	c.activeField = core.CallsignField
	c.view.SetCallsign("")
	if c.selectedMode == core.ModeSSB {
		c.view.SetTheirReport("59")
		c.view.SetMyReport("59")
	} else {
		c.view.SetMyReport("599")
		c.view.SetTheirReport("599")
	}
	c.view.SetTheirNumber("")
	c.view.SetTheirXchange("")
	if c.selectedBand != core.NoBand {
		c.view.SetBand(c.selectedBand.String())
	}
	if c.selectedMode != core.NoMode {
		c.view.SetMode(c.selectedMode.String())
	}
	c.view.SetMyNumber(nextNumber.String())
	c.view.SetActiveField(c.activeField)
	c.view.SetDuplicateMarker(false)
	c.view.ClearError()
}
