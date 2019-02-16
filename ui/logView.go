package ui

import (
	"log"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	columnUTC int = iota
	columnCallsign
	columnBand
	columnMode
	columnMyReport
	columnMyNumber
	columnMyXchange
	columnTheirReport
	columnTheirNumber
	columnTheirXchange
)

type logView struct {
	view *gtk.TreeView
	list *gtk.ListStore
	log  core.Log
}

func setupLogView(builder *gtk.Builder) *logView {
	result := new(logView)

	result.view = getUI(builder, "logView").(*gtk.TreeView)

	result.view.AppendColumn(createColumn("UTC", columnUTC))
	result.view.AppendColumn(createColumn("Callsign", columnCallsign))
	result.view.AppendColumn(createColumn("Band", columnBand))
	result.view.AppendColumn(createColumn("Mode", columnMode))
	result.view.AppendColumn(createColumn("My RST", columnMyReport))
	result.view.AppendColumn(createColumn("My #", columnMyNumber))
	result.view.AppendColumn(createColumn("My XChg", columnMyXchange))
	result.view.AppendColumn(createColumn("Th RST", columnTheirReport))
	result.view.AppendColumn(createColumn("Th #", columnTheirNumber))
	result.view.AppendColumn(createColumn("Th XChg", columnTheirXchange))

	var err error
	result.list, err = gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		log.Fatalf("Cannot create QSO list store: %v", err)
	}
	result.view.SetModel(result.list)
	return result
}

func createColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatalf("Cannot create text cell renderer for column %s: %v", title, err)
	}

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)
	if err != nil {
		log.Fatalf("Cannot create column %s: %v", title, err)
	}
	return column
}

func (v *logView) SetLog(log core.Log) {
	v.log = log
}

func (v *logView) UpdateAllRows(qsos []core.QSO) {
	v.list.Clear()
	for _, qso := range qsos {
		v.RowAdded(qso)
	}
}

func (v *logView) RowAdded(qso core.QSO) {
	newRow := v.list.Append()
	err := v.list.Set(newRow,
		[]int{
			columnUTC,
			columnCallsign,
			columnBand,
			columnMode,
			columnMyReport,
			columnMyNumber,
			columnMyXchange,
			columnTheirReport,
			columnTheirNumber,
			columnTheirXchange,
		},
		[]interface{}{
			qso.Time.In(time.UTC).Format("15:04"),
			qso.Callsign.String(),
			qso.Band.String(),
			qso.Mode.String(),
			qso.MyReport.String(),
			qso.MyNumber.String(),
			qso.MyXchange,
			qso.TheirReport.String(),
			qso.TheirNumber.String(),
			qso.TheirXchange,
		})
	if err != nil {
		log.Printf("Cannot add QSO row %s: %v", qso.String(), err)
	}
	path, err := v.list.GetPath(newRow)
	if err != nil {
		log.Printf("Cannot get path for list item: %s", err)
	}
	v.view.SetCursorOnCell(path, v.view.GetColumn(1), nil, false)
}
