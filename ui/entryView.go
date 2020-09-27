package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

// EntryController controls the entry of QSO data.
type EntryController interface {
	GotoNextField() core.EntryField
	SetActiveField(core.EntryField)

	BandSelected(string)
	ModeSelected(string)
	EnterCallsign(string)
	SendQuestion()

	Log()
	Reset()
}

type entryView struct {
	controller EntryController

	style             *style
	ignoreComboChange bool

	entryRoot    *gtk.Grid
	callsign     *gtk.Entry
	theirReport  *gtk.Entry
	theirNumber  *gtk.Entry
	theirXchange *gtk.Entry
	band         *gtk.ComboBoxText
	mode         *gtk.ComboBoxText
	myReport     *gtk.Entry
	myNumber     *gtk.Entry
	myXchange    *gtk.Entry
	logButton    *gtk.Button
	resetButton  *gtk.Button
	messageLabel *gtk.Label
}

func setupEntryView(builder *gtk.Builder) *entryView {
	result := new(entryView)

	result.entryRoot = getUI(builder, "entryGrid").(*gtk.Grid)
	result.callsign = getUI(builder, "callsignEntry").(*gtk.Entry)
	result.theirReport = getUI(builder, "theirReportEntry").(*gtk.Entry)
	result.theirNumber = getUI(builder, "theirNumberEntry").(*gtk.Entry)
	result.theirXchange = getUI(builder, "theirXchangeEntry").(*gtk.Entry)
	result.band = getUI(builder, "bandCombo").(*gtk.ComboBoxText)
	result.mode = getUI(builder, "modeCombo").(*gtk.ComboBoxText)
	result.myReport = getUI(builder, "myReportEntry").(*gtk.Entry)
	result.myNumber = getUI(builder, "myNumberEntry").(*gtk.Entry)
	result.myXchange = getUI(builder, "myXchangeEntry").(*gtk.Entry)
	result.logButton = getUI(builder, "logButton").(*gtk.Button)
	result.resetButton = getUI(builder, "resetButton").(*gtk.Button)
	result.messageLabel = getUI(builder, "messageLabel").(*gtk.Label)

	result.addEntryTraversal(result.callsign)
	result.addEntryTraversal(result.theirReport)
	result.addEntryTraversal(result.theirNumber)
	result.addEntryTraversal(result.theirXchange)
	result.addEntryTraversal(result.myReport)
	result.addEntryTraversal(result.myNumber)
	result.addEntryTraversal(result.myXchange)
	result.addOtherWidgetTraversal(&result.band.Widget)
	result.addOtherWidgetTraversal(&result.mode.Widget)

	result.callsign.Connect("changed", result.onCallsignChanged)
	result.band.Connect("changed", result.onBandChanged)
	result.mode.Connect("changed", result.onModeChanged)
	result.logButton.Connect("clicked", result.onLogButtonClicked)
	result.resetButton.Connect("clicked", result.onResetButtonClicked)

	setupBandCombo(result.band)
	setupModeCombo(result.mode)

	result.style = newStyle(`
	.duplicate {
		background-color: #FF0000; 
		color: #FFFFFF;
	}
	.editing {
		background-color: #66AAFF;
	}
	`)
	result.style.applyTo(&result.entryRoot.Widget)

	return result
}

func setupBandCombo(combo *gtk.ComboBoxText) {
	combo.RemoveAll()
	for _, value := range core.Bands {
		combo.Append(value.String(), value.String())
	}
	combo.SetActive(0)
}

func setupModeCombo(combo *gtk.ComboBoxText) {
	combo.RemoveAll()
	for _, value := range core.Modes {
		combo.Append(value.String(), value.String())
	}
	combo.SetActive(0)
}

func (v *entryView) SetEntryController(controller EntryController) {
	v.controller = controller
}

func (v *entryView) addEntryTraversal(entry *gtk.Entry) {
	entry.Connect("key_press_event", v.onEntryKeyPress)
	entry.Connect("focus_in_event", v.onEntryFocusIn)
	entry.Connect("focus_out_event", v.onEntryFocusOut)
}

func (v *entryView) onEntryKeyPress(widget interface{}, event *gdk.Event) bool {
	keyEvent := gdk.EventKeyNewFromEvent(event)
	switch keyEvent.KeyVal() {
	case gdk.KEY_Tab:
		v.controller.GotoNextField()
		return true
	case gdk.KEY_space:
		v.controller.GotoNextField()
		return true
	case gdk.KEY_Return:
		v.controller.Log()
		return true
	case gdk.KEY_Escape:
		v.controller.Reset()
		return true
	case gdk.KEY_question:
		v.controller.SendQuestion()
		return true
	default:
		return false
	}
}

func (v *entryView) onEntryFocusIn(entry *gtk.Entry, event *gdk.Event) bool {
	entryField := v.entryToField(entry)
	v.controller.SetActiveField(entryField)
	return false
}

func (v *entryView) onEntryFocusOut(entry *gtk.Entry, event *gdk.Event) bool {
	entry.SelectRegion(0, 0)
	return false
}

func (v *entryView) addOtherWidgetTraversal(widget *gtk.Widget) {
	widget.Connect("key_press_event", v.onOtherWidgetKeyPress)
}

