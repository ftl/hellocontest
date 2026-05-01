package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

// SettingsController is the callback surface the settings dialog uses.
type SettingsController interface {
	Save()
	Reset()
	OpenConfigurationFile()

	EnterStationCallsign(string)
	EnterStationOperator(string)
	EnterStationLocator(string)

	SelectContestIdentifier(string)
	OpenContestRulesPage()
	OpenContestUploadPage()
	OpenCallHistoryArchivePage()
	ClearCallHistory()

	EnterContestExchangeValue(core.EntryField, string)
	EnterContestGenerateSerialExchange(bool)
	EnterContestGenerateReport(bool)

	EnterContestName(string)
	EnterContestStartTime(string)
	SetContestStartTimeToday()
	SetContestStartTimeNow()
	SetOperationModeSprint(bool)
	SetContestEnableQTCs(bool)
	EnterContestCallHistoryFile(string)
	EnterContestCallHistoryFieldName(core.EntryField, string)
	EnterQSOsGoal(string)
	EnterPointsGoal(string)
	EnterMultisGoal(string)
}

type exchangeRow struct {
	label *qtlib.QLabel
	entry *qtlib.QLineEdit
	field core.EntryField
}

type callHistoryFieldRow struct {
	label *qtlib.QLabel
	combo *qtlib.QComboBox
	field core.EntryField
}

type settingsView struct {
	dialog *qtlib.QDialog

	root *qtlib.QWidget

	stationCallsign *qtlib.QLineEdit
	stationOperator *qtlib.QLineEdit
	stationLocator  *qtlib.QLineEdit

	contestCombo *qtlib.QComboBox
	contestIDs   []string
	contestTexts []string

	contestName       *qtlib.QLineEdit
	contestStartTime  *qtlib.QLineEdit
	startTimeTodayBtn *qtlib.QPushButton
	startTimeNowBtn   *qtlib.QPushButton
	sprintMode        *qtlib.QCheckBox
	enableQTCs        *qtlib.QCheckBox

	exchangeGrid        *qtlib.QGridLayout
	exchangeRows        []exchangeRow
	generateSerialChk   *qtlib.QCheckBox
	generateReportChk   *qtlib.QCheckBox
	serialExchangeEntry *qtlib.QLineEdit
	reportExchangeEntry *qtlib.QLineEdit

	callHistoryPath                *qtlib.QLineEdit
	callHistoryBrowseBtn           *qtlib.QPushButton
	callHistoryClearBtn            *qtlib.QPushButton
	callHistoryArchiveBtn          *qtlib.QPushButton
	callHistoryFieldGrid           *qtlib.QGridLayout
	callHistoryFieldRows           []callHistoryFieldRow
	availableCallHistoryFieldNames []string

	qsosGoal   *qtlib.QLineEdit
	pointsGoal *qtlib.QLineEdit
	multisGoal *qtlib.QLineEdit

	messageLabel *qtlib.QLabel

	resetBtn *qtlib.QPushButton
	closeBtn *qtlib.QPushButton

	controller         SettingsController
	ignoreChangedEvent bool
}

func (v *settingsView) doIgnore(f func()) {
	v.ignoreChangedEvent = true
	defer func() { v.ignoreChangedEvent = false }()
	f()
}

func newSettingsView(dialog *qtlib.QDialog, controller SettingsController) *settingsView {
	v := &settingsView{dialog: dialog, controller: controller}

	v.root = qtlib.NewQWidget2()
	root := qtlib.NewQVBoxLayout(v.root)

	root.AddWidget(v.buildStationGroup().QWidget)
	root.AddWidget(v.buildContestGroup().QWidget)
	root.AddWidget(v.buildExchangeGroup().QWidget)
	root.AddWidget(v.buildCallHistoryGroup().QWidget)
	root.AddWidget(v.buildGoalsGroup().QWidget)

	v.messageLabel = qtlib.NewQLabel3("")
	root.AddWidget(v.messageLabel.QWidget)

	root.AddWidget(v.buildButtonRow())

	return v
}

