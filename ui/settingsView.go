package ui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

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
	ClearCallHistory()

	EnterContestExchangeValue(core.EntryField, string)
	EnterContestGenerateSerialExchange(bool)
	EnterContestGenerateReport(bool)

	EnterContestName(string)
	EnterContestStartTime(string)
	SetContestStartTimeToday()
	SetContestStartTimeNow()
	SetOperationModeSprint(bool)
	EnterContestCallHistoryFile(string)
	EnterContestCallHistoryFieldName(core.EntryField, string)
	EnterQSOsGoal(string)
	EnterPointsGoal(string)
	EnterMultisGoal(string)
}

type fieldID string

const (
	stationCallsign        fieldID = "stationCallsign"
	stationOperator        fieldID = "stationOperator"
	stationLocator         fieldID = "stationLocator"
	contestIdentifier      fieldID = "contestIdentifier"
	contestName            fieldID = "contestName"
	contestStartTime       fieldID = "contestStartTime"
	operationModeSprint    fieldID = "operationModeSprint"
	contestCallHistoryFile fieldID = "contestCallHistoryFile"
	qsosGoal               fieldID = "qsosGoal"
	pointsGoal             fieldID = "pointsGoal"
	multisGoal             fieldID = "multisGoal"
)

type settingsView struct {
	parent     *gtk.Dialog
	controller SettingsController

	ignoreChangedEvent bool

	message               *gtk.Label
	openContestRulesPage  *gtk.Button
	openContestUploadPage *gtk.Button
	openConfigurationFile *gtk.Button
	reset                 *gtk.Button
	close                 *gtk.Button

	fields map[fieldID]interface{}

	exchangeFieldsParent           *gtk.Grid
	exchangeFieldCount             int
	generateSerialExchangeButton   *gtk.CheckButton
	generateReportButton           *gtk.CheckButton
	serialExchangeEntry            *gtk.Entry
	reportEntry                    *gtk.Entry
	contestStartTimeTodayButton    *gtk.Button
	contestStartTimeNowButton      *gtk.Button
	callHistoryFieldNamesParent    *gtk.Grid
	clearCallHistorySettingsButton *gtk.Button
	availableCallHistoryFieldNames []string
}

func setupSettingsView(builder *gtk.Builder, parent *gtk.Dialog, controller SettingsController) *settingsView {
	result := new(settingsView)
	result.parent = parent
	result.controller = controller
	result.fields = make(map[fieldID]interface{})

	result.message = getUI(builder, "settingsMessageLabel").(*gtk.Label)
	result.openContestRulesPage = getUI(builder, "openContestRulesPageButton").(*gtk.Button)
	result.openContestRulesPage.Connect("clicked", result.onOpenContestRulesPagePressed)
	result.openContestUploadPage = getUI(builder, "openContestUploadPageButton").(*gtk.Button)
	result.openContestUploadPage.Connect("clicked", result.onOpenContestUploadPagePressed)
	result.exchangeFieldsParent = getUI(builder, "contestExchangeFieldsGrid").(*gtk.Grid)
	result.contestStartTimeTodayButton = getUI(builder, "contestStartTimeTodayButton").(*gtk.Button)
	result.contestStartTimeTodayButton.Connect("clicked", result.onContestStartTimeTodayPressed)
	result.contestStartTimeNowButton = getUI(builder, "contestStartTimeNowButton").(*gtk.Button)
	result.contestStartTimeNowButton.Connect("clicked", result.onContestStartTimeNowPressed)
	result.callHistoryFieldNamesParent = getUI(builder, "contestCallHistoryFieldNamesGrid").(*gtk.Grid)
	result.clearCallHistorySettingsButton = getUI(builder, "contestCallHistoryClearButton").(*gtk.Button)
	result.clearCallHistorySettingsButton.Connect("clicked", result.onClearCallHistoryPressed)

	result.openConfigurationFile = getUI(builder, "openConfigurationButton").(*gtk.Button)
	result.openConfigurationFile.Connect("clicked", result.onOpenConfigurationFilePressed)
	result.reset = getUI(builder, "resetButton").(*gtk.Button)
	result.reset.Connect("clicked", result.onResetPressed)
	result.close = getUI(builder, "closeButton").(*gtk.Button)
	result.close.Connect("clicked", result.onClosePressed)

	result.addEntry(builder, stationCallsign)
	result.addEntry(builder, stationOperator)
	result.addEntry(builder, stationLocator)
	result.addCombo(builder, contestIdentifier)
	result.addEntry(builder, contestName)
	result.addEntry(builder, contestStartTime)
	result.addCheckButton(builder, operationModeSprint)
	result.addFileChooser(builder, contestCallHistoryFile)
	result.addEntry(builder, qsosGoal)
	result.addEntry(builder, pointsGoal)
	result.addEntry(builder, multisGoal)

	result.parent.Connect("destroy", result.onDestroy)

	return result
}

