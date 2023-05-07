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
)

// Run the application
func Run(version string, args []string) {
	var err error
	app := &application{id: "ft.hellocontest", version: version}

	gdk.SetAllowedBackends("x11")

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
	id             string
	version        string
	app            *gtk.Application
	builder        *gtk.Builder
	windowGeometry *gmtry.Geometry
	mainWindow     *mainWindow
	callinfoWindow *callinfoWindow
	scoreWindow    *scoreWindow
	rateWindow     *rateWindow
	bandmapWindow  *bandmapWindow
	settingsDialog *settingsDialog

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

func (a *application) activate() {
	a.builder = setupBuilder()

	configuration, err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}
	a.controller = app.NewController(a.version, clock.New(), a.app, a.runAsync, configuration)
	a.controller.Startup()

	a.mainWindow = setupMainWindow(a.builder, a.app)
	a.callinfoWindow = setupCallinfoWindow(a.windowGeometry)
	a.scoreWindow = setupScoreWindow(a.windowGeometry)
	a.rateWindow = setupRateWindow(a.windowGeometry)
	a.bandmapWindow = setupBandmapWindow(a.windowGeometry, a.controller.Bandmap)
	a.settingsDialog = setupSettingsDialog(a.controller.Settings)

	a.mainWindow.SetMainMenuController(a.controller)
	a.mainWindow.SetSpotSourceMenuController(a.controller)
	a.mainWindow.SetStopKeyController(a.controller)
	a.mainWindow.SetLogbookController(a.controller.QSOList)
	a.mainWindow.SetEntryController(a.controller.Entry)
	a.mainWindow.SetWorkmodeController(a.controller.Workmode)
	a.mainWindow.SetKeyerController(a.controller.Keyer)

	a.controller.SetView(a.mainWindow)
	a.controller.QSOList.Notify(a.mainWindow)
	a.controller.Entry.SetView(a.mainWindow)
	a.controller.Workmode.SetView(a.mainWindow)
	a.controller.Workmode.Notify(a.mainWindow)
	a.controller.Keyer.SetView(a.mainWindow)
	a.controller.ServiceStatus.Notify(a.mainWindow)
	a.controller.Callinfo.SetView(a.callinfoWindow)
	a.controller.Score.SetView(a.scoreWindow)
	a.controller.Rate.SetView(a.rateWindow)
	a.controller.Rate.Notify(a.scoreWindow)
	a.controller.Bandmap.SetView(a.bandmapWindow)
	a.controller.Settings.SetView(a.settingsDialog)
	a.controller.Clusters.SetView(a.mainWindow)

	a.mainWindow.ConnectToGeometry(a.windowGeometry)
	err = a.windowGeometry.Restore()
	if err != nil {
		a.useDefaultWindowGeometry(err)
	}

	a.mainWindow.Show()
	a.callinfoWindow.RestoreVisibility()
	a.scoreWindow.RestoreVisibility()
	a.rateWindow.RestoreVisibility()
	a.bandmapWindow.RestoreVisibility()

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
