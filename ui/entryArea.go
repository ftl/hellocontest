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
	result.root.SetStyleSheet(CentralAreaStyle) // TODO: do we need a separate style?

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

	result.vfoArea[core.VFO1] = result.newVFOArea(core.VFO1, "vfo1", "VFO 1")
	result.vfoArea[core.VFO2] = result.newVFOArea(core.VFO2, "vfo2", "VFO 2")

	rootLayout.AddWidget(result.myDataContainer)
	rootLayout.AddWidget(result.vfoArea[core.VFO1].root)
	rootLayout.AddWidget(result.vfoArea[core.VFO2].root)

	result.vfoArea[core.VFO2].setEnabled(false)

	entry.setRootWidget(result.root)
	entry.onVFO2Enabled = result.setVFO2Enabled
	entry.onSerialClaim = callinfo.SetSerialClaim

	return result
}

func (a *entryArea) newVFOArea(vfo core.VFOID, widgetPrefix string, vfoName string) *vfoArea {
	return newVFOArea(vfo, widgetPrefix, vfoName, &a.entry.vfo[vfo], &a.callinfo.vfo[vfo])
}

func (a *entryArea) SetExchangeFields(myExchangeFields, theirExchangeFields []core.ExchangeField, generateSerialExchange bool) {
	a.removeWidgetsFromLayout()
	a.entry.SetExchangeFields(myExchangeFields, theirExchangeFields, generateSerialExchange)
	a.callinfo.SetPredictedExchangeFields(theirExchangeFields)
	a.addWidgetsToLayout()
}

func (a *entryArea) setVFO2Enabled(enabled bool) {
	a.vfoArea[core.VFO2].setEnabled(enabled)
	a.entry.setVFO2Enabled(enabled)
	a.callinfo.SetVFOEnabled(core.VFO2, enabled)
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
