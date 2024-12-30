package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type exportCabrilloDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller ExportCabrilloController
	view       *exportCabrilloView

	categoryBand     string
	categoryMode     string
	categoryOperator string
	categoryPower    string
	categoryAssisted string
	name             string
	email            string
	openAfterExport  bool
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
	d.view = newExportCabrilloView(d.controller)
	d.view.categoriesCombo.SetActiveID("")
	d.view.categoryBandCombo.SetActiveID(d.categoryBand)
	d.view.categoryModeCombo.SetActiveID(d.categoryMode)
	d.view.categoryOperatorCombo.SetActiveID(d.categoryOperator)
	d.view.categoryPowerCombo.SetActiveID(d.categoryPower)
	d.view.categoryAssistedCombo.SetActiveID(d.categoryAssisted)
	d.view.nameEntry.SetText(d.name)
	d.view.emailEntry.SetText(d.email)
	d.view.openAfterExportCheckButton.SetActive(d.openAfterExport)

	dialog, _ := gtk.DialogNew()
	d.dialog = dialog
	d.dialog.SetDefaultSize(400, 400)
	d.dialog.SetTransientFor(nil)
	d.dialog.SetPosition(gtk.WIN_POS_CENTER)
	d.dialog.Connect("destroy", d.onDestroy)
	d.dialog.SetTitle("Export Cabrillo")
	d.dialog.SetDefaultResponse(gtk.RESPONSE_OK)
	d.dialog.SetModal(true)
	contentArea, _ := d.dialog.GetContentArea()
	contentArea.Add(d.view.root)
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

func (d *exportCabrilloDialog) SetCategoryBand(band string) {
	d.categoryBand = band
	if d.view != nil {
		d.view.categoryBandCombo.SetActiveID(band)
	}
}

func (d *exportCabrilloDialog) SetCategoryMode(mode string) {
	d.categoryMode = mode
	if d.view != nil {
		d.view.categoryModeCombo.SetActiveID(mode)
	}
}

func (d *exportCabrilloDialog) SetCategoryOperator(operator string) {
	d.categoryOperator = operator
	if d.view != nil {
		d.view.categoryOperatorCombo.SetActiveID(operator)
	}
}

func (d *exportCabrilloDialog) SetCategoryPower(power string) {
	d.categoryPower = power
	if d.view != nil {
		d.view.categoryPowerCombo.SetActiveID(power)
	}
}

func (d *exportCabrilloDialog) SetCategoryAssisted(assisted string) {
	d.categoryAssisted = assisted
	if d.view != nil {
		d.view.categoryAssistedCombo.SetActiveID(assisted)
	}
}

func (d *exportCabrilloDialog) SetName(name string) {
	d.name = name
	if d.view != nil {
		d.view.nameEntry.SetText(name)
	}
}

func (d *exportCabrilloDialog) SetEmail(email string) {
	d.email = email
	if d.view != nil {
		d.view.emailEntry.SetText(email)
	}
}

func (d *exportCabrilloDialog) SetOpenAfterExport(open bool) {
	d.openAfterExport = open
}
