package entry

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/parse"
	"github.com/ftl/hellocontest/core/ticker"
)

// View represents the visual part of the QSO data entry.
type View interface {
	SetUTC(string)
	SetMyCall(string)
	SetFrequency(core.Frequency)
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
	LastXchange() string
	Log(core.QSO)
}

// QSOList functionality used for QSO entry.
type QSOList interface {
	Find(callsign.Callsign, core.Band, core.Mode) []core.QSO
	FindDuplicateQSOs(callsign.Callsign, core.Band, core.Mode) []core.QSO
	SelectQSO(core.QSO)
	SelectLastQSO()
}

// Keyer functionality used for QSO entry.
type Keyer interface {
	SendQuestion(q string)
	Stop()
}

// Callinfo functionality used for QSO entry.
type Callinfo interface {
	ShowInfo(call string, band core.Band, mode core.Mode, xchange string)
	PredictedXchange() string
}

// VFO functionality used for QSO entry.
type VFO interface {
	Active() bool
	SetFrequency(core.Frequency)
	SetBand(core.Band)
	SetMode(core.Mode)
}

// NewController returns a new entry controller.
func NewController(settings core.Settings, clock core.Clock, qsoList QSOList, asyncRunner core.AsyncRunner) *Controller {
	result := &Controller{
		clock:       clock,
		view:        new(nullView),
		logbook:     new(nullLogbook),
		callinfo:    new(nullCallinfo),
		vfo:         new(nullVFO),
		asyncRunner: asyncRunner,
		qsoList:     qsoList,

		stationCallsign:     settings.Station().Callsign.String(),
		enableTheirNumber:   settings.Contest().EnterTheirNumber,
		enableTheirXchange:  settings.Contest().EnterTheirXchange,
		requireTheirXchange: settings.Contest().RequireTheirXchange,
	}
	result.refreshTicker = ticker.New(result.refreshUTC)
	return result
}

type Controller struct {
	clock    core.Clock
	view     View
	logbook  Logbook
	qsoList  QSOList
	keyer    Keyer
	callinfo Callinfo
	vfo      VFO

	asyncRunner   core.AsyncRunner
	refreshTicker *ticker.Ticker

	stationCallsign     string
	enableTheirNumber   bool
	enableTheirXchange  bool
	requireTheirXchange bool

	input              input
	activeField        core.EntryField
	errorField         core.EntryField
	selectedFrequency  core.Frequency
	selectedBand       core.Band
	selectedMode       core.Mode
	editing            bool
	editQSO            core.QSO
	ignoreQSOSelection bool
}

func (c *Controller) SetView(view View) {
	if view == nil {
		c.view = &nullView{}
		return
	}
	c.view = view
	c.Clear()
	c.refreshUTC()
	c.view.EnableExchangeFields(c.enableTheirNumber, c.enableTheirXchange)
}

func (c *Controller) SetLogbook(logbook Logbook) {
	c.logbook = logbook
	if c.selectedBand == core.NoBand || !c.vfo.Active() {
		lastBand := c.logbook.LastBand()
		if lastBand != core.NoBand {
			c.selectedBand = lastBand
			c.input.band = lastBand.String()
		} else {
			c.selectedBand = core.Band160m
			c.input.band = c.selectedBand.String()
		}
	}
	if c.selectedMode == core.NoMode || !c.vfo.Active() {
		lastMode := c.logbook.LastMode()
		if lastMode != core.NoMode {
			c.selectedMode = lastMode
			c.input.mode = lastMode.String()
		} else {
			c.selectedMode = core.ModeCW
			c.input.mode = c.selectedMode.String()
		}
	}
	c.input.myXchange = c.logbook.LastXchange()

	c.showInput()
}

func (c *Controller) SetKeyer(keyer Keyer) {
	c.keyer = keyer
}

func (c *Controller) SetCallinfo(callinfo Callinfo) {
	c.callinfo = callinfo
}

