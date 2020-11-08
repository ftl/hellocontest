package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

// MainMenuController provides the functionality for the main menu.
type MainMenuController interface {
	About()
	New()
	Open()
	SaveAs()
	ExportCabrillo()
	ExportADIF()
	ExportCSV()
	Quit()
	ShowCallinfo()
	ShowScore()
}

type mainMenu struct {
	controller MainMenuController

	fileNew            *gtk.MenuItem
	fileOpen           *gtk.MenuItem
	fileSaveAs         *gtk.MenuItem
	fileExportCabrillo *gtk.MenuItem
	fileExportADIF     *gtk.MenuItem
	fileExportCSV      *gtk.MenuItem
	fileQuit           *gtk.MenuItem
	windowCallinfo     *gtk.MenuItem
	windowScore        *gtk.MenuItem
	helpAbout          *gtk.MenuItem
}

func setupMainMenu(builder *gtk.Builder) *mainMenu {
	result := new(mainMenu)

	result.fileNew = getUI(builder, "menuFileNew").(*gtk.MenuItem)
	result.fileOpen = getUI(builder, "menuFileOpen").(*gtk.MenuItem)
	result.fileSaveAs = getUI(builder, "menuFileSaveAs").(*gtk.MenuItem)
	result.fileExportCabrillo = getUI(builder, "menuFileExportCabrillo").(*gtk.MenuItem)
	result.fileExportADIF = getUI(builder, "menuFileExportADIF").(*gtk.MenuItem)
	result.fileExportCSV = getUI(builder, "menuFileExportCSV").(*gtk.MenuItem)
	result.fileQuit = getUI(builder, "menuFileQuit").(*gtk.MenuItem)
	result.windowCallinfo = getUI(builder, "menuWindowCallinfo").(*gtk.MenuItem)
	result.windowScore = getUI(builder, "menuWindowScore").(*gtk.MenuItem)
	result.helpAbout = getUI(builder, "menuHelpAbout").(*gtk.MenuItem)

	result.fileNew.Connect("activate", result.onNew)
	result.fileOpen.Connect("activate", result.onOpen)
	result.fileSaveAs.Connect("activate", result.onSaveAs)
	result.fileExportCabrillo.Connect("activate", result.onExportCabrillo)
	result.fileExportADIF.Connect("activate", result.onExportADIF)
	result.fileExportCSV.Connect("activate", result.onExportCSV)
	result.fileQuit.Connect("activate", result.onQuit)
	result.windowCallinfo.Connect("activate", result.onCallinfo)
	result.windowScore.Connect("activate", result.onScore)
	result.helpAbout.Connect("activate", result.onAbout)

	return result
}

func (m *mainMenu) SetMainMenuController(controller MainMenuController) {
	m.controller = controller
}

func (m *mainMenu) onAbout() {
	m.controller.About()
}

func (m *mainMenu) onNew() {
	m.controller.New()
}

func (m *mainMenu) onOpen() {
	m.controller.Open()
}

func (m *mainMenu) onSaveAs() {
	m.controller.SaveAs()
}

func (m *mainMenu) onExportCabrillo() {
	m.controller.ExportCabrillo()
}

func (m *mainMenu) onExportADIF() {
	m.controller.ExportADIF()
}

func (m *mainMenu) onExportCSV() {
	m.controller.ExportCSV()
}

func (m *mainMenu) onQuit() {
	m.controller.Quit()
}

func (m *mainMenu) onCallinfo() {
	m.controller.ShowCallinfo()
}

func (m *mainMenu) onScore() {
	m.controller.ShowScore()
}
