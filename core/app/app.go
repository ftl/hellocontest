package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/hamradio/bandplan"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/bandmap"
	"github.com/ftl/hellocontest/core/callhistory"
	"github.com/ftl/hellocontest/core/callinfo"
	"github.com/ftl/hellocontest/core/cfg"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/cluster"
	"github.com/ftl/hellocontest/core/dxcc"
	"github.com/ftl/hellocontest/core/entry"
	"github.com/ftl/hellocontest/core/export/adif"
	"github.com/ftl/hellocontest/core/export/cabrillo"
	"github.com/ftl/hellocontest/core/export/csv"
	"github.com/ftl/hellocontest/core/export/summary"
	"github.com/ftl/hellocontest/core/hamdxmap"
	"github.com/ftl/hellocontest/core/keyer"
	"github.com/ftl/hellocontest/core/logbook"
	"github.com/ftl/hellocontest/core/newcontest"
	"github.com/ftl/hellocontest/core/parrot"
	"github.com/ftl/hellocontest/core/qtc"
	"github.com/ftl/hellocontest/core/radio"
	"github.com/ftl/hellocontest/core/rate"
	"github.com/ftl/hellocontest/core/remote"
	"github.com/ftl/hellocontest/core/score"
	"github.com/ftl/hellocontest/core/scp"
	"github.com/ftl/hellocontest/core/session"
	"github.com/ftl/hellocontest/core/settings"
	"github.com/ftl/hellocontest/core/store"
	"github.com/ftl/hellocontest/core/vfo"
	"github.com/ftl/hellocontest/core/workmode"
)

// NewController returns a new instance of the AppController interface.
func NewController(version string, clock core.Clock, quitter Quitter, asyncRunner core.AsyncRunner, configuration Configuration, sponsors string) *Controller {
	return &Controller{
		version:       version,
		clock:         clock,
		quitter:       quitter,
		asyncRunner:   asyncRunner,
		configuration: configuration,
		sponsors:      sponsors,
	}
}

type Controller struct {
	view View

	filename string

	version       string
	clock         core.Clock
	session       *session.Session
	configuration Configuration
	sponsors      string
	quitter       Quitter
	asyncRunner   core.AsyncRunner
	store         *store.FileStore

	bandplan          bandplan.Bandplan
	dxccFinder        *dxcc.Finder
	scpFinder         *scp.Finder
	callHistoryFinder *callhistory.Finder
	hamDXMap          *hamdxmap.HamDXMap

	VFOs                     []*vfo.VFO
	Logbook                  *logbook.Logbook
	QSOList                  *logbook.QSOList
	QTCList                  *logbook.QTCList
	Entry                    *entry.Controller
	ClockView                *clock.Controller
	ScoreController          *score.Controller
	Workmode                 *workmode.Controller
	Radio                    *radio.Controller
	Keyer                    *keyer.Keyer
	Callinfo                 *callinfo.Callinfo
	Rate                     *rate.Counter
	ServiceStatus            *ServiceStatus
	NewContestController     *newcontest.Controller
	SummaryController        *summary.Controller
	ExportCabrilloController *cabrillo.Controller
	QTCController            *qtc.Controller
	Settings                 *settings.Settings
	Bandmap                  *bandmap.Bandmap
	Clusters                 *cluster.Clusters
	Parrot                   *parrot.Parrot

	remoteServer *remote.Server
}

// View defines the visual functionality of the main application window.
type View interface {
	BringToFront()
	ShowFilename(string)
	SelectOpenFile(title string, dir string, patterns ...string) (string, bool, error)
	SelectSaveFile(title string, dir string, filename string, patterns ...string) (string, bool, error)
	ShowInfoDialog(string, ...any)
	ShowQuestionDialog(string, ...any) bool
	ShowErrorDialog(string, ...any)
}

// Configuration provides read access to the configuration data.
type Configuration interface {
	LogDirectory() string
	HamDXMapPort() int
	Station() core.Station
	Contest() core.Contest
	KeyerSettings() core.KeyerSettings
	KeyerPresets() []core.KeyerPreset
	SpotLifetime() time.Duration
	SpotSources() []core.SpotSource
	Radios() []core.Radio
	Keyers() []core.Keyer
	RemoteServerSettings() core.RemoteServerSettings
}

