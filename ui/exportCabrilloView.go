package ui

import "github.com/gotk3/gotk3/gtk"

type ExportCabrilloController interface {
	Categories() []string
	CategoryBands() []string
	CategoryModes() []string
	CategoryOperators() []string
	CategoryPowers() []string

	SetCategory(string)
	SetCategoryAssisted(bool)
	SetCategoryBand(string)
	SetCategoryMode(string)
	SetCategoryOperator(string)
	SetCategoryPower(string)
	SetName(string)
	SetEmail(string)
	SetOpenAfterExport(bool)
}

type exportCabrilloView struct {
	controller ExportCabrilloController

	categoriesCombo             *gtk.ComboBoxText
	categoryAssistedCheckButton *gtk.CheckButton
	categoryBandCombo           *gtk.ComboBoxText
	categoryModeCombo           *gtk.ComboBoxText
	categoryOperatorCombo       *gtk.ComboBoxText
	categoryPowerCombo          *gtk.ComboBoxText
	nameEntry                   *gtk.Entry
	emailEntry                  *gtk.Entry
	openAfterExportCheckButton  *gtk.CheckButton
}

func (v *exportCabrilloView) setup(controller ExportCabrilloController) {
	v.controller = controller
	v.categoriesCombo.Connect("changed", v.onCategoryChanged)
	v.categoryAssistedCheckButton.Connect("toggled", v.onCategoryAssistedToggled)
	v.categoryBandCombo.Connect("changed", v.onCategoryBandChanged)
	v.categoryModeCombo.Connect("changed", v.onCategoryModeChanged)
	v.categoryOperatorCombo.Connect("changed", v.onCategoryOperatorChanged)
	v.categoryPowerCombo.Connect("changed", v.onCategoryPowerChanged)
	v.nameEntry.Connect("changed", v.onNameChanged)
	v.emailEntry.Connect("changed", v.onEmailChanged)
	v.openAfterExportCheckButton.Connect("toggled", v.onOpenAfterExportToggled)
}

func (v *exportCabrilloView) onCategoryChanged() {
	v.controller.SetCategory(v.categoriesCombo.GetActiveText())
}

func (v *exportCabrilloView) onCategoryAssistedToggled() {
	v.controller.SetCategoryAssisted(v.categoryAssistedCheckButton.GetActive())
}

func (v *exportCabrilloView) onCategoryBandChanged() {
	v.controller.SetCategoryBand(v.categoryBandCombo.GetActiveText())
}

func (v *exportCabrilloView) onCategoryModeChanged() {
	v.controller.SetCategoryMode(v.categoryModeCombo.GetActiveText())
}

func (v *exportCabrilloView) onCategoryOperatorChanged() {
	v.controller.SetCategoryOperator(v.categoryOperatorCombo.GetActiveText())
}

func (v *exportCabrilloView) onCategoryPowerChanged() {
	v.controller.SetCategoryPower(v.categoryPowerCombo.GetActiveText())
}

func (v *exportCabrilloView) onNameChanged() {
	text, _ := v.nameEntry.GetText()
	v.controller.SetName(text)
}

func (v *exportCabrilloView) onEmailChanged() {
	text, _ := v.emailEntry.GetText()
	v.controller.SetEmail(text)
}

func (v *exportCabrilloView) onOpenAfterExportToggled() {
	v.controller.SetOpenAfterExport(v.openAfterExportCheckButton.GetActive())
}
