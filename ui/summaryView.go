package ui

import "github.com/gotk3/gotk3/gtk"

type SummaryController interface {
	OperatorModes() []string
	Overlays() []string
	PowerModes() []string

	SetOperatorMode(string)
	SetOverlay(string)
	SetPowerMode(string)
	SetAssisted(bool)

	SetOpenAfterExport(bool)
}

type summaryView struct {
	controller SummaryController

	root *gtk.Grid

	contestNameEntry  *gtk.Entry
	cabrilloNameEntry *gtk.Entry
	startTimeEntry    *gtk.Entry
	callsignEntry     *gtk.Entry
	myExchangesEntry  *gtk.Entry

	operatorModeCombo   *gtk.ComboBoxText
	overlayCombo        *gtk.ComboBoxText
	powerModeCombo      *gtk.ComboBoxText
	assistedCheckButton *gtk.CheckButton

	workedModesEntry   *gtk.Entry
	workedBandsEntry   *gtk.Entry
	operatingTimeEntry *gtk.Entry
	breakTimeEntry     *gtk.Entry
	breaksEntry        *gtk.Entry

	scoreTable *scoreTable

	openAfterExportCheckButton *gtk.CheckButton
}

func newSummaryView(controller SummaryController) *summaryView {
	result := &summaryView{
		controller: controller,
	}

	result.root, _ = gtk.GridNew()
	result.root.SetOrientation(gtk.ORIENTATION_VERTICAL)
	result.root.SetHExpand(true)
	result.root.SetVExpand(true)
	result.root.SetColumnSpacing(5)
	result.root.SetRowSpacing(5)
	result.root.SetMarginStart(5)
	result.root.SetMarginEnd(5)

	result.contestNameEntry = buildLabeledEntry(result.root, 0, "Contest Name", nil)
	result.cabrilloNameEntry = buildLabeledEntry(result.root, 1, "Cabrillo Name", nil)
	result.startTimeEntry = buildLabeledEntry(result.root, 2, "Start Time", nil)
	result.callsignEntry = buildLabeledEntry(result.root, 3, "Callsign", nil)
	result.myExchangesEntry = buildLabeledEntry(result.root, 4, "My Exchanges", nil)

	buildSeparator(result.root, 5, 2)

	result.operatorModeCombo = buildLabeledCombo(result.root, 6, "Operator Mode", false, result.controller.OperatorModes(), result.onOperatorModeChanged)
	result.overlayCombo = buildLabeledCombo(result.root, 7, "Overlay", false, result.controller.Overlays(), result.onOverlayChanged)
	result.powerModeCombo = buildLabeledCombo(result.root, 8, "Power", false, result.controller.PowerModes(), result.onPowerModeChanged)
	result.assistedCheckButton = buildCheckButtonInColumn(result.root, 9, 1, 1, "Assisted", result.onAssistedToggled)

	buildSeparator(result.root, 10, 2)

	result.workedModesEntry = buildLabeledEntry(result.root, 11, "Worked Modes", nil)
	result.workedBandsEntry = buildLabeledEntry(result.root, 12, "Worked Bands", nil)
	result.operatingTimeEntry = buildLabeledEntry(result.root, 13, "Operating Time", nil)
	result.breakTimeEntry = buildLabeledEntry(result.root, 14, "Break Time", nil)
	result.breaksEntry = buildLabeledEntry(result.root, 15, "Breaks", nil)

	buildSeparator(result.root, 16, 2)

	result.scoreTable = newScoreTable(nil)
	result.root.Attach(result.scoreTable.Table(), 0, 17, 2, 1)

	buildSeparator(result.root, 18, 2)

	result.openAfterExportCheckButton = buildCheckButtonInColumn(result.root, 19, 0, 2, "Open the file after export", result.onOpenAfterExportToggled)

	return result
}

func (v *summaryView) onOperatorModeChanged() {
	v.controller.SetOperatorMode(v.operatorModeCombo.GetActiveText())
}

func (v *summaryView) onOverlayChanged() {
	v.controller.SetOverlay(v.overlayCombo.GetActiveText())
}

func (v *summaryView) onPowerModeChanged() {
	v.controller.SetPowerMode(v.powerModeCombo.GetActiveText())
}

func (v *summaryView) onAssistedToggled() {
	v.controller.SetAssisted(v.assistedCheckButton.GetActive())
}

func (v *summaryView) onOpenAfterExportToggled() {
	v.controller.SetOpenAfterExport(v.openAfterExportCheckButton.GetActive())
}
