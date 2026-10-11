package ui

import (
	"github.com/ftl/hellocontest/core"
	qtlib "github.com/mappu/miqt/qt6"
)

type centralArea struct {
	root *qtlib.QWidget

	entryArea *entryArea

	esmWorkmodeContainer *qtlib.QWidget
}

func newCentralArea(entry *entryView, callinfo *callinfoView, esm *esmView, workmode *workmodeView, keyer *keyerView) *centralArea {
	result := &centralArea{
		entryArea: newEntryArea(entry, callinfo),
	}

	result.root = qtlib.NewQWidget2()
	result.root.SetObjectName(*qtlib.NewQAnyStringView3("centralArea"))
	result.root.SetAttribute(qtlib.WA_StyledBackground)
	result.root.SetStyleSheet(CentralAreaStyle)
	rootLayout := qtlib.NewQVBoxLayout(result.root)
	rootLayout.SetContentsMargins(0, 0, 0, 0)

	entrySeparator := qtlib.NewQFrame2()
	entrySeparator.SetFrameShape(qtlib.QFrame__HLine)
	entrySeparator.SetFrameShadow(qtlib.QFrame__Sunken)

	result.esmWorkmodeContainer = qtlib.NewQWidget2()
	result.esmWorkmodeContainer.SetObjectName(*qtlib.NewQAnyStringView3("esmWorkmode"))
	esmWorkmodeLayout := qtlib.NewQHBoxLayout(result.esmWorkmodeContainer)
	esmWorkmodeLayout.SetContentsMargins(0, 0, 0, 0)
	esmWorkmodeLayout.AddWidget(esm.widget)
	esmWorkmodeLayout.AddStretch()
	esmWorkmodeLayout.AddWidget(keyer.lastTransmissionLabel.QWidget)
	esmWorkmodeLayout.AddStretch()
	esmWorkmodeLayout.AddWidget(workmode.widget)

	rootLayout.AddWidget(result.entryArea.root)
	rootLayout.AddWidget(entrySeparator.QWidget)
	rootLayout.AddWidget(result.esmWorkmodeContainer)
	rootLayout.AddWidget(keyer.widget)
	rootLayout.AddStretch()

	return result
}

func (a *centralArea) RepaintForThemeChange() {
	a.root.SetStyleSheet(CentralAreaStyle)
	a.root.Update()
}

func (a *centralArea) SetExchangeFields(myExchangeFields, theirExchangeFields []core.ExchangeField, generateSerialExchange bool) {
	a.entryArea.SetExchangeFields(myExchangeFields, theirExchangeFields, generateSerialExchange)
}
