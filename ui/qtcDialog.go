package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type qtcDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller QTCController
	view       *qtcView

	// data fields
}

func setupQTCDialog(parent gtk.IWidget, controller QTCController) *qtcDialog {
	result := &qtcDialog{
		parent:     parent,
		controller: controller,
	}
	return result
}

func (d *qtcDialog) onDestroy() {
	d.dialog = nil
	d.view = nil
}

func (d *qtcDialog) QuestionQTCCount(max int) (int, bool) {
	// TODO: implement modal dialog
	return 10, true
}

func (d *qtcDialog) Show(qtcMode core.QTCMode, qtcSeries core.QTCSeries) {
	d.view = newQTCView(d.controller, qtcMode)

	// setup the dialog
	dialog, _ := gtk.DialogNew()
	d.dialog = dialog
	d.dialog.SetDefaultSize(400, 400)
	d.dialog.SetTransientFor(nil)
	d.dialog.SetPosition(gtk.WIN_POS_CENTER)
	d.dialog.Connect("destroy", d.onDestroy)
	d.dialog.SetTitle("QTC")
	d.dialog.SetDefaultResponse(gtk.RESPONSE_OK)
	d.dialog.SetModal(true)
	contentArea, _ := d.dialog.GetContentArea()
	contentArea.Add(d.view.root)
	d.dialog.AddButton("Log", gtk.RESPONSE_OK)
	// TODO: add a check before closing the dialog
	d.dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)
	d.dialog.ShowAll()

	// put the QTC series data into the view's widgets
	// IMPORTANT: This needs to happen after ShowAll, otherwise the
	// show/hide of the qtcRows does not work (done in setQTCs).
	d.view.setHeader(qtcSeries.TheirCallsign(), qtcSeries.Header)
	d.view.setQTCs(qtcSeries.QTCs)

	// run the dialog
	d.dialog.Run()
	d.dialog.Close()
	d.dialog.Destroy()
	d.dialog = nil
	d.view = nil
}

func (d *qtcDialog) Update(core.QTCSeries) {
	// TODO: implement
}

func (d *qtcDialog) Close() {
	// TODO: implement
}

func (d *qtcDialog) SetActiveField(core.QTCField) {
	// TODO: implement
}
