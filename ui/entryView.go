package ui

import (
	"fmt"
	"strings"
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

const parrot = "🦜"

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
	SetXITActive(bool)

	EnterPressed()
	Log()
	Clear()

	LogVFO(core.VFOID)
	ClearVFO(core.VFOID)
}

type entryView struct {
	root *qtlib.QWidget // root widget containing the grid layout

	utcLabel    *qtlib.QLabel
	myCallLabel *qtlib.QLabel

	vfoLabel       *qtlib.QLabel
	topSeparator   *qtlib.QFrame
	frequencyLabel *qtlib.QLabel
	txIndicator    *qtlib.QLabel
	messageLabel   *qtlib.QLabel

	vfo2Label             *qtlib.QLabel
	vfo2FrequencyLabel    *qtlib.QLabel
	vfo2Band              *qtlib.QComboBox
	vfo2Mode              *qtlib.QComboBox
	vfo2XITIndicator      *qtlib.QLabel
	vfo2TXIndicator       *qtlib.QLabel
	vfo2BandModeContainer *qtlib.QWidget

	vfo2TheirLabel          *qtlib.QLabel
	vfo2Callsign            *qtlib.QLineEdit
	vfo2TheirExchangeFields []*qtlib.QLineEdit
	vfo2LogButton           *qtlib.QPushButton
	vfo2ClearButton         *qtlib.QPushButton

	theirLabel        *qtlib.QLabel
	callsign          *qtlib.QLineEdit
	bandModeContainer *qtlib.QWidget
	band              *qtlib.QComboBox
	mode              *qtlib.QComboBox
	xit               *qtlib.QCheckBox

	myExchangeFields    []*qtlib.QLineEdit
	theirExchangeFields []*qtlib.QLineEdit

	logButton   *qtlib.QPushButton
	clearButton *qtlib.QPushButton

	vfo2MessageLabel *qtlib.QLabel

	theirEntryStyle string
	vfo2Enabled     bool

	ignoreInput bool
	isDuplicate bool
	isEditing   bool
	controller  EntryController
}

