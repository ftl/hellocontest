package ui

import (
	"fmt"
	"strings"
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core/model"
)

const (
	parrot          = "🦜"
	txIndicatorText = "TX"
)

// EntryController controls the entry of QSO data.
type EntryController interface {
	GotoNextField() model.EntryField
	GotoNextFieldSkippingReport() model.EntryField
	GotoNextPlaceholder()
	SetFocusedVFO(model.VFOID)
	SetActiveField(model.EntryField)

	Enter(string)
	SelectMatch(int)
	SelectBestMatchOnFrequency()
	SendQuestion()
	RepeatLastTransmission()
	StopTX()

	EnterPressed()
	Log()
	Clear()

	LogVFO(model.VFOID)
	ClearVFO(model.VFOID)
}

type entryVFOWidgets struct {
	// radio control
	vfoLabel       *qtlib.QLabel
	frequencyLabel *qtlib.QLabel
	band           *qtlib.QComboBox
	mode           *qtlib.QComboBox
	xit            *qtlib.QCheckBox
	rit            *qtlib.QCheckBox
	txIndicator    *qtlib.QLabel

	// QSO entry
	callsign            *qtlib.QLineEdit
	theirExchangeFields []*qtlib.QLineEdit
	logButton           *qtlib.QPushButton
	clearButton         *qtlib.QPushButton

	messageLabel *qtlib.QLabel
}

// entryView handles interaction and content related to core/entry. It owns all related widgets.
type entryView struct {
	myExchangeFields []*qtlib.QLineEdit

	vfo [model.VFOCount]entryVFOWidgets

	vfoWorkmode [model.VFOCount]model.Workmode

	area *entryArea

	ignoreInput bool

	controller             EntryController
	incrementalTuningInput IncrementalTuningController
}

type IncrementalTuningController interface {
	SetIncrementalTuningActive(model.VFOID, model.IncrementalTuningKind, bool)
}

func newEntryView() *entryView {
	v := &entryView{}

	v.vfo[model.VFO1] = newEntryVFOWidgets("VFO 1")
	v.registerVFOWidgets(model.VFO1, &v.vfo[model.VFO1])

	v.vfo[model.VFO2] = newEntryVFOWidgets("VFO 2")
	v.registerVFOWidgets(model.VFO2, &v.vfo[model.VFO2])

	return v
}

func newEntryVFOWidgets(vfoName string) entryVFOWidgets {
	w := entryVFOWidgets{}

	w.vfoLabel = qtlib.NewQLabel3(vfoName)
	w.txIndicator = qtlib.NewQLabel3(txIndicatorText)
	w.frequencyLabel = qtlib.NewQLabel3("- kHz")

	w.band = qtlib.NewQComboBox2()
	setupBandCombo(w.band)

	w.mode = qtlib.NewQComboBox2()
	setupModeCombo(w.mode)

	w.xit = qtlib.NewQCheckBox3("XIT")
	w.xit.SetFocusPolicy(qtlib.NoFocus)

	w.rit = qtlib.NewQCheckBox3("RIT")
	w.rit.SetFocusPolicy(qtlib.NoFocus)

	w.callsign = qtlib.NewQLineEdit2()
	w.callsign.SetPlaceholderText("Call")

	w.logButton = qtlib.NewQPushButton3("Log")
	w.logButton.SetFocusPolicy(qtlib.NoFocus)

	w.clearButton = qtlib.NewQPushButton3("Clear")
	w.clearButton.SetFocusPolicy(qtlib.NoFocus)

	w.messageLabel = qtlib.NewQLabel3("")

	return w
}

func (v *entryView) registerVFOWidgets(vfo model.VFOID, widgets *entryVFOWidgets) {
	v.connectEditSignals(widgets.callsign, vfo, model.CallsignField, true)
	v.connectComboSignals(widgets.band, vfo, model.BandField)
	v.connectComboSignals(widgets.mode, vfo, model.ModeField)
	widgets.logButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.LogVFO(vfo)
		}
	})
	widgets.clearButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.ClearVFO(vfo)
		}
	})
	widgets.xit.OnStateChanged(func(state int) {
		if v.incrementalTuningInput != nil {
			v.incrementalTuningInput.SetIncrementalTuningActive(vfo, model.XIT, state != 0)
		}
	})
	widgets.rit.OnStateChanged(func(state int) {
		if v.incrementalTuningInput != nil {
			v.incrementalTuningInput.SetIncrementalTuningActive(vfo, model.RIT, state != 0)
		}
	})
}

