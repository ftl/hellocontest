package ui

import (
	qtlib "github.com/mappu/miqt/qt6"
)

type mainMenu struct {
	menuBar  *qtlib.QMenuBar
	actions  *actions
	fileMenu *qtlib.QMenu
	editMenu *qtlib.QMenu
}

func newMainMenu(
	window *qtlib.QMainWindow,
	a *actions,
	radioMenu *radioMenu,
	spotSourceMenu *spotSourceMenu,
) *mainMenu {
	m := &mainMenu{actions: a}
	m.menuBar = qtlib.NewQMenuBar2()

	// ── File ─────────────────────────────────────────────────────────────
	m.fileMenu = m.menuBar.AddMenuWithTitle("&File")
	m.fileMenu.SetObjectName(*qtlib.NewQAnyStringView3("fileMenu"))
	fileMenu := m.fileMenu
	fileMenu.AddAction(a.newFileAction)
	fileMenu.AddAction(a.openFileAction)
	fileMenu.AddAction(a.saveAsAction)
	fileMenu.AddSeparator()
	exportMenu := fileMenu.AddMenuWithTitle("E&xport")
	exportMenu.AddAction(a.exportSummaryAction)
	exportMenu.AddAction(a.exportCabrilloAction)
	exportMenu.AddAction(a.exportADIFAction)
	exportMenu.AddAction(a.exportCSVAction)
	exportMenu.AddAction(a.exportCallhistoryAction)
	fileMenu.AddSeparator()
	fileMenu.AddAction(a.openRulesAction)
	fileMenu.AddAction(a.openUploadAction)
	fileMenu.AddSeparator()
	fileMenu.AddAction(a.settingsAction)
	fileMenu.AddAction(a.configFileAction)
	fileMenu.AddSeparator()
	fileMenu.AddAction(a.quitAction)

	// ── Edit ─────────────────────────────────────────────────────────────
	m.editMenu = m.menuBar.AddMenuWithTitle("&Edit")
	m.editMenu.SetObjectName(*qtlib.NewQAnyStringView3("editMenu"))
	editMenu := m.editMenu
	editMenu.AddAction(a.clearEntryAction)
	editMenu.AddAction(a.gotoEntryAction)
	editMenu.AddAction(a.editLastAction)
	editMenu.AddAction(a.refreshPredAction)
	editMenu.AddSeparator()
	editMenu.AddAction(a.logQSOAction)
	editMenu.AddAction(a.startParrotAction)
	editMenu.AddAction(a.esmAction)
	editMenu.AddSeparator()
	editMenu.AddAction(a.spAction)
	editMenu.AddAction(a.runAction)
	editMenu.AddSeparator()
	editMenu.AddAction(a.offerQTCAction)
	editMenu.AddAction(a.requestQTCAction)

	// ── Radio ────────────────────────────────────────────────────────────
	radioSubmenu := m.menuBar.AddMenuWithTitle("&Radio")
	radioSubmenu.AddAction(a.xitActiveAction)
	radioSubmenu.AddAction(a.ritActiveAction)
	radioSubmenu.AddSeparator()
	radioMenu.menu = radioSubmenu

	// ── Bandmap ──────────────────────────────────────────────────────────
	bandmapMenu := m.menuBar.AddMenuWithTitle("&Bandmap")
	bandmapMenu.AddAction(a.markBandmapAction)
	bandmapMenu.AddAction(a.highestSpotAction)
	bandmapMenu.AddAction(a.nearestSpotAction)
	bandmapMenu.AddAction(a.nextSpotUpAction)
	bandmapMenu.AddAction(a.nextSpotDownAction)
	bandmapMenu.AddSeparator()
	bandmapMenu.AddAction(a.sendSpotsToTciAction)
	bandmapMenu.AddSeparator()
	spotSourceMenu.menu = bandmapMenu

	// ── Window ───────────────────────────────────────────────────────────
	windowMenu := m.menuBar.AddMenuWithTitle("&Window")
	windowMenu.AddAction(a.showQSOsAction)
	windowMenu.AddAction(a.showQTCsAction)
	windowMenu.AddAction(a.showScoreGraphAction)
	windowMenu.AddAction(a.showScoreTableAction)
	windowMenu.AddAction(a.showRateAction)
	windowMenu.AddAction(a.showSpotsAction)
	windowMenu.AddAction(a.showClockAction)

	// ── Help ─────────────────────────────────────────────────────────────
	helpMenu := m.menuBar.AddMenuWithTitle("&Help")
	helpMenu.AddAction(a.wikiAction)
	helpMenu.AddAction(a.sponsorsAction)
	helpMenu.AddAction(a.aboutAction)

	_ = window
	return m
}
