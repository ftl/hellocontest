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
	Show(core.QTCMode, core.QTCSeries)
	Update(core.QTCSeries)
	Close()
	SetActiveField(core.QTCField)
}

type Controller struct {
	clock           core.Clock
	logbook         Logbook
	qtcList         QTCList
	entryController EntryController
	keyer           Keyer

	view View

	activeField core.QTCField

	currentMode   core.QTCMode
	currentSeries core.QTCSeries
	currentQTC    int

	vfoFrequency core.Frequency
	vfoBand      core.Band
	vfoMode      core.Mode
}

func NewController(clock core.Clock, logbook Logbook, qtcList QTCList, entryController EntryController, keyer Keyer) *Controller {
	return &Controller{
		clock:           clock,
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

func (c *Controller) VFOFrequencyChanged(frequency core.Frequency) {
	c.vfoFrequency = frequency
}

func (c *Controller) VFOBandChanged(band core.Band) {
	c.vfoBand = band
}

func (c *Controller) VFOModeChanged(mode core.Mode) {
	c.vfoMode = mode
}

func (c *Controller) Proceed() {
	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		switch {
		case c.activeField.IsHeader():
			c.SendHeader()
		case c.activeField.IsQTC():
			c.SendQTC()
		case c.activeField == core.CompleteField:
			c.CompleteQTCSeries()
		default:
			return
		}
	} else {
		// TODO: check if all fields of the current QTC are filled with valid data
		// if not -> focus the first field of the current QTC and request repeat
		c.keyer.SendText(ConfirmText)
	}
	c.GotoNextField()
}

func (c *Controller) Repeat() {
	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		c.keyer.Repeat()
	} else {
		c.keyer.SendText(RequestRepeatText)
	}
}

func (c *Controller) GotoNextField() {
	var (
		nextField core.QTCField
		ok        bool
	)
	qtcIndex := c.activeField.QTCIndex()

	// TODO: use polymorphism for the two modes
	if c.currentMode == core.ProvideQTC {
		switch {
		case c.activeField.IsHeader():
			nextField = core.QTCSendField(0)
		case c.activeField.IsQTC():
			nextField, ok = c.nextQTCField(qtcIndex)
			if !ok {
				return
			}
		default:
			return
		}
	} else {
		switch {
		case c.activeField == core.HeaderSequenceField:
			nextField = core.HeaderCountField
		case c.activeField == core.HeaderCountField:
			nextField = core.QTCTimeField(0)
		case c.activeField.IsTime():
			nextField = core.QTCCallsignField(qtcIndex)
		case c.activeField.IsCallsign():
			nextField = core.QTCNumberField(qtcIndex)
		case c.activeField.IsNumber():
			nextField, ok = c.nextQTCField(qtcIndex)
			if !ok {
				return
			}
		default:
			return
		}
	}

	c.SetActiveField(nextField)
	c.view.SetActiveField(nextField)
}

func (c *Controller) nextQTCField(index int) (core.QTCField, bool) {
	if c.currentSeries.IsLastQTCIndex(index) {
		return core.CompleteField, true
	}

	nextIndex, ok := c.nextQTCIndex(index)
	if !ok {
		return core.NoQTCField, false
	}

	if c.currentMode == core.ProvideQTC {
		return core.QTCSendField(nextIndex), true
	} else {
		return core.QTCTimeField(nextIndex), true
	}

}

func (c *Controller) nextQTCIndex(index int) (int, bool) {
	if !c.currentSeries.IsValidQTCIndex(index) {
		return core.NoQTCIndex, false
	}
	nextIndex := index + 1
	if !c.currentSeries.IsValidQTCIndex(nextIndex) {
		return core.NoQTCIndex, false
	}
	return nextIndex, true
}

func (c *Controller) SetActiveField(field core.QTCField) {
	qtcIndex := field.QTCIndex()
	if !(qtcIndex == core.NoQTCIndex || c.currentSeries.IsValidQTCIndex(qtcIndex)) {
		return
	}
	c.activeField = field
	c.currentQTC = qtcIndex
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
	c.currentMode = core.ProvideQTC
	c.currentSeries = qtcSeries
	c.currentQTC = core.NoQTCIndex

	// 5. show QTC window for sending
	c.view.Show(c.currentMode, c.currentSeries)
	c.SetActiveField(core.HeaderCountField)

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
	c.keyer.SendText(SendHeaderTemplate, c.currentSeries.Header)
}

// SendQTC sends the current QTC.
func (c *Controller) SendQTC() {
	if !c.currentSeries.IsValidQTCIndex(c.currentQTC) {
		return
	}

	qtc := c.currentSeries.QTCs[c.currentQTC]
	time := qtc.QTCTime.String()
	call := qtc.QTCCallsign.String()
	exchange := strconv.Itoa(int(qtc.QTCNumber)) // TODO: shorten numbers

	// shorten time if the last QTC qso was in the same hour
	if c.currentQTC > 0 {
		lastQTC := c.currentSeries.QTCs[c.currentQTC-1]
		if lastQTC.QTCTime.Hour == qtc.QTCTime.Hour {
			// TODO: time = shortened time
		}
	}

	c.keyer.SendText("%s %s %s", time, call, exchange)

	// add transmission data and mark the QTC as transmitted
	qtc.Timestamp = c.clock.Now()
	qtc.Frequency = c.vfoFrequency
	qtc.Band = c.vfoBand
	qtc.Mode = c.vfoMode
	c.currentSeries.QTCs[c.currentQTC] = qtc
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
			if qtc.WasTransmitted() {
				continue
			}

			c.currentQTC = i
			c.SetActiveField(core.QTCSendField(i))
			c.view.ShowError(fmt.Errorf("Not all QTCs have been transmitted, the QTC series cannot be completed. Abort the series to close the window or transmit the remaining QTCs."))
			return
		}

		for _, qtc := range c.currentSeries.QTCs {
			c.logbook.LogQTC(qtc)
		}

		c.keyer.SendText(CompleteQTCSeriesText)

		c.view.Close()
	} else {
		// TODO: implement c.currentMode == core.ReceiveQTC
	}
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

func (*nullView) QuestionInvalidQSOData() bool      { return false }
func (*nullView) QuestionQTCCount(int) (int, bool)  { return 0, false }
func (*nullView) QuestionConfirmAbort() bool        { return false }
func (*nullView) ShowError(error)                   {}
func (*nullView) Show(core.QTCMode, core.QTCSeries) {}
func (*nullView) Update(core.QTCSeries)             {}
func (*nullView) Close()                            {}
func (*nullView) SetActiveField(core.QTCField)      {}
