package ui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type SettingsController interface {
	Save()
	Reset()
	OpenDefaults()

	EnterStationCallsign(string)
	EnterStationOperator(string)
	EnterStationLocator(string)
	SelectContestIdentifier(string)
	OpenContestRulesPage()
	OpenContestUploadPage()
	EnterContestExchangeValue(core.EntryField, string)
	EnterContestGenerateSerialExchange(bool)

	EnterContestName(string)
	EnterContestEnterTheirNumber(bool)
	EnterContestEnterTheirXchange(bool)
	EnterContestRequireTheirXchange(bool)
	EnterContestAllowMultiBand(bool)
	EnterContestAllowMultiMode(bool)
	EnterContestSameCountryPoints(string)
	EnterContestSameContinentPoints(string)
	EnterContestSpecificCountryPoints(string)
	EnterContestSpecificCountryPrefixes(string)
	EnterContestOtherPoints(string)
	EnterContestMultis(dxcc, wpx, xchange bool)
	EnterContestXchangeMultiPattern(string)
	EnterContestTestXchangeValue(string)
	EnterContestCountPerBand(bool)
	EnterContestCallHistoryFile(string)
	EnterContestCallHistoryFieldName(core.EntryField, string)
	EnterContestCabrilloQSOTemplate(string)
}

type fieldID string

const (
	stationCallsign                fieldID = "stationCallsign"
	stationOperator                fieldID = "stationOperator"
	stationLocator                 fieldID = "stationLocator"
	contestIdentifier              fieldID = "contestIdentifier"
	contestName                    fieldID = "contestName"
	contestEnterTheirNumber        fieldID = "contestEnterTheirNumber"
	contestEnterTheirXchange       fieldID = "contestEnterTheirXchange"
	contestRequireTheirXchange     fieldID = "contestRequireTheirXchange"
	contestAllowMultiBand          fieldID = "contestAllowMultiBand"
	contestAllowMultiMode          fieldID = "contestAllowMultiMode"
	contestSameCountryPoints       fieldID = "contestSameCountryPoints"
	contestSameContinentPoints     fieldID = "contestSameContinentPoints"
	contestSpecificCountryPoints   fieldID = "contestSpecificCountryPoints"
	contestSpecificCountryPrefixes fieldID = "contestSpecificCountryPrefixes"
	contestOtherPoints             fieldID = "contestOtherPoints"
	contestMultiDXCC               fieldID = "contestMultiDXCC"
	contestMultiWPX                fieldID = "contestMultiWPX"
	contestMultiXchange            fieldID = "contestMultiXchange"
	contestXchangeMultiPattern     fieldID = "contestXchangeMultiPattern"
	contestTestXchangeMultiPattern fieldID = "contestTestXchangeMultiPattern"
	contestCountPerBand            fieldID = "contestCountPerBand"
	contestCallHistoryFile         fieldID = "contestCallHistoryFile"
	contestCabrilloQSOTemplate     fieldID = "contestCabrilloQSOTemplate"
)

type settingsView struct {
	parent     *gtk.Dialog
	controller SettingsController

	ignoreChangedEvent bool

	message               *gtk.Label
	openContestRulesPage  *gtk.Button
	openContestUploadPage *gtk.Button
	openDefaults          *gtk.Button
	reset                 *gtk.Button
	close                 *gtk.Button

	xchangeMultiValue *gtk.Label

	fields                       map[fieldID]interface{}
	exchangeFieldsParent         *gtk.Grid
	exchangeFieldCount           int
	generateSerialExchangeButton *gtk.CheckButton
	serialExchangeEntry          *gtk.Entry
	callHistoryFieldNamesParent  *gtk.Grid
}

