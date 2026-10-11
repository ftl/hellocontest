package ui

import (
	"fmt"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
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

	vfoContainer     *qtlib.QWidget
	vfoContainerName string
	infoContainer    *qtlib.QWidget

	txVFO        bool
	parrotActive bool
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

	result.setupVFOContainer()
	result.setupInfoContainer()
	result.setupEntryWidgets()
	result.setupCallinfoWidgets()

	return result
}

func (a *vfoArea) setupVFOContainer() {
	a.vfoContainerName = a.widgetPrefix + "VFOContainer"

	a.vfoContainer = qtlib.NewQWidget2()
	a.vfoContainer.SetObjectName(*qtlib.NewQAnyStringView3(a.vfoContainerName))
	a.vfoContainer.SetAttribute(qtlib.WA_StyledBackground)

	layout := qtlib.NewQHBoxLayout(a.vfoContainer)
	layout.AddWidget(a.entry.vfoLabel.QWidget)
	layout.AddWidget(a.entry.txIndicator.QWidget)
	layout.AddWidget(a.entry.frequencyLabel.QWidget)
	layout.AddWidget(a.entry.band.QWidget)
	layout.AddWidget(a.entry.mode.QWidget)
	layout.AddWidget(a.entry.xit.QWidget)
	layout.AddWidget(a.entry.rit.QWidget)
	layout.AddStretch()
}

func (a *vfoArea) setupInfoContainer() {
	a.infoContainer = qtlib.NewQWidget2()

	layout := qtlib.NewQHBoxLayout(a.infoContainer)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.AddWidget(a.callinfo.dxccLabel.QWidget)
	layout.AddWidget(a.callinfo.userInfoLabel.QWidget)
}

func (a *vfoArea) setupEntryWidgets() {
	a.entry.vfoLabel.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "Label"))
	a.entry.vfoLabel.SetAlignment(qtlib.AlignLeading | qtlib.AlignVCenter)
	setFixedTextWidth(a.entry.vfoLabel.QWidget, "VFO 2 S&P", RoundedLabelPadding)

	a.entry.txIndicator.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "TX"))
	a.entry.txIndicator.SetStyleSheet(TXIndicatorInactiveStyle)
	a.entry.txIndicator.SetAlignment(qtlib.AlignCenter | qtlib.AlignVCenter)
	setFixedBoldTextWidth(a.entry.txIndicator.QWidget, txIndicatorWidthSample, RoundedLabelPadding)
	retainSizeWhenHidden(a.entry.txIndicator.QWidget)

	a.entry.frequencyLabel.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "FrequencyLabel"))
	a.entry.frequencyLabel.SetAlignment(qtlib.AlignTrailing | qtlib.AlignVCenter)
	setFixedTextWidth(a.entry.frequencyLabel.QWidget, "999999.99 kHz", 0)

	a.entry.band.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "BandCombo"))
	a.entry.mode.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "ModeCombo"))

	a.entry.xit.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "XIT"))
	a.entry.xit.SetVisible(false)
	a.entry.rit.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "RIT"))
	a.entry.rit.SetVisible(false)

	a.entry.callsign.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "CallsignEntry"))
	a.entry.callsign.SetStyleSheet(EntryFieldStyle)
	setFixedTextWidth(a.entry.callsign.QWidget, callsignFieldWidthSample, exchangeFieldPadding)
}

