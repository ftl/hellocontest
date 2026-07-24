package clock

import (
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/ticker"
)

// View represents the visual part that shows the current UTC time.
type View interface {
	Show()
	SetTime(string)
}

type nullView struct{}

func (n *nullView) Show()          {}
func (n *nullView) SetTime(string) {}

// NewController returns a new clock view controller driven by the given clock.
func NewController(clock core.Clock, asyncRunner core.AsyncRunner) *Controller {
	result := &Controller{
		clock:       clock,
		asyncRunner: asyncRunner,
		view:        &nullView{},
	}
	result.ticker = ticker.New(clock, result.refresh)
	return result
}

type Controller struct {
	clock       core.Clock
	asyncRunner core.AsyncRunner
	ticker      *ticker.Ticker
	view        View
}

func (c *Controller) SetView(view View) {
	if view == nil {
		c.view = &nullView{}
		return
	}
	c.view = view
	c.refresh()
}

// StartAutoRefresh starts the per-second refresh of the displayed time.
func (c *Controller) StartAutoRefresh() {
	c.ticker.Start()
}

func (c *Controller) Show() {
	c.view.Show()
}

func (c *Controller) refresh() {
	c.asyncRunner(func() {
		c.view.SetTime(c.clock.Now().UTC().Format("15:04:05"))
	})
}
