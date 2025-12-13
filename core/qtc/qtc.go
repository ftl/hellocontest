package qtc

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

// text prompts to communicate with the opposite station
const (
	OfferQTCText          = "qtc"
	SendHeaderTemplate    = "qtc %s"
	CompleteQTCSeriesText = "tu"

	RequestQTCText    = "qtc?"
	QRVText           = "qrv"
	ConfirmText       = "r"
	RequestRepeatText = "agn"
)

type Logbook interface {
	NextSeriesNumber() int
	LastCallsign() callsign.Callsign
	LogQTC(core.QTC)
}

type QTCList interface {
	PrepareFor(callsign.Callsign, int) []core.QTC
}

type EntryController interface {
	CurrentQSOState() (callsign.Callsign, core.QSODataState)
	Log()
}

type Keyer interface {
	SendText(string, ...any)
	Repeat()
	Stop()
	Cut(string) string
}

type InfoDialogs interface {
	ShowInfo(format string, args ...any)
	ShowQuestion(format string, args ...any) bool
	ShowError(format string, args ...any)
}

type View interface {
	QuestionQTCCount(max int) (int, bool)
	ShowError(string)
	ClearError()
	Show(core.QTCMode, core.QTCSeries)
	UpdateQTC(int, core.QTC)
	Close()
	ClearDataInputs()
	SetActivePhase(core.QTCWorkflowPhase)
	SetActiveField(core.QTCField)
	SetActiveQTC(int)
}

type Controller struct {
	clock           core.Clock
	logbook         Logbook
	qtcList         QTCList
	entryController EntryController
	keyer           Keyer

	infoDialogs InfoDialogs
	view        View

	activePhase core.QTCWorkflowPhase
	activeField core.QTCField

	currentMode   core.QTCMode
	currentSeries core.QTCSeries
	currentQTC    int
	currentInput  map[core.QTCField]string

	vfoFrequency core.Frequency
	vfoBand      core.Band
	vfoMode      core.Mode
}

func NewController(clock core.Clock, infoDialogs InfoDialogs, qtcList QTCList, entryController EntryController, keyer Keyer) *Controller {
	return &Controller{
		clock:           clock,
		logbook:         new(nullLogbook),
		qtcList:         qtcList,
		entryController: entryController,
		keyer:           keyer,
		infoDialogs:     infoDialogs,
		view:            new(nullView),
		currentInput:    make(map[core.QTCField]string),
	}
}

func (c *Controller) SetLogbook(logbook Logbook) {
	if logbook == nil {
		c.logbook = new(nullLogbook)
		return
	}
	c.logbook = logbook
}

func (c *Controller) SetView(view View) {
	if view == nil {
		c.view = new(nullView)
		return
	}
	c.view = view
}

func (c *Controller) questionInvalidQSOData(call callsign.Callsign) bool {
	return c.infoDialogs.ShowQuestion("The entered callsign is valid, but the QSO data is invalid. Proceed with the entered callsign of %s?", call.String())
}

func (c *Controller) questionConfirmAbort() bool {
	return c.infoDialogs.ShowQuestion("The QTC exchange is incomplete. Do you want to cancel it?")
}

func (c *Controller) showErrorDialog(format string, args ...any) {
	c.infoDialogs.ShowError(format, args...)
}

func (c *Controller) showErrorMessage(format string, args ...any) {
	c.view.ShowError(fmt.Sprintf(format, args...))
}

func (c *Controller) clearErrorMessage() {
	c.view.ClearError()
}

func (c *Controller) VFOFrequencyChanged(frequency core.Frequency) {
	c.vfoFrequency = frequency
}

func (c *Controller) VFOBandChanged(band core.Band) {
	c.vfoBand = band
}

func (c *Controller) VFOModeChanged(mode core.Mode) {
	c.vfoMode = mode
}

// *****************************************************
// NEW WORKFLOW METHODS
// *****************************************************

// func OfferQTC stays in the ProvideQTC section for now
// func RequestQTC stays in the ReceiveQTC section for now