func (v *entryView) connectComboSignals(combo *qtlib.QComboBox, vfo model.VFOID, field model.EntryField) {
	combo.OnCurrentTextChanged(func(text string) {
		if v.controller == nil || v.ignoreInput {
			return
		}
		v.controller.SetFocusedVFO(vfo)
		v.controller.SetActiveField(field)
		v.controller.Enter(text)
	})
	combo.OnFocusInEvent(func(super func(ev *qtlib.QFocusEvent), ev *qtlib.QFocusEvent) {
		super(ev)
		if v.controller != nil {
			v.controller.SetFocusedVFO(vfo)
			v.controller.SetActiveField(field)
		}
	})
	combo.OnKeyPressEvent(func(super func(ev *qtlib.QKeyEvent), ev *qtlib.QKeyEvent) {
		v.handleKeyPress(super, ev, false)
	})
}

func (v *entryView) connectEditSignals(edit *qtlib.QLineEdit, vfo model.VFOID, field model.EntryField, isTheirRow bool) {
	edit.OnTextChanged(func(text string) {
		if v.controller == nil || v.ignoreInput {
			return
		}
		v.controller.Enter(text)
	})
	edit.OnFocusInEvent(func(super func(ev *qtlib.QFocusEvent), ev *qtlib.QFocusEvent) {
		super(ev)
		edit.SelectAll()
		if v.controller != nil {
			v.controller.SetFocusedVFO(vfo)
			v.controller.SetActiveField(field)
		}
	})
	edit.OnFocusOutEvent(func(super func(ev *qtlib.QFocusEvent), ev *qtlib.QFocusEvent) {
		super(ev)
		edit.Deselect()
	})
	edit.OnKeyPressEvent(func(super func(ev *qtlib.QKeyEvent), ev *qtlib.QKeyEvent) {
		v.handleKeyPress(super, ev, isTheirRow)
	})
	if isTheirRow {
		edit.OnFocusNextPrevChild(func(super func(next bool) bool, next bool) bool {
			if v.controller == nil {
				return super(next)
			}
			v.controller.GotoNextField()
			return true
		})
	}
}

func (v *entryView) handleKeyPress(super func(ev *qtlib.QKeyEvent), ev *qtlib.QKeyEvent, isTheirRow bool) {
	if v.controller == nil {
		super(ev)
		return
	}

	key := ev.Key()
	ctrl := ev.Modifiers()&qtlib.ControlModifier != 0
	alt := ev.Modifiers()&qtlib.AltModifier != 0

	switch {
	case alt && key >= int(qtlib.Key_1) && key <= int(qtlib.Key_9):
		index := key - int(qtlib.Key_1)
		v.controller.SelectMatch(index)
	case key == int(qtlib.Key_Tab) || key == int(qtlib.Key_Backtab):
		if !isTheirRow {
			super(ev)
			return
		}
		v.controller.GotoNextField()
	case key == int(qtlib.Key_Space) && ctrl:
		v.controller.GotoNextPlaceholder()
	case key == int(qtlib.Key_Space):
		v.controller.GotoNextFieldSkippingReport()
	case key == int(qtlib.Key_Return) || key == int(qtlib.Key_Enter):
		if !(alt || ctrl) {
			v.controller.EnterPressed()
		}
	case key == int(qtlib.Key_Question):
		v.controller.SendQuestion()
	case key == int(qtlib.Key_Equal):
		v.controller.RepeatLastTransmission()
	default:
		super(ev)
		return
	}
	// All handled cases consume the event (don't call super)
}

func (v *entryView) SetEntryController(controller EntryController) {
	v.controller = controller
}

func (v *entryView) SetIncrementalTuningController(controller IncrementalTuningController) {
	v.incrementalTuningInput = controller
}

