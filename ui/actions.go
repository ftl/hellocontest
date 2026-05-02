package ui

import (
	"fmt"
	"strings"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/app"
)

const (
	ActionFileNew               = "file.new"
	ActionFileOpen              = "file.open"
	ActionFileSaveAs            = "file.save_as"
	ActionFileExportSummary     = "file.export_summary"
	ActionFileExportCabrillo    = "file.export_cabrillo"
	ActionFileExportADIF        = "file.export_adif"
	ActionFileExportCSV         = "file.export_csv"
	ActionFileExportCallhistory = "file.export_callhistory"
	ActionFileOpenRules         = "file.open_rules"
	ActionFileOpenUpload        = "file.open_upload"
	ActionFileSettings          = "file.settings"
	ActionFileConfigFile        = "file.config_file"
	ActionFileQuit              = "file.quit"

	ActionEntryClear             = "entry.clear"
	ActionEntryGotoEntryField    = "entry.goto_entry_field"
	ActionEntryEditLastQSO       = "entry.edit_last_qso"
	ActionEntryRefreshPrediction = "entry.refresh_prediction"
	ActionEntrySelectBestMatch   = "entry.select_best_match"
	ActionEntryLogQSO            = "entry.log_qso"
	ActionEntryStartParrot       = "entry.start_parrot"
	ActionEntryEnableESM         = "entry.enable_esm"
	ActionEntryNextESMStep       = "entry.next_esm_step"
	ActionEntryWorkmodeSP        = "entry.workmode_sp"
	ActionEntryWorkmodeRun       = "entry.workmode_run"
	ActionEntryOfferQTC          = "entry.offer_qtc"
	ActionEntryRequestQTC        = "entry.request_qtc"

	ActionRadioXITActive = "radio.xit_active"

	ActionBandmapMark             = "bandmap.mark"
	ActionBandmapGotoHighestSpot  = "bandmap.goto_highest_spot"
	ActionBandmapGotoNearestSpot  = "bandmap.goto_nearest_spot"
	ActionBandmapGotoNextSpotUp   = "bandmap.goto_next_spot_up"
	ActionBandmapGotoNextSpotDown = "bandmap.goto_next_spot_down"
	ActionBandmapSendSpotsToTci   = "bandmap.send_spots_to_tci"

	ActionWindowShowQSOs       = "window.show_qsos"
	ActionWindowShowQTCs       = "window.show_qtcs"
	ActionWindowShowScoreGraph = "window.show_score_graph"
	ActionWindowShowScoreTable = "window.show_score_table"
	ActionWindowShowRate       = "window.show_rate"
	ActionWindowShowSpots      = "window.show_spots"

	ActionHelpWiki     = "help.wiki"
	ActionHelpSponsors = "help.sponsors"
	ActionHelpAbout    = "help.about"

	ActionKeyerSendMacro1 = "keyer.send_macro_1"
	ActionKeyerSendMacro2 = "keyer.send_macro_2"
	ActionKeyerSendMacro3 = "keyer.send_macro_3"
	ActionKeyerSendMacro4 = "keyer.send_macro_4"
)

