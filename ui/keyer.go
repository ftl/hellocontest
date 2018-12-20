package ui

import (
	"fmt"
	"log"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

type keyer struct {
	controller core.KeyerController

	buttons []*gtk.Button
	entries []*gtk.Entry
}

func setupKeyer(builder *gtk.Builder, parent *mainWindow) *keyer {
	result := new(keyer)

	result.buttons = make([]*gtk.Button, 4)
	result.entries = make([]*gtk.Entry, 4)
	for i := 0; i < len(result.buttons); i++ {
		result.buttons[i] = getUI(builder, fmt.Sprintf("f%dButton", i+1)).(*gtk.Button)
		result.entries[i] = getUI(builder, fmt.Sprintf("f%dEntry", i+1)).(*gtk.Entry)
		result.buttons[i].Connect("clicked", result.onButton(i))
		result.entries[i].Connect("changed", result.onEntryChanged(i))
	}

	return result
}

func (k *keyer) onButton(index int) func(button *gtk.Button) bool {
	return func(button *gtk.Button) bool {
		if k.controller == nil {
			log.Printf("no keyer controller")
			return false
		}
		k.controller.Send(index)
		return true
	}
}

func (k *keyer) onEntryChanged(index int) func(entry *gtk.Entry) bool {
	return func(entry *gtk.Entry) bool {
		if k.controller == nil {
			log.Printf("no keyer controller")
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

func (k *keyer) SetKeyerController(controller core.KeyerController) {
	k.controller = controller
}

func (k *keyer) GetPattern(index int) string {
	text, _ := k.entries[index].GetText()
	return text
}

func (k *keyer) SetPattern(index int, text string) {
	k.entries[index].SetText(text)
}
