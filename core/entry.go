package core

import (
	"fmt"
	"strconv"

	"github.com/ftl/hamradio/callsign"
)

// EntryController controls the entry of QSO data.
type EntryController interface {
	SetView(EntryView)

	GotoNextField() EntryField
	GetActiveField() EntryField
	SetActiveField(EntryField)

	Log()
	Reset()
}

// EntryView represents the visual part of the QSO data entry.
type EntryView interface {
	SetEntryController(EntryController)

	GetCallsign() string
	SetCallsign(string)
	GetTheirReport() string
	SetTheirReport(string)
	GetTheirNumber() string
	SetTheirNumber(string)
	GetBand() string
	SetBand(text string)
	GetMode() string
	SetMode(text string)
	GetMyReport() string
	SetMyReport(string)
	GetMyNumber() string
	SetMyNumber(string)

	SetActiveField(EntryField)
	SetDuplicateMarker(bool)
	ShowError(error)
	ClearError()
}

// EntryField represents an entry field in the visual part.
type EntryField int

// The entry fields.
const (
	CallsignField EntryField = iota
	TheirReportField
	TheirNumberField
	MyReportField
	MyNumberField
	OtherField
)

// NewEntryController returns a new EntryController.
func NewEntryController(clock Clock, log Log) EntryController {
	return &entryController{
		clock: clock,
		log:   log,
	}
}

type entryController struct {
	clock       Clock
	log         Log
	view        EntryView
	activeField EntryField
}

func (c *entryController) SetView(view EntryView) {
	c.view = view
	c.view.SetEntryController(c)
	c.Reset()
}

func (c *entryController) GotoNextField() EntryField {
	switch c.activeField {
	case CallsignField:
		c.leaveCallsignField()
	}

	transitions := map[EntryField]EntryField{
		CallsignField:    TheirReportField,
		TheirReportField: TheirNumberField,
		TheirNumberField: CallsignField,
		MyReportField:    CallsignField,
		MyNumberField:    CallsignField,
	}
	c.activeField = transitions[c.activeField]
	c.view.SetActiveField(c.activeField)
	return c.activeField
}

func (c *entryController) leaveCallsignField() {
	callsign, err := callsign.Parse(c.view.GetCallsign())
	if err != nil {
		fmt.Println(err)
		return
	}

	qso, found := c.log.Find(callsign)
	if !found {
		c.view.SetDuplicateMarker(false)
		return
	}

	c.view.SetBand(string(qso.Band))
	c.view.SetMode(string(qso.Mode))
	c.view.SetTheirReport(string(qso.TheirReport))
	c.view.SetTheirNumber(qso.TheirNumber.String())
	c.view.SetMyReport(string(qso.MyReport))
	c.view.SetMyNumber(qso.MyNumber.String())
	c.view.SetDuplicateMarker(true)
}

func (c *entryController) GetActiveField() EntryField {
	return c.activeField
}

func (c *entryController) SetActiveField(field EntryField) {
	c.activeField = field
}

func (c *entryController) Log() {
	var err error
	qso := QSO{}
	qso.Callsign, err = callsign.Parse(c.view.GetCallsign())
	if err != nil {
		c.showErrorOnField(err, CallsignField)
		return
	}
	qso.Time = c.clock.Now()

	qso.Band, err = ParseBand(c.view.GetBand())
	if err != nil {
		c.view.ShowError(err)
		return
	}

	qso.Mode, err = ParseMode(c.view.GetMode())
	if err != nil {
		c.view.ShowError(err)
		return
	}

	qso.TheirReport, err = ParseRST(c.view.GetTheirReport())
	if err != nil {
		c.showErrorOnField(err, TheirReportField)
		return
	}

	theirNumber, err := strconv.Atoi(c.view.GetTheirNumber())
	if err != nil {
		c.showErrorOnField(err, TheirNumberField)
		return
	}
	qso.TheirNumber = QSONumber(theirNumber)

	qso.MyReport, err = ParseRST(c.view.GetMyReport())
	if err != nil {
		c.showErrorOnField(err, MyReportField)
		return
	}

	myNumber, err := strconv.Atoi(c.view.GetMyNumber())
	if err != nil {
		c.showErrorOnField(err, MyNumberField)
		return
	}
	qso.MyNumber = QSONumber(myNumber)

	duplicateQso, duplicate := c.log.Find(qso.Callsign)
	if duplicate && duplicateQso.MyNumber != qso.MyNumber {
		c.showErrorOnField(fmt.Errorf("%s was worked before in QSO #%s", qso.Callsign, duplicateQso.MyNumber.String()), CallsignField)
		return
	}

	c.log.Log(qso)
	c.Reset()
}

func (c *entryController) showErrorOnField(err error, field EntryField) {
	c.activeField = field
	c.view.SetActiveField(c.activeField)
	c.view.ShowError(err)
}

func (c *entryController) Reset() {
	nextNumber := c.log.GetNextNumber()
	c.activeField = CallsignField
	c.view.SetCallsign("")
	c.view.SetTheirReport("599")
	c.view.SetTheirNumber("")
	c.view.SetMyReport("599")
	c.view.SetMyNumber(nextNumber.String())
	c.view.SetActiveField(c.activeField)
	c.view.SetDuplicateMarker(false)
	c.view.ClearError()
}