func (c *Controller) SetVFO(vfo VFO) {
	if vfo == nil {
		c.vfo = new(nullVFO)
	}
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
	if c.enableTheirNumber && c.enableTheirXchange {
		transitions[core.TheirReportField] = core.TheirNumberField
		transitions[core.TheirNumberField] = core.TheirXchangeField
	} else if !c.enableTheirNumber && c.enableTheirXchange {
		transitions[core.TheirReportField] = core.TheirXchangeField
		transitions[core.TheirNumberField] = core.CallsignField
	} else if c.enableTheirNumber && !c.enableTheirXchange {
		transitions[core.TheirReportField] = core.TheirNumberField
		transitions[core.TheirNumberField] = core.CallsignField
	} else if !c.enableTheirNumber && !c.enableTheirXchange {
		transitions[core.TheirReportField] = core.CallsignField
		transitions[core.TheirNumberField] = core.CallsignField
	}

	nextField := transitions[c.activeField]
	if nextField == "" {
		nextField = core.CallsignField
	}

	c.activeField = nextField
	c.view.SetActiveField(c.activeField)
	return c.activeField
}

func (c *Controller) leaveCallsignField() {
	callsign, err := callsign.Parse(c.input.callsign)
	if err != nil {
		fmt.Println(err)
		return
	}
	if c.enableTheirXchange && c.input.theirXchange == "" {
		c.setTheirXchangePrediction(c.callinfo.PredictedXchange())
	}

	_, found := c.isDuplicate(callsign)
	if !found {
		c.view.SetDuplicateMarker(false)
		return
	}
	if c.editing {
		c.view.SetDuplicateMarker(c.editQSO.Callsign != callsign)
		return
	}

	c.view.SetDuplicateMarker(true)
}

func (c *Controller) StartAutoRefresh() {
	c.refreshTicker.Start()
}