// Script can be used to automate things
type Script interface {
	Step(ctx context.Context, app *Controller, ui func(func())) bool
}

// Quitter allows to quit the application. This interface is used to call the actual application framework to quit.
type Quitter interface {
	Quit()
}

const (
	wikiURL     = "https://github.com/ftl/hellocontest/wiki"
	sponsorsURL = "https://github.com/sponsors/ftl"
)

func (c *Controller) SetView(view View) {
	c.view = view
	c.view.ShowFilename(c.filename)
}

func (c *Controller) Startup() {
	c.session = session.NewDefaultSession()
	err := c.session.Restore()
	if err != nil {
		log.Printf("Cannot restore session: %v", err)
		c.session = session.NewDefaultSession()
	}

	c.ServiceStatus = newServiceStatus(c.asyncRunner)

	c.callHistoryFinder = callhistory.New(c.ServiceStatus.StatusChanged)
	c.Settings = settings.New(
		c.OpenConfigurationFile,
		c.clock,
		c.openWithExternalApplication,
		c.callHistoryFinder,
		c.configuration.Station(),
		c.configuration.Contest(),
	)
	c.callHistoryFinder.Notify(c.Settings)
	c.NewContestController = newcontest.NewController(c.Settings, c.configuration.LogDirectory())
	c.ExportCabrilloController = cabrillo.NewController()

	c.bandplan = core.BandplanByID(c.configuration.Station().Bandplan)
	c.dxccFinder = dxcc.New()
	c.scpFinder = scp.New()
	c.hamDXMap = hamdxmap.NewHamDXMap(c.configuration.HamDXMapPort())

	c.QSOList = logbook.NewQSOList(c.Settings)
	c.QTCList = logbook.NewQTCList()
	c.ScoreController = score.NewController(c.Settings)

	c.Logbook = logbook.NewLogbook(c.clock, c.Settings, c.dxccFinder)
	c.Logbook.Notify(c.QSOList)
	c.Logbook.Notify(c.QTCList)
	c.Logbook.Notify(c.ScoreController)

	c.Bandmap = bandmap.NewBandmap(c.clock, c.Settings, c.Logbook, c.asyncRunner, bandmap.DefaultUpdatePeriod, c.configuration.SpotLifetime())
	c.Logbook.Notify(c.Bandmap)

	c.Clusters = cluster.NewClusters(c.configuration.SpotSources(), c.Bandmap, c.bandplan, c.dxccFinder, c.clock)
	c.Entry = entry.NewController(
		c.Settings,
		c.clock,
		c.Logbook,
		c.QSOList,
		c.Bandmap,
		c.asyncRunner,
	)
	c.Entry.SetESMEnabled(c.session.ESM())
	c.Entry.Notify(c.session)
	c.Entry.Notify(c.hamDXMap)
	c.Bandmap.Notify(c.Entry)
	c.Logbook.Notify(c.Entry)
	c.QSOList.Notify(c.Entry)

	c.ClockView = clock.NewController(c.clock, c.asyncRunner)

	c.SummaryController = summary.NewController(c.Logbook)

	c.Workmode = workmode.NewController()
	c.Workmode.Notify(c.Entry)
	c.Entry.Notify(c.Workmode)
	c.QSOList.Notify(c.Workmode)

	c.Callinfo = callinfo.New(c.dxccFinder, c.scpFinder, c.callHistoryFinder, c.Logbook, c.Logbook, c.Logbook)
	c.Entry.SetCallinfo(c.Callinfo)
	c.Callinfo.Notify(c.Entry)

	c.Radio = radio.NewController(c.configuration.Radios(), c.configuration.Keyers(), c.bandplan)
	c.Radio.Notify(c.ServiceStatus)
	c.Radio.Notify(c.Entry)
	c.Radio.Notify(c.Callinfo)
	c.Radio.Notify(c.Workmode)
	c.Entry.SetVFOSwitcher(c.Radio)
	c.Bandmap.Notify(c.Radio) // TODO implement Entry... in radio.Controller

	c.VFOs = make([]*vfo.VFO, core.VFOCount)
	for vfoID := range len(c.VFOs) {
		v := vfo.NewVFO(core.VFOID(vfoID), fmt.Sprintf("VFO %d", vfoID+1), c.bandplan, c.Logbook, c.asyncRunner)
		v.SetClient(c.Radio)
		c.VFOs[vfoID] = v
		c.Entry.SetVFO(core.VFOID(vfoID), v)
		c.Logbook.Notify(v)
	}
	c.Bandmap.SetVFO(c.VFOs[core.VFO1])
	c.Workmode.Notify(c.VFOs[core.VFO1])

	c.Radio.SetSendSpotsToTci(c.session.SendSpotsToTci())
	c.Radio.SelectRadio(c.session.Radio1())
	c.Radio.SelectKeyer(c.session.Keyer1())

	c.Keyer = keyer.New(c.Settings, c.Radio, c.configuration.KeyerSettings(), c.Workmode.Workmode(), c.configuration.KeyerPresets())
	c.Keyer.SetVFOSwitcher(c.Radio)
	c.Keyer.SetValues(c.Entry.CurrentValues)
	c.Keyer.Notify(c.ServiceStatus)
	c.Keyer.Notify(c.Entry)
	c.Entry.Notify(c.Keyer)
	c.Workmode.Notify(c.Keyer)
	c.Entry.SetKeyer(c.Keyer)

	remoteSettings := c.configuration.RemoteServerSettings()
	if remoteSettings.Enabled {
		c.remoteServer = remote.NewServer(c, c.Keyer, remoteSettings.Port, c.asyncRunner)
		if err := c.remoteServer.Start(); err != nil {
			log.Printf("remote server failed to start: %v", err)
			c.remoteServer = nil
			c.ServiceStatus.StatusChanged(core.RemoteService, false)
		} else {
			c.ServiceStatus.StatusChanged(core.RemoteService, true)
		}
	} else {
		c.ServiceStatus.StatusChanged(core.RemoteService, false)
	}

	c.QTCController = qtc.NewController(c.clock, c, c.Logbook, c.QTCList, c.Entry, c.Keyer)

	c.Rate = rate.NewCounter(c.clock, c.asyncRunner)
	c.QSOList.Notify(logbook.QSOsClearedListenerFunc(c.Rate.Clear))
	c.QSOList.Notify(logbook.QSOAddedListenerFunc(c.Rate.Add))

	c.Bandmap.SetCallinfo(c.Callinfo)
	c.Bandmap.Notify(c.Callinfo)
	c.Logbook.Notify(c.Callinfo)

	c.Parrot = parrot.New(c.Workmode, c.Keyer, c.asyncRunner)
	c.Keyer.SetParrot(c.Parrot)
	c.Keyer.Notify(c.Parrot)
	c.Workmode.Notify(c.Parrot)
	c.Entry.Notify(c.Parrot)
	c.Parrot.Notify(c.Entry)

	c.Settings.Notify(c.Entry)
	c.Settings.Notify(c.Workmode)
	c.Settings.Notify(c.Keyer)
	c.Settings.Notify(c.Logbook)
	c.Settings.Notify(c.QSOList)
	c.Settings.Notify(c.QTCController)
	c.Settings.Notify(c.ScoreController)
	c.Settings.Notify(c.Rate)
	c.Settings.Notify(c.Callinfo)
	c.Settings.Notify(c.Clusters)
	c.Settings.Notify(c.Bandmap)
	c.Settings.Notify(c.dxccFinder)
	c.Settings.Notify(c.callHistoryFinder)
	c.Settings.Notify(settings.SettingsListenerFunc(func(s core.Settings) {
		if !c.dxccFinder.Available() {
			return
		}
		if !c.Logbook.Valid() {
			c.Logbook.Refresh()
		}
	}))
	c.Settings.Notify(settings.StationListenerFunc(func(station core.Station) {
		bp := core.BandplanByID(station.Bandplan)
		for _, v := range c.VFOs {
			v.SetBandplan(bp)
		}
		c.Radio.SetBandplan(bp)
		c.Clusters.SetBandplan(bp)
		c.Refresh()
	}))

	c.dxccFinder.WhenAvailable(func() {
		c.asyncRunner(func() {
			// TODO: check this
			if !c.Logbook.Valid() {
				c.Logbook.StationChanged(c.Settings.Station())
				c.Logbook.ContestChanged(c.Settings.Contest())
			}
			if !c.QSOList.Valid() {
				c.QSOList.ContestChanged(c.Settings.Contest())
			}
			if !c.Clusters.Valid() {
				c.Clusters.StationChanged(c.Settings.Station())
			}
			c.Refresh()
		})
		c.ServiceStatus.StatusChanged(core.DXCCService, true)
	})
	c.scpFinder.WhenAvailable(func() {
		c.ServiceStatus.StatusChanged(core.SCPService, true)
	})
	c.ServiceStatus.StatusChanged(core.MapService, true)
	c.hamDXMap.WhenStopped(func() {
		c.ServiceStatus.StatusChanged(core.MapService, false)
	})

	// Toggle the workmode to make sure all the listeners are notified about the current workmode.
	c.Workmode.SetWorkmode(core.Run)
	c.Workmode.SetWorkmode(core.SearchPounce)

	c.ClockView.StartAutoRefresh()
	c.Rate.StartAutoRefresh()

	err = c.openCurrentLog()
	if err != nil {
		c.Quit()
	}
}

