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

	ActionBandmapMark                 = "bandmap.mark"
	ActionBandmapGotoHighestValueSpot = "bandmap.goto_highest_value_spot"
	ActionBandmapGotoNearestSpot      = "bandmap.goto_nearest_spot"
	ActionBandmapGotoNextSpotUp       = "bandmap.goto_next_spot_up"
	ActionBandmapGotoNextSpotDown     = "bandmap.goto_next_spot_down"
	ActionBandmapSendSpotsToTci       = "bandmap.send_spots_to_tci"

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
	allInfos    []ActionInfo

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
	a.newFileAction = a.makeTriggerAction(ActionFileNew, "&New...", "Create a new contest log", "Ctrl+N", controller.New)
	a.openFileAction = a.makeTriggerAction(ActionFileOpen, "&Open...", "Open a contest log", "Ctrl+O", controller.Open)
	a.saveAsAction = a.makeTriggerAction(ActionFileSaveAs, "Save &As...", "", "", controller.SaveAs)
	a.exportSummaryAction = a.makeTriggerAction(ActionFileExportSummary, "&Summary...", "Export the contest summary", "", controller.ExportSummary)
	a.exportCabrilloAction = a.makeTriggerAction(ActionFileExportCabrillo, "&Cabrillo...", "Export the log as Cabrillo file", "", controller.ExportCabrillo)
	a.exportADIFAction = a.makeTriggerAction(ActionFileExportADIF, "&ADIF...", "Export the log as ADIF file", "", controller.ExportADIF)
	a.exportCSVAction = a.makeTriggerAction(ActionFileExportCSV, "C&SV...", "Export the log as CSV file", "", controller.ExportCSV)
	a.exportCallhistoryAction = a.makeTriggerAction(ActionFileExportCallhistory, "Call &History...", "Export a call history file based on the contest log", "", controller.ExportCallhistory)
	a.openRulesAction = a.makeTriggerAction(ActionFileOpenRules, "Open Contest &Rules Page", "", "", controller.OpenContestRulesPage)
	a.openUploadAction = a.makeTriggerAction(ActionFileOpenUpload, "Open Contest &Upload Page", "", "", controller.OpenContestUploadPage)
	a.settingsAction = a.makeTriggerAction(ActionFileSettings, "Contest &Settings...", "", "Ctrl+.", controller.OpenSettings)
	a.configFileAction = a.makeTriggerAction(ActionFileConfigFile, "Open &Configuration File...", "Open the configuration file in the external editor", "Ctrl+Shift+.", controller.OpenConfigurationFile)
	a.quitAction = a.makeTriggerAction(ActionFileQuit, "&Quit", "", "Ctrl+Q", controller.Quit)

	// Edit menu
	a.clearEntryAction = a.makeTriggerAction(ActionEntryClear, "&Clear Entry Fields", "Clear the entry fields", "", controller.ClearEntryFields)
	a.gotoEntryAction = a.makeTriggerAction(ActionEntryGotoEntryField, "&Goto Entry Fields", "Focus the entry field", "Ctrl+E", controller.GotoEntryFields)
	a.editLastAction = a.makeTriggerAction(ActionEntryEditLastQSO, "Edit &Last QSO", "Select the last QSO for editing", "Ctrl+L", controller.EditLastQSO)
	a.refreshPredAction = a.makeTriggerAction(ActionEntryRefreshPrediction, "Refresh &Prediction", "Refresh the exchange prediction", "Ctrl+P", controller.RefreshPrediction)
	a.logQSOAction = a.makeTriggerAction(ActionEntryLogQSO, "Log QSO", "", "Ctrl+Return", controller.LogQSO)
	a.startParrotAction = a.makeTriggerAction(ActionEntryStartParrot, "Start Parrot", "Start the keyer parrot", "Ctrl+F1", controller.StartParrot)
	a.esmAction = a.makeCheckAction(ActionEntryEnableESM, "ESM", "Enable ESM", "Ctrl+Shift+M", controller.SetESMEnabled)
	a.spAction = a.makeToggleAction(ActionEntryWorkmodeSP, "S&P", "", "Ctrl+S", controller.SwitchToSPWorkmode)
	a.workModeGroup.AddAction(a.spAction)
	a.runAction = a.makeToggleAction(ActionEntryWorkmodeRun, "Run", "", "Ctrl+R", controller.SwitchToRunWorkmode)
	a.workModeGroup.AddAction(a.runAction)
	a.offerQTCAction = a.makeTriggerAction(ActionEntryOfferQTC, "Offer QTC", "", "F5", controller.OfferQTC)
	a.requestQTCAction = a.makeTriggerAction(ActionEntryRequestQTC, "Request QTC", "", "F6", controller.RequestQTC)

	// Radio menu
	a.xitActiveAction = a.makeCheckAction(ActionRadioXITActive, "XIT Active", "Activate the XIT for the search & pounce workmode", "Ctrl+Shift+X", controller.SetXITActive)

	// Bandmap menu
	a.markBandmapAction = a.makeTriggerAction(ActionBandmapMark, "Mark In Bandmap", "", "Ctrl+M", controller.MarkInBandmap)
	a.highestSpotAction = a.makeTriggerAction(ActionBandmapGotoHighestValueSpot, "Goto Highest Value Spot", "Go to the spot with the highest predicted value", "Ctrl+Shift+N", controller.GotoHighestValueSpot)
	a.nearestSpotAction = a.makeTriggerAction(ActionBandmapGotoNearestSpot, "Goto Nearest Spot", "Go to the spot closest to the current frequency", "Ctrl+N", controller.GotoNearestSpot)
	a.nextSpotUpAction = a.makeTriggerAction(ActionBandmapGotoNextSpotUp, "Goto Next Spot Up", "", "Ctrl+Up", controller.GotoNextSpotUp)
	a.nextSpotDownAction = a.makeTriggerAction(ActionBandmapGotoNextSpotDown, "Goto Next Spot Down", "", "Ctrl+Down", controller.GotoNextSpotDown)
	a.sendSpotsToTciAction = a.makeCheckAction(ActionBandmapSendSpotsToTci, "Send Spots to TCI", "", "", controller.SetSendSpotsToTci)

	// Window menu
	a.showQSOsAction = a.makeTriggerAction(ActionWindowShowQSOs, "&QSOs", "Show the QSOs list", "", controller.ShowQSOs)
	a.showQTCsAction = a.makeTriggerAction(ActionWindowShowQTCs, "Q&TCs", "Show the QTCs list", "", controller.ShowQTCs)
	a.showScoreGraphAction = a.makeTriggerAction(ActionWindowShowScoreGraph, "Score &Graph", "Show the score graph", "", controller.ShowScoreGraph)
	a.showScoreTableAction = a.makeTriggerAction(ActionWindowShowScoreTable, "&Score Table", "Show the score table", "", controller.ShowScoreTable)
	a.showRateAction = a.makeTriggerAction(ActionWindowShowRate, "&Rate", "Show the QSO rate", "", controller.ShowRate)
	a.showSpotsAction = a.makeTriggerAction(ActionWindowShowSpots, "S&pots", "Show the spots table", "", controller.ShowSpots)

	// Help menu
	a.wikiAction = a.makeTriggerAction(ActionHelpWiki, "&Wiki", "Open the manual in the webbrowser", "", controller.OpenWiki)
	a.sponsorsAction = a.makeTriggerAction(ActionHelpSponsors, "&Sponsors", "Open the sponsors page in the webbrowser", "", controller.Sponsors)
	a.aboutAction = a.makeTriggerAction(ActionHelpAbout, "&About", "Show the about dialog", "", controller.About)

	// Other actions
	a.sendMacro1Action = a.makeTriggerAction(ActionKeyerSendMacro1, "Send Macro 1", "Send the keyer macro #1", "F1", func() { controller.Keyer.Send(0) })
	a.sendMacro2Action = a.makeTriggerAction(ActionKeyerSendMacro2, "Send Macro 2", "Send the keyer macro #2", "F2", func() { controller.Keyer.Send(1) })
	a.sendMacro3Action = a.makeTriggerAction(ActionKeyerSendMacro3, "Send Macro 3", "Send the keyer macro #3", "F3", func() { controller.Keyer.Send(2) })
	a.sendMacro4Action = a.makeTriggerAction(ActionKeyerSendMacro4, "Send Macro 4", "Send the keyer macro #4", "F4", func() { controller.Keyer.Send(3) })
	a.selectBestMatchAction = a.makeTriggerAction(ActionEntrySelectBestMatch, "Select Best Match", "Select the best matching callsign", "Alt+Return", func() { controller.Entry.SelectBestMatchOnFrequency() })
	a.nextESMStepAction = a.makeTriggerAction(ActionEntryNextESMStep, "Next ESM Step", "Execute the next ESM step", "", func() { controller.Entry.NextESMStep() })
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