func (v *settingsView) SetSettingsController(controller SettingsController) {
	v.controller = controller
}
func (v *settingsView) ShowMessage(message string) {
	v.message.SetText(message)
	v.message.Show()
}

func (v *settingsView) HideMessage() {
	v.message.Hide()
}

func (v *settingsView) Ready() bool {
	return v != nil
}

func (v *settingsView) addEntry(builder *gtk.Builder, id fieldID) {
	entry := getUI(builder, string(id)+"Entry").(*gtk.Entry)
	field, _ := entry.GetName()
	v.fields[fieldID(field)] = entry

	widget := &entry.Widget
	widget.Connect("changed", v.onFieldChanged)
}

func (v *settingsView) addCheckButton(builder *gtk.Builder, id fieldID) {
	button := getUI(builder, string(id)+"Button").(*gtk.CheckButton)
	field, _ := button.GetName()
	v.fields[fieldID(field)] = button

	widget := &button.Widget
	widget.Connect("toggled", v.onFieldChanged)
}

func (v *settingsView) addCombo(builder *gtk.Builder, id fieldID) {
	entry := getUI(builder, string(id)+"Combo").(*gtk.ComboBoxText)
	field, _ := entry.GetName()
	v.fields[fieldID(field)] = entry

	widget := &entry.Widget
	widget.Connect("changed", v.onFieldChanged)
}

func (v *settingsView) addFileChooser(builder *gtk.Builder, id fieldID) {
	button := getUI(builder, string(id)+"Chooser").(*gtk.FileChooserButton)
	field, _ := button.GetName()
	v.fields[fieldID(field)] = button

	widget := &button.Widget
	widget.Connect("file-set", v.onFieldChanged)
}

func (v *settingsView) onFieldChanged(w any) bool {
	if v.ignoreChangedEvent {
		return false
	}

	var field string
	var value interface{}
	switch widget := w.(type) {
	case *gtk.Entry:
		field, _ = widget.GetName()
		value, _ = widget.GetText()
	case *gtk.ComboBoxText:
		field, _ = widget.GetName()
		value = widget.GetActiveID()
	case *gtk.CheckButton:
		field, _ = widget.GetName()
		value = widget.GetActive()
	case *gtk.FileChooserButton:
		field, _ = widget.GetName()
		value = widget.GetFilename()
	default:
		return false
	}

	switch fieldID(field) {
	case stationCallsign:
		v.controller.EnterStationCallsign(value.(string))
	case stationOperator:
		v.controller.EnterStationOperator(value.(string))
	case stationLocator:
		v.controller.EnterStationLocator(value.(string))
	case contestIdentifier:
		v.controller.SelectContestIdentifier(value.(string))
	case contestName:
		v.controller.EnterContestName(value.(string))
	case contestStartTime:
		v.controller.EnterContestStartTime(value.(string))
	case operationModeSprint:
		v.controller.SetOperationModeSprint(value.(bool))
	case contestCallHistoryFile:
		v.controller.EnterContestCallHistoryFile(value.(string))
	case qsosGoal:
		v.controller.EnterQSOsGoal(value.(string))
	case pointsGoal:
		v.controller.EnterPointsGoal(value.(string))
	case multisGoal:
		v.controller.EnterMultisGoal(value.(string))
	default:
		log.Printf("enter unknown field %s: %v", field, value)
	}

	return false
}

