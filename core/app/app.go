package app

import (
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/ftl/hamradio/cwclient"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/callinfo"
	"github.com/ftl/hellocontest/core/entry"
	"github.com/ftl/hellocontest/core/export/adif"
	"github.com/ftl/hellocontest/core/export/cabrillo"
	"github.com/ftl/hellocontest/core/keyer"
	"github.com/ftl/hellocontest/core/logbook"
	"github.com/ftl/hellocontest/core/store"
	"github.com/ftl/hellocontest/core/workmode"
)

// NewController returns a new instance of the AppController interface.
func NewController(clock core.Clock, quitter core.Quitter, configuration core.Configuration) *Controller {
	return &Controller{
		clock:         clock,
		quitter:       quitter,
		configuration: configuration,
	}
}

type Controller struct {
	view View

	filename string

	clock         core.Clock
	configuration core.Configuration
	store         core.Store
	cwclient      core.CWClient
	quitter       core.Quitter

	Logbook  *logbook.Logbook
	Entry    *entry.Controller
	Workmode *workmode.Controller
	Keyer    *keyer.Keyer
	Callinfo *callinfo.Callinfo

	OnLogbookChanged func()
}

// View defines the visual functionality of the main application window.
type View interface {
	BringToFront()
	ShowFilename(string)
	SelectOpenFile(string, ...string) (string, bool, error)
	SelectSaveFile(string, ...string) (string, bool, error)
	ShowInfoDialog(string, ...interface{})
	ShowErrorDialog(string, ...interface{})
}

func (c *Controller) SetView(view View) {
	c.view = view
	c.view.ShowFilename(c.filename)
}

func (c *Controller) Startup() {
	var err error
	filename := "current.log"

	store := store.New(filename)
	newLogbook, err := logbook.Load(c.clock, store)
	if err != nil {
		log.Println(err)
		newLogbook = logbook.New(c.clock)
	}

	c.cwclient, _ = cwclient.New(c.configuration.KeyerHost(), c.configuration.KeyerPort())

	c.Keyer = keyer.New(c.cwclient, c.configuration.MyCall(), c.configuration.KeyerWPM())
	c.Keyer.SetPatterns(c.configuration.KeyerSPPatterns())

	c.Workmode = workmode.NewController(c.configuration.KeyerSPPatterns(), c.configuration.KeyerRunPatterns())
	c.Workmode.SetKeyer(c.Keyer)

	c.Callinfo = callinfo.New(setupDXCC(), setupSupercheck())

	c.changeLogbook(filename, store, newLogbook)
}

func setupDXCC() *dxcc.Prefixes {
	localFilename, err := dxcc.LocalFilename()
	if err != nil {
		log.Print(err)
		return nil
	}
	updated, err := dxcc.Update(dxcc.DefaultURL, localFilename)
	if err != nil {
		log.Printf("update of local copy of DXCC prefixes failed: %v", err)
	}
	if updated {
		log.Printf("updated local copy of DXCC prefixes: %v", localFilename)
	}

	result, err := dxcc.LoadLocal(localFilename)
	if err != nil {
		log.Printf("cannot load DXCC prefixes: %v", err)
		return nil
	}
	return result
}

func setupSupercheck() *scp.Database {
	localFilename, err := scp.LocalFilename()
	if err != nil {
		log.Print(err)
		return nil
	}
	updated, err := scp.Update(scp.DefaultURL, localFilename)
	if err != nil {
		log.Printf("update of local copy of Supercheck database failed: %v", err)
	}
	if updated {
		log.Printf("updated local copy of Supercheck database: %v", localFilename)
	}

	result, err := scp.LoadLocal(localFilename)
	if err != nil {
		log.Printf("cannot load Supercheck database: %v", err)
		return nil
	}
	return result
}

func (c *Controller) Shutdown() {
	c.cwclient.Disconnect()
}

func (c *Controller) Quit() {
	c.quitter.Quit()
}

func (c *Controller) New() {
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

	c.changeLogbook(filename, store, logbook.New(c.clock))
}

func (c *Controller) Open() {
	filename, ok, err := c.view.SelectOpenFile("Open Logfile", "*.log")
	if !ok {
		return
	}
	if err != nil {
		c.view.ShowErrorDialog("Cannot select a file: %v", err)
		return
	}

	store := store.New(filename)
	log, err := logbook.Load(c.clock, store)
	if err != nil {
		c.view.ShowErrorDialog("Cannot open %s: %v", filepath.Base(filename), err)
		return
	}

	c.changeLogbook(filename, store, log)
}

func (c *Controller) changeLogbook(filename string, store core.Store, logbook *logbook.Logbook) {
	c.filename = filename
	c.store = store

	if c.Logbook != nil {
		c.Logbook.SetView(nil)
		c.Logbook.ClearRowSelectedListeners()
	}
	if c.Entry != nil {
		c.Entry.SetView(nil)
	}

	c.Logbook = logbook
	c.Logbook.OnRowAdded(c.store.Write)
	c.Entry = entry.NewController(
		c.clock,
		c.Logbook,
		c.configuration.EnterTheirNumber(),
		c.configuration.EnterTheirXchange(),
		c.configuration.AllowMultiBand(),
		c.configuration.AllowMultiMode(),
	)
	c.Logbook.OnRowSelected(c.Entry.QSOSelected)

	c.Entry.SetKeyer(c.Keyer)
	c.Entry.SetCallinfo(c.Callinfo)

	c.Keyer.SetValues(c.Entry.CurrentValues)
	c.Callinfo.SetDupChecker(c.Entry)

	if c.view != nil {
		c.view.ShowFilename(c.filename)
	}
	if c.OnLogbookChanged != nil {
		c.OnLogbookChanged()
	}
}

func (c *Controller) SaveAs() {
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
	err = c.Logbook.WriteAll(store)
	if err != nil {
		c.view.ShowErrorDialog("Cannot save as %s: %v", filepath.Base(filename), err)
		return
	}

	c.Logbook.ClearRowAddedListeners()
	c.filename = filename
	c.store = store
	c.Logbook.OnRowAdded(c.store.Write)

	c.view.ShowFilename(c.filename)
}

func (c *Controller) ExportCabrillo() {
	filename, ok, err := c.view.SelectSaveFile("Export Cabrillo File", "*.cabrillo")
	if !ok {
		return
	}
	if err != nil {
		c.view.ShowErrorDialog("Cannot select a file: %v", err)
		return
	}

	template, err := template.New("").Parse(c.configuration.CabrilloQSOTemplate())
	if err != nil {
		c.view.ShowErrorDialog("Cannot parse the QSO template: %v", err)
		return
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		c.view.ShowErrorDialog("Cannot open file %s: %v", filename, err)
		return
	}
	defer file.Close()
	err = cabrillo.Export(
		file,
		template,
		c.configuration.MyCall(),
		c.Logbook.UniqueQsosOrderedByMyNumber()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export Cabrillo to %s: %v", filename, err)
		return
	}
}

func (c *Controller) ExportADIF() {
	filename, ok, err := c.view.SelectSaveFile("Export ADIF File", "*.adif")
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
	err = adif.Export(file, c.Logbook.UniqueQsosOrderedByMyNumber()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export ADIF to %s: %v", filename, err)
		return
	}
}

func (c *Controller) ShowCallinfo() {
	c.Callinfo.Show()
	c.view.BringToFront()
}
