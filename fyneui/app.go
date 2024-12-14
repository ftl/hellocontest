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

	app             fyne.App
	shortcuts       *Shortcuts
	mainWindow      *mainWindow
	mainMenu        *mainMenu
	qsoList         *qsoList
	workmodeControl *workmodeControl
	keyerControl    *keyerControl
	statusBar       *statusBar

	controller *app.Controller
}

func (a *application) activate() {
	a.mainWindow = setupMainWindow(a.app.NewWindow("Hello Contest"))
	a.mainWindow.UseDefaultWindowGeometry() // TODO: store/restore the window geometry
	a.mainWindow.Show()

	configuration, err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}
	a.controller = app.NewController(a.version, clock.New(), a.app, a.runAsync, configuration, a.sponsors)
	a.controller.Startup()

	a.shortcuts = setupShortcuts(a.controller, a.controller.Keyer)
	a.mainMenu = setupMainMenu(a.controller, a.shortcuts)
	a.qsoList = setupQSOList()
	a.workmodeControl = setupWorkmodeControl()
	a.keyerControl = setupKeyerControl()
	a.statusBar = setupStatusBar()

	a.qsoList.SetLogbookController(a.controller.QSOList)
	a.workmodeControl.SetWorkmodeController(a.controller.Workmode)
	a.keyerControl.SetKeyerController(a.controller.Keyer)

	a.controller.SetView(a.mainWindow)
	a.controller.QSOList.Notify(a.qsoList)
	a.controller.Workmode.SetView(a.workmodeControl)
	a.controller.Workmode.Notify(a.mainMenu)
	a.controller.Keyer.SetView(a.keyerControl)
	a.controller.ServiceStatus.Notify(a.statusBar)

	a.mainWindow.setContent(a.qsoList.container, a.workmodeControl.container, a.keyerControl.container, a.statusBar.container)
	a.mainWindow.window.SetMainMenu(a.mainMenu.root)
	a.shortcuts.AddTo(a.mainWindow.window.Canvas())

	a.controller.Refresh()
}

func (a *application) runAsync(f func()) {
	f() // TODO: this is probably sufficient with Fyne, Fyne is thread safe afaik
}
