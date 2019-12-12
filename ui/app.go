package ui

import (
	logger "log"
	"os"
	"path/filepath"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/ftl/gmtry"

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

	app.app.Connect("startup", app.startup)
	app.app.Connect("activate", app.activate)
	app.app.Connect("shutdown", app.shutdown)

	app.app.Run(args)
}

type application struct {
	id             string
	app            *gtk.Application
	builder        *gtk.Builder
	windowGeometry gmtry.Windows
	mainWindow     *mainWindow
	controller     core.AppController
}

func (app *application) startup() {
	filename := filepath.Join(cfg.Directory(), "hellocontest.geometry")
	logger.Printf("Loading window geometry from %s", filename)

	f, err := os.Open(filename)
	if err != nil {
		app.useDefaultWindowGeometry(err)
		return
	}
	defer f.Close()

	app.windowGeometry, err = gmtry.LoadWindows(f)
	if err != nil {
		app.useDefaultWindowGeometry(err)
	}
}

func (app *application) useDefaultWindowGeometry(cause error) {
	logger.Printf("Cannot load window geometry, using defaults instead: %v", cause)
	app.windowGeometry = gmtry.NewWindows()
	app.windowGeometry["main"] = &gmtry.Window{
		ID:     "main",
		X:      300,
		Y:      100,
		Width:  569,
		Height: 700,
	}
}

func (app *application) activate() {
	app.builder = setupBuilder()

	configuration, err := cfg.Load()
	if err != nil {
		logger.Println(err)
	}
	app.controller = coreapp.NewController(clock.New(), app.app, configuration)
	app.mainWindow = setupMainWindow(app.builder, app.app, app.windowGeometry)

	app.controller.Startup()
	app.controller.SetView(app.mainWindow)
	app.controller.SetLogView(app.mainWindow)
	app.controller.SetEntryView(app.mainWindow)
	app.controller.SetKeyerView(app.mainWindow)

	app.mainWindow.Show()
}

func (app *application) shutdown() {
	app.controller.Shutdown()

	filename := filepath.Join(cfg.Directory(), "hellocontest.geometry")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Printf("Cannot open window geometry file: %v", err)
		return
	}
	defer f.Close()

	err = app.windowGeometry.Store(f)
	if err != nil {
		logger.Printf("Cannot store window geometry: %v", err)
	} else {
		logger.Printf("Stored window geometry in %s", f.Name())
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