func (c *Controller) openCurrentLog() error {
	filename := c.session.LastFilename()
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

	qsos, station, contest, keyerSettings, qtcs, err := store.ReadAll()
	if err != nil {
		log.Printf("Cannot load %s: %v", filepath.Base(filename), err)
		qsos = []core.QSO{}
		qtcs = []core.QTC{}
	} else {
		c.Settings.SetWriter(store)
		if station != nil {
			c.Settings.SetStation(*station)
		}
		if contest != nil {
			c.Settings.SetContest(*contest)
		}
		c.Keyer.SetWriter(store)
		if keyerSettings != nil {
			c.Keyer.SetSettings(*keyerSettings, "")
		}
	}
	c.loadLogbook(filename, store, qsos, qtcs)
	return nil
}

func (c *Controller) loadLogbook(filename string, store *store.FileStore, qsos []core.QSO, qtcs []core.QTC) {
	c.filename = filename
	c.store = store
	c.Logbook.Load(store, qsos, qtcs)

	if c.view != nil {
		c.view.ShowFilename(c.filename)
	}

	err := c.session.SetLastFilename(c.filename)
	if err != nil {
		log.Println(err)
	}
}

func (c *Controller) Shutdown() {
	if c.remoteServer != nil {
		c.remoteServer.Stop()
	}
	c.Radio.Stop()
}