func setupSettingsView(builder *gtk.Builder, parent *gtk.Dialog, controller SettingsController) *settingsView {
	result := new(settingsView)
	result.parent = parent
	result.controller = controller
	result.fields = make(map[fieldID]interface{})

	result.message = getUI(builder, "settingsMessageLabel").(*gtk.Label)
	result.xchangeMultiValue = getUI(builder, "xchangeMultiValueLabel").(*gtk.Label)
	result.openContestRulesPage = getUI(builder, "openContestRulesPageButton").(*gtk.Button)
	result.openContestRulesPage.Connect("clicked", result.onOpenContestRulesPagePressed)
	result.openContestUploadPage = getUI(builder, "openContestUploadPageButton").(*gtk.Button)
	result.openContestUploadPage.Connect("clicked", result.onOpenContestUploadPagePressed)
	result.exchangeFieldsParent = getUI(builder, "contestExchangeFieldsGrid").(*gtk.Grid)
	result.callHistoryFieldNamesParent = getUI(builder, "contestCallHistoryFieldNamesGrid").(*gtk.Grid)

	result.openDefaults = getUI(builder, "openDefaultsButton").(*gtk.Button)
	result.openDefaults.Connect("clicked", result.onOpenDefaultsPressed)
	result.reset = getUI(builder, "resetButton").(*gtk.Button)
	result.reset.Connect("clicked", result.onResetPressed)
	result.close = getUI(builder, "closeButton").(*gtk.Button)
	result.close.Connect("clicked", result.onClosePressed)

	result.addEntry(builder, stationCallsign)
	result.addEntry(builder, stationOperator)
	result.addEntry(builder, stationLocator)
	result.addCombo(builder, contestIdentifier)
	result.addEntry(builder, contestName)
	result.addCheckButton(builder, contestEnterTheirNumber)
	result.addCheckButton(builder, contestEnterTheirXchange)
	result.addCheckButton(builder, contestRequireTheirXchange)
	result.addCheckButton(builder, contestAllowMultiBand)
	result.addCheckButton(builder, contestAllowMultiMode)
	result.addEntry(builder, contestSameCountryPoints)
	result.addEntry(builder, contestSameContinentPoints)
	result.addEntry(builder, contestSpecificCountryPoints)
	result.addEntry(builder, contestSpecificCountryPrefixes)
	result.addEntry(builder, contestOtherPoints)
	result.addCheckButton(builder, contestMultiDXCC)
	result.addCheckButton(builder, contestMultiWPX)
	result.addCheckButton(builder, contestMultiXchange)
	result.addEntry(builder, contestXchangeMultiPattern)
	result.addEntry(builder, contestTestXchangeMultiPattern)
	result.addCheckButton(builder, contestCountPerBand)
	result.addFileChooser(builder, contestCallHistoryFile)
	result.addEntry(builder, contestCabrilloQSOTemplate)

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

func (v *settingsView) addCombo(builder *gtk.Builder, id fieldID) {
	entry := getUI(builder, string(id)+"Combo").(*gtk.ComboBoxText)
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
	case contestEnterTheirNumber:
		v.controller.EnterContestEnterTheirNumber(value.(bool))
	case contestEnterTheirXchange:
		v.controller.EnterContestEnterTheirXchange(value.(bool))
	case contestRequireTheirXchange:
		v.controller.EnterContestRequireTheirXchange(value.(bool))
	case contestAllowMultiBand:
		v.controller.EnterContestAllowMultiBand(value.(bool))
	case contestAllowMultiMode:
		v.controller.EnterContestAllowMultiMode(value.(bool))
	case contestSameCountryPoints:
		v.controller.EnterContestSameCountryPoints(value.(string))
	case contestSameContinentPoints:
		v.controller.EnterContestSameContinentPoints(value.(string))
	case contestSpecificCountryPoints:
		v.controller.EnterContestSpecificCountryPoints(value.(string))
	case contestSpecificCountryPrefixes:
		v.controller.EnterContestSpecificCountryPrefixes(value.(string))
	case contestOtherPoints:
		v.controller.EnterContestOtherPoints(value.(string))
	case contestMultiDXCC, contestMultiWPX, contestMultiXchange:
		v.controller.EnterContestMultis(v.multis())
	case contestXchangeMultiPattern:
		v.controller.EnterContestXchangeMultiPattern(value.(string))
	case contestTestXchangeMultiPattern:
		v.controller.EnterContestTestXchangeValue(value.(string))
	case contestCountPerBand:
		v.controller.EnterContestCountPerBand(value.(bool))
	case contestCallHistoryFile:
		v.controller.EnterContestCallHistoryFile(value.(string))
	case contestCabrilloQSOTemplate:
		v.controller.EnterContestCabrilloQSOTemplate(value.(string))
	default:
		log.Printf("enter unknown field %s: %v", field, value)
	}

	return false
}

func (v *settingsView) multis() (dxcc, wpx, xchange bool) {
	dxcc = v.fields[contestMultiDXCC].(*gtk.CheckButton).GetActive()
	wpx = v.fields[contestMultiWPX].(*gtk.CheckButton).GetActive()
	xchange = v.fields[contestMultiXchange].(*gtk.CheckButton).GetActive()
	return
}

func (v *settingsView) onOpenContestRulesPagePressed(_ *gtk.Button) {
	v.controller.OpenContestRulesPage()
}