func (v *settingsView) buildStationGroup() *qtlib.QGroupBox {
	box := qtlib.NewQGroupBox3("Station")
	form := qtlib.NewQFormLayout(box.QWidget)

	v.stationCallsign = qtlib.NewQLineEdit2()
	v.stationCallsign.OnTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.EnterStationCallsign(text)
	})
	form.AddRow3("Callsign:", v.stationCallsign.QWidget)

	v.stationOperator = qtlib.NewQLineEdit2()
	v.stationOperator.OnTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.EnterStationOperator(text)
	})
	form.AddRow3("Operator:", v.stationOperator.QWidget)

	v.stationLocator = qtlib.NewQLineEdit2()
	v.stationLocator.OnTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.EnterStationLocator(text)
	})
	form.AddRow3("Locator:", v.stationLocator.QWidget)

	return box
}

func (v *settingsView) buildContestGroup() *qtlib.QGroupBox {
	box := qtlib.NewQGroupBox3("Contest")
	form := qtlib.NewQFormLayout(box.QWidget)

	v.contestCombo = makeFilterableCombo()
	v.contestCombo.OnActivated(func(index int) {
		if v.ignoreChangedEvent || index < 0 || index >= len(v.contestIDs) {
			return
		}
		v.controller.SelectContestIdentifier(v.contestIDs[index])
	})
	form.AddRow3("Contest:", v.contestCombo.QWidget)

	v.contestName = qtlib.NewQLineEdit2()
	v.contestName.OnTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.EnterContestName(text)
	})
	form.AddRow3("Name:", v.contestName.QWidget)

	// Start time: entry + Today + Now inside an HBox.
	startTimeLine := qtlib.NewQWidget2()
	startTimeHBox := qtlib.NewQHBoxLayout(startTimeLine)
	startTimeHBox.SetContentsMargins(0, 0, 0, 0)
	v.contestStartTime = qtlib.NewQLineEdit2()
	v.contestStartTime.SetPlaceholderText("DD-MM-YYYY HH:mm")
	v.contestStartTime.OnTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.EnterContestStartTime(text)
	})
	startTimeHBox.AddWidget(v.contestStartTime.QWidget)
	v.startTimeTodayBtn = qtlib.NewQPushButton3("Today")
	v.startTimeTodayBtn.OnClicked(func() { v.controller.SetContestStartTimeToday() })
	startTimeHBox.AddWidget(v.startTimeTodayBtn.QWidget)
	v.startTimeNowBtn = qtlib.NewQPushButton3("Now")
	v.startTimeNowBtn.OnClicked(func() { v.controller.SetContestStartTimeNow() })
	startTimeHBox.AddWidget(v.startTimeNowBtn.QWidget)
	form.AddRow3("Start Time (UTC):", startTimeLine)

	v.sprintMode = qtlib.NewQCheckBox3("Sprint mode (auto-switch workmode after each QSO)")
	v.sprintMode.OnToggled(func(checked bool) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetOperationModeSprint(checked)
	})
	form.AddRow3("Operation Mode:", v.sprintMode.QWidget)

	v.enableQTCs = qtlib.NewQCheckBox3("Enable QTCs")
	v.enableQTCs.OnToggled(func(checked bool) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SetContestEnableQTCs(checked)
	})
	form.AddRow3("QTCs:", v.enableQTCs.QWidget)

	return box
}

func (v *settingsView) buildExchangeGroup() *qtlib.QGroupBox {
	box := qtlib.NewQGroupBox3("My Exchange")
	box.SetObjectName(*qtlib.NewQAnyStringView3("myExchangeGroup"))
	v.exchangeGrid = qtlib.NewQGridLayout(box.QWidget)
	return box
}

