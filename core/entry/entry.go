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
	SetCallsign(string)
	SetTheirReport(string)
	SetTheirNumber(string)
	SetTheirXchange(string)
	SetBand(text string)
	SetMode(text string)
	SetMyReport(string)
	SetMyNumber(string)
	SetMyXchange(string)

	EnableExchangeFields(bool, bool)
	SetActiveField(core.EntryField)
	SetDuplicateMarker(bool)
	SetEditingMarker(bool)
	ShowMessage(...interface{})
	ClearMessage()
}

type input struct {
	callsign     string
	theirReport  string
	theirNumber  string
	theirXchange string
	myReport     string
	myNumber     string
	myXchange    string
	band         string
	mode         string
}

// Logbook functionality used for QSO entry.
type Logbook interface {
	NextNumber() core.QSONumber
	LastBand() core.Band
	LastMode() core.Mode
	Log(core.QSO)
	FindAll(callsign.Callsign, core.Band, core.Mode) []core.QSO
	SelectQSO(core.QSO)
	SelectLastQSO()
}

// Keyer functionality used for QSO entry.
type Keyer interface {
	SendQuestion(q string)
}

// Callinfo functionality used for QSO entry.
type Callinfo interface {
	ShowCallsign(string)
}

// VFO functionality used for QSO entry.
type VFO interface {
	SetBand(core.Band)
	SetMode(core.Mode)
}

