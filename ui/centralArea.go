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

func (a *centralArea) addWidgetsToLayout() {
	lastColumn := 4 + len(a.entry.theirExchangeFields)

	// row 0: my data: call, exchange
	a.entryLayout.AddWidget2(a.entry.utcLabel.QWidget, 0, 0)
	a.entryLayout.AddWidget2(a.entry.myCallLabel.QWidget, 0, 1)
	for i := range a.entry.myExchangeFields {
		a.entryLayout.AddWidget2(a.entry.myExchangeFields[i].QWidget, 0, i+2)
	}

	// row 1: horizontal separator
	a.entryLayout.AddWidget3(a.entry.topSeparator.QWidget, 1, 0, 1, -1)

	// row 2: Frequency, Band, Mode, [Serial Claim], XIT, TRX
	a.entryLayout.AddWidget2(a.entry.vfoLabel.QWidget, 2, 0)
	a.entryLayout.AddWidget2(a.entry.frequencyLabel.QWidget, 2, 1)
	a.entryLayout.AddWidget2(a.entry.bandModeContainer, 2, 2)
	if a.entry.serialClaimLabel != nil {
		a.entryLayout.AddWidget2(a.entry.serialClaimLabel.QWidget, 2, 3)
	}
	a.entryLayout.AddWidget2(a.entry.xit.QWidget, 2, lastColumn-1)
	a.entryLayout.AddWidget2(a.entry.txIndicator.QWidget, 2, lastColumn)

	// row 3: callinfo: best match, exchange, value, qtcs
	a.entryLayout.AddWidget2(a.callinfo.vfo[core.VFO1].callsignLabel.QWidget, 3, 1)
	for i := range a.callinfo.vfo[core.VFO1].predictedExchangeLabels {
		a.entryLayout.AddWidget2(a.callinfo.vfo[core.VFO1].predictedExchangeLabels[i].QWidget, 3, i+2)
	}
	a.entryLayout.AddWidget2(a.callinfo.vfo[core.VFO1].valueLabel.QWidget, 3, lastColumn-1)
	a.entryLayout.AddWidget2(a.callinfo.vfo[core.VFO1].qtcStatusLabel.QWidget, 3, lastColumn)

	// row 4: their data: call, exchange, log-button, clear-button
	a.entryLayout.AddWidget2(a.entry.theirLabel.QWidget, 4, 0)
	a.entryLayout.AddWidget2(a.entry.callsign.QWidget, 4, 1)
	for i := range a.entry.theirExchangeFields {
		a.entryLayout.AddWidget2(a.entry.theirExchangeFields[i].QWidget, 4, i+2)
	}
	a.entryLayout.AddWidget2(a.entry.logButton.QWidget, 4, lastColumn-1)
	a.entryLayout.AddWidget2(a.entry.clearButton.QWidget, 4, lastColumn)

	// row 5: supercheck
	a.entryLayout.AddWidget3(a.callinfo.vfo[core.VFO1].supercheckLabel.QWidget, 5, 1, 1, -1)

	// row 6: dxcc, personal info
	a.entryLayout.AddWidget3(a.callinfo.vfo[core.VFO1].infoContainer, 6, 0, 1, -1)

	// row 7: message
	a.entryLayout.AddWidget3(a.entry.messageLabel.QWidget, 7, 0, 1, -1)

	// row 8: horizontal separator
	a.entryLayout.AddWidget3(a.entry.vfoSeparator.QWidget, 8, 0, 1, -1)

	// row 9: VFO2 frequency / band / mode / [serial claim] / XIT / TX
	a.entryLayout.AddWidget2(a.entry.vfo2Label.QWidget, 9, 0)
	a.entryLayout.AddWidget2(a.entry.vfo2FrequencyLabel.QWidget, 9, 1)
	a.entryLayout.AddWidget2(a.entry.vfo2BandModeContainer, 9, 2)
	if a.entry.vfo2SerialClaimLabel != nil {
		a.entryLayout.AddWidget2(a.entry.vfo2SerialClaimLabel.QWidget, 9, 3)
	}
	a.entryLayout.AddWidget2(a.entry.vfo2XITIndicator.QWidget, 9, lastColumn-1)
	a.entryLayout.AddWidget2(a.entry.vfo2TXIndicator.QWidget, 9, lastColumn)

	// row 10: VFO2 callinfo: best match, predicted exchange, value, qtcs
	a.entryLayout.AddWidget2(a.callinfo.vfo[core.VFO2].callsignLabel.QWidget, 10, 1)
	for i := range a.callinfo.vfo[core.VFO2].predictedExchangeLabels {
		a.entryLayout.AddWidget2(a.callinfo.vfo[core.VFO2].predictedExchangeLabels[i].QWidget, 10, i+2)
	}
	a.entryLayout.AddWidget2(a.callinfo.vfo[core.VFO2].valueLabel.QWidget, 10, lastColumn-1)
	a.entryLayout.AddWidget2(a.callinfo.vfo[core.VFO2].qtcStatusLabel.QWidget, 10, lastColumn)

	// row 11: VFO2 their data: call, exchange, log, clear
	a.entryLayout.AddWidget2(a.entry.vfo2TheirLabel.QWidget, 11, 0)
	a.entryLayout.AddWidget2(a.entry.vfo2Callsign.QWidget, 11, 1)
	for i := range a.entry.vfo2TheirExchangeFields {
		a.entryLayout.AddWidget2(a.entry.vfo2TheirExchangeFields[i].QWidget, 11, i+2)
	}
	a.entryLayout.AddWidget2(a.entry.vfo2LogButton.QWidget, 11, lastColumn-1)
	a.entryLayout.AddWidget2(a.entry.vfo2ClearButton.QWidget, 11, lastColumn)

	// row 12: VFO2 supercheck
	a.entryLayout.AddWidget3(a.callinfo.vfo[core.VFO2].supercheckLabel.QWidget, 12, 1, 1, -1)

	// row 13: VFO2 dxcc, personal info
	a.entryLayout.AddWidget3(a.callinfo.vfo[core.VFO2].infoContainer, 13, 0, 1, -1)

	// row 14: VFO2 message
	a.entryLayout.AddWidget3(a.entry.vfo2MessageLabel.QWidget, 14, 0, 1, -1)
}

