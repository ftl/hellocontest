package fyneui

import (
	"fyne.io/fyne/v2"

	"github.com/ftl/hellocontest/core"
)

type MainMenuController interface {
	Open()
	SaveAs()
	ExportCabrillo()
	ExportADIF()
	ExportCSV()
	ExportCallhistory()
	OpenContestRulesPage()
	OpenContestUploadPage()
	OpenConfigurationFile()
	Quit()

	SwitchToSPWorkmode()
	SwitchToRunWorkmode()

	OpenWiki()
	Sponsors()
	About()

	NotYetImplemented()
}

type ShortcutProvider interface {
	Get(ShortcutID) fyne.Shortcut
}

type mainMenu struct {
	controller MainMenuController
	shortcuts  ShortcutProvider
	root       *fyne.MainMenu

	fileMenu
	editMenu
	radioMenu
	bandmapMenu
	windowMenu
	helpMenu
}

type fileMenu struct {
	fileNew,
	fileOpen,
	fileSaveAs,
	fileExportCabrillo,
	fileExportADIF,
	fileExportCSV,
	fileExportCallhistory,
	fileOpenRules,
	fileOpenUpload,
	fileOpenSettings,
	fileOpenConfigurationFile,
	fileQuit *fyne.MenuItem
}

type editMenu struct {
	editSearchPounce,
	editRun *fyne.MenuItem
}

type radioMenu struct {
}

type bandmapMenu struct {
}

type windowMenu struct {
}

type helpMenu struct {
	helpWiki,
	helpSponsors,
	helpAbout *fyne.MenuItem
}

func setupMainMenu(mainWindow fyne.Window, controller MainMenuController, shortcuts ShortcutProvider) *mainMenu {
	result := &mainMenu{
		controller: controller,
		shortcuts:  shortcuts,
	}

	result.root = fyne.NewMainMenu(
		fyne.NewMenu("File", result.setupFileMenu()...),
		fyne.NewMenu("Edit", result.setupEditMenu()...),
		fyne.NewMenu("Radio", result.setupRadioMenu()...),
		fyne.NewMenu("Bandmap", result.setupBandmapMenu()...),
		fyne.NewMenu("Window", result.setupWindowMenu()...),
		fyne.NewMenu("Help", result.setupHelpMenu()...),
	)
	mainWindow.SetMainMenu(result.root)

	return result
}

// FILE

func (m *mainMenu) setupFileMenu() []*fyne.MenuItem {
	m.fileNew = fyne.NewMenuItem("New...", m.onFileNew)

	m.fileOpen = fyne.NewMenuItem("Open...", m.onFileOpen)
	m.fileOpen.Shortcut = m.shortcuts.Get(OpenShortcut)

	m.fileSaveAs = fyne.NewMenuItem("Save As...", m.onFileSaveAs)
	m.fileExportCabrillo = fyne.NewMenuItem("Export Cabrillo...", m.onFileExportCabrillo)
	m.fileExportADIF = fyne.NewMenuItem("Export ADIF...", m.onFileExportADIF)
	m.fileExportCSV = fyne.NewMenuItem("Export CSV...", m.onFileExportCSV)
	m.fileExportCallhistory = fyne.NewMenuItem("Export Call History...", m.onFileExportCallhistory)
	m.fileOpenRules = fyne.NewMenuItem("Open Contest Rules...", m.onFileOpenRules)
	m.fileOpenUpload = fyne.NewMenuItem("Open Upload Page...", m.onFileOpenUpload)
	m.fileOpenConfigurationFile = fyne.NewMenuItem("Open Configuration File...", m.onFileOpenConfigurationFile)

	m.fileOpenSettings = fyne.NewMenuItem("Settings...", m.onFileOpenSettings)
	m.fileOpenSettings.Shortcut = m.shortcuts.Get(OpenSettingsShortcut)

	m.fileQuit = fyne.NewMenuItem("Quit", m.onFileQuit)
	m.fileQuit.Shortcut = m.shortcuts.Get(QuitShortcut)
	m.fileQuit.IsQuit = true

	return []*fyne.MenuItem{
		m.fileNew,
		m.fileOpen,
		m.fileSaveAs,
		fyne.NewMenuItemSeparator(),
		m.fileExportCabrillo,
		m.fileExportADIF,
		m.fileExportCSV,
		m.fileExportCallhistory,
		fyne.NewMenuItemSeparator(),
		m.fileOpenRules,
		m.fileOpenUpload,
		fyne.NewMenuItemSeparator(),
		m.fileOpenConfigurationFile,
		m.fileOpenSettings,
		fyne.NewMenuItemSeparator(),
		m.fileQuit,
	}
}