func (v *settingsView) onOpenContestUploadPagePressed(_ *gtk.Button) {
	v.controller.OpenContestUploadPage()
}

func (v *settingsView) onOpenDefaultsPressed(_ *gtk.Button) {
	v.controller.OpenDefaults()
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

func (v *settingsView) selectComboField(field fieldID, value string) {
	v.doIgnoreChanges(func() {
		v.fields[field].(*gtk.ComboBoxText).SetActiveID(value)
	})
}

func (v *settingsView) setCheckButtonField(field fieldID, value bool) {
	v.doIgnoreChanges(func() {
		v.fields[field].(*gtk.CheckButton).SetActive(value)
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
		fieldName, _ := gtk.EntryNew()
		fieldName.SetName(string(field.Field))
		fieldName.SetWidthChars(6)
		fieldName.SetTooltipText(field.Short) // TODO use field.Hint
		fieldName.SetHAlign(gtk.ALIGN_FILL)
		fieldName.SetHExpand(true)
		fieldName.Connect("changed", v.onCallHistoryFieldNameChanged)
		v.callHistoryFieldNamesParent.Attach(fieldName, i, 0, 1, 1)

		if !field.CanContainSerial || v.generateSerialExchangeButton != nil {
			continue
		}

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

func (v *settingsView) onCallHistoryFieldNameChanged(entry *gtk.Entry) bool {
	if v.ignoreChangedEvent {
		return false
	}

	name, _ := entry.GetName()
	entryField := core.EntryField(name)

	value, _ := entry.GetText()

	v.controller.EnterContestCallHistoryFieldName(entryField, value)

	return false
}

func (v *settingsView) SetContestCallHistoryFieldName(i int, value string) {
	child, _ := v.callHistoryFieldNamesParent.GetChildAt(i, 0)
	entry, ok := child.(*gtk.Entry)
	if !ok {
		return
	}

	v.doIgnoreChanges(func() {
		entry.SetText(value)
	})
}

func (v *settingsView) SetContestName(value string) {
	v.setEntryField(contestName, value)
}

func (v *settingsView) SetContestEnterTheirNumber(value bool) {
	v.setCheckButtonField(contestEnterTheirNumber, value)
}

func (v *settingsView) SetContestEnterTheirXchange(value bool) {
	v.setCheckButtonField(contestEnterTheirXchange, value)
}

func (v *settingsView) SetContestRequireTheirXchange(value bool) {
	v.setCheckButtonField(contestRequireTheirXchange, value)
}

func (v *settingsView) SetContestAllowMultiBand(value bool) {
	v.setCheckButtonField(contestAllowMultiBand, value)
}

func (v *settingsView) SetContestAllowMultiMode(value bool) {
	v.setCheckButtonField(contestAllowMultiMode, value)
}

func (v *settingsView) SetContestSameCountryPoints(value string) {
	v.setEntryField(contestSameCountryPoints, value)
}

func (v *settingsView) SetContestSameContinentPoints(value string) {
	v.setEntryField(contestSameContinentPoints, value)
}

func (v *settingsView) SetContestSpecificCountryPoints(value string) {
	v.setEntryField(contestSpecificCountryPoints, value)
}

func (v *settingsView) SetContestSpecificCountryPrefixes(value string) {
	v.setEntryField(contestSpecificCountryPrefixes, value)
}

func (v *settingsView) SetContestOtherPoints(value string) {
	v.setEntryField(contestOtherPoints, value)
}

func (v *settingsView) SetContestMultis(dxcc, wpx, xchange bool) {
	v.setCheckButtonField(contestMultiDXCC, dxcc)
	v.setCheckButtonField(contestMultiWPX, wpx)
	v.setCheckButtonField(contestMultiXchange, xchange)
}

func (v *settingsView) SetContestXchangeMultiPattern(value string) {
	v.setEntryField(contestXchangeMultiPattern, value)
}

func (v *settingsView) SetContestXchangeMultiPatternResult(value string) {
	v.xchangeMultiValue.SetText(value)
}

func (v *settingsView) SetContestCountPerBand(value bool) {
	v.setCheckButtonField(contestCountPerBand, value)
}

func (v *settingsView) SetContestCallHistoryFile(value string) {
	v.setFileChooserField(contestCallHistoryFile, value)
}

func (v *settingsView) SetContestCabrilloQSOTemplate(value string) {
	v.setEntryField(contestCabrilloQSOTemplate, value)
}
