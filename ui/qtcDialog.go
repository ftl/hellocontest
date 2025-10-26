package ui

import (
	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
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
	return 0, false
}

func (d *qtcDialog) Show(qtcMode core.QTCMode, qtcSeries core.QTCSeries) bool {
	// TODO: provide the qtcMode to generate the corresponding UI details
	d.view = newQTCView(d.controller)

	dialog, _ := gtk.DialogNew()
	d.dialog = dialog
	d.dialog.SetDefaultSize(400, 400)
	d.dialog.SetTransientFor(nil)
	d.dialog.SetPosition(gtk.WIN_POS_CENTER)
	d.dialog.Connect("destroy", d.onDestroy)
	d.dialog.SetTitle("QTC")
	d.dialog.SetDefaultResponse(gtk.RESPONSE_OK)
	d.dialog.SetModal(true)
	// contentArea, _ := d.dialog.GetContentArea()
	// contentArea.Add(d.view.root)
	d.dialog.AddButton("Log", gtk.RESPONSE_OK)
	d.dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)

	// TODO: put the data from qtcSeries into the view

	d.dialog.ShowAll()
	result := d.dialog.Run() == gtk.RESPONSE_OK
	d.dialog.Close()
	d.dialog.Destroy()
	d.dialog = nil
	d.view = nil

	return result
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
