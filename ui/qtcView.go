package ui

import (
	"strconv"

	"github.com/ftl/hamradio/callsign"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type QTCController interface {
	// TBD
}

type qtcView struct {
	controller QTCController
	mode       core.QTCMode

	// widgets
	root           *gtk.Grid
	theirCallEntry *gtk.Entry
	seriesEntry    *gtk.Entry
	qtcTable       *qtcTable
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

	buildHeaderLabel(contentGrid, 0, "Header")
	theirCallLabel, _ := gtk.LabelNew("Their Call")
	contentGrid.Attach(theirCallLabel, 0, 1, 1, 1)
	seriesLabel, _ := gtk.LabelNew("Series/QTC Count")
	contentGrid.Attach(seriesLabel, 1, 1, 1, 1)
	actionLabel, _ := gtk.LabelNew("Action")
	contentGrid.Attach(actionLabel, 2, 1, 2, 1)
	result.theirCallEntry, _ = gtk.EntryNew()
	result.theirCallEntry.SetPlaceholderText("Call")
	if mode == core.ProvideQTC {
		result.theirCallEntry.SetSensitive(false)
	} else {
		// TODO: add handler callback
	}
	contentGrid.Attach(result.theirCallEntry, 0, 2, 1, 1)
	result.seriesEntry, _ = gtk.EntryNew()
	result.seriesEntry.SetPlaceholderText("Series/QTC Count")
	if mode == core.ProvideQTC {
		result.seriesEntry.SetSensitive(false)
	} else {
		// TODO: add handler callback
	}
	contentGrid.Attach(result.seriesEntry, 1, 2, 1, 1)
	sendHeaderButton, _ := gtk.ButtonNewWithLabel("Send")
	sendHeaderButton.SetHAlign(gtk.ALIGN_END)
	contentGrid.Attach(sendHeaderButton, 2, 2, 1, 1)
	confirmHeaderButton, _ := gtk.ButtonNewWithLabel("R")
	confirmHeaderButton.SetHAlign(gtk.ALIGN_END)
	contentGrid.Attach(confirmHeaderButton, 3, 2, 1, 1)

	separator, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	separator.SetVExpand(false)
	contentGrid.Attach(separator, 0, 3, 4, 1)

	buildHeaderLabel(contentGrid, 4, "QTCs")

	result.qtcTable = newQTCTable()
	contentGrid.Attach(result.qtcTable.Table(), 0, 5, 2, 4)
	sendQTCButton, _ := gtk.ButtonNewWithLabel("Send")
	sendQTCButton.SetHAlign(gtk.ALIGN_END)
	sendQTCButton.SetVAlign(gtk.ALIGN_START)
	sendQTCButton.SetVExpand(false)
	contentGrid.Attach(sendQTCButton, 2, 5, 1, 1)
	confirmQTCButton, _ := gtk.ButtonNewWithLabel("R")
	confirmQTCButton.SetHAlign(gtk.ALIGN_END)
	confirmQTCButton.SetVAlign(gtk.ALIGN_START)
	confirmQTCButton.SetVExpand(false)
	contentGrid.Attach(confirmQTCButton, 3, 5, 1, 1)

	result.root = contentGrid

	return result
}

func (v *qtcView) setHeader(theirCall callsign.Callsign, qtcHeader core.QTCHeader) {
	v.theirCallEntry.SetText(theirCall.String())
	v.seriesEntry.SetText(qtcHeader.String())
}

func (v *qtcView) setQTCs(qtcs []core.QTC) {
	v.qtcTable.ShowQTCs(qtcs)
}

func (v *qtcView) setQTC(index int, qtc core.QTC) {
	v.qtcTable.UpdateQTC(index, qtc)
}

func (v *qtcView) focusHeader() {
	// TODO: implement
}

func (v *qtcView) focusQTC(index int) {
	// TODO: implement
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
		"", // TODO: qtc confirmed -> show check mark
	}

	t.tableContent.Set(row, columns, values)
}

func (t *qtcTable) refreshTableStyle() {
	t.showInTable(t.qtcs)
}
