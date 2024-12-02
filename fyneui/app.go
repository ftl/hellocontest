package fyneui

import (
	"log"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	"github.com/ftl/hellocontest/core/app"
	"github.com/ftl/hellocontest/core/cfg"
	"github.com/ftl/hellocontest/core/clock"
)

const AppID = "ft.hellocontest"

func Run(version string, sponsors string, args []string) {
	app := &application{
		id:       AppID,
		version:  version,
		sponsors: sponsors,
	}

	app.app = fyneapp.NewWithID(app.id)

	app.app.Lifecycle().SetOnStarted(app.activate)
	app.app.Run()
}

type application struct {
	id       string
	version  string
	sponsors string

	app        fyne.App
	mainWindow *mainWindow
	mainMenu   *mainMenu

	controller *app.Controller
}

func (a *application) activate() {
	configuration, err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}
	a.controller = app.NewController(a.version, clock.New(), a.app, a.runAsync, configuration, a.sponsors)
	a.controller.Startup()

	a.mainWindow = setupMainWindow(a.app.NewWindow("Hello Contest"))
	a.mainMenu = setupMainMenu(a.mainWindow.window, a.controller)

	a.controller.SetView(a.mainWindow)

	a.mainWindow.UseDefaultWindowGeometry() // TODO: store/restore the window geometry
	a.mainWindow.Show()
}

func (a *application) runAsync(f func()) {
	f() // TODO: this is probably sufficient with Fyne, Fyne is thread safe afaik
}