func (v *settingsView) buildCallHistoryGroup() *qtlib.QGroupBox {
	box := qtlib.NewQGroupBox3("Call History")
	outer := qtlib.NewQVBoxLayout(box.QWidget)

	fileLine := qtlib.NewQWidget2()
	fileHBox := qtlib.NewQHBoxLayout(fileLine)
	fileHBox.SetContentsMargins(0, 0, 0, 0)
	v.callHistoryPath = qtlib.NewQLineEdit2()
	v.callHistoryPath.SetReadOnly(true)
	fileHBox.AddWidget(v.callHistoryPath.QWidget)

	v.callHistoryBrowseBtn = qtlib.NewQPushButton3("Browse…")
	v.callHistoryBrowseBtn.OnClicked(func() {
		dlg := qtlib.NewQFileDialog4(v.dialog.QWidget, "Select Call History File")
		dlg.SetAcceptMode(qtlib.QFileDialog__AcceptOpen)
		dlg.SetFileMode(qtlib.QFileDialog__ExistingFile)
		dlg.SetNameFilter("Call history files (*.txt *.csv);;All Files (*)")
		dlg.SetWindowFlags(
			qtlib.Window |
				qtlib.CustomizeWindowHint |
				qtlib.WindowTitleHint |
				qtlib.WindowCloseButtonHint,
		)
		if dlg.Exec() != int(qtlib.QDialog__Accepted) {
			return
		}
		files := dlg.SelectedFiles()
		if len(files) == 0 {
			return
		}
		v.controller.EnterContestCallHistoryFile(files[0])
	})
	fileHBox.AddWidget(v.callHistoryBrowseBtn.QWidget)

	v.callHistoryClearBtn = qtlib.NewQPushButton3("Clear")
	v.callHistoryClearBtn.OnClicked(func() { v.controller.ClearCallHistory() })
	fileHBox.AddWidget(v.callHistoryClearBtn.QWidget)

	v.callHistoryArchiveBtn = qtlib.NewQPushButton3("🌐")
	v.callHistoryArchiveBtn.OnClicked(func() { v.controller.OpenCallHistoryArchivePage() })
	fileHBox.AddWidget(v.callHistoryArchiveBtn.QWidget)

	outer.AddWidget(fileLine)

	fieldsContainer := qtlib.NewQWidget2()
	v.callHistoryFieldGrid = qtlib.NewQGridLayout(fieldsContainer)
	outer.AddWidget(fieldsContainer)

	return box
}

func (v *settingsView) buildGoalsGroup() *qtlib.QGroupBox {
	box := qtlib.NewQGroupBox3("Goals")
	layout := qtlib.NewQHBoxLayout(box.QWidget)

	layout.AddWidget(qtlib.NewQLabel3("QSOs/hour:").QWidget)
	v.qsosGoal = qtlib.NewQLineEdit2()
	v.qsosGoal.OnTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.EnterQSOsGoal(text)
	})
	layout.AddWidget(v.qsosGoal.QWidget)

	layout.AddWidget(qtlib.NewQLabel3("Points/hour:").QWidget)
	v.pointsGoal = qtlib.NewQLineEdit2()
	v.pointsGoal.OnTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.EnterPointsGoal(text)
	})
	layout.AddWidget(v.pointsGoal.QWidget)

	layout.AddWidget(qtlib.NewQLabel3("Multis/hour:").QWidget)
	v.multisGoal = qtlib.NewQLineEdit2()
	v.multisGoal.OnTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.EnterMultisGoal(text)
	})
	layout.AddWidget(v.multisGoal.QWidget)

	return box
}

func (v *settingsView) buildButtonRow() *qtlib.QWidget {
	line := qtlib.NewQWidget2()
	hbox := qtlib.NewQHBoxLayout(line)
	hbox.SetContentsMargins(0, 0, 0, 0)

	v.resetBtn = qtlib.NewQPushButton3("Reset")
	v.resetBtn.OnClicked(func() { v.controller.Reset() })
	hbox.AddWidget(v.resetBtn.QWidget)

	hbox.AddStretch()

	v.closeBtn = qtlib.NewQPushButton3("Close")
	v.closeBtn.OnClicked(func() { v.dialog.Reject() })
	hbox.AddWidget(v.closeBtn.QWidget)

	return line
}

// settings.View implementation

func (v *settingsView) ShowMessage(message string) {
	v.messageLabel.SetText(message)
	v.messageLabel.SetVisible(true)
}

func (v *settingsView) HideMessage() {
	v.messageLabel.SetVisible(false)
}

func (v *settingsView) SetStationCallsign(value string) {
	v.doIgnore(func() { v.stationCallsign.SetText(value) })
}

func (v *settingsView) SetStationOperator(value string) {
	v.doIgnore(func() { v.stationOperator.SetText(value) })
}

