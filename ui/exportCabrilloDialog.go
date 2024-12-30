package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type exportCabrilloDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller ExportCabrilloController
	view       *exportCabrilloView

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
	d.view = newExportCabrilloView(d.controller)
	d.view.categoriesCombo.SetActiveID("")
	d.view.categoryBandCombo.SetActiveID(d.categoryBand)
	d.view.categoryModeCombo.SetActiveID(d.categoryMode)
	d.view.categoryOperatorCombo.SetActiveID(d.categoryOperator)
	d.view.categoryPowerCombo.SetActiveID(d.categoryPower)
	d.view.categoryAssistedCombo.SetActiveID(d.categoryAssisted)
	d.view.categoryStationCombo.SetActiveID(d.categoryStation)
	d.view.categoryTransmitterCombo.SetActiveID(d.categoryTransmitter)
	d.view.categoryOverlayCombo.SetActiveID(d.categoryOverlay)
	d.view.categoryTimeCombo.SetActiveID(d.categoryTime)
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
	d.view.certificateCheckButton.SetActive(d.certificate)
	buffer, _ := d.view.soapBoxEntry.GetBuffer()
	buffer.SetText(d.soapBox)
	// d.view.soapBoxEntry.SetBuffer(buffer)
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

func (d *exportCabrilloDialog) SetCategoryStation(station string) {
	d.categoryStation = station
	if d.view != nil {
		d.view.categoryStationCombo.SetActiveID(station)
	}
}
func (d *exportCabrilloDialog) SetCategoryTransmitter(transmitter string) {
	d.categoryTransmitter = transmitter
	if d.view != nil {
		d.view.categoryTransmitterCombo.SetActiveID(transmitter)
	}
}
func (d *exportCabrilloDialog) SetCategoryOverlay(overlay string) {
	d.categoryOverlay = overlay
	if d.view != nil {
		d.view.categoryOverlayCombo.SetActiveID(overlay)
	}
}
func (d *exportCabrilloDialog) SetCategoryTime(time string) {
	d.categoryTime = time
	if d.view != nil {
		d.view.categoryTimeCombo.SetActiveID(time)
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

func (d *exportCabrilloDialog) SetLocation(location string) {
	d.location = location
	if d.view != nil {
		d.view.locationEntry.SetText(location)
	}
}

func (d *exportCabrilloDialog) SetAddressText(addressText string) {
	d.addressText = addressText
	if d.view != nil {
		d.view.addressTextEntry.SetText(addressText)
	}
}

func (d *exportCabrilloDialog) SetAddressCity(addressCity string) {
	d.addressCity = addressCity
	if d.view != nil {
		d.view.addressCityEntry.SetText(addressCity)
	}
}

func (d *exportCabrilloDialog) SetAddressPostalCode(addressPostalCode string) {
	d.addressPostalCode = addressPostalCode
	if d.view != nil {
		d.view.addressPostalCodeEntry.SetText(addressPostalCode)
	}
}

func (d *exportCabrilloDialog) SetAddressStateProvince(addressStateProvince string) {
	d.addressStateProvince = addressStateProvince
	if d.view != nil {
		d.view.addressStateProvinceEntry.SetText(addressStateProvince)
	}
}

func (d *exportCabrilloDialog) SetAddressCountry(addressCountry string) {
	d.addressCountry = addressCountry
	if d.view != nil {
		d.view.addressCountryEntry.SetText(addressCountry)
	}
}

func (d *exportCabrilloDialog) SetClub(club string) {
	d.club = club
	if d.view != nil {
		d.view.clubEntry.SetText(club)
	}
}

func (d *exportCabrilloDialog) SetSpecific(specific string) {
	d.specific = specific
	if d.view != nil {
		d.view.specificEntry.SetText(specific)
	}
}

func (d *exportCabrilloDialog) SetCertificate(certificate bool) {
	d.certificate = certificate
	if d.view != nil {
		d.view.certificateCheckButton.SetActive(certificate)
	}
}

func (d *exportCabrilloDialog) SetSoapBox(soapBox string) {
	d.soapBox = soapBox
	if d.view != nil {
		buffer, _ := d.view.soapBoxEntry.GetBuffer()
		buffer.SetText(soapBox)
	}
}

func (d *exportCabrilloDialog) SetOpenAfterExport(open bool) {
	d.openAfterExport = open
}
