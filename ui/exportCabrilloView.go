package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type ExportCabrilloController interface {
	Categories() []string
	CategoryBands() []string
	CategoryModes() []string
	CategoryOperators() []string
	CategoryPowers() []string
	CategoryAssisted() []string
	CategoryStations() []string
	CategoryTransmitters() []string
	CategoryOverlays() []string
	CategoryTimes() []string

	SetCategory(string)
	SetCategoryBand(string)
	SetCategoryMode(string)
	SetCategoryOperator(string)
	SetCategoryPower(string)
	SetCategoryAssisted(string)
	SetCategoryStation(string)
	SetCategoryTransmitter(string)
	SetCategoryOverlay(string)
	SetCategoryTime(string)
	SetName(string)
	SetEmail(string)
	SetOpenAfterExport(bool)
}

type exportCabrilloView struct {
	controller ExportCabrilloController

	root *gtk.Grid

	categoriesCombo          *gtk.ComboBoxText
	categoryBandCombo        *gtk.ComboBoxText
	categoryModeCombo        *gtk.ComboBoxText
	categoryOperatorCombo    *gtk.ComboBoxText
	categoryPowerCombo       *gtk.ComboBoxText
	categoryAssistedCombo    *gtk.ComboBoxText
	categoryStationCombo     *gtk.ComboBoxText
	categoryTransmitterCombo *gtk.ComboBoxText
	categoryOverlayCombo     *gtk.ComboBoxText
	categoryTimeCombo        *gtk.ComboBoxText

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
	result.root.SetRowSpacing(5)

	columns, _ := gtk.GridNew()
	columns.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	columns.SetHExpand(true)
	columns.SetVExpand(true)
	columns.SetColumnSpacing(10)
	result.root.Attach(columns, 0, 1, 1, 1)

	leftColumn, _ := gtk.GridNew()
	leftColumn.SetOrientation(gtk.ORIENTATION_VERTICAL)
	leftColumn.SetHExpand(false)
	leftColumn.SetVExpand(true)
	leftColumn.SetColumnSpacing(5)
	leftColumn.SetRowSpacing(5)
	columns.Attach(leftColumn, 0, 0, 1, 1)

	rightColumn, _ := gtk.GridNew()
	rightColumn.SetOrientation(gtk.ORIENTATION_VERTICAL)
	rightColumn.SetHExpand(true)
	rightColumn.SetVExpand(true)
	rightColumn.SetColumnSpacing(5)
	rightColumn.SetRowSpacing(5)
	columns.Attach(rightColumn, 1, 0, 1, 1)

	buildHeaderLabel(leftColumn, 0, "Category")
	result.categoriesCombo = buildLabeledCombo(leftColumn, 1, "Category", false, result.controller.Categories(), result.onCategoryChanged)
	categoryExplanation := buildExplanationLabel(leftColumn, 2, "Choose one of the categories defined in the contest rules to fill out the Cabrillo category fields.")
	categoryExplanation.SetHExpand(false)
	categoryExplanation.SetLineWrap(true)
	result.categoryBandCombo = buildLabeledCombo(leftColumn, 3, "Band", false, result.controller.CategoryBands(), result.onCategoryBandChanged)
	result.categoryModeCombo = buildLabeledCombo(leftColumn, 4, "Mode", false, result.controller.CategoryModes(), result.onCategoryModeChanged)
	result.categoryOperatorCombo = buildLabeledCombo(leftColumn, 5, "Operator", false, result.controller.CategoryOperators(), result.onCategoryOperatorChanged)
	result.categoryPowerCombo = buildLabeledCombo(leftColumn, 6, "Power", false, result.controller.CategoryPowers(), result.onCategoryPowerChanged)
	result.categoryAssistedCombo = buildLabeledCombo(leftColumn, 7, "Assisted", false, result.controller.CategoryAssisted(), result.onCategoryAssistedChanged)
	buildSeparator(leftColumn, 8)
	result.categoryStationCombo = buildLabeledCombo(leftColumn, 9, "Station", false, result.controller.CategoryStations(), result.onCategoryStationChanged)
	result.categoryTransmitterCombo = buildLabeledCombo(leftColumn, 10, "Transmitter", false, result.controller.CategoryTransmitters(), result.onCategoryTransmitterChanged)
	result.categoryOverlayCombo = buildLabeledCombo(leftColumn, 11, "Overlay", true, result.controller.CategoryOverlays(), result.onCategoryOverlayChanged)
	result.categoryTimeCombo = buildLabeledCombo(leftColumn, 12, "Time", true, result.controller.CategoryTimes(), result.onCategoryTimeChanged)

	buildHeaderLabel(rightColumn, 0, "Personal Information")
	result.nameEntry = buildLabeledEntry(rightColumn, 1, "Name", result.onNameChanged)
	result.emailEntry = buildLabeledEntry(rightColumn, 2, "Email", result.onEmailChanged)

	buildSeparator(result.root, 2)

	result.openAfterExportCheckButton = buildCheckButton(result.root, 3, "Open the file after export", result.onOpenAfterExportToggled)

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

func (v *exportCabrilloView) onCategoryStationChanged() {
	v.controller.SetCategoryStation(v.categoryStationCombo.GetActiveText())
}

func (v *exportCabrilloView) onCategoryTransmitterChanged() {
	v.controller.SetCategoryTransmitter(v.categoryTransmitterCombo.GetActiveText())
}

func (v *exportCabrilloView) onCategoryOverlayChanged() {
	v.controller.SetCategoryOverlay(v.categoryOverlayCombo.GetActiveText())
}

func (v *exportCabrilloView) onCategoryTimeChanged() {
	v.controller.SetCategoryTime(v.categoryTimeCombo.GetActiveText())
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
