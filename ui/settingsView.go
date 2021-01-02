package ui

import (
	"fmt"
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
}

type settingsView struct {
	parent     *gtk.Dialog
	controller SettingsController

	ignoreChangedEvent bool

	message *gtk.Label
	reset   *gtk.Button
	close   *gtk.Button
	fields  map[string]interface{}
}

func setupSettingsView(builder *gtk.Builder, parent *gtk.Dialog, controller SettingsController) *settingsView {
	result := new(settingsView)
	result.parent = parent
	result.controller = controller
	result.fields = make(map[string]interface{})

	result.message = getUI(builder, "settingsMessageLabel").(*gtk.Label)
	result.reset = getUI(builder, "resetButton").(*gtk.Button)
	result.reset.Connect("clicked", result.onResetPressed)
	result.close = getUI(builder, "closeButton").(*gtk.Button)
	result.close.Connect("clicked", result.onClosePressed)

	result.addEntry(builder, "stationCallsignEntry")
	result.addEntry(builder, "stationOperatorEntry")
	result.addEntry(builder, "stationLocatorEntry")
	result.addEntry(builder, "contestNameEntry")
	result.addCheckButton(builder, "contestEnterTheirNumberButton")
	result.addCheckButton(builder, "contestEnterTheirXchangeButton")
	result.addCheckButton(builder, "contestRequireTheirXchangeButton")

	result.parent.Connect("destroy", result.onDestroy)

	return result
}

func (v *settingsView) SetSettingsController(controller SettingsController) {
	v.controller = controller
}
func (v *settingsView) ShowMessage(message string) {
	v.message.SetMarkup(fmt.Sprintf("<span foreground='red'>%s</span>", message))
	v.message.Show()
}

func (v *settingsView) HideMessage() {
	v.message.Hide()
}

func (v *settingsView) addEntry(builder *gtk.Builder, id string) {
	entry := getUI(builder, id).(*gtk.Entry)
	name, _ := entry.GetName()
	v.fields[name] = entry

	widget := &entry.Widget
	widget.Connect("changed", v.onFieldChanged)
}

func (v *settingsView) addCheckButton(builder *gtk.Builder, id string) {
	button := getUI(builder, id).(*gtk.CheckButton)
	name, _ := button.GetName()
	v.fields[name] = button

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

	switch field {
	case "stationCallsign":
		v.controller.EnterStationCallsign(value.(string))
	case "stationOperator":
		v.controller.EnterStationOperator(value.(string))
	case "stationLocator":
		v.controller.EnterStationLocator(value.(string))
	case "contestName":
		v.controller.EnterContestName(value.(string))
	case "contestEnterTheirNumber":
		v.controller.EnterContestEnterTheirNumber(value.(bool))
	case "contestEnterTheirXchange":
		v.controller.EnterContestEnterTheirXchange(value.(bool))
	case "contestRequireTheirXchange":
		v.controller.EnterContestRequireTheirXchange(value.(bool))
	default:
		log.Printf("enter unknown field %s: %v", field, value)
	}

	return false
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

func (v *settingsView) setEntryField(field string, value string) {
	v.doIgnoreChanges(func() {
		v.fields[field].(*gtk.Entry).SetText(value)
	})
}

func (v *settingsView) setCheckButtonField(field string, value bool) {
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
	v.setEntryField("stationCallsign", value)
}

func (v *settingsView) SetStationOperator(value string) {
	v.setEntryField("stationOperator", value)
}

func (v *settingsView) SetStationLocator(value string) {
	v.setEntryField("stationLocator", value)
}

func (v *settingsView) SetContestName(value string) {
	v.setEntryField("contestName", value)
}

func (v *settingsView) SetContestEnterTheirNumber(value bool) {
	v.setCheckButtonField("contestEnterTheirNumber", value)
}

func (v *settingsView) SetContestEnterTheirXchange(value bool) {
	v.setCheckButtonField("contestEnterTheirXchange", value)
}

func (v *settingsView) SetContestRequireTheirXchange(value bool) {
	v.setCheckButtonField("contestRequireTheirXchange", value)
}
