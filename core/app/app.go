package app

import (
	logger "log"
	"os"
	"path/filepath"

	"github.com/ftl/hamradio/cwclient"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/entry"
	"github.com/ftl/hellocontest/core/export/cabrillo"
	"github.com/ftl/hellocontest/core/keyer"
	"github.com/ftl/hellocontest/core/log"
	"github.com/ftl/hellocontest/core/store"
)

// NewController returns a new instance of the AppController interface.
func NewController(clock core.Clock, configuration core.Configuration) core.AppController {
	return &controller{
		clock:         clock,
		configuration: configuration,
	}
}

type controller struct {
	view core.AppView

	filename string

	clock         core.Clock
	configuration core.Configuration
	log           core.Log
	store         core.Store
	cwclient      core.CWClient
	entry         core.EntryController
	keyer         core.KeyerController

	logView   core.LogView
	entryView core.EntryView
	keyerView core.KeyerView
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
		logger.Println(err)
		c.log = log.New(c.clock)
	}
	c.log.OnRowAdded(c.store.Write)
	c.cwclient, _ = cwclient.New(c.configuration.KeyerHost(), c.configuration.KeyerPort())

	c.entry = entry.NewController(
		c.clock,
		c.log,
		c.configuration.EnterTheirNumber(),
		c.configuration.EnterTheirXchange(),
	)
	c.keyer = keyer.NewController(c.cwclient, c.configuration.MyCall(), c.entry.CurrentValues)
	c.keyer.SetPatterns(c.configuration.KeyerSPPatterns())
}

func (c *controller) Shutdown() {
	c.cwclient.Disconnect()
}

func (c *controller) SetLogView(view core.LogView) {
	c.logView = view
	c.log.SetView(c.logView)
}

func (c *controller) SetEntryView(view core.EntryView) {
	c.entryView = view
	c.entry.SetView(c.entryView)
}

func (c *controller) SetKeyerView(view core.KeyerView) {
	c.keyerView = view
	c.keyer.SetView(c.keyerView)
}

func (c *controller) New() {
	filename, ok, err := c.view.SelectSaveFile("New Logfile", "*.log")
	if !ok {
		return
	}
	if err != nil {
		c.view.ShowErrorDialog("Cannot select a file: %v", err)
		return
	}
	store := store.New(filename)
	err = store.Clear()
	if err != nil {
		c.view.ShowErrorDialog("Cannot create %s: %v", filepath.Base(filename), err)
		return
	}

	c.filename = filename
	c.store = store
	c.log = log.New(c.clock)
	c.log.OnRowAdded(c.store.Write)
	c.entry = entry.NewController(
		c.clock,
		c.log,
		c.configuration.EnterTheirNumber(),
		c.configuration.EnterTheirXchange(),
	)

	c.view.ShowFilename(c.filename)
	c.log.SetView(c.logView)
	c.entry.SetView(c.entryView)
}

func (c *controller) Open() {
	filename, ok, err := c.view.SelectOpenFile("Open Logfile", "*.log")
	if !ok {
		return
	}
	if err != nil {
		c.view.ShowErrorDialog("Cannot select a file: %v", err)
		return
	}

	store := store.New(filename)
	log, err := log.Load(c.clock, store)
	if err != nil {
		c.view.ShowErrorDialog("Cannot open %s: %v", filepath.Base(filename), err)
		return
	}

	c.filename = filename
	c.store = store
	c.log = log
	c.log.OnRowAdded(c.store.Write)
	c.entry = entry.NewController(
		c.clock,
		c.log,
		c.configuration.EnterTheirNumber(),
		c.configuration.EnterTheirXchange(),
	)

	c.view.ShowFilename(c.filename)
	c.log.SetView(c.logView)
	c.entry.SetView(c.entryView)
}

func (c *controller) SaveAs() {
	filename, ok, err := c.view.SelectSaveFile("Save Logfile As", "*.log")
	if !ok {
		return
	}
	if err != nil {
		c.view.ShowErrorDialog("Cannot select a file: %v", err)
		return
	}

	store := store.New(filename)
	err = store.Clear()
	if err != nil {
		c.view.ShowErrorDialog("Cannot create %s: %v", filepath.Base(filename), err)
		return
	}
	err = c.log.WriteAll(store)
	if err != nil {
		c.view.ShowErrorDialog("Cannot save as %s: %v", filepath.Base(filename), err)
		return
	}

	c.log.ClearRowAddedListeners()
	c.filename = filename
	c.store = store
	c.log.OnRowAdded(c.store.Write)

	c.view.ShowFilename(c.filename)
}

func (c *controller) ExportCabrillo() {
	filename, ok, err := c.view.SelectSaveFile("Export Cabrillo File", "*.cabrillo")
	if !ok {
		return
	}
	if err != nil {
		c.view.ShowErrorDialog("Cannot select a file: %v", err)
		return
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		c.view.ShowErrorDialog("Cannot open file %s: %v", filename, err)
		return
	}
	defer file.Close()
	err = cabrillo.Export(file,
		c.configuration.MyCall(),
		c.configuration.MyExchanger(),
		c.configuration.TheirExchanger(),
		c.log.UniqueQsosOrderedByMyNumber()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export Cabrillo to %s: %v", filename, err)
		return
	}
}
