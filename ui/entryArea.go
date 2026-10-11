package ui

import (
	"github.com/ftl/hellocontest/core"
	qtlib "github.com/mappu/miqt/qt6"
)

// entryArea handles the visible UI related to core/entry and core/callinfo. It takes care of visibility, layout and style of the related widgets.
type entryArea struct {
	entry    *entryView
	callinfo *callinfoView

	root *qtlib.QWidget

	myDataContainer *qtlib.QWidget
	myDataLayout    *qtlib.QHBoxLayout
	myDataLabel     *qtlib.QLabel

	vfoArea map[core.VFOID]*vfoArea

	vfo2Enabled bool
	activeVFO   core.VFOID
	txVFO       core.VFOID
}

func newEntryArea(entry *entryView, callinfo *callinfoView) *entryArea {
	result := &entryArea{
		entry:    entry,
		callinfo: callinfo,
		vfoArea:  make(map[core.VFOID]*vfoArea, 2),
	}

	entryFrame := qtlib.NewQFrame2()
	entryFrame.SetFrameShape(qtlib.QFrame__StyledPanel)
	entryFrame.SetFrameShadow(qtlib.QFrame__Plain)

	result.root = entryFrame.QWidget
	result.root.SetObjectName(*qtlib.NewQAnyStringView3("entryArea"))
	result.root.SetAttribute(qtlib.WA_StyledBackground)
	result.root.SetStyleSheet(EntryAreaStyle)

	rootLayout := qtlib.NewQVBoxLayout(result.root)
	rootLayout.SetContentsMargins(5, 5, 5, 5)
	rootLayout.SetSpacing(5)

	result.myDataContainer = qtlib.NewQWidget2()
	result.myDataContainer.SetObjectName(*qtlib.NewQAnyStringView3("myDataContainer"))
	result.myDataLayout = qtlib.NewQHBoxLayout(result.myDataContainer)
	result.myDataLayout.SetContentsMargins(0, 0, 0, 0)
	result.myDataLayout.SetSpacing(5)

	result.myDataLabel = qtlib.NewQLabel3("My Data:")
	result.myDataLabel.SetObjectName(*qtlib.NewQAnyStringView3("myDataLabel"))
	result.myDataLayout.AddWidget(result.myDataLabel.QWidget)
	result.myDataLayout.AddStretch()

	result.vfoArea[core.VFO1] = result.newVFOArea(core.VFO1, "vfo1")
	result.vfoArea[core.VFO2] = result.newVFOArea(core.VFO2, "vfo2")

	rootLayout.AddWidget(result.myDataContainer)
	rootLayout.AddWidget(result.vfoArea[core.VFO1].root)
	rootLayout.AddWidget(result.vfoArea[core.VFO2].root)

	result.vfoArea[core.VFO2].setEnabled(false)
	result.applyActiveVFOStyle()
	result.applyTXIndicatorVisibility()

	entry.area = result
	callinfo.area = result

	return result
}

func (a *entryArea) newVFOArea(vfo core.VFOID, widgetPrefix string) *vfoArea {
	return newVFOArea(vfo, widgetPrefix, &a.entry.vfo[vfo], &a.callinfo.vfo[vfo])
}

func (a *entryArea) SetExchangeFields(myExchangeFields, theirExchangeFields []core.ExchangeField) {
	a.removeWidgetsFromLayout()

	a.entry.SetExchangeFields(myExchangeFields, theirExchangeFields)
	a.callinfo.SetPredictedExchangeFields(theirExchangeFields)

	a.setupMyExchangeFields(myExchangeFields)
	for _, vfoArea := range a.vfoArea {
		vfoArea.setupExchangeFields(theirExchangeFields)
	}

	a.addWidgetsToLayout()
}

func (a *entryArea) setupMyExchangeFields(fields []core.ExchangeField) {
	for i, field := range fields {
		if i >= len(a.entry.myExchangeFields) {
			break
		}
		editField := a.entry.myExchangeFields[i]
		editField.SetObjectName(*qtlib.NewQAnyStringView3(string(field.Field)))
		setFixedTextWidth(editField.QWidget, exchangeFieldWidthSample, exchangeFieldPadding)
	}
}

func (a *entryArea) setVFO2Enabled(enabled bool) {
	if a.vfo2Enabled == enabled {
		return
	}
	a.vfo2Enabled = enabled

	a.vfoArea[core.VFO2].setEnabled(enabled)
	a.applyActiveVFOStyle()
	a.applyTXIndicatorVisibility()
}

func (a *entryArea) setActiveVFO(vfo core.VFOID) {
	a.activeVFO = vfo
	a.applyActiveVFOStyle()
}

func (a *entryArea) applyActiveVFOStyle() {
	for vfo, vfoArea := range a.vfoArea {
		vfoArea.setActive(a.vfo2Enabled && vfo == a.activeVFO)
	}
}

func (a *entryArea) setTXVFO(vfo core.VFOID) {
	a.txVFO = vfo
	a.applyTXIndicatorVisibility()
}

func (a *entryArea) applyTXIndicatorVisibility() {
	for vfo, vfoArea := range a.vfoArea {
		vfoArea.setTXVFO(a.vfo2Enabled && vfo == a.txVFO)
	}
}

func (a *entryArea) setTXState(vfo core.VFOID, ptt bool, parrotActive bool) {
	a.vfoArea[vfo].setTXState(ptt, parrotActive)
}

func (a *entryArea) setIncrementalTuningVisible(vfo core.VFOID, kind core.IncrementalTuningKind, visible bool) {
	a.vfoArea[vfo].setIncrementalTuningVisible(kind, visible)
}

func (a *entryArea) setDuplicateMarker(vfo core.VFOID, duplicate bool) {
	a.vfoArea[vfo].setDuplicateMarker(duplicate)
}

func (a *entryArea) setEditingMarker(vfo core.VFOID, editing bool) {
	a.vfoArea[vfo].setEditingMarker(editing)
}

func (a *entryArea) setMessageVisible(vfo core.VFOID, visible bool) {
	a.vfoArea[vfo].setMessageVisible(visible)
}

func (a *entryArea) setSerialClaim(vfo core.VFOID, serial core.QSONumber, committed bool) {
	a.callinfo.SetSerialClaim(vfo, serial, committed)
}

func (a *entryArea) addWidgetsToLayout() {
	for i := range a.entry.myExchangeFields {
		a.myDataLayout.InsertWidget(i+1, a.entry.myExchangeFields[i].QWidget)
	}
	a.vfoArea[core.VFO1].addWidgetsToLayout()
	a.vfoArea[core.VFO2].addWidgetsToLayout()
}

func (a *entryArea) removeWidgetsFromLayout() {
	for i := range a.entry.myExchangeFields {
		a.myDataLayout.RemoveWidget(a.entry.myExchangeFields[i].QWidget)
	}
	a.vfoArea[core.VFO1].removeWidgetsFromLayout()
	a.vfoArea[core.VFO2].removeWidgetsFromLayout()
}
