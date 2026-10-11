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
	rootLayout.SetContentsMargins(0, 0, 0, 0)

	// TODO: myDataContainer, myDataLayout

	result.vfoArea[core.VFO1] = result.newVFOArea(core.VFO1, "vfo1", "VFO 1")
	result.vfoArea[core.VFO2] = result.newVFOArea(core.VFO2, "vfo2", "VFO 2")

	return result
}

func (a *entryArea) newVFOArea(vfo core.VFOID, widgetPrefix string, vfoName string) *vfoArea {
	return newVFOArea(vfo, widgetPrefix, vfoName, &a.entry.vfo[vfo], &a.callinfo.vfo[vfo])
}

func (a *entryArea) addWidgetsToLayout() {
	// TODO: my data
	a.vfoArea[core.VFO1].addWidgetsToLayout()
	a.vfoArea[core.VFO2].addWidgetsToLayout()
}

func (a *entryArea) removeWidgetsFromlayout() {
	// TODO: my data
	a.vfoArea[core.VFO1].removeWidgetsFromLayout()
	a.vfoArea[core.VFO2].removeWidgetsFromLayout()
}
