package ui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
)

type SettingsController interface {
	Save()
	Reset()

	EnterStationCallsign(string)
	EnterStationOperator(string)
	EnterStationLocator(string)
	EnterContestName(string)
	EnterContestEnterTheirNumber(bool)
	EnterContestEnterTheirXchange(bool)
	EnterContestRequireTheirXchange(bool)
	EnterContestAllowMultiBand(bool)
	EnterContestAllowMultiMode(bool)
	EnterContestSameCountryPoints(string)
	EnterContestSameContinentPoints(string)
	EnterContestSpecificCountryPoints(string)
	EnterContestSpecificCountryPrefixes(string)
	EnterContestOtherPoints(string)
	EnterContestMultis(dxcc, wpx, xchange bool)
	EnterContestXchangeMultiPattern(string)
	EnterContestCountPerBand(bool)
}

type fieldID string

const (
	stationCallsign                fieldID = "stationCallsign"
	stationOperator                fieldID = "stationOperator"
	stationLocator                 fieldID = "stationLocator"
	contestName                    fieldID = "contestName"
	contestEnterTheirNumber        fieldID = "contestEnterTheirNumber"
	contestEnterTheirXchange       fieldID = "contestEnterTheirXchange"
	contestRequireTheirXchange     fieldID = "contestRequireTheirXchange"
	contestAllowMultiBand          fieldID = "contestAllowMultiBand"
	contestAllowMultiMode          fieldID = "contestAllowMultiMode"
	contestSameCountryPoints       fieldID = "contestSameCountryPoints"
	contestSameContinentPoints     fieldID = "contestSameContinentPoints"
	contestSpecificCountryPoints   fieldID = "contestSpecificCountryPoints"
	contestSpecificCountryPrefixes fieldID = "contestSpecificCountryPrefixes"
	contestOtherPoints             fieldID = "contestOtherPoints"
	contestMultiDXCC               fieldID = "contestMultiDXCC"
	contestMultiWPX                fieldID = "contestMultiWPX"
	contestMultiXchange            fieldID = "contestMultiXchange"
	contestXchangeMultiPattern     fieldID = "contestXchangeMultiPattern"
	contestCountPerBand            fieldID = "contestCountPerBand"
)

type settingsView struct {
	parent     *gtk.Dialog
	controller SettingsController

	ignoreChangedEvent bool

	message *gtk.Label
	reset   *gtk.Button
	close   *gtk.Button
	fields  map[fieldID]interface{}
}

func setupSettingsView(builder *gtk.Builder, parent *gtk.Dialog, controller SettingsController) *settingsView {
	result := new(settingsView)
	result.parent = parent
	result.controller = controller
	result.fields = make(map[fieldID]interface{})

	result.message = getUI(builder, "settingsMessageLabel").(*gtk.Label)
	result.reset = getUI(builder, "resetButton").(*gtk.Button)
	result.reset.Connect("clicked", result.onResetPressed)
	result.close = getUI(builder, "closeButton").(*gtk.Button)
	result.close.Connect("clicked", result.onClosePressed)

	result.addEntry(builder, stationCallsign)
	result.addEntry(builder, stationOperator)
	result.addEntry(builder, stationLocator)
	result.addEntry(builder, contestName)
	result.addCheckButton(builder, contestEnterTheirNumber)
	result.addCheckButton(builder, contestEnterTheirXchange)
	result.addCheckButton(builder, contestRequireTheirXchange)
	result.addCheckButton(builder, contestAllowMultiBand)
	result.addCheckButton(builder, contestAllowMultiMode)
	result.addEntry(builder, contestSameCountryPoints)
	result.addEntry(builder, contestSameContinentPoints)
	result.addEntry(builder, contestSpecificCountryPoints)
	result.addEntry(builder, contestSpecificCountryPrefixes)
	result.addEntry(builder, contestOtherPoints)
	result.addCheckButton(builder, contestMultiDXCC)
	result.addCheckButton(builder, contestMultiWPX)
	result.addCheckButton(builder, contestMultiXchange)
	result.addEntry(builder, contestXchangeMultiPattern)
	result.addCheckButton(builder, contestCountPerBand)

	result.parent.Connect("destroy", result.onDestroy)

	return result
}

func (v *settingsView) SetSettingsController(controller SettingsController) {
	v.controller = controller
}
func (v *settingsView) ShowMessage(message string) {
	v.message.SetText(message)
	v.message.Show()
}

func (v *settingsView) HideMessage() {
	v.message.Hide()
}

func (v *settingsView) addEntry(builder *gtk.Builder, id fieldID) {
	entry := getUI(builder, string(id)+"Entry").(*gtk.Entry)
	field, _ := entry.GetName()
	v.fields[fieldID(field)] = entry

	widget := &entry.Widget
	widget.Connect("changed", v.onFieldChanged)
}

func (v *settingsView) addCheckButton(builder *gtk.Builder, id fieldID) {
	button := getUI(builder, string(id)+"Button").(*gtk.CheckButton)
	field, _ := button.GetName()
	v.fields[fieldID(field)] = button

	widget := &button.Widget
	widget.Connect("toggled", v.onFieldChanged)
}

