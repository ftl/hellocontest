package summary

import (
	"strings"
	"time"

	"github.com/ftl/hellocontest/core"
)

type View interface {
	Show() bool

	SetContestName(string)
	SetCabrilloName(string)
	SetCallsign(string)
	SetMyExchanges(string)

	// SetOperatorMode(string)
	// SetOverlay(string)
	// SetPowerMode(string)
	// SetAssisted(bool)

	SetWorkedModes(string)
	SetWorkedBands(string)
	SetOperatingTime(time.Duration)
	SetBreakTime(time.Duration)
	SetBreaks(int)

	// SetScore(core.Score)

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
	counter ScoreCounter
	view    View

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

	workedModes := make([]string, len(c.summary.WorkedModes))
	for i := range c.summary.WorkedModes {
		workedModes[i] = string(c.summary.WorkedModes[i])
	}
	c.view.SetWorkedModes(strings.Join(workedModes, ", "))
	workedBands := make([]string, len(c.summary.WorkedBands))
	for i := range c.summary.WorkedBands {
		workedBands[i] = string(c.summary.WorkedBands[i])
	}
	c.view.SetWorkedBands(strings.Join(workedBands, ", "))
	c.view.SetOperatingTime(c.summary.TimeReport.OperationTime())
	c.view.SetBreakTime(time.Duration(c.summary.TimeReport.BreakMinutes) * time.Minute)
	c.view.SetBreaks(c.summary.TimeReport.Breaks)

	// TODO: move all the data that should be visible into the view

	c.view.SetOpenAfterExport(c.openAfterExport)
}

func (c *Controller) createSummary(settings core.Settings) *core.Summary {
	result := &core.Summary{
		ContestName:  settings.Contest().Definition.Name,
		CabrilloName: string(settings.Contest().Definition.Identifier),
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
	// TODO: update the
}

func (c *Controller) SetOpenAfterExport(open bool) {
	c.openAfterExport = open
}
