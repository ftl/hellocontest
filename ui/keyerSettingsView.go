package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type KeyerSettingsController interface {
	EnterLabel(core.Workmode, int, string)
	EnterMacro(core.Workmode, int, string)
	SelectPreset(string)
	EnterParrotIntervalSeconds(int)
	Save()
}

type keyerSettingsView struct {
	parent     *gtk.Dialog
	controller KeyerSettingsController

	spLabels              []*gtk.Entry
	spMacros              []*gtk.Entry
	runLabels             []*gtk.Entry
	runMacros             []*gtk.Entry
	labels                [][]*gtk.Entry
	macros                [][]*gtk.Entry
	presetCombo           *gtk.ComboBoxText
	messageLabel          *gtk.Label
	parrotIntervalSeconds *gtk.SpinButton
	close                 *gtk.Button

	ignoreChangedEvent bool
}

func setupKeyerSettingsView(builder *gtk.Builder, parent *gtk.Dialog, controller KeyerSettingsController) *keyerSettingsView {
	result := new(keyerSettingsView)
	result.parent = parent
	result.controller = controller

	const macroCount = 4
	result.spLabels = make([]*gtk.Entry, macroCount)
	result.spMacros = make([]*gtk.Entry, macroCount)
	result.runLabels = make([]*gtk.Entry, macroCount)
	result.runMacros = make([]*gtk.Entry, macroCount)
	for i := 0; i < macroCount; i++ {
		result.spLabels[i] = result.setupEntry(builder, "spF%dLabel", core.SearchPounce, i, result.onLabelChanged)
		result.spMacros[i] = result.setupEntry(builder, "spF%dMacro", core.SearchPounce, i, result.onMacroChanged)
		result.runLabels[i] = result.setupEntry(builder, "runF%dLabel", core.Run, i, result.onLabelChanged)
		result.runMacros[i] = result.setupEntry(builder, "runF%dMacro", core.Run, i, result.onMacroChanged)
	}

	result.labels = [][]*gtk.Entry{result.spLabels, result.runLabels}
	result.macros = [][]*gtk.Entry{result.spMacros, result.runMacros}

	result.presetCombo = getUI(builder, "keyerPresetCombo").(*gtk.ComboBoxText)
	result.presetCombo.Connect("changed", result.onPresetChanged)

	result.parrotIntervalSeconds = getUI(builder, "parrotIntervalSpin").(*gtk.SpinButton)
	result.parrotIntervalSeconds.Connect("value-changed", result.onParrotIntervalChanged)
	result.parrotIntervalSeconds.Connect("focus_out_event", result.onEntryFocusOut)

	result.messageLabel = getUI(builder, "keyerSettingsMessageLabel").(*gtk.Label)

	result.close = getUI(builder, "keyerSettingsCloseButton").(*gtk.Button)
	result.close.Connect("clicked", result.onClosePressed)
	result.close.SetCanDefault(true)

	return result
}

func (v *keyerSettingsView) setupEntry(builder *gtk.Builder, idPattern string, workmode core.Workmode, i int, listenerFactory func(core.Workmode, int) func(*gtk.Entry) bool) *gtk.Entry {
	result := getUI(builder, fmt.Sprintf(idPattern, i+1)).(*gtk.Entry)
	result.Connect("changed", listenerFactory(workmode, i))
	result.Connect("focus_out_event", v.onEntryFocusOut)
	return result
}

func (v *keyerSettingsView) doIgnoreChanges(f func()) {
	if v == nil {
		return
	}

	v.ignoreChangedEvent = true
	defer func() {
		v.ignoreChangedEvent = false
	}()
	f()
}

func (v *keyerSettingsView) onClosePressed(_ *gtk.Button) {
	v.controller.Save()
	v.parent.Close()
}

func (v *keyerSettingsView) onEntryFocusOut(_ any, _ *gdk.Event) bool {
	v.controller.Save()
	return false
}

func (v *keyerSettingsView) onLabelChanged(workmode core.Workmode, index int) func(entry *gtk.Entry) bool {
	return func(entry *gtk.Entry) bool {
		if v.ignoreChangedEvent {
			return false
		}
		if v.controller == nil {
			log.Println("onLabelChanged: no keyer controller")
			return false
		}
		text, err := entry.GetText()
		if err != nil {
			log.Println(err)
			return false
		}
		v.controller.EnterLabel(workmode, index, text)
		return false
	}
}

func (v *keyerSettingsView) onMacroChanged(workmode core.Workmode, index int) func(entry *gtk.Entry) bool {
	return func(entry *gtk.Entry) bool {
		if v.ignoreChangedEvent {
			return false
		}
		if v.controller == nil {
			log.Println("onMacroChanged: no keyer controller")
			return false
		}
		text, err := entry.GetText()
		if err != nil {
			log.Println(err)
			return false
		}
		v.controller.EnterMacro(workmode, index, text)
		return false
	}
}

func (v *keyerSettingsView) onPresetChanged(combo *gtk.ComboBoxText) bool {
	if v.ignoreChangedEvent {
		return false
	}
	if v.controller == nil {
		log.Println("onPresetChanged: no keyer controller")
		return false
	}

	v.controller.SelectPreset(combo.GetActiveText())
	return true
}

func (v *keyerSettingsView) onParrotIntervalChanged(spin *gtk.SpinButton) bool {
	if v.ignoreChangedEvent {
		return false
	}
	if v.controller == nil {
		log.Println("onParrotIntervalChanged: no keyer controller")
		return false
	}

	v.controller.EnterParrotIntervalSeconds(int(spin.GetValue()))
	return true
}

func (v *keyerSettingsView) SetKeyerController(controller KeyerSettingsController) {
	v.controller = controller
}

func (v *keyerSettingsView) SetLabel(workmode core.Workmode, index int, text string) {
	v.labels[int(workmode)-1][index].SetText(text)
}

func (v *keyerSettingsView) SetMacro(workmode core.Workmode, index int, text string) {
	v.macros[int(workmode)-1][index].SetText(text)
}

func (v *keyerSettingsView) SetPresetNames(names []string) {
	v.doIgnoreChanges(func() {
		v.presetCombo.RemoveAll()
		v.presetCombo.Append("", "")
		for _, name := range names {
			v.presetCombo.Append(name, name)
		}
	})
}

func (v *keyerSettingsView) SetPreset(name string) {
	v.doIgnoreChanges(func() {
		v.presetCombo.SetActiveID(name)
	})
}

func (v *keyerSettingsView) SetParrotIntervalSeconds(interval int) {
	v.doIgnoreChanges(func() {
		v.parrotIntervalSeconds.SetValue(float64(interval))
	})
}

func (v *keyerSettingsView) ShowMessage(args ...interface{}) {
	v.messageLabel.SetText(fmt.Sprint(args...))
}

func (v *keyerSettingsView) ClearMessage() {
	v.messageLabel.SetText("")
}
