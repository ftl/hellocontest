package ui

import (
	"fmt"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/app"
)

type actions struct {
	parent *qtlib.QWidget

	controller  *app.Controller
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
	sendMacro1Action *qtlib.QAction
	sendMacro2Action *qtlib.QAction
	sendMacro3Action *qtlib.QAction
	sendMacro4Action *qtlib.QAction
}

func newActions(parent *qtlib.QWidget, controller *app.Controller) *actions {
	a := &actions{
		parent:            parent,
		controller:        controller,
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
	a.newFileAction = a.makeTriggerAction("&New...", "", controller.New)
	a.openFileAction = a.makeTriggerAction("&Open...", "", controller.Open)
	a.saveAsAction = a.makeTriggerAction("Save &As...", "", controller.SaveAs)
	a.exportSummaryAction = a.makeTriggerAction("&Summary...", "", controller.ExportSummary)
	a.exportCabrilloAction = a.makeTriggerAction("&Cabrillo...", "", controller.ExportCabrillo)
	a.exportADIFAction = a.makeTriggerAction("&ADIF...", "", controller.ExportADIF)
	a.exportCSVAction = a.makeTriggerAction("C&SV...", "", controller.ExportCSV)
	a.exportCallhistoryAction = a.makeTriggerAction("Call &History...", "", controller.ExportCallhistory)
	a.openRulesAction = a.makeTriggerAction("Open Contest &Rules Page", "", controller.OpenContestRulesPage)
	a.openUploadAction = a.makeTriggerAction("Open Contest &Upload Page", "", controller.OpenContestUploadPage)
	a.settingsAction = a.makeTriggerAction("Contest &Settings...", "Ctrl+.", controller.OpenSettings)
	a.configFileAction = a.makeTriggerAction("&Configuration File...", "", controller.OpenConfigurationFile)
	a.quitAction = a.makeTriggerAction("&Quit", "Ctrl+Q", controller.Quit)

	// Edit menu
	a.clearEntryAction = a.makeTriggerAction("&Clear Entry Fields", "", controller.ClearEntryFields)
	a.gotoEntryAction = a.makeTriggerAction("&Goto Entry Fields", "Ctrl+E", controller.GotoEntryFields)
	a.editLastAction = a.makeTriggerAction("Edit &Last QSO", "Ctrl+L", controller.EditLastQSO)
	a.refreshPredAction = a.makeTriggerAction("Refresh &Prediction", "Ctrl+P", controller.RefreshPrediction)
	a.logQSOAction = a.makeTriggerAction("Log QSO", "", controller.LogQSO)
	a.startParrotAction = a.makeTriggerAction("Start Parrot", "Ctrl+F1", controller.StartParrot)
	a.esmAction = a.makeCheckAction("ESM", "Ctrl+Shift+M", controller.SetESMEnabled)
	a.spAction = a.makeToggleAction("S&P", "Ctrl+S", controller.SwitchToSPWorkmode)
	a.workModeGroup.AddAction(a.spAction)
	a.runAction = a.makeToggleAction("Run", "Ctrl+R", controller.SwitchToRunWorkmode)
	a.workModeGroup.AddAction(a.runAction)
	a.offerQTCAction = a.makeTriggerAction("Offer QTC", "F5", controller.OfferQTC)
	a.requestQTCAction = a.makeTriggerAction("Request QTC", "F6", controller.RequestQTC)

	// Radio menu
	a.xitActiveAction = a.makeCheckAction("XIT Active", "Ctrl+Shift+X", controller.SetXITActive)

	// Bandmap menu
	a.markBandmapAction = a.makeTriggerAction("Mark In Bandmap", "Ctrl+M", controller.MarkInBandmap)
	a.highestSpotAction = a.makeTriggerAction("Goto Highest Value Spot", "Ctrl+Shift+N", controller.GotoHighestValueSpot)
	a.nearestSpotAction = a.makeTriggerAction("Goto Nearest Spot", "Ctrl+N", controller.GotoNearestSpot)
	a.nextSpotUpAction = a.makeTriggerAction("Goto Next Spot Up", "Ctrl+Up", controller.GotoNextSpotUp)
	a.nextSpotDownAction = a.makeTriggerAction("Goto Next Spot Down", "Ctrl+Down", controller.GotoNextSpotDown)
	a.sendSpotsToTciAction = a.makeCheckAction("Send Spots to TCI", "", controller.SetSendSpotsToTci)

	// Window menu
	a.showQSOsAction = a.makeTriggerAction("&QSOs", "", controller.ShowQSOs)
	a.showQTCsAction = a.makeTriggerAction("Q&TCs", "", controller.ShowQTCs)
	a.showScoreGraphAction = a.makeTriggerAction("Score &Graph", "", controller.ShowScoreGraph)
	a.showScoreTableAction = a.makeTriggerAction("&Score Table", "", controller.ShowScoreTable)
	a.showRateAction = a.makeTriggerAction("&Rate", "", controller.ShowRate)
	a.showSpotsAction = a.makeTriggerAction("S&pots", "", controller.ShowSpots)

	// Help menu
	a.wikiAction = a.makeTriggerAction("&Wiki", "", controller.OpenWiki)
	a.sponsorsAction = a.makeTriggerAction("&Sponsors", "", controller.Sponsors)
	a.aboutAction = a.makeTriggerAction("&About", "", controller.About)

	// Other actions
	a.sendMacro1Action = a.makeTriggerAction("Macro 1", "F1", func() { controller.Keyer.Send(0) })
	a.sendMacro2Action = a.makeTriggerAction("Macro 2", "F2", func() { controller.Keyer.Send(1) })
	a.sendMacro3Action = a.makeTriggerAction("Macro 3", "F3", func() { controller.Keyer.Send(2) })
	a.sendMacro4Action = a.makeTriggerAction("Macro 4", "F4", func() { controller.Keyer.Send(3) })
	a.parent.AddActions([]*qtlib.QAction{a.sendMacro1Action, a.sendMacro2Action, a.sendMacro3Action, a.sendMacro4Action})

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
	if existing, ok := a.radioActions[name]; ok {
		return existing
	}
	action := qtlib.NewQAction()
	action.SetText(name)
	action.SetCheckable(true)
	a.radioGroup.AddAction(action)
	a.radioActions[name] = action
	action.OnToggled(func(checked bool) {
		if a.ignoreInput || !checked {
			return
		}
		a.controller.Radio.SelectRadio(name)
	})
	return action
}

func (a *actions) AddKeyerAction(name string) *qtlib.QAction {
	if existing, ok := a.keyerActions[name]; ok {
		return existing
	}
	action := qtlib.NewQAction()
	action.SetText(name)
	action.SetCheckable(true)
	a.keyerGroup.AddAction(action)
	a.keyerActions[name] = action
	action.OnToggled(func(checked bool) {
		if a.ignoreInput || !checked {
			return
		}
		a.controller.Radio.SelectKeyer(name)
	})
	return action
}

func (a *actions) AddSpotSourceAction(name string) *qtlib.QAction {
	if existing, ok := a.spotSourceActions[name]; ok {
		return existing
	}
	action := qtlib.NewQAction()
	action.SetText(fmt.Sprintf("Use %s", name))
	action.SetCheckable(true)
	a.spotSourceActions[name] = action
	action.OnToggled(func(checked bool) {
		if a.ignoreInput {
			return
		}
		a.controller.Clusters.SetSpotSourceEnabled(name, checked)
	})
	return action
}

func (a *actions) SetRadioSelected(name string) {
	a.ignoreInput = true
	defer func() { a.ignoreInput = false }()
	for n, action := range a.radioActions {
		action.SetChecked(n == name)
	}
}

func (a *actions) SetKeyerSelected(name string) {
	a.ignoreInput = true
	defer func() { a.ignoreInput = false }()
	for n, action := range a.keyerActions {
		action.SetChecked(n == name)
	}
}

func (a *actions) SetSpotSourceEnabled(name string, enabled bool) {
	action, ok := a.spotSourceActions[name]
	if !ok {
		return
	}
	a.ignoreInput = true
	defer func() { a.ignoreInput = false }()
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