type actions struct {
	parent *qtlib.QWidget

	controller  *app.Controller
	keybindings map[string]string
	ignoreInput bool

	// Action groups
	workModeGroup *qtlib.QActionGroup
	radioGroup    *qtlib.QActionGroup
	keyerGroup    *qtlib.QActionGroup

	// File menu
	newFileAction           *qtlib.QAction
	openFileAction          *qtlib.QAction
	saveAsAction            *qtlib.QAction
	exportSummaryAction     *qtlib.QAction
	exportCabrilloAction    *qtlib.QAction
	exportADIFAction        *qtlib.QAction
	exportCSVAction         *qtlib.QAction
	exportCallhistoryAction *qtlib.QAction
	openRulesAction         *qtlib.QAction
	openUploadAction        *qtlib.QAction
	settingsAction          *qtlib.QAction
	configFileAction        *qtlib.QAction
	quitAction              *qtlib.QAction

	// Edit menu
	clearEntryAction  *qtlib.QAction
	gotoEntryAction   *qtlib.QAction
	editLastAction    *qtlib.QAction
	refreshPredAction *qtlib.QAction
	logQSOAction      *qtlib.QAction
	startParrotAction *qtlib.QAction
	esmAction         *qtlib.QAction
	spAction          *qtlib.QAction
	runAction         *qtlib.QAction
	offerQTCAction    *qtlib.QAction
	requestQTCAction  *qtlib.QAction

	// Radio menu
	xitActiveAction *qtlib.QAction

	// Bandmap menu
	markBandmapAction    *qtlib.QAction
	highestSpotAction    *qtlib.QAction
	nearestSpotAction    *qtlib.QAction
	nextSpotUpAction     *qtlib.QAction
	nextSpotDownAction   *qtlib.QAction
	sendSpotsToTciAction *qtlib.QAction

	// Window menu
	showQSOsAction       *qtlib.QAction
	showQTCsAction       *qtlib.QAction
	showScoreGraphAction *qtlib.QAction
	showScoreTableAction *qtlib.QAction
	showRateAction       *qtlib.QAction
	showSpotsAction      *qtlib.QAction

	// Help menu
	wikiAction     *qtlib.QAction
	sponsorsAction *qtlib.QAction
	aboutAction    *qtlib.QAction

	// Dynamic actions (radio names, keyer names, spot source names)
	radioActions      map[string]*qtlib.QAction
	keyerActions      map[string]*qtlib.QAction
	spotSourceActions map[string]*qtlib.QAction

	// Other actions
	sendMacro1Action      *qtlib.QAction
	sendMacro2Action      *qtlib.QAction
	sendMacro3Action      *qtlib.QAction
	sendMacro4Action      *qtlib.QAction
	selectBestMatchAction *qtlib.QAction
	nextESMStepAction     *qtlib.QAction
}

