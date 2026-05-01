package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core/export/cabrillo"
)

var _ cabrillo.View = (*exportCabrilloDialog)(nil)

type exportCabrilloDialog struct {
	parent     *qtlib.QMainWindow
	controller ExportCabrilloController

	dialog *qtlib.QDialog
	view   *exportCabrilloView

	// Cached state, retained across Show calls.
	categoryBand        string
	categoryMode        string
	categoryOperator    string
	categoryPower       string
	categoryAssisted    string
	categoryStation     string
	categoryTransmitter string
	categoryOverlay     string
	categoryTime        string

	name                 string
	email                string
	location             string
	addressText          string
	addressCity          string
	addressPostalCode    string
	addressStateProvince string
	addressCountry       string
	club                 string
	specific             string

	certificate bool
	soapBox     string

	openUploadAfterExport bool
	openAfterExport       bool
}

func newExportCabrilloDialog(parent *qtlib.QMainWindow, controller ExportCabrilloController) *exportCabrilloDialog {
	return &exportCabrilloDialog{parent: parent, controller: controller}
}

func (d *exportCabrilloDialog) Show() bool {
	d.view = newExportCabrilloView(d.controller)

	d.view.doIgnore(func() {
		selectComboValue(d.view.categoriesCombo, "")
		selectComboValue(d.view.categoryBandCombo, d.categoryBand)
		selectComboValue(d.view.categoryModeCombo, d.categoryMode)
		selectComboValue(d.view.categoryOperatorCombo, d.categoryOperator)
		selectComboValue(d.view.categoryPowerCombo, d.categoryPower)
		selectComboValue(d.view.categoryAssistedCombo, d.categoryAssisted)
		selectComboValue(d.view.categoryStationCombo, d.categoryStation)
		selectComboValue(d.view.categoryTransmitterCombo, d.categoryTransmitter)
		selectComboValue(d.view.categoryOverlayCombo, d.categoryOverlay)
		selectComboValue(d.view.categoryTimeCombo, d.categoryTime)

		d.view.nameEntry.SetText(d.name)
		d.view.emailEntry.SetText(d.email)
		d.view.locationEntry.SetText(d.location)
		d.view.addressTextEntry.SetText(d.addressText)
		d.view.addressCityEntry.SetText(d.addressCity)
		d.view.addressPostalCodeEntry.SetText(d.addressPostalCode)
		d.view.addressStateProvinceEntry.SetText(d.addressStateProvince)
		d.view.addressCountryEntry.SetText(d.addressCountry)
		d.view.clubEntry.SetText(d.club)
		d.view.specificEntry.SetText(d.specific)

		d.view.certificateChk.SetChecked(d.certificate)
		d.view.soapBoxEdit.SetPlainText(d.soapBox)

		d.view.openUploadAfterExport.SetChecked(d.openUploadAfterExport)
		d.view.openAfterExport.SetChecked(d.openAfterExport)
	})

	d.dialog = qtlib.NewQDialog(d.parent.QWidget)
	d.dialog.SetWindowTitle("Export Cabrillo")
	d.dialog.SetModal(true)
	d.dialog.SetMinimumSize(qtlib.NewQSize2(800, 600))
	d.dialog.SetWindowFlags(
		qtlib.Window |
			qtlib.CustomizeWindowHint |
			qtlib.WindowTitleHint |
			qtlib.WindowCloseButtonHint,
	)

	root := qtlib.NewQVBoxLayout(d.dialog.QWidget)
	root.AddWidget(d.view.root)

	buttons := qtlib.NewQDialogButtonBox4(
		qtlib.QDialogButtonBox__Ok | qtlib.QDialogButtonBox__Cancel,
	)
	buttons.Button(qtlib.QDialogButtonBox__Ok).SetText("Export")
	buttons.OnAccepted(func() { d.dialog.Accept() })
	buttons.OnRejected(func() { d.dialog.Reject() })
	root.AddWidget(buttons.QWidget)

	accepted := d.dialog.Exec() == int(qtlib.QDialog__Accepted)
	d.dialog.DeleteLater()
	d.dialog = nil
	d.view = nil
	return accepted
}

func (d *exportCabrilloDialog) SetCategoryBand(v string) {
	d.categoryBand = v
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.categoryBandCombo, v) })
	}
}

func (d *exportCabrilloDialog) SetCategoryMode(v string) {
	d.categoryMode = v
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.categoryModeCombo, v) })
	}
}

func (d *exportCabrilloDialog) SetCategoryOperator(v string) {
	d.categoryOperator = v
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.categoryOperatorCombo, v) })
	}
}

func (d *exportCabrilloDialog) SetCategoryPower(v string) {
	d.categoryPower = v
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.categoryPowerCombo, v) })
	}
}

func (d *exportCabrilloDialog) SetCategoryAssisted(v string) {
	d.categoryAssisted = v
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.categoryAssistedCombo, v) })
	}
}

func (d *exportCabrilloDialog) SetCategoryStation(v string) {
	d.categoryStation = v
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.categoryStationCombo, v) })
	}
}

func (d *exportCabrilloDialog) SetCategoryTransmitter(v string) {
	d.categoryTransmitter = v
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.categoryTransmitterCombo, v) })
	}
}

func (d *exportCabrilloDialog) SetCategoryOverlay(v string) {
	d.categoryOverlay = v
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.categoryOverlayCombo, v) })
	}
}

func (d *exportCabrilloDialog) SetCategoryTime(v string) {
	d.categoryTime = v
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.categoryTimeCombo, v) })
	}
}

func (d *exportCabrilloDialog) SetName(v string) {
	d.name = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.nameEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetEmail(v string) {
	d.email = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.emailEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetLocation(v string) {
	d.location = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.locationEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetAddressText(v string) {
	d.addressText = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.addressTextEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetAddressCity(v string) {
	d.addressCity = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.addressCityEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetAddressPostalCode(v string) {
	d.addressPostalCode = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.addressPostalCodeEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetAddressStateProvince(v string) {
	d.addressStateProvince = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.addressStateProvinceEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetAddressCountry(v string) {
	d.addressCountry = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.addressCountryEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetClub(v string) {
	d.club = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.clubEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetSpecific(v string) {
	d.specific = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.specificEntry.SetText(v) })
	}
}

func (d *exportCabrilloDialog) SetCertificate(v bool) {
	d.certificate = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.certificateChk.SetChecked(v) })
	}
}

func (d *exportCabrilloDialog) SetSoapBox(v string) {
	d.soapBox = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.soapBoxEdit.SetPlainText(v) })
	}
}

func (d *exportCabrilloDialog) SetOpenUploadAfterExport(v bool) {
	d.openUploadAfterExport = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.openUploadAfterExport.SetChecked(v) })
	}
}

func (d *exportCabrilloDialog) SetOpenAfterExport(v bool) {
	d.openAfterExport = v
	if d.view != nil {
		d.view.doIgnore(func() { d.view.openAfterExport.SetChecked(v) })
	}
}