func (v *settingsView) onOpenContestRulesPagePressed(_ *gtk.Button) {
	v.controller.OpenContestRulesPage()
}

func (v *settingsView) onOpenContestUploadPagePressed(_ *gtk.Button) {
	v.controller.OpenContestUploadPage()
}

func (v *settingsView) onContestStartTimeTodayPressed(_ *gtk.Button) {
	v.controller.SetContestStartTimeToday()
}

func (v *settingsView) onContestStartTimeNowPressed(_ *gtk.Button) {
	v.controller.SetContestStartTimeNow()
}

func (v *settingsView) onClearCallHistoryPressed(_ *gtk.Button) {
	v.controller.ClearCallHistory()
}

func (v *settingsView) onOpenConfigurationFilePressed(_ *gtk.Button) {
	v.controller.OpenConfigurationFile()
}

func (v *settingsView) onResetPressed(_ *gtk.Button) {
	v.controller.Reset()
}

func (v *settingsView) onClosePressed(_ *gtk.Button) {
	v.parent.Close()
}

func (v *settingsView) onDestroy() {
	v.controller.Save()
}

func (v *settingsView) setEntryField(field fieldID, value string) {
	v.doIgnoreChanges(func() {
		v.fields[field].(*gtk.Entry).SetText(value)
	})
}

func (v *settingsView) setCheckboxField(field fieldID, value bool) {
	v.doIgnoreChanges(func() {
		v.fields[field].(*gtk.CheckButton).SetActive(value)
	})
}

func (v *settingsView) selectComboField(field fieldID, value string) {
	v.doIgnoreChanges(func() {
		v.fields[field].(*gtk.ComboBoxText).SetActiveID(value)
	})
}

func (v *settingsView) setFileChooserField(field fieldID, value string) {
	v.doIgnoreChanges(func() {
		v.fields[field].(*gtk.FileChooserButton).SetFilename(value)
	})
}

func (v *settingsView) doIgnoreChanges(f func()) {
	if v == nil {
		return
	}

	v.ignoreChangedEvent = true
	defer func() {
		v.ignoreChangedEvent = false
	}()
	f()
}

func (v *settingsView) SetStationCallsign(value string) {
	v.setEntryField(stationCallsign, value)
}

func (v *settingsView) SetStationOperator(value string) {
	v.setEntryField(stationOperator, value)
}

func (v *settingsView) SetStationLocator(value string) {
	v.setEntryField(stationLocator, value)
}

func (v *settingsView) SetContestIdentifiers(ids []string, texts []string) {
	if len(ids) != len(texts) {
		panic("contest identifiers and names are not in sync")
	}

	v.doIgnoreChanges(func() {
		combo := v.fields[contestIdentifier].(*gtk.ComboBoxText)
		combo.RemoveAll()
		for i, value := range ids {
			combo.Append(value, texts[i])
		}
		combo.SetActive(0)
	})
}

func (v *settingsView) SetContestPagesAvailable(rulesPageAvailable bool, uploadPageAvailable bool) {
	v.openContestRulesPage.SetSensitive(rulesPageAvailable)
	v.openContestUploadPage.SetSensitive(uploadPageAvailable)
}

func (v *settingsView) SelectContestIdentifier(value string) {
	v.selectComboField(contestIdentifier, value)
}