func (c *Controller) StartAction() {
	if c.activePhase != core.QTCStart {
		return
	}

	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		c.SendQTCOffer()
	} else {
		c.SendQTCRequest()
	}
}

func (c *Controller) HeaderAction() {
	if c.activePhase != core.QTCExchangeHeader {
		return
	}

	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		c.SendHeader()
	} else {
		c.RequestRepeat()
	}
}

func (c *Controller) DataAction() {
	if c.activePhase != core.QTCExchangeData {
		return
	}

	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		c.SendCurrentQTC()
	} else {
		c.RequestRepeat()
	}
}

func (c *Controller) ConfirmStart() {
	if c.activePhase != core.QTCStart {
		return
	}
	c.SetActivePhase(core.QTCExchangeHeader)
	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ReceiveQTC {
		c.SetActiveField(core.QTCHeaderField)
		c.SendQRV()
	}
}

func (c *Controller) ConfirmHeader() {
	if c.activePhase != core.QTCExchangeHeader {
		return
	}
	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		c.SetActivePhase(core.QTCExchangeData)
	} else {
		// parse and validate the header
		header, err := c.currentHeader()
		if err != nil {
			c.showErrorMessage("%v", err)
			c.SetActiveField(core.QTCHeaderField)
			return
		}
		c.currentSeries.Header = header
		c.clearErrorMessage()

		// progress in the workflow
		c.currentQTC = 0
		c.SetActivePhase(core.QTCExchangeData)
		c.SetActiveField(core.QTCTimestampField)
		c.SendConfirm()
	}
}

func (c *Controller) ConfirmData() {
	if c.activePhase != core.QTCExchangeData {
		return
	}

	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		if c.currentSeries.IsValidQTCIndex(c.currentQTC) {
			qtc := c.currentSeries.QTCs[c.currentQTC]
			qtc.Confirmed = true
			c.currentSeries.QTCs[c.currentQTC] = qtc
			c.view.UpdateQTC(c.currentQTC, qtc)
		}
		if c.currentSeries.IsValidQTCIndex(c.currentQTC + 1) {
			c.SetActiveQTC(c.currentQTC + 1)
		} else {
			c.SetActivePhase(core.QTCFinish)
		}
	} else {
		// parse the entered QTC Data
		qtc := core.QTC{
			Kind:      core.ReceivedQTC,
			Timestamp: c.clock.Now(),
		}

		qtcTime, err := c.currentQTCTimestamp()
		if err != nil {
			c.showErrorMessage("%v", err)
			c.SetActiveField(core.QTCTimestampField)
			return
		}
		qtc.QTCTime = qtcTime

		qtcCallsign, err := c.currentQTCCallsign()
		if err != nil {
			c.showErrorMessage("%v", err)
			c.SetActiveField(core.QTCCallsignField)
			return
		}
		qtc.QTCCallsign = qtcCallsign

		qtcNumber, err := c.currentQTCNumber()
		if err != nil {
			c.showErrorMessage("%v", err)
			c.SetActiveField(core.QTCExchangeField)
			return
		}
		qtc.QTCNumber = qtcNumber
		c.clearErrorMessage()

		// log the entered QTC data in the current series
		c.currentSeries.SetData(c.currentQTC, qtc)

		// clear the current input
		delete(c.currentInput, core.QTCTimestampField)
		delete(c.currentInput, core.QTCCallsignField)
		delete(c.currentInput, core.QTCExchangeField)
		// TODO: clear the input fields also in the view

		// progress in the workflow
		c.currentQTC += 1
		c.SetActiveField(core.QTCTimestampField)
		c.SendConfirm()
	}
}

func (c *Controller) currentHeader() (core.QTCHeader, error) {
	headerStr, ok := c.currentInput[core.QTCHeaderField]
	if !ok || (headerStr == "") {
		return core.QTCHeader{}, fmt.Errorf("the header field is empty")
	}
	return core.ParseQTCHeader(headerStr)
}

