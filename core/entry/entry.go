package entry

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/parse"
)

// NewController returns a new EntryController.
func NewController(clock core.Clock, logbook core.Logbook, enterTheirNumber, enterTheirXchange, allowMultiBand, allowMultiMode bool) core.EntryController {
	return &controller{
		clock:             clock,
		logbook:           logbook,
		enterTheirNumber:  enterTheirNumber,
		enterTheirXchange: enterTheirXchange,
		allowMultiBand:    allowMultiBand,
		allowMultiMode:    allowMultiMode,
		selectedBand:      logbook.LastBand(),
		selectedMode:      logbook.LastMode(),
	}
}

type controller struct {
	clock    core.Clock
	logbook  core.Logbook
	keyer    core.KeyerController
	callinfo core.CallinfoController

	enterTheirNumber  bool
	enterTheirXchange bool
	allowMultiBand    bool
	allowMultiMode    bool
	view              core.EntryView
	activeField       core.EntryField
	selectedBand      core.Band
	selectedMode      core.Mode
	editing           bool
	editQSO           core.QSO
}

func (c *controller) SetView(view core.EntryView) {
	c.view = view
	c.view.SetEntryController(c)
	c.view.SetBand(c.selectedBand.String())
	c.view.SetMode(c.selectedMode.String())
	c.view.EnableExchangeFields(c.enterTheirNumber, c.enterTheirXchange)
	c.Reset()
}

func (c *controller) SetKeyer(keyer core.KeyerController) {
	c.keyer = keyer
}

func (c *controller) SetCallinfo(callinfo core.CallinfoController) {
	c.callinfo = callinfo
}

func (c *controller) GotoNextField() core.EntryField {
	switch c.activeField {
	case core.CallsignField:
		c.leaveCallsignField()
	}

	transitions := map[core.EntryField]core.EntryField{
		core.CallsignField:     core.TheirReportField,
		core.TheirXchangeField: core.CallsignField,
		core.MyReportField:     core.CallsignField,
		core.MyNumberField:     core.CallsignField,
	}
	if c.enterTheirNumber && c.enterTheirXchange {
		transitions[core.TheirReportField] = core.TheirNumberField
		transitions[core.TheirNumberField] = core.TheirXchangeField
	} else if !c.enterTheirNumber && c.enterTheirXchange {
		transitions[core.TheirReportField] = core.TheirXchangeField
		transitions[core.TheirNumberField] = core.CallsignField
	} else if c.enterTheirNumber && !c.enterTheirXchange {
		transitions[core.TheirReportField] = core.TheirNumberField
		transitions[core.TheirNumberField] = core.CallsignField
	} else if !c.enterTheirNumber && !c.enterTheirXchange {
		transitions[core.TheirReportField] = core.CallsignField
		transitions[core.TheirNumberField] = core.CallsignField
	}
	c.activeField = transitions[c.activeField]
	c.view.SetActiveField(c.activeField)
	return c.activeField
}