func (a *vfoArea) setupCallinfoWidgets() {
	a.callinfo.callsignLabel.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "predictedBestMatch"))
	a.callinfo.callsignLabel.SetTextFormat(qtlib.RichText)

	a.callinfo.valueLabel.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "predictedValue"))
	a.callinfo.valueLabel.SetTextFormat(qtlib.RichText)

	a.callinfo.qtcStatusLabel.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "qtcStatus"))
	a.callinfo.qtcStatusLabel.SetTextFormat(qtlib.RichText)
	a.callinfo.qtcStatusLabel.SetAlignment(qtlib.AlignVCenter | qtlib.AlignTrailing)

	a.callinfo.supercheckLabel.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "supercheckLabel"))
	a.callinfo.supercheckLabel.SetTextFormat(qtlib.RichText)
	a.callinfo.supercheckLabel.QWidget.SetSizePolicy2(qtlib.QSizePolicy__Ignored, qtlib.QSizePolicy__Preferred)
	a.callinfo.supercheckLabel.SetMinimumWidth(0)

	a.callinfo.dxccLabel.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "dxccLabel"))
	a.callinfo.dxccLabel.QWidget.SetSizePolicy2(qtlib.QSizePolicy__Ignored, qtlib.QSizePolicy__Preferred)
	a.callinfo.dxccLabel.SetMinimumWidth(0)

	a.callinfo.userInfoLabel.SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "userInfoLabel"))
	a.callinfo.userInfoLabel.QWidget.SetSizePolicy2(qtlib.QSizePolicy__Ignored, qtlib.QSizePolicy__Preferred)
	a.callinfo.userInfoLabel.SetMinimumWidth(0)
	a.callinfo.userInfoLabel.SetAlignment(qtlib.AlignVCenter | qtlib.AlignTrailing)
}

func (a *vfoArea) setupExchangeFields(fields []core.ExchangeField) {
	for i, field := range fields {
		if i >= len(a.entry.theirExchangeFields) {
			break
		}
		editField := a.entry.theirExchangeFields[i]
		editField.SetObjectName(*qtlib.NewQAnyStringView3(a.exchangeFieldName(field)))
		editField.SetStyleSheet(EntryFieldStyle)
		setFixedTextWidth(editField.QWidget, exchangeFieldWidthSample, exchangeFieldPadding)
	}

	if len(a.callinfo.predictedExchangeLabels) > 0 {
		a.callinfo.predictedExchangeLabels[0].SetObjectName(*qtlib.NewQAnyStringView3(a.widgetPrefix + "predictedExchange"))
	}
}

func (a *vfoArea) exchangeFieldName(field core.ExchangeField) string {
	if a.vfo == core.VFO2 {
		return "vfo2_" + string(field.Field)
	}
	return string(field.Field)
}

func (a *vfoArea) setEnabled(enabled bool) {
	a.root.SetVisible(enabled)
}

func (a *vfoArea) setIncrementalTuningVisible(kind core.IncrementalTuningKind, visible bool) {
	incrementalTuningWidget(a.entry, kind).SetVisible(visible)
}

func (a *vfoArea) setTXVFO(txVFO bool) {
	a.txVFO = txVFO
	a.applyTXIndicatorVisibility()
}

func (a *vfoArea) setTXState(ptt bool, parrotActive bool) {
	a.parrotActive = parrotActive
	a.applyTXIndicatorVisibility()

	// TODO: use a property with a selective style
	style := TXIndicatorInactiveStyle
	if ptt {
		style = TXIndicatorActiveStyle
	}
	a.entry.txIndicator.SetStyleSheet(style)
}

func (a *vfoArea) applyTXIndicatorVisibility() {
	a.entry.txIndicator.SetVisible(a.txVFO || a.parrotActive)
}

func (a *vfoArea) setActive(active bool) {
	// TODO: use a property with a selective style
	style := VFOInactiveStyle
	if active {
		style = fmt.Sprintf(VFOActiveStyle, a.vfoContainerName)
	}
	a.vfoContainer.SetStyleSheet(style)
}

func (a *vfoArea) addWidgetsToLayout() {
	lastColumn := 3 + len(a.entry.theirExchangeFields)
	a.setFillerColumn(lastColumn - 2)

	a.entryLayout.AddWidget3(a.vfoContainer, 0, 0, 1, -1)
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
	a.entryLayout.AddWidget3(a.infoContainer, 4, 0, 1, -1)
	a.entryLayout.AddWidget3(a.entry.messageLabel.QWidget, 4, 0, 1, -1)
}

func (a *vfoArea) setFillerColumn(column int) {
	for i := 0; i < a.entryLayout.ColumnCount(); i++ {
		a.entryLayout.SetColumnStretch(i, 0)
	}
	a.entryLayout.SetColumnStretch(column, 1)
}

func (a *vfoArea) removeWidgetsFromLayout() {
	a.entryLayout.RemoveWidget(a.vfoContainer)
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
	a.entryLayout.RemoveWidget(a.infoContainer)
	a.entryLayout.RemoveWidget(a.entry.messageLabel.QWidget)
}
