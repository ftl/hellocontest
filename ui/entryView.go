package ui

import (
	"fmt"
	"strings"
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

const (
	parrot                 = "🦜"
	txIndicatorText        = "TX"
	txIndicatorWidthSample = parrot + ": 10m0s"
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

	// radio
	vfoContainer     *qtlib.QWidget
	vfoContainerName string
	vfoLabel         *qtlib.QLabel
	frequencyLabel   *qtlib.QLabel
	band             *qtlib.QComboBox
	mode             *qtlib.QComboBox
	xit              *qtlib.QCheckBox
	rit              *qtlib.QCheckBox
	txIndicator      *qtlib.QLabel

	// serial claim
	serialClaimLabel *qtlib.QLabel

	// QSO
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
	activeVFO   core.VFOID

	parrotActive [core.VFOCount]bool

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
	v.registerVFOWidgets(core.VFO1, &v.vfo[core.VFO1])

	v.vfo[core.VFO2] = newEntryVFOWidgets("vfo2", "VFO 2")
	v.registerVFOWidgets(core.VFO2, &v.vfo[core.VFO2])

	v.SetTXVFO(core.VFO1)

	return v
}

func newEntryVFOWidgets(prefix string, vfoName string) entryVFOWidgets {
	w := entryVFOWidgets{}

	w.topSeparator = qtlib.NewQFrame2()
	w.topSeparator.SetFrameShape(qtlib.QFrame__HLine)
	w.topSeparator.SetFrameShadow(qtlib.QFrame__Sunken)

	w.vfoContainer = qtlib.NewQWidget2()
	w.vfoContainerName = prefix + "VFOContainer"
	w.vfoContainer.SetObjectName(*qtlib.NewQAnyStringView3(w.vfoContainerName))
	w.vfoContainer.SetAttribute(qtlib.WA_StyledBackground)
	// w.vfoContainer.SetStyleSheet(fmt.Sprintf("QWidget#%sVFOContainer { background-color: red; }", prefix))
	vfoContainerLayout := qtlib.NewQHBoxLayout(w.vfoContainer)
	w.vfoLabel = qtlib.NewQLabel3(vfoName)
	w.vfoLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "Label"))
	w.vfoLabel.SetAlignment(qtlib.AlignLeading | qtlib.AlignVCenter)
	setFixedTextWidth(w.vfoLabel.QWidget, "VFO 2 S&P", RoundedLabelPadding)
	vfoContainerLayout.AddWidget(w.vfoLabel.QWidget)

	w.txIndicator = qtlib.NewQLabel3(txIndicatorText)
	w.txIndicator.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "TX"))
	w.txIndicator.SetStyleSheet(TXIndicatorInactiveStyle)
	w.txIndicator.SetAlignment(qtlib.AlignCenter | qtlib.AlignVCenter)
	setFixedBoldTextWidth(w.txIndicator.QWidget, txIndicatorWidthSample, RoundedLabelPadding)
	retainSizeWhenHidden(w.txIndicator.QWidget)
	vfoContainerLayout.AddWidget(w.txIndicator.QWidget)

	w.frequencyLabel = qtlib.NewQLabel3("- kHz")
	w.frequencyLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "FrequencyLabel"))
	w.frequencyLabel.SetAlignment(qtlib.AlignTrailing | qtlib.AlignVCenter)
	setFixedTextWidth(w.frequencyLabel.QWidget, "999999.99 kHz", 0)
	vfoContainerLayout.AddWidget(w.frequencyLabel.QWidget)

	w.band = qtlib.NewQComboBox2()
	w.band.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "BandCombo"))
	setupBandCombo(w.band)
	vfoContainerLayout.AddWidget(w.band.QWidget)

	w.mode = qtlib.NewQComboBox2()
	w.mode.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "ModeCombo"))
	setupModeCombo(w.mode)
	vfoContainerLayout.AddWidget(w.mode.QWidget)

	w.xit = qtlib.NewQCheckBox3("XIT")
	w.xit.SetFocusPolicy(qtlib.NoFocus)
	w.xit.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "XIT"))
	w.xit.SetVisible(false)
	vfoContainerLayout.AddWidget(w.xit.QWidget)

	w.rit = qtlib.NewQCheckBox3("RIT")
	w.rit.SetFocusPolicy(qtlib.NoFocus)
	w.rit.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "RIT"))
	w.rit.SetVisible(false)
	vfoContainerLayout.AddWidget(w.rit.QWidget)

	vfoContainerLayout.AddStretch()

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

