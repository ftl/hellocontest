package qtc

import (
	"fmt"
	"strconv"

	"github.com/ftl/hellocontest/core/model"
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
	LastCallsign() model.Callsign
	AddQTCSeries(model.QTCSeries)
	PrepareFor(model.Callsign, int) []model.QTC
}

type QTCList interface {
	SelectLastQTC()
	SetQTCsEnabled(bool)
}

type EntryController interface {
	CurrentQSOState() (model.Callsign, model.QSODataState)
	CurrentVFOState() (model.Frequency, model.Band, model.Mode)
	FocusedVFO() (string, bool)
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
	ShowFieldError(model.QTCField, string)
	ClearFieldError()
	Show(model.QTCMode, model.QTCSeries, string)
	UpdateQTC(int, model.QTC)
	Close()
	ClearDataInputs()
	SetActivePhase(model.QTCWorkflowPhase)
	SetActiveField(model.QTCField)
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

	clock           model.Clock
	logbook         Logbook
	qtcList         QTCList
	entryController EntryController
	keyer           Keyer

	infoDialogs InfoDialogs
	view        View

	activePhase model.QTCWorkflowPhase
	activeField model.QTCField

	currentMode   model.QTCMode
	currentSeries model.QTCSeries
	currentQTC    int
	currentInput  map[model.QTCField]string
}

