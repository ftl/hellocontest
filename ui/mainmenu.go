package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type mainMenu struct {
	controller core.MainMenuController

	fileNew            *gtk.MenuItem
	fileOpen           *gtk.MenuItem
	fileSaveAs         *gtk.MenuItem
	fileExportCabrillo *gtk.MenuItem
	fileExportADIF     *gtk.MenuItem
	fileQuit           *gtk.MenuItem
	windowCallinfo     *gtk.MenuItem
}

func setupMainMenu(builder *gtk.Builder) *mainMenu {
	result := new(mainMenu)

	result.fileNew = getUI(builder, "menuFileNew").(*gtk.MenuItem)
	result.fileOpen = getUI(builder, "menuFileOpen").(*gtk.MenuItem)
	result.fileSaveAs = getUI(builder, "menuFileSaveAs").(*gtk.MenuItem)
	result.fileExportCabrillo = getUI(builder, "menuFileExportCabrillo").(*gtk.MenuItem)
	result.fileExportADIF = getUI(builder, "menuFileExportADIF").(*gtk.MenuItem)
	result.fileQuit = getUI(builder, "menuFileQuit").(*gtk.MenuItem)
	result.windowCallinfo = getUI(builder, "menuWindowCallinfo").(*gtk.MenuItem)

	result.fileNew.Connect("activate", result.onNew)
	result.fileOpen.Connect("activate", result.onOpen)
	result.fileSaveAs.Connect("activate", result.onSaveAs)
	result.fileExportCabrillo.Connect("activate", result.onExportCabrillo)
	result.fileExportADIF.Connect("activate", result.onExportADIF)
	result.fileQuit.Connect("activate", result.onQuit)
	result.windowCallinfo.Connect("activate", result.onCallinfo)

	return result
}

func (m *mainMenu) SetMainMenuController(controller core.MainMenuController) {
	m.controller = controller
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

func (m *mainMenu) onQuit() {
	m.controller.Quit()
}

func (m *mainMenu) onCallinfo() {
	m.controller.Callinfo()
}