func (v *entryView) SetFrequency(vfo model.VFOID, frequency model.Frequency) {
	v.vfo[vfo].frequencyLabel.SetText(fmt.Sprintf("%.2f kHz", frequency/1000.0))
}

func (v *entryView) SetSerialClaim(vfo model.VFOID, serial model.QSONumber, committed bool) {
	v.area.setSerialClaim(vfo, serial, committed)
}

func (v *entryView) SetCallsign(vfo model.VFOID, text string) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	widget := v.vfo[vfo].callsign
	if widget == nil {
		return
	}

	widget.SetText(text)
}

func (v *entryView) SetBand(vfo model.VFOID, text string) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	combo := v.vfo[vfo].band
	if combo == nil {
		return
	}

	idx := combo.FindText(text)
	if idx >= 0 {
		combo.SetCurrentIndex(idx)
	}
}

func (v *entryView) SetMode(vfo model.VFOID, text string) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	combo := v.vfo[vfo].mode
	if combo == nil {
		return
	}

	idx := combo.FindText(text)
	if idx >= 0 {
		combo.SetCurrentIndex(idx)
	}
}

func incrementalTuningWidget(widgets *entryVFOWidgets, kind model.IncrementalTuningKind) *qtlib.QCheckBox {
	if kind == model.RIT {
		return widgets.rit
	}
	return widgets.xit
}

func (v *entryView) VFOIncrementalTuningActiveChanged(vfo model.VFOID, kind model.IncrementalTuningKind, active bool) {
	widget := incrementalTuningWidget(&v.vfo[vfo], kind)
	if widget == nil {
		return
	}
	if widget.IsChecked() == active {
		return
	}
	widget.SetChecked(active)
}

func (v *entryView) VFOIncrementalTuningChanged(vfo model.VFOID, kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	widget := incrementalTuningWidget(&v.vfo[vfo], kind)
	if widget == nil {
		return
	}

	if active {
		widget.SetText(fmt.Sprintf("%s %s", kind, offset))
	} else {
		widget.SetText(kind.String())
	}
}

func (v *entryView) VFOIncrementalTuningVisibilityChanged(vfo model.VFOID, kind model.IncrementalTuningKind, visible bool) {
	v.area.setIncrementalTuningVisible(vfo, kind, visible)
}

func (v *entryView) SetTXState(vfo model.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration) {
	widget := v.vfo[vfo].txIndicator
	if widget == nil {
		return
	}

	text := txIndicatorText
	if parrotActive {
		text = parrot
		if parrotTimeLeft > 0 {
			text += fmt.Sprintf(": %v", parrotTimeLeft)
		}
	}
	widget.SetText(text)

	v.area.setTXState(vfo, ptt, parrotActive)
}

func (v *entryView) SetMyExchange(index int, text string) {
	i := index - 1
	if i < 0 || i >= len(v.myExchangeFields) {
		return
	}
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	v.myExchangeFields[i].SetText(text)
}

func (v *entryView) SetTheirExchange(vfo model.VFOID, index int, text string) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	fields := v.vfo[vfo].theirExchangeFields
	i := index - 1
	if i < 0 || i >= len(fields) {
		return
	}
	fields[i].SetText(text)
}

func (v *entryView) SetExchangeFields(myExchangeFields, theirExchangeFields []model.ExchangeField) {
	v.setExchangeFields(myExchangeFields, &v.myExchangeFields, false, model.VFO1)
	v.setExchangeFields(theirExchangeFields, &v.vfo[model.VFO1].theirExchangeFields, true, model.VFO1)
	v.setExchangeFields(theirExchangeFields, &v.vfo[model.VFO2].theirExchangeFields, true, model.VFO2)
}

func (v *entryView) setExchangeFields(fields []model.ExchangeField, editFields *[]*qtlib.QLineEdit, isTheirRow bool, vfo model.VFOID) {
	// Remove old fields
	for _, f := range *editFields {
		f.SetParent(nil)
		f.Delete()
	}

	// Create new fields
	*editFields = make([]*qtlib.QLineEdit, len(fields))
	for i, field := range fields {
		editField := qtlib.NewQLineEdit2()
		editField.SetPlaceholderText(field.Short)
		editField.SetEnabled(!field.ReadOnly)

		v.connectEditSignals(editField, vfo, model.TheirExchangeField(i+1), isTheirRow)
		(*editFields)[i] = editField
	}
}

