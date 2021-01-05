package app

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/ftl/hamradio/cwclient"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/callinfo"
	"github.com/ftl/hellocontest/core/cfg"
	"github.com/ftl/hellocontest/core/dxcc"
	"github.com/ftl/hellocontest/core/entry"
	"github.com/ftl/hellocontest/core/export/adif"
	"github.com/ftl/hellocontest/core/export/cabrillo"
	"github.com/ftl/hellocontest/core/export/csv"
	"github.com/ftl/hellocontest/core/hamlib"
	"github.com/ftl/hellocontest/core/keyer"
	"github.com/ftl/hellocontest/core/logbook"
	"github.com/ftl/hellocontest/core/rate"
	"github.com/ftl/hellocontest/core/score"
	"github.com/ftl/hellocontest/core/scp"
	"github.com/ftl/hellocontest/core/settings"
	"github.com/ftl/hellocontest/core/store"
	"github.com/ftl/hellocontest/core/workmode"
)

// NewController returns a new instance of the AppController interface.
func NewController(version string, clock core.Clock, quitter Quitter, asyncRunner core.AsyncRunner, configuration Configuration) *Controller {
	return &Controller{
		version:       version,
		clock:         clock,
		quitter:       quitter,
		asyncRunner:   asyncRunner,
		configuration: configuration,
	}
}

