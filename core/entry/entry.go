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

// View represents the visual part of the QSO data entry.
type View interface {
	Callsign() string
	SetCallsign(string)
	TheirReport() string
	SetTheirReport(string)
	TheirNumber() string
	SetTheirNumber(string)
	TheirXchange() string
	SetTheirXchange(string)
	Band() string
	SetBand(text string)
	Mode() string
	SetMode(text string)
	MyReport() string
	SetMyReport(string)
	MyNumber() string
	SetMyNumber(string)
	MyXchange() string
	SetMyXchange(string)

	EnableExchangeFields(bool, bool)
	SetActiveField(core.EntryField)
	SetDuplicateMarker(bool)
	SetEditingMarker(bool)
	ShowMessage(...interface{})
	ClearMessage()
}

var instance = 0

// Logbook functionality used for QSO entry.
type Logbook interface {
	NextNumber() core.QSONumber
	LastBand() core.Band
	LastMode() core.Mode
	Log(core.QSO)
	FindAll(callsign.Callsign, core.Band, core.Mode) []core.QSO
}

// Keyer functionality used for QSO entry.
type Keyer interface {
	SendQuestion(q string)
}

// Callinfo functionality used for QSO entry.
type Callinfo interface {
	ShowCallsign(string)
}

// NewController returns a new entry controller.
func NewController(clock core.Clock, logbook Logbook, enterTheirNumber, enterTheirXchange, allowMultiBand, allowMultiMode bool) *Controller {
	instance++
	return &Controller{
		instance:          instance,
		clock:             clock,
		view:              &nullView{},
		logbook:           logbook,
		enterTheirNumber:  enterTheirNumber,
		enterTheirXchange: enterTheirXchange,
		allowMultiBand:    allowMultiBand,
		allowMultiMode:    allowMultiMode,
		selectedBand:      logbook.LastBand(),
		selectedMode:      logbook.LastMode(),
	}
}

type Controller struct {
	instance int
	clock    core.Clock
	view     View
	logbook  Logbook
	keyer    Keyer
	callinfo Callinfo

	enterTheirNumber  bool
	enterTheirXchange bool
	allowMultiBand    bool
	allowMultiMode    bool

	activeField  core.EntryField
	selectedBand core.Band
	selectedMode core.Mode
	editing      bool
	editQSO      core.QSO
}

func (c *Controller) SetView(view View) {
	if view == nil {
		c.view = &nullView{}
		return
	}
	c.view = view
	c.view.SetBand(c.selectedBand.String())
	c.view.SetMode(c.selectedMode.String())
	c.view.EnableExchangeFields(c.enterTheirNumber, c.enterTheirXchange)
	c.Reset()
}

func (c *Controller) SetKeyer(keyer Keyer) {
	c.keyer = keyer
}

func (c *Controller) SetCallinfo(callinfo Callinfo) {
	c.callinfo = callinfo
}

func (c *Controller) GotoNextField() core.EntryField {
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

func (c *Controller) leaveCallsignField() {
	callsign, err := callsign.Parse(c.view.Callsign())
	if err != nil {
		fmt.Println(err)
		return
	}

	qso, found := c.IsDuplicate(callsign)
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

func (c *Controller) IsDuplicate(callsign callsign.Callsign) (core.QSO, bool) {
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

func (c *Controller) GetActiveField() core.EntryField {
	return c.activeField
}

func (c *Controller) SetActiveField(field core.EntryField) {
	c.activeField = field
}

func (c *Controller) BandSelected(s string) {
	if band, err := parse.Band(s); err == nil {
		log.Printf("Band selected: %v", band)
		c.selectedBand = band
		c.EnterCallsign(c.view.Callsign())
	}
}

func (c *Controller) ModeSelected(s string) {
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

func (c *Controller) SendQuestion() {
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

func (c *Controller) EnterCallsign(s string) {
	if c.callinfo != nil {
		c.callinfo.ShowCallsign(s)
	}

	callsign, err := callsign.Parse(s)
	if err != nil {
		return
	}

	qso, found := c.IsDuplicate(callsign)
	if !found {
		c.view.ClearMessage()
		return
	}

	c.view.ShowMessage(fmt.Sprintf("%s was worked before in QSO #%s", qso.Callsign, qso.MyNumber.String()))
}

func (c *Controller) QSOSelected(qso core.QSO) {
	log.Printf("%d: QSO selected: %v", c.instance, qso)
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

func (c *Controller) Log() {
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
			return
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

	duplicateQso, duplicate := c.IsDuplicate(qso.Callsign)
	if duplicate && duplicateQso.MyNumber != qso.MyNumber {
		c.showErrorOnField(fmt.Errorf("%s was worked before in QSO #%s", qso.Callsign, duplicateQso.MyNumber.String()), core.CallsignField)
		return
	}

	c.logbook.Log(qso)
	c.Reset()
}

func (c *Controller) showErrorOnField(err error, field core.EntryField) {
	c.activeField = field
	c.view.SetActiveField(c.activeField)
	c.view.ShowMessage(err)
}

func (c *Controller) Reset() {
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

func (c *Controller) CurrentValues() core.KeyerValues {
	myNumber, _ := strconv.Atoi(c.view.MyNumber())

	values := core.KeyerValues{}
	values.MyReport, _ = parse.RST(c.view.MyReport())
	values.MyNumber = core.QSONumber(myNumber)
	values.MyXchange = c.view.MyXchange()
	values.TheirCall = c.view.Callsign()

	return values
}

type nullView struct{}

func (n *nullView) Callsign() string                { return "" }
func (n *nullView) SetCallsign(string)              {}
func (n *nullView) TheirReport() string             { return "" }
func (n *nullView) SetTheirReport(string)           {}
func (n *nullView) TheirNumber() string             { return "" }
func (n *nullView) SetTheirNumber(string)           {}
func (n *nullView) TheirXchange() string            { return "" }
func (n *nullView) SetTheirXchange(string)          {}
func (n *nullView) Band() string                    { return "" }
func (n *nullView) SetBand(text string)             {}
func (n *nullView) Mode() string                    { return "" }
func (n *nullView) SetMode(text string)             {}
func (n *nullView) MyReport() string                { return "" }
func (n *nullView) SetMyReport(string)              {}
func (n *nullView) MyNumber() string                { return "" }
func (n *nullView) SetMyNumber(string)              {}
func (n *nullView) MyXchange() string               { return "" }
func (n *nullView) SetMyXchange(string)             {}
func (n *nullView) EnableExchangeFields(bool, bool) {}
func (n *nullView) SetActiveField(core.EntryField)  {}
func (n *nullView) SetDuplicateMarker(bool)         {}
func (n *nullView) SetEditingMarker(bool)           {}
func (n *nullView) ShowMessage(...interface{})      {}
func (n *nullView) ClearMessage()                   {}