func (a *centralArea) removeWidgetsFromLayout() {
	// row 0: my data: call, exchange
	a.entryLayout.RemoveWidget(a.entry.utcLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.myCallLabel.QWidget)
	for i := range a.entry.myExchangeFields {
		a.entryLayout.RemoveWidget(a.entry.myExchangeFields[i].QWidget)
	}

	// row 1: horizontal separator
	a.entryLayout.RemoveWidget(a.entry.topSeparator.QWidget)

	// row 2: Frequency, Band, Mode, [Serial Claim], XIT, TRX
	a.entryLayout.RemoveWidget(a.entry.vfoLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.frequencyLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.bandModeContainer)
	if a.entry.serialClaimLabel != nil {
		a.entryLayout.RemoveWidget(a.entry.serialClaimLabel.QWidget)
	}
	a.entryLayout.RemoveWidget(a.entry.xit.QWidget)
	a.entryLayout.RemoveWidget(a.entry.txIndicator.QWidget)

	// row 3: callinfo: best match, exchange, value, qtcs
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO1].callsignLabel.QWidget)
	for i := range a.callinfo.vfo[core.VFO1].predictedExchangeLabels {
		a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO1].predictedExchangeLabels[i].QWidget)
	}
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO1].valueLabel.QWidget)
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO1].qtcStatusLabel.QWidget)

	// row 4: their data: call, exchange, log-button, clear-button
	a.entryLayout.RemoveWidget(a.entry.theirLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.callsign.QWidget)
	for i := range a.entry.theirExchangeFields {
		a.entryLayout.RemoveWidget(a.entry.theirExchangeFields[i].QWidget)
	}
	a.entryLayout.RemoveWidget(a.entry.logButton.QWidget)
	a.entryLayout.RemoveWidget(a.entry.clearButton.QWidget)

	// row 5: supercheck
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO1].supercheckLabel.QWidget)

	// row 6: dxcc, personal info
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO1].infoContainer)

	// row 7: message
	a.entryLayout.RemoveWidget(a.entry.messageLabel.QWidget)

	// row 8: horizontal separator
	a.entryLayout.RemoveWidget(a.entry.vfoSeparator.QWidget)

	// row 9: VFO2
	a.entryLayout.RemoveWidget(a.entry.vfo2Label.QWidget)
	a.entryLayout.RemoveWidget(a.entry.vfo2FrequencyLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.vfo2BandModeContainer)
	if a.entry.vfo2SerialClaimLabel != nil {
		a.entryLayout.RemoveWidget(a.entry.vfo2SerialClaimLabel.QWidget)
	}
	a.entryLayout.RemoveWidget(a.entry.vfo2XITIndicator.QWidget)
	a.entryLayout.RemoveWidget(a.entry.vfo2TXIndicator.QWidget)

	// row 10: VFO2 callinfo
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO2].callsignLabel.QWidget)
	for i := range a.callinfo.vfo[core.VFO2].predictedExchangeLabels {
		a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO2].predictedExchangeLabels[i].QWidget)
	}
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO2].valueLabel.QWidget)
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO2].qtcStatusLabel.QWidget)

	// row 11: VFO2 their data
	a.entryLayout.RemoveWidget(a.entry.vfo2TheirLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.vfo2Callsign.QWidget)
	for i := range a.entry.vfo2TheirExchangeFields {
		a.entryLayout.RemoveWidget(a.entry.vfo2TheirExchangeFields[i].QWidget)
	}
	a.entryLayout.RemoveWidget(a.entry.vfo2LogButton.QWidget)
	a.entryLayout.RemoveWidget(a.entry.vfo2ClearButton.QWidget)

	// row 12: VFO2 supercheck
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO2].supercheckLabel.QWidget)

	// row 13: VFO2 dxcc, personal info
	a.entryLayout.RemoveWidget(a.callinfo.vfo[core.VFO2].infoContainer)

	// row 14: VFO2 message
	a.entryLayout.RemoveWidget(a.entry.vfo2MessageLabel.QWidget)
}
