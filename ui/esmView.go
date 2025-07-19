package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type ESMController interface {
	SetESMEnabled(enabled bool)
}

type esmView struct {
	controller ESMController

	enableButton *gtk.CheckButton
	messageLabel *gtk.Label
}

func setupESMView(builder *gtk.Builder) *esmView {
	result := new(esmView)

	result.enableButton = getUI(builder, "esmCheckButton").(*gtk.CheckButton)
	result.messageLabel = getUI(builder, "esmMessageLabel").(*gtk.Label)

	result.enableButton.Connect("toggled", result.onEnableButtonToggled)

	return result
}

func (v *esmView) SetESMController(controller ESMController) {
	v.controller = controller
}

func (v *esmView) onEnableButtonToggled(button *gtk.CheckButton) bool {
	v.controller.SetESMEnabled(button.GetActive())
	return true
}

func (v *esmView) SetESMEnabled(enabled bool) {
	if v.enableButton.GetActive() == enabled {
		return
	}
	v.enableButton.SetActive(enabled)
}

func (v *esmView) SetMessage(message string) {
	v.messageLabel.SetText(message)
}
