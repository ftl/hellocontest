package ui

import (
	qtlib "github.com/mappu/miqt/qt6"
)

// SummaryController is the callback surface for the summary dialog.
type SummaryController interface {
	OperatorModes() []string
	Overlays() []string
	PowerModes() []string

	SetOperatorMode(string)
	SetOverlay(string)
	SetPowerMode(string)
	SetAssisted(bool)

	SetOpenAfterExport(bool)
}

type summaryView struct {
	controller SummaryController

	root *qtlib.QWidget

	contestName  *qtlib.QLineEdit
	cabrilloName *qtlib.QLineEdit
	startTime    *qtlib.QLineEdit
	callsign     *qtlib.QLineEdit
	myExchanges  *qtlib.QLineEdit

	operatorModeCombo *qtlib.QComboBox
	overlayCombo      *qtlib.QComboBox
	powerModeCombo    *qtlib.QComboBox
	assisted          *qtlib.QCheckBox

	workedModes   *qtlib.QLineEdit
	workedBands   *qtlib.QLineEdit
	operatingTime *qtlib.QLineEdit
	breakTime     *qtlib.QLineEdit
	breaks        *qtlib.QLineEdit

	scoreTable *scoreTable

	openAfterExport *qtlib.QCheckBox

	ignoreChangedEvent bool
}

func (v *summaryView) doIgnore(f func()) {
	v.ignoreChangedEvent = true
	defer func() { v.ignoreChangedEvent = false }()
	f()
}

func newSummaryView(controller SummaryController) *summaryView {
	v := &summaryView{controller: controller}

	v.root = qtlib.NewQWidget2()
	root := qtlib.NewQVBoxLayout(v.root)

	columns := qtlib.NewQWidget2()
	columnsLayout := qtlib.NewQHBoxLayout(columns)
	columnsLayout.SetContentsMargins(0, 0, 0, 0)
	columnsLayout.SetSpacing(20)

	columnsLayout.AddWidget(v.buildLeftColumn())
	columnsLayout.AddWidget(v.buildRightColumn())

	root.AddWidget(columns)

	v.openAfterExport = qtlib.NewQCheckBox3("Open the file after export")
	v.openAfterExport.OnToggled(func(checked bool) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetOpenAfterExport(checked)
	})
	root.AddWidget(v.openAfterExport.QWidget)

	return v
}

func (v *summaryView) buildLeftColumn() *qtlib.QWidget {
	column := qtlib.NewQWidget2()
	col := qtlib.NewQVBoxLayout(column)
	col.SetContentsMargins(0, 0, 0, 0)

	form := qtlib.NewQFormLayout2()
	v.contestName = newReadOnlyLineEdit()
	form.AddRow3("Contest Name:", v.contestName.QWidget)
	v.cabrilloName = newReadOnlyLineEdit()
	form.AddRow3("Cabrillo Name:", v.cabrilloName.QWidget)
	v.startTime = newReadOnlyLineEdit()
	form.AddRow3("Start Time:", v.startTime.QWidget)
	v.callsign = newReadOnlyLineEdit()
	form.AddRow3("Callsign:", v.callsign.QWidget)
	v.myExchanges = newReadOnlyLineEdit()
	form.AddRow3("My Exchanges:", v.myExchanges.QWidget)
	col.AddLayout(form.QLayout)

	header := qtlib.NewQLabel3("Working Condition")
	header.SetStyleSheet(BoldSectionStyle)
	col.AddWidget(header.QWidget)

	form2 := qtlib.NewQFormLayout2()
	v.operatorModeCombo = qtlib.NewQComboBox2()
	for _, item := range v.controller.OperatorModes() {
		v.operatorModeCombo.AddItem(item)
	}
	v.operatorModeCombo.OnCurrentTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetOperatorMode(text)
	})
	form2.AddRow3("Operator Mode:", v.operatorModeCombo.QWidget)

	v.overlayCombo = qtlib.NewQComboBox2()
	for _, item := range v.controller.Overlays() {
		v.overlayCombo.AddItem(item)
	}
	v.overlayCombo.OnCurrentTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetOverlay(text)
	})
	form2.AddRow3("Overlay:", v.overlayCombo.QWidget)

	v.powerModeCombo = qtlib.NewQComboBox2()
	for _, item := range v.controller.PowerModes() {
		v.powerModeCombo.AddItem(item)
	}
	v.powerModeCombo.OnCurrentTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetPowerMode(text)
	})
	form2.AddRow3("Power:", v.powerModeCombo.QWidget)

	v.assisted = qtlib.NewQCheckBox3("Assisted")
	v.assisted.OnToggled(func(checked bool) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetAssisted(checked)
	})
	form2.AddRowWithWidget(v.assisted.QWidget)
	col.AddLayout(form2.QLayout)

	col.AddStretch()
	return column
}

func (v *summaryView) buildRightColumn() *qtlib.QWidget {
	column := qtlib.NewQWidget2()
	col := qtlib.NewQVBoxLayout(column)
	col.SetContentsMargins(0, 0, 0, 0)

	form := qtlib.NewQFormLayout2()
	v.workedModes = newReadOnlyLineEdit()
	form.AddRow3("Worked Modes:", v.workedModes.QWidget)
	v.workedBands = newReadOnlyLineEdit()
	form.AddRow3("Worked Bands:", v.workedBands.QWidget)
	v.operatingTime = newReadOnlyLineEdit()
	form.AddRow3("Operating Time:", v.operatingTime.QWidget)
	v.breakTime = newReadOnlyLineEdit()
	form.AddRow3("Break Time:", v.breakTime.QWidget)
	v.breaks = newReadOnlyLineEdit()
	form.AddRow3("Breaks:", v.breaks.QWidget)
	col.AddLayout(form.QLayout)

	header := qtlib.NewQLabel3("Claimed Score")
	header.SetStyleSheet(BoldSectionStyle)
	col.AddWidget(header.QWidget)

	v.scoreTable = newScoreTable()
	col.AddWidget(v.scoreTable.widget.QWidget)

	col.AddStretch()
	return column
}

func newReadOnlyLineEdit() *qtlib.QLineEdit {
	edit := qtlib.NewQLineEdit2()
	edit.SetReadOnly(true)
	return edit
}