func (v *settingsView) onFieldChanged(w interface{}) bool {
	if v.ignoreChangedEvent {
		return false
	}

	var field string
	var value interface{}
	switch widget := w.(type) {
	case *gtk.Entry:
		field, _ = widget.GetName()
		value, _ = widget.GetText()
	case *gtk.CheckButton:
		field, _ = widget.GetName()
		value = widget.GetActive()
	default:
		return false
	}

	switch fieldID(field) {
	case stationCallsign:
		v.controller.EnterStationCallsign(value.(string))
	case stationOperator:
		v.controller.EnterStationOperator(value.(string))
	case stationLocator:
		v.controller.EnterStationLocator(value.(string))
	case contestName:
		v.controller.EnterContestName(value.(string))
	case contestEnterTheirNumber:
		v.controller.EnterContestEnterTheirNumber(value.(bool))
	case contestEnterTheirXchange:
		v.controller.EnterContestEnterTheirXchange(value.(bool))
	case contestRequireTheirXchange:
		v.controller.EnterContestRequireTheirXchange(value.(bool))
	case contestAllowMultiBand:
		v.controller.EnterContestAllowMultiBand(value.(bool))
	case contestAllowMultiMode:
		v.controller.EnterContestAllowMultiMode(value.(bool))
	case contestSameCountryPoints:
		v.controller.EnterContestSameCountryPoints(value.(string))
	case contestSameContinentPoints:
		v.controller.EnterContestSameContinentPoints(value.(string))
	case contestSpecificCountryPoints:
		v.controller.EnterContestSpecificCountryPoints(value.(string))
	case contestSpecificCountryPrefixes:
		v.controller.EnterContestSpecificCountryPrefixes(value.(string))
	case contestOtherPoints:
		v.controller.EnterContestOtherPoints(value.(string))
	case contestMultiDXCC, contestMultiWPX, contestMultiXchange:
		v.controller.EnterContestMultis(v.multis())
	case contestXchangeMultiPattern:
		v.controller.EnterContestXchangeMultiPattern(value.(string))
	case contestCountPerBand:
		v.controller.EnterContestCountPerBand(value.(bool))
	default:
		log.Printf("enter unknown field %s: %v", field, value)
	}

	return false
}

func (v *settingsView) multis() (dxcc, wpx, xchange bool) {
	dxcc = v.fields[contestMultiDXCC].(*gtk.CheckButton).GetActive()
	wpx = v.fields[contestMultiWPX].(*gtk.CheckButton).GetActive()
	xchange = v.fields[contestMultiXchange].(*gtk.CheckButton).GetActive()
	return
}

func (v *settingsView) onResetPressed(_ *gtk.Button) {
	v.controller.Reset()
}

func (v *settingsView) onClosePressed(_ *gtk.Button) {
	v.parent.Close()
}

func (v *settingsView) onDestroy() {
	v.controller.Save()
}

func (v *settingsView) setEntryField(field fieldID, value string) {
	v.doIgnoreChanges(func() {
		v.fields[field].(*gtk.Entry).SetText(value)
	})
}

func (v *settingsView) setCheckButtonField(field fieldID, value bool) {
	v.doIgnoreChanges(func() {
		v.fields[field].(*gtk.CheckButton).SetActive(value)
	})
}

func (v *settingsView) doIgnoreChanges(f func()) {
	if v == nil {
		return
	}

	v.ignoreChangedEvent = true
	defer func() {
		v.ignoreChangedEvent = false
	}()
	f()
}

func (v *settingsView) SetStationCallsign(value string) {
	v.setEntryField(stationCallsign, value)
}

func (v *settingsView) SetStationOperator(value string) {
	v.setEntryField(stationOperator, value)
}

func (v *settingsView) SetStationLocator(value string) {
	v.setEntryField(stationLocator, value)
}

func (v *settingsView) SetContestName(value string) {
	v.setEntryField(contestName, value)
}

func (v *settingsView) SetContestEnterTheirNumber(value bool) {
	v.setCheckButtonField(contestEnterTheirNumber, value)
}

func (v *settingsView) SetContestEnterTheirXchange(value bool) {
	v.setCheckButtonField(contestEnterTheirXchange, value)
}

func (v *settingsView) SetContestRequireTheirXchange(value bool) {
	v.setCheckButtonField(contestRequireTheirXchange, value)
}

func (v *settingsView) SetContestAllowMultiBand(value bool) {
	v.setCheckButtonField(contestAllowMultiBand, value)
}

func (v *settingsView) SetContestAllowMultiMode(value bool) {
	v.setCheckButtonField(contestAllowMultiMode, value)
}

func (v *settingsView) SetContestSameCountryPoints(value string) {
	v.setEntryField(contestSameCountryPoints, value)
}

func (v *settingsView) SetContestSameContinentPoints(value string) {
	v.setEntryField(contestSameContinentPoints, value)
}

func (v *settingsView) SetContestSpecificCountryPoints(value string) {
	v.setEntryField(contestSpecificCountryPoints, value)
}

func (v *settingsView) SetContestSpecificCountryPrefixes(value string) {
	v.setEntryField(contestSpecificCountryPrefixes, value)
}

func (v *settingsView) SetContestOtherPoints(value string) {
	v.setEntryField(contestOtherPoints, value)
}

func (v *settingsView) SetContestMultis(dxcc, wpx, xchange bool) {
	v.setCheckButtonField(contestMultiDXCC, dxcc)
	v.setCheckButtonField(contestMultiWPX, wpx)
	v.setCheckButtonField(contestMultiXchange, xchange)
}

func (v *settingsView) SetContestXchangeMultiPattern(value string) {
	v.setEntryField(contestXchangeMultiPattern, value)
}

func (v *settingsView) SetContestCountPerBand(value bool) {
	v.setCheckButtonField(contestCountPerBand, value)
}