func NewController(clock model.Clock, infoDialogs InfoDialogs, logbook Logbook, qtcList QTCList, entryController EntryController, keyer Keyer) *Controller {
	return &Controller{
		workflow:        new(nullWorkflow),
		clock:           clock,
		logbook:         logbook,
		qtcList:         qtcList,
		entryController: entryController,
		keyer:           keyer,
		infoDialogs:     infoDialogs,
		view:            new(nullView),
		currentInput:    make(map[model.QTCField]string),
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

func (c *Controller) ContestChanged(contest model.Contest) {
	c.enabled = contest.EnableQTCs
	c.qtcList.SetQTCsEnabled(contest.EnableQTCs)
}

func (c *Controller) questionInvalidQSOData(call model.Callsign) bool {
	return c.infoDialogs.ShowQuestion("The entered callsign is valid, but the QSO data is invalid. Proceed with the entered callsign of %s?", call.String())
}

func (c *Controller) questionConfirmAbort() bool {
	return c.infoDialogs.ShowQuestion("The QTC exchange is incomplete. Do you want to cancel it?")
}

func (c *Controller) showErrorDialog(format string, args ...any) {
	c.infoDialogs.ShowError(format, args...)
}

func (c *Controller) showFieldError(field model.QTCField, err error) {
	c.view.ShowFieldError(field, err.Error())
}

func (c *Controller) clearErrorMessage() {
	c.view.ClearFieldError()
}

// OfferQTC initiates the QTC dialog with mode model.ProvideQTC.
func (c *Controller) OfferQTC() {
	// 1. find out their callsign
	theirCall, ok := c.findOutTheirCallsign()
	if !ok {
		return
	}

	// 2. get available QTCs
	qtcs := c.logbook.PrepareFor(theirCall, model.MaxQTCsPerCall)
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
	qtcSeries, err := model.NewQTCSeries(c.logbook.NextSeriesNumber(), qtcs)
	if err != nil {
		c.showErrorDialog("%v", err)
		return
	}
	c.workflow = &provide{Controller: c}
	c.currentMode = model.ProvideQTC
	c.currentSeries = qtcSeries
	c.currentQTC = model.NoQTCIndex

	// 5. enter the first phase: send "qtc"
	c.setActivePhase(model.QTCStart)
	c.sendQTCOffer()

	// 6. show and run the QTC dialog
	c.view.Show(c.currentMode, c.currentSeries, c.so2vVFOName())
}

func (c *Controller) so2vVFOName() string {
	name, so2v := c.entryController.FocusedVFO()
	if !so2v {
		return ""
	}
	return name
}

func (c *Controller) findOutTheirCallsign() (model.Callsign, bool) {
	theirCall, currentQSOState := c.entryController.CurrentQSOState()
	switch currentQSOState {
	case model.QSODataValid: // a) there is currently a valid QSO in the entry fields that is not yet logged -> log this QSO and take their callsign
		c.entryController.Log()
	case model.QSODataInvalid: // b) there is currently a valid callsign and some QSO data (but not valid) in the entry fields -> show info about invalid QSO data, ask if the callsign should be used -> use the callsign
		if !c.questionInvalidQSOData(theirCall) {
			return model.Callsign{}, false
		}
	case model.QSODataEmpty: // c) there is currently a valid callsign in the entry field, but no QSO data at all-> use this callsign
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

// OfferQTC initiates the QTC dialog with mode model.ReceiveQTC.
func (c *Controller) RequestQTC() {
	// find out their callsign
	theirCall, ok := c.findOutTheirCallsign()
	if !ok {
		return
	}

	// TODO: find out how many QTCs may be received from theirCall (is this relevant, maybe we should just take what we get)

	// create a new and empty QTCSeries
	qtcSeries := model.NewReceivingQTCSeries(theirCall)
	c.workflow = &receive{Controller: c}
	c.currentMode = model.ReceiveQTC
	c.currentSeries = qtcSeries
	c.currentQTC = model.NoQTCIndex

	// enter the first phase
	c.setActivePhase(model.QTCStart)

	// show and run the QTC dialog
	c.view.Show(c.currentMode, c.currentSeries, c.so2vVFOName())
}

// *****************************************************
// COMMON WORKFLOW METHODS
// *****************************************************

func (c *Controller) StartAction() {
	if c.activePhase != model.QTCStart {
		return
	}

	c.workflow.StartAction()
}

func (c *Controller) ConfirmStart() {
	if c.activePhase != model.QTCStart {
		return
	}
	c.setActivePhase(model.QTCExchangeHeader)

	c.workflow.ConfirmStart()
}

func (c *Controller) HeaderAction() {
	if c.activePhase != model.QTCExchangeHeader {
		return
	}

	c.workflow.HeaderAction()
}

func (c *Controller) ConfirmHeader() {
	if c.activePhase != model.QTCExchangeHeader {
		return
	}

	c.workflow.ConfirmHeader()
}

func (c *Controller) DataAction() {
	if c.activePhase != model.QTCExchangeData {
		return
	}

	c.workflow.DataAction()
}

func (c *Controller) ConfirmData() {
	if c.activePhase != model.QTCExchangeData {
		return
	}

	c.workflow.ConfirmData()
}

func (c *Controller) currentHeader() (model.QTCHeader, error) {
	headerStr, ok := c.currentInput[model.QTCHeaderField]
	if !ok || (headerStr == "") {
		return model.QTCHeader{}, fmt.Errorf("the header field is empty")
	}
	return model.ParseQTCHeader(headerStr)
}

func (c *Controller) currentQTCTimestamp() (model.QTCTime, error) {
	qtcTimeStr, ok := c.currentInput[model.QTCTimestampField]
	if !ok || (qtcTimeStr == "") {
		return model.ZeroQTCTime, fmt.Errorf("the QTC timestamp field is empty")
	}
	return model.ParseQTCTime(qtcTimeStr, c.currentReferenceTime())
}

func (c *Controller) currentReferenceTime() model.QTCTime {
	previousQTC := c.currentQTC - 1
	if previousQTC >= len(c.currentSeries.QTCs) {
		previousQTC = len(c.currentSeries.QTCs) - 1
	}
	if previousQTC < 0 {
		return model.ZeroQTCTime
	}
	return c.currentSeries.QTCs[previousQTC].QTCTime
}

func (c *Controller) currentQTCCallsign() (model.Callsign, error) {
	qtcCallsignStr, ok := c.currentInput[model.QTCCallsignField]
	if !ok || (qtcCallsignStr == "") {
		return model.Callsign{}, fmt.Errorf("the QTC callsign field is empty")
	}
	qtcCallsign, err := model.ParseCallsign(qtcCallsignStr)
	if err != nil {
		return model.Callsign{}, err
	}
	return qtcCallsign, nil
}

func (c *Controller) currentQTCNumber() (model.QSONumber, error) {
	qtcExchangeStr, ok := c.currentInput[model.QTCExchangeField]
	if !ok || (qtcExchangeStr == "") {
		return 0, fmt.Errorf("the QTC exchange field is empty")
	}
	qtcExchange, err := strconv.Atoi(qtcExchangeStr)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid QTC exchange", qtcExchangeStr)
	}
	return model.QSONumber(qtcExchange), nil
}

func (c *Controller) DoAction() {
	switch c.activePhase {
	case model.QTCStart:
		c.StartAction()
	case model.QTCExchangeHeader:
		c.HeaderAction()
	case model.QTCExchangeData:
		c.DataAction()
	}
}

func (c *Controller) DoConfirm() {
	switch c.activePhase {
	case model.QTCStart:
		c.ConfirmStart()
	case model.QTCExchangeHeader:
		c.ConfirmHeader()
	case model.QTCExchangeData:
		c.ConfirmData()
	case model.QTCFinish:
		c.CompleteQTCSeries()
	}
}

func (c *Controller) Enter(s string) {
	c.currentInput[c.activeField] = s
}

func (c *Controller) GotoNextField() {
	c.workflow.GotoNextField()
}

func (c *Controller) SetActiveField(field model.QTCField) {
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

func (c *Controller) setActivePhase(phase model.QTCWorkflowPhase) {
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
	currentQTC.Frequency, currentQTC.Band, currentQTC.Mode = c.entryController.CurrentVFOState()
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
func (c *Controller) sendQTC(qtc model.QTC, shortenTime bool) {
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

func (*nullView) QuestionQTCCount(int) (int, bool)            { return 0, false }
func (*nullView) ShowFieldError(model.QTCField, string)       {}
func (*nullView) ClearFieldError()                            {}
func (*nullView) Show(model.QTCMode, model.QTCSeries, string) {}
func (*nullView) UpdateQTC(int, model.QTC)                    {}
func (*nullView) Close()                                      {}
func (*nullView) ClearDataInputs()                            {}
func (*nullView) SetActivePhase(model.QTCWorkflowPhase)       {}
func (*nullView) SetActiveField(model.QTCField)               {}
func (*nullView) SetActiveQTC(int)                            {}

// nullWorkflow

var _ qtcWorkflow = &nullWorkflow{}

type nullWorkflow struct{}

func (*nullWorkflow) StartAction()                          {}
func (*nullWorkflow) ConfirmStart()                         {}
func (*nullWorkflow) HeaderAction()                         {}
func (*nullWorkflow) ConfirmHeader()                        {}
func (*nullWorkflow) DataAction()                           {}
func (*nullWorkflow) ConfirmData()                          {}
func (*nullWorkflow) GotoNextField()                        {}
func (*nullWorkflow) CompleteQTCSeries()                    {}
func (*nullWorkflow) SetActivePhase(model.QTCWorkflowPhase) {}
