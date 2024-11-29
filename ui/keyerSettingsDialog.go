//go:build !fyne

package ui

import (
	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

type keyerSettingsDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller KeyerSettingsController
	*keyerSettingsView
}

func setupKeyerSettingsDialog(parent gtk.IWidget, controller KeyerSettingsController) *keyerSettingsDialog {
	result := &keyerSettingsDialog{
		parent:     parent,
		controller: controller,
	}
	return result
}

func (d *keyerSettingsDialog) onDestroy() {
	d.dialog = nil
	d.keyerSettingsView = nil
}

func (d *keyerSettingsDialog) Show() {
	if d.dialog == nil {
		builder := setupBuilder()
		d.dialog = getUI(builder, "keyerSettingsDialog").(*gtk.Dialog)
		d.dialog.SetPosition(gtk.WIN_POS_CENTER)
		d.dialog.Connect("destroy", d.onDestroy)
		d.keyerSettingsView = setupKeyerSettingsView(builder, d.dialog, d.controller)
	}
	d.dialog.ShowAll()
	d.dialog.Present()
}

func (d *keyerSettingsDialog) ShowMessage(message ...any) {
	if d.keyerSettingsView == nil {
		return
	}
	d.keyerSettingsView.ShowMessage(message...)
}

func (d *keyerSettingsDialog) ClearMessage() {
	if d.keyerSettingsView == nil {
		return
	}
	d.keyerSettingsView.ClearMessage()
}

func (d *keyerSettingsDialog) SetLabel(workmode core.Workmode, index int, text string) {
	if d.keyerSettingsView == nil {
		return
	}
	d.keyerSettingsView.SetLabel(workmode, index, text)
}

func (d *keyerSettingsDialog) SetMacro(workmode core.Workmode, index int, text string) {
	if d.keyerSettingsView == nil {
		return
	}
	d.keyerSettingsView.SetMacro(workmode, index, text)
}

func (d *keyerSettingsDialog) SetPresetNames(names []string) {
	if d.keyerSettingsView == nil {
		return
	}
	d.keyerSettingsView.SetPresetNames(names)
}

func (d *keyerSettingsDialog) SetPreset(name string) {
	if d.keyerSettingsView == nil {
		return
	}
	d.keyerSettingsView.SetPreset(name)
}