func (c *Controller) currentQTCTimestamp() (core.QTCTime, error) {
	qtcTimeStr, ok := c.currentInput[core.QTCTimestampField]
	if !ok || (qtcTimeStr == "") {
		return core.ZeroQTCTime, fmt.Errorf("the QTC timestamp field is empty")
	}
	return core.ParseQTCTime(qtcTimeStr, c.currentReferenceTime())
}

func (c *Controller) currentReferenceTime() core.QTCTime {
	previousQTC := c.currentQTC - 1
	if previousQTC >= len(c.currentSeries.QTCs) {
		previousQTC = len(c.currentSeries.QTCs) - 1
	}
	if previousQTC < 0 {
		return core.ZeroQTCTime
	}
	return c.currentSeries.QTCs[previousQTC].QTCTime
}

func (c *Controller) currentQTCCallsign() (callsign.Callsign, error) {
	qtcCallsignStr, ok := c.currentInput[core.QTCCallsignField]
	if !ok || (qtcCallsignStr == "") {
		return callsign.Callsign{}, fmt.Errorf("the QTC callsign field is empty")
	}
	qtcCallsign, err := callsign.Parse(qtcCallsignStr)
	if err != nil {
		return callsign.Callsign{}, err
	}
	return qtcCallsign, nil
}

func (c *Controller) currentQTCNumber() (core.QSONumber, error) {
	qtcExchangeStr, ok := c.currentInput[core.QTCExchangeField]
	if !ok || (qtcExchangeStr == "") {
		return 0, fmt.Errorf("the QTC exchange field is empty")
	}
	qtcExchange, err := strconv.Atoi(qtcExchangeStr)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid QTC exchange", qtcExchangeStr)
	}
	return core.QSONumber(qtcExchange), nil
}

func (c *Controller) Enter(s string) {
	c.currentInput[c.activeField] = s
}

func (c *Controller) GotoNextField() {
	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		// TODO: implement
	} else {
		nextField := c.activeField
		switch c.activeField {
		case core.QTCHeaderField:
		// ignore, use ConfirmHeader to progress
		case core.QTCTimestampField:
			nextField = core.QTCCallsignField
		case core.QTCCallsignField:
			nextField = core.QTCExchangeField
		case core.QTCExchangeField:
			nextField = core.QTCTimestampField
		}
		c.activeField = nextField
		c.view.SetActiveField(c.activeField)
	}
}

func (c *Controller) SetActiveField(field core.QTCField) {
	// TODO: make sure the active phase matches the field's phase

	// set the active field
	c.activeField = field
	// update the view
	c.view.SetActiveField(field)
}

// func CompleteQTCSeries() stays below for now
// func AbortQTCSeries() stays below for now

// *****************************************************
// OLD WORKFLOW METHODS
// *****************************************************

func (c *Controller) RequestRepeat() {
	c.keyer.SendText(RequestRepeatText)
}

func (c *Controller) SetActivePhase(phase core.QTCWorkflowPhase) {
	c.view.SetActivePhase(phase)
	if c.activePhase == phase {
		return
	}

	// enter the phase
	c.activePhase = phase

	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		switch c.activePhase {
		case core.QTCStart:
			c.SendQTCOffer()
		case core.QTCExchangeHeader:
			c.SendHeader()
		case core.QTCExchangeData:
			c.SetActiveQTC(0)
		}
	} else {
		log.Printf("set active phase to %v", c.activePhase)
	}
}

func (c *Controller) SetActiveQTC(index int) {
	if !c.currentSeries.IsValidQTCIndex(index) {
		return
	}
	if c.activePhase != core.QTCExchangeData {
		c.activePhase = core.QTCExchangeData
		c.view.SetActivePhase(core.QTCExchangeData)
	}

	c.currentQTC = index
	c.view.SetActiveQTC(c.currentQTC)

	c.SendCurrentQTC()
}