func (c *controller) leaveCallsignField() {
	callsign, err := callsign.Parse(c.view.Callsign())
	if err != nil {
		fmt.Println(err)
		return
	}

	qso, found := c.isDuplicate(callsign)
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

func (c *controller) isDuplicate(callsign callsign.Callsign) (core.QSO, bool) {
	band := core.NoBand
	if c.allowMultiBand {
		band = c.selectedBand
	}
	mode := core.NoMode
	if c.allowMultiMode {
		mode = c.selectedMode
	}
	qsos := c.logbook.FindAll(callsign, band, mode)
	if len(qsos) == 0 {
		return core.QSO{}, false
	}
	return qsos[len(qsos)-1], true
}

func (c *controller) GetActiveField() core.EntryField {
	return c.activeField
}

func (c *controller) SetActiveField(field core.EntryField) {
	c.activeField = field
}

func (c *controller) BandSelected(s string) {
	if band, err := parse.Band(s); err == nil {
		log.Printf("Band selected: %v", band)
		c.selectedBand = band
		c.EnterCallsign(c.view.Callsign())
	}
}

func (c *controller) ModeSelected(s string) {
	if mode, err := parse.Mode(s); err == nil {
		log.Printf("Mode selected: %v", mode)
		c.selectedMode = mode

		if c.selectedMode == core.ModeSSB {
			c.view.SetTheirReport("59")
			c.view.SetMyReport("59")
		} else {
			c.view.SetMyReport("599")
			c.view.SetTheirReport("599")
		}

		c.EnterCallsign(c.view.Callsign())
	}
}

func (c *controller) SendQuestion() {
	if c.keyer == nil {
		return
	}

	switch c.activeField {
	case core.TheirReportField, core.TheirNumberField, core.TheirXchangeField:
		c.keyer.SendQuestion("nr")
	default:
		callsign := c.view.Callsign()
		c.keyer.SendQuestion(callsign)
	}
}

func (c *controller) EnterCallsign(s string) {
	if c.callinfo != nil {
		c.callinfo.ShowCallsign(s)
	}

	callsign, err := callsign.Parse(s)
	if err != nil {
		return
	}

	qso, found := c.isDuplicate(callsign)
	if !found {
		c.view.ClearMessage()
		return
	}

	c.view.ShowMessage(fmt.Sprintf("%s was worked before in QSO #%s", qso.Callsign, qso.MyNumber.String()))
}

func (c *controller) QSOSelected(qso core.QSO) {
	log.Printf("QSO selected: %v", qso)
	c.editing = true
	c.editQSO = qso

	c.view.SetBand(string(qso.Band))
	c.view.SetMode(string(qso.Mode))
	c.view.SetCallsign(qso.Callsign.String())
	c.view.SetTheirReport(string(qso.TheirReport))
	c.view.SetTheirNumber(qso.TheirNumber.String())
	c.view.SetTheirXchange(qso.TheirXchange)
	c.view.SetMyReport(string(qso.MyReport))
	c.view.SetMyNumber(qso.MyNumber.String())
	c.view.SetMyXchange(qso.MyXchange)
	c.view.SetActiveField(core.CallsignField)
	c.view.SetEditingMarker(true)
}

func (c *controller) Log() {
	var err error
	qso := core.QSO{}
	qso.Callsign, err = callsign.Parse(c.view.Callsign())
	if err != nil {
		c.showErrorOnField(err, core.CallsignField)
		return
	}
	if c.editing {
		qso.Time = c.editQSO.Time
	} else {
		qso.Time = c.clock.Now()
	}

	qso.Band, err = parse.Band(c.view.Band())
	if err != nil {
		c.view.ShowMessage(err)
		return
	}

	qso.Mode, err = parse.Mode(c.view.Mode())
	if err != nil {
		c.view.ShowMessage(err)
		return
	}

	qso.TheirReport, err = parse.RST(c.view.TheirReport())
	if err != nil {
		c.showErrorOnField(err, core.TheirReportField)
		return
	}

	if c.enterTheirNumber {
		value := c.view.TheirNumber()
		if value == "" {
			c.showErrorOnField(errors.New("their number is missing"), core.TheirNumberField)
			return
		}

		theirNumber, err := strconv.Atoi(value)
		if err != nil {
			c.showErrorOnField(err, core.TheirNumberField)
			return
		}
		qso.TheirNumber = core.QSONumber(theirNumber)
	}

	if c.enterTheirXchange {
		qso.TheirXchange = c.view.TheirXchange()
		if qso.TheirXchange == "" {
			c.showErrorOnField(errors.New("their exchange is missing"), core.TheirXchangeField)
		}
	}

	qso.MyReport, err = parse.RST(c.view.MyReport())
	if err != nil {
		c.showErrorOnField(err, core.MyReportField)
		return
	}

	myNumber, err := strconv.Atoi(c.view.MyNumber())
	if err != nil {
		c.showErrorOnField(err, core.MyNumberField)
		return
	}
	qso.MyNumber = core.QSONumber(myNumber)

	qso.MyXchange = c.view.MyXchange()

	duplicateQso, duplicate := c.isDuplicate(qso.Callsign)
	if duplicate && duplicateQso.MyNumber != qso.MyNumber {
		c.showErrorOnField(fmt.Errorf("%s was worked before in QSO #%s", qso.Callsign, duplicateQso.MyNumber.String()), core.CallsignField)
		return
	}

	c.logbook.Log(qso)
	c.Reset()
}

func (c *controller) showErrorOnField(err error, field core.EntryField) {
	c.activeField = field
	c.view.SetActiveField(c.activeField)
	c.view.ShowMessage(err)
}

func (c *controller) Reset() {
	c.editing = false
	c.editQSO = core.QSO{}

	nextNumber := c.logbook.NextNumber()
	c.activeField = core.CallsignField
	c.view.SetCallsign("")
	if c.selectedMode == core.ModeSSB {
		c.view.SetMyReport("59")
		c.view.SetTheirReport("59")
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
	c.view.SetEditingMarker(false)
	c.view.ClearMessage()
}

func (c *controller) CurrentValues() core.KeyerValues {
	myNumber, _ := strconv.Atoi(c.view.MyNumber())

	values := core.KeyerValues{}
	values.MyReport, _ = parse.RST(c.view.MyReport())
	values.MyNumber = core.QSONumber(myNumber)
	values.MyXchange = c.view.MyXchange()
	values.TheirCall = c.view.Callsign()

	return values
}