func (c *Controller) RunScript(ctx context.Context, script Script) {
	go func() {
		cont := true
		for cont {
			select {
			case <-ctx.Done():
				return
			default:
				cont = script.Step(ctx, c, c.asyncRunner)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (c *Controller) OpenWiki() {
	c.openWithExternalApplication(wikiURL)
}

func (c *Controller) About() {
	var sponsorText string

	sponsors := strings.Split(c.sponsors, "\n")
	if len(sponsors) > 0 {
		for sponsors[len(sponsors)-1] == "" {
			sponsors = sponsors[:len(sponsors)-1]
		}
		sponsorsCSV := strings.Join(sponsors, ", ")
		sponsorText = fmt.Sprintf("sponsored by:\n%s\n\n", sponsorsCSV)
	}

	c.view.ShowInfoDialog("Hello Contest\n\nVersion %s\n\n%sThis software is published under the MIT License.\n(c) Florian Thienel/DL3NEY", c.version, sponsorText)
}

func (c *Controller) ShowInfo(format string, args ...any) {
	c.view.ShowInfoDialog(format, args...)
}

func (c *Controller) ShowQuestion(format string, args ...any) bool {
	return c.view.ShowQuestionDialog(format, args...)
}

func (c *Controller) ShowError(format string, args ...any) {
	c.view.ShowErrorDialog(format, args...)
}

func (c *Controller) Sponsors() {
	c.openWithExternalApplication(sponsorsURL)
}

func (c *Controller) OpenContestRulesPage() {
	c.Settings.OpenContestRulesPage()
}

func (c *Controller) OpenContestUploadPage() {
	c.Settings.OpenContestUploadPage()
}

func (c *Controller) OpenSettings() {
	c.Settings.Show()
}

func (c *Controller) OpenConfigurationFile() {
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
	return proposeFilename(c.Settings.Contest().Name, c.Settings.Station().Callsign.BaseCall)
}

func proposeFilename(contestName, callsign string) string {
	result := strings.Join([]string{contestName, callsign}, " ")
	result = strings.TrimSpace(result)
	result = strings.ToUpper(result)
	result = strings.ReplaceAll(result, " ", "_")
	return result
}

func (c *Controller) New() {
	var err error
	newContest, ok := c.NewContestController.Run()
	if !ok {
		return
	}

	var proposedName string
	if newContest.Name == "" {
		proposedName = c.Settings.ProposeContestName(newContest.Identifier)
	} else {
		proposedName = newContest.Name
	}
	proposedFilename := proposeFilename(proposedName, c.Settings.Station().Callsign.BaseCall) + ".log"

	filename, ok, err := c.view.SelectSaveFile("New Logfile", c.configuration.LogDirectory(), proposedFilename, "*.log")
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
		c.view.ShowErrorDialog("Cannot create %s: %v", filepath.Base(newContest.Filename), err)
		return
	}

	c.Settings.Reset()
	c.Settings.SelectContestIdentifier(newContest.Identifier)
	c.Settings.EnterContestName(newContest.Name)
	c.Keyer.SetSettings(c.configuration.KeyerSettings(), newContest.Identifier)

	err = store.WriteStation(c.Settings.Station())
	if err != nil {
		c.view.ShowErrorDialog("Cannot save as %s: %v", filepath.Base(newContest.Filename), err)
		return
	}
	err = store.WriteContest(c.Settings.Contest())
	if err != nil {
		c.view.ShowErrorDialog("Cannot save as %s: %v", filepath.Base(newContest.Filename), err)
		return
	}
	err = store.WriteKeyer(c.Keyer.KeyerSettings())
	if err != nil {
		c.view.ShowErrorDialog("Cannot save as %s: %v", filepath.Base(newContest.Filename), err)
		return
	}

	c.Settings.SetWriter(store)
	c.Keyer.SetWriter(store)
	c.loadLogbook(filename, store, nil, nil)

	c.OpenSettings()
}

func (c *Controller) Open() {
	filename, ok, err := c.view.SelectOpenFile("Open Logfile", c.configuration.LogDirectory(), "*.log")
	if !ok {
		return
	}
	if err != nil {
		c.view.ShowErrorDialog("Cannot select a file: %v", err)
		return
	}

	store := store.NewFileStore(filename)
	qsos, station, contest, keyerSettings, qtcs, err := store.ReadAll()
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
	if keyerSettings != nil {
		c.Keyer.SetSettings(*keyerSettings, "")
	}
	c.loadLogbook(filename, store, qsos, qtcs)
}

func (c *Controller) SaveAs() {
	proposedName := c.proposeFilename() + ".log"
	filename, ok, err := c.view.SelectSaveFile("Save Logfile As", c.configuration.LogDirectory(), proposedName, "*.log")
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

	err = c.session.SetLastFilename(c.filename)
	if err != nil {
		log.Println(err)
	}
}

func (c *Controller) ExportSummary() {
	result, ok := c.SummaryController.Run(c.Settings)
	if !ok {
		return
	}

	proposedName := c.proposeFilename() + "_summary.txt"
	filename, ok, err := c.view.SelectSaveFile("Export Summary File", c.configuration.LogDirectory(), proposedName, "*.txt")
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

	err = summary.Export(file, result.Summary)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export summary to %s: %v", filename, err)
		return
	}

	if result.OpenAfterExport {
		c.openWithExternalApplication(filename)
	}
}

func (c *Controller) ExportCabrillo() {
	result, ok := c.ExportCabrilloController.Run(c.Settings, c.Logbook.Total(), c.Logbook.AllQSOs(), c.Logbook.AllQTCs())
	if !ok {
		return
	}

	proposedName := c.proposeFilename() + ".cabrillo"
	filename, ok, err := c.view.SelectSaveFile("Export Cabrillo File", c.configuration.LogDirectory(), proposedName, "*.cabrillo")
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

	err = cabrillo.Export(file, result.Export)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export Cabrillo to %s: %v", filename, err)
		return
	}

	if result.OpenUploadAfterExport {
		c.OpenContestUploadPage()
	}
	if result.OpenAfterExport {
		c.openWithExternalApplication(filename)
	}
}

func (c *Controller) ExportADIF() {
	proposedName := c.proposeFilename() + ".adif"
	filename, ok, err := c.view.SelectSaveFile("Export ADIF File", c.configuration.LogDirectory(), proposedName, "*.adif")
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
	err = adif.Export(file, c.Logbook.AllQSOs()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export ADIF to %s: %v", filename, err)
		return
	}
}

func (c *Controller) ExportCSV() {
	proposedName := c.proposeFilename() + ".csv"
	filename, ok, err := c.view.SelectSaveFile("Export CSV File", c.configuration.LogDirectory(), proposedName, "*.csv")
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
		c.Logbook.AllQSOs()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export Cabrillo to %s: %v", filename, err)
		return
	}
}

