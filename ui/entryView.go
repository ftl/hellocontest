package ui

import (
	"fmt"
	"strings"
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

const (
	parrot         = "🦜"
	txVFOIndicator = `<span style="color: red; font-size: 16pt;">&#x25CF;</span>`
)

// EntryController controls the entry of QSO data.
type EntryController interface {
	GotoNextField() core.EntryField
	GotoNextPlaceholder()
	SetFocusedVFO(core.VFOID)
	SetActiveField(core.EntryField)

	Enter(string)
	SelectMatch(int)
	SelectBestMatchOnFrequency()
	SendQuestion()
	RepeatLastTransmission()
	StopTX()

	EnterPressed()
	Log()
	Clear()

	LogVFO(core.VFOID)
	ClearVFO(core.VFOID)
}

type entryVFOWidgets struct {
	topSeparator *qtlib.QFrame

	vfoContainer   *qtlib.QWidget
	vfoLabel       *qtlib.QLabel
	frequencyLabel *qtlib.QLabel
	band           *qtlib.QComboBox
	mode           *qtlib.QComboBox

	serialClaimLabel *qtlib.QLabel
	xit              *qtlib.QCheckBox
	rit              *qtlib.QCheckBox
	txIndicator      *qtlib.QLabel

	callsign            *qtlib.QLineEdit
	theirExchangeFields []*qtlib.QLineEdit
	logButton           *qtlib.QPushButton
	clearButton         *qtlib.QPushButton

	messageLabel *qtlib.QLabel
}

type entryView struct {
	root *qtlib.QWidget // root widget containing the grid layout

	myCallLabel      *qtlib.QLabel
	myExchangeFields []*qtlib.QLineEdit

	vfo [core.VFOCount]entryVFOWidgets

	vfoWorkmode [core.VFOCount]core.Workmode
	itVisible   [core.VFOCount][2]bool
	txVFO       core.VFOID

	vfo2Enabled   bool
	onVFO2Enabled func(bool) // callback to centralArea for layout add/remove

	ignoreInput bool
	isDuplicate bool
	isEditing   bool

	controller             EntryController
	incrementalTuningInput IncrementalTuningController
}

type IncrementalTuningController interface {
	SetIncrementalTuningActive(core.VFOID, core.IncrementalTuningKind, bool)
}

func newEntryView() *entryView {
	v := &entryView{}

	// Row 0: myCall label, myExchanges container (cols 2-3, span 2)
	v.myCallLabel = qtlib.NewQLabel3("DL0ABC")

	v.vfo[core.VFO1] = newEntryVFOWidgets("vfo1", "VFO 1")
	v.vfo[core.VFO2] = newEntryVFOWidgets("vfo2", "VFO 2")

	// Connect signals for static widgets
	v.connectEditSignals(v.vfo[core.VFO1].callsign, core.VFO1, core.CallsignField, true)
	v.connectEditSignals(v.vfo[core.VFO2].callsign, core.VFO2, core.CallsignField, true)
	v.connectComboSignals(v.vfo[core.VFO1].band, core.VFO1, core.BandField)
	v.connectComboSignals(v.vfo[core.VFO2].band, core.VFO2, core.BandField)
	v.connectComboSignals(v.vfo[core.VFO1].mode, core.VFO1, core.ModeField)
	v.connectComboSignals(v.vfo[core.VFO2].mode, core.VFO2, core.ModeField)

	// Connect button/checkbox signals
	v.vfo[core.VFO1].logButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.LogVFO(core.VFO1)
		}
	})
	v.vfo[core.VFO2].logButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.LogVFO(core.VFO2)
		}
	})
	v.vfo[core.VFO1].clearButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.ClearVFO(core.VFO1)
		}
	})
	v.vfo[core.VFO2].clearButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.ClearVFO(core.VFO2)
		}
	})
	v.vfo[core.VFO1].xit.OnStateChanged(func(state int) {
		if v.incrementalTuningInput != nil {
			v.incrementalTuningInput.SetIncrementalTuningActive(core.VFO1, core.XIT, state != 0)
		}
	})
	v.vfo[core.VFO2].xit.OnStateChanged(func(state int) {
		if v.incrementalTuningInput != nil {
			v.incrementalTuningInput.SetIncrementalTuningActive(core.VFO2, core.XIT, state != 0)
		}
	})
	v.vfo[core.VFO1].rit.OnStateChanged(func(state int) {
		if v.incrementalTuningInput != nil {
			v.incrementalTuningInput.SetIncrementalTuningActive(core.VFO1, core.RIT, state != 0)
		}
	})
	v.vfo[core.VFO2].rit.OnStateChanged(func(state int) {
		if v.incrementalTuningInput != nil {
			v.incrementalTuningInput.SetIncrementalTuningActive(core.VFO2, core.RIT, state != 0)
		}
	})

	return v
}

