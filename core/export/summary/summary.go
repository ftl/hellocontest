package summary

import (
	"fmt"
	"io"

	"github.com/ftl/hellocontest/core"
)

type View interface {
	// TODO add view methods
}

type Result struct {
	Summary         core.Summary
	OpenAfterExport bool
}

type Controller struct {
	view View
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
	return Result{Summary: core.Summary{}}, false
}

func Export(w io.Writer, summary core.Summary) error {
	return fmt.Errorf("summary.Export is not yet implemented")
}