func newEntryView() *entryView {
	v := &entryView{}

	// Row 0: UTC label, myCall label, myExchanges container (cols 2-3, span 2)
	v.utcLabel = qtlib.NewQLabel3("00:00")

	v.myCallLabel = qtlib.NewQLabel3("DL0ABC")

	// Row 1: Horizontal separator (span all 6 columns)
	v.topSeparator = qtlib.NewQFrame2()
	v.topSeparator.SetFrameShape(qtlib.QFrame__HLine)
	v.topSeparator.SetFrameShadow(qtlib.QFrame__Sunken)

	// Row 2: "VFO:" label, frequency label, band combo, mode combo, XIT checkbox, TX indicator
	v.vfoLabel = qtlib.NewQLabel3("VFO 1:")

	v.frequencyLabel = qtlib.NewQLabel3("- kHz")
	v.frequencyLabel.SetObjectName(*qtlib.NewQAnyStringView3("frequencyLabel"))
	v.frequencyLabel.SetAlignment(qtlib.AlignTrailing | qtlib.AlignVCenter)

	v.bandModeContainer = qtlib.NewQWidget2()
	bandModeLayout := qtlib.NewQHBoxLayout(v.bandModeContainer)

	v.band = qtlib.NewQComboBox2()
	v.band.SetObjectName(*qtlib.NewQAnyStringView3("bandCombo"))
	bandModeLayout.AddWidget2(v.band.QWidget, 1)
	v.mode = qtlib.NewQComboBox2()
	v.mode.SetObjectName(*qtlib.NewQAnyStringView3("modeCombo"))
	bandModeLayout.AddWidget2(v.mode.QWidget, 2)

	v.xit = qtlib.NewQCheckBox3("XIT")
	v.xit.SetObjectName(*qtlib.NewQAnyStringView3("xit"))

	v.txIndicator = qtlib.NewQLabel3("")

	// Row 3: Reserved for callinfo (later step)

	// Row 4: "Their:" label, callsign QLineEdit, theirExchanges container, Log button, Clear button
	v.theirLabel = qtlib.NewQLabel3("Their:")

	v.callsign = qtlib.NewQLineEdit2()
	v.callsign.SetObjectName(*qtlib.NewQAnyStringView3("callsignEntry"))
	v.callsign.SetPlaceholderText("Call")

	fi := qtlib.NewQFontInfo(v.callsign.Font())
	if pt := fi.PointSizeF(); pt > 0 {
		v.theirEntryStyle = fmt.Sprintf("QLineEdit { font-size: %.1fpt; }", pt*2)
	} else if px := fi.PixelSize(); px > 0 {
		v.theirEntryStyle = fmt.Sprintf("QLineEdit { font-size: %dpx; }", px*2)
	}
	v.callsign.SetStyleSheet(v.theirEntryStyle)

	v.logButton = qtlib.NewQPushButton3("Log")
	v.logButton.SetFocusPolicy(qtlib.NoFocus)

	v.clearButton = qtlib.NewQPushButton3("Clear")
	v.clearButton.SetFocusPolicy(qtlib.NoFocus)

	// Row 7: Message label (span all 6 columns)
	v.messageLabel = qtlib.NewQLabel3("")

	// VFO2 message label
	v.vfo2MessageLabel = qtlib.NewQLabel3("")

	// Row 8: VFO2
	v.vfo2Label = qtlib.NewQLabel3("VFO 2:")
	v.vfo2FrequencyLabel = qtlib.NewQLabel3("- kHz")
	v.vfo2FrequencyLabel.SetObjectName(*qtlib.NewQAnyStringView3("vfo2FrequencyLabel"))
	v.vfo2FrequencyLabel.SetAlignment(qtlib.AlignTrailing | qtlib.AlignVCenter)

	v.vfo2BandModeContainer = qtlib.NewQWidget2()
	vfo2BandModeLayout := qtlib.NewQHBoxLayout(v.vfo2BandModeContainer)

	v.vfo2Band = qtlib.NewQComboBox2()
	v.vfo2Band.SetObjectName(*qtlib.NewQAnyStringView3("vfo2BandCombo"))
	vfo2BandModeLayout.AddWidget2(v.vfo2Band.QWidget, 1)
	v.vfo2Mode = qtlib.NewQComboBox2()
	v.vfo2Mode.SetObjectName(*qtlib.NewQAnyStringView3("vfo2ModeCombo"))
	vfo2BandModeLayout.AddWidget2(v.vfo2Mode.QWidget, 2)

	v.vfo2XITIndicator = qtlib.NewQLabel3("XIT")
	v.vfo2XITIndicator.SetObjectName(*qtlib.NewQAnyStringView3("vfo2XITIndicator"))
	v.vfo2TXIndicator = qtlib.NewQLabel3("RX")
	v.vfo2TXIndicator.SetObjectName(*qtlib.NewQAnyStringView3("vfo2TXIndicator"))

	// VFO2 input row: their label + callsign field + (later) their-exchange fields + log/clear
	v.vfo2TheirLabel = qtlib.NewQLabel3("Their:")
	v.vfo2Callsign = qtlib.NewQLineEdit2()
	v.vfo2Callsign.SetObjectName(*qtlib.NewQAnyStringView3("vfo2CallsignEntry"))
	v.vfo2Callsign.SetPlaceholderText("Call")
	if v.theirEntryStyle != "" {
		v.vfo2Callsign.SetStyleSheet(v.theirEntryStyle)
	}
	v.vfo2LogButton = qtlib.NewQPushButton3("Log")
	v.vfo2LogButton.SetFocusPolicy(qtlib.NoFocus)
	v.vfo2ClearButton = qtlib.NewQPushButton3("Clear")
	v.vfo2ClearButton.SetFocusPolicy(qtlib.NoFocus)

	// Initialize combos
	SetupBandCombo(v.band)
	SetupModeCombo(v.mode)
	SetupBandCombo(v.vfo2Band)
	SetupModeCombo(v.vfo2Mode)

	// Connect signals for static widgets
	v.connectEditSignals(v.callsign, core.VFO1, core.CallsignField, true)
	v.connectEditSignals(v.vfo2Callsign, core.VFO2, core.CallsignField, true)
	v.connectComboSignals(v.band, core.VFO1, core.BandField)
	v.connectComboSignals(v.mode, core.VFO1, core.ModeField)
	v.connectComboSignals(v.vfo2Band, core.VFO2, core.BandField)
	v.connectComboSignals(v.vfo2Mode, core.VFO2, core.ModeField)

	// Connect button/checkbox signals
	v.logButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.LogVFO(core.VFO1)
		}
	})
	v.clearButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.ClearVFO(core.VFO1)
		}
	})
	v.vfo2LogButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.LogVFO(core.VFO2)
		}
	})
	v.vfo2ClearButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.ClearVFO(core.VFO2)
		}
	})
	v.xit.OnStateChanged(func(state int) {
		if v.controller != nil {
			v.controller.SetXITActive(state != 0)
		}
	})

	return v
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

