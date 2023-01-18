package app

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ftl/hamradio/cwclient"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/callhistory"
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
	"github.com/ftl/hellocontest/core/tci"
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

	version           string
	clock             core.Clock
	configuration     Configuration
	quitter           Quitter
	asyncRunner       core.AsyncRunner
	store             *store.FileStore
	tciClient         *tci.Client
	cwclient          *cwclient.Client
	hamlibClient      *hamlib.Client
	dxccFinder        *dxcc.Finder
	scpFinder         *scp.Finder
	callHistoryFinder *callhistory.Finder

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
	SelectSaveFile(string, string, ...string) (string, bool, error)
	ShowInfoDialog(string, ...interface{})
	ShowErrorDialog(string, ...interface{})
}

// Configuration provides read access to the configuration data.
type Configuration interface {
	Station() core.Station
	Contest() core.Contest
	Keyer() core.Keyer

	KeyerType() core.KeyerType
	KeyerHost() string
	KeyerPort() int
	HamlibAddress() string
	TCIAddress() string
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
		c.openWithExternalApplication,
		c.configuration.Station(),
		c.configuration.Contest(),
	)

	c.ServiceStatus = newServiceStatus(c.asyncRunner)

	c.dxccFinder = dxcc.New()
	c.scpFinder = scp.New()
	c.callHistoryFinder = callhistory.New(c.Settings, c.ServiceStatus.StatusChanged)

	c.Score = score.NewCounter(c.Settings, c.dxccFinder)
	c.QSOList = logbook.NewQSOList(c.Settings, c.Score)
	c.Entry = entry.NewController(
		c.Settings,
		c.clock,
		c.QSOList,
		c.asyncRunner,
	)
	c.QSOList.Notify(c.Entry)

	tciAddress := c.configuration.TCIAddress()
	hamlibAddress := c.configuration.HamlibAddress()
	var keyerCWClient keyer.CWClient
	if tciAddress != "" {
		tciClient, err := tci.NewClient(tciAddress)
		if err != nil {
			log.Printf("cannot open TCI connection: %v", err)
		} else {
			c.tciClient = tciClient
			c.tciClient.Notify(c.ServiceStatus)
			c.Entry.SetVFO(c.tciClient)
			c.tciClient.SetVFOController(c.Entry)
			keyerCWClient = c.tciClient
		}
	} else if hamlibAddress != "" {
		c.hamlibClient = hamlib.New(hamlibAddress)
		c.hamlibClient.Notify(c.ServiceStatus)
		c.hamlibClient.KeepOpen()
		c.Entry.SetVFO(c.hamlibClient)
		c.hamlibClient.SetVFOController(c.Entry)
		if c.configuration.KeyerType() == core.KeyerTypeHamlib {
			keyerCWClient = c.hamlibClient
			log.Println("using the hamlib client for CW")
		}
	}

	if keyerCWClient == nil || c.configuration.KeyerType() == core.KeyerTypeCWDaemon {
		c.cwclient, _ = cwclient.New(c.configuration.KeyerHost(), c.configuration.KeyerPort())
		keyerCWClient = c.cwclient
		log.Println("using the CWDaemon for CW")
	}

	c.Workmode = workmode.NewController()

	c.Keyer = keyer.New(c.Settings, keyerCWClient, c.configuration.Keyer(), c.Workmode.Workmode())
	c.Keyer.SetValues(c.Entry.CurrentValues)
	c.Keyer.Notify(c.ServiceStatus)
	c.Workmode.Notify(c.Keyer)
	c.Entry.SetKeyer(c.Keyer)

	c.Rate = rate.NewCounter(c.asyncRunner)
	c.QSOList.Notify(logbook.QSOsClearedListenerFunc(c.Rate.Clear))
	c.QSOList.Notify(logbook.QSOAddedListenerFunc(c.Rate.Add))

	c.Callinfo = callinfo.New(c.dxccFinder, c.scpFinder, c.callHistoryFinder, c.QSOList, c.Score, c.Entry)
	c.Entry.SetCallinfo(c.Callinfo)

	c.Settings.Notify(c.Entry)
	c.Settings.Notify(c.Workmode)
	c.Settings.Notify(c.Keyer)
	c.Settings.Notify(c.QSOList)
	c.Settings.Notify(c.Score)
	c.Settings.Notify(c.Rate)
	c.Settings.Notify(c.Callinfo)
	c.Settings.Notify(c.callHistoryFinder)
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
			if !c.Score.Valid() {
				c.Score.StationChanged(c.Settings.Station())
				c.Score.ContestChanged(c.Settings.Contest())
			}
			if !c.QSOList.Valid() {
				c.QSOList.ContestChanged(c.Settings.Contest())
			}
			c.Refresh()
		})
		c.ServiceStatus.StatusChanged(core.DXCCService, true)
	})
	c.scpFinder.WhenAvailable(func() {
		c.ServiceStatus.StatusChanged(core.SCPService, true)
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
		err = store.WriteKeyer(c.Keyer.KeyerSettings())
		if err != nil {
			log.Printf("Cannot write contest settings to %s: %v", filepath.Base(filename), err)
			return err
		}
	}

	var newLogbook *logbook.Logbook
	qsos, station, contest, keyer, err := store.ReadAll()
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
		c.Keyer.SetWriter(store)
		if keyer != nil {
			c.Keyer.SetKeyer(*keyer)
		}
		newLogbook = logbook.Load(c.clock, qsos)
	}
	c.changeLogbook(filename, store, newLogbook)
	return nil
}

