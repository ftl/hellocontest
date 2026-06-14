package ui

import (
	"github.com/ftl/hellocontest/core"
	qtlib "github.com/mappu/miqt/qt6"
)

type centralArea struct {
	root        *qtlib.QWidget
	entryRoot   *qtlib.QWidget
	entryLayout *qtlib.QGridLayout

	entry    *entryView
	callinfo *callinfoView
	esm      *esmView
	workmode *workmodeView
	keyer    *keyerView

	esmWorkmodeContainer *qtlib.QWidget
}

func newCentralArea(entry *entryView, callinfo *callinfoView, esm *esmView, workmode *workmodeView, keyer *keyerView) *centralArea {
	result := &centralArea{
		entry:    entry,
		callinfo: callinfo,
		esm:      esm,
		workmode: workmode,
		keyer:    keyer,
	}

	result.root = qtlib.NewQWidget2()
	result.root.SetObjectName(*qtlib.NewQAnyStringView3("centralArea"))
	result.root.SetAttribute(qtlib.WA_StyledBackground)
	result.root.SetStyleSheet(CentralAreaStyle)
	rootLayout := qtlib.NewQVBoxLayout(result.root)
	rootLayout.SetContentsMargins(0, 0, 0, 0)

	entryFrame := qtlib.NewQFrame2()
	entryFrame.SetFrameShape(qtlib.QFrame__StyledPanel)
	entryFrame.SetFrameShadow(qtlib.QFrame__Plain)
	result.entryRoot = entryFrame.QWidget
	entry.setRootWidget(result.entryRoot)

	result.entryLayout = qtlib.NewQGridLayout(result.entryRoot)
	result.entryLayout.SetContentsMargins(5, 5, 5, 5)
	result.entryLayout.SetSpacing(5)

	entrySeparator := qtlib.NewQFrame2()
	entrySeparator.SetFrameShape(qtlib.QFrame__HLine)
	entrySeparator.SetFrameShadow(qtlib.QFrame__Sunken)

	result.esmWorkmodeContainer = qtlib.NewQWidget2()
	result.esmWorkmodeContainer.SetObjectName(*qtlib.NewQAnyStringView3("esmWorkmode"))
	esmWorkmodeLayout := qtlib.NewQHBoxLayout(result.esmWorkmodeContainer)
	esmWorkmodeLayout.SetContentsMargins(0, 0, 0, 0)
	esmWorkmodeLayout.AddWidget(esm.widget)
	esmWorkmodeLayout.AddStretch()
	esmWorkmodeLayout.AddWidget(workmode.widget)

	rootLayout.AddWidget(result.entryRoot)
	rootLayout.AddWidget(entrySeparator.QWidget)
	rootLayout.AddWidget(result.esmWorkmodeContainer)
	rootLayout.AddWidget(result.keyer.widget)
	rootLayout.AddStretch()

	entry.onVFO2Enabled = result.setVFO2Enabled

	return result
}

func (a *centralArea) RepaintForThemeChange() {
	a.root.SetStyleSheet(CentralAreaStyle)
	a.root.Update()
}

func (a *centralArea) SetExchangeFields(myExchangeFields, theirExchangeFields []core.ExchangeField, generateSerialExchange bool) {
	a.removeWidgetsFromLayout()
	a.entry.SetExchangeFields(myExchangeFields, theirExchangeFields)
	a.entry.SetSerialClaimLabelsVisible(generateSerialExchange)
	a.callinfo.SetPredictedExchangeFields(theirExchangeFields)
	a.addWidgetsToLayout()
}

func (a *centralArea) setVFO2Enabled(enabled bool) {
	if enabled {
		a.addVFOWidgetsToLayout(core.VFO2, 8)
	} else {
		a.removeVFOWidgetsFromLayout(core.VFO2)
	}
	a.entry.setVFO2Enabled(enabled)
	a.callinfo.SetVFOEnabled(core.VFO2, enabled)
}

func (a *centralArea) addWidgetsToLayout() {
	a.entryLayout.AddWidget2(a.entry.myCallLabel.QWidget, 0, 0)
	for i := range a.entry.myExchangeFields {
		a.entryLayout.AddWidget2(a.entry.myExchangeFields[i].QWidget, 0, i+1)
	}

	a.addVFOWidgetsToLayout(core.VFO1, 1)
	if a.entry.vfo2Enabled {
		a.addVFOWidgetsToLayout(core.VFO2, 8)
	}
}

