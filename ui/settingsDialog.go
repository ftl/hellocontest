//go:build !fyne

package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type settingsDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller SettingsController
	*settingsView
}

func setupSettingsDialog(parent gtk.IWidget, controller SettingsController) *settingsDialog {
	result := &settingsDialog{
		parent:     parent,
		controller: controller,
	}
	return result
}

func (d *settingsDialog) onDestroy() {
	d.dialog = nil
	d.settingsView = nil
}

func (d *settingsDialog) Show() {
	if d.dialog == nil {
		builder := setupBuilder()
		d.dialog = getUI(builder, "settingsDialog").(*gtk.Dialog)
		d.dialog.SetPosition(gtk.WIN_POS_CENTER)
		d.dialog.Connect("destroy", d.onDestroy)
		d.settingsView = setupSettingsView(builder, d.dialog, d.controller)
	}
	d.dialog.ShowAll()
	d.dialog.Present()
}
