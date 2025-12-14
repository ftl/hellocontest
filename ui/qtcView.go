package ui

import (
	"fmt"
	"strconv"

	"github.com/ftl/hamradio/callsign"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

const (
	qtcActivePhaseClass style.Class = "active-phase"

	seriesName    = "series"
	timestampName = "timestamp"
	callsignName  = "callsign"
	exchangeName  = "exchange"
)

type QTCController interface {
	StartAction()
	ConfirmStart()
	HeaderAction()
	ConfirmHeader()
	DataAction()
	ConfirmData()

	DoAction()
	DoConfirm()
	RepeatLastTransmission()

	Enter(string)
	GotoNextField()
	SetActiveField(core.QTCField)

	CompleteQTCSeries()
	AbortQTCSeries()
}

type qtcView struct {
	controller QTCController
	mode       core.QTCMode

	// widgets
	root                *gtk.Grid
	startHeadingLabel   *gtk.Label
	theirCallLabel      *gtk.Label
	startActionButton   *gtk.Button
	confirmStartButton  *gtk.Button
	headerHeadingLabel  *gtk.Label
	seriesEntry         *gtk.Entry
	headerActionButton  *gtk.Button
	confirmHeaderButton *gtk.Button
	dataHeadingLabel    *gtk.Label
	qtcTable            *qtcTable
	qtcTimeEntry        *gtk.Entry
	qtcCallEntry        *gtk.Entry
	qtcExchangeEntry    *gtk.Entry
	dataActionButton    *gtk.Button
	confirmDataButton   *gtk.Button
}

func newQTCView(controller QTCController, mode core.QTCMode) *qtcView {
	var (
		modeText string
		sendText string
	)
	if mode == core.ProvideQTC {
		modeText = "Offer QTC"
		sendText = "Send"
	} else {
		modeText = "Request QTC"
		sendText = "AGN"
	}

	result := &qtcView{
		controller: controller,
		mode:       mode,
	}

	contentGrid, _ := gtk.GridNew()
	contentGrid.SetVExpand(true)
	contentGrid.SetColumnSpacing(COLUMN_SPACING)
	contentGrid.SetRowSpacing(ROW_SPACING)
	contentGrid.SetMarginStart(MARGIN)
	contentGrid.SetMarginEnd(MARGIN)

	result.theirCallLabel, _ = gtk.LabelNew("") // the actual text is set in SetHeader
	result.theirCallLabel.SetHAlign(gtk.ALIGN_CENTER)
	contentGrid.Attach(result.theirCallLabel, 0, 0, 5, 1)

	result.startHeadingLabel = buildHeaderLabel(contentGrid, 1, "1. Start")
	startExchangeLabel, _ := gtk.LabelNew(modeText)
	startExchangeLabel.SetHAlign(gtk.ALIGN_START)
	contentGrid.Attach(startExchangeLabel, 0, 2, 3, 1)
	result.startActionButton, _ = gtk.ButtonNewWithLabel("Send")
	result.startActionButton.SetHAlign(gtk.ALIGN_FILL)
	result.startActionButton.Connect("clicked", controller.StartAction)
	result.startActionButton.Connect("key_press_event", result.onButtonKeyPress)
	contentGrid.Attach(result.startActionButton, 3, 2, 1, 1)
	result.confirmStartButton, _ = gtk.ButtonNewWithLabel("QRV")
	result.confirmStartButton.SetHAlign(gtk.ALIGN_FILL)
	result.confirmStartButton.Connect("clicked", controller.ConfirmStart)
	result.confirmStartButton.Connect("key_press_event", result.onButtonKeyPress)
	contentGrid.Attach(result.confirmStartButton, 4, 2, 1, 1)

	result.headerHeadingLabel = buildHeaderLabel(contentGrid, 3, "2. Header")
	seriesLabel, _ := gtk.LabelNew("Series/QTC Count")
	seriesLabel.SetHAlign(gtk.ALIGN_START)
	contentGrid.Attach(seriesLabel, 0, 4, 1, 1)
	result.seriesEntry, _ = gtk.EntryNew()
	result.seriesEntry.SetName(seriesName)
	result.seriesEntry.SetHExpand(true)
	result.seriesEntry.SetSizeRequest(200, 0)
	result.seriesEntry.SetSensitive(true)
	result.seriesEntry.SetEditable(mode == core.ReceiveQTC)
	result.addEntryEventHandlers(&result.seriesEntry.Widget)
	contentGrid.Attach(result.seriesEntry, 1, 4, 2, 1)

	result.headerActionButton, _ = gtk.ButtonNewWithLabel(sendText)
	result.headerActionButton.SetHAlign(gtk.ALIGN_FILL)
	result.headerActionButton.Connect("clicked", controller.HeaderAction)
	result.headerActionButton.Connect("key_press_event", result.onButtonKeyPress)
	contentGrid.Attach(result.headerActionButton, 3, 4, 1, 1)
	result.confirmHeaderButton, _ = gtk.ButtonNewWithLabel("R")
	result.confirmHeaderButton.SetHAlign(gtk.ALIGN_FILL)
	result.confirmHeaderButton.Connect("clicked", controller.ConfirmHeader)
	result.confirmHeaderButton.Connect("key_press_event", result.onButtonKeyPress)
	contentGrid.Attach(result.confirmHeaderButton, 4, 4, 1, 1)

	result.dataHeadingLabel = buildHeaderLabel(contentGrid, 6, "3. QTCs")
	result.qtcTable = newQTCTable()
	contentGrid.Attach(result.qtcTable.Table(), 0, 7, 3, 4)
	result.dataActionButton, _ = gtk.ButtonNewWithLabel(sendText)
	result.dataActionButton.SetHAlign(gtk.ALIGN_FILL)
	result.dataActionButton.SetVAlign(gtk.ALIGN_START)
	result.dataActionButton.SetVExpand(false)
	result.dataActionButton.Connect("clicked", controller.DataAction)
	result.dataActionButton.Connect("key_press_event", result.onButtonKeyPress)
	result.confirmDataButton, _ = gtk.ButtonNewWithLabel("R")
	result.confirmDataButton.SetHAlign(gtk.ALIGN_FILL)
	result.confirmDataButton.SetVAlign(gtk.ALIGN_START)
	result.confirmDataButton.SetVExpand(false)
	result.confirmDataButton.Connect("clicked", controller.ConfirmData)
	result.confirmDataButton.Connect("key_press_event", result.onButtonKeyPress)

	if mode == core.ProvideQTC {
		contentGrid.Attach(result.dataActionButton, 3, 7, 1, 1)
		contentGrid.Attach(result.confirmDataButton, 4, 7, 1, 1)
	} else {
		result.qtcTimeEntry, _ = gtk.EntryNew()
		result.qtcTimeEntry.SetName(timestampName)
		result.qtcTimeEntry.SetSizeRequest(75, 0)
		result.qtcTimeEntry.SetPlaceholderText("Time")
		result.addEntryEventHandlers(&result.qtcTimeEntry.Widget)
		contentGrid.Attach(result.qtcTimeEntry, 0, 11, 1, 1)
		result.qtcCallEntry, _ = gtk.EntryNew()
		result.qtcCallEntry.SetName(callsignName)
		result.qtcCallEntry.SetSizeRequest(150, 0)
		result.qtcCallEntry.SetPlaceholderText("Call")
		result.addEntryEventHandlers(&result.qtcCallEntry.Widget)
		contentGrid.Attach(result.qtcCallEntry, 1, 11, 1, 1)
		result.qtcExchangeEntry, _ = gtk.EntryNew()
		result.qtcExchangeEntry.SetName(exchangeName)
		result.qtcExchangeEntry.SetSizeRequest(75, 0)
		result.qtcExchangeEntry.SetPlaceholderText("Exch.")
		result.addEntryEventHandlers(&result.qtcExchangeEntry.Widget)
		contentGrid.Attach(result.qtcExchangeEntry, 2, 11, 1, 1)

		contentGrid.Attach(result.dataActionButton, 3, 11, 1, 1)
		contentGrid.Attach(result.confirmDataButton, 4, 11, 1, 1)
	}

	result.root = contentGrid

	return result
}

func (v *qtcView) setHeader(theirCall callsign.Callsign, qtcHeader core.QTCHeader) {
	v.theirCallLabel.SetText(fmt.Sprintf("Exchanging QTCs with %s", theirCall.String()))
	v.seriesEntry.SetText(qtcHeader.String())
}

func (v *qtcView) setQTCs(qtcs []core.QTC) {
	v.qtcTable.ShowQTCs(qtcs)
}

func (v *qtcView) setQTC(index int, qtc core.QTC) {
	v.qtcTable.UpdateQTC(index, qtc)
}

func (v *qtcView) enableWidgets(phase core.QTCWorkflowPhase) {
	v.startActionButton.SetSensitive(phase == core.QTCStart)
	v.confirmStartButton.SetSensitive(phase == core.QTCStart)

	v.seriesEntry.SetSensitive(phase == core.QTCExchangeHeader)
	v.headerActionButton.SetSensitive(phase == core.QTCExchangeHeader)
	v.confirmHeaderButton.SetSensitive(phase == core.QTCExchangeHeader)

	v.qtcTable.Table().SetSensitive(phase == core.QTCExchangeData)
	if v.qtcTimeEntry != nil {
		v.qtcTimeEntry.SetSensitive(phase == core.QTCExchangeData)
	}
	if v.qtcCallEntry != nil {
		v.qtcCallEntry.SetSensitive(phase == core.QTCExchangeData)
	}
	if v.qtcExchangeEntry != nil {
		v.qtcExchangeEntry.SetSensitive(phase == core.QTCExchangeData)
	}
	v.dataActionButton.SetSensitive(phase == core.QTCExchangeData)
	v.confirmDataButton.SetSensitive(phase == core.QTCExchangeData)
}

func (v *qtcView) focusStart() {
	style.AddClass(&v.startHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.headerHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.dataHeadingLabel.Widget, qtcActivePhaseClass)
	v.enableWidgets(core.QTCStart)
	v.confirmStartButton.GrabFocus()
}

func (v *qtcView) focusHeader() {
	style.RemoveClass(&v.startHeadingLabel.Widget, qtcActivePhaseClass)
	style.AddClass(&v.headerHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.dataHeadingLabel.Widget, qtcActivePhaseClass)
	v.enableWidgets(core.QTCExchangeHeader)
	v.confirmHeaderButton.GrabFocus()
}

func (v *qtcView) focusData() {
	style.RemoveClass(&v.startHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.headerHeadingLabel.Widget, qtcActivePhaseClass)
	style.AddClass(&v.dataHeadingLabel.Widget, qtcActivePhaseClass)
	v.enableWidgets(core.QTCExchangeData)
	v.confirmDataButton.GrabFocus()
}

func (v *qtcView) focusEntry(field core.QTCField) {
	switch field {
	case core.QTCHeaderField:
		v.seriesEntry.GrabFocus()
	case core.QTCTimestampField:
		v.qtcTimeEntry.GrabFocus()
	case core.QTCCallsignField:
		v.qtcCallEntry.GrabFocus()
	case core.QTCExchangeField:
		v.qtcExchangeEntry.GrabFocus()
	}
}

func (v *qtcView) clearExchangeEntry() {
	v.qtcTimeEntry.SetText("")
	v.qtcCallEntry.SetText("")
	v.qtcExchangeEntry.SetText("")
}

func (v *qtcView) focusNone() {
	style.RemoveClass(&v.startHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.headerHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.dataHeadingLabel.Widget, qtcActivePhaseClass)
	v.enableWidgets(core.QTCNone)
	v.qtcTable.ClearSelection()
}

func (v *qtcView) focusQTC(index int) {
	v.qtcTable.SelectRow(index)
	v.confirmDataButton.GrabFocus()
}

func (v *qtcView) addEntryEventHandlers(w *gtk.Widget) {
	w.Connect("key_press_event", v.onEntryKeyPress)
	w.Connect("focus_in_event", v.onEntryFocusIn)
	w.Connect("focus_out_event", v.onEntryFocusOut)
	w.Connect("changed", v.onEntryChanged)
}

func (v *qtcView) onButtonKeyPress(_ any, event *gdk.Event) bool {
	keyEvent := gdk.EventKeyNewFromEvent(event)
	state := gdk.ModifierType(keyEvent.State())
	shift := state&gdk.SHIFT_MASK != 0
	key := keyEvent.KeyVal()

	switch key {
	case gdk.KEY_Tab, gdk.KEY_ISO_Left_Tab:
		if shift {
			// goto previous field
		} else {
			v.controller.GotoNextField()
		}
		return true
	case gdk.KEY_Return:
		v.controller.DoConfirm()
		return true
	case gdk.KEY_question:
		v.controller.DoAction()
		return true
	case gdk.KEY_equal:
		v.controller.RepeatLastTransmission()
		return true
	default:
		return false
	}
}

func (v *qtcView) onEntryKeyPress(_ any, event *gdk.Event) bool {
	keyEvent := gdk.EventKeyNewFromEvent(event)
	state := gdk.ModifierType(keyEvent.State())
	ctrl := state&gdk.CONTROL_MASK != 0
	alt := state&gdk.MOD1_MASK != 0 // MOD1 = ALT right
	shift := state&gdk.SHIFT_MASK != 0
	key := keyEvent.KeyVal()

	_ = ctrl
	_ = alt
	_ = shift

	switch key {
	// TODO: select a certain QTC?
	case gdk.KEY_0, gdk.KEY_1, gdk.KEY_2, gdk.KEY_3, gdk.KEY_4, gdk.KEY_5, gdk.KEY_6, gdk.KEY_7, gdk.KEY_8, gdk.KEY_9:
		if alt {
			var index int
			if key == gdk.KEY_0 {
				index = 9
			} else {
				index = int(key - gdk.KEY_1)
			}
			_ = index
			// v.controller.SetActiveQTC(index)
			return true
		} else {
			return false
		}
	case gdk.KEY_Tab, gdk.KEY_ISO_Left_Tab:
		if shift {
			// goto previous field
		} else {
			v.controller.GotoNextField()
		}
		return true
	case gdk.KEY_space:
		if shift {
			// goto previous field
		} else {
			v.controller.GotoNextField()
		}
		return true
	case gdk.KEY_Return:
		v.controller.DoConfirm()
		return true
	case gdk.KEY_question:
		v.controller.DoAction()
		return true
	case gdk.KEY_equal:
		v.controller.RepeatLastTransmission()
		return true
	default:
		return false
	}
}

func (v *qtcView) onEntryFocusIn(widget any, _ *gdk.Event) bool {
	entry, ok := widget.(*gtk.Entry)
	if !ok {
		return false
	}

	field := v.widgetToField(&entry.Widget)
	v.controller.SetActiveField(field)

	return false
}

func (v *qtcView) onEntryFocusOut(widget any, _ *gdk.Event) bool {
	if entry, ok := widget.(*gtk.Entry); ok {
		entry.SelectRegion(0, 0)
	}
	return false
}

func (v *qtcView) onEntryChanged(widget any) bool {
	entry, ok := widget.(*gtk.Entry)
	if !ok {
		return false
	}

	text, _ := entry.GetText()
	v.controller.Enter(text)

	return false
}

func (v *qtcView) widgetToField(widget *gtk.Widget) core.QTCField {
	name, _ := widget.GetName()
	switch name {
	case seriesName:
		return core.QTCHeaderField
	case timestampName:
		return core.QTCTimestampField
	case callsignName:
		return core.QTCCallsignField
	case exchangeName:
		return core.QTCExchangeField
	default:
		return core.QTCNoneField
	}
}

// qtcTable

const (
	qtcColumnNumber int = iota
	qtcColumnTime
	qtcColumnCall
	qtcColumnExchange
	qtcColumnConfirmed

	qtcColumnCount
)

type qtcTable struct {
	table        *gtk.TreeView
	tableContent *gtk.ListStore

	qtcs []core.QTC
}

func newQTCTable() *qtcTable {
	result := &qtcTable{
		tableContent: createQTCListStore(qtcColumnCount),
	}

	result.table, _ = gtk.TreeViewNew()
	result.table.SetHExpand(true)
	result.table.SetVExpand(true)
	result.table.SetHAlign(gtk.ALIGN_FILL)
	result.table.SetVAlign(gtk.ALIGN_FILL)
	result.table.SetCanFocus(false)
	result.table.SetModel(result.tableContent)
	result.table.AppendColumn(createQTCColumn("#", qtcColumnNumber))
	result.table.AppendColumn(createQTCColumn("Time", qtcColumnTime))
	result.table.AppendColumn(createQTCColumn("Call", qtcColumnCall))
	result.table.AppendColumn(createQTCColumn("Exch.", qtcColumnExchange))
	result.table.AppendColumn(createQTCColumn("Cfm.", qtcColumnConfirmed))

	return result
}

func createQTCListStore(columnCount int) *gtk.ListStore {
	types := make([]glib.Type, columnCount)
	for i := range types {
		types[i] = glib.TYPE_STRING // TODO: use better fitting types?
	}
	result, _ := gtk.ListStoreNew(types...)
	return result
}

func createQTCColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, _ := gtk.CellRendererTextNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "markup", id)
	return column
}