func (c *Controller) SendCurrentQTC() {
	if !c.currentSeries.IsValidQTCIndex(c.currentQTC) {
		return
	}
	currentQTC := c.currentSeries.QTCs[c.currentQTC]

	// shorten time if the last QTC qso was in the same hour
	// TODO: make this optional?
	shortenTime := false
	if c.currentQTC > 0 {
		lastQTC := c.currentSeries.QTCs[c.currentQTC-1]
		shortenTime = lastQTC.QTCTime.Hour == currentQTC.QTCTime.Hour
	}

	c.SendQTC(currentQTC, shortenTime)

	// add transmission data and mark the QTC as transmitted
	currentQTC.Timestamp = c.clock.Now()
	currentQTC.Frequency = c.vfoFrequency
	currentQTC.Band = c.vfoBand
	currentQTC.Mode = c.vfoMode
	c.currentSeries.QTCs[c.currentQTC] = currentQTC
}

// Workflow for providing QTCs

func (c *Controller) OfferQTC() {
	// 1. find out their callsign
	theirCall, ok := c.findOutTheirCallsign()
	if !ok {
		return
	}

	// 2. get available QTCs
	qtcs := c.qtcList.PrepareFor(theirCall, core.MaxQTCsPerCall)
	if len(qtcs) == 0 {
		c.showErrorDialog("No QTCs available for %s", theirCall)
		return
	}

	// 3. enter the number of QTCs to send and reduce the qtcs slice accordingly
	qtcCount, ok := c.view.QuestionQTCCount(len(qtcs))
	if !ok {
		return
	}
	qtcCount = min(qtcCount, len(qtcs))
	qtcs = qtcs[:qtcCount]

	// 4. create new QTCSeries
	qtcSeries, err := core.NewQTCSeries(c.logbook.NextSeriesNumber(), qtcs)
	if err != nil {
		c.showErrorDialog("%v", err)
		return
	}
	c.currentMode = core.ProvideQTC
	c.currentSeries = qtcSeries
	c.currentQTC = core.NoQTCIndex

	// 5. enter the first phase: send "qtc"
	c.SetActivePhase(core.QTCStart)

	// 6. show and run the QTC dialog
	c.view.Show(c.currentMode, c.currentSeries)
}

func (c *Controller) findOutTheirCallsign() (callsign.Callsign, bool) {
	theirCall, currentQSOState := c.entryController.CurrentQSOState()
	switch currentQSOState {
	case core.QSODataValid: // a) there is currently a valid QSO in the entry fields that is not yet logged -> log this QSO and take their callsign
		c.entryController.Log()
	case core.QSODataInvalid: // b) there is currently a valid callsign and some QSO data (but not valid) in the entry fields -> show info about invalid QSO data, ask if the callsign should be used -> use the callsign
		if !c.questionInvalidQSOData(theirCall) {
			return callsign.Callsign{}, false
		}
	case core.QSODataEmpty: // c) there is currently a valid callsign in the entry field, but no QSO data at all-> use this callsign
	default:
		panic(fmt.Errorf("unknown QSODataState: %d", currentQSOState))
	}
	if theirCall.BaseCall != "" {
		return theirCall, true
	}

	// d) otherwise -> use the last logged callsign
	theirCall = c.logbook.LastCallsign()

	return theirCall, (theirCall.BaseCall != "")
}

// CompleteQTCSeries completes the current QTC series.
//
// mode == ProvideQTC: stores all QTCs to the log, sends "tu", and closes the QTC window.
// The series can only be completed when all QTCs have been transmitted. Otherwise, an
// error message is presented to the user, the QTC window stays open.
//
// mode == ReceiveQTC: not yet implemented
func (c *Controller) CompleteQTCSeries() {
	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		// check if all QTCs have actually been transmitted
		for i, qtc := range c.currentSeries.QTCs {
			if qtc.Confirmed {
				continue
			}

			c.showErrorDialog("Not all QTCs have been confirmed, the QTC series cannot be completed. Abort the series to close the window or transmit the remaining QTCs.")
			c.SetActiveQTC(i)
			return
		}

		for _, qtc := range c.currentSeries.QTCs {
			c.logbook.LogQTC(qtc)
		}

		c.keyer.SendText(CompleteQTCSeriesText)

		c.view.Close()
	} else {
		// fill common data from the series and log the QTC
		for i, qtc := range c.currentSeries.QTCs {
			qtc.TheirCallsign = c.currentSeries.TheirCallsign
			qtc.Header = c.currentSeries.Header
			qtc.Frequency = c.vfoFrequency
			qtc.Band = c.vfoBand
			qtc.Mode = c.vfoMode
			c.currentSeries.QTCs[i] = qtc
			c.logbook.LogQTC(qtc)
		}

		c.view.Close()
	}
}

