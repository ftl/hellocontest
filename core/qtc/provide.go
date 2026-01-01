package qtc

import "github.com/ftl/hellocontest/core"

var _ qtcWorkflow = &provide{}

type provide struct {
	*Controller
}

func (c *provide) StartAction() {
	c.sendQTCOffer()
}

func (c *provide) ConfirmStart() {
	c.sendHeader()
}

func (c *provide) HeaderAction() {
	c.sendHeader()
}

func (c *provide) ConfirmHeader() {
	c.setActivePhase(core.QTCExchangeData)
	c.setActiveQTC(0)
}

func (c *provide) DataAction() {
	c.sendCurrentQTC()
}

func (c *provide) ConfirmData() {
	if c.currentSeries.IsValidQTCIndex(c.currentQTC) {
		qtc := c.currentSeries.QTCs[c.currentQTC]
		qtc.Confirmed = true
		c.currentSeries.SetData(c.currentQTC, qtc)
		c.view.UpdateQTC(c.currentQTC, qtc)
	}
	if c.currentSeries.IsValidQTCIndex(c.currentQTC + 1) {
		c.setActiveQTC(c.currentQTC + 1)
	} else {
		c.setActivePhase(core.QTCFinish)
	}
}

func (c *provide) setActiveQTC(index int) {
	if !c.currentSeries.IsValidQTCIndex(index) {
		return
	}
	c.setActivePhase(core.QTCExchangeData)

	c.currentQTC = index
	c.view.SetActiveQTC(c.currentQTC)

	c.sendCurrentQTC()
}

func (c *provide) GotoNextField() {
	// no-op???
}

func (c *provide) CompleteQTCSeries() {
	// check if all QTCs have actually been transmitted
	for i, qtc := range c.currentSeries.QTCs {
		if qtc.Confirmed {
			continue
		}

		c.showErrorDialog("Not all QTCs have been confirmed, the QTC series cannot be completed. Abort the series to close the window or transmit the remaining QTCs.")
		c.setActiveQTC(i)
		return
	}

	c.logbook.AddQTCSeries(c.currentSeries)

	c.keyer.SendText(CompleteQTCSeriesText)

	c.view.Close()
}
