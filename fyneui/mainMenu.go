package fyneui

import "fyne.io/fyne/v2"

type MainMenuController interface {
	Open()
	Quit()

	OpenWiki()
	Sponsors()
	About()
}

type mainMenu struct {
	controller MainMenuController

	fileMenu
	editMenu
	bandmapMenu
	windowMenu
	helpMenu
}

type fileMenu struct {
	fileOpen *fyne.MenuItem
	fileQuit *fyne.MenuItem
}

type editMenu struct {
}

type bandmapMenu struct {
}

type windowMenu struct {
}

type helpMenu struct {
	helpWiki     *fyne.MenuItem
	helpSponsors *fyne.MenuItem
	helpAbout    *fyne.MenuItem
}

func setupMainMenu(mainWindow fyne.Window, controller MainMenuController) *mainMenu {
	result := &mainMenu{
		controller: controller,
	}

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File", result.setupFileMenu()...),
		fyne.NewMenu("Edit", result.setupEditMenu()...),
		fyne.NewMenu("Bandmap", result.setupBandmapMenu()...),
		fyne.NewMenu("Window", result.setupWindowMenu()...),
		fyne.NewMenu("Help", result.setupHelpMenu()...),
	)
	mainWindow.SetMainMenu(mainMenu)

	return result
}

// FILE

func (m *mainMenu) setupFileMenu() []*fyne.MenuItem {
	m.fileOpen = fyne.NewMenuItem("Open...", m.onFileOpen)
	m.fileQuit = fyne.NewMenuItem("Quit", m.onFileQuit)
	m.fileQuit.IsQuit = true

	return []*fyne.MenuItem{
		m.fileOpen,
		m.fileQuit,
	}
}

func (m *mainMenu) onFileOpen() {
	m.controller.Open()
}

func (m *mainMenu) onFileQuit() {
	m.controller.Quit()
}

// EDIT

func (m *mainMenu) setupEditMenu() []*fyne.MenuItem {
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