func newActions(parent *qtlib.QWidget, controller *app.Controller, keybindings map[string]string) *actions {
	a := &actions{
		parent:            parent,
		controller:        controller,
		keybindings:       keybindings,
		radioActions:      make(map[string]*qtlib.QAction),
		keyerActions:      make(map[string]*qtlib.QAction),
		spotSourceActions: make(map[string]*qtlib.QAction),
	}

	// Action groups
	a.workModeGroup = qtlib.NewQActionGroup(parent.QObject)
	a.workModeGroup.SetExclusive(true)
	a.radioGroup = qtlib.NewQActionGroup(parent.QObject)
	a.radioGroup.SetExclusive(true)
	a.keyerGroup = qtlib.NewQActionGroup(parent.QObject)
	a.keyerGroup.SetExclusive(true)

	// File menu
	a.newFileAction = a.makeTriggerAction("&New...", a.shortcutFor(ActionFileNew, "Ctrl+N"), controller.New)
	a.openFileAction = a.makeTriggerAction("&Open...", a.shortcutFor(ActionFileOpen, "Ctrl+O"), controller.Open)
	a.saveAsAction = a.makeTriggerAction("Save &As...", a.shortcutFor(ActionFileSaveAs, ""), controller.SaveAs)
	a.exportSummaryAction = a.makeTriggerAction("&Summary...", a.shortcutFor(ActionFileExportSummary, ""), controller.ExportSummary)
	a.exportCabrilloAction = a.makeTriggerAction("&Cabrillo...", a.shortcutFor(ActionFileExportCabrillo, ""), controller.ExportCabrillo)
	a.exportADIFAction = a.makeTriggerAction("&ADIF...", a.shortcutFor(ActionFileExportADIF, ""), controller.ExportADIF)
	a.exportCSVAction = a.makeTriggerAction("C&SV...", a.shortcutFor(ActionFileExportCSV, ""), controller.ExportCSV)
	a.exportCallhistoryAction = a.makeTriggerAction("Call &History...", a.shortcutFor(ActionFileExportCallhistory, ""), controller.ExportCallhistory)
	a.openRulesAction = a.makeTriggerAction("Open Contest &Rules Page", a.shortcutFor(ActionFileOpenRules, ""), controller.OpenContestRulesPage)
	a.openUploadAction = a.makeTriggerAction("Open Contest &Upload Page", a.shortcutFor(ActionFileOpenUpload, ""), controller.OpenContestUploadPage)
	a.settingsAction = a.makeTriggerAction("Contest &Settings...", a.shortcutFor(ActionFileSettings, "Ctrl+."), controller.OpenSettings)
	a.configFileAction = a.makeTriggerAction("Open &Configuration File...", a.shortcutFor(ActionFileConfigFile, "Ctrl+Shift+."), controller.OpenConfigurationFile)
	a.quitAction = a.makeTriggerAction("&Quit", a.shortcutFor(ActionFileQuit, "Ctrl+Q"), controller.Quit)

	// Edit menu
	a.clearEntryAction = a.makeTriggerAction("&Clear Entry Fields", a.shortcutFor(ActionEntryClear, ""), controller.ClearEntryFields)
	a.gotoEntryAction = a.makeTriggerAction("&Goto Entry Fields", a.shortcutFor(ActionEntryGotoEntryField, "Ctrl+E"), controller.GotoEntryFields)
	a.editLastAction = a.makeTriggerAction("Edit &Last QSO", a.shortcutFor(ActionEntryEditLastQSO, "Ctrl+L"), controller.EditLastQSO)
	a.refreshPredAction = a.makeTriggerAction("Refresh &Prediction", a.shortcutFor(ActionEntryRefreshPrediction, "Ctrl+P"), controller.RefreshPrediction)
	a.logQSOAction = a.makeTriggerAction("Log QSO", a.shortcutFor(ActionEntryLogQSO, "Ctrl+Return"), controller.LogQSO)
	a.startParrotAction = a.makeTriggerAction("Start Parrot", a.shortcutFor(ActionEntryStartParrot, "Ctrl+F1"), controller.StartParrot)
	a.esmAction = a.makeCheckAction("ESM", a.shortcutFor(ActionEntryEnableESM, "Ctrl+Shift+M"), controller.SetESMEnabled)
	a.spAction = a.makeToggleAction("S&P", a.shortcutFor(ActionEntryWorkmodeSP, "Ctrl+S"), controller.SwitchToSPWorkmode)
	a.workModeGroup.AddAction(a.spAction)
	a.runAction = a.makeToggleAction("Run", a.shortcutFor(ActionEntryWorkmodeRun, "Ctrl+R"), controller.SwitchToRunWorkmode)
	a.workModeGroup.AddAction(a.runAction)
	a.offerQTCAction = a.makeTriggerAction("Offer QTC", a.shortcutFor(ActionEntryOfferQTC, "F5"), controller.OfferQTC)
	a.requestQTCAction = a.makeTriggerAction("Request QTC", a.shortcutFor(ActionEntryRequestQTC, "F6"), controller.RequestQTC)

	// Radio menu
	a.xitActiveAction = a.makeCheckAction("XIT Active", a.shortcutFor(ActionRadioXITActive, "Ctrl+Shift+X"), controller.SetXITActive)

	// Bandmap menu
	a.markBandmapAction = a.makeTriggerAction("Mark In Bandmap", a.shortcutFor(ActionBandmapMark, "Ctrl+M"), controller.MarkInBandmap)
	a.highestSpotAction = a.makeTriggerAction("Goto Highest Value Spot", a.shortcutFor(ActionBandmapGotoHighestSpot, "Ctrl+Shift+N"), controller.GotoHighestValueSpot)
	a.nearestSpotAction = a.makeTriggerAction("Goto Nearest Spot", a.shortcutFor(ActionBandmapGotoNearestSpot, "Ctrl+N"), controller.GotoNearestSpot)
	a.nextSpotUpAction = a.makeTriggerAction("Goto Next Spot Up", a.shortcutFor(ActionBandmapGotoNextSpotUp, "Ctrl+Up"), controller.GotoNextSpotUp)
	a.nextSpotDownAction = a.makeTriggerAction("Goto Next Spot Down", a.shortcutFor(ActionBandmapGotoNextSpotDown, "Ctrl+Down"), controller.GotoNextSpotDown)
	a.sendSpotsToTciAction = a.makeCheckAction("Send Spots to TCI", a.shortcutFor(ActionBandmapSendSpotsToTci, ""), controller.SetSendSpotsToTci)

	// Window menu
	a.showQSOsAction = a.makeTriggerAction("&QSOs", a.shortcutFor(ActionWindowShowQSOs, ""), controller.ShowQSOs)
	a.showQTCsAction = a.makeTriggerAction("Q&TCs", a.shortcutFor(ActionWindowShowQTCs, ""), controller.ShowQTCs)
	a.showScoreGraphAction = a.makeTriggerAction("Score &Graph", a.shortcutFor(ActionWindowShowScoreGraph, ""), controller.ShowScoreGraph)
	a.showScoreTableAction = a.makeTriggerAction("&Score Table", a.shortcutFor(ActionWindowShowScoreTable, ""), controller.ShowScoreTable)
	a.showRateAction = a.makeTriggerAction("&Rate", a.shortcutFor(ActionWindowShowRate, ""), controller.ShowRate)
	a.showSpotsAction = a.makeTriggerAction("S&pots", a.shortcutFor(ActionWindowShowSpots, ""), controller.ShowSpots)

	// Help menu
	a.wikiAction = a.makeTriggerAction("&Wiki", a.shortcutFor(ActionHelpWiki, ""), controller.OpenWiki)
	a.sponsorsAction = a.makeTriggerAction("&Sponsors", a.shortcutFor(ActionHelpSponsors, ""), controller.Sponsors)
	a.aboutAction = a.makeTriggerAction("&About", a.shortcutFor(ActionHelpAbout, ""), controller.About)

	// Other actions
	a.sendMacro1Action = a.makeTriggerAction("Macro 1", a.shortcutFor(ActionKeyerSendMacro1, "F1"), func() { controller.Keyer.Send(0) })
	a.sendMacro2Action = a.makeTriggerAction("Macro 2", a.shortcutFor(ActionKeyerSendMacro2, "F2"), func() { controller.Keyer.Send(1) })
	a.sendMacro3Action = a.makeTriggerAction("Macro 3", a.shortcutFor(ActionKeyerSendMacro3, "F3"), func() { controller.Keyer.Send(2) })
	a.sendMacro4Action = a.makeTriggerAction("Macro 4", a.shortcutFor(ActionKeyerSendMacro4, "F4"), func() { controller.Keyer.Send(3) })
	a.selectBestMatchAction = a.makeTriggerAction("Select Best Match", a.shortcutFor(ActionEntrySelectBestMatch, "Alt+Return"), controller.Entry.SelectBestMatchOnFrequency)
	a.nextESMStepAction = a.makeTriggerAction("Next ESM Step", a.shortcutFor(ActionEntryNextESMStep, ""), controller.Entry.NextESMStep)
	a.parent.AddActions([]*qtlib.QAction{
		a.sendMacro1Action,
		a.sendMacro2Action,
		a.sendMacro3Action,
		a.sendMacro4Action,
		a.selectBestMatchAction,
		a.nextESMStepAction,
	})

	// setup initial action state from controller
	a.updateFromController()

	return a
}

