package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/settings"
)

var _ settings.View = (*settingsDialog)(nil)

type settingsDialog struct {
	parent     *qtlib.QMainWindow
	controller SettingsController

	dialog  *qtlib.QDialog
	view    *settingsView
	created bool
}

func newSettingsDialog(parent *qtlib.QMainWindow, controller SettingsController) *settingsDialog {
	return &settingsDialog{parent: parent, controller: controller}
}

func (d *settingsDialog) Show() {
	if !d.created {
		d.create()
		d.created = true
	}
	d.dialog.Show()
	d.dialog.Raise()
	d.dialog.ActivateWindow()
}

func (d *settingsDialog) Ready() bool {
	return d.created && d.view != nil
}

func (d *settingsDialog) create() {
	d.dialog = qtlib.NewQDialog(d.parent.QWidget)
	d.dialog.SetObjectName(*qtlib.NewQAnyStringView3("settingsDialog"))
	d.dialog.SetWindowTitle("Contest Settings")
	d.dialog.SetModal(false)
	d.dialog.SetWindowFlags(
		qtlib.Window |
			qtlib.CustomizeWindowHint |
			qtlib.WindowTitleHint |
			qtlib.WindowCloseButtonHint,
	)

	d.view = newSettingsView(d.dialog, d.controller)

	layout := qtlib.NewQVBoxLayout(d.dialog.QWidget)
	layout.AddWidget(d.view.root)

	d.dialog.AdjustSize()

	d.dialog.OnFinished(func(result int) {
		d.controller.Save()
		d.dialog.Hide()
	})
}

// settings.View delegation

func (d *settingsDialog) ShowMessage(msg string) {
	if d.view == nil {
		return
	}
	d.view.ShowMessage(msg)
}

func (d *settingsDialog) HideMessage() {
	if d.view == nil {
		return
	}
	d.view.HideMessage()
}

func (d *settingsDialog) SetStationCallsign(v string) {
	if d.view == nil {
		return
	}
	d.view.SetStationCallsign(v)
}

func (d *settingsDialog) SetStationOperator(v string) {
	if d.view == nil {
		return
	}
	d.view.SetStationOperator(v)
}

func (d *settingsDialog) SetStationLocator(v string) {
	if d.view == nil {
		return
	}
	d.view.SetStationLocator(v)
}

func (d *settingsDialog) SetContestIdentifiers(ids []string, texts []string) {
	if d.view == nil {
		return
	}
	d.view.SetContestIdentifiers(ids, texts)
}

func (d *settingsDialog) SelectContestIdentifier(v string) {
	if d.view == nil {
		return
	}
	d.view.SelectContestIdentifier(v)
}

func (d *settingsDialog) SetContestExchangeFields(fields []core.ExchangeField) {
	if d.view == nil {
		return
	}
	d.view.SetContestExchangeFields(fields)
}

func (d *settingsDialog) SetContestExchangeValue(index int, value string) {
	if d.view == nil {
		return
	}
	d.view.SetContestExchangeValue(index, value)
}

func (d *settingsDialog) SetContestGenerateSerialExchange(active bool, sensitive bool) {
	if d.view == nil {
		return
	}
	d.view.SetContestGenerateSerialExchange(active, sensitive)
}

func (d *settingsDialog) SetContestGenerateReport(active bool, sensitive bool) {
	if d.view == nil {
		return
	}
	d.view.SetContestGenerateReport(active, sensitive)
}

func (d *settingsDialog) SetContestEnableQTCs(v bool) {
	if d.view == nil {
		return
	}
	d.view.SetContestEnableQTCs(v)
}

func (d *settingsDialog) SetContestName(v string) {
	if d.view == nil {
		return
	}
	d.view.SetContestName(v)
}

func (d *settingsDialog) SetContestStartTime(v string) {
	if d.view == nil {
		return
	}
	d.view.SetContestStartTime(v)
}

func (d *settingsDialog) SetOperationModeSprint(v bool) {
	if d.view == nil {
		return
	}
	d.view.SetOperationModeSprint(v)
}

func (d *settingsDialog) SetSwitchTXVFOOnFocus(v bool) {
	if d.view == nil {
		return
	}
	d.view.SetSwitchTXVFOOnFocus(v)
}

func (d *settingsDialog) SetContestCallHistoryFile(v string) {
	if d.view == nil {
		return
	}
	d.view.SetContestCallHistoryFile(v)
}

func (d *settingsDialog) SetContestCallHistoryFieldName(i int, value string) {
	if d.view == nil {
		return
	}
	d.view.SetContestCallHistoryFieldName(i, value)
}

func (d *settingsDialog) SetContestAvailableCallHistoryFieldNames(names []string) {
	if d.view == nil {
		return
	}
	d.view.SetContestAvailableCallHistoryFieldNames(names)
}

func (d *settingsDialog) SetQSOsGoal(v string) {
	if d.view == nil {
		return
	}
	d.view.SetQSOsGoal(v)
}

func (d *settingsDialog) SetPointsGoal(v string) {
	if d.view == nil {
		return
	}
	d.view.SetPointsGoal(v)
}

func (d *settingsDialog) SetMultisGoal(v string) {
	if d.view == nil {
		return
	}
	d.view.SetMultisGoal(v)
}
