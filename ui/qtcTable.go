package ui

import (
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

type qtcTable struct {
	widget *qtlib.QTableView
	model  *qtlib.QStandardItemModel
}

func newQTCTable() *qtcTable {
	t := &qtcTable{}

	headers := []string{"UTC", "Callsign", "Band", "Mode", "K", "Hdr", "QTC Time", "QTC Call", "QTC Exch."}
	t.model = qtlib.NewQStandardItemModel2(0, len(headers))
	t.model.SetHorizontalHeaderLabels(headers)

	t.widget = qtlib.NewQTableView2()
	t.widget.SetModel(t.model.QAbstractItemModel)
	ConfigureReadOnlyTable(t.widget)

	SetColumnSampleWidth(t.widget, 0, "00:00 ")
	SetColumnSampleWidth(t.widget, 2, "00000")
	SetColumnSampleWidth(t.widget, 3, "00000")
	t.widget.HorizontalHeader().SetSectionResizeMode2(4, qtlib.QHeaderView__ResizeToContents)
	SetColumnSampleWidth(t.widget, 6, "QTC Time")

	return t
}

func (t *qtcTable) QTCAdded(qtc core.QTC) {
	row := t.model.RowCount(qtlib.NewQModelIndex())
	items := []*qtlib.QStandardItem{
		qtlib.NewQStandardItem2(qtc.Timestamp.In(time.UTC).Format("15:04")),
		qtlib.NewQStandardItem2(qtc.TheirCallsign.String()),
		qtlib.NewQStandardItem2(qtc.Band.String()),
		qtlib.NewQStandardItem2(qtc.Mode.String()),
		qtlib.NewQStandardItem2(KindToString(qtc.Kind)),
		qtlib.NewQStandardItem2(qtc.Header.String()),
		qtlib.NewQStandardItem2(qtc.QTCTime.String()),
		qtlib.NewQStandardItem2(qtc.QTCCallsign.String()),
		qtlib.NewQStandardItem2(qtc.QTCNumber.String()),
	}
	for col, item := range items {
		t.model.SetItem(row, col, item)
	}

	index := t.model.Index(row, 0, qtlib.NewQModelIndex())
	t.widget.ScrollTo(index, qtlib.QAbstractItemView__EnsureVisible)
}

func (t *qtcTable) QTCsCleared() {
	ClearTableRows(t.model)
}

func (t *qtcTable) QTCRowSelected(index int) {
	SelectTableRow(t.widget, index, t.model)
}
