package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"
)

// KeyerController controls the keyer.
type KeyerController interface {
	Send(int)
	Stop()
	EnterPattern(int, string)
	EnterSpeed(int)
}

type keyerView struct {
	controller KeyerController

	buttons    []*gtk.Button
	entries    []*gtk.Entry
	stopButton *gtk.Button
	speedEntry *gtk.SpinButton
}

func setupKeyerView(builder *gtk.Builder) *keyerView {
	result := new(keyerView)

	result.buttons = make([]*gtk.Button, 4)
	result.entries = make([]*gtk.Entry, 4)
	for i := 0; i < len(result.buttons); i++ {
		result.buttons[i] = getUI(builder, fmt.Sprintf("f%dButton", i+1)).(*gtk.Button)
		result.entries[i] = getUI(builder, fmt.Sprintf("f%dEntry", i+1)).(*gtk.Entry)
		result.buttons[i].Connect("clicked", result.onButton(i))
		result.entries[i].Connect("changed", result.onEntryChanged(i))
	}
	result.stopButton = getUI(builder, "stopButton").(*gtk.Button)
	result.speedEntry = getUI(builder, "speedEntry").(*gtk.SpinButton)

	result.stopButton.Connect("clicked", result.onStop)
	result.speedEntry.Connect("value-changed", result.onSpeedChanged)

	return result
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

func (k *keyerView) onEntryChanged(index int) func(entry *gtk.Entry) bool {
	return func(entry *gtk.Entry) bool {
		if k.controller == nil {
			log.Println("onEntryChanged: no keyer controller")
			return false
		}
		text, err := entry.GetText()
		if err != nil {
			log.Println(err)
			return false
		}
		k.controller.EnterPattern(index, text)
		return false
	}
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

func (k *keyerView) Pattern(index int) string {
	text, _ := k.entries[index].GetText()
	return text
}

func (k *keyerView) SetPattern(index int, text string) {
	k.entries[index].SetText(text)
}

func (k *keyerView) Speed() int {
	return int(k.speedEntry.GetValue())
}

func (k *keyerView) SetSpeed(speed int) {
	k.speedEntry.SetValue(float64(speed))
}