type Controller struct {
	view View

	filename string

	version       string
	clock         core.Clock
	configuration Configuration
	quitter       Quitter
	asyncRunner   core.AsyncRunner
	store         *store.FileStore
	cwclient      *cwclient.Client
	hamlibClient  *hamlib.Client
	dxccFinder    *dxcc.Finder
	scpFinder     *scp.Finder

	Logbook       *logbook.Logbook
	QSOList       *logbook.QSOList
	Entry         *entry.Controller
	Workmode      *workmode.Controller
	Keyer         *keyer.Keyer
	Callinfo      *callinfo.Callinfo
	Score         *score.Counter
	Rate          *rate.Counter
	ServiceStatus *ServiceStatus
	Settings      *settings.Settings

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

// Configuration provides read access to the configuration data.
type Configuration interface {
	Station() core.Station
	Contest() core.Contest
	Keyer() core.Keyer
	Cabrillo() core.Cabrillo

	CabrilloQSOTemplate() string
	KeyerHost() string
	KeyerPort() int
	HamlibAddress() string
}

// Quitter allows to quit the application. This interface is used to call the actual application framework to quit.
type Quitter interface {
	Quit()
}

func (c *Controller) SetView(view View) {
	c.view = view
	c.view.ShowFilename(c.filename)
}

func (c *Controller) Startup() {
	c.Settings = settings.New(
		c.OpenDefaultConfigurationFile,
		score.MatchXchange,
		c.configuration.Station(),
		c.configuration.Contest(),
	)

	c.ServiceStatus = newServiceStatus()

	c.dxccFinder = dxcc.New()
	c.scpFinder = scp.New()

	hamlibAddress := c.configuration.HamlibAddress()
	c.hamlibClient = hamlib.New(hamlibAddress)
	c.hamlibClient.Notify(c.ServiceStatus)
	if hamlibAddress != "" {
		c.hamlibClient.KeepOpen()
	}

	c.QSOList = logbook.NewQSOList(c.Settings)
	c.QSOList.Notify(logbook.QSOFillerFunc(c.fillQSO))
	c.Entry = entry.NewController(
		c.Settings,
		c.clock,
		c.QSOList,
		c.asyncRunner,
	)
	c.QSOList.Notify(c.Entry)

	c.Entry.SetVFO(c.hamlibClient)
	c.hamlibClient.SetVFOController(c.Entry)

	c.cwclient, _ = cwclient.New(c.configuration.KeyerHost(), c.configuration.KeyerPort())
	c.Keyer = keyer.New(c.Settings, c.cwclient, c.configuration.Keyer().WPM)
	c.Keyer.SetPatterns(c.configuration.Keyer().SPMacros)
	c.Keyer.SetValues(c.Entry.CurrentValues)
	c.Keyer.Notify(c.ServiceStatus)
	c.Entry.SetKeyer(c.Keyer)

	c.Workmode = workmode.NewController(c.configuration.Keyer().SPMacros, c.configuration.Keyer().RunMacros)
	c.Workmode.SetKeyer(c.Keyer)

	c.Score = score.NewCounter(c.Settings, c.dxccFinder)
	c.QSOList.Notify(logbook.QSOsClearedListenerFunc(c.Score.Clear))
	c.QSOList.Notify(logbook.QSOAddedListenerFunc(c.Score.Add))
	c.QSOList.Notify(logbook.QSOUpdatedListenerFunc(func(_ int, o, n core.QSO) { c.Score.Update(o, n) }))

	c.Rate = rate.NewCounter(c.asyncRunner)
	c.QSOList.Notify(logbook.QSOsClearedListenerFunc(c.Rate.Clear))
	c.QSOList.Notify(logbook.QSOAddedListenerFunc(c.Rate.Add))
	c.QSOList.Notify(logbook.QSOUpdatedListenerFunc(func(_ int, o, n core.QSO) { c.Rate.Update(o, n) }))

	c.Callinfo = callinfo.New(c.dxccFinder, c.scpFinder, c.QSOList, c.Score)
	c.Entry.SetCallinfo(c.Callinfo)

	c.Settings.Notify(c.Entry)
	c.Settings.Notify(c.Keyer)
	c.Settings.Notify(c.QSOList)
	c.Settings.Notify(c.Score)
	c.Settings.Notify(settings.SettingsListenerFunc(func(s core.Settings) {
		if !c.dxccFinder.Available() {
			return
		}
		if !c.QSOList.Valid() || !c.Score.Valid() {
			c.Refresh()
		}
	}))

	c.dxccFinder.WhenAvailable(func() {
		c.asyncRunner(func() {
			if !c.QSOList.Valid() {
				c.QSOList.ContestChanged(c.Settings.Contest())
			}
			if !c.Score.Valid() {
				c.Score.StationChanged(c.Settings.Station())
				c.Score.ContestChanged(c.Settings.Contest())
			}
			c.Refresh()
			c.ServiceStatus.StatusChanged(core.DXCCService, true)
		})
	})
	c.scpFinder.WhenAvailable(func() {
		c.asyncRunner(func() {
			c.ServiceStatus.StatusChanged(core.SCPService, true)
		})
	})

	c.Entry.StartAutoRefresh()
	c.Rate.StartAutoRefresh()

	err := c.openCurrentLog()
	if err != nil {
		c.Quit()
	}
}

func (c *Controller) openCurrentLog() error {
	filename := "current.log"
	store := store.NewFileStore(filename)
	if !store.Exists() {
		err := store.Clear()
		if err != nil {
			log.Printf("Cannot create %s: %v", filepath.Base(filename), err)
			return err
		}
		err = store.WriteStation(c.Settings.Station())
		if err != nil {
			log.Printf("Cannot write station settings to %s: %v", filepath.Base(filename), err)
			return err
		}
		err = store.WriteContest(c.Settings.Contest())
		if err != nil {
			log.Printf("Cannot write contest settings to %s: %v", filepath.Base(filename), err)
			return err
		}
	}

	var newLogbook *logbook.Logbook
	qsos, station, contest, err := store.ReadAll()
	if err != nil {
		log.Printf("Cannot load %s: %v", filepath.Base(filename), err)
		newLogbook = logbook.New(c.clock)
	} else {
		c.Settings.SetWriter(store)
		if station != nil {
			c.Settings.SetStation(*station)
		}
		if contest != nil {
			c.Settings.SetContest(*contest)
		}
		newLogbook = logbook.Load(c.clock, qsos)
	}
	c.changeLogbook(filename, store, newLogbook)
	return nil
}

func (c *Controller) fillQSO(qso *core.QSO) {
	if entity, found := c.dxccFinder.Find(qso.Callsign.String()); found {
		qso.DXCC = entity
	}
	qso.Points, _ = c.Score.Value(qso.Callsign, qso.DXCC, qso.Band, qso.Mode, qso.TheirXchange)
}

func (c *Controller) changeLogbook(filename string, store *store.FileStore, logbook *logbook.Logbook) {
	c.QSOList.Clear()

	c.filename = filename
	c.store = store
	c.Logbook = logbook
	c.Logbook.SetWriter(c.store)
	c.Logbook.OnRowAdded(c.QSOList.Put)
	c.Entry.SetLogbook(c.Logbook)

	if c.view != nil {
		c.view.ShowFilename(c.filename)
	}
	if c.OnLogbookChanged != nil {
		c.OnLogbookChanged()
	}

	c.hamlibClient.Refresh()
	c.Entry.Clear()
}

func (c *Controller) Shutdown() {
	c.cwclient.Disconnect()
}

func (c *Controller) About() {
	c.view.ShowInfoDialog("Hello Contest\n\nVersion %s\n\nThis software is published under the MIT License.\n(c) Florian Thienel/DL3NEY", c.version)
}

func (c *Controller) OpenSettings() {
	c.Settings.Show()
}

func (c *Controller) OpenDefaultConfigurationFile() {
	c.openTextFile(cfg.AbsoluteFilename())
}

func (c *Controller) openTextFile(filename string) {
	cmd := exec.Command("xdg-open", filename)
	err := cmd.Run()
	if err != nil {
		c.view.ShowErrorDialog("Cannot open the settings file: %v", err)
	}
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
	store := store.NewFileStore(filename)
	err = store.Clear()
	if err != nil {
		c.view.ShowErrorDialog("Cannot create %s: %v", filepath.Base(filename), err)
		return
	}

	c.Settings.SetWriter(store)
	c.changeLogbook(filename, store, logbook.New(c.clock))
	c.Refresh()

	c.OpenSettings()
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

	store := store.NewFileStore(filename)
	qsos, station, contest, err := store.ReadAll()
	if err != nil {
		c.view.ShowErrorDialog("Cannot open %s: %v", filepath.Base(filename), err)
		return
	}

	c.Settings.SetWriter(store)
	if station != nil {
		c.Settings.SetStation(*station)
	}
	if contest != nil {
		c.Settings.SetContest(*contest)
	}
	log := logbook.Load(c.clock, qsos)
	c.changeLogbook(filename, store, log)
	c.Refresh()
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

	store := store.NewFileStore(filename)
	err = store.Clear()
	if err != nil {
		c.view.ShowErrorDialog("Cannot create %s: %v", filepath.Base(filename), err)
		return
	}
	err = store.WriteStation(c.Settings.Station())
	if err != nil {
		c.view.ShowErrorDialog("Cannot save as %s: %v", filepath.Base(filename), err)
		return
	}
	err = store.WriteContest(c.Settings.Contest())
	if err != nil {
		c.view.ShowErrorDialog("Cannot save as %s: %v", filepath.Base(filename), err)
		return
	}
	err = c.Logbook.WriteAll(store)
	if err != nil {
		c.view.ShowErrorDialog("Cannot save as %s: %v", filepath.Base(filename), err)
		return
	}

	c.filename = filename
	c.store = store
	c.Settings.SetWriter(store)
	c.Logbook.SetWriter(store)

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
		c.Settings,
		c.Score.Result(),
		c.QSOList.All()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export Cabrillo to %s: %v", filename, err)
		return
	}
	c.openTextFile(filename)
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
	err = adif.Export(file, c.QSOList.All()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export ADIF to %s: %v", filename, err)
		return
	}
}

