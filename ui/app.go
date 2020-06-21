package ui

import (
	logger "log"
	"path/filepath"

	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	coreapp "github.com/ftl/hellocontest/core/app"
	"github.com/ftl/hellocontest/core/cfg"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/ui/glade"
)

// Run the application
func Run(args []string) {
	var err error
	app := &application{id: "ft.hellocontest"}
	app.app, err = gtk.ApplicationNew(app.id, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		logger.Fatal("Cannot create application: ", err)
	}

	gdk.SetAllowedBackends("x11")

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
	controller     core.AppController
}

func (app *application) startup() {
	filename := filepath.Join(cfg.Directory(), "hellocontest.geometry")

	app.windowGeometry = gmtry.NewGeometry(filename)
}

func (app *application) useDefaultWindowGeometry(cause error) {
	logger.Printf("Cannot load window geometry, using defaults instead: %v", cause)
	app.mainWindow.Window.Move(300, 100)
	app.mainWindow.Window.Window.Resize(569, 700)
}

func (app *application) activate() {
	app.builder = setupBuilder()

	configuration, err := cfg.Load()
	if err != nil {
		logger.Println(err)
	}
	app.controller = coreapp.NewController(clock.New(), app.app, configuration)
	app.mainWindow = setupMainWindow(app.builder, app.app)

	app.controller.Startup()
	app.controller.SetView(app.mainWindow)
	app.controller.SetLogView(app.mainWindow)
	app.controller.SetEntryView(app.mainWindow)
	app.controller.SetKeyerView(app.mainWindow)

	connectToGeometry(app.windowGeometry, "main", &app.mainWindow.Window.Window)
	err = app.windowGeometry.Restore()
	if err != nil {
		app.useDefaultWindowGeometry(err)
	}

	app.mainWindow.Show()
}

func (app *application) shutdown() {
	app.controller.Shutdown()

	err := app.windowGeometry.Store()
	if err != nil {
		logger.Printf("Cannot store window geometry: %v", err)
	}
}

func setupBuilder() *gtk.Builder {
	builder, err := gtk.BuilderNew()
	if err != nil {
		logger.Fatal("Cannot create builder: ", err)
	}

	// builder.AddFromFile("ui/glade/contest.glade")
	builder.AddFromString(glade.MustAssetString("contest.glade"))

	return builder
}

func connectToGeometry(geometry *gmtry.Geometry, id gmtry.ID, window *gtk.Window) {
	geometry.Add(id, window)

	window.Connect("configure-event", func(_ interface{}, event *gdk.Event) {
		e := gdk.EventConfigureNewFromEvent(event)
		w := geometry.Get(id)
		w.SetPosition(e.X(), e.Y())
		w.SetSize(e.Width(), e.Height())
	})
	window.Connect("window-state-event", func(_ interface{}, event *gdk.Event) {
		e := gdk.EventWindowStateNewFromEvent(event)
		if e.ChangedMask()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED {
			geometry.Get(id).SetMaximized(e.NewWindowState()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED)
		}
	})
}
