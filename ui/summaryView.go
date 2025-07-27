package ui

import "github.com/gotk3/gotk3/gtk"

type SummaryController interface {
	SetOpenAfterExport(bool)
}

type summaryView struct {
	controller SummaryController

	root *gtk.Grid

	contestNameEntry  *gtk.Entry
	cabrilloNameEntry *gtk.Entry
	callsignEntry     *gtk.Entry
	myExchangesEntry  *gtk.Entry

	workedModesEntry   *gtk.Entry
	workedBandsEntry   *gtk.Entry
	operatingTimeEntry *gtk.Entry
	breakTimeEntry     *gtk.Entry
	breaksEntry        *gtk.Entry

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
	result.callsignEntry = buildLabeledEntry(result.root, 2, "Callsign", nil)
	result.myExchangesEntry = buildLabeledEntry(result.root, 3, "My Exchanges", nil)

	buildSeparator(result.root, 4, 2)

	result.workedModesEntry = buildLabeledEntry(result.root, 5, "Worked Modes", nil)
	result.workedBandsEntry = buildLabeledEntry(result.root, 6, "Worked Bands", nil)
	result.operatingTimeEntry = buildLabeledEntry(result.root, 7, "Operating Time", nil)
	result.breakTimeEntry = buildLabeledEntry(result.root, 8, "Break Time", nil)
	result.breaksEntry = buildLabeledEntry(result.root, 9, "Breaks", nil)

	buildSeparator(result.root, 10, 2)
	result.openAfterExportCheckButton = buildCheckButton(result.root, 11, "Open the file after export", result.onOpenAfterExportToggled)

	return result
}

func (v *summaryView) onOpenAfterExportToggled() {
	v.controller.SetOpenAfterExport(v.openAfterExportCheckButton.GetActive())
}
