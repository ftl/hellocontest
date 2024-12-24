package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type exportCabrilloDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller ExportCabrilloController
	view       *exportCabrilloView

	openAfterExport bool
}

func setupExportCabrilloDialog(parent gtk.IWidget, controller ExportCabrilloController) *exportCabrilloDialog {
	result := &exportCabrilloDialog{
		parent:     parent,
		controller: controller,
	}
	return result
}

func (d *exportCabrilloDialog) onDestroy() {
	d.dialog = nil
	d.view = nil
}

func (d *exportCabrilloDialog) Show() bool {
	d.view = &exportCabrilloView{}
	grid, _ := gtk.GridNew()
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	label, _ := gtk.LabelNew("Export the log as Cabrillo format.")
	grid.Attach(label, 0, 0, 2, 1)

	d.view.openAfterExportCheckButton, _ = gtk.CheckButtonNewWithLabel("Open the file after export")
	d.view.openAfterExportCheckButton.SetActive(d.openAfterExport)
	grid.Attach(d.view.openAfterExportCheckButton, 0, 1, 2, 1)

	d.view.setup(d.controller)

	dialog, _ := gtk.DialogNew()
	d.dialog = dialog
	d.dialog.SetDefaultSize(400, 300)
	d.dialog.SetTransientFor(nil)
	d.dialog.SetPosition(gtk.WIN_POS_CENTER)
	d.dialog.Connect("destroy", d.onDestroy)
	d.dialog.SetTitle("Export Log as Cabrillo")
	d.dialog.SetDefaultResponse(gtk.RESPONSE_OK)
	d.dialog.SetModal(true)
	contentArea, _ := d.dialog.GetContentArea()
	contentArea.Add(grid)
	d.dialog.AddButton("Export", gtk.RESPONSE_OK)
	d.dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)

	d.dialog.ShowAll()
	result := d.dialog.Run() == gtk.RESPONSE_OK
	d.dialog.Close()
	d.dialog.Destroy()
	d.dialog = nil
	d.view = nil

	return result
}

func (d *exportCabrilloDialog) SetOpenAfterExport(open bool) {
	d.openAfterExport = open
}