func (v *entryView) SetVFOWorkmode(vfo model.VFOID, workmode model.Workmode) {
	v.vfoWorkmode[vfo] = workmode
	v.updateVFOLabel(vfo)
}

func (v *entryView) SetTXVFO(vfo model.VFOID) {
	v.area.setTXVFO(vfo)
}

func (v *entryView) updateVFOLabel(vfo model.VFOID) {
	label := v.vfo[vfo].vfoLabel
	if label == nil {
		return
	}
	text := fmt.Sprintf("VFO %d", vfo+1)
	switch v.vfoWorkmode[vfo] {
	case model.Run:
		text += " RUN"
	case model.SearchPounce:
		text += " S&P"
	}
	label.SetTextFormat(qtlib.PlainText)
	label.SetText(text)
}

func (v *entryView) SetActiveVFO(vfo model.VFOID) {
	v.area.setActiveVFO(vfo)
}

func (v *entryView) SetActiveField(vfo model.VFOID, field model.EntryField) {
	callsign := v.vfo[vfo].callsign
	band := v.vfo[vfo].band
	mode := v.vfo[vfo].mode
	theirExchange := v.vfo[vfo].theirExchangeFields

	switch field {
	case model.CallsignField, model.OtherField:
		if callsign != nil {
			callsign.SetFocus()
		}
	case model.BandField:
		if band != nil {
			band.SetFocus()
		}
	case model.ModeField:
		if mode != nil {
			mode.SetFocus()
		}
	default:
		switch {
		case field.IsTheirExchange():
			i := field.ExchangeIndex() - 1
			if i >= 0 && i < len(theirExchange) {
				theirExchange[i].SetFocus()
			}
		case field.IsMyExchange():
			i := field.ExchangeIndex() - 1
			if i >= 0 && i < len(v.myExchangeFields) {
				v.myExchangeFields[i].SetFocus()
			}
		}
	}
}

func (v *entryView) SelectText(vfo model.VFOID, field model.EntryField, s string) {
	edit := v.fieldToEntry(vfo, field)
	if edit == nil {
		return
	}
	text := edit.Text()
	index := strings.Index(strings.ToUpper(text), strings.ToUpper(s))
	if index == -1 {
		return
	}
	edit.SetSelection(index, len(s))
}

func (v *entryView) fieldToEntry(vfo model.VFOID, field model.EntryField) *qtlib.QLineEdit {
	callsign := v.vfo[vfo].callsign
	theirExchange := v.vfo[vfo].theirExchangeFields
	switch field {
	case model.CallsignField, model.OtherField:
		return callsign
	}
	switch {
	case field.IsMyExchange():
		i := field.ExchangeIndex() - 1
		if i >= 0 && i < len(v.myExchangeFields) {
			return v.myExchangeFields[i]
		}
	case field.IsTheirExchange():
		i := field.ExchangeIndex() - 1
		if i >= 0 && i < len(theirExchange) {
			return theirExchange[i]
		}
	}
	return nil
}

func (v *entryView) SetDuplicateMarker(vfo model.VFOID, duplicate bool) {
	v.area.setDuplicateMarker(vfo, duplicate)
}

func (v *entryView) SetEditingMarker(vfo model.VFOID, editing bool) {
	v.area.setEditingMarker(vfo, editing)
}

func (v *entryView) ShowMessage(vfo model.VFOID, args ...any) {
	widget := v.vfo[vfo].messageLabel
	if widget == nil {
		return
	}
	text := fmt.Sprint(args...)
	widget.SetText(text)
	v.area.setMessageVisible(vfo, text != "")
}

func (v *entryView) ClearMessage(vfo model.VFOID) {
	widget := v.vfo[vfo].messageLabel
	if widget == nil {
		return
	}
	widget.SetText("")
	v.area.setMessageVisible(vfo, false)
}

func (v *entryView) SetVFOEnabled(vfo model.VFOID, enabled bool) {
	if vfo == model.VFO1 {
		return
	}
	v.area.setVFO2Enabled(enabled)
}