func (c *Controller) ExportCSV() {
	filename, ok, err := c.view.SelectSaveFile("Export CSV File", "*.csv")
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
	err = csv.Export(
		file,
		c.Settings.Station().Callsign,
		c.QSOList.All()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export Cabrillo to %s: %v", filename, err)
		return
	}
}

func (c *Controller) ShowCallinfo() {
	c.Callinfo.Show()
	c.view.BringToFront()
}

func (c *Controller) ShowScore() {
	c.Score.Show()
	c.view.BringToFront()
}

func (c *Controller) ShowRate() {
	c.Rate.Show()
	c.view.BringToFront()
}

func (c *Controller) Refresh() {
	c.QSOList.Clear()
	c.Logbook.ReplayAll()
	c.Entry.Clear()
}

func (c *Controller) ClearEntryFields() {
	c.Entry.Clear()
}

func (c *Controller) GotoEntryFields() {
	c.Entry.Activate()
}

func (c *Controller) EditLastQSO() {
	c.Entry.EditLastQSO()
}

func (c *Controller) LogQSO() {
	c.Entry.Log()
}

func (c *Controller) SwitchToSPWorkmode() {
	c.Workmode.SetWorkmode(core.SearchPounce)
}

func (c *Controller) SwitchToRunWorkmode() {
	c.Workmode.SetWorkmode(core.Run)
}
