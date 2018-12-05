package ui

import (
	logger "log"

	"github.com/ftl/hellocontest/core"
	coreapp "github.com/ftl/hellocontest/core/app"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// Run the application
func Run(args []string) {
	var err error
	app := &application{id: "ft.hellocontest"}
	app.app, err = gtk.ApplicationNew(app.id, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		logger.Fatal("Cannot create application: ", err)
	}

	app.app.Connect("startup", app.startup)
	app.app.Connect("activate", app.activate)
	app.app.Connect("shutdown", app.shutdown)

	app.app.Run(args)
}

type application struct {
	id         string
	app        *gtk.Application
	builder    *gtk.Builder
	mainWindow *mainWindow
	clock      core.Clock
	controller core.AppController
}

func (app *application) startup() {
}

func (app *application) activate() {
	app.builder = setupBuilder()

	app.mainWindow = setupMainWindow(app.builder, app.app)
	app.mainWindow.Show()

	app.clock = clock.New()
	app.controller = coreapp.NewController(app.clock)
	app.controller.Startup()
	app.controller.SetView(app.mainWindow)
	app.controller.SetLogView(app.mainWindow)
	app.controller.SetEntryView(app.mainWindow)
}

func (app *application) shutdown() {
}

func setupBuilder() *gtk.Builder {
	builder, err := gtk.BuilderNew()
	if err != nil {
		logger.Fatal("Cannot create builder: ", err)
	}

	builder.AddFromFile("ui/glade/contest.glade")

	return builder
}
