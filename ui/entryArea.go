package ui

import (
	"github.com/ftl/hellocontest/core/model"
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

	vfoArea map[model.VFOID]*vfoArea

	vfo2Enabled bool
	activeVFO   model.VFOID
	txVFO       model.VFOID
}

func newEntryArea(entry *entryView, callinfo *callinfoView) *entryArea {
	result := &entryArea{
		entry:    entry,
		callinfo: callinfo,
		vfoArea:  make(map[model.VFOID]*vfoArea, 2),
	}

	entryFrame := qtlib.NewQFrame2()
	entryFrame.SetFrameShape(qtlib.QFrame__StyledPanel)
	entryFrame.SetFrameShadow(qtlib.QFrame__Plain)

	result.root = entryFrame.QWidget
	setObjectName(result.root, "entryArea")
	result.root.SetAttribute(qtlib.WA_StyledBackground)
	result.root.SetStyleSheet(EntryAreaStyle)

	rootLayout := qtlib.NewQVBoxLayout(result.root)
	rootLayout.SetContentsMargins(5, 5, 5, 5)
	rootLayout.SetSpacing(5)

	result.myDataContainer = qtlib.NewQWidget2()
	setObjectName(result.myDataContainer, "myDataContainer")
	result.myDataLayout = qtlib.NewQHBoxLayout(result.myDataContainer)
	result.myDataLayout.SetContentsMargins(0, 0, 0, 0)
	result.myDataLayout.SetSpacing(5)

	result.myDataLabel = qtlib.NewQLabel3("My Data:")
	setObjectName(result.myDataLabel, "myDataLabel")
	result.myDataLayout.AddWidget(result.myDataLabel.QWidget)
	result.myDataLayout.AddStretch()

	result.vfoArea[model.VFO1] = result.newVFOArea(model.VFO1, "vfo1")
	result.vfoArea[model.VFO2] = result.newVFOArea(model.VFO2, "vfo2")

	rootLayout.AddWidget(result.myDataContainer)
	rootLayout.AddWidget(result.vfoArea[model.VFO1].root)
	rootLayout.AddWidget(result.vfoArea[model.VFO2].root)

	result.vfoArea[model.VFO2].setEnabled(false)
	result.applyActiveVFOStyle()
	result.applyTXIndicatorVisibility()

	entry.area = result
	callinfo.area = result

	return result
}

func (a *entryArea) newVFOArea(vfo model.VFOID, widgetPrefix string) *vfoArea {
	return newVFOArea(vfo, widgetPrefix, &a.entry.vfo[vfo], &a.callinfo.vfo[vfo])
}

func (a *entryArea) SetExchangeFields(myExchangeFields, theirExchangeFields []model.ExchangeField, showSerialExchangeIndicator bool) {
	a.removeWidgetsFromLayout()

	a.entry.SetExchangeFields(myExchangeFields, theirExchangeFields)
	a.callinfo.SetPredictedExchangeFields(theirExchangeFields, showSerialExchangeIndicator)

	a.setupMyExchangeFields(myExchangeFields)
	for _, vfoArea := range a.vfoArea {
		vfoArea.setupExchangeFields(theirExchangeFields)
	}

	a.addWidgetsToLayout()
}

func (a *entryArea) setupMyExchangeFields(fields []model.ExchangeField) {
	for i, field := range fields {
		if i >= len(a.entry.myExchangeFields) {
			break
		}
		editField := a.entry.myExchangeFields[i]
		setObjectName(editField, string(field.Field))
		setFixedTextWidth(editField.QWidget, exchangeFieldWidthSample, exchangeFieldPadding)
	}
}

func (a *entryArea) setVFO2Enabled(enabled bool) {
	if a.vfo2Enabled == enabled {
		return
	}
	a.vfo2Enabled = enabled

	a.vfoArea[model.VFO2].setEnabled(enabled)
	a.applyActiveVFOStyle()
	a.applyTXIndicatorVisibility()
}

func (a *entryArea) setActiveVFO(vfo model.VFOID) {
	a.activeVFO = vfo
	a.applyActiveVFOStyle()
}

func (a *entryArea) applyActiveVFOStyle() {
	for vfo, vfoArea := range a.vfoArea {
		vfoArea.setActive(a.vfo2Enabled && vfo == a.activeVFO)
	}
}

func (a *entryArea) setTXVFO(vfo model.VFOID) {
	a.txVFO = vfo
	a.applyTXIndicatorVisibility()
}

func (a *entryArea) applyTXIndicatorVisibility() {
	for vfo, vfoArea := range a.vfoArea {
		vfoArea.setTXVFO(a.vfo2Enabled && vfo == a.txVFO)
	}
}

func (a *entryArea) setTXState(vfo model.VFOID, ptt bool, parrotActive bool) {
	a.vfoArea[vfo].setTXState(ptt, parrotActive)
}

func (a *entryArea) setIncrementalTuningVisible(vfo model.VFOID, kind model.IncrementalTuningKind, visible bool) {
	a.vfoArea[vfo].setIncrementalTuningVisible(kind, visible)
}

func (a *entryArea) setDuplicateMarker(vfo model.VFOID, duplicate bool) {
	a.vfoArea[vfo].setDuplicateMarker(duplicate)
}

func (a *entryArea) setEditingMarker(vfo model.VFOID, editing bool) {
	a.vfoArea[vfo].setEditingMarker(editing)
}

func (a *entryArea) setMessageVisible(vfo model.VFOID, visible bool) {
	a.vfoArea[vfo].setMessageVisible(visible)
}

func (a *entryArea) setSerialClaim(vfo model.VFOID, serial model.QSONumber, committed bool) {
	a.callinfo.SetSerialClaim(vfo, serial, committed)
}

func (a *entryArea) addWidgetsToLayout() {
	for i := range a.entry.myExchangeFields {
		a.myDataLayout.InsertWidget(i+1, a.entry.myExchangeFields[i].QWidget)
	}
	a.vfoArea[model.VFO1].addWidgetsToLayout()
	a.vfoArea[model.VFO2].addWidgetsToLayout()
}

func (a *entryArea) removeWidgetsFromLayout() {
	for i := range a.entry.myExchangeFields {
		a.myDataLayout.RemoveWidget(a.entry.myExchangeFields[i].QWidget)
	}
	a.vfoArea[model.VFO1].removeWidgetsFromLayout()
	a.vfoArea[model.VFO2].removeWidgetsFromLayout()
}
