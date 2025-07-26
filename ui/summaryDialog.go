package ui

import "github.com/gotk3/gotk3/gtk"

type summaryDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller SummaryController
	view       *summaryView

	openAfterExport bool
}

func setupSummaryDialog(parent gtk.IWidget, controller SummaryController) *summaryDialog {
	result := &summaryDialog{
		parent:     parent,
		controller: controller,
	}
	return result
}

func (d *summaryDialog) onDestroy() {
	d.dialog = nil
	d.view = nil
}

func (d *summaryDialog) Show() bool {
	d.view = newSummaryView(d.controller)

	d.view.openAfterExportCheckButton.SetActive(d.openAfterExport)

	dialog, _ := gtk.DialogNew()
	d.dialog = dialog
	d.dialog.SetDefaultSize(400, 400)
	d.dialog.SetTransientFor(nil)
	d.dialog.SetPosition(gtk.WIN_POS_CENTER)
	d.dialog.Connect("destroy", d.onDestroy)
	d.dialog.SetTitle("Summary")
	d.dialog.SetDefaultResponse(gtk.RESPONSE_OK)
	d.dialog.SetModal(true)
	contentArea, _ := d.dialog.GetContentArea()
	contentArea.Add(d.view.root)
	d.dialog.AddButton("Export", gtk.RESPONSE_OK)
	d.dialog.AddButton("Close", gtk.RESPONSE_CANCEL)

	d.dialog.ShowAll()
	result := d.dialog.Run() == gtk.RESPONSE_OK
	d.dialog.Close()
	d.dialog.Destroy()
	d.dialog = nil
	d.view = nil

	return result
}

func (d *summaryDialog) SetOpenAfterExport(open bool) {
	d.openAfterExport = open
}