func newEntryVFOWidgets(prefix string, vfoName string) entryVFOWidgets {
	w := entryVFOWidgets{}

	w.topSeparator = qtlib.NewQFrame2()
	w.topSeparator.SetFrameShape(qtlib.QFrame__HLine)
	w.topSeparator.SetFrameShadow(qtlib.QFrame__Sunken)

	w.vfoContainer = qtlib.NewQWidget2()
	vfoContainerLayout := qtlib.NewQHBoxLayout(w.vfoContainer)

	w.vfoLabel = qtlib.NewQLabel3(vfoName)
	w.vfoLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "Label"))
	w.vfoLabel.SetAlignment(qtlib.AlignCenter | qtlib.AlignVCenter)
	vfoContainerLayout.AddWidget(w.vfoLabel.QWidget)

	w.frequencyLabel = qtlib.NewQLabel3("- kHz")
	w.frequencyLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "FrequencyLabel"))
	w.frequencyLabel.SetAlignment(qtlib.AlignTrailing | qtlib.AlignVCenter)
	vfoContainerLayout.AddWidget2(w.frequencyLabel.QWidget, 2)

	w.band = qtlib.NewQComboBox2()
	w.band.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "BandCombo"))
	setupBandCombo(w.band)
	vfoContainerLayout.AddWidget(w.band.QWidget)

	w.mode = qtlib.NewQComboBox2()
	w.mode.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "ModeCombo"))
	setupModeCombo(w.mode)
	vfoContainerLayout.AddWidget(w.mode.QWidget)

	w.xit = qtlib.NewQCheckBox3("XIT")
	w.xit.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "XIT"))
	w.xit.SetVisible(false)

	w.rit = qtlib.NewQCheckBox3("RIT")
	w.rit.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "RIT"))
	w.rit.SetVisible(false)

	w.txIndicator = qtlib.NewQLabel3("")
	w.txIndicator.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "TX"))

	w.callsign = qtlib.NewQLineEdit2()
	w.callsign.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "CallsignEntry"))
	w.callsign.SetPlaceholderText("Call")
	w.callsign.SetStyleSheet(EntryFieldStyle)

	w.logButton = qtlib.NewQPushButton3("Log")
	w.logButton.SetFocusPolicy(qtlib.NoFocus)

	w.clearButton = qtlib.NewQPushButton3("Clear")
	w.clearButton.SetFocusPolicy(qtlib.NoFocus)

	w.messageLabel = qtlib.NewQLabel3("")

	return w
}

// setRootWidgets sets the widget that is used to show the duplicate and editing marks
func (v *entryView) setRootWidget(root *qtlib.QWidget) {
	v.root = root
	v.root.SetObjectName(*qtlib.NewQAnyStringView3("entryWidget"))
}

