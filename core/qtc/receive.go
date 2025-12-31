package qtc

import (
	"github.com/ftl/hellocontest/core"
)

var _ qtcWorkflow = &receive{}

type receive struct {
	*Controller
}

func (c *receive) StartAction() {
	c.sendQTCRequest()
}

func (c *receive) ConfirmStart() {
	c.SetActiveField(core.QTCHeaderField)
	c.sendQRV()
}

func (c *receive) HeaderAction() {
	c.sendRequestRepeat()
}

func (c *receive) ConfirmHeader() {
	// parse and validate the header
	header, err := c.currentHeader()
	if err != nil {
		c.showFieldError(core.QTCHeaderField, err)
		c.SetActiveField(core.QTCHeaderField)
		return
	}
	c.currentSeries.Header = header
	c.clearErrorMessage()

	// progress in the workflow
	c.currentQTC = 0
	c.setActivePhase(core.QTCExchangeData)
	c.SetActiveField(core.QTCTimestampField)
	c.sendConfirm()
}

func (c *receive) DataAction() {
	c.sendRequestRepeat()
}

func (c *receive) ConfirmData() {
	// parse the entered QTC Data
	qtc := core.QTC{
		Kind:      core.ReceivedQTC,
		Timestamp: c.clock.Now(),
		Confirmed: true,
	}

	qtcTime, err := c.currentQTCTimestamp()
	if err != nil {
		c.showFieldError(core.QTCTimestampField, err)
		c.SetActiveField(core.QTCTimestampField)
		return
	}
	qtc.QTCTime = qtcTime

	qtcCallsign, err := c.currentQTCCallsign()
	if err != nil {
		c.showFieldError(core.QTCCallsignField, err)
		c.SetActiveField(core.QTCCallsignField)
		return
	}
	qtc.QTCCallsign = qtcCallsign

	qtcNumber, err := c.currentQTCNumber()
	if err != nil {
		c.showFieldError(core.QTCExchangeField, err)
		c.SetActiveField(core.QTCExchangeField)
		return
	}
	qtc.QTCNumber = qtcNumber
	c.clearErrorMessage()

	// log the entered QTC data in the current series
	c.currentSeries.SetData(c.currentQTC, qtc)
	c.view.UpdateQTC(c.currentQTC, qtc)

	// clear the current input
	delete(c.currentInput, core.QTCTimestampField)
	delete(c.currentInput, core.QTCCallsignField)
	delete(c.currentInput, core.QTCExchangeField)
	c.view.ClearDataInputs()

	// progress in the workflow
	c.currentQTC += 1
	if c.currentSeries.IsComplete() {
		c.setActivePhase(core.QTCFinish)
	} else {
		c.SetActiveField(core.QTCTimestampField)
	}
	c.sendConfirm()
}

func (c *receive) GotoNextField() {
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

func (c *receive) CompleteQTCSeries() {
	// fill common data from the series and log the QTC series
	for i, qtc := range c.currentSeries.QTCs {
		qtc.TheirCallsign = c.currentSeries.TheirCallsign
		qtc.Header = c.currentSeries.Header
		qtc.Frequency = c.vfoFrequency
		qtc.Band = c.vfoBand
		qtc.Mode = c.vfoMode
		c.currentSeries.QTCs[i] = qtc
	}

	c.logbook.LogQTCSeries(c.currentSeries)

	c.view.Close()
}