func (t *qtcTable) Table() *gtk.TreeView {
	return t.table
}

func (t *qtcTable) SelectRow(index int) {
	row, _ := t.tableContent.GetIterFromString(strconv.Itoa(index))
	path, _ := t.tableContent.GetPath(row)
	selection, _ := t.table.GetSelection()
	selection.SelectPath(path)
}

func (t *qtcTable) ClearSelection() {
	selection, _ := t.table.GetSelection()
	selection.UnselectAll()
}

func (t *qtcTable) ShowQTCs(qtcs []core.QTC) {
	t.qtcs = qtcs
	t.showInTable(qtcs)
}

func (t *qtcTable) AppendQTC(qtc core.QTC) {
	t.qtcs = append(t.qtcs, qtc)
	row := t.tableContent.Append()
	t.fillRow(row, len(t.qtcs)-1, qtc)
}

func (t *qtcTable) UpdateQTC(index int, qtc core.QTC) {
	if index < 0 {
		return
	}
	if index >= len(t.qtcs) {
		t.AppendQTC(qtc)
	}

	row, _ := t.tableContent.GetIterFromString(strconv.Itoa(index))
	t.fillRow(row, index, qtc)
}

func (t *qtcTable) showInTable(qtcs []core.QTC) {
	t.tableContent.Clear()
	for i, qtc := range qtcs {
		row := t.tableContent.Append()
		t.fillRow(row, i, qtc)
	}
}

func (t *qtcTable) fillRow(row *gtk.TreeIter, index int, qtc core.QTC) {
	columns := []int{
		qtcColumnNumber,
		qtcColumnTime,
		qtcColumnCall,
		qtcColumnExchange,
		qtcColumnConfirmed,
	}

	values := []any{
		strconv.Itoa(index + 1),
		qtc.QTCTime.String(),
		qtc.QTCCallsign.String(),
		qtc.QTCNumber.String(),
		boolToCheckmark(qtc.Confirmed),
	}

	t.tableContent.Set(row, columns, values)
}
