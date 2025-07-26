package summary

import (
	"github.com/ftl/hellocontest/core"
)

type View interface {
	Show() bool

	SetOpenAfterExport(bool)
}

type Result struct {
	Summary         core.Summary
	OpenAfterExport bool
}

type Controller struct {
	view View

	openAfterExport bool
}

func NewController() *Controller {
	result := &Controller{}

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

func (c *Controller) Run() (Result, bool) {
	// TODO: add parameters for contest definition, contest settings, and the score counter (as interface)
	summary := createSummary()

	// TODO: move all the data that should be visible into the view
	c.view.SetOpenAfterExport(c.openAfterExport)

	accepted := c.view.Show()
	if !accepted {
		return Result{}, false
	}

	result := Result{
		Summary: summary,

		OpenAfterExport: c.openAfterExport,
	}
	return result, true
}

func createSummary() core.Summary {
	// TODO: fill in the data from the contest definition, the contest settings, and the score counter
	return core.Summary{}
}

func (c *Controller) SetOpenAfterExport(open bool) {
	c.openAfterExport = open
}