func (c *Controller) ExportCallhistory() {
	proposedName := c.proposeFilename() + ".txt"
	filename, ok, err := c.view.SelectSaveFile("Export Call History File", c.configuration.LogDirectory(), proposedName, "*.txt")
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

	err = callhistory.Export(file, c.Settings.Contest().CallHistoryFieldNames, c.Logbook.AllQSOs()...)
	if err != nil {
		c.view.ShowErrorDialog("Cannot export call history to %s: %v", filename, err)
		return
	}
	c.openWithExternalApplication(filename)
}

func (c *Controller) ShowQSOs() {
	c.QSOList.Show()
	c.view.BringToFront()
}

func (c *Controller) ShowQTCs() {
	c.QTCList.Show()
	c.view.BringToFront()
}

func (c *Controller) ShowScoreGraph() {
	c.ScoreController.ShowGraph()
	c.view.BringToFront()
}

func (c *Controller) ShowScoreTable() {
	c.ScoreController.ShowTable()
	c.view.BringToFront()
}

func (c *Controller) ShowRate() {
	c.Rate.Show()
	c.view.BringToFront()
}

func (c *Controller) ShowSpots() {
	c.Bandmap.Show()
	c.view.BringToFront()
}

func (c *Controller) ShowClock() {
	c.ClockView.Show()
	c.view.BringToFront()
}

