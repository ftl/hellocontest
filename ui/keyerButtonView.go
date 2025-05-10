package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// KeyerController controls the keyer.
type KeyerController interface {
	Send(int)
	Stop()
	EnterSpeed(int)
	Save()
	OpenKeyerSettings()
}

type keyerView struct {
	controller KeyerController

	buttons                 []*gtk.Button
	stopButton              *gtk.Button
	openKeyerSettingsButton *gtk.Button
	speedEntry              *gtk.SpinButton

	cqLabelText  string
	parrotActive bool

	ignoreChangedEvent bool
}

func setupKeyerView(builder *gtk.Builder) *keyerView {
	result := new(keyerView)

	result.buttons = make([]*gtk.Button, 4)
	for i := 0; i < len(result.buttons); i++ {
		result.buttons[i] = getUI(builder, fmt.Sprintf("f%dButton", i+1)).(*gtk.Button)
		result.buttons[i].Connect("clicked", result.onButton(i))
	}

	result.stopButton = getUI(builder, "stopButton").(*gtk.Button)
	result.stopButton.Connect("clicked", result.onStop)

	result.openKeyerSettingsButton = getUI(builder, "openKeyerSettingsButton").(*gtk.Button)
	result.openKeyerSettingsButton.Connect("clicked", result.onOpenKeyerSettings)

	result.speedEntry = getUI(builder, "speedEntry").(*gtk.SpinButton)
	result.speedEntry.Connect("value-changed", result.onSpeedChanged)
	result.speedEntry.Connect("focus_out_event", result.onEntryFocusOut)

	return result
}

func (v *keyerView) doIgnoreChanges(f func()) {
	if v == nil {
		return
	}

	v.ignoreChangedEvent = true
	defer func() {
		v.ignoreChangedEvent = false
	}()
	f()
}

func (v *keyerView) onEntryFocusOut(widget interface{}, _ *gdk.Event) bool {
	v.controller.Save()
	return false
}

func (k *keyerView) onButton(index int) func(button *gtk.Button) bool {
	return func(button *gtk.Button) bool {
		if k.controller == nil {
			log.Println("onButton: no keyer controller")
			return false
		}
		k.controller.Send(index)
		return true
	}
}

func (k *keyerView) onOpenKeyerSettings(button *gtk.Button) bool {
	if k.controller == nil {
		log.Println("onOpenKeyerSettings: no keyer controller")
		return false
	}
	k.controller.OpenKeyerSettings()
	return true
}

func (k *keyerView) onStop(button *gtk.Button) bool {
	if k.controller == nil {
		log.Println("onStop: no keyer controller")
		return false
	}
	k.controller.Stop()
	return true
}

func (k *keyerView) onSpeedChanged(button *gtk.SpinButton) bool {
	if k.ignoreChangedEvent {
		return false
	}
	if k.controller == nil {
		log.Println("onSpeedChanged: no keyer controller")
		return false
	}

	k.controller.EnterSpeed(int(button.GetValue()))
	return true
}

func (k *keyerView) SetKeyerController(controller KeyerController) {
	k.controller = controller
}

func (k *keyerView) SetParrotActive(active bool) {
	k.parrotActive = active
	k.updateCQButtonLabel()
}

func (k *keyerView) updateCQButtonLabel() {
	k.buttons[0].SetLabel(k.buildLabel(0, k.cqLabelText))
}

func (k *keyerView) SetLabel(index int, text string) {
	if index == 0 {
		k.cqLabelText = text
	}

	label := k.buildLabel(index, text)
	k.buttons[index].SetLabel(label)
}

func (k *keyerView) buildLabel(index int, text string) string {
	var decoration string
	if index == 0 && k.parrotActive {
		decoration = parrot
	} else {
		decoration = fmt.Sprintf("F%d", index+1)
	}

	if text == "" {
		return decoration
	}
	return fmt.Sprintf("%s: %s", decoration, text)
}

func (k *keyerView) SetPattern(index int, text string) {
	k.buttons[index].SetTooltipText(fmt.Sprintf("F%d: %s", index+1, text))
}

func (k *keyerView) Speed() int {
	return int(k.speedEntry.GetValue())
}

func (k *keyerView) SetSpeed(speed int) {
	k.speedEntry.SetValue(float64(speed))
}
