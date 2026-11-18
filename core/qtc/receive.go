package qtc

import (
	"github.com/ftl/hellocontest/core/model"
)

var _ qtcWorkflow = &receive{}

type receive struct {
	*Controller
}

func (c *receive) StartAction() {
	c.sendQTCRequest()
}

func (c *receive) ConfirmStart() {
	c.SetActiveField(model.QTCHeaderField)
	c.sendQRV()
}

func (c *receive) HeaderAction() {
	c.sendRequestRepeat()
}

func (c *receive) ConfirmHeader() {
	// parse and validate the header
	header, err := c.currentHeader()
	if err != nil {
		c.showFieldError(model.QTCHeaderField, err)
		c.SetActiveField(model.QTCHeaderField)
		return
	}
	c.currentSeries.Header = header
	c.clearErrorMessage()

	// progress in the workflow
	c.currentQTC = 0
	c.setActivePhase(model.QTCExchangeData)
	c.SetActiveField(model.QTCTimestampField)
	c.sendConfirm()
}

func (c *receive) DataAction() {
	c.sendRequestRepeat()
}

func (c *receive) ConfirmData() {
	// parse the entered QTC Data
	qtc := model.QTC{
		Kind:      model.ReceivedQTC,
		Timestamp: c.clock.Now(),
		Confirmed: true,
	}

	qtcTime, err := c.currentQTCTimestamp()
	if err != nil {
		c.showFieldError(model.QTCTimestampField, err)
		c.SetActiveField(model.QTCTimestampField)
		return
	}
	qtc.QTCTime = qtcTime

	qtcCallsign, err := c.currentQTCCallsign()
	if err != nil {
		c.showFieldError(model.QTCCallsignField, err)
		c.SetActiveField(model.QTCCallsignField)
		return
	}
	qtc.QTCCallsign = qtcCallsign

	qtcNumber, err := c.currentQTCNumber()
	if err != nil {
		c.showFieldError(model.QTCExchangeField, err)
		c.SetActiveField(model.QTCExchangeField)
		return
	}
	qtc.QTCNumber = qtcNumber
	c.clearErrorMessage()

	// log the entered QTC data in the current series
	c.currentSeries.SetData(c.currentQTC, qtc)
	c.view.UpdateQTC(c.currentQTC, qtc)

	// clear the current input
	delete(c.currentInput, model.QTCTimestampField)
	delete(c.currentInput, model.QTCCallsignField)
	delete(c.currentInput, model.QTCExchangeField)
	c.view.ClearDataInputs()

	// progress in the workflow
	c.currentQTC += 1
	if c.currentSeries.IsComplete() {
		c.setActivePhase(model.QTCFinish)
	} else {
		c.SetActiveField(model.QTCTimestampField)
	}
	c.sendConfirm()
}

func (c *receive) GotoNextField() {
	nextField := c.activeField
	switch c.activeField {
	case model.QTCHeaderField:
	// ignore, use ConfirmHeader to progress
	case model.QTCTimestampField:
		nextField = model.QTCCallsignField
	case model.QTCCallsignField:
		nextField = model.QTCExchangeField
	case model.QTCExchangeField:
		nextField = model.QTCTimestampField
	}
	c.activeField = nextField
	c.view.SetActiveField(c.activeField)
}

func (c *receive) CompleteQTCSeries() {
	// fill common data from the series and log the QTC series
	frequency, band, mode := c.entryController.CurrentVFOState()
	for i, qtc := range c.currentSeries.QTCs {
		qtc.TheirCallsign = c.currentSeries.TheirCallsign
		qtc.Header = c.currentSeries.Header
		qtc.Frequency = frequency
		qtc.Band = band
		qtc.Mode = mode
		c.currentSeries.QTCs[i] = qtc
	}

	c.logbook.AddQTCSeries(c.currentSeries)

	c.view.Close()
}
