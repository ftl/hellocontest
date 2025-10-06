package qtc

import (
	"fmt"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

// text prompts to communicate with the opposite station
const (
	OfferQTCText          = "qtc"
	SendHeaderTemplate    = "qtc %s"
	CompleteQTCSeriesText = "tu"

	RequestQTCText = "qtc?"
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
	SendText(text string, args ...any)
	Repeat()
	Stop()
}

type View interface {
	QuestionInvalidQSOData() bool
	QuestionQTCCount(max int) (int, bool)
	QuestionConfirmAbort() bool
	ShowError(error)
	ShowSendWindow(core.QTCSeries)
	Update(core.QTCSeries)
	Close()
}

type Controller struct {
	logbook         Logbook
	qtcList         QTCList
	entryController EntryController
	keyer           Keyer

	view View

	currentSeries core.QTCSeries
	currentQTC    int
}

func NewController(logbook Logbook, qtcList QTCList, entryController EntryController, keyer Keyer) *Controller {
	return &Controller{
		logbook:         logbook,
		qtcList:         qtcList,
		entryController: entryController,
		keyer:           keyer,
		view:            new(nullView),
	}
}

func (c *Controller) SetView(view View) {
	if view == nil {
		c.view = new(nullView)
		return
	}
	c.view = view
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
		c.view.ShowError(err)
		return
	}
	c.currentSeries = qtcSeries
	c.currentQTC = 0

	// 5. show QTC window for sending
	c.view.ShowSendWindow(c.currentSeries)

	// 6. send "qtc"
	c.keyer.SendText(OfferQTCText)
}

func (c *Controller) findOutTheirCallsign() (callsign.Callsign, bool) {
	theirCall, currentQSOState := c.entryController.CurrentQSOState()
	switch currentQSOState {
	case core.QSODataValid: // a) there is currently a valid QSO in the entry fields that is not yet logged -> log this QSO and take their callsign
		c.entryController.Log()
	case core.QSODataInvalid: // b) there is currently a valid callsign and some QSO data (but not valid) in the entry fields -> show info about invalid QSO data, ask if the callsign should be used -> use the callsign
		if !c.view.QuestionInvalidQSOData() {
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

// SendHeader sends the header of the current QTC series.
func (c *Controller) SendHeader() {
	// send the header
	c.keyer.SendText(SendHeaderTemplate, c.currentSeries.Header)

	// TODO: advance UI focus to the first QTC?
}

// SendQTC sends the current QTC.
func (c *Controller) SendQTC() {
	qtc := c.currentSeries.QTCs[c.currentQTC]
	time := qtc.QTCTime.String()
	call := qtc.QTCCallsign.String()
	exchange := qtc.QTCNumber

	// shorten time if the last QTC qso was in the same hour
	if c.currentQTC > 0 {
		lastQTC := c.currentSeries.QTCs[c.currentQTC-1]
		if lastQTC.QTCTime.Hour == qtc.QTCTime.Hour {
			// TODO: time = shortened time
		}
	}

	c.keyer.SendText("%s %s %d", time, call, exchange)

	// TODO: advance UI focus to next QTC?
}

// CompleteQTCSeries completes the current QTC series, stores all QTCs to the log, sends "tu",
// and closes the QTC window.
// The series can only be completed when all QTCs have been transmitted. Otherwise, an
// error message is presented to the user, the QTC window stays open.
func (c *Controller) CompleteQTCSeries() {
	// TODO: check if all QTCs have been transmitted -> otherwise c.view.ShowError("Not all QTCs have been transmitted. The QTC series cannot be completed")
	// and focus the first QTC that has not been transmitted yet

	for _, qtc := range c.currentSeries.QTCs {
		c.logbook.LogQTC(qtc)
	}

	c.keyer.SendText(CompleteQTCSeriesText)

	c.view.Close()
}

// AbortQTCSeries aborts the current QTC series: no QTCs are logged, the QTC window is closed.
// To prevent data loss due to an accidental abort, the user is asked for confirmation first.
func (c *Controller) AbortQTCSeries() {
	if !c.view.QuestionConfirmAbort() {
		return
	}

	c.view.Close()
}

// Workflow for receiving QTCs

func (c *Controller) RequestQTC() {
	// TODO implement workflow for receiving QTCs
}

// nullView

var _ View = &nullView{}

type nullView struct{}

func (*nullView) QuestionInvalidQSOData() bool     { return false }
func (*nullView) QuestionQTCCount(int) (int, bool) { return 0, false }
func (*nullView) QuestionConfirmAbort() bool       { return false }
func (*nullView) ShowError(error)                  {}
func (*nullView) ShowSendWindow(core.QTCSeries)    {}
func (*nullView) Update(core.QTCSeries)            {}
func (*nullView) Close()                           {}
