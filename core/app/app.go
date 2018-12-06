package app

import (
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/entry"
	"github.com/ftl/hellocontest/core/log"
	"github.com/ftl/hellocontest/core/store"
)

// NewController returns a new instance of the AppController interface.
func NewController(clock core.Clock) core.AppController {
	return &controller{
		clock: clock,
	}
}

type controller struct {
	view core.AppView

	filename string

	clock core.Clock
	log   core.Log
	store core.Store
	entry core.EntryController

	logView   core.LogView
	entryView core.EntryView
}

func (c *controller) SetView(view core.AppView) {
	c.view = view
	c.view.SetAppController(c)
	c.view.ShowFilename(c.filename)
}

func (c *controller) Startup() {
	var err error
	c.filename = "current.log"

	c.store = store.New(c.filename)
	c.log, err = log.Load(c.clock, c.store)
	if err != nil {
		c.log = log.New(c.clock)
	}
	c.log.OnRowAdded(c.store.Write)
	c.entry = entry.NewController(c.clock, c.log)
}

func (c *controller) SetLogView(view core.LogView) {
	c.logView = view
	c.log.SetView(c.logView)
}

func (c *controller) SetEntryView(view core.EntryView) {
	c.entryView = view
	c.entry.SetView(c.entryView)
}

func (c *controller) New() {
	filename, ok, err := c.view.SelectSaveFile("New Logfile", "*.log")
	if !ok {
		return
	}
	if err != nil {
		c.view.ShowErrorMessage("Cannot select a file: %v", err)
		return
	}

	c.filename = filename
	c.store = store.New(c.filename)
	c.log = log.New(c.clock)
	c.log.OnRowAdded(c.store.Write)
	c.entry = entry.NewController(c.clock, c.log)

	c.view.ShowFilename(c.filename)
	c.log.SetView(c.logView)
	c.entry.SetView(c.entryView)
}

func (c *controller) Open() {
	c.view.ShowErrorMessage("Opening a log is not yet implemented.")
}

func (c *controller) SaveAs() {
	c.view.ShowErrorMessage("Saving a log under another name is not yet implemented.")
}
