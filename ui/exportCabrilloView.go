package ui

import "github.com/gotk3/gotk3/gtk"

type ExportCabrilloController interface {
	Categories() []string
	CategoryBands() []string
	CategoryModes() []string
	CategoryOperators() []string
	CategoryPowers() []string
	CategoryAssisted() []string

	SetCategory(string)
	SetCategoryBand(string)
	SetCategoryMode(string)
	SetCategoryOperator(string)
	SetCategoryPower(string)
	SetCategoryAssisted(string)
	SetName(string)
	SetEmail(string)
	SetOpenAfterExport(bool)
}

type exportCabrilloView struct {
	controller ExportCabrilloController

	root *gtk.Grid

	categoriesCombo            *gtk.ComboBoxText
	categoryBandCombo          *gtk.ComboBoxText
	categoryModeCombo          *gtk.ComboBoxText
	categoryOperatorCombo      *gtk.ComboBoxText
	categoryPowerCombo         *gtk.ComboBoxText
	categoryAssistedCombo      *gtk.ComboBoxText
	nameEntry                  *gtk.Entry
	emailEntry                 *gtk.Entry
	openAfterExportCheckButton *gtk.CheckButton
}

func newExportCabrilloView(controller ExportCabrilloController) *exportCabrilloView {
	result := &exportCabrilloView{
		controller: controller,
	}

	result.root, _ = gtk.GridNew()
	result.root.SetOrientation(gtk.ORIENTATION_VERTICAL)
	result.root.SetHExpand(true)
	result.root.SetVExpand(true)
	result.root.SetColumnSpacing(5)

	buildExplanationLabel(result.root, 0, "Export the log as Cabrillo file.")

	result.categoriesCombo = buildLabeledCombo(result.root, 1, "Category", result.controller.Categories(), result.onCategoryChanged)
	result.categoryBandCombo = buildLabeledCombo(result.root, 2, "Band", result.controller.CategoryBands(), result.onCategoryBandChanged)
	result.categoryModeCombo = buildLabeledCombo(result.root, 3, "Mode", result.controller.CategoryModes(), result.onCategoryModeChanged)
	result.categoryOperatorCombo = buildLabeledCombo(result.root, 4, "Operator", result.controller.CategoryOperators(), result.onCategoryOperatorChanged)
	result.categoryPowerCombo = buildLabeledCombo(result.root, 5, "Power", result.controller.CategoryPowers(), result.onCategoryPowerChanged)
	result.categoryAssistedCombo = buildLabeledCombo(result.root, 6, "Assisted", result.controller.CategoryAssisted(), result.onCategoryAssistedChanged)

	buildSeparator(result.root, 7)

	result.nameEntry = buildLabeledEntry(result.root, 8, "Name", result.onNameChanged)
	result.emailEntry = buildLabeledEntry(result.root, 9, "Email", result.onEmailChanged)

	buildSeparator(result.root, 10)

	result.openAfterExportCheckButton = buildCheckButton(result.root, 11, "Open the file after export", result.onOpenAfterExportToggled)

	return result
}

func (v *exportCabrilloView) onCategoryChanged() {
	v.controller.SetCategory(v.categoriesCombo.GetActiveText())
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

func (v *exportCabrilloView) onCategoryAssistedChanged() {
	v.controller.SetCategoryAssisted(v.categoryAssistedCombo.GetActiveText())
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