func (v *entryView) SetUTC(text string) {
	v.utcLabel.SetText(text)
}

func (v *entryView) SetMyCall(text string) {
	v.myCallLabel.SetText(text)
}

func (v *entryView) SetFrequency(vfo core.VFOID, frequency core.Frequency) {
	switch vfo {
	case core.VFO1:
		v.frequencyLabel.SetText(fmt.Sprintf("%.2f kHz", frequency/1000.0))
	case core.VFO2:
		v.vfo2FrequencyLabel.SetText(fmt.Sprintf("%.2f kHz", frequency/1000.0))
	}
}

func (v *entryView) SetCallsign(vfo core.VFOID, text string) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	switch vfo {
	case core.VFO1:
		v.callsign.SetText(text)
	case core.VFO2:
		if v.vfo2Callsign != nil {
			v.vfo2Callsign.SetText(text)
		}
	}
}

func (v *entryView) SetBand(vfo core.VFOID, text string) {
	var combo *qtlib.QComboBox
	switch vfo {
	case core.VFO1:
		combo = v.band
	case core.VFO2:
		combo = v.vfo2Band
	default:
		return
	}

	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	idx := combo.FindText(text)
	if idx >= 0 {
		combo.SetCurrentIndex(idx)
	}
}

func (v *entryView) SetMode(vfo core.VFOID, text string) {
	var combo *qtlib.QComboBox
	switch vfo {
	case core.VFO1:
		combo = v.mode
	case core.VFO2:
		combo = v.vfo2Mode
	default:
		return
	}

	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	idx := combo.FindText(text)
	if idx >= 0 {
		combo.SetCurrentIndex(idx)
	}
}

func (v *entryView) SetXITActive(vfo core.VFOID, active bool) {
	switch vfo {
	case core.VFO1:
		if v.xit.IsChecked() == active {
			return
		}
		v.xit.SetChecked(active)
	case core.VFO2:
		// TODO: handle event
	}
}

func (v *entryView) SetXIT(vfo core.VFOID, active bool, offset core.Frequency) {
	var text string
	if active {
		text = fmt.Sprintf("XIT %s", offset)
	} else {
		text = "XIT"
	}

	switch vfo {
	case core.VFO1:
		v.xit.SetText(text)
	case core.VFO2:
		v.vfo2XITIndicator.SetText(text)
	}
}