func (m *mainMenu) onFileNew() {
	m.controller.NotYetImplemented()
}

func (m *mainMenu) onFileOpen() {
	m.controller.Open()
}

func (m *mainMenu) onFileSaveAs() {
	m.controller.SaveAs()
}

func (m *mainMenu) onFileExportCabrillo() {
	m.controller.ExportCabrillo()
}

func (m *mainMenu) onFileExportADIF() {
	m.controller.ExportADIF()
}

func (m *mainMenu) onFileExportCSV() {
	m.controller.ExportCSV()
}

func (m *mainMenu) onFileExportCallhistory() {
	m.controller.ExportCallhistory()
}

func (m *mainMenu) onFileOpenRules() {
	m.controller.OpenContestRulesPage()
}

func (m *mainMenu) onFileOpenUpload() {
	m.controller.OpenContestUploadPage()
}

func (m *mainMenu) onFileOpenConfigurationFile() {
	m.controller.OpenConfigurationFile()
}

func (m *mainMenu) onFileOpenSettings() {
}

func (m *mainMenu) onFileQuit() {
	m.controller.Quit()
}

// EDIT

func (m *mainMenu) setupEditMenu() []*fyne.MenuItem {
	m.editSearchPounce = fyne.NewMenuItem("Search & Pounce", m.onEditSearchPounce)
	m.editSearchPounce.Shortcut = m.shortcuts.Get(WorkmodeSearchPounceShortcut)
	m.editSearchPounce.Checked = true // this is the default workmode when starting up
	m.editRun = fyne.NewMenuItem("Run", m.onEditRun)
	m.editRun.Shortcut = m.shortcuts.Get(WorkmodeRunShortcut)

	return []*fyne.MenuItem{
		m.editSearchPounce,
		m.editRun,
	}
}

func (m *mainMenu) onEditSearchPounce() {
	m.controller.SwitchToSPWorkmode()
}

func (m *mainMenu) onEditRun() {
	m.controller.SwitchToRunWorkmode()
}

func (m *mainMenu) WorkmodeChanged(workmode core.Workmode) {
	m.editSearchPounce.Checked = (workmode == core.SearchPounce)
	m.editRun.Checked = (workmode == core.Run)
	m.root.Items[1].Refresh() // TODO clean-up the whole menu object structure and use composition instead of embedding
}

// RADIO

func (m *mainMenu) setupRadioMenu() []*fyne.MenuItem {
	return nil
}

// BANDMAP

func (m *mainMenu) setupBandmapMenu() []*fyne.MenuItem {
	return nil
}

// WINDOW

func (m *mainMenu) setupWindowMenu() []*fyne.MenuItem {
	return nil
}

// HELP

func (m *mainMenu) setupHelpMenu() []*fyne.MenuItem {
	m.helpWiki = fyne.NewMenuItem("Wiki", m.onHelpWiki)
	m.helpSponsors = fyne.NewMenuItem("Sponsors", m.onHelpSponsors)
	m.helpAbout = fyne.NewMenuItem("About", m.onHelpAbout)

	return []*fyne.MenuItem{
		m.helpWiki,
		m.helpSponsors,
		m.helpAbout,
	}
}

func (m *mainMenu) onHelpWiki() {
	m.controller.OpenWiki()
}

func (m *mainMenu) onHelpSponsors() {
	m.controller.Sponsors()
}

func (m *mainMenu) onHelpAbout() {
	m.controller.About()
}
