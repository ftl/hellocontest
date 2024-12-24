package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type exportCabrilloDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller ExportCabrilloController
	view       *exportCabrilloView

	categoryAssisted bool
	categoryBand     string
	categoryMode     string
	categoryOperator string
	categoryPower    string
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
	d.view = &exportCabrilloView{}
	grid, _ := gtk.GridNew()
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	grid.SetHExpand(true)
	grid.SetVExpand(true)
	grid.SetColumnSpacing(5)

	label, _ := gtk.LabelNew("Export the log as Cabrillo file.")
	label.SetHAlign(gtk.ALIGN_START)
	label.SetMarginBottom(10)
	grid.Attach(label, 0, 0, 2, 1)

	label, _ = gtk.LabelNew("Category")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 1, 1, 1)
	d.view.categoriesCombo, _ = gtk.ComboBoxTextNew()
	d.view.categoriesCombo.Append("", "")
	for _, category := range d.controller.Categories() {
		d.view.categoriesCombo.Append(category, category)
	}
	d.view.categoriesCombo.SetActive(0)
	grid.Attach(d.view.categoriesCombo, 1, 1, 1, 1)

	d.view.categoryAssistedCheckButton, _ = gtk.CheckButtonNewWithLabel("Assisted")
	d.view.categoryAssistedCheckButton.SetActive(d.categoryAssisted)
	grid.Attach(d.view.categoryAssistedCheckButton, 1, 2, 1, 1)

	label, _ = gtk.LabelNew("Band")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 3, 1, 1)
	d.view.categoryBandCombo, _ = gtk.ComboBoxTextNew()
	d.view.categoryBandCombo.Append("", "")
	for _, band := range d.controller.CategoryBands() {
		d.view.categoryBandCombo.Append(band, band)
	}
	d.view.categoryBandCombo.SetActiveID(d.categoryBand)
	grid.Attach(d.view.categoryBandCombo, 1, 3, 1, 1)

	label, _ = gtk.LabelNew("Mode")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 4, 1, 1)
	d.view.categoryModeCombo, _ = gtk.ComboBoxTextNew()
	d.view.categoryModeCombo.Append("", "")
	for _, mode := range d.controller.CategoryModes() {
		d.view.categoryModeCombo.Append(mode, mode)
	}
	d.view.categoryModeCombo.SetActiveID(d.categoryMode)
	grid.Attach(d.view.categoryModeCombo, 1, 4, 1, 1)

	label, _ = gtk.LabelNew("Operator")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 5, 1, 1)
	d.view.categoryOperatorCombo, _ = gtk.ComboBoxTextNew()
	d.view.categoryOperatorCombo.Append("", "")
	for _, operator := range d.controller.CategoryOperators() {
		d.view.categoryOperatorCombo.Append(operator, operator)
	}
	d.view.categoryOperatorCombo.SetActiveID(d.categoryOperator)
	grid.Attach(d.view.categoryOperatorCombo, 1, 5, 1, 1)

	label, _ = gtk.LabelNew("Power")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 6, 1, 1)
	d.view.categoryPowerCombo, _ = gtk.ComboBoxTextNew()
	d.view.categoryPowerCombo.Append("", "")
	for _, power := range d.controller.CategoryPowers() {
		d.view.categoryPowerCombo.Append(power, power)
	}
	d.view.categoryPowerCombo.SetActiveID(d.categoryPower)
	grid.Attach(d.view.categoryPowerCombo, 1, 6, 1, 1)

	separator, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	separator.SetHExpand(true)
	separator.SetVExpand(true)
	grid.Attach(separator, 0, 7, 2, 1)

	d.view.openAfterExportCheckButton, _ = gtk.CheckButtonNewWithLabel("Open the file after export")
	d.view.openAfterExportCheckButton.SetActive(d.openAfterExport)
	grid.Attach(d.view.openAfterExportCheckButton, 0, 8, 2, 1)

	d.view.setup(d.controller)

	dialog, _ := gtk.DialogNew()
	d.dialog = dialog
	d.dialog.SetDefaultSize(400, 300)
	d.dialog.SetTransientFor(nil)
	d.dialog.SetPosition(gtk.WIN_POS_CENTER)
	d.dialog.Connect("destroy", d.onDestroy)
	d.dialog.SetTitle("Export Cabrillo")
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

func (d *exportCabrilloDialog) SetCategoryAssisted(assisted bool) {
	d.categoryAssisted = assisted
	if d.view != nil {
		d.view.categoryAssistedCheckButton.SetActive(assisted)
	}
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

func (d *exportCabrilloDialog) SetOpenAfterExport(open bool) {
	d.openAfterExport = open
}
