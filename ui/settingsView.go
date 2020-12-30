package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"
)

type SettingsController interface {
	EnterStationCallsign(string)
	EnterContestEnterTheirNumber(bool)
}

type settingsView struct {
	controller SettingsController

	ignoreChangedEvent bool

	message *gtk.Label
	fields  map[string]interface{}
}

func setupSettingsView(builder *gtk.Builder, controller SettingsController) *settingsView {
	log.Println("setting up the settings view")
	result := new(settingsView)
	result.controller = controller
	result.fields = make(map[string]interface{})

	result.message = getUI(builder, "settingsMessageLabel").(*gtk.Label)

	result.addEntry(builder, "stationCallsignEntry")
	result.addEntry(builder, "stationOperatorEntry")
	result.addEntry(builder, "stationLocatorEntry")
	result.addEntry(builder, "contestNameEntry")
	result.addCheckButton(builder, "contestEnterTheirNumberButton")
	result.addCheckButton(builder, "contestEnterTheirXchangeButton")

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
	widget.Connect("changed", v.onChanged)
}

func (v *settingsView) addCheckButton(builder *gtk.Builder, id string) {
	button := getUI(builder, id).(*gtk.CheckButton)
	name, _ := button.GetName()
	v.fields[name] = button

	widget := &button.Widget
	widget.Connect("toggled", v.onChanged)
}

func (v *settingsView) onChanged(w interface{}) bool {
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
	case "contestEnterTheirNumber":
		v.controller.EnterContestEnterTheirNumber(value.(bool))
	default:
		log.Printf("enter unknown field %s: %v", field, value)
	}

	return false
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

func (v *settingsView) SetStationCallsign(value string) {
	v.setEntryField("stationCallsign", value)
}

func (v *settingsView) SetContestEnterTheirNumber(value bool) {
	v.setCheckButtonField("contestEnterTheirNumber", value)
}