func (v *entryView) connectComboSignals(combo *qtlib.QComboBox, vfo core.VFOID, field core.EntryField) {
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

func (v *entryView) connectEditSignals(edit *qtlib.QLineEdit, vfo core.VFOID, field core.EntryField, isTheirRow bool) {
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
		v.controller.GotoNextField()
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

func (v *entryView) SetMyCall(text string) {
	v.myCallLabel.SetText(text)
}

func (v *entryView) SetFrequency(vfo core.VFOID, frequency core.Frequency) {
	v.vfo[vfo].frequencyLabel.SetText(fmt.Sprintf("%.2f kHz", frequency/1000.0))
}

func (v *entryView) SetSerialClaim(vfo core.VFOID, serial core.QSONumber, committed bool) {
	label := v.vfo[vfo].serialClaimLabel
	if label == nil {
		return
	}

	if serial == 0 {
		label.SetText("")
	} else {
		text := fmt.Sprintf("#%s", serial.String())
		if committed {
			text = fmt.Sprintf("<b>%s</b>", text)
		}
		label.SetText(text)
	}
}

func (v *entryView) SetCallsign(vfo core.VFOID, text string) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	widget := v.vfo[vfo].callsign
	if widget == nil {
		return
	}

	widget.SetText(text)
}

func (v *entryView) SetBand(vfo core.VFOID, text string) {
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

func (v *entryView) SetMode(vfo core.VFOID, text string) {
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

func (v *entryView) incrementalTuningWidget(vfo core.VFOID, kind core.IncrementalTuningKind) *qtlib.QCheckBox {
	if kind == core.RIT {
		return v.vfo[vfo].rit
	}
	return v.vfo[vfo].xit
}

func (v *entryView) IncrementalTuningActiveChanged(vfo core.VFOID, kind core.IncrementalTuningKind, active bool) {
	widget := v.incrementalTuningWidget(vfo, kind)
	if widget == nil {
		return
	}
	if widget.IsChecked() == active {
		return
	}
	widget.SetChecked(active)
}

func (v *entryView) VFOIncrementalTuningChanged(vfo core.VFOID, kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	widget := v.incrementalTuningWidget(vfo, kind)
	if widget == nil {
		return
	}

	if active {
		widget.SetText(fmt.Sprintf("%s %s", kind, offset))
	} else {
		widget.SetText(kind.String())
	}
}

func (v *entryView) IncrementalTuningVisibilityChanged(vfo core.VFOID, kind core.IncrementalTuningKind, visible bool) {
	v.itVisible[vfo][kind] = visible
	v.applyIncrementalTuningVisibility(vfo, kind)
}

func (v *entryView) applyIncrementalTuningVisibility(vfo core.VFOID, kind core.IncrementalTuningKind) {
	widget := v.incrementalTuningWidget(vfo, kind)
	if widget == nil {
		return
	}
	vfoEnabled := vfo == core.VFO1 || v.vfo2Enabled
	widget.SetVisible(v.itVisible[vfo][kind] && vfoEnabled)
}

func (v *entryView) SetTXState(vfo core.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration) {
	widget := v.vfo[vfo].txIndicator
	if widget == nil {
		return
	}

	var text string
	switch {
	case parrotActive:
		text = parrot
		if parrotTimeLeft > 0 {
			text += fmt.Sprintf(": %v", parrotTimeLeft)
		}
	case ptt:
		text = "On Air"
	default:
		text = ""
	}

	// TODO: use a property with a selective style
	if ptt {
		widget.SetStyleSheet(TXIndicatorActiveStyle)
	} else {
		widget.SetStyleSheet(TXIndicatorInactiveStyle)
	}
	widget.SetText(text)
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

func (v *entryView) SetTheirExchange(vfo core.VFOID, index int, text string) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	fields := v.vfo[vfo].theirExchangeFields
	i := index - 1
	if i < 0 || i >= len(fields) {
		return
	}
	fields[i].SetText(text)
}

func (v *entryView) SetSerialClaimLabelsVisible(visible bool) {
	for vfo := range core.VFOCount {
		widget := v.vfo[vfo].serialClaimLabel
		prefix := fmt.Sprintf("vfo%d", vfo+1)
		if visible && v.vfo2Enabled {
			if widget == nil {
				widget = qtlib.NewQLabel3("")
				widget.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "SerialClaim"))
				widget.SetAlignment(qtlib.AlignCenter | qtlib.AlignVCenter)
				v.vfo[vfo].serialClaimLabel = widget
			}
		} else {
			if widget != nil {
				widget.SetParent(nil)
				widget.Delete()
				v.vfo[vfo].serialClaimLabel = nil
			}
		}
	}
}

func (v *entryView) SetExchangeFields(myExchangeFields, theirExchangeFields []core.ExchangeField) {
	v.setExchangeFields(myExchangeFields, &v.myExchangeFields, false, core.VFO1)
	v.setExchangeFields(theirExchangeFields, &v.vfo[core.VFO1].theirExchangeFields, true, core.VFO1)
	v.setExchangeFields(theirExchangeFields, &v.vfo[core.VFO2].theirExchangeFields, true, core.VFO2)
}

func (v *entryView) setExchangeFields(fields []core.ExchangeField, editFields *[]*qtlib.QLineEdit, isTheirRow bool, vfo core.VFOID) {
	// Remove old fields
	for _, f := range *editFields {
		f.SetParent(nil)
		f.Delete()
	}

	// Create new fields
	*editFields = make([]*qtlib.QLineEdit, len(fields))
	for i, field := range fields {
		editField := qtlib.NewQLineEdit2()
		objName := string(field.Field)
		if vfo == core.VFO2 {
			objName = "vfo2_" + objName
		}
		editField.SetObjectName(*qtlib.NewQAnyStringView3(objName))
		editField.SetPlaceholderText(field.Short)
		editField.SetEnabled(!field.ReadOnly)

		if isTheirRow {
			editField.SetStyleSheet(EntryFieldStyle)
		}

		if vfo == core.VFO2 && !v.vfo2Enabled {
			editField.SetVisible(false)
			editField.SetEnabled(false)
		}

		v.connectEditSignals(editField, vfo, core.TheirExchangeField(i+1), isTheirRow)
		(*editFields)[i] = editField
	}
}

func (v *entryView) SetVFOWorkmode(vfo core.VFOID, workmode core.Workmode) {
	v.vfoWorkmode[vfo] = workmode
	v.updateVFOLabel(vfo)
}

func (v *entryView) SetTXVFO(vfo core.VFOID) {
	v.txVFO = vfo
	for id := core.VFOID(0); id < core.VFOCount; id++ {
		v.updateVFOLabel(id)
	}
}

func (v *entryView) updateVFOLabel(vfo core.VFOID) {
	label := v.vfo[vfo].vfoLabel
	if label == nil {
		return
	}
	text := fmt.Sprintf("VFO %d", vfo+1)
	switch v.vfoWorkmode[vfo] {
	case core.Run:
		text += " RUN"
	case core.SearchPounce:
		text += " S&amp;P"
	}
	if vfo == v.txVFO {
		text += " " + txVFOIndicator
	}
	label.SetTextFormat(qtlib.RichText)
	label.SetText(text)
}

func (v *entryView) SetActiveVFO(vfo core.VFOID) {
	// TODO: use a property with a selective style
	switch vfo {
	case core.VFO1:
		v.vfo[core.VFO1].vfoLabel.SetStyleSheet(VFOActiveStyle)
		v.vfo[core.VFO2].vfoLabel.SetStyleSheet(VFOInactiveStyle)
	case core.VFO2:
		v.vfo[core.VFO1].vfoLabel.SetStyleSheet(VFOInactiveStyle)
		v.vfo[core.VFO2].vfoLabel.SetStyleSheet(VFOActiveStyle)
	}
}

func (v *entryView) SetActiveField(vfo core.VFOID, field core.EntryField) {
	callsign := v.vfo[vfo].callsign
	band := v.vfo[vfo].band
	mode := v.vfo[vfo].mode
	theirExchange := v.vfo[vfo].theirExchangeFields

	switch field {
	case core.CallsignField, core.OtherField:
		if callsign != nil {
			callsign.SetFocus()
		}
	case core.BandField:
		if band != nil {
			band.SetFocus()
		}
	case core.ModeField:
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

func (v *entryView) SelectText(vfo core.VFOID, field core.EntryField, s string) {
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

func (v *entryView) fieldToEntry(vfo core.VFOID, field core.EntryField) *qtlib.QLineEdit {
	callsign := v.vfo[vfo].callsign
	theirExchange := v.vfo[vfo].theirExchangeFields
	switch field {
	case core.CallsignField, core.OtherField:
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

func (v *entryView) SetDuplicateMarker(vfo core.VFOID, duplicate bool) {
	// TODO step 6 follow-up: per-VFO duplicate marker. For now, only VFO1 drives the root style.
	if vfo != core.VFO1 {
		return
	}
	v.isDuplicate = duplicate
	v.updateMarkerStyle()
}

func (v *entryView) SetEditingMarker(vfo core.VFOID, editing bool) {
	if vfo != core.VFO1 {
		return
	}
	v.isEditing = editing
	v.updateMarkerStyle()
}

func (v *entryView) updateMarkerStyle() {
	if v.root == nil {
		return
	}

	switch {
	case v.isDuplicate && v.isEditing:
		v.root.SetStyleSheet(EntryDuplicateStyle)
	case v.isDuplicate:
		v.root.SetStyleSheet(EntryDuplicateStyle)
	case v.isEditing:
		v.root.SetStyleSheet(EntryEditingStyle)
	default:
		v.root.SetStyleSheet(EntryNormalStyle)
	}
}

func (v *entryView) ShowMessage(vfo core.VFOID, args ...any) {
	widget := v.vfo[vfo].messageLabel
	if widget == nil {
		return
	}
	widget.SetText(fmt.Sprint(args...))
}

func (v *entryView) ClearMessage(vfo core.VFOID) {
	widget := v.vfo[vfo].messageLabel
	if widget == nil {
		return
	}
	widget.SetText("")
}

// SetVFOEnabled toggles the visibility/enabled state of a VFO's row of widgets.
// VFO1 is always enabled. VFO2 widgets are shown/hidden as a group.
// If onVFO2Enabled is set, it delegates to centralArea for layout add/remove first.
func (v *entryView) SetVFOEnabled(vfo core.VFOID, enabled bool) {
	if vfo == core.VFO1 {
		return
	}
	if v.vfo2Enabled == enabled {
		return
	}
	if v.onVFO2Enabled != nil {
		v.onVFO2Enabled(enabled)
		return
	}
	v.setVFO2Enabled(enabled)
}

func (v *entryView) setVFO2Enabled(enabled bool) {
	v.vfo2Enabled = enabled
	widgets := v.vfo[core.VFO2]
	if widgets.vfoContainer != nil {
		widgets.vfoContainer.SetVisible(enabled)
	}
	v.applyIncrementalTuningVisibility(core.VFO2, core.XIT)
	v.applyIncrementalTuningVisibility(core.VFO2, core.RIT)
	if widgets.txIndicator != nil {
		widgets.txIndicator.SetVisible(enabled)
	}
	if widgets.serialClaimLabel != nil {
		widgets.serialClaimLabel.SetVisible(enabled)
	}
	if widgets.callsign != nil {
		widgets.callsign.SetVisible(enabled)
		widgets.callsign.SetEnabled(enabled)
	}
	for _, f := range widgets.theirExchangeFields {
		f.SetVisible(enabled)
		f.SetEnabled(enabled)
	}
	if widgets.logButton != nil {
		widgets.logButton.SetVisible(enabled)
	}
	if widgets.clearButton != nil {
		widgets.clearButton.SetVisible(enabled)
	}
	if widgets.messageLabel != nil {
		widgets.messageLabel.SetVisible(enabled)
	}
}
