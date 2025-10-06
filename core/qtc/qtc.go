package qtc

import (
	"fmt"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

// text prompts to communicate with the opposite station
const (
	OfferQTCText   = "qtc"
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
	SendText(text string)
	Repeat()
	Stop()
}

type View interface {
	QuestionInvalidQSOData() bool
	QuestionQTCCount(max int) (int, bool)
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

// Workflow for receiving QTCs

func (c *Controller) RequestQTC() {
	// TODO implement workflow for receiving QTCs
}

// nullView

var _ View = &nullView{}

type nullView struct{}

func (*nullView) QuestionInvalidQSOData() bool     { return false }
func (*nullView) QuestionQTCCount(int) (int, bool) { return 0, false }
func (*nullView) ShowError(error)                  {}
func (*nullView) ShowSendWindow(core.QTCSeries)    {}
func (*nullView) Update(core.QTCSeries)            {}
func (*nullView) Close()                           {}