func (v *settingsView) SetStationLocator(value string) {
	v.doIgnore(func() { v.stationLocator.SetText(value) })
}

func (v *settingsView) SetContestIdentifiers(ids []string, texts []string) {
	if len(ids) != len(texts) {
		return
	}
	v.contestIDs = append(v.contestIDs[:0], ids...)
	v.contestTexts = append(v.contestTexts[:0], texts...)
	v.doIgnore(func() {
		v.contestCombo.Clear()
		for _, t := range texts {
			v.contestCombo.AddItem(t)
		}
	})
}

func (v *settingsView) SelectContestIdentifier(value string) {
	v.doIgnore(func() {
		for i, id := range v.contestIDs {
			if id == value {
				v.contestCombo.SetCurrentIndex(i)
				return
			}
		}
		v.contestCombo.SetCurrentIndex(-1)
	})
}

func (v *settingsView) SetContestExchangeFields(fields []core.ExchangeField) {
	v.doIgnore(func() {
		// Wipe existing exchange rows.
		for _, row := range v.exchangeRows {
			v.exchangeGrid.RemoveWidget(row.label.QWidget)
			row.label.QWidget.SetParent(nil)
			row.label.QWidget.DeleteLater()
			v.exchangeGrid.RemoveWidget(row.entry.QWidget)
			row.entry.QWidget.SetParent(nil)
			row.entry.QWidget.DeleteLater()
		}
		v.exchangeRows = v.exchangeRows[:0]

		if v.generateSerialChk != nil {
			v.exchangeGrid.RemoveWidget(v.generateSerialChk.QWidget)
			v.generateSerialChk.QWidget.SetParent(nil)
			v.generateSerialChk.QWidget.DeleteLater()
			v.generateSerialChk = nil
			v.serialExchangeEntry = nil
		}
		if v.generateReportChk != nil {
			v.exchangeGrid.RemoveWidget(v.generateReportChk.QWidget)
			v.generateReportChk.QWidget.SetParent(nil)
			v.generateReportChk.QWidget.DeleteLater()
			v.generateReportChk = nil
			v.reportExchangeEntry = nil
		}

		// Wipe existing call-history field rows.
		for _, row := range v.callHistoryFieldRows {
			v.callHistoryFieldGrid.RemoveWidget(row.label.QWidget)
			row.label.QWidget.SetParent(nil)
			row.label.QWidget.DeleteLater()
			v.callHistoryFieldGrid.RemoveWidget(row.combo.QWidget)
			row.combo.QWidget.SetParent(nil)
			row.combo.QWidget.DeleteLater()
		}
		v.callHistoryFieldRows = v.callHistoryFieldRows[:0]

		// Build new rows.
		for i, field := range fields {
			label := qtlib.NewQLabel3(field.Short)
			entry := qtlib.NewQLineEdit2()
			entry.SetToolTip(field.Short)
			entryField := field.Field
			entry.OnTextChanged(func(text string) {
				if v.ignoreChangedEvent {
					return
				}
				v.controller.EnterContestExchangeValue(entryField, text)
			})
			v.exchangeGrid.AddWidget2(label.QWidget, i, 0)
			v.exchangeGrid.AddWidget2(entry.QWidget, i, 1)
			v.exchangeRows = append(v.exchangeRows, exchangeRow{label: label, entry: entry, field: entryField})

			if field.CanContainSerial && v.generateSerialChk == nil {
				chk := qtlib.NewQCheckBox3("Gen. Serial Number")
				chk.SetToolTip("Check this to automatically generate a serial number for this exchange field.")
				chk.OnToggled(func(checked bool) {
					if v.ignoreChangedEvent {
						return
					}
					if v.serialExchangeEntry != nil {
						v.serialExchangeEntry.SetEnabled(!checked)
					}
					v.controller.EnterContestGenerateSerialExchange(checked)
				})
				v.exchangeGrid.AddWidget2(chk.QWidget, i, 2)
				v.generateSerialChk = chk
				v.serialExchangeEntry = entry
			}
			if field.CanContainReport && v.generateReportChk == nil {
				chk := qtlib.NewQCheckBox3("Gen. Report")
				chk.SetToolTip("Check this to automatically generate a report based on the current mode.")
				chk.OnToggled(func(checked bool) {
					if v.ignoreChangedEvent {
						return
					}
					if v.reportExchangeEntry != nil {
						v.reportExchangeEntry.SetEnabled(!checked)
					}
					v.controller.EnterContestGenerateReport(checked)
				})
				v.exchangeGrid.AddWidget2(chk.QWidget, i, 2)
				v.generateReportChk = chk
				v.reportExchangeEntry = entry
			}

			chLabel := qtlib.NewQLabel3(field.Short)
			chCombo := qtlib.NewQComboBox2()
			chCombo.AddItem("")
			for _, name := range v.availableCallHistoryFieldNames {
				chCombo.AddItem(name)
			}
			chCombo.SetToolTip(field.Short)
			chField := field.Field
			chCombo.OnCurrentTextChanged(func(text string) {
				if v.ignoreChangedEvent {
					return
				}
				v.controller.EnterContestCallHistoryFieldName(chField, text)
			})
			v.callHistoryFieldGrid.AddWidget2(chLabel.QWidget, i, 0)
			v.callHistoryFieldGrid.AddWidget2(chCombo.QWidget, i, 1)
			v.callHistoryFieldRows = append(v.callHistoryFieldRows, callHistoryFieldRow{label: chLabel, combo: chCombo, field: chField})
		}
	})
}

