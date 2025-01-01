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
	SetLocation(string)
	SetAddressText(string)
	SetAddressCity(string)
	SetAddressPostalCode(string)
	SetAddressStateProvince(string)
	SetAddressCountry(string)
	SetClub(string)
	SetSpecific(string)
	SetCertificate(bool)
	SetSoapBox(string)
	SetOpenUploadAfterExport(bool)
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

	nameEntry                 *gtk.Entry
	emailEntry                *gtk.Entry
	locationEntry             *gtk.Entry
	addressTextEntry          *gtk.Entry
	addressCityEntry          *gtk.Entry
	addressPostalCodeEntry    *gtk.Entry
	addressStateProvinceEntry *gtk.Entry
	addressCountryEntry       *gtk.Entry
	clubEntry                 *gtk.Entry
	specificEntry             *gtk.Entry

	certificateCheckButton *gtk.CheckButton
	soapBoxEntry           *gtk.TextView

	openUploadAfterExportCheckButton *gtk.CheckButton
	openAfterExportCheckButton       *gtk.CheckButton
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
	result.root.SetMarginStart(5)
	result.root.SetMarginEnd(5)

	columns, _ := gtk.GridNew()
	columns.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	columns.SetHExpand(true)
	columns.SetVExpand(false)
	columns.SetColumnSpacing(20)
	result.root.Attach(columns, 0, 1, 1, 1)

	leftColumn, _ := gtk.GridNew()
	leftColumn.SetOrientation(gtk.ORIENTATION_VERTICAL)
	leftColumn.SetHExpand(true)
	leftColumn.SetVExpand(false)
	leftColumn.SetColumnSpacing(5)
	leftColumn.SetRowSpacing(5)
	columns.Attach(leftColumn, 0, 0, 1, 1)

	rightColumn, _ := gtk.GridNew()
	rightColumn.SetOrientation(gtk.ORIENTATION_VERTICAL)
	rightColumn.SetHExpand(true)
	rightColumn.SetVExpand(false)
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
	buildSeparator(leftColumn, 8, 2)
	result.categoryStationCombo = buildLabeledCombo(leftColumn, 9, "Station", false, result.controller.CategoryStations(), result.onCategoryStationChanged)
	result.categoryTransmitterCombo = buildLabeledCombo(leftColumn, 10, "Transmitter", false, result.controller.CategoryTransmitters(), result.onCategoryTransmitterChanged)
	result.categoryOverlayCombo = buildLabeledCombo(leftColumn, 11, "Overlay", true, result.controller.CategoryOverlays(), result.onCategoryOverlayChanged)
	result.categoryTimeCombo = buildLabeledCombo(leftColumn, 12, "Time", true, result.controller.CategoryTimes(), result.onCategoryTimeChanged)

	buildHeaderLabel(rightColumn, 0, "Personal Information")
	result.nameEntry = buildLabeledEntry(rightColumn, 1, "Name", result.onNameChanged)
	result.emailEntry = buildLabeledEntry(rightColumn, 2, "Email", result.onEmailChanged)
	result.locationEntry = buildLabeledEntry(rightColumn, 3, "Location", result.onLocationChanged)
	buildSeparator(rightColumn, 4, 2)
	result.addressTextEntry = buildLabeledEntry(rightColumn, 5, "Address", result.onAddressTextChanged)
	result.addressCityEntry = buildLabeledEntry(rightColumn, 6, "City", result.onAddressCityChanged)
	result.addressPostalCodeEntry = buildLabeledEntry(rightColumn, 7, "Postal Code", result.onAddressPostalCodeChanged)
	result.addressStateProvinceEntry = buildLabeledEntry(rightColumn, 8, "State/Province", result.onAddressStateProvinceChanged)
	result.addressCountryEntry = buildLabeledEntry(rightColumn, 9, "Country", result.onAddressCountryChanged)
	buildSeparator(rightColumn, 10, 2)
	result.clubEntry = buildLabeledEntry(rightColumn, 11, "Club", result.onClubChanged)
	result.specificEntry = buildLabeledEntry(rightColumn, 12, "Specific", result.onSpecificChanged)

	buildSeparator(result.root, 2, 1)

	result.certificateCheckButton = buildCheckButton(result.root, 3, "Request a certificate", result.onCertificateToggled)
	result.soapBoxEntry = buildLabeledTextView(result.root, 4, "Soap Box", result.onSoapBoxChanged)

	buildSeparator(result.root, 6, 1)

	result.openUploadAfterExportCheckButton = buildCheckButton(result.root, 7, "Open the upload URL after export", result.onOpenUploadAfterExportToggled)
	result.openAfterExportCheckButton = buildCheckButton(result.root, 8, "Open the file after export", result.onOpenAfterExportToggled)

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

func (v *exportCabrilloView) onLocationChanged() {
	text, _ := v.locationEntry.GetText()
	v.controller.SetLocation(text)
}

func (v *exportCabrilloView) onAddressTextChanged() {
	text, _ := v.addressTextEntry.GetText()
	v.controller.SetAddressText(text)
}

func (v *exportCabrilloView) onAddressCityChanged() {
	text, _ := v.addressCityEntry.GetText()
	v.controller.SetAddressCity(text)
}

func (v *exportCabrilloView) onAddressPostalCodeChanged() {
	text, _ := v.addressPostalCodeEntry.GetText()
	v.controller.SetAddressPostalCode(text)
}

func (v *exportCabrilloView) onAddressStateProvinceChanged() {
	text, _ := v.addressStateProvinceEntry.GetText()
	v.controller.SetAddressStateProvince(text)
}

func (v *exportCabrilloView) onAddressCountryChanged() {
	text, _ := v.addressCountryEntry.GetText()
	v.controller.SetAddressCountry(text)
}

func (v *exportCabrilloView) onClubChanged() {
	text, _ := v.clubEntry.GetText()
	v.controller.SetClub(text)
}

func (v *exportCabrilloView) onSpecificChanged() {
	text, _ := v.specificEntry.GetText()
	v.controller.SetSpecific(text)
}

func (v *exportCabrilloView) onCertificateToggled() {
	v.controller.SetCertificate(v.certificateCheckButton.GetActive())
}

func (v *exportCabrilloView) onSoapBoxChanged() {
	buffer, _ := v.soapBoxEntry.GetBuffer()
	text, _ := buffer.GetText(buffer.GetStartIter(), buffer.GetEndIter(), true)
	v.controller.SetSoapBox(text)
}

func (v *exportCabrilloView) onOpenUploadAfterExportToggled() {
	v.controller.SetOpenUploadAfterExport(v.openUploadAfterExportCheckButton.GetActive())
}

func (v *exportCabrilloView) onOpenAfterExportToggled() {
	v.controller.SetOpenAfterExport(v.openAfterExportCheckButton.GetActive())
}
