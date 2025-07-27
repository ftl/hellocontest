package summary

import (
	"strings"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hellocontest/core"
)

type View interface {
	Show() bool

	SetContestName(string)
	SetCabrilloName(string)
	SetCallsign(string)
	SetMyExchanges(string)

	SetOperatorMode(string)
	SetOverlay(string)
	SetPowerMode(string)
	SetAssisted(bool)

	SetWorkedModes(string)
	SetWorkedBands(string)
	SetOperatingTime(time.Duration)
	SetBreakTime(time.Duration)
	SetBreaks(int)

	SetScore(core.Score)

	SetOpenAfterExport(bool)
}

type ScoreCounter interface {
	FillSummary(*core.Summary)
}

type Result struct {
	Summary         core.Summary
	OpenAfterExport bool
}

type Controller struct {
	counter  ScoreCounter
	settings core.Settings
	view     View

	summary *core.Summary

	openAfterExport bool
}

func NewController(counter ScoreCounter) *Controller {
	result := &Controller{
		counter: counter,
	}

	return result
}

func (c *Controller) SetView(view View) {
	if view == nil {
		panic("summary.Controller.SetView must not be called with nil")
	}
	if c.view != nil {
		panic("summary.Controller.SetView was already called")
	}

	c.view = view
}

func (c *Controller) Run(settings core.Settings) (Result, bool) {
	c.settings = settings
	c.summary = c.createSummary(settings)

	c.showSummary()

	accepted := c.view.Show()
	if !accepted {
		return Result{}, false
	}

	result := Result{
		Summary: *c.summary,

		OpenAfterExport: c.openAfterExport,
	}
	return result, true
}

func (c *Controller) showSummary() {
	c.view.SetContestName(c.summary.ContestName)
	c.view.SetCabrilloName(c.summary.CabrilloName)
	c.view.SetCallsign(c.summary.Callsign.String())
	c.view.SetMyExchanges(c.summary.MyExchanges)

	c.view.SetOperatorMode(string(c.summary.OperatorMode))
	c.view.SetOverlay(string(c.summary.Overlay))
	c.view.SetPowerMode(string(c.summary.PowerMode))
	c.view.SetAssisted(c.summary.Assisted)

	c.view.SetWorkedModes(strings.Join(c.summary.WorkedModes, ", "))
	c.view.SetWorkedBands(strings.Join(c.summary.WorkedBands, ", "))
	c.view.SetOperatingTime(c.summary.TimeReport.OperationTime())
	c.view.SetBreakTime(time.Duration(c.summary.TimeReport.BreakMinutes) * time.Minute)
	c.view.SetBreaks(c.summary.TimeReport.Breaks)

	c.view.SetScore(c.summary.Score)

	c.view.SetOpenAfterExport(c.openAfterExport)
}

func (c *Controller) createSummary(settings core.Settings) *core.Summary {
	result := &core.Summary{
		ContestName:  settings.Contest().Definition.Name,
		CabrilloName: string(settings.Contest().Definition.Identifier),
		StartTime:    settings.Contest().StartTime,
		Callsign:     settings.Station().Callsign,
	}

	// calculate MyExchanges
	myExchanges := make([]string, 0, len(settings.Contest().MyExchangeFields))
	for _, value := range settings.Contest().ExchangeValues {
		if value != "" {
			myExchanges = append(myExchanges, value)
		}
	}
	result.MyExchanges = strings.Join(myExchanges, ", ")

	c.counter.FillSummary(result)

	return result
}

func (c *Controller) updateSummary() {
	c.counter.FillSummary(c.summary)
	c.showSummary()
}

func (c *Controller) OperatorModes() []string {
	return []string{
		string(conval.SingleOperator),
		string(conval.MultiOperator),
	}
}

func (c *Controller) SetOperatorMode(operatorMode string) {
	c.summary.OperatorMode = conval.OperatorMode(operatorMode)
	c.updateSummary()
}

func (c *Controller) Overlays() []string {
	result := make([]string, len(c.settings.Contest().Definition.Overlays))
	for i, overlay := range c.settings.Contest().Definition.Overlays {
		result[i] = string(overlay)
	}
	return result
}

func (c *Controller) SetOverlay(overlay string) {
	c.summary.Overlay = conval.Overlay(overlay)
	c.updateSummary()
}

func (c *Controller) PowerModes() []string {
	return []string{
		string(conval.QRPPower),
		string(conval.LowPower),
		string(conval.HighPower),
	}
}

func (c *Controller) SetAssisted(assisted bool) {
	c.summary.Assisted = assisted
	c.updateSummary()
}

func (c *Controller) SetPowerMode(powerMode string) {
	c.summary.PowerMode = conval.PowerMode(powerMode)
	c.updateSummary()
}

func (c *Controller) SetOpenAfterExport(open bool) {
	c.openAfterExport = open
}