func (v *entryView) SetTXState(vfo core.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration) {
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

	switch vfo {
	case core.VFO1:
		if ptt {
			v.txIndicator.SetStyleSheet(TXIndicatorActiveStyle)
		} else {
			v.txIndicator.SetStyleSheet(TXIndicatorInactiveStyle)
		}
		v.txIndicator.SetText(text)
	case core.VFO2:
		v.vfo2TXIndicator.SetText(text)
	}
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
	i := index - 1
	var fields []*qtlib.QLineEdit
	switch vfo {
	case core.VFO1:
		fields = v.theirExchangeFields
	case core.VFO2:
		fields = v.vfo2TheirExchangeFields
	default:
		return
	}
	if i < 0 || i >= len(fields) {
		return
	}
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	fields[i].SetText(text)
}

func (v *entryView) SetExchangeFields(myExchangeFields, theirExchangeFields []core.ExchangeField) {
	v.setExchangeFields(myExchangeFields, &v.myExchangeFields, false, core.VFO1)
	v.setExchangeFields(theirExchangeFields, &v.theirExchangeFields, true, core.VFO1)
	v.setExchangeFields(theirExchangeFields, &v.vfo2TheirExchangeFields, true, core.VFO2)
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
			editField.SetStyleSheet(v.theirEntryStyle)
		}

		if vfo == core.VFO2 && !v.vfo2Enabled {
			editField.SetVisible(false)
			editField.SetEnabled(false)
		}

		v.connectEditSignals(editField, vfo, core.TheirExchangeField(i+1), isTheirRow)
		(*editFields)[i] = editField
	}
}

func (v *entryView) SetActiveField(vfo core.VFOID, field core.EntryField) {
	callsign := v.callsign
	band := v.band
	mode := v.mode
	theirExchange := v.theirExchangeFields
	if vfo == core.VFO2 {
		callsign = v.vfo2Callsign
		band = v.vfo2Band
		mode = v.vfo2Mode
		theirExchange = v.vfo2TheirExchangeFields
	}

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
	callsign := v.callsign
	theirExchange := v.theirExchangeFields
	if vfo == core.VFO2 {
		callsign = v.vfo2Callsign
		theirExchange = v.vfo2TheirExchangeFields
	}
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
	label := v.messageLabel
	if vfo == core.VFO2 {
		label = v.vfo2MessageLabel
	}
	label.SetText(fmt.Sprint(args...))
}

func (v *entryView) ClearMessage(vfo core.VFOID) {
	label := v.messageLabel
	if vfo == core.VFO2 {
		label = v.vfo2MessageLabel
	}
	label.SetText("")
}

// SetVFOEnabled toggles the visibility/enabled state of a VFO's row of widgets.
// VFO1 is always enabled. VFO2 widgets are shown/hidden as a group.
func (v *entryView) SetVFOEnabled(vfo core.VFOID, enabled bool) {
	if vfo == core.VFO1 {
		return
	}
	v.vfo2Enabled = enabled
	if v.vfo2Label != nil {
		v.vfo2Label.SetVisible(enabled)
	}
	if v.vfo2FrequencyLabel != nil {
		v.vfo2FrequencyLabel.SetVisible(enabled)
	}
	if v.vfo2BandModeContainer != nil {
		v.vfo2BandModeContainer.SetVisible(enabled)
	}
	if v.vfo2XITIndicator != nil {
		v.vfo2XITIndicator.SetVisible(enabled)
	}
	if v.vfo2TXIndicator != nil {
		v.vfo2TXIndicator.SetVisible(enabled)
	}
	if v.vfo2TheirLabel != nil {
		v.vfo2TheirLabel.SetVisible(enabled)
	}
	if v.vfo2Callsign != nil {
		v.vfo2Callsign.SetVisible(enabled)
		v.vfo2Callsign.SetEnabled(enabled)
	}
	for _, f := range v.vfo2TheirExchangeFields {
		f.SetVisible(enabled)
		f.SetEnabled(enabled)
	}
	if v.vfo2LogButton != nil {
		v.vfo2LogButton.SetVisible(enabled)
	}
	if v.vfo2ClearButton != nil {
		v.vfo2ClearButton.SetVisible(enabled)
	}
	if v.vfo2MessageLabel != nil {
		v.vfo2MessageLabel.SetVisible(enabled)
	}
}
