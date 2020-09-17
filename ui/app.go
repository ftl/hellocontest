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
func Run(args []string) {
	var err error
	app := &application{id: "ft.hellocontest"}

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
	app            *gtk.Application
	builder        *gtk.Builder
	windowGeometry *gmtry.Geometry
	mainWindow     *mainWindow
	callinfoWindow *callinfoWindow
	controller     *app.Controller
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
		log.Println(err)
	}
	a.mainWindow = setupMainWindow(a.builder, a.app)
	a.callinfoWindow = setupCallinfoWindow(a.builder)

	a.controller = app.NewController(clock.New(), a.app, configuration)
	a.controller.Startup()
	a.controller.SetView(a.mainWindow)
	a.controller.SetLogbookView(a.mainWindow)
	a.controller.SetEntryView(a.mainWindow)
	a.controller.SetWorkmodeView(a.mainWindow)
	a.controller.SetKeyerView(a.mainWindow)
	a.controller.SetCallinfoView(a.callinfoWindow)

	a.mainWindow.SetMainMenuController(a.controller)

	a.mainWindow.ConnectToGeometry(a.windowGeometry)
	a.callinfoWindow.ConnectToGeometry(a.windowGeometry)
	err = a.windowGeometry.Restore()
	if err != nil {
		a.useDefaultWindowGeometry(err)
	}

	a.mainWindow.Show()
}

func (a *application) shutdown() {
	a.controller.Shutdown()

	err := a.windowGeometry.Store()
	if err != nil {
		log.Printf("Cannot store window geometry: %v", err)
	}
}

func setupBuilder() *gtk.Builder {
	builder, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Cannot create builder: ", err)
	}

	// builder.AddFromFile("ui/glade/contest.glade")
	builder.AddFromString(glade.MustAssetString("contest.glade"))

	return builder
}

func connectToGeometry(geometry *gmtry.Geometry, id gmtry.ID, window *gtk.Window) {
	geometry.Add(id, window)

	window.Connect("configure-event", func(_ interface{}, event *gdk.Event) {
		if !window.IsVisible() {
			// return
		}
		e := gdk.EventConfigureNewFromEvent(event)
		w := geometry.Get(id)
		w.SetPosition(window.GetPosition())
		w.SetSize(e.Width(), e.Height())
	})
	window.Connect("window-state-event", func(_ interface{}, event *gdk.Event) {
		if !window.IsVisible() {
			// return
		}
		e := gdk.EventWindowStateNewFromEvent(event)
		if e.ChangedMask()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED {
			geometry.Get(id).SetMaximized(e.NewWindowState()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED)
		}
	})
}