func (v *entryView) registerVFOWidgets(vfo core.VFOID, widgets *entryVFOWidgets) {
	v.connectEditSignals(widgets.callsign, vfo, core.CallsignField, true)
	v.connectComboSignals(widgets.band, vfo, core.BandField)
	v.connectComboSignals(widgets.mode, vfo, core.ModeField)
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
			v.incrementalTuningInput.SetIncrementalTuningActive(vfo, core.XIT, state != 0)
		}
	})
	widgets.rit.OnStateChanged(func(state int) {
		if v.incrementalTuningInput != nil {
			v.incrementalTuningInput.SetIncrementalTuningActive(vfo, core.RIT, state != 0)
		}
	})
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
		text := fmt.Sprintf("%s claimed", serial.String())
		if committed {
			text = fmt.Sprintf("<b>%s committed</b>", serial.String())
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

func (v *entryView) VFOIncrementalTuningActiveChanged(vfo core.VFOID, kind core.IncrementalTuningKind, active bool) {
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

func (v *entryView) VFOIncrementalTuningVisibilityChanged(vfo core.VFOID, kind core.IncrementalTuningKind, visible bool) {
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

	v.parrotActive[vfo] = parrotActive
	v.applyTXIndicatorVisibility()

	text := txIndicatorText
	if parrotActive {
		text = parrot
		if parrotTimeLeft > 0 {
			text += fmt.Sprintf(": %v", parrotTimeLeft)
		}
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

func (v *entryView) SetExchangeFields(myExchangeFields, theirExchangeFields []core.ExchangeField, generateSerialExchange bool) {
	v.setExchangeFields(myExchangeFields, &v.myExchangeFields, false, core.VFO1)
	v.setExchangeFields(theirExchangeFields, &v.vfo[core.VFO1].theirExchangeFields, true, core.VFO1)
	v.setExchangeFields(theirExchangeFields, &v.vfo[core.VFO2].theirExchangeFields, true, core.VFO2)
	v.setSerialClaimLabelsVisible(generateSerialExchange)
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

func (v *entryView) setSerialClaimLabelsVisible(visible bool) {
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

func (v *entryView) SetVFOWorkmode(vfo core.VFOID, workmode core.Workmode) {
	v.vfoWorkmode[vfo] = workmode
	v.updateVFOLabel(vfo)
}

func (v *entryView) SetTXVFO(vfo core.VFOID) {
	v.txVFO = vfo
	v.applyTXIndicatorVisibility()
}

func (v *entryView) applyTXIndicatorVisibility() {
	for id := core.VFOID(0); id < core.VFOCount; id++ {
		widget := v.vfo[id].txIndicator
		if widget == nil {
			continue
		}
		widget.SetVisible((v.vfo2Enabled && id == v.txVFO) || v.parrotActive[id])
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
		text += " S&P"
	}
	label.SetTextFormat(qtlib.PlainText)
	label.SetText(text)
}

func (v *entryView) SetActiveVFO(vfo core.VFOID) {
	v.activeVFO = vfo
	v.applyActiveVFOStyle()
}

func (v *entryView) applyActiveVFOStyle() {
	// TODO: use a property with a selective style
	for id := core.VFOID(0); id < core.VFOCount; id++ {
		widgets := v.vfo[id]
		if widgets.vfoContainer == nil {
			continue
		}
		style := VFOInactiveStyle
		if v.vfo2Enabled && id == v.activeVFO {
			style = fmt.Sprintf(VFOActiveStyle, widgets.vfoContainerName)
		}
		widgets.vfoContainer.SetStyleSheet(style)
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
	v.applyTXIndicatorVisibility()
	v.applyActiveVFOStyle()
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
