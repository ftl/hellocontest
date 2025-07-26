package ui

import "github.com/gotk3/gotk3/gtk"

type SummaryController interface {
	SetOpenAfterExport(bool)
}

type summaryView struct {
	controller SummaryController

	root *gtk.Grid

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

	result.openAfterExportCheckButton = buildCheckButton(result.root, 1, "Open the file after export", result.onOpenAfterExportToggled)

	return result
}

func (v *summaryView) onOpenAfterExportToggled() {
	v.controller.SetOpenAfterExport(v.openAfterExportCheckButton.GetActive())
}