func (c *Controller) changeLogbook(filename string, store *store.FileStore, logbook *logbook.Logbook) {
	c.QSOList.Clear()

	c.filename = filename
	c.store = store
	c.Logbook = logbook
	c.Logbook.SetWriter(c.store)
	c.Logbook.OnRowAdded(c.QSOList.Put)
	c.Logbook.OnRowAdded(c.Workmode.RowAdded)
	c.Entry.SetLogbook(c.Logbook)

	if c.view != nil {
		c.view.ShowFilename(c.filename)
	}
	if c.OnLogbookChanged != nil {
		c.OnLogbookChanged()
	}

	if c.tciClient != nil {
		c.tciClient.Refresh()
	}
	if c.hamlibClient != nil {
		c.hamlibClient.Refresh()
	}
	c.Entry.Clear()
}

func (c *Controller) Shutdown() {
	if c.tciClient != nil {
		c.tciClient.Disconnect()
	}
	if c.hamlibClient != nil {
		c.hamlibClient.Disconnect()
	}
	if c.cwclient != nil {
		c.cwclient.Disconnect()
	}
}

func (c *Controller) About() {
	c.view.ShowInfoDialog("Hello Contest\n\nVersion %s\n\nThis software is published under the MIT License.\n(c) Florian Thienel/DL3NEY", c.version)
}

func (c *Controller) OpenSettings() {
	c.Settings.Show()
}

func (c *Controller) OpenDefaultConfigurationFile() {
	c.openWithExternalApplication(cfg.AbsoluteFilename())
}

func (c *Controller) openWithExternalApplication(filename string) {
	cmd := exec.Command("xdg-open", filename)
	err := cmd.Run()
	if err != nil {
		c.view.ShowErrorDialog("Cannot open the file %s with its external application: %v", filename, err)
	}
}

func (c *Controller) Quit() {
	c.quitter.Quit()
}

func (c *Controller) proposeFilename() string {
	result := strings.Join([]string{c.Settings.Contest().Name, c.Settings.Station().Callsign.BaseCall}, " ")
	result = strings.TrimSpace(result)
	result = strings.ToUpper(result)
	result = strings.ReplaceAll(result, " ", "_")
	return result
}

func (c *Controller) New() {
	proposedName := c.proposeFilename() + ".log"
	filename, ok, err := c.view.SelectSaveFile("New Logfile", proposedName, "*.log")
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

	c.Settings.Reset()

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
	err = store.WriteKeyer(c.Keyer.KeyerSettings())
	if err != nil {
		c.view.ShowErrorDialog("Cannot save as %s: %v", filepath.Base(filename), err)
		return
	}

	c.Settings.SetWriter(store)
	c.Keyer.SetWriter(store)
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
	qsos, station, contest, keyer, err := store.ReadAll()
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
	c.Keyer.SetWriter(store)
	if keyer != nil {
		c.Keyer.SetKeyer(*keyer)
	}
	log := logbook.Load(c.clock, qsos)
	c.changeLogbook(filename, store, log)
	c.Refresh()
}

func (c *Controller) SaveAs() {
	proposedName := c.proposeFilename() + ".log"
	filename, ok, err := c.view.SelectSaveFile("Save Logfile As", proposedName, "*.log")
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
	err = store.WriteKeyer(c.Keyer.KeyerSettings())
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
	proposedName := c.proposeFilename() + ".cabrillo"
	filename, ok, err := c.view.SelectSaveFile("Export Cabrillo File", proposedName, "*.cabrillo")
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

	err = cabrillo.Export(
		file,
		c.Settings,
		c.Score.Result(),
		c.QSOList.All()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export Cabrillo to %s: %v", filename, err)
		return
	}
	c.openWithExternalApplication(filename)
}

func (c *Controller) ExportADIF() {
	proposedName := c.proposeFilename() + ".adif"
	filename, ok, err := c.view.SelectSaveFile("Export ADIF File", proposedName, "*.adif")
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
	proposedName := c.proposeFilename() + ".csv"
	filename, ok, err := c.view.SelectSaveFile("Export CSV File", proposedName, "*.csv")
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

func (c *Controller) ExportCallhistory() {
	proposedName := c.proposeFilename() + ".txt"
	filename, ok, err := c.view.SelectSaveFile("Export Call History File", proposedName, "*.txt")
	if !ok {
		return
	}
	if err != nil {
		c.view.ShowErrorDialog("Cannot select a file: %v", err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		c.view.ShowErrorDialog("Cannot open file %s: %v", filename, err)
	}
	defer file.Close()

	err = callhistory.Export(file, c.Settings.Contest().CallHistoryFieldNames, c.QSOList.All()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export call history to %s: %v", filename, err)
		return
	}
	c.openWithExternalApplication(filename)
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
	c.QSOList.Fill(c.Logbook.All())
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

func (c *Controller) Stop() {
	c.Keyer.Stop()
}

func (c *Controller) DoubleStop() {
	c.Entry.Clear()
}
