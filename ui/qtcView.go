package ui

import (
	"fmt"
	"strconv"

	"github.com/ftl/hamradio/callsign"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

const (
	qtcActivePhaseClass style.Class = "active-phase"
)

type QTCController interface {
	SendStart()
	SendHeader()
	SendQTC()
	ConfirmStart()
	ConfirmHeader()
	ConfirmQTC()
	CompleteQTCSeries()
}

type qtcView struct {
	controller QTCController
	mode       core.QTCMode

	// widgets
	root                *gtk.Grid
	startHeadingLabel   *gtk.Label
	theirCallLabel      *gtk.Label
	qrvButton           *gtk.Button
	headerHeadingLabel  *gtk.Label
	seriesEntry         *gtk.Entry
	confirmHeaderButton *gtk.Button
	dataHeadingLabel    *gtk.Label
	qtcTable            *qtcTable
	confirmQTCButton    *gtk.Button
}

func newQTCView(controller QTCController, mode core.QTCMode) *qtcView {
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
	contentGrid.Attach(result.theirCallLabel, 0, 0, 4, 1)

	result.startHeadingLabel = buildHeaderLabel(contentGrid, 1, "1. Start")
	var modeText string
	switch mode {
	case core.ProvideQTC:
		modeText = "Offer QTC"
	case core.ReceiveQTC:
		modeText = "Request QTC"
	default:
		modeText = fmt.Sprintf("UNKNOWN MODE %d", mode)
	}
	startExchangeLabel, _ := gtk.LabelNew(modeText)
	startExchangeLabel.SetHAlign(gtk.ALIGN_START)
	contentGrid.Attach(startExchangeLabel, 0, 2, 2, 1)
	sendStartButton, _ := gtk.ButtonNewWithLabel("Send")
	sendStartButton.SetHAlign(gtk.ALIGN_FILL)
	sendStartButton.Connect("clicked", controller.SendStart)
	contentGrid.Attach(sendStartButton, 2, 2, 1, 1)
	result.qrvButton, _ = gtk.ButtonNewWithLabel("QRV")
	result.qrvButton.SetHAlign(gtk.ALIGN_FILL)
	result.qrvButton.Connect("clicked", controller.ConfirmStart)
	contentGrid.Attach(result.qrvButton, 3, 2, 1, 1)

	result.headerHeadingLabel = buildHeaderLabel(contentGrid, 3, "2. Header")
	result.seriesEntry = buildLabeledEntry(contentGrid, 4, "Series/QTC Count", nil) // TODO: add callback if needed
	result.seriesEntry.SetSensitive(true)
	result.seriesEntry.SetEditable(false)
	sendHeaderButton, _ := gtk.ButtonNewWithLabel("Send")
	sendHeaderButton.SetHAlign(gtk.ALIGN_FILL)
	sendHeaderButton.Connect("clicked", controller.SendHeader)
	contentGrid.Attach(sendHeaderButton, 2, 4, 1, 1)
	result.confirmHeaderButton, _ = gtk.ButtonNewWithLabel("R")
	result.confirmHeaderButton.SetHAlign(gtk.ALIGN_FILL)
	result.confirmHeaderButton.Connect("clicked", controller.ConfirmHeader)
	contentGrid.Attach(result.confirmHeaderButton, 3, 4, 1, 1)

	result.dataHeadingLabel = buildHeaderLabel(contentGrid, 6, "3. QTCs")
	result.qtcTable = newQTCTable()
	contentGrid.Attach(result.qtcTable.Table(), 0, 7, 2, 4)
	sendQTCButton, _ := gtk.ButtonNewWithLabel("Send")
	sendQTCButton.SetHAlign(gtk.ALIGN_FILL)
	sendQTCButton.SetVAlign(gtk.ALIGN_START)
	sendQTCButton.SetVExpand(false)
	sendQTCButton.Connect("clicked", controller.SendQTC)
	contentGrid.Attach(sendQTCButton, 2, 7, 1, 1)
	result.confirmQTCButton, _ = gtk.ButtonNewWithLabel("R")
	result.confirmQTCButton.SetHAlign(gtk.ALIGN_FILL)
	result.confirmQTCButton.SetVAlign(gtk.ALIGN_START)
	result.confirmQTCButton.SetVExpand(false)
	result.confirmQTCButton.Connect("clicked", controller.ConfirmQTC)
	contentGrid.Attach(result.confirmQTCButton, 3, 7, 1, 1)

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

func (v *qtcView) focusStart() {
	style.AddClass(&v.startHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.headerHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.dataHeadingLabel.Widget, qtcActivePhaseClass)
	v.qrvButton.GrabFocus()
}

func (v *qtcView) focusHeader() {
	style.RemoveClass(&v.startHeadingLabel.Widget, qtcActivePhaseClass)
	style.AddClass(&v.headerHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.dataHeadingLabel.Widget, qtcActivePhaseClass)
	v.confirmHeaderButton.GrabFocus()
}

func (v *qtcView) focusData() {
	style.RemoveClass(&v.startHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.headerHeadingLabel.Widget, qtcActivePhaseClass)
	style.AddClass(&v.dataHeadingLabel.Widget, qtcActivePhaseClass)
	v.confirmQTCButton.GrabFocus()
}

func (v *qtcView) focusEntry() {
	// TODO: implement
}

func (v *qtcView) focusNone() {
	style.RemoveClass(&v.startHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.headerHeadingLabel.Widget, qtcActivePhaseClass)
	style.RemoveClass(&v.dataHeadingLabel.Widget, qtcActivePhaseClass)
	v.qtcTable.ClearSelection()
}

func (v *qtcView) focusQTC(index int) {
	v.qtcTable.SelectRow(index)
	v.confirmQTCButton.GrabFocus()
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
	result.table.Connect("style-updated", result.refreshTableStyle)

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
	if index >= len(t.qtcs) {
		t.AppendQTC(qtc)
	}
	// TODO: update the row #index
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
		boolToCheckmark(qtc.WasTransmitted()), // TODO: transmitted != confirmed
	}

	t.tableContent.Set(row, columns, values)
}

func (t *qtcTable) refreshTableStyle() {
	t.showInTable(t.qtcs)
}