func (v *entryView) onOtherWidgetKeyPress(widget interface{}, event *gdk.Event) bool {
	keyEvent := gdk.EventKeyNewFromEvent(event)
	switch keyEvent.KeyVal() {
	case gdk.KEY_Tab:
		v.controller.SetActiveField(core.CallsignField)
		v.SetActiveField(core.CallsignField)
		return true
	case gdk.KEY_space:
		v.controller.SetActiveField(core.CallsignField)
		v.SetActiveField(core.CallsignField)
		return true
	default:
		return false
	}
}

func (v *entryView) onCallsignChanged(widget *gtk.Entry) bool {
	if v.controller == nil {
		return false
	}
	callsign, err := widget.GetText()
	if err != nil {
		return false
	}
	v.controller.EnterCallsign(callsign)
	return false
}

func (v *entryView) onBandChanged(widget *gtk.ComboBoxText) bool {
	if v.controller != nil && !v.ignoreComboChange {
		v.controller.BandSelected(widget.GetActiveText())
	}
	return false
}

func (v *entryView) onModeChanged(widget *gtk.ComboBoxText) bool {
	if v.controller != nil && !v.ignoreComboChange {
		v.controller.ModeSelected(widget.GetActiveText())
	}
	return false
}

func (v *entryView) onLogButtonClicked(button *gtk.Button) bool {
	v.controller.Log()
	return true
}

func (v *entryView) onResetButtonClicked(button *gtk.Button) bool {
	v.controller.Reset()
	return true
}

func (v *entryView) Callsign() string {
	text, err := v.callsign.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (v *entryView) SetCallsign(text string) {
	v.callsign.SetText(text)
}

func (v *entryView) TheirReport() string {
	text, err := v.theirReport.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (v *entryView) SetTheirReport(text string) {
	v.theirReport.SetText(text)
}

func (v *entryView) TheirNumber() string {
	text, err := v.theirNumber.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (v *entryView) SetTheirNumber(text string) {
	v.theirNumber.SetText(text)
}

func (v *entryView) TheirXchange() string {
	text, err := v.theirXchange.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (v *entryView) SetTheirXchange(text string) {
	v.theirXchange.SetText(text)
}

func (v *entryView) Band() string {
	return v.band.GetActiveText()
}

func (v *entryView) SetBand(text string) {
	v.ignoreComboChange = true
	defer func() { v.ignoreComboChange = false }()
	v.band.SetActiveID(text)
}

func (v *entryView) Mode() string {
	return v.mode.GetActiveText()
}

func (v *entryView) SetMode(text string) {
	v.ignoreComboChange = true
	defer func() { v.ignoreComboChange = false }()
	v.mode.SetActiveID(text)
}

func (v *entryView) MyReport() string {
	text, err := v.myReport.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (v *entryView) SetMyReport(text string) {
	v.myReport.SetText(text)
}

func (v *entryView) MyNumber() string {
	text, err := v.myNumber.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (v *entryView) SetMyNumber(text string) {
	v.myNumber.SetText(text)
}

func (v *entryView) MyXchange() string {
	text, err := v.myXchange.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (v *entryView) SetMyXchange(text string) {
	v.myXchange.SetText(text)
}

func (v *entryView) EnableExchangeFields(theirNumber, theirXchange bool) {
	v.theirNumber.SetSensitive(theirNumber)
	v.theirXchange.SetSensitive(theirXchange)
}

func (v *entryView) SetActiveField(field core.EntryField) {
	entry := v.fieldToEntry(field)
	entry.GrabFocus()
}

func (v *entryView) fieldToEntry(field core.EntryField) *gtk.Entry {
	switch field {
	case core.CallsignField:
		return v.callsign
	case core.TheirReportField:
		return v.theirReport
	case core.TheirNumberField:
		return v.theirNumber
	case core.TheirXchangeField:
		return v.theirXchange
	case core.MyReportField:
		return v.myReport
	case core.MyNumberField:
		return v.myNumber
	case core.MyXchangeField:
		return v.myXchange
	case core.OtherField:
		return v.callsign
	default:
		log.Fatalf("Unknown entry field %d", field)
	}
	panic("this is never reached")
}

func (v *entryView) entryToField(entry *gtk.Entry) core.EntryField {
	name, _ := entry.GetName()
	switch name {
	case "callsignEntry":
		return core.CallsignField
	case "theirReportEntry":
		return core.TheirReportField
	case "theirNumberEntry":
		return core.TheirNumberField
	case "theirXchangeEntry":
		return core.TheirXchangeField
	case "myReportEntry":
		return core.MyReportField
	case "myNumberEntry":
		return core.MyNumberField
	case "myXchangeEntry":
		return core.MyXchangeField
	default:
		return core.OtherField
	}
}

func (v *entryView) SetDuplicateMarker(duplicate bool) {
	if duplicate {
		addStyleClass(&v.entryRoot.Widget, "duplicate")
	} else {
		removeStyleClass(&v.entryRoot.Widget, "duplicate")
	}
}

func (v *entryView) SetEditingMarker(editing bool) {
	if editing {
		addStyleClass(&v.entryRoot.Widget, "editing")
	} else {
		removeStyleClass(&v.entryRoot.Widget, "editing")
	}
}

func (v *entryView) ShowMessage(args ...interface{}) {
	v.messageLabel.SetText(fmt.Sprint(args...))
}

func (v *entryView) ClearMessage() {
	v.messageLabel.SetText("")
}
