package fyneui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const parrot = "🦜"

type KeyerController interface {
	Send(int)
	Stop()
	EnterSpeed(int)
	OpenKeyerSettings()
}

func setupKeyerControl() *keyerControl {
	result := &keyerControl{}

	result.stop = widget.NewButton("ESC: Stop TX", result.onStop)
	result.buttons = []*widget.Button{
		widget.NewButton("F1", result.onMacro(0)),
		widget.NewButton("F2", result.onMacro(1)),
		widget.NewButton("F3", result.onMacro(2)),
		widget.NewButton("F4", result.onMacro(3)),
	}
	result.macros = widget.NewButton("Macros...", result.onMacros)

	result.container = container.NewGridWithColumns(5,
		result.stop, result.buttons[0], result.buttons[1], result.buttons[2], result.buttons[3],
		result.macros,
	)

	return result
}

type keyerControl struct {
	container  *fyne.Container
	controller KeyerController

	cqLabelText  string
	parrotActive bool

	stop    *widget.Button
	buttons []*widget.Button
	macros  *widget.Button
	// TODO: Speed Control
}

func (c *keyerControl) SetKeyerController(controller KeyerController) {
	c.controller = controller
}

func (c *keyerControl) SetParrotActive(active bool) {
	c.parrotActive = active
	c.updateCQButtonLabel()
}

func (c *keyerControl) ShowMessage(...any) {
	// TODO: show this message somewhere
}

func (c *keyerControl) updateCQButtonLabel() {
	c.buttons[0].SetText(c.buildLabel(0, c.cqLabelText))
}

func (c *keyerControl) SetLabel(index int, text string) {
	if index == 0 {
		c.cqLabelText = text
	}

	label := c.buildLabel(index, text)
	c.buttons[index].SetText(label)
}

func (c *keyerControl) buildLabel(index int, text string) string {
	var decoration string
	if index == 0 && c.parrotActive {
		decoration = parrot
	} else {
		decoration = fmt.Sprintf("F%d", index+1)
	}

	if text == "" {
		return decoration
	}
	return fmt.Sprintf("%s: %s", decoration, text)
}

func (c *keyerControl) SetPattern(index int, text string) {
	// TODO: add tooltip support from https://github.com/dweymouth/fyne-tooltip
	// c.buttons[index].SetTooltip(fmt.Sprintf("F%d: %s", index+1, text))
}

func (c *keyerControl) onStop() {
	if c.controller == nil {
		log.Println("onStop: no keyer controller")
		return
	}
	c.controller.Stop()
}

func (c *keyerControl) onMacro(index int) func() {
	return func() {
		if c.controller == nil {
			log.Printf("onMacro(%d): no keyer controller", index)
			return
		}
		c.controller.Send(index)
	}
}

func (c *keyerControl) onMacros() {
	if c.controller == nil {
		log.Println("onMacros: no keyer controller")
		return
	}
	c.controller.OpenKeyerSettings()
}

func (c *keyerControl) onChangeSpeed(delta int) func() {
	// TODO: speed control
	return func() {}
}

func (c *keyerControl) onSetSpeed(speed int) func() {
	// TODO: speed control
	return func() {}
}

func (c *keyerControl) SetSpeed(speed int) {
	// TODO: speed control
}