func (a *actions) makeTriggerAction(id, text, tooltip, defaultShortcut string, handler func()) *qtlib.QAction {
	a.allInfos = append(a.allInfos, ActionInfo{ID: id, Name: stripAccelerator(text), Tooltip: tooltip, DefaultShortcut: defaultShortcut})
	action := a.makeAction(text, a.shortcutFor(id, defaultShortcut), tooltip)
	if handler != nil {
		action.OnTriggered(handler)
	}
	return action
}

func (a *actions) makeCheckAction(id, text, tooltip, defaultShortcut string, handler func(bool)) *qtlib.QAction {
	a.allInfos = append(a.allInfos, ActionInfo{ID: id, Name: stripAccelerator(text), Tooltip: tooltip, DefaultShortcut: defaultShortcut})
	action := a.makeAction(text, a.shortcutFor(id, defaultShortcut), tooltip)
	action.SetCheckable(true)
	action.OnToggled(func(checked bool) {
		if a.ignoreInput {
			return
		}
		if handler != nil {
			handler(checked)
		}
	})
	return action
}

func (a *actions) makeToggleAction(id, text, tooltip, defaultShortcut string, handler func()) *qtlib.QAction {
	a.allInfos = append(a.allInfos, ActionInfo{ID: id, Name: stripAccelerator(text), Tooltip: tooltip, DefaultShortcut: defaultShortcut})
	action := a.makeAction(text, a.shortcutFor(id, defaultShortcut), tooltip)
	action.SetCheckable(true)
	action.OnToggled(func(checked bool) {
		if a.ignoreInput || !checked {
			return
		}
		if handler != nil {
			handler()
		}
	})
	return action
}

func (a *actions) makeAction(text, shortcut, tooltip string) *qtlib.QAction {
	action := qtlib.NewQAction()
	action.SetText(text)
	if shortcut != "" {
		action.SetShortcut(qtlib.NewQKeySequence2(shortcut))
	}
	if tooltip != "" {
		action.SetToolTip(tooltip)
	}
	return action
}

func stripAccelerator(text string) string {
	return strings.TrimSuffix(strings.ReplaceAll(text, "&", ""), "...")
}

func (a *actions) shortcutFor(id, defaultShortcut string) string {
	if s, ok := a.keybindings[id]; ok && s != "" {
		return s
	}
	return defaultShortcut
}

func (a *actions) updateFromController() {
	if a.controller == nil {
		return
	}
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
