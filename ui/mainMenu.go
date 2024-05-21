package ui

import (
	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

// MainMenuController provides the functionality for the main menu.
type MainMenuController interface {
	OpenWiki()
	Sponsors()
	About()
	New()
	Open()
	SaveAs()
	ExportCabrillo()
	ExportADIF()
	ExportCSV()
	ExportCallhistory()
	OpenContestRulesPage()
	OpenContestUploadPage()
	OpenSettings()
	OpenConfigurationFile()
	Quit()
	ShowScore()
	ShowRate()
	ShowSpots()
	ClearEntryFields()
	GotoEntryFields()
	EditLastQSO()
	RefreshPrediction()
	LogQSO()
	SwitchToSPWorkmode()
	SwitchToRunWorkmode()
	MarkInBandmap()
	GotoHighestValueSpot()
	GotoNearestSpot()
	GotoNextSpotUp()
	GotoNextSpotDown()
	SendSpotsToTci() bool
	SetSendSpotsToTci(bool)
}

type AcceptFocusFunc func(bool)

type mainMenu struct {
	controller     MainMenuController
	setAcceptFocus AcceptFocusFunc

	fileNew               *gtk.MenuItem
	fileOpen              *gtk.MenuItem
	fileSaveAs            *gtk.MenuItem
	fileExportCabrillo    *gtk.MenuItem
	fileExportADIF        *gtk.MenuItem
	fileExportCSV         *gtk.MenuItem
	fileExportCallhistory *gtk.MenuItem
	fileOpenRules         *gtk.MenuItem
	fileOpenUpload        *gtk.MenuItem
	fileSettings          *gtk.MenuItem
	fileConfiguration     *gtk.MenuItem
	fileQuit              *gtk.MenuItem

	editClearEntryFields  *gtk.MenuItem
	editGotoEntryFields   *gtk.MenuItem
	editEditLastQSO       *gtk.MenuItem
	editRefreshPrediction *gtk.MenuItem
	editLogQSO            *gtk.MenuItem
	editSP                *gtk.RadioMenuItem
	editRun               *gtk.RadioMenuItem

	bandmapMark                 *gtk.MenuItem
	bandmapGotoHighestValueSpot *gtk.MenuItem
	bandmapGotoNearestSpot      *gtk.MenuItem
	bandmapGotoNextSpotUp       *gtk.MenuItem
	bandmapGotoNextSpotDown     *gtk.MenuItem
	bandmapSendSpotsToTci       *gtk.CheckMenuItem

	windowScore       *gtk.MenuItem
	windowRate        *gtk.MenuItem
	windowSpots       *gtk.MenuItem
	windowAcceptFocus *gtk.CheckMenuItem

	helpWiki     *gtk.MenuItem
	helpSponsors *gtk.MenuItem
	helpAbout    *gtk.MenuItem
}

func setupMainMenu(builder *gtk.Builder, setAcceptFocus AcceptFocusFunc) *mainMenu {
	result := new(mainMenu)
	result.setAcceptFocus = setAcceptFocus

	result.fileNew = getUI(builder, "menuFileNew").(*gtk.MenuItem)
	result.fileOpen = getUI(builder, "menuFileOpen").(*gtk.MenuItem)
	result.fileSaveAs = getUI(builder, "menuFileSaveAs").(*gtk.MenuItem)
	result.fileExportCabrillo = getUI(builder, "menuFileExportCabrillo").(*gtk.MenuItem)
	result.fileExportADIF = getUI(builder, "menuFileExportADIF").(*gtk.MenuItem)
	result.fileExportCSV = getUI(builder, "menuFileExportCSV").(*gtk.MenuItem)
	result.fileExportCallhistory = getUI(builder, "menuFileExportCallhistory").(*gtk.MenuItem)
	result.fileOpenRules = getUI(builder, "menuFileOpenRules").(*gtk.MenuItem)
	result.fileOpenUpload = getUI(builder, "menuFileOpenUpload").(*gtk.MenuItem)
	result.fileSettings = getUI(builder, "menuFileSettings").(*gtk.MenuItem)
	result.fileConfiguration = getUI(builder, "menuFileConfiguration").(*gtk.MenuItem)
	result.fileQuit = getUI(builder, "menuFileQuit").(*gtk.MenuItem)
	result.editClearEntryFields = getUI(builder, "menuEditClearEntryFields").(*gtk.MenuItem)
	result.editGotoEntryFields = getUI(builder, "menuEditGotoEntryFields").(*gtk.MenuItem)
	result.editEditLastQSO = getUI(builder, "menuEditEditLastQSO").(*gtk.MenuItem)
	result.editRefreshPrediction = getUI(builder, "menuEditRefreshPrediction").(*gtk.MenuItem)
	result.editLogQSO = getUI(builder, "menuEditLogQSO").(*gtk.MenuItem)
	result.editSP = getUI(builder, "menuEditSP").(*gtk.RadioMenuItem)
	result.editRun = getUI(builder, "menuEditRun").(*gtk.RadioMenuItem)
	result.bandmapMark = getUI(builder, "menuBandmapMark").(*gtk.MenuItem)
	result.bandmapGotoHighestValueSpot = getUI(builder, "menuBandmapGotoHighestValueSpot").(*gtk.MenuItem)
	result.bandmapGotoNearestSpot = getUI(builder, "menuBandmapGotoNearestSpot").(*gtk.MenuItem)
	result.bandmapGotoNextSpotUp = getUI(builder, "menuBandmapGotoNextSpotUp").(*gtk.MenuItem)
	result.bandmapGotoNextSpotDown = getUI(builder, "menuBandmapGotoNextSpotDown").(*gtk.MenuItem)
	result.bandmapSendSpotsToTci = getUI(builder, "menuBandmapSendSpotsToTci").(*gtk.CheckMenuItem)
	result.windowScore = getUI(builder, "menuWindowScore").(*gtk.MenuItem)
	result.windowRate = getUI(builder, "menuWindowRate").(*gtk.MenuItem)
	result.windowSpots = getUI(builder, "menuWindowSpots").(*gtk.MenuItem)
	result.windowAcceptFocus = getUI(builder, "menuWindowAcceptFocus").(*gtk.CheckMenuItem)
	result.helpWiki = getUI(builder, "menuHelpWiki").(*gtk.MenuItem)
	result.helpSponsors = getUI(builder, "menuHelpSponsors").(*gtk.MenuItem)
	result.helpAbout = getUI(builder, "menuHelpAbout").(*gtk.MenuItem)

	result.fileNew.Connect("activate", result.onNew)
	result.fileOpen.Connect("activate", result.onOpen)
	result.fileSaveAs.Connect("activate", result.onSaveAs)
	result.fileExportCabrillo.Connect("activate", result.onExportCabrillo)
	result.fileExportADIF.Connect("activate", result.onExportADIF)
	result.fileExportCSV.Connect("activate", result.onExportCSV)
	result.fileExportCallhistory.Connect("activate", result.onExportCallhistory)
	result.fileOpenRules.Connect("activate", result.onOpenRules)
	result.fileOpenUpload.Connect("activate", result.onOpenUpload)
	result.fileSettings.Connect("activate", result.onSettings)
	result.fileConfiguration.Connect("activate", result.onConfiguration)
	result.fileQuit.Connect("activate", result.onQuit)
	result.editClearEntryFields.Connect("activate", result.onClearEntryFields)
	result.editGotoEntryFields.Connect("activate", result.onGotoEntryFields)
	result.editEditLastQSO.Connect("activate", result.onEditLastQSO)
	result.editRefreshPrediction.Connect("activate", result.onEditRefreshPrediction)
	result.editLogQSO.Connect("activate", result.onLogQSO)
	result.editSP.Connect("toggled", result.onSP)
	result.editRun.Connect("toggled", result.onRun)
	result.bandmapMark.Connect("activate", result.onMarkInBandmap)
	result.bandmapGotoHighestValueSpot.Connect("activate", result.onGotoHighestValueSpot)
	result.bandmapGotoNearestSpot.Connect("activate", result.onGotoNearestSpot)
	result.bandmapGotoNextSpotUp.Connect("activate", result.onGotoNextSpotUp)
	result.bandmapGotoNextSpotDown.Connect("activate", result.onGotoNextSpotDown)
	result.bandmapSendSpotsToTci.Connect("toggled", result.onSendSpotsToTci)
	result.windowScore.Connect("activate", result.onScore)
	result.windowRate.Connect("activate", result.onRate)
	result.windowSpots.Connect("activate", result.onSpots)
	result.windowAcceptFocus.Connect("activate", result.onAcceptFocus)
	result.helpWiki.Connect("activate", result.onWiki)
	result.helpSponsors.Connect("activate", result.onSponsors)
	result.helpAbout.Connect("activate", result.onAbout)

	return result
}

func (m *mainMenu) SetMainMenuController(controller MainMenuController) {
	m.bandmapSendSpotsToTci.SetActive(controller.SendSpotsToTci())
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

func (m *mainMenu) ContestPagesChanged(rulesAvailable bool, uploadAvailable bool) {
	m.fileOpenRules.SetSensitive(rulesAvailable)
	m.fileOpenUpload.SetSensitive(uploadAvailable)
}

func (m *mainMenu) onWiki() {
	m.controller.OpenWiki()
}

func (m *mainMenu) onSponsors() {
	m.controller.Sponsors()
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

func (m *mainMenu) onOpenRules() {
	m.controller.OpenContestRulesPage()
}

func (m *mainMenu) onOpenUpload() {
	m.controller.OpenContestUploadPage()
}

func (m *mainMenu) onSettings() {
	m.controller.OpenSettings()
}

func (m *mainMenu) onConfiguration() {
	m.controller.OpenConfigurationFile()
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

func (m *mainMenu) onEditRefreshPrediction() {
	m.controller.RefreshPrediction()
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

func (m *mainMenu) onMarkInBandmap() {
	m.controller.MarkInBandmap()
}

func (m *mainMenu) onGotoHighestValueSpot() {
	m.controller.GotoHighestValueSpot()
}

func (m *mainMenu) onGotoNearestSpot() {
	m.controller.GotoNearestSpot()
}

func (m *mainMenu) onGotoNextSpotUp() {
	m.controller.GotoNextSpotUp()
}

func (m *mainMenu) onGotoNextSpotDown() {
	m.controller.GotoNextSpotDown()
}

func (m *mainMenu) onSendSpotsToTci() {
	if m.controller == nil {
		return
	}
	m.controller.SetSendSpotsToTci(m.bandmapSendSpotsToTci.GetActive())
}

func (m *mainMenu) onScore() {
	m.controller.ShowScore()
}

func (m *mainMenu) onRate() {
	m.controller.ShowRate()
}

func (m *mainMenu) onSpots() {
	m.controller.ShowSpots()
}

func (m *mainMenu) onAcceptFocus() {
	if m.setAcceptFocus == nil {
		return
	}
	m.setAcceptFocus(m.windowAcceptFocus.GetActive())
}