func (a *centralArea) addVFOWidgetsToLayout(vfo core.VFOID, firstRow int) {
	entry := a.entry.vfo[vfo]
	callinfo := a.callinfo.vfo[vfo]
	lastColumn := 3 + len(entry.theirExchangeFields)

	a.entryLayout.AddWidget3(entry.topSeparator.QWidget, firstRow+0, 0, 1, -1)
	a.entryLayout.AddWidget3(entry.vfoContainer, firstRow+1, 0, 1, 2)
	if entry.serialClaimLabel != nil {
		a.entryLayout.AddWidget2(entry.serialClaimLabel.QWidget, firstRow+1, 2)
	}
	a.entryLayout.AddWidget2(entry.xit.QWidget, firstRow+1, lastColumn-1)
	a.entryLayout.AddWidget2(entry.txIndicator.QWidget, firstRow+1, lastColumn)
	a.entryLayout.AddWidget2(callinfo.callsignLabel.QWidget, firstRow+2, 0)
	for i := range callinfo.predictedExchangeLabels {
		a.entryLayout.AddWidget2(callinfo.predictedExchangeLabels[i].QWidget, firstRow+2, i+1)
	}
	a.entryLayout.AddWidget2(callinfo.valueLabel.QWidget, firstRow+2, lastColumn-1)
	a.entryLayout.AddWidget2(callinfo.qtcStatusLabel.QWidget, firstRow+2, lastColumn)
	a.entryLayout.AddWidget2(entry.callsign.QWidget, firstRow+3, 0)
	for i := range entry.theirExchangeFields {
		a.entryLayout.AddWidget2(entry.theirExchangeFields[i].QWidget, firstRow+3, i+1)
	}
	a.entryLayout.AddWidget2(entry.logButton.QWidget, firstRow+3, lastColumn-1)
	a.entryLayout.AddWidget2(entry.clearButton.QWidget, firstRow+3, lastColumn)
	a.entryLayout.AddWidget3(callinfo.supercheckLabel.QWidget, firstRow+4, 0, 1, -1)
	a.entryLayout.AddWidget3(callinfo.infoContainer, firstRow+5, 0, 1, -1)
	a.entryLayout.AddWidget3(entry.messageLabel.QWidget, firstRow+5, 0, 1, -1)
}

func (a *centralArea) removeWidgetsFromLayout() {
	// row 0: my data: call, exchange
	a.entryLayout.RemoveWidget(a.entry.utcLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.myCallLabel.QWidget)
	for i := range a.entry.myExchangeFields {
		a.entryLayout.RemoveWidget(a.entry.myExchangeFields[i].QWidget)
	}
	a.removeVFOWidgetsFromLayout(core.VFO1)
	if a.entry.vfo2Enabled {
		a.removeVFOWidgetsFromLayout(core.VFO2)
	}
}

func (a *centralArea) removeVFOWidgetsFromLayout(vfo core.VFOID) {
	entry := a.entry.vfo[vfo]
	callinfo := a.callinfo.vfo[vfo]

	a.entryLayout.RemoveWidget(entry.topSeparator.QWidget)
	a.entryLayout.RemoveWidget(entry.vfoLabel.QWidget)
	a.entryLayout.RemoveWidget(entry.frequencyLabel.QWidget)
	a.entryLayout.RemoveWidget(entry.vfoContainer)
	if entry.serialClaimLabel != nil {
		a.entryLayout.RemoveWidget(entry.serialClaimLabel.QWidget)
	}
	a.entryLayout.RemoveWidget(entry.xit.QWidget)
	a.entryLayout.RemoveWidget(entry.txIndicator.QWidget)
	a.entryLayout.RemoveWidget(callinfo.callsignLabel.QWidget)
	for i := range callinfo.predictedExchangeLabels {
		a.entryLayout.RemoveWidget(callinfo.predictedExchangeLabels[i].QWidget)
	}
	a.entryLayout.RemoveWidget(callinfo.valueLabel.QWidget)
	a.entryLayout.RemoveWidget(callinfo.qtcStatusLabel.QWidget)
	a.entryLayout.RemoveWidget(entry.callsign.QWidget)
	for i := range entry.theirExchangeFields {
		a.entryLayout.RemoveWidget(entry.theirExchangeFields[i].QWidget)
	}
	a.entryLayout.RemoveWidget(entry.logButton.QWidget)
	a.entryLayout.RemoveWidget(entry.clearButton.QWidget)
	a.entryLayout.RemoveWidget(callinfo.supercheckLabel.QWidget)
	a.entryLayout.RemoveWidget(callinfo.infoContainer)
	a.entryLayout.RemoveWidget(entry.messageLabel.QWidget)
}
