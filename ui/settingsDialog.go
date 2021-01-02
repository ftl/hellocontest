package ui

import "github.com/gotk3/gotk3/gtk"

type settingsDialog struct {
	dialog *gtk.Dialog

	controller SettingsController
	*settingsView
}

func setupSettingsDialog(controller SettingsController) *settingsDialog {
	result := &settingsDialog{
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
		d.dialog.SetTitle("Settings")
		d.dialog.Connect("destroy", d.onDestroy)
		d.settingsView = setupSettingsView(builder, d.dialog, d.controller)
	}
	d.dialog.ShowAll()
	d.dialog.Present()
}