func (c *Controller) Refresh() {
	c.Logbook.Refresh()
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

func (c *Controller) RefreshPrediction() {
	c.Entry.RefreshPrediction()
}

func (c *Controller) LogQSO() {
	c.Entry.Log()
}

func (c *Controller) StartParrot() {
	c.Parrot.Start()
}

func (c *Controller) ESMEnabled() bool {
	return c.Entry.ESMEnabled()
}

func (c *Controller) SetESMEnabled(enabled bool) {
	c.Entry.SetESMEnabled(enabled)
}

func (c *Controller) SwitchToSPWorkmode() {
	c.Workmode.SetWorkmode(core.SearchPounce)
}

func (c *Controller) SwitchToRunWorkmode() {
	c.Workmode.SetWorkmode(core.Run)
}

func (c *Controller) OfferQTC() {
	c.QTCController.OfferQTC()
}

func (c *Controller) RequestQTC() {
	c.QTCController.RequestQTC()
}

func (c *Controller) XITActive() bool {
	return c.VFOs[core.VFO1].XITActive()
}

func (c *Controller) SetXITActive(active bool) {
	c.VFOs[core.VFO1].SetXITActive(active)
}

func (c *Controller) ShiftXIT(delta core.Frequency) {
	c.VFOs[core.VFO1].ShiftXITOffset(delta)
}

func (c *Controller) XITUp() {
	c.ShiftXIT(core.DefaultXITShift)
}

func (c *Controller) XITDown() {
	c.ShiftXIT(-core.DefaultXITShift)
}

func (c *Controller) MuteAudio(vfo core.VFOID) {
	c.VFOs[vfo].MuteAudio()
}

func (c *Controller) UnmuteAudio(vfo core.VFOID) {
	c.VFOs[vfo].UnmuteAudio()
}

func (c *Controller) ShiftFrequency(delta core.Frequency) {
	c.Entry.ShiftFrequency(delta)
}

func (c *Controller) FrequencyUp() {
	c.ShiftFrequency(core.DefaultFrequencyShift)
}

func (c *Controller) FrequencyDown() {
	c.ShiftFrequency(-core.DefaultFrequencyShift)
}

func (c *Controller) KeyerShiftSpeed(delta int) {
	c.Keyer.ShiftSpeed(delta)
}

func (c *Controller) KeyerSpeedUp() {
	c.KeyerShiftSpeed(core.DefaultSpeedShift)
}

func (c *Controller) KeyerSpeedDown() {
	c.KeyerShiftSpeed(-core.DefaultSpeedShift)
}

func (c *Controller) ToggleAudio(vfo core.VFOID) {
	c.VFOs[vfo].ToggleAudio()
}

func (c *Controller) MarkInBandmap() {
	c.Entry.MarkInBandmap()
}

func (c *Controller) GotoHighestValueSpot() {
	c.Bandmap.GotoHighestValueEntry()
}

func (c *Controller) GotoNearestSpot() {
	c.Bandmap.GotoNearestEntry()
}

func (c *Controller) GotoNextSpotUp() {
	c.Bandmap.GotoNextEntryUp()
}

func (c *Controller) GotoNextSpotDown() {
	c.Bandmap.GotoNextEntryDown()
}

func (c *Controller) SendSpotsToTci() bool {
	return c.session.SendSpotsToTci()
}

func (c *Controller) SetSendSpotsToTci(sendSpotsToTci bool) {
	if c.Radio == nil {
		return
	}
	c.Radio.SetSendSpotsToTci(sendSpotsToTci)

	err := c.session.SetSendSpotsToTci(sendSpotsToTci)
	if err != nil {
		log.Println(err)
	}
}

func (c *Controller) SetSpotSourceEnabled(name string, enabled bool) {
	c.Clusters.SetSpotSourceEnabled(name, enabled)
}

func (c *Controller) SelectRadio(name string) {
	err := c.Radio.SelectRadio(name)
	if err != nil {
		return
	}

	err = c.session.SetRadio1(c.Radio.Radio())
	if err != nil {
		log.Println(err)
	}
	err = c.session.SetKeyer1(c.Radio.Keyer())
	if err != nil {
		log.Println(err)
	}
}

func (c *Controller) SelectKeyer(name string) {
	err := c.Radio.SelectKeyer(name)
	if err != nil {
		log.Println(err) // TODO show an error dialog
		return
	}

	err = c.session.SetKeyer1(c.Radio.Keyer())
	if err != nil {
		log.Println(err)
	}
}

func (c *Controller) Stop() {
	c.Keyer.Stop()
}

func (c *Controller) DoubleStop() {
	c.Entry.Clear()
}

// shiftAmount reads the "amount" parameter and parses it into an int. It
// returns def if the parameter is missing or empty, and an error if it is
// present but cannot be parsed.
func shiftAmount(params map[string]string, def int) (int, error) {
	s, ok := params["amount"]
	if !ok || s == "" {
		return def, nil
	}
	return strconv.Atoi(s)
}

func (c *Controller) DoAction(id string, params map[string]string) error {
	switch id {
	case core.ActionRadioShiftFrequency:
		amount, err := shiftAmount(params, int(core.DefaultFrequencyShift))
		if err != nil {
			return err
		}
		c.ShiftFrequency(core.Frequency(amount))
	case core.ActionRadioFrequencyUp:
		c.FrequencyUp()
	case core.ActionRadioFrequencyDown:
		c.FrequencyDown()
	case core.ActionRadioShiftXIT:
		amount, err := shiftAmount(params, int(core.DefaultXITShift))
		if err != nil {
			return err
		}
		c.ShiftXIT(core.Frequency(amount))
	case core.ActionRadioXITUp:
		c.XITUp()
	case core.ActionRadioXITDown:
		c.XITDown()
	case core.ActionKeyerShiftSpeed:
		amount, err := shiftAmount(params, core.DefaultSpeedShift)
		if err != nil {
			return err
		}
		c.KeyerShiftSpeed(amount)
	case core.ActionKeyerSpeedUp:
		c.KeyerSpeedUp()
	case core.ActionKeyerSpeedDown:
		c.KeyerSpeedDown()
	case core.ActionEntryClear:
		c.ClearEntryFields()
	case core.ActionEntryGotoEntryField:
		c.GotoEntryFields()
	case core.ActionEntryEditLastQSO:
		c.EditLastQSO()
	case core.ActionEntryRefreshPrediction:
		c.RefreshPrediction()
	case core.ActionEntrySelectBestMatch:
		c.Entry.SelectBestMatchOnFrequency()
	case core.ActionEntryLogQSO:
		c.LogQSO()
	case core.ActionEntryStartParrot:
		c.StartParrot()
	case core.ActionEntryNextESMStep:
		c.Entry.NextESMStep()
	case core.ActionEntryToggleFocusedVFO:
		c.Entry.ToggleFocusedVFO()
	case core.ActionEntryFocusVFO1:
		c.Entry.FocusVFO1()
	case core.ActionEntryFocusVFO2:
		c.Entry.FocusVFO2()
	case core.ActionEntryLogVFO1:
		c.Entry.LogVFO(core.VFO1)
	case core.ActionEntryLogVFO2:
		c.Entry.LogVFO(core.VFO2)
	case core.ActionEntryClearVFO1:
		c.Entry.ClearVFO(core.VFO1)
	case core.ActionEntryClearVFO2:
		c.Entry.ClearVFO(core.VFO2)
	case core.ActionEntryOfferQTC:
		c.OfferQTC()
	case core.ActionEntryRequestQTC:
		c.RequestQTC()
	case core.ActionBandmapMark:
		c.MarkInBandmap()
	case core.ActionBandmapGotoHighestValueSpot:
		c.GotoHighestValueSpot()
	case core.ActionBandmapGotoNearestSpot:
		c.GotoNearestSpot()
	case core.ActionBandmapGotoNextSpotUp:
		c.GotoNextSpotUp()
	case core.ActionBandmapGotoNextSpotDown:
		c.GotoNextSpotDown()
	case core.ActionWindowShowQSOs:
		c.ShowQSOs()
	case core.ActionWindowShowQTCs:
		c.ShowQTCs()
	case core.ActionWindowShowScoreGraph:
		c.ShowScoreGraph()
	case core.ActionWindowShowScoreTable:
		c.ShowScoreTable()
	case core.ActionWindowShowRate:
		c.ShowRate()
	case core.ActionWindowShowSpots:
		c.ShowSpots()
	case core.ActionWindowShowClock:
		c.ShowClock()
	case core.ActionKeyerSendMacro1:
		c.Keyer.SendMacro(0)
	case core.ActionKeyerSendMacro2:
		c.Keyer.SendMacro(1)
	case core.ActionKeyerSendMacro3:
		c.Keyer.SendMacro(2)
	case core.ActionKeyerSendMacro4:
		c.Keyer.SendMacro(3)
	case core.ActionRadioMuteAudioVFO1:
		c.MuteAudio(core.VFO1)
	case core.ActionRadioMuteAudioVFO2:
		c.MuteAudio(core.VFO2)
	case core.ActionRadioUnmuteAudioVFO1:
		c.UnmuteAudio(core.VFO1)
	case core.ActionRadioUnmuteAudioVFO2:
		c.UnmuteAudio(core.VFO2)
	case core.ActionRadioToggleAudioVFO1:
		c.ToggleAudio(core.VFO1)
	case core.ActionRadioToggleAudioVFO2:
		c.ToggleAudio(core.VFO2)
	default:
		return fmt.Errorf("unknown action: %s", id)
	}
	return nil
}
