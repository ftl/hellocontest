package ui

import (
	"fmt"
	"strings"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/app"
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
	toggleVFOAction       *qtlib.QAction
	focusVFO1Action       *qtlib.QAction
	focusVFO2Action       *qtlib.QAction
	muteAudioVFO1Action   *qtlib.QAction
	muteAudioVFO2Action   *qtlib.QAction
	unmuteAudioVFO1Action *qtlib.QAction
	unmuteAudioVFO2Action *qtlib.QAction
	toggleAudioVFO1Action *qtlib.QAction
	toggleAudioVFO2Action *qtlib.QAction
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
	a.newFileAction = a.makeTriggerAction(core.ActionFileNew, "&New...", "Create a new contest log", "Ctrl+N", controller.New)
	a.openFileAction = a.makeTriggerAction(core.ActionFileOpen, "&Open...", "Open a contest log", "Ctrl+O", controller.Open)
	a.saveAsAction = a.makeTriggerAction(core.ActionFileSaveAs, "Save &As...", "", "", controller.SaveAs)
	a.exportSummaryAction = a.makeTriggerAction(core.ActionFileExportSummary, "&Summary...", "Export the contest summary", "", controller.ExportSummary)
	a.exportCabrilloAction = a.makeTriggerAction(core.ActionFileExportCabrillo, "&Cabrillo...", "Export the log as Cabrillo file", "", controller.ExportCabrillo)
	a.exportADIFAction = a.makeTriggerAction(core.ActionFileExportADIF, "&ADIF...", "Export the log as ADIF file", "", controller.ExportADIF)
	a.exportCSVAction = a.makeTriggerAction(core.ActionFileExportCSV, "C&SV...", "Export the log as CSV file", "", controller.ExportCSV)
	a.exportCallhistoryAction = a.makeTriggerAction(core.ActionFileExportCallhistory, "Call &History...", "Export a call history file based on the contest log", "", controller.ExportCallhistory)
	a.openRulesAction = a.makeTriggerAction(core.ActionFileOpenRules, "Open Contest &Rules Page", "", "", controller.OpenContestRulesPage)
	a.openUploadAction = a.makeTriggerAction(core.ActionFileOpenUpload, "Open Contest &Upload Page", "", "", controller.OpenContestUploadPage)
	a.settingsAction = a.makeTriggerAction(core.ActionFileSettings, "Contest &Settings...", "", "Ctrl+.", controller.OpenSettings)
	a.configFileAction = a.makeTriggerAction(core.ActionFileConfigFile, "Open &Configuration File...", "Open the configuration file in the external editor", "Ctrl+Shift+.", controller.OpenConfigurationFile)
	a.quitAction = a.makeTriggerAction(core.ActionFileQuit, "&Quit", "", "Ctrl+Q", controller.Quit)

	// Edit menu
	a.clearEntryAction = a.makeTriggerAction(core.ActionEntryClear, "&Clear Entry Fields", "Clear the entry fields", "", controller.ClearEntryFields)
	a.gotoEntryAction = a.makeTriggerAction(core.ActionEntryGotoEntryField, "&Goto Entry Fields", "Focus the entry field", "Ctrl+E", controller.GotoEntryFields)
	a.editLastAction = a.makeTriggerAction(core.ActionEntryEditLastQSO, "Edit &Last QSO", "Select the last QSO for editing", "Ctrl+L", controller.EditLastQSO)
	a.refreshPredAction = a.makeTriggerAction(core.ActionEntryRefreshPrediction, "Refresh &Prediction", "Refresh the exchange prediction", "Ctrl+P", controller.RefreshPrediction)
	a.logQSOAction = a.makeTriggerAction(core.ActionEntryLogQSO, "Log QSO", "", "Ctrl+Return", controller.LogQSO)
	a.startParrotAction = a.makeTriggerAction(core.ActionEntryStartParrot, "Start Parrot", "Start the keyer parrot", "Ctrl+F1", controller.StartParrot)
	a.esmAction = a.makeCheckAction(core.ActionEntryEnableESM, "ESM", "Enable ESM", "Ctrl+Shift+M", controller.SetESMEnabled)
	a.spAction = a.makeToggleAction(core.ActionEntryWorkmodeSP, "S&P", "", "Ctrl+S", controller.SwitchToSPWorkmode)
	a.workModeGroup.AddAction(a.spAction)
	a.runAction = a.makeToggleAction(core.ActionEntryWorkmodeRun, "Run", "", "Ctrl+R", controller.SwitchToRunWorkmode)
	a.workModeGroup.AddAction(a.runAction)
	a.offerQTCAction = a.makeTriggerAction(core.ActionEntryOfferQTC, "Offer QTC", "", "F5", controller.OfferQTC)
	a.requestQTCAction = a.makeTriggerAction(core.ActionEntryRequestQTC, "Request QTC", "", "F6", controller.RequestQTC)

	// Radio menu
	a.xitActiveAction = a.makeCheckAction(core.ActionRadioXITActive, "XIT Active", "Activate the XIT for the search & pounce workmode", "Ctrl+Shift+X", controller.SetXITActive)

	// Bandmap menu
	a.markBandmapAction = a.makeTriggerAction(core.ActionBandmapMark, "Mark In Bandmap", "", "Ctrl+M", controller.MarkInBandmap)
	a.highestSpotAction = a.makeTriggerAction(core.ActionBandmapGotoHighestValueSpot, "Goto Highest Value Spot", "Go to the spot with the highest predicted value", "Ctrl+Shift+N", controller.GotoHighestValueSpot)
	a.nearestSpotAction = a.makeTriggerAction(core.ActionBandmapGotoNearestSpot, "Goto Nearest Spot", "Go to the spot closest to the current frequency", "Ctrl+N", controller.GotoNearestSpot)
	a.nextSpotUpAction = a.makeTriggerAction(core.ActionBandmapGotoNextSpotUp, "Goto Next Spot Up", "", "Ctrl+Up", controller.GotoNextSpotUp)
	a.nextSpotDownAction = a.makeTriggerAction(core.ActionBandmapGotoNextSpotDown, "Goto Next Spot Down", "", "Ctrl+Down", controller.GotoNextSpotDown)
	a.sendSpotsToTciAction = a.makeCheckAction(core.ActionBandmapSendSpotsToTci, "Send Spots to TCI", "", "", controller.SetSendSpotsToTci)

	// Window menu
	a.showQSOsAction = a.makeTriggerAction(core.ActionWindowShowQSOs, "&QSOs", "Show the QSOs list", "", controller.ShowQSOs)
	a.showQTCsAction = a.makeTriggerAction(core.ActionWindowShowQTCs, "Q&TCs", "Show the QTCs list", "", controller.ShowQTCs)
	a.showScoreGraphAction = a.makeTriggerAction(core.ActionWindowShowScoreGraph, "Score &Graph", "Show the score graph", "", controller.ShowScoreGraph)
	a.showScoreTableAction = a.makeTriggerAction(core.ActionWindowShowScoreTable, "&Score Table", "Show the score table", "", controller.ShowScoreTable)
	a.showRateAction = a.makeTriggerAction(core.ActionWindowShowRate, "&Rate", "Show the QSO rate", "", controller.ShowRate)
	a.showSpotsAction = a.makeTriggerAction(core.ActionWindowShowSpots, "S&pots", "Show the spots table", "", controller.ShowSpots)

	// Help menu
	a.wikiAction = a.makeTriggerAction(core.ActionHelpWiki, "&Wiki", "Open the manual in the webbrowser", "", controller.OpenWiki)
	a.sponsorsAction = a.makeTriggerAction(core.ActionHelpSponsors, "&Sponsors", "Open the sponsors page in the webbrowser", "", controller.Sponsors)
	a.aboutAction = a.makeTriggerAction(core.ActionHelpAbout, "&About", "Show the about dialog", "", controller.About)

	// Other actions
	a.sendMacro1Action = a.makeTriggerAction(core.ActionKeyerSendMacro1, "Send Macro 1", "Send the keyer macro #1", "F1", func() { controller.Keyer.Send(0) })
	a.sendMacro2Action = a.makeTriggerAction(core.ActionKeyerSendMacro2, "Send Macro 2", "Send the keyer macro #2", "F2", func() { controller.Keyer.Send(1) })
	a.sendMacro3Action = a.makeTriggerAction(core.ActionKeyerSendMacro3, "Send Macro 3", "Send the keyer macro #3", "F3", func() { controller.Keyer.Send(2) })
	a.sendMacro4Action = a.makeTriggerAction(core.ActionKeyerSendMacro4, "Send Macro 4", "Send the keyer macro #4", "F4", func() { controller.Keyer.Send(3) })
	a.selectBestMatchAction = a.makeTriggerAction(core.ActionEntrySelectBestMatch, "Select Best Match", "Select the best matching callsign", "Alt+Return", func() { controller.Entry.SelectBestMatchOnFrequency() })
	a.nextESMStepAction = a.makeTriggerAction(core.ActionEntryNextESMStep, "Next ESM Step", "Execute the next ESM step", "", func() { controller.Entry.NextESMStep() })
	a.toggleVFOAction = a.makeTriggerAction(core.ActionEntryToggleFocusedVFO, "Toggle Focused VFO", "Switch focus between VFO 1 and VFO 2", "F8", func() { controller.Entry.ToggleFocusedVFO() })
	a.focusVFO1Action = a.makeTriggerAction(core.ActionEntryFocusVFO1, "Focus VFO 1", "Set the focused VFO to VFO 1", "F9", func() { controller.Entry.FocusVFO1() })
	a.focusVFO2Action = a.makeTriggerAction(core.ActionEntryFocusVFO2, "Focus VFO 2", "Set the focused VFO to VFO 2", "F10", func() { controller.Entry.FocusVFO2() })
	a.muteAudioVFO1Action = a.makeTriggerAction(core.ActionRadioMuteAudioVFO1, "Mute VFO 1", "Mute audio on VFO 1", "", func() { controller.MuteAudio(core.VFO1) })
	a.muteAudioVFO2Action = a.makeTriggerAction(core.ActionRadioMuteAudioVFO2, "Mute VFO 2", "Mute audio on VFO 2", "", func() { controller.MuteAudio(core.VFO2) })
	a.unmuteAudioVFO1Action = a.makeTriggerAction(core.ActionRadioUnmuteAudioVFO1, "Unmute VFO 1", "Unmute audio on VFO 1", "", func() { controller.UnmuteAudio(core.VFO1) })
	a.unmuteAudioVFO2Action = a.makeTriggerAction(core.ActionRadioUnmuteAudioVFO2, "Unmute VFO 2", "Unmute audio on VFO 2", "", func() { controller.UnmuteAudio(core.VFO2) })
	a.toggleAudioVFO1Action = a.makeTriggerAction(core.ActionRadioToggleAudioVFO1, "Toggle Audio VFO 1", "Toggle audio on VFO 1", "", func() { controller.ToggleAudio(core.VFO1) })
	a.toggleAudioVFO2Action = a.makeTriggerAction(core.ActionRadioToggleAudioVFO2, "Toggle Audio VFO 2", "Toggle audio on VFO 2", "", func() { controller.ToggleAudio(core.VFO2) })
	a.parent.AddActions([]*qtlib.QAction{
		a.sendMacro1Action,
		a.sendMacro2Action,
		a.sendMacro3Action,
		a.sendMacro4Action,
		a.selectBestMatchAction,
		a.nextESMStepAction,
		a.toggleVFOAction,
		a.focusVFO1Action,
		a.focusVFO2Action,
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
