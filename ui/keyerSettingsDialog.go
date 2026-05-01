package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/keyer"
)

var _ keyer.SettingsView = (*keyerSettingsDialog)(nil)

type keyerSettingsDialog struct {
	parent     *qtlib.QMainWindow
	controller KeyerSettingsController

	dialog  *qtlib.QDialog
	view    *keyerSettingsView
	created bool
}

func newKeyerSettingsDialog(parent *qtlib.QMainWindow, controller KeyerSettingsController) *keyerSettingsDialog {
	return &keyerSettingsDialog{parent: parent, controller: controller}
}

func (d *keyerSettingsDialog) Show() {
	if !d.created {
		d.create()
		d.created = true
	}
	d.dialog.Show()
	d.dialog.Raise()
	d.dialog.ActivateWindow()
}

func (d *keyerSettingsDialog) create() {
	d.dialog = qtlib.NewQDialog(d.parent.QWidget)
	d.dialog.SetObjectName(*qtlib.NewQAnyStringView3("keyerSettingsDialog"))
	d.dialog.SetWindowTitle("Keyer Macros")
	d.dialog.SetModal(false)
	d.dialog.SetMinimumWidth(600)
	d.dialog.SetWindowFlags(
		qtlib.Window |
			qtlib.CustomizeWindowHint |
			qtlib.WindowTitleHint |
			qtlib.WindowCloseButtonHint,
	)

	d.view = newKeyerSettingsView(d.dialog, d.controller)

	layout := qtlib.NewQVBoxLayout(d.dialog.QWidget)
	layout.AddWidget(d.view.root)

	d.dialog.AdjustSize()

	d.dialog.OnFinished(func(result int) {
		d.controller.Save()
		d.dialog.Hide()
	})
}

// keyer.SettingsView delegation

func (d *keyerSettingsDialog) ShowMessage(args ...any) {
	if d.view == nil {
		return
	}
	d.view.ShowMessage(args...)
}

func (d *keyerSettingsDialog) ClearMessage() {
	if d.view == nil {
		return
	}
	d.view.ClearMessage()
}

func (d *keyerSettingsDialog) SetLabel(workmode core.Workmode, index int, text string) {
	if d.view == nil {
		return
	}
	d.view.SetLabel(workmode, index, text)
}

func (d *keyerSettingsDialog) SetMacro(workmode core.Workmode, index int, text string) {
	if d.view == nil {
		return
	}
	d.view.SetMacro(workmode, index, text)
}

func (d *keyerSettingsDialog) SetPresetNames(names []string) {
	if d.view == nil {
		return
	}
	d.view.SetPresetNames(names)
}

func (d *keyerSettingsDialog) SetPreset(name string) {
	if d.view == nil {
		return
	}
	d.view.SetPreset(name)
}

func (d *keyerSettingsDialog) SetParrotIntervalSeconds(interval int) {
	if d.view == nil {
		return
	}
	d.view.SetParrotIntervalSeconds(interval)
}