// NewController returns a new entry controller.
func NewController(clock core.Clock, logbook Logbook, enterTheirNumber, enterTheirXchange, allowMultiBand, allowMultiMode bool) *Controller {
	return &Controller{
		clock:             clock,
		view:              &nullView{},
		vfo:               &nullVFO{},
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
	clock    core.Clock
	view     View
	logbook  Logbook
	keyer    Keyer
	callinfo Callinfo
	vfo      VFO

	enterTheirNumber  bool
	enterTheirXchange bool
	allowMultiBand    bool
	allowMultiMode    bool

	input        input
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
	c.Reset()
	c.view.EnableExchangeFields(c.enterTheirNumber, c.enterTheirXchange)
}

func (c *Controller) SetKeyer(keyer Keyer) {
	c.keyer = keyer
}

func (c *Controller) SetCallinfo(callinfo Callinfo) {
	c.callinfo = callinfo
}

func (c *Controller) SetVFO(vfo VFO) {
	c.vfo = vfo
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
		core.BandField:         core.CallsignField,
		core.ModeField:         core.CallsignField,
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
	callsign, err := callsign.Parse(c.input.callsign)
	if err != nil {
		fmt.Println(err)
		return
	}

	qso, found := c.IsDuplicate(callsign)
	if !found {
		c.view.SetDuplicateMarker(false)
		return
	}
	if c.editing {
		c.view.SetDuplicateMarker(c.editQSO.Callsign != qso.Callsign)
		return
	}

	c.showQSO(qso)
	c.view.SetDuplicateMarker(true)
	c.logbook.SelectQSO(qso)
}

func (c *Controller) showQSO(qso core.QSO) {
	c.input.callsign = qso.Callsign.String()
	c.input.theirReport = qso.TheirReport.String()
	c.input.theirNumber = qso.TheirNumber.String()
	c.input.theirXchange = qso.TheirXchange
	c.input.myReport = qso.MyReport.String()
	c.input.myNumber = qso.MyNumber.String()
	c.input.myXchange = qso.MyXchange
	c.input.band = qso.Band.String()
	c.input.mode = qso.Mode.String()

	c.selectedBand = qso.Band
	c.selectedMode = qso.Mode

	c.showInput()
}

func (c *Controller) showInput() {
	c.view.SetCallsign(c.input.callsign)
	c.view.SetTheirReport(c.input.theirReport)
	c.view.SetTheirNumber(c.input.theirNumber)
	c.view.SetTheirXchange(c.input.theirXchange)
	c.view.SetMyReport(c.input.myReport)
	c.view.SetMyNumber(c.input.myNumber)
	c.view.SetMyXchange(c.input.myXchange)
	c.view.SetBand(c.input.band)
	c.view.SetMode(c.input.mode)
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

func (c *Controller) SetActiveField(field core.EntryField) {
	c.activeField = field
}

func (c *Controller) Enter(text string) {
	switch c.activeField {
	case core.CallsignField:
		c.input.callsign = text
		c.enterCallsign(text)
	case core.TheirReportField:
		c.input.theirReport = text
	case core.TheirNumberField:
		c.input.theirNumber = text
	case core.TheirXchangeField:
		c.input.theirXchange = text
	case core.MyReportField:
		c.input.myReport = text
	case core.MyNumberField:
		c.input.myNumber = text
	case core.MyXchangeField:
		c.input.myXchange = text
	case core.BandField:
		c.input.band = text
		c.bandSelected(text)
	case core.ModeField:
		c.input.mode = text
		c.modeSelected(text)
	}
}

func (c *Controller) bandSelected(s string) {
	if band, err := parse.Band(s); err == nil {
		log.Printf("Band selected: %v", band)
		c.selectedBand = band
		c.vfo.SetBand(band)
		c.enterCallsign(c.input.callsign)
	}
}

func (c *Controller) SetBand(band core.Band) {
	if c.selectedBand == band {
		return
	}
	c.selectedBand = band
	c.input.band = c.selectedBand.String()
	c.view.SetBand(c.input.band)
}

func (c *Controller) modeSelected(s string) {
	if mode, err := parse.Mode(s); err == nil {
		log.Printf("Mode selected: %v", mode)
		c.selectedMode = mode
		c.vfo.SetMode(mode)

		if c.selectedMode == core.ModeSSB {
			c.input.theirReport = "59"
			c.input.myReport = "59"
		} else {
			c.input.myReport = "599"
			c.input.theirReport = "599"
		}

		c.enterCallsign(c.input.callsign)
	}
}

func (c *Controller) SetMode(mode core.Mode) {
	if c.selectedMode == mode {
		return
	}
	c.selectedMode = mode
	c.input.mode = c.selectedMode.String()
	c.view.SetMode(c.input.mode)
}

func (c *Controller) SendQuestion() {
	if c.keyer == nil {
		return
	}

	switch c.activeField {
	case core.TheirReportField, core.TheirNumberField, core.TheirXchangeField:
		c.keyer.SendQuestion("nr")
	default:
		c.keyer.SendQuestion(c.input.callsign)
	}
}

func (c *Controller) enterCallsign(s string) {
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
	log.Printf("QSO selected: %v", qso)
	c.editing = true
	c.editQSO = qso

	c.showQSO(qso)
	c.view.SetActiveField(core.CallsignField)
	c.view.SetEditingMarker(true)
}

func (c *Controller) Log() {
	var err error
	qso := core.QSO{}
	qso.Callsign, err = callsign.Parse(c.input.callsign)
	if err != nil {
		c.showErrorOnField(err, core.CallsignField)
		return
	}
	if c.editing {
		qso.Time = c.editQSO.Time
	} else {
		qso.Time = c.clock.Now()
	}

	qso.Band, err = parse.Band(c.input.band)
	if err != nil {
		c.view.ShowMessage(err)
		return
	}

	qso.Mode, err = parse.Mode(c.input.mode)
	if err != nil {
		c.view.ShowMessage(err)
		return
	}

	qso.TheirReport, err = parse.RST(c.input.theirReport)
	if err != nil {
		c.showErrorOnField(err, core.TheirReportField)
		return
	}

	if c.enterTheirNumber {
		value := c.input.theirNumber
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
		qso.TheirXchange = c.input.theirXchange
		if qso.TheirXchange == "" {
			c.showErrorOnField(errors.New("their exchange is missing"), core.TheirXchangeField)
			return
		}
	}

	qso.MyReport, err = parse.RST(c.input.myReport)
	if err != nil {
		c.showErrorOnField(err, core.MyReportField)
		return
	}

	myNumber, err := strconv.Atoi(c.input.myNumber)
	if err != nil {
		c.showErrorOnField(err, core.MyNumberField)
		return
	}
	qso.MyNumber = core.QSONumber(myNumber)

	qso.MyXchange = c.input.myXchange

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
	c.input.callsign = ""
	if c.selectedMode == core.ModeSSB {
		c.input.myReport = "59"
		c.input.theirReport = "59"
	} else {
		c.input.myReport = "599"
		c.input.theirReport = "599"
	}
	c.input.theirNumber = ""
	c.input.theirXchange = ""
	if c.selectedBand != core.NoBand {
		c.input.band = c.selectedBand.String()
	}
	if c.selectedMode != core.NoMode {
		c.input.mode = c.selectedMode.String()
	}
	c.input.myNumber = nextNumber.String()

	c.showInput()
	c.view.SetActiveField(c.activeField)
	c.view.SetDuplicateMarker(false)
	c.view.SetEditingMarker(false)
	c.view.ClearMessage()
	c.logbook.SelectLastQSO()
}

func (c *Controller) CurrentValues() core.KeyerValues {
	myNumber, _ := strconv.Atoi(c.input.myNumber)

	values := core.KeyerValues{}
	values.MyReport, _ = parse.RST(c.input.myReport)
	values.MyNumber = core.QSONumber(myNumber)
	values.MyXchange = c.input.myXchange
	values.TheirCall = c.input.callsign

	return values
}

type nullView struct{}

func (n *nullView) SetCallsign(string)              {}
func (n *nullView) SetTheirReport(string)           {}
func (n *nullView) SetTheirNumber(string)           {}
func (n *nullView) SetTheirXchange(string)          {}
func (n *nullView) SetBand(text string)             {}
func (n *nullView) SetMode(text string)             {}
func (n *nullView) SetMyReport(string)              {}
func (n *nullView) SetMyNumber(string)              {}
func (n *nullView) SetMyXchange(string)             {}
func (n *nullView) EnableExchangeFields(bool, bool) {}
func (n *nullView) SetActiveField(core.EntryField)  {}
func (n *nullView) SetDuplicateMarker(bool)         {}
func (n *nullView) SetEditingMarker(bool)           {}
func (n *nullView) ShowMessage(...interface{})      {}
func (n *nullView) ClearMessage()                   {}

type nullVFO struct{}

func (n *nullVFO) SetBand(core.Band) {}
func (n *nullVFO) SetMode(core.Mode) {}
func (n *nullVFO) Refresh()          {}