func (v *settingsView) SetContestExchangeValue(index int, value string) {
	i := index - 1
	if i < 0 || i >= len(v.exchangeRows) {
		return
	}
	v.doIgnore(func() { v.exchangeRows[i].entry.SetText(value) })
}

func (v *settingsView) SetContestGenerateSerialExchange(active bool, sensitive bool) {
	if v.generateSerialChk == nil {
		return
	}
	v.doIgnore(func() {
		v.generateSerialChk.SetChecked(active)
		v.generateSerialChk.SetEnabled(sensitive)
		if v.serialExchangeEntry != nil {
			v.serialExchangeEntry.SetEnabled(!active)
		}
	})
}

func (v *settingsView) SetContestGenerateReport(active bool, sensitive bool) {
	if v.generateReportChk == nil {
		return
	}
	v.doIgnore(func() {
		v.generateReportChk.SetChecked(active)
		v.generateReportChk.SetEnabled(sensitive)
		if v.reportExchangeEntry != nil {
			v.reportExchangeEntry.SetEnabled(!active)
		}
	})
}

func (v *settingsView) SetContestEnableQTCs(value bool) {
	v.doIgnore(func() { v.enableQTCs.SetChecked(value) })
}

func (v *settingsView) SetContestName(value string) {
	v.doIgnore(func() { v.contestName.SetText(value) })
}

func (v *settingsView) SetContestStartTime(value string) {
	v.doIgnore(func() { v.contestStartTime.SetText(value) })
}

func (v *settingsView) SetOperationModeSprint(value bool) {
	v.doIgnore(func() { v.sprintMode.SetChecked(value) })
}

func (v *settingsView) SetContestCallHistoryFile(value string) {
	v.doIgnore(func() { v.callHistoryPath.SetText(value) })
}

func (v *settingsView) SetContestCallHistoryFieldName(i int, value string) {
	if i < 0 || i >= len(v.callHistoryFieldRows) {
		return
	}
	v.doIgnore(func() {
		combo := v.callHistoryFieldRows[i].combo
		idx := combo.FindText(value)
		if idx < 0 {
			idx = 0
		}
		combo.SetCurrentIndex(idx)
	})
}

func (v *settingsView) SetContestAvailableCallHistoryFieldNames(names []string) {
	v.availableCallHistoryFieldNames = append(v.availableCallHistoryFieldNames[:0], names...)
}

func (v *settingsView) SetQSOsGoal(value string) {
	v.doIgnore(func() { v.qsosGoal.SetText(value) })
}

func (v *settingsView) SetPointsGoal(value string) {
	v.doIgnore(func() { v.pointsGoal.SetText(value) })
}

func (v *settingsView) SetMultisGoal(value string) {
	v.doIgnore(func() { v.multisGoal.SetText(value) })
}
