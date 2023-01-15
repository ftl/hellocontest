package ui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

// WorkmodeController controls the workmode handling.
type WorkmodeController interface {
	SetWorkmode(core.Workmode)
}

type workmodeView struct {
	controller WorkmodeController

	searchPounceModeButton *gtk.RadioButton
	runModeButton          *gtk.RadioButton
	operationModeLabel     *gtk.Label
}

func setupWorkmodeView(builder *gtk.Builder) *workmodeView {
	result := new(workmodeView)

	result.searchPounceModeButton = getUI(builder, "searchPounceModeButton").(*gtk.RadioButton)
	result.runModeButton = getUI(builder, "runModeButton").(*gtk.RadioButton)
	result.operationModeLabel = getUI(builder, "operationModeLabel").(*gtk.Label)

	result.searchPounceModeButton.Connect("toggled", result.onSearchPounceModeButtonToggled)
	result.runModeButton.Connect("toggled", result.onRunModeButtonToggled)

	return result
}

func (v *workmodeView) onSearchPounceModeButtonToggled(button *gtk.RadioButton) bool {
	if button.GetActive() {
		v.controller.SetWorkmode(core.SearchPounce)
	}
	return true
}

func (v *workmodeView) onRunModeButtonToggled(button *gtk.RadioButton) bool {
	if button.GetActive() {
		v.controller.SetWorkmode(core.Run)
	}
	return true
}

func (v *workmodeView) SetWorkmodeController(controller WorkmodeController) {
	v.controller = controller
}

func (v *workmodeView) SetWorkmode(workmode core.Workmode) {
	var activeButton *gtk.RadioButton
	switch workmode {
	case core.SearchPounce:
		activeButton = v.searchPounceModeButton
	case core.Run:
		activeButton = v.runModeButton
	default:
		activeButton = nil
	}

	if activeButton != nil && !activeButton.GetActive() {
		name, _ := activeButton.GetLabel()
		log.Printf("UI: set %s active", name)
		activeButton.SetActive(true)
	}
}

func (v *workmodeView) SetOperationModeHint(hint string) {
	v.operationModeLabel.SetText(hint)
}