func (v *settingsView) SetContestExchangeFields(fields []core.ExchangeField) {
	for i := 0; i < v.exchangeFieldCount; i++ {
		label, _ := v.exchangeFieldsParent.GetChildAt(0, 0)
		if label != nil {
			label.ToWidget().Destroy()
		}
		entry, _ := v.exchangeFieldsParent.GetChildAt(1, 0)
		if entry != nil {
			entry.ToWidget().Destroy()
		}
		v.exchangeFieldsParent.RemoveRow(0)

		fieldName, _ := v.callHistoryFieldNamesParent.GetChildAt(0, 0)
		if fieldName != nil {
			fieldName.ToWidget().Destroy()
		}
		v.callHistoryFieldNamesParent.RemoveColumn(0)
	}
	if v.generateSerialExchangeButton != nil {
		v.generateSerialExchangeButton.Destroy()
		v.generateSerialExchangeButton = nil
		v.serialExchangeEntry = nil
	}
	if v.generateReportButton != nil {
		v.generateReportButton.Destroy()
		v.generateReportButton = nil
		v.reportEntry = nil
	}

	for i, field := range fields {
		v.exchangeFieldsParent.InsertRow(i)
		label, _ := gtk.LabelNew(field.Short)
		label.SetHAlign(gtk.ALIGN_START)
		label.SetHExpand(false)
		v.exchangeFieldsParent.Attach(label, 0, i, 1, 1)

		entry, _ := gtk.EntryNew()
		entry.SetName(string(field.Field))
		entry.SetWidthChars(4)
		entry.SetTooltipText(field.Short) // TODO use field.Hint
		entry.SetHAlign(gtk.ALIGN_FILL)
		entry.SetHExpand(false)
		entry.Connect("changed", v.onExchangeFieldChanged)
		v.exchangeFieldsParent.Attach(entry, 1, i, 1, 1)

		v.callHistoryFieldNamesParent.InsertColumn(i)
		fieldName, _ := gtk.ComboBoxTextNew()
		fieldName.SetName(string(field.Field))
		fieldName.Append("", "")
		for _, t := range v.availableCallHistoryFieldNames {
			fieldName.Append(t, t)
		}
		fieldName.SetTooltipText(field.Short) // TODO use field.Hint
		fieldName.SetHAlign(gtk.ALIGN_FILL)
		fieldName.SetHExpand(true)
		fieldName.Connect("changed", v.onCallHistoryFieldNameChanged)
		v.callHistoryFieldNamesParent.Attach(fieldName, i, 0, 1, 1)

		if field.CanContainSerial && v.generateSerialExchangeButton == nil {
			serialCheckButton, _ := gtk.CheckButtonNew()
			serialCheckButton.SetLabel("Gen. Serial Number")
			serialCheckButton.SetTooltipText("Check this if you want to automatically generate a serial number as your exchange for this field.")
			serialCheckButton.SetHAlign(gtk.ALIGN_START)
			serialCheckButton.SetHExpand(true)
			serialCheckButton.Connect("toggled", v.onGenerateSerialExchangeChanged)
			v.exchangeFieldsParent.Attach(serialCheckButton, 2, i, 1, 1)
			v.generateSerialExchangeButton = serialCheckButton
			v.serialExchangeEntry = entry
		}

		if field.CanContainReport && v.generateReportButton == nil {
			reportCheckButton, _ := gtk.CheckButtonNew()
			reportCheckButton.SetLabel("Gen. Report")
			reportCheckButton.SetTooltipText("Check this if you want to automatically generate a report based on the currently selected mode.")
			reportCheckButton.SetHAlign(gtk.ALIGN_START)
			reportCheckButton.SetHExpand(true)
			reportCheckButton.Connect("toggled", v.onGenerateReportChanged)
			v.exchangeFieldsParent.Attach(reportCheckButton, 2, i, 1, 1)
			v.generateReportButton = reportCheckButton
			v.reportEntry = entry
		}
	}

	v.exchangeFieldsParent.ShowAll()
	v.callHistoryFieldNamesParent.ShowAll()
	v.exchangeFieldCount = len(fields)
}

