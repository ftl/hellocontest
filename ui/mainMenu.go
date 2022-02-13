package ui

import (
	"github.com/ftl/hellocontest/core"
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
	ExportCallhistory()
	OpenSettings()
	Quit()
	ShowCallinfo()
	ShowScore()
	ShowRate()
	ClearEntryFields()
	GotoEntryFields()
	EditLastQSO()
	LogQSO()
	SwitchToSPWorkmode()
	SwitchToRunWorkmode()
}

type mainMenu struct {
	controller MainMenuController

	fileNew               *gtk.MenuItem
	fileOpen              *gtk.MenuItem
	fileSaveAs            *gtk.MenuItem
	fileExportCabrillo    *gtk.MenuItem
	fileExportADIF        *gtk.MenuItem
	fileExportCSV         *gtk.MenuItem
	fileExportCallhistory *gtk.MenuItem
	fileSettings          *gtk.MenuItem
	fileQuit              *gtk.MenuItem

	editClearEntryFields *gtk.MenuItem
	editGotoEntryFields  *gtk.MenuItem
	editEditLastQSO      *gtk.MenuItem
	editLogQSO           *gtk.MenuItem
	editSP               *gtk.RadioMenuItem
	editRun              *gtk.RadioMenuItem

	windowCallinfo *gtk.MenuItem
	windowScore    *gtk.MenuItem
	windowRate     *gtk.MenuItem
	helpAbout      *gtk.MenuItem
}

func setupMainMenu(builder *gtk.Builder) *mainMenu {
	result := new(mainMenu)

	result.fileNew = getUI(builder, "menuFileNew").(*gtk.MenuItem)
	result.fileOpen = getUI(builder, "menuFileOpen").(*gtk.MenuItem)
	result.fileSaveAs = getUI(builder, "menuFileSaveAs").(*gtk.MenuItem)
	result.fileExportCabrillo = getUI(builder, "menuFileExportCabrillo").(*gtk.MenuItem)
	result.fileExportADIF = getUI(builder, "menuFileExportADIF").(*gtk.MenuItem)
	result.fileExportCSV = getUI(builder, "menuFileExportCSV").(*gtk.MenuItem)
	result.fileExportCallhistory = getUI(builder, "menuFileExportCallhistory").(*gtk.MenuItem)
	result.fileSettings = getUI(builder, "menuFileSettings").(*gtk.MenuItem)
	result.fileQuit = getUI(builder, "menuFileQuit").(*gtk.MenuItem)
	result.editClearEntryFields = getUI(builder, "menuEditClearEntryFields").(*gtk.MenuItem)
	result.editGotoEntryFields = getUI(builder, "menuEditGotoEntryFields").(*gtk.MenuItem)
	result.editEditLastQSO = getUI(builder, "menuEditEditLastQSO").(*gtk.MenuItem)
	result.editLogQSO = getUI(builder, "menuEditLogQSO").(*gtk.MenuItem)
	result.editSP = getUI(builder, "menuEditSP").(*gtk.RadioMenuItem)
	result.editRun = getUI(builder, "menuEditRun").(*gtk.RadioMenuItem)
	result.windowCallinfo = getUI(builder, "menuWindowCallinfo").(*gtk.MenuItem)
	result.windowScore = getUI(builder, "menuWindowScore").(*gtk.MenuItem)
	result.windowRate = getUI(builder, "menuWindowRate").(*gtk.MenuItem)
	result.helpAbout = getUI(builder, "menuHelpAbout").(*gtk.MenuItem)

	result.fileNew.Connect("activate", result.onNew)
	result.fileOpen.Connect("activate", result.onOpen)
	result.fileSaveAs.Connect("activate", result.onSaveAs)
	result.fileExportCabrillo.Connect("activate", result.onExportCabrillo)
	result.fileExportADIF.Connect("activate", result.onExportADIF)
	result.fileExportCSV.Connect("activate", result.onExportCSV)
	result.fileExportCallhistory.Connect("activate", result.onExportCallhistory)
	result.fileSettings.Connect("activate", result.onSettings)
	result.fileQuit.Connect("activate", result.onQuit)
	result.editClearEntryFields.Connect("activate", result.onClearEntryFields)
	result.editGotoEntryFields.Connect("activate", result.onGotoEntryFields)
	result.editEditLastQSO.Connect("activate", result.onEditLastQSO)
	result.editLogQSO.Connect("activate", result.onLogQSO)
	result.editSP.Connect("toggled", result.onSP)
	result.editRun.Connect("toggled", result.onRun)
	result.windowCallinfo.Connect("activate", result.onCallinfo)
	result.windowScore.Connect("activate", result.onScore)
	result.windowRate.Connect("activate", result.onRate)
	result.helpAbout.Connect("activate", result.onAbout)

	return result
}

func (m *mainMenu) SetMainMenuController(controller MainMenuController) {
	m.controller = controller
}

func (m *mainMenu) WorkmodeChanged(workmode core.Workmode) {
	switch workmode {
	case core.SearchPounce:
		if m.editSP.GetActive() {
			return
		}
		m.editSP.SetActive(true)
	case core.Run:
		if m.editRun.GetActive() {
			return
		}
		m.editRun.SetActive(true)
	}
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

func (m *mainMenu) onExportCallhistory() {
	m.controller.ExportCallhistory()
}

func (m *mainMenu) onSettings() {
	m.controller.OpenSettings()
}

func (m *mainMenu) onQuit() {
	m.controller.Quit()
}

func (m *mainMenu) onClearEntryFields() {
	m.controller.ClearEntryFields()
}

func (m *mainMenu) onGotoEntryFields() {
	m.controller.GotoEntryFields()
}

func (m *mainMenu) onEditLastQSO() {
	m.controller.EditLastQSO()
}

func (m *mainMenu) onLogQSO() {
	m.controller.LogQSO()
}

func (m *mainMenu) onSP() {
	if m.editSP.GetActive() {
		m.controller.SwitchToSPWorkmode()
	}
}

func (m *mainMenu) onRun() {
	if m.editRun.GetActive() {
		m.controller.SwitchToRunWorkmode()
	}
}

func (m *mainMenu) onCallinfo() {
	m.controller.ShowCallinfo()
}

func (m *mainMenu) onScore() {
	m.controller.ShowScore()
}

func (m *mainMenu) onRate() {
	m.controller.ShowRate()
}
