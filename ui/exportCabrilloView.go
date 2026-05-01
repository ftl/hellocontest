package ui

import (
	qtlib "github.com/mappu/miqt/qt6"
)

// ExportCabrilloController is the callback surface for the export-cabrillo dialog.
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

	root *qtlib.QWidget

	categoriesCombo          *qtlib.QComboBox
	categoryBandCombo        *qtlib.QComboBox
	categoryModeCombo        *qtlib.QComboBox
	categoryOperatorCombo    *qtlib.QComboBox
	categoryPowerCombo       *qtlib.QComboBox
	categoryAssistedCombo    *qtlib.QComboBox
	categoryStationCombo     *qtlib.QComboBox
	categoryTransmitterCombo *qtlib.QComboBox
	categoryOverlayCombo     *qtlib.QComboBox
	categoryTimeCombo        *qtlib.QComboBox

	nameEntry                 *qtlib.QLineEdit
	emailEntry                *qtlib.QLineEdit
	locationEntry             *qtlib.QLineEdit
	addressTextEntry          *qtlib.QLineEdit
	addressCityEntry          *qtlib.QLineEdit
	addressPostalCodeEntry    *qtlib.QLineEdit
	addressStateProvinceEntry *qtlib.QLineEdit
	addressCountryEntry       *qtlib.QLineEdit
	clubEntry                 *qtlib.QLineEdit
	specificEntry             *qtlib.QLineEdit

	certificateChk        *qtlib.QCheckBox
	soapBoxEdit           *qtlib.QTextEdit
	openUploadAfterExport *qtlib.QCheckBox
	openAfterExport       *qtlib.QCheckBox

	ignoreChangedEvent bool
}

func (v *exportCabrilloView) doIgnore(f func()) {
	v.ignoreChangedEvent = true
	defer func() { v.ignoreChangedEvent = false }()
	f()
}

func newExportCabrilloView(controller ExportCabrilloController) *exportCabrilloView {
	v := &exportCabrilloView{controller: controller}

	v.root = qtlib.NewQWidget2()
	root := qtlib.NewQVBoxLayout(v.root)

	columns := qtlib.NewQWidget2()
	columnsLayout := qtlib.NewQHBoxLayout(columns)
	columnsLayout.SetContentsMargins(0, 0, 0, 0)
	columnsLayout.SetSpacing(20)

	columnsLayout.AddWidget(v.buildLeftColumn())
	columnsLayout.AddWidget(v.buildRightColumn())

	root.AddWidget(columns)

	// Soap box section
	soapBoxLabel := qtlib.NewQLabel3("Soap Box:")
	root.AddWidget(soapBoxLabel.QWidget)
	v.soapBoxEdit = qtlib.NewQTextEdit2()
	v.soapBoxEdit.OnTextChanged(func() {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetSoapBox(v.soapBoxEdit.ToPlainText())
	})
	root.AddWidget(v.soapBoxEdit.QWidget)

	v.certificateChk = qtlib.NewQCheckBox3("Request a certificate")
	v.certificateChk.OnToggled(func(checked bool) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetCertificate(checked)
	})
	root.AddWidget(v.certificateChk.QWidget)

	v.openUploadAfterExport = qtlib.NewQCheckBox3("Open the upload URL after export")
	v.openUploadAfterExport.OnToggled(func(checked bool) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetOpenUploadAfterExport(checked)
	})
	root.AddWidget(v.openUploadAfterExport.QWidget)

	v.openAfterExport = qtlib.NewQCheckBox3("Open the file after export")
	v.openAfterExport.OnToggled(func(checked bool) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetOpenAfterExport(checked)
	})
	root.AddWidget(v.openAfterExport.QWidget)

	return v
}