func (a *actions) makeTriggerAction(text, shortcut string, handler func()) *qtlib.QAction {
	action := a.makeAction(text, shortcut)
	action.OnTriggered(handler)
	return action
}

func (a *actions) makeCheckAction(text, shortcut string, handler func(bool)) *qtlib.QAction {
	action := a.makeAction(text, shortcut)
	action.SetCheckable(true)
	action.OnToggled(func(checked bool) {
		if a.ignoreInput {
			return
		}
		handler(checked)
	})
	return action
}

func (a *actions) makeToggleAction(text, shortcut string, handler func()) *qtlib.QAction {
	action := a.makeAction(text, shortcut)
	action.SetCheckable(true)
	action.OnToggled(func(checked bool) {
		if a.ignoreInput || !checked {
			return
		}
		handler()
	})
	return action
}

func (a *actions) makeAction(text, shortcut string) *qtlib.QAction {
	action := qtlib.NewQAction()
	action.SetText(text)
	if shortcut != "" {
		action.SetShortcut(qtlib.NewQKeySequence2(shortcut))
	}
	return action
}

func (a *actions) shortcutFor(id, defaultShortcut string) string {
	if s, ok := a.keybindings[id]; ok && s != "" {
		return s
	}
	return defaultShortcut
}

func (a *actions) updateFromController() {
	a.ignoreInput = true

	a.esmAction.SetChecked(a.controller.ESMEnabled())
	a.xitActiveAction.SetChecked(a.controller.XITActive())
	a.sendSpotsToTciAction.SetChecked(a.controller.SendSpotsToTci())
	switch a.controller.Workmode.Workmode() {
	case core.SearchPounce:
		a.spAction.SetChecked(true)
	case core.Run:
		a.runAction.SetChecked(true)
	}
	qtcsEnabled := a.controller.QTCList.QTCsEnabled()
	a.offerQTCAction.SetEnabled(qtcsEnabled)
	a.requestQTCAction.SetEnabled(qtcsEnabled)
	a.showQTCsAction.SetEnabled(qtcsEnabled)

	a.ignoreInput = false
}

