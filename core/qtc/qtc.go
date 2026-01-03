package qtc

import (
	"fmt"
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
	AddQTCSeries(core.QTCSeries)
	PrepareFor(callsign.Callsign, int) []core.QTC
}

type QTCList interface {
	SelectLastQTC()
	SetQTCsEnabled(bool)
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
	ShowFieldError(core.QTCField, string)
	ClearFieldError()
	Show(core.QTCMode, core.QTCSeries)
	UpdateQTC(int, core.QTC)
	Close()
	ClearDataInputs()
	SetActivePhase(core.QTCWorkflowPhase)
	SetActiveField(core.QTCField)
	SetActiveQTC(int)
}

type qtcWorkflow interface {
	StartAction()
	ConfirmStart()
	HeaderAction()
	ConfirmHeader()
	DataAction()
	ConfirmData()
	GotoNextField()
	CompleteQTCSeries()
}

type Controller struct {
	workflow qtcWorkflow
	enabled  bool

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

func NewController(clock core.Clock, infoDialogs InfoDialogs, logbook Logbook, qtcList QTCList, entryController EntryController, keyer Keyer) *Controller {
	return &Controller{
		workflow:        new(nullWorkflow),
		clock:           clock,
		logbook:         logbook,
		qtcList:         qtcList,
		entryController: entryController,
		keyer:           keyer,
		infoDialogs:     infoDialogs,
		view:            new(nullView),
		currentInput:    make(map[core.QTCField]string),
	}
}

func (c *Controller) SetView(view View) {
	if view == nil {
		c.view = new(nullView)
		return
	}
	c.view = view
	c.qtcList.SetQTCsEnabled(c.enabled)
}

func (c *Controller) ContestChanged(contest core.Contest) {
	c.enabled = contest.EnableQTCs
	c.qtcList.SetQTCsEnabled(contest.EnableQTCs)
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

func (c *Controller) showFieldError(field core.QTCField, err error) {
	c.view.ShowFieldError(field, err.Error())
}

func (c *Controller) clearErrorMessage() {
	c.view.ClearFieldError()
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

// OfferQTC initiates the QTC dialog with mode core.ProvideQTC.
func (c *Controller) OfferQTC() {
	// 1. find out their callsign
	theirCall, ok := c.findOutTheirCallsign()
	if !ok {
		return
	}

	// 2. get available QTCs
	qtcs := c.logbook.PrepareFor(theirCall, core.MaxQTCsPerCall)
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
	c.workflow = &provide{Controller: c}
	c.currentMode = core.ProvideQTC
	c.currentSeries = qtcSeries
	c.currentQTC = core.NoQTCIndex

	// 5. enter the first phase: send "qtc"
	c.setActivePhase(core.QTCStart)
	c.sendQTCOffer()

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

// OfferQTC initiates the QTC dialog with mode core.ReceiveQTC.
func (c *Controller) RequestQTC() {
	// find out their callsign
	theirCall, ok := c.findOutTheirCallsign()
	if !ok {
		return
	}

	// TODO: find out how many QTCs may be received from theirCall (is this relevant, maybe we should just take what we get)

	// create a new and empty QTCSeries
	qtcSeries := core.NewReceivingQTCSeries(theirCall)
	c.workflow = &receive{Controller: c}
	c.currentMode = core.ReceiveQTC
	c.currentSeries = qtcSeries
	c.currentQTC = core.NoQTCIndex

	// enter the first phase
	c.setActivePhase(core.QTCStart)

	// show and run the QTC dialog
	c.view.Show(c.currentMode, c.currentSeries)
}

// *****************************************************
// COMMON WORKFLOW METHODS
// *****************************************************

func (c *Controller) StartAction() {
	if c.activePhase != core.QTCStart {
		return
	}

	c.workflow.StartAction()
}

func (c *Controller) ConfirmStart() {
	if c.activePhase != core.QTCStart {
		return
	}
	c.setActivePhase(core.QTCExchangeHeader)

	c.workflow.ConfirmStart()
}

func (c *Controller) HeaderAction() {
	if c.activePhase != core.QTCExchangeHeader {
		return
	}

	c.workflow.HeaderAction()
}

func (c *Controller) ConfirmHeader() {
	if c.activePhase != core.QTCExchangeHeader {
		return
	}

	c.workflow.ConfirmHeader()
}

func (c *Controller) DataAction() {
	if c.activePhase != core.QTCExchangeData {
		return
	}

	c.workflow.DataAction()
}

func (c *Controller) ConfirmData() {
	if c.activePhase != core.QTCExchangeData {
		return
	}

	c.workflow.ConfirmData()
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

func (c *Controller) DoAction() {
	switch c.activePhase {
	case core.QTCStart:
		c.StartAction()
	case core.QTCExchangeHeader:
		c.HeaderAction()
	case core.QTCExchangeData:
		c.DataAction()
	}
}

func (c *Controller) DoConfirm() {
	switch c.activePhase {
	case core.QTCStart:
		c.ConfirmStart()
	case core.QTCExchangeHeader:
		c.ConfirmHeader()
	case core.QTCExchangeData:
		c.ConfirmData()
	case core.QTCFinish:
		c.CompleteQTCSeries()
	}
}

func (c *Controller) Enter(s string) {
	c.currentInput[c.activeField] = s
}

func (c *Controller) GotoNextField() {
	c.workflow.GotoNextField()
}

func (c *Controller) SetActiveField(field core.QTCField) {
	// TODO: make sure the active phase matches the field's phase

	// set the active field
	c.activeField = field
	// update the view
	c.view.SetActiveField(field)
}

// CompleteQTCSeries completes the current QTC series.
//
// mode == ProvideQTC: stores all QTCs to the log, sends "tu", and closes the QTC window.
// The series can only be completed when all QTCs have been transmitted. Otherwise, an
// error message is presented to the user, the QTC window stays open.
//
// mode == ReceiveQTC: stores all received QTCs to the log and closes the QTC window.
func (c *Controller) CompleteQTCSeries() {
	c.workflow.CompleteQTCSeries()
	c.qtcList.SelectLastQTC()
}

// AbortQTCSeries aborts the current QTC series: no QTCs are logged, the QTC window is closed.
// To prevent data loss due to an accidental abort, the user is asked for confirmation first.
func (c *Controller) AbortQTCSeries() {
	if !c.questionConfirmAbort() {
		return
	}

	c.view.Close()
}

func (c *Controller) Stop() {
	c.keyer.Stop()
}

func (c *Controller) DoubleStop() {
	c.view.Close()
}

// *****************************************************
// WORKFLOW HELPER METHODS
// *****************************************************

func (c *Controller) setActivePhase(phase core.QTCWorkflowPhase) {
	c.view.SetActivePhase(phase)
	if c.activePhase == phase {
		return
	}

	// enter the phase
	c.activePhase = phase
}

func (c *Controller) sendCurrentQTC() {
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

	c.sendQTC(currentQTC, shortenTime)

	// add transmission data and mark the QTC as transmitted
	currentQTC.Timestamp = c.clock.Now()
	currentQTC.Frequency = c.vfoFrequency
	currentQTC.Band = c.vfoBand
	currentQTC.Mode = c.vfoMode
	c.currentSeries.QTCs[c.currentQTC] = currentQTC
}

// *****************************************************
// SENDING METHODS - NO WORKFLOW BELOW HERE
// *****************************************************

func (c *Controller) RepeatLastTransmission() {
	c.keyer.Repeat()
}

func (c *Controller) sendQTCOffer() {
	c.keyer.SendText(OfferQTCText)
}

func (c *Controller) sendQTCRequest() {
	c.keyer.SendText(RequestQTCText)
}

func (c *Controller) sendRequestRepeat() {
	c.keyer.SendText(RequestRepeatText)
}

func (c *Controller) sendQRV() {
	c.keyer.SendText(QRVText)
}

func (c *Controller) sendConfirm() {
	c.keyer.SendText(ConfirmText)
}

// sendHeader sends the header of the current QTC series.
func (c *Controller) sendHeader() {
	c.keyer.SendText(SendHeaderTemplate, c.currentSeries.Header)
}

// sendQTC sends the given QTC.
func (c *Controller) sendQTC(qtc core.QTC, shortenTime bool) {
	time := qtc.QTCTime.String()
	if shortenTime {
		time = qtc.QTCTime.ShortString()
	}
	call := qtc.QTCCallsign.String()
	exchange := c.keyer.Cut(qtc.QTCNumber.String())

	c.keyer.SendText("%s %s %s", time, call, exchange)
}

// nullView

var _ View = &nullView{}

type nullView struct{}

func (*nullView) QuestionQTCCount(int) (int, bool)     { return 0, false }
func (*nullView) ShowFieldError(core.QTCField, string) {}
func (*nullView) ClearFieldError()                     {}
func (*nullView) Show(core.QTCMode, core.QTCSeries)    {}
func (*nullView) UpdateQTC(int, core.QTC)              {}
func (*nullView) Close()                               {}
func (*nullView) ClearDataInputs()                     {}
func (*nullView) SetActivePhase(core.QTCWorkflowPhase) {}
func (*nullView) SetActiveField(core.QTCField)         {}
func (*nullView) SetActiveQTC(int)                     {}

// nullWorkflow

var _ qtcWorkflow = &nullWorkflow{}

type nullWorkflow struct{}

func (*nullWorkflow) StartAction()                         {}
func (*nullWorkflow) ConfirmStart()                        {}
func (*nullWorkflow) HeaderAction()                        {}
func (*nullWorkflow) ConfirmHeader()                       {}
func (*nullWorkflow) DataAction()                          {}
func (*nullWorkflow) ConfirmData()                         {}
func (*nullWorkflow) GotoNextField()                       {}
func (*nullWorkflow) CompleteQTCSeries()                   {}
func (*nullWorkflow) SetActivePhase(core.QTCWorkflowPhase) {}