func (v *exportCabrilloView) buildLeftColumn() *qtlib.QWidget {
	column := qtlib.NewQWidget2()
	col := qtlib.NewQVBoxLayout(column)
	col.SetContentsMargins(0, 0, 0, 0)

	header := qtlib.NewQLabel3("Category")
	header.SetStyleSheet(BoldSectionStyle)
	col.AddWidget(header.QWidget)

	form := qtlib.NewQFormLayout2()

	v.categoriesCombo = v.makeComboNoCallback("")
	for _, item := range v.controller.Categories() {
		v.categoriesCombo.AddItem(item)
	}
	v.categoriesCombo.OnCurrentTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetCategory(text)
	})
	form.AddRow3("Category:", v.categoriesCombo.QWidget)
	categoryHint := qtlib.NewQLabel3("Choose one of the categories defined in the contest rules to fill out the Cabrillo category fields.")
	categoryHint.SetWordWrap(true)
	form.AddRow3("", categoryHint.QWidget)

	v.categoryBandCombo = v.buildCategoryCombo(v.controller.CategoryBands(), v.controller.SetCategoryBand)
	form.AddRow3("Band:", v.categoryBandCombo.QWidget)

	v.categoryModeCombo = v.buildCategoryCombo(v.controller.CategoryModes(), v.controller.SetCategoryMode)
	form.AddRow3("Mode:", v.categoryModeCombo.QWidget)

	v.categoryOperatorCombo = v.buildCategoryCombo(v.controller.CategoryOperators(), v.controller.SetCategoryOperator)
	form.AddRow3("Operator:", v.categoryOperatorCombo.QWidget)

	v.categoryPowerCombo = v.buildCategoryCombo(v.controller.CategoryPowers(), v.controller.SetCategoryPower)
	form.AddRow3("Power:", v.categoryPowerCombo.QWidget)

	v.categoryAssistedCombo = v.buildCategoryCombo(v.controller.CategoryAssisted(), v.controller.SetCategoryAssisted)
	form.AddRow3("Assisted:", v.categoryAssistedCombo.QWidget)

	v.categoryStationCombo = v.buildCategoryCombo(v.controller.CategoryStations(), v.controller.SetCategoryStation)
	form.AddRow3("Station:", v.categoryStationCombo.QWidget)

	v.categoryTransmitterCombo = v.buildCategoryCombo(v.controller.CategoryTransmitters(), v.controller.SetCategoryTransmitter)
	form.AddRow3("Transmitter:", v.categoryTransmitterCombo.QWidget)

	v.categoryOverlayCombo = v.buildCategoryCombo(v.controller.CategoryOverlays(), v.controller.SetCategoryOverlay)
	form.AddRow3("Overlay:", v.categoryOverlayCombo.QWidget)

	v.categoryTimeCombo = v.buildCategoryCombo(v.controller.CategoryTimes(), v.controller.SetCategoryTime)
	form.AddRow3("Time:", v.categoryTimeCombo.QWidget)

	col.AddLayout(form.QLayout)
	col.AddStretch()
	return column
}

func (v *exportCabrilloView) buildRightColumn() *qtlib.QWidget {
	column := qtlib.NewQWidget2()
	col := qtlib.NewQVBoxLayout(column)
	col.SetContentsMargins(0, 0, 0, 0)

	header := qtlib.NewQLabel3("Personal Information")
	header.SetStyleSheet(BoldSectionStyle)
	col.AddWidget(header.QWidget)

	form := qtlib.NewQFormLayout2()

	v.nameEntry = v.buildEntry(v.controller.SetName)
	form.AddRow3("Name:", v.nameEntry.QWidget)
	v.emailEntry = v.buildEntry(v.controller.SetEmail)
	form.AddRow3("Email:", v.emailEntry.QWidget)
	v.locationEntry = v.buildEntry(v.controller.SetLocation)
	form.AddRow3("Location:", v.locationEntry.QWidget)

	v.addressTextEntry = v.buildEntry(v.controller.SetAddressText)
	form.AddRow3("Address:", v.addressTextEntry.QWidget)
	v.addressCityEntry = v.buildEntry(v.controller.SetAddressCity)
	form.AddRow3("City:", v.addressCityEntry.QWidget)
	v.addressPostalCodeEntry = v.buildEntry(v.controller.SetAddressPostalCode)
	form.AddRow3("Postal Code:", v.addressPostalCodeEntry.QWidget)
	v.addressStateProvinceEntry = v.buildEntry(v.controller.SetAddressStateProvince)
	form.AddRow3("State/Province:", v.addressStateProvinceEntry.QWidget)
	v.addressCountryEntry = v.buildEntry(v.controller.SetAddressCountry)
	form.AddRow3("Country:", v.addressCountryEntry.QWidget)

	v.clubEntry = v.buildEntry(v.controller.SetClub)
	form.AddRow3("Club:", v.clubEntry.QWidget)
	v.specificEntry = v.buildEntry(v.controller.SetSpecific)
	form.AddRow3("Specific:", v.specificEntry.QWidget)

	col.AddLayout(form.QLayout)
	col.AddStretch()
	return column
}

func (v *exportCabrilloView) makeComboNoCallback(_ string) *qtlib.QComboBox {
	return qtlib.NewQComboBox2()
}

func (v *exportCabrilloView) buildCategoryCombo(items []string, setter func(string)) *qtlib.QComboBox {
	combo := qtlib.NewQComboBox2()
	combo.AddItem("")
	for _, item := range items {
		combo.AddItem(item)
	}
	combo.OnCurrentTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		setter(text)
	})
	return combo
}

func (v *exportCabrilloView) buildEntry(setter func(string)) *qtlib.QLineEdit {
	edit := qtlib.NewQLineEdit2()
	edit.OnTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		setter(text)
	})
	return edit
}
