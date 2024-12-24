package ui

import (
	"log"
	"path/filepath"

	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core/app"
	"github.com/ftl/hellocontest/core/cfg"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/ui/glade"
	"github.com/ftl/hellocontest/ui/style"
)

const AppID = "ft.hellocontest"

// Run the application
func Run(version string, sponsors string, args []string) {
	var err error
	app := &application{id: AppID, version: version, sponsors: sponsors}

	gdk.SetAllowedBackends("x11")
	gtk.WindowSetDefaultIconName("hellocontest")

	app.app, err = gtk.ApplicationNew(app.id, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Cannot create application: ", err)
	}
	app.app.Connect("startup", app.startup)
	app.app.Connect("activate", app.activate)
	app.app.Connect("shutdown", app.shutdown)

	app.app.Run(args)
}

type application struct {
	id       string
	version  string
	sponsors string

	app                  *gtk.Application
	builder              *gtk.Builder
	style                *style.Style
	windowGeometry       *gmtry.Geometry
	mainWindow           *mainWindow
	scoreWindow          *scoreWindow
	rateWindow           *rateWindow
	spotsWindow          *spotsWindow
	newContestDialog     *newContestDialog
	exportCabrilloDialog *exportCabrilloDialog
	settingsDialog       *settingsDialog
	keyerSettingsDialog  *keyerSettingsDialog

	controller *app.Controller
}

func (a *application) startup() {
	filename := filepath.Join(cfg.Directory(), "hellocontest.geometry")

	a.windowGeometry = gmtry.NewGeometry(filename)
}

func (a *application) useDefaultWindowGeometry(cause error) {
	log.Printf("Cannot load window geometry, using defaults instead: %v", cause)
	a.mainWindow.UseDefaultWindowGeometry()
}

func (a *application) setAcceptFocus(acceptFocus bool) {
	a.rateWindow.SetAcceptFocus(acceptFocus)
	a.scoreWindow.SetAcceptFocus(acceptFocus)
	a.spotsWindow.SetAcceptFocus(acceptFocus)
}

func (a *application) activate() {
	a.builder = setupBuilder()

	configuration, err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}
	a.controller = app.NewController(a.version, clock.New(), a.app, a.runAsync, configuration, a.sponsors)
	a.controller.Startup()

	a.mainWindow = setupMainWindow(a.builder, a.app, a.style, a.setAcceptFocus)
	screen := a.mainWindow.window.GetScreen()
	a.style = style.New()
	a.style.AddToScreen(screen)

	a.scoreWindow = setupScoreWindow(a.windowGeometry, a.style)
	a.rateWindow = setupRateWindow(a.windowGeometry, a.style)
	a.spotsWindow = setupSpotsWindow(a.windowGeometry, a.style, a.controller.Bandmap)
	a.settingsDialog = setupSettingsDialog(a.mainWindow.window, a.controller.Settings)
	a.keyerSettingsDialog = setupKeyerSettingsDialog(a.mainWindow.window, a.controller.Keyer)
	a.newContestDialog = setupNewContestDialog(a.mainWindow.window, a.controller.NewContestController)
	a.exportCabrilloDialog = setupExportCabrilloDialog(a.mainWindow.window, a.controller.ExportCabrilloController)

	a.mainWindow.SetMainMenuController(a.controller)
	a.mainWindow.SetRadioMenuController(a.controller)
	a.mainWindow.SetSpotSourceMenuController(a.controller)
	a.mainWindow.SetStopKeyController(a.controller)
	a.mainWindow.SetLogbookController(a.controller.QSOList)
	a.mainWindow.SetEntryController(a.controller.Entry)
	a.mainWindow.SetWorkmodeController(a.controller.Workmode)
	a.mainWindow.SetKeyerController(a.controller.Keyer)
	a.mainWindow.SetCallinfoController(a.controller.Callinfo)

	a.controller.SetView(a.mainWindow)
	a.controller.QSOList.Notify(a.mainWindow)
	a.controller.Entry.SetView(a.mainWindow)
	a.controller.Workmode.SetView(a.mainWindow)
	a.controller.Workmode.Notify(a.mainWindow)
	a.controller.Radio.SetView(a.mainWindow)
	a.controller.Keyer.SetView(a.mainWindow)
	a.controller.Keyer.SetSettingsView(a.keyerSettingsDialog)
	a.controller.ServiceStatus.Notify(a.mainWindow)
	a.controller.Callinfo.SetView(a.mainWindow)
	a.controller.Score.SetView(a.scoreWindow)
	a.controller.Rate.SetView(a.rateWindow)
	a.controller.Rate.Notify(a.scoreWindow)
	// TODO: use a listener model for the bandmap to allow multiple views on the bandmap (scope, spots list, mini-scope)
	a.controller.Bandmap.SetView(a.spotsWindow)
	a.controller.Settings.SetView(a.settingsDialog)
	a.controller.Settings.Notify(a.mainWindow)
	a.controller.NewContestController.SetView(a.newContestDialog)
	a.controller.ExportCabrilloController.SetView(a.exportCabrilloDialog)
	a.controller.Clusters.SetView(a.mainWindow)
	a.controller.Parrot.SetView(a.mainWindow)

	a.mainWindow.ConnectToGeometry(a.windowGeometry)
	err = a.windowGeometry.Restore()
	if err != nil {
		a.useDefaultWindowGeometry(err)
	}

	a.mainWindow.Show()
	a.scoreWindow.RestoreVisibility()
	a.rateWindow.RestoreVisibility()
	a.spotsWindow.RestoreVisibility()

	a.controller.Refresh()
}

func (a *application) shutdown() {
	a.controller.Shutdown()

	err := a.windowGeometry.Store()
	if err != nil {
		log.Printf("Cannot store window geometry: %v", err)
	}
}

func (a *application) runAsync(f func()) {
	runAsync(f)
}

func runAsync(f func()) {
	glib.IdleAdd(func() bool {
		f()
		return false
	})
}

func setupBuilder() *gtk.Builder {
	builder, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Cannot create builder: ", err)
	}

	builder.AddFromString(glade.Assets)

	return builder
}

func connectToGeometry(geometry *gmtry.Geometry, id gmtry.ID, window *gtk.Window) {
	geometry.Add(id, window)

	window.Connect("configure-event", func(_ interface{}, event *gdk.Event) {
		e := gdk.EventConfigureNewFromEvent(event)
		w := geometry.Get(id)
		w.SetPosition(window.GetPosition())
		w.SetSize(e.Width(), e.Height())
	})
	window.Connect("window-state-event", func(_ interface{}, event *gdk.Event) {
		e := gdk.EventWindowStateNewFromEvent(event)
		if e.ChangedMask()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED {
			geometry.Get(id).SetMaximized(e.NewWindowState()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED)
		}
	})
	window.Connect("show", func() {
		w := geometry.Get(id)
		w.SetVisible(true)
	})
	window.Connect("hide", func() {
		w := geometry.Get(id)
		w.SetVisible(false)
	})
}