// AbortQTCSeries aborts the current QTC series: no QTCs are logged, the QTC window is closed.
// To prevent data loss due to an accidental abort, the user is asked for confirmation first.
func (c *Controller) AbortQTCSeries() {
	if !c.questionConfirmAbort() {
		return
	}

	c.view.Close()
}

// Workflow for receiving QTCs

func (c *Controller) RequestQTC() {
	// find out their callsign
	theirCall, ok := c.findOutTheirCallsign()
	if !ok {
		return
	}

	// TODO: find out how many QTCs may be received from theirCall (is this relevant, maybe we should just take what we get)

	// create a new and empty QTCSeries
	qtcSeries := core.NewReceivingQTCSeries(theirCall)
	c.currentMode = core.ReceiveQTC
	c.currentSeries = qtcSeries
	c.currentQTC = core.NoQTCIndex

	// enter the first phase: send "qtc?"
	c.SetActivePhase(core.QTCStart)

	// show and run the QTC dialog
	c.view.Show(c.currentMode, c.currentSeries)
}

// *****************************************************
// SENDING METHODS - NO WORKFLOW BELOW HERE
// *****************************************************

// SendQTCOffer sends the offer for a QTC exchange.
func (c *Controller) SendQTCOffer() {
	c.keyer.SendText(OfferQTCText)
}

// SendQTCRequest sends the request for a QTC exchange.
func (c *Controller) SendQTCRequest() {
	c.keyer.SendText(RequestQTCText)
}

func (c *Controller) SendQRV() {
	c.keyer.SendText(QRVText)
}

func (c *Controller) SendConfirm() {
	c.keyer.SendText(ConfirmText)
}

// SendHeader sends the header of the current QTC series.
func (c *Controller) SendHeader() {
	c.keyer.SendText(SendHeaderTemplate, c.currentSeries.Header)
}

// SendQTC sends the given QTC.
func (c *Controller) SendQTC(qtc core.QTC, shortenTime bool) {
	time := qtc.QTCTime.String()
	if shortenTime {
		time = qtc.QTCTime.ShortString()
	}
	call := qtc.QTCCallsign.String()
	exchange := c.keyer.Cut(qtc.QTCNumber.String())

	c.keyer.SendText("%s %s %s", time, call, exchange)
}

// nullLogbook

var _ Logbook = &nullLogbook{}

type nullLogbook struct{}

func (*nullLogbook) NextSeriesNumber() int           { return 0 }
func (*nullLogbook) LastCallsign() callsign.Callsign { return callsign.Callsign{} }
func (*nullLogbook) LogQTC(core.QTC)                 {}

// nullView

var _ View = &nullView{}

type nullView struct{}

func (*nullView) QuestionQTCCount(int) (int, bool)     { return 0, false }
func (*nullView) ShowError(string)                     {}
func (*nullView) ClearError()                          {}
func (*nullView) Show(core.QTCMode, core.QTCSeries)    {}
func (*nullView) UpdateQTC(int, core.QTC)              {}
func (*nullView) Close()                               {}
func (*nullView) ClearDataInputs()                     {}
func (*nullView) SetActivePhase(core.QTCWorkflowPhase) {}
func (*nullView) SetActiveField(core.QTCField)         {}
func (*nullView) SetActiveQTC(int)                     {}
