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

	columns, _ := gtk.GridNew()
	columns.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	columns.SetHExpand(true)
	columns.SetVExpand(false)
	columns.SetColumnSpacing(20)
	result.root.Attach(columns, 0, 1, 1, 1)

	leftColumn, _ := gtk.GridNew()
	leftColumn.SetOrientation(gtk.ORIENTATION_VERTICAL)
	leftColumn.SetHExpand(true)
	leftColumn.SetVExpand(false)
	leftColumn.SetColumnSpacing(5)
	leftColumn.SetRowSpacing(5)
	columns.Attach(leftColumn, 0, 0, 1, 1)

	rightColumn, _ := gtk.GridNew()
	rightColumn.SetOrientation(gtk.ORIENTATION_VERTICAL)
	rightColumn.SetHExpand(true)
	rightColumn.SetVExpand(false)
	rightColumn.SetColumnSpacing(5)
	rightColumn.SetRowSpacing(5)
	columns.Attach(rightColumn, 1, 0, 1, 1)

	// left

	result.contestNameEntry = buildLabeledEntry(leftColumn, 0, "Contest Name", nil)
	result.cabrilloNameEntry = buildLabeledEntry(leftColumn, 1, "Cabrillo Name", nil)
	result.startTimeEntry = buildLabeledEntry(leftColumn, 2, "Start Time", nil)
	result.callsignEntry = buildLabeledEntry(leftColumn, 3, "Callsign", nil)
	result.myExchangesEntry = buildLabeledEntry(leftColumn, 4, "My Exchanges", nil)

	buildHeaderLabel(leftColumn, 5, "Working Condition")
	result.operatorModeCombo = buildLabeledCombo(leftColumn, 6, "Operator Mode", false, result.controller.OperatorModes(), result.onOperatorModeChanged)
	result.overlayCombo = buildLabeledCombo(leftColumn, 7, "Overlay", false, result.controller.Overlays(), result.onOverlayChanged)
	result.powerModeCombo = buildLabeledCombo(leftColumn, 8, "Power", false, result.controller.PowerModes(), result.onPowerModeChanged)
	result.assistedCheckButton = buildCheckButtonInColumn(leftColumn, 9, 1, 1, "Assisted", result.onAssistedToggled)

	//right

	result.workedModesEntry = buildLabeledEntry(rightColumn, 0, "Worked Modes", nil)
	result.workedBandsEntry = buildLabeledEntry(rightColumn, 1, "Worked Bands", nil)
	result.operatingTimeEntry = buildLabeledEntry(rightColumn, 2, "Operating Time", nil)
	result.breakTimeEntry = buildLabeledEntry(rightColumn, 3, "Break Time", nil)
	result.breaksEntry = buildLabeledEntry(rightColumn, 4, "Breaks", nil)

	buildHeaderLabel(rightColumn, 5, "Claimed Score")
	result.scoreTable = newScoreTable(nil)
	rightColumn.Attach(result.scoreTable.Table(), 0, 6, 2, 1)

	buildSeparator(result.root, 2, 1)

	result.openAfterExportCheckButton = buildCheckButton(result.root, 3, "Open the file after export", result.onOpenAfterExportToggled)

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
