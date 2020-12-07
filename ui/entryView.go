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

	Enter(string)
	SendQuestion()

	Log()
	Reset()
}

type entryView struct {
	controller EntryController

	style       *style
	ignoreInput bool

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

	result.addEntryEventHandlers(&result.callsign.Widget)
	result.addEntryEventHandlers(&result.theirReport.Widget)
	result.addEntryEventHandlers(&result.theirNumber.Widget)
	result.addEntryEventHandlers(&result.theirXchange.Widget)
	result.addEntryEventHandlers(&result.myReport.Widget)
	result.addEntryEventHandlers(&result.myNumber.Widget)
	result.addEntryEventHandlers(&result.myXchange.Widget)
	result.addEntryEventHandlers(&result.band.Widget)
	result.addEntryEventHandlers(&result.mode.Widget)

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

func (v *entryView) addEntryEventHandlers(w *gtk.Widget) {
	w.Connect("key_press_event", v.onEntryKeyPress)
	w.Connect("focus_in_event", v.onEntryFocusIn)
	w.Connect("focus_out_event", v.onEntryFocusOut)
	w.Connect("changed", v.onEntryChanged)
}

func (v *entryView) onEntryKeyPress(_ interface{}, event *gdk.Event) bool {
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

func (v *entryView) onEntryFocusIn(widget interface{}, _ *gdk.Event) bool {
	var field core.EntryField
	switch w := widget.(type) {
	case *gtk.Entry:
		field = v.widgetToField(&w.Widget)
	case *gtk.ComboBoxText:
		field = v.widgetToField(&w.Widget)
	default:
		field = core.OtherField
	}
	v.controller.SetActiveField(field)
	return false
}

func (v *entryView) onEntryFocusOut(widget interface{}, _ *gdk.Event) bool {
	if entry, ok := widget.(*gtk.Entry); ok {
		entry.SelectRegion(0, 0)
	}
	return false
}

func (v *entryView) onEntryChanged(widget interface{}) bool {
	if v.controller == nil {
		return false
	}
	if v.ignoreInput {
		return false
	}

	switch w := widget.(type) {
	case *gtk.Entry:
		text, err := w.GetText()
		if err != nil {
			return false
		}
		v.controller.Enter(text)
	case *gtk.ComboBoxText:
		activeField := v.widgetToField(&w.Widget)
		v.controller.SetActiveField(activeField)
		text := w.GetActiveText()
		v.controller.Enter(text)
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

func (v *entryView) setTextWithoutChangeEvent(f func(string), value string) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	f(value)
}

func (v *entryView) SetCallsign(text string) {
	v.setTextWithoutChangeEvent(v.callsign.SetText, text)
}

func (v *entryView) SetTheirReport(text string) {
	v.setTextWithoutChangeEvent(v.theirReport.SetText, text)
}

func (v *entryView) SetTheirNumber(text string) {
	v.setTextWithoutChangeEvent(v.theirNumber.SetText, text)
}

func (v *entryView) SetTheirXchange(text string) {
	v.setTextWithoutChangeEvent(v.theirXchange.SetText, text)
}

func (v *entryView) SetFrequency(frequency core.Frequency) {
	// ignore
}

func (v *entryView) SetBand(text string) {
	v.setTextWithoutChangeEvent(func(s string) { v.band.SetActiveID(s) }, text)
}

func (v *entryView) SetMode(text string) {
	v.setTextWithoutChangeEvent(func(s string) { v.mode.SetActiveID(s) }, text)
}

func (v *entryView) SetMyReport(text string) {
	v.setTextWithoutChangeEvent(v.myReport.SetText, text)
}

func (v *entryView) SetMyNumber(text string) {
	v.setTextWithoutChangeEvent(v.myNumber.SetText, text)
}

func (v *entryView) SetMyXchange(text string) {
	v.setTextWithoutChangeEvent(v.myXchange.SetText, text)
}

func (v *entryView) EnableExchangeFields(theirNumber, theirXchange bool) {
	v.theirNumber.SetSensitive(theirNumber)
	v.theirXchange.SetSensitive(theirXchange)
}

func (v *entryView) SetActiveField(field core.EntryField) {
	widget := v.fieldToWidget(field)
	widget.GrabFocus()
}

func (v *entryView) fieldToWidget(field core.EntryField) *gtk.Widget {
	switch field {
	case core.CallsignField:
		return &v.callsign.Widget
	case core.TheirReportField:
		return &v.theirReport.Widget
	case core.TheirNumberField:
		return &v.theirNumber.Widget
	case core.TheirXchangeField:
		return &v.theirXchange.Widget
	case core.MyReportField:
		return &v.myReport.Widget
	case core.MyNumberField:
		return &v.myNumber.Widget
	case core.MyXchangeField:
		return &v.myXchange.Widget
	case core.BandField:
		return &v.band.Widget
	case core.ModeField:
		return &v.mode.Widget
	case core.OtherField:
		return &v.callsign.Widget
	default:
		log.Fatalf("Unknown entry field %d", field)
	}
	panic("this is never reached")
}

func (v *entryView) widgetToField(widget *gtk.Widget) core.EntryField {
	name, _ := widget.GetName()
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
	case "bandCombo":
		return core.BandField
	case "modeCombo":
		return core.ModeField
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
