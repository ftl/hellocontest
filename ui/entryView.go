//go:build !fyne

package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

// EntryController controls the entry of QSO data.
type EntryController interface {
	GotoNextField() core.EntryField
	GotoNextPlaceholder()
	SetActiveField(core.EntryField)

	Enter(string)
	SelectMatch(int)
	SelectBestMatch()
	SendQuestion()
	StopTX()

	Log()
	Clear()
}

type entryView struct {
	controller EntryController

	// style       *style
	ignoreInput bool

	entryRoot            *gtk.Grid
	utc                  *gtk.Label
	myCall               *gtk.Label
	frequency            *gtk.Label
	callsign             *gtk.Entry
	theirExchangesParent *gtk.Grid
	theirExchanges       []*gtk.Entry
	band                 *gtk.ComboBoxText
	mode                 *gtk.ComboBoxText
	myExchangesParent    *gtk.Grid
	myExchanges          []*gtk.Entry
	logButton            *gtk.Button
	clearButton          *gtk.Button
	messageLabel         *gtk.Label
}

func setupEntryView(builder *gtk.Builder) *entryView {
	result := new(entryView)

	result.entryRoot = getUI(builder, "entryGrid").(*gtk.Grid)
	result.utc = getUI(builder, "utcLabel").(*gtk.Label)
	result.myCall = getUI(builder, "myCallLabel").(*gtk.Label)
	result.frequency = getUI(builder, "frequencyLabel").(*gtk.Label)
	result.callsign = getUI(builder, "callsignEntry").(*gtk.Entry)
	result.theirExchangesParent = getUI(builder, "theirExchangesGrid").(*gtk.Grid)
	result.band = getUI(builder, "bandCombo").(*gtk.ComboBoxText)
	result.mode = getUI(builder, "modeCombo").(*gtk.ComboBoxText)
	result.myExchangesParent = getUI(builder, "myExchangesGrid").(*gtk.Grid)
	result.logButton = getUI(builder, "logButton").(*gtk.Button)
	result.clearButton = getUI(builder, "clearButton").(*gtk.Button)
	result.messageLabel = getUI(builder, "messageLabel").(*gtk.Label)

	result.addEntryEventHandlers(&result.callsign.Widget)
	result.addEntryEventHandlers(&result.band.Widget)
	result.addEntryEventHandlers(&result.mode.Widget)

	result.logButton.Connect("clicked", result.onLogButtonClicked)
	result.clearButton.Connect("clicked", result.onClearButtonClicked)

	setupBandCombo(result.band)
	setupModeCombo(result.mode)

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
	ctrl := keyEvent.State()&gdk.CONTROL_MASK != 0
	alt := keyEvent.State()&gdk.MOD1_MASK != 0 // MOD1 = ALT right
	key := keyEvent.KeyVal()

	switch key {
	case gdk.KEY_1, gdk.KEY_2, gdk.KEY_3, gdk.KEY_4, gdk.KEY_5, gdk.KEY_6, gdk.KEY_7, gdk.KEY_8, gdk.KEY_9:
		if alt {
			index := int(key - gdk.KEY_1)
			v.controller.SelectMatch(index)
			return true
		} else {
			return false
		}
	case gdk.KEY_Tab:
		v.controller.GotoNextField()
		return true
	case gdk.KEY_space:
		if ctrl {
			v.controller.GotoNextPlaceholder()
		} else {
			v.controller.GotoNextField()
		}
		return true
	case gdk.KEY_Return:
		if alt {
			v.controller.SelectBestMatch()
		} else {
			v.controller.Log()
		}
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

func (v *entryView) onClearButtonClicked(button *gtk.Button) bool {
	v.controller.Clear()
	return true
}

func (v *entryView) setTextWithoutChangeEvent(f func(string), value string) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	f(value)
}

func (v *entryView) SetUTC(text string) {
	v.utc.SetText(text)
}

func (v *entryView) SetMyCall(text string) {
	v.myCall.SetText(text)
}

func (v *entryView) SetFrequency(frequency core.Frequency) {
	v.frequency.SetText(fmt.Sprintf("%.2f kHz", frequency/1000.0))
}

func (v *entryView) SetCallsign(text string) {
	v.setTextWithoutChangeEvent(v.callsign.SetText, text)
}

func (v *entryView) SetTheirExchange(index int, text string) {
	i := index - 1
	if i < 0 || i >= len(v.theirExchanges) {
		return
	}
	v.setTextWithoutChangeEvent(v.theirExchanges[i].SetText, text)
}

func (v *entryView) SetBand(text string) {
	v.setTextWithoutChangeEvent(func(s string) { v.band.SetActiveID(s) }, text)
}

func (v *entryView) SetMode(text string) {
	v.setTextWithoutChangeEvent(func(s string) { v.mode.SetActiveID(s) }, text)
}

func (v *entryView) SetMyExchange(index int, text string) {
	i := index - 1
	if i < 0 || i >= len(v.myExchanges) {
		return
	}
	v.setTextWithoutChangeEvent(v.myExchanges[i].SetText, text)
}

func (v *entryView) SetMyExchangeFields(fields []core.ExchangeField) {
	v.setExchangeFields(fields, v.myExchangesParent, &v.myExchanges)
}

func (v *entryView) SetTheirExchangeFields(fields []core.ExchangeField) {
	v.setExchangeFields(fields, v.theirExchangesParent, &v.theirExchanges)
}

func (v *entryView) setExchangeFields(fields []core.ExchangeField, parent *gtk.Grid, entries *[]*gtk.Entry) {
	for _, entry := range *entries {
		entry.Destroy()
		parent.RemoveColumn(0)
	}

	*entries = make([]*gtk.Entry, len(fields))
	for i, field := range fields {
		entry, err := gtk.EntryNew()
		if err != nil {
			log.Printf("cannot create entry for %s: %v", field.Field, err)
			break
		}
		entry.SetName(string(field.Field))
		entry.SetPlaceholderText(field.Short)
		entry.SetTooltipText(field.Short) // TODO use field.Hint
		entry.SetHExpand(true)
		entry.SetHAlign(gtk.ALIGN_FILL)
		entry.SetWidthChars(4)
		entry.SetSensitive(!field.ReadOnly)

		(*entries)[i] = entry
		parent.Add(entry)
		v.addEntryEventHandlers(&entry.Widget)
	}
	parent.ShowAll()
}

func (v *entryView) SetActiveField(field core.EntryField) {
	widget := v.fieldToWidget(field)
	widget.GrabFocus()
}

func (v *entryView) SelectText(field core.EntryField, s string) {
	entry := v.fieldToEntry(field)
	if entry == nil {
		return
	}
	text, err := entry.GetText()
	if err != nil {
		return
	}
	index := strings.Index(strings.ToUpper(text), strings.ToUpper(s))
	if index == -1 {
		return
	}
	entry.SelectRegion(index, index+len(s))
}

func (v *entryView) fieldToWidget(field core.EntryField) *gtk.Widget {
	switch field {
	case core.CallsignField:
		return &v.callsign.Widget
	case core.BandField:
		return &v.band.Widget
	case core.ModeField:
		return &v.mode.Widget
	case core.OtherField:
		return &v.callsign.Widget
	}
	switch {
	case field.IsMyExchange():
		i := field.ExchangeIndex() - 1
		return &v.myExchanges[i].Widget
	case field.IsTheirExchange():
		i := field.ExchangeIndex() - 1
		return &v.theirExchanges[i].Widget
	default:
		log.Fatalf("Unknown entry field %s", field)
	}
	panic("this is never reached")
}

func (v *entryView) fieldToEntry(field core.EntryField) *gtk.Entry {
	switch field {
	case core.CallsignField:
		return v.callsign
	case core.OtherField:
		return v.callsign
	}
	switch {
	case field.IsMyExchange():
		i := field.ExchangeIndex() - 1
		return v.myExchanges[i]
	case field.IsTheirExchange():
		i := field.ExchangeIndex() - 1
		return v.theirExchanges[i]
	}
	return nil
}

func (v *entryView) widgetToField(widget *gtk.Widget) core.EntryField {
	name, _ := widget.GetName()
	switch name {
	case "callsignEntry":
		return core.CallsignField
	case "bandCombo":
		return core.BandField
	case "modeCombo":
		return core.ModeField
	default:
		if core.IsExchangeField(name) {
			return core.EntryField(name)
		}
		return core.OtherField
	}
}

const (
	entryDuplicateClass style.Class = "entry-duplicate"
	entryEditingClass   style.Class = "entry-editing"
)

func (v *entryView) SetDuplicateMarker(duplicate bool) {
	if duplicate {
		style.AddClass(&v.entryRoot.Widget, entryDuplicateClass)
	} else {
		style.RemoveClass(&v.entryRoot.Widget, entryDuplicateClass)
	}
}

func (v *entryView) SetEditingMarker(editing bool) {
	if editing {
		style.AddClass(&v.entryRoot.Widget, entryEditingClass)
	} else {
		style.RemoveClass(&v.entryRoot.Widget, entryEditingClass)
	}
}

func (v *entryView) ShowMessage(args ...any) {
	v.messageLabel.SetText(fmt.Sprint(args...))
}

func (v *entryView) ClearMessage() {
	v.messageLabel.SetText("")
}