// Dynamic action management

func (a *actions) AddRadioAction(name string) *qtlib.QAction {
	normalizedName := strings.ToLower(name)
	if existing, ok := a.radioActions[normalizedName]; ok {
		return existing
	}
	action := qtlib.NewQAction()
	action.SetText(name)
	action.SetCheckable(true)
	a.radioGroup.AddAction(action)
	a.radioActions[normalizedName] = action
	action.OnToggled(func(checked bool) {
		if a.ignoreInput || !checked {
			return
		}
		a.controller.SelectRadio(normalizedName)
	})
	return action
}

func (a *actions) AddKeyerAction(name string) *qtlib.QAction {
	normalizedName := strings.ToLower(name)
	if existing, ok := a.keyerActions[normalizedName]; ok {
		return existing
	}
	action := qtlib.NewQAction()
	action.SetText(name)
	action.SetCheckable(true)
	a.keyerGroup.AddAction(action)
	a.keyerActions[normalizedName] = action
	action.OnToggled(func(checked bool) {
		if a.ignoreInput || !checked {
			return
		}
		a.controller.SelectKeyer(normalizedName)
	})
	return action
}

func (a *actions) AddSpotSourceAction(name string) *qtlib.QAction {
	normalizedName := strings.ToLower(name)
	if existing, ok := a.spotSourceActions[normalizedName]; ok {
		return existing
	}
	action := qtlib.NewQAction()
	action.SetText(fmt.Sprintf("Use %s", name))
	action.SetCheckable(true)
	a.spotSourceActions[normalizedName] = action
	action.OnToggled(func(checked bool) {
		if a.ignoreInput {
			return
		}
		a.controller.Clusters.SetSpotSourceEnabled(normalizedName, checked)
	})
	return action
}

func (a *actions) SetRadioSelected(name string) {
	a.ignoreInput = true
	defer func() { a.ignoreInput = false }()

	name = strings.ToLower(name)
	for n, action := range a.radioActions {
		action.SetChecked(n == name)
	}
}

func (a *actions) SetKeyerSelected(name string) {
	a.ignoreInput = true
	defer func() { a.ignoreInput = false }()

	name = strings.ToLower(name)
	for n, action := range a.keyerActions {
		action.SetChecked(n == name)
	}
}

func (a *actions) SetSpotSourceEnabled(name string, enabled bool) {
	a.ignoreInput = true
	defer func() { a.ignoreInput = false }()

	name = strings.ToLower(name)
	action, ok := a.spotSourceActions[name]
	if !ok {
		return
	}
	action.SetChecked(enabled)
}

// Listener-side state updates

func (a *actions) WorkmodeChanged(wm core.Workmode) {
	a.ignoreInput = true
	defer func() { a.ignoreInput = false }()
	switch wm {
	case core.SearchPounce:
		a.spAction.SetChecked(true)
	case core.Run:
		a.runAction.SetChecked(true)
	}
}

func (a *actions) ESMEnabled(enabled bool) {
	a.ignoreInput = true
	defer func() { a.ignoreInput = false }()
	a.esmAction.SetChecked(enabled)
}

func (a *actions) XITActiveChanged(active bool) {
	a.ignoreInput = true
	defer func() { a.ignoreInput = false }()
	a.xitActiveAction.SetChecked(active)
}

func (a *actions) ContestPagesChanged(rulesAvailable, uploadAvailable bool) {
	a.openRulesAction.SetEnabled(rulesAvailable)
	a.openUploadAction.SetEnabled(uploadAvailable)
}

func (a *actions) SetQTCsEnabled(enabled bool) {
	a.offerQTCAction.SetEnabled(enabled)
	a.requestQTCAction.SetEnabled(enabled)
	a.showQTCsAction.SetEnabled(enabled)
}