func (c *Controller) refreshUTC() {
	c.asyncRunner(func() {
		if c.view == nil {
			return
		}

		utc := c.clock.Now().UTC()
		c.view.SetUTC(utc.Format("15:04"))
	})
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

	c.selectedFrequency = qso.Frequency
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

func (c *Controller) setTheirXchangePrediction(predictedXchange string) {
	c.input.theirXchange = predictedXchange
	c.view.SetTheirXchange(c.input.theirXchange)
}

func (c *Controller) selectQSO(qso core.QSO) {
	c.ignoreQSOSelection = true
	c.qsoList.SelectQSO(qso)
	c.ignoreQSOSelection = false
}

func (c *Controller) isDuplicate(callsign callsign.Callsign) (core.QSO, bool) {
	qsos := c.qsoList.FindDuplicateQSOs(callsign, c.selectedBand, c.selectedMode)
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
		c.enterTheirXchange(text)
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

func (c *Controller) frequencySelected(frequency core.Frequency) {
	log.Printf("Frequency selected: %s", frequency)
	c.selectedFrequency = frequency
	c.vfo.SetFrequency(frequency)
	c.input.callsign = ""
	c.enterCallsign(c.input.callsign)
	c.view.SetCallsign(c.input.callsign)
	c.view.SetFrequency(frequency)
}

func (c *Controller) SetFrequency(frequency core.Frequency) {
	if c.selectedFrequency == frequency {
		return
	}
	c.selectedFrequency = frequency
	c.view.SetFrequency(c.selectedFrequency)
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
		c.callinfo.ShowInfo(c.input.callsign, c.selectedBand, c.selectedMode, c.input.theirXchange)
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

	c.showErrorOnField(fmt.Errorf("%s was worked before in QSO #%s", qso.Callsign, qso.MyNumber.String()), core.CallsignField)
}

func (c *Controller) enterTheirXchange(s string) {
	if c.callinfo == nil {
		return
	}
	c.callinfo.ShowInfo(c.input.callsign, c.selectedBand, c.selectedMode, c.input.theirXchange)
	c.clearErrorOnField(core.TheirXchangeField)
}

func (c *Controller) QSOSelected(qso core.QSO) {
	if c.ignoreQSOSelection {
		return
	}

	log.Printf("QSO selected: %v", qso)
	c.editing = true
	c.editQSO = qso

	c.showQSO(qso)
	c.view.SetActiveField(core.CallsignField)
	c.view.SetEditingMarker(true)
	c.callinfo.ShowInfo(qso.Callsign.String(), qso.Band, qso.Mode, qso.TheirXchange)
}

func (c *Controller) Log() {
	if f, ok := parseKilohertz(c.input.callsign); ok && c.activeField == core.CallsignField {
		c.frequencySelected(f)
		return
	}

	var err error
	qso := core.QSO{}
	if c.editing {
		qso.Time = c.editQSO.Time
	} else {
		qso.Time = c.clock.Now()
	}

	qso.Callsign, err = callsign.Parse(c.input.callsign)
	if err != nil {
		c.showErrorOnField(err, core.CallsignField)
		return
	}

	qso.Frequency = c.selectedFrequency

	qso.Band, err = parse.Band(c.input.band)
	if err != nil {
		c.showErrorOnField(err, core.BandField)
		return
	}

	qso.Mode, err = parse.Mode(c.input.mode)
	if err != nil {
		c.showErrorOnField(err, core.ModeField)
		return
	}

	qso.TheirReport, err = parse.RST(c.input.theirReport)
	if err != nil {
		c.showErrorOnField(err, core.TheirReportField)
		return
	}

	if c.enableTheirNumber {
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

	if c.enableTheirXchange {
		qso.TheirXchange = c.input.theirXchange
		if qso.TheirXchange == "" && c.requireTheirXchange {
			predictedXchange := c.callinfo.PredictedXchange()
			if predictedXchange != "" {
				c.setTheirXchangePrediction(predictedXchange)
				c.showErrorOnField(fmt.Errorf("check their exhange"), core.TheirXchangeField)
				return
			}
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

	c.logbook.Log(qso)
	c.Clear()
}

func parseKilohertz(s string) (core.Frequency, bool) {
	kHz, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return core.Frequency(kHz * 1000), true
}

func (c *Controller) showErrorOnField(err error, field core.EntryField) {
	c.activeField = field
	c.errorField = field
	c.view.SetActiveField(c.activeField)
	c.view.ShowMessage(err)
}

func (c *Controller) clearErrorOnField(field core.EntryField) {
	if c.errorField != field {
		return
	}
	c.view.ClearMessage()
}

func (c *Controller) Clear() {
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
	c.view.SetMyCall(c.stationCallsign)
	c.view.SetFrequency(c.selectedFrequency)
	c.view.SetActiveField(c.activeField)
	c.view.SetDuplicateMarker(false)
	c.view.SetEditingMarker(false)
	c.view.ClearMessage()
	c.selectLastQSO()
	if c.callinfo != nil {
		c.callinfo.ShowInfo("", core.NoBand, core.NoMode, "")
	}
}

func (c *Controller) Activate() {
	c.view.SetActiveField(c.activeField)
}

func (c *Controller) EditLastQSO() {
	c.activeField = core.CallsignField
	c.qsoList.SelectLastQSO()
}

func (c *Controller) StopTX() {
	c.keyer.Stop()
}

func (c *Controller) selectLastQSO() {
	c.ignoreQSOSelection = true
	c.qsoList.SelectLastQSO()
	c.ignoreQSOSelection = false
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

func (c *Controller) StationChanged(station core.Station) {
	c.stationCallsign = station.Callsign.String()
	c.view.SetMyCall(c.stationCallsign)
}

func (c *Controller) ContestChanged(contest core.Contest) {
	c.enableTheirNumber = contest.EnterTheirNumber
	c.enableTheirXchange = contest.EnterTheirXchange
	c.requireTheirXchange = contest.RequireTheirXchange
	c.view.EnableExchangeFields(c.enableTheirNumber, c.enableTheirXchange)
}

type nullView struct{}

func (n *nullView) SetUTC(string)                   {}
func (n *nullView) SetMyCall(string)                {}
func (n *nullView) SetFrequency(core.Frequency)     {}
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

func (n *nullVFO) Active() bool                { return false }
func (n *nullVFO) SetFrequency(core.Frequency) {}
func (n *nullVFO) SetBand(core.Band)           {}
func (n *nullVFO) SetMode(core.Mode)           {}

type nullLogbook struct{}

func (n *nullLogbook) NextNumber() core.QSONumber { return 0 }
func (n *nullLogbook) LastBand() core.Band        { return core.NoBand }
func (n *nullLogbook) LastMode() core.Mode        { return core.NoMode }
func (n *nullLogbook) LastXchange() string        { return "" }
func (n *nullLogbook) Log(core.QSO)               {}

type nullCallinfo struct{}

func (n *nullCallinfo) ShowInfo(string, core.Band, core.Mode, string) {}
func (n *nullCallinfo) PredictedXchange() string                      { return "" }