func (v *settingsView) onExchangeFieldChanged(entry *gtk.Entry) bool {
	if v.ignoreChangedEvent {
		return false
	}

	name, _ := entry.GetName()
	entryField := core.EntryField(name)

	value, _ := entry.GetText()

	v.controller.EnterContestExchangeValue(entryField, value)

	return false
}

func (v *settingsView) onGenerateSerialExchangeChanged(checkButton *gtk.CheckButton) bool {
	if v.ignoreChangedEvent {
		return false
	}

	value := checkButton.GetActive()
	v.serialExchangeEntry.SetSensitive(!value)
	v.controller.EnterContestGenerateSerialExchange(value)

	return false
}

func (v *settingsView) onGenerateReportChanged(checkButton *gtk.CheckButton) bool {
	if v.ignoreChangedEvent {
		return false
	}

	value := checkButton.GetActive()
	v.reportEntry.SetSensitive(!value)
	v.controller.EnterContestGenerateReport(value)

	return false
}

func (v *settingsView) SetContestExchangeValue(index int, value string) {
	child, _ := v.exchangeFieldsParent.GetChildAt(1, index-1)
	entry, ok := child.(*gtk.Entry)
	if !ok {
		return
	}

	v.doIgnoreChanges(func() {
		entry.SetText(value)
	})
}

func (v *settingsView) SetContestGenerateSerialExchange(active bool, sensitive bool) {
	if v.generateSerialExchangeButton == nil {
		return
	}

	v.doIgnoreChanges(func() {
		v.generateSerialExchangeButton.SetActive(active)
		v.generateSerialExchangeButton.SetSensitive(sensitive)
		v.serialExchangeEntry.SetSensitive(!active)
	})
}

func (v *settingsView) SetContestGenerateReport(active bool, sensitive bool) {
	if v.generateReportButton == nil {
		return
	}

	v.doIgnoreChanges(func() {
		v.generateReportButton.SetActive(active)
		v.generateReportButton.SetSensitive(sensitive)
		v.reportEntry.SetSensitive(!active)
	})
}

func (v *settingsView) onCallHistoryFieldNameChanged(entry *gtk.ComboBoxText) bool {
	if v.ignoreChangedEvent {
		return false
	}

	name, _ := entry.GetName()
	entryField := core.EntryField(name)

	value := entry.GetActiveText()

	v.controller.EnterContestCallHistoryFieldName(entryField, value)

	return false
}

func (v *settingsView) SetContestCallHistoryFieldName(i int, value string) {
	child, _ := v.callHistoryFieldNamesParent.GetChildAt(i, 0)
	entry, ok := child.(*gtk.ComboBoxText)
	if !ok {
		return
	}

	v.doIgnoreChanges(func() {
		entry.SetActiveID(value)
	})
}

func (v *settingsView) SetContestAvailableCallHistoryFieldNames(fieldNames []string) {
	v.availableCallHistoryFieldNames = fieldNames
}

func (v *settingsView) SetContestName(value string) {
	v.setEntryField(contestName, value)
}

func (v *settingsView) SetContestStartTime(value string) {
	v.setEntryField(contestStartTime, value)
}

func (v *settingsView) SetOperationModeSprint(value bool) {
	v.setCheckboxField(operationModeSprint, value)
}

func (v *settingsView) SetContestCallHistoryFile(value string) {
	v.setFileChooserField(contestCallHistoryFile, value)
}

func (v *settingsView) SetQSOsGoal(value string) {
	v.setEntryField(qsosGoal, value)
}

func (v *settingsView) SetPointsGoal(value string) {
	v.setEntryField(pointsGoal, value)
}

func (v *settingsView) SetMultisGoal(value string) {
	v.setEntryField(multisGoal, value)
}
