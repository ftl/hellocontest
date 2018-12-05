package ui

import (
	logger "log"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/entry"
	"github.com/ftl/hellocontest/core/log"
	"github.com/ftl/hellocontest/core/store"
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
	log        core.Log
	entry      core.EntryController
	store      core.Store
}

func (app *application) startup() {
}

func (app *application) activate() {
	app.builder = setupBuilder()

	app.mainWindow = setupMainWindow(app.builder, app.app)
	app.mainWindow.Show()

	app.clock = clock.New()
	app.store = store.New("current.log")
	app.log = loadLog(app)
	app.log.SetView(app.mainWindow)
	app.log.OnRowAdded(app.store.Write)
	app.entry = entry.NewController(app.clock, app.log)
	app.entry.SetView(app.mainWindow)
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

func loadLog(app *application) core.Log {
	log, err := log.Load(app.clock, app.store)
	if err != nil {
		panic(err)
	}
	return log
}
