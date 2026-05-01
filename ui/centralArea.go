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

func (a *centralArea) SetExchangeFields(myExchangeFields, theirExchangeFields []core.ExchangeField) {
	a.removeWidgetsFromLayout()
	a.entry.SetExchangeFields(myExchangeFields, theirExchangeFields)
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

	// row 1: VFO label + horizontal separator
	a.entryLayout.AddWidget3(a.entry.topSeparator.QWidget, 1, 0, 1, lastColumn)

	// row 2: Frequency, Band, Mode, XIT, TRX
	a.entryLayout.AddWidget2(a.entry.vfoLabel.QWidget, 2, 0)
	a.entryLayout.AddWidget2(a.entry.frequencyLabel.QWidget, 2, 1)
	a.entryLayout.AddWidget2(a.entry.bandModeContainer, 2, 2)
	a.entryLayout.AddWidget2(a.entry.xit.QWidget, 2, lastColumn-1)
	a.entryLayout.AddWidget2(a.entry.txIndicator.QWidget, 2, lastColumn)

	// row 3: callinfo: best match, exchange, value, qtcs
	a.entryLayout.AddWidget2(a.callinfo.callsignLabel.QWidget, 3, 1)
	for i := range a.callinfo.predictedExchangeLabels {
		a.entryLayout.AddWidget2(a.callinfo.predictedExchangeLabels[i].QWidget, 3, i+2)
	}
	a.entryLayout.AddWidget2(a.callinfo.valueLabel.QWidget, 3, lastColumn-1)
	a.entryLayout.AddWidget2(a.callinfo.qtcStatusLabel.QWidget, 3, lastColumn)

	// row 4: their data: call, exchange, log-button, clear-button
	a.entryLayout.AddWidget2(a.entry.theirLabel.QWidget, 4, 0)
	a.entryLayout.AddWidget2(a.entry.callsign.QWidget, 4, 1)
	for i := range a.entry.theirExchangeFields {
		a.entryLayout.AddWidget2(a.entry.theirExchangeFields[i].QWidget, 4, i+2)
	}
	a.entryLayout.AddWidget2(a.entry.logButton.QWidget, 4, lastColumn-1)
	a.entryLayout.AddWidget2(a.entry.clearButton.QWidget, 4, lastColumn)

	// row 5: supercheck
	a.entryLayout.AddWidget3(a.callinfo.supercheckLabel.QWidget, 5, 1, 1, -1)

	// row 6: dxcc, personal info
	a.entryLayout.AddWidget3(a.callinfo.infoContainer, 6, 0, 1, -1)

	// row 7: message
	a.entryLayout.AddWidget3(a.entry.messageLabel.QWidget, 7, 0, 1, -1)
}

func (a *centralArea) removeWidgetsFromLayout() {
	// row 0: my data: call, exchange
	a.entryLayout.RemoveWidget(a.entry.utcLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.myCallLabel.QWidget)
	for i := range a.entry.myExchangeFields {
		a.entryLayout.RemoveWidget(a.entry.myExchangeFields[i].QWidget)
	}

	// row 1: VFO label + horizontal separator
	a.entryLayout.RemoveWidget(a.entry.vfoLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.topSeparator.QWidget)

	// row 2: Frequency, Band, Mode, XIT, TRX
	a.entryLayout.RemoveWidget(a.entry.frequencyLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.bandModeContainer)
	a.entryLayout.RemoveWidget(a.entry.xit.QWidget)
	a.entryLayout.RemoveWidget(a.entry.txIndicator.QWidget)

	// row 3: callinfo: best match, exchange, value, qtcs
	a.entryLayout.RemoveWidget(a.callinfo.callsignLabel.QWidget)
	for i := range a.callinfo.predictedExchangeLabels {
		a.entryLayout.RemoveWidget(a.callinfo.predictedExchangeLabels[i].QWidget)
	}
	a.entryLayout.RemoveWidget(a.callinfo.valueLabel.QWidget)

	// row 4: their data: call, exchange, log-button, clear-button
	a.entryLayout.RemoveWidget(a.entry.theirLabel.QWidget)
	a.entryLayout.RemoveWidget(a.entry.callsign.QWidget)
	for i := range a.entry.theirExchangeFields {
		a.entryLayout.RemoveWidget(a.entry.theirExchangeFields[i].QWidget)
	}
	a.entryLayout.RemoveWidget(a.entry.logButton.QWidget)
	a.entryLayout.RemoveWidget(a.entry.clearButton.QWidget)

	// row 5: supercheck
	a.entryLayout.RemoveWidget(a.callinfo.supercheckLabel.QWidget)

	// row 6: dxcc, personal info
	a.entryLayout.RemoveWidget(a.callinfo.infoContainer)

	// row 7: message
	a.entryLayout.RemoveWidget(a.entry.messageLabel.QWidget)
}
