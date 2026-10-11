package ui

import (
	"github.com/ftl/hellocontest/core"
	qtlib "github.com/mappu/miqt/qt6"
)

// vfoArea handles the visible UI related to core/entry and core/callinfo for a specific VFO. It takes care of visibility, layout and style of the related widgets.
type vfoArea struct {
	vfo          core.VFOID
	widgetPrefix string
	vfoName      string

	entry    *entryVFOWidgets
	callinfo *callinfoVFOWidgets

	root        *qtlib.QWidget
	entryLayout *qtlib.QGridLayout
}

func newVFOArea(vfo core.VFOID, widgetPrefix string, vfoName string, entry *entryVFOWidgets, callinfo *callinfoVFOWidgets) *vfoArea {
	result := &vfoArea{
		vfo:          vfo,
		widgetPrefix: widgetPrefix,
		vfoName:      vfoName,
		entry:        entry,
		callinfo:     callinfo,
	}

	result.root = qtlib.NewQWidget2()
	result.root.SetObjectName(*qtlib.NewQAnyStringView3(widgetPrefix + "VFOArea"))
	result.root.SetAttribute(qtlib.WA_StyledBackground)

	result.entryLayout = qtlib.NewQGridLayout(result.root)
	result.entryLayout.SetContentsMargins(0, 0, 0, 0)
	result.entryLayout.SetSpacing(5)

	return result
}

func (a *vfoArea) setEnabled(enabled bool) {
	a.root.SetVisible(enabled)
}

func (a *vfoArea) addWidgetsToLayout() {
	lastColumn := 3 + len(a.entry.theirExchangeFields)

	a.entryLayout.AddWidget3(a.entry.vfoContainer, 0, 0, 1, -1)
	a.entryLayout.AddWidget2(a.callinfo.callsignLabel.QWidget, 1, 0)
	for i := range a.callinfo.predictedExchangeLabels {
		a.entryLayout.AddWidget2(a.callinfo.predictedExchangeLabels[i].QWidget, 1, i+1)
	}
	a.entryLayout.AddWidget2(a.callinfo.valueLabel.QWidget, 1, lastColumn-1)
	a.entryLayout.AddWidget2(a.callinfo.qtcStatusLabel.QWidget, 1, lastColumn)
	a.entryLayout.AddWidget2(a.entry.callsign.QWidget, 2, 0)
	for i := range a.entry.theirExchangeFields {
		a.entryLayout.AddWidget2(a.entry.theirExchangeFields[i].QWidget, 2, i+1)
	}
	a.entryLayout.AddWidget2(a.entry.logButton.QWidget, 2, lastColumn-1)
	a.entryLayout.AddWidget2(a.entry.clearButton.QWidget, 2, lastColumn)
	a.entryLayout.AddWidget3(a.callinfo.supercheckLabel.QWidget, 3, 0, 1, -1)
	a.entryLayout.AddWidget3(a.callinfo.infoContainer, 4, 0, 1, -1)
	a.entryLayout.AddWidget3(a.entry.messageLabel.QWidget, 4, 0, 1, -1)
}

func (a *vfoArea) removeWidgetsFromLayout() {
	a.entryLayout.RemoveWidget(a.entry.vfoContainer)
	a.entryLayout.RemoveWidget(a.callinfo.callsignLabel.QWidget)
	for i := range a.callinfo.predictedExchangeLabels {
		a.entryLayout.RemoveWidget(a.callinfo.predictedExchangeLabels[i].QWidget)
	}
	a.entryLayout.RemoveWidget(a.callinfo.valueLabel.QWidget)
	a.entryLayout.RemoveWidget(a.callinfo.qtcStatusLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.callsign.QWidget)
	for i := range a.entry.theirExchangeFields {
		a.entryLayout.RemoveWidget(a.entry.theirExchangeFields[i].QWidget)
	}
	a.entryLayout.RemoveWidget(a.entry.logButton.QWidget)
	a.entryLayout.RemoveWidget(a.entry.clearButton.QWidget)
	a.entryLayout.RemoveWidget(a.callinfo.supercheckLabel.QWidget)
	a.entryLayout.RemoveWidget(a.callinfo.infoContainer)
	a.entryLayout.RemoveWidget(a.entry.messageLabel.QWidget)
}
