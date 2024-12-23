package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type exportCabrilloDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget
}

func setupExportCabrilloDialog(parent gtk.IWidget, controller ExportCabrilloController) *exportCabrilloDialog {
	result := &exportCabrilloDialog{
		parent: parent,
	}
	return result
}

func (d *exportCabrilloDialog) onDestroy() {
	d.dialog = nil
}

func (d *exportCabrilloDialog) Show() bool {
	label, _ := gtk.LabelNew("Export the log as Cabrillo format.")

	grid, _ := gtk.GridNew()
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	grid.Add(label)

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

	return result
}
