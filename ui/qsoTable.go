package ui

import (
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

// QSOListController is implemented by *logbook.QSOList.
type QSOListController interface {
	GetExchangeFields() ([]core.ExchangeField, []core.ExchangeField)
	SelectRow(int)
}

type qsoTable struct {
	table *qtlib.QTableView
	model *qtlib.QStandardItemModel

	myExchangeFields    []core.ExchangeField
	theirExchangeFields []core.ExchangeField

	controller QSOListController
}

func newQSOTable() *qsoTable {
	v := &qsoTable{}
	v.model = qtlib.NewQStandardItemModel2(0, 0)
	v.table = qtlib.NewQTableView2()
	v.table.SetModel(v.model.QAbstractItemModel)
	ConfigureReadOnlyTable(v.table)
	return v
}

func (t *qsoTable) SetQSOListController(controller QSOListController) {
	t.controller = controller
	t.ExchangeFieldsChanged(t.controller.GetExchangeFields())
}

func (t *qsoTable) ExchangeFieldsChanged(myFields, theirFields []core.ExchangeField) {
	t.myExchangeFields = myFields
	t.theirExchangeFields = theirFields
	t.rebuildColumns()
}

func (t *qsoTable) rebuildColumns() {
	headers := []string{"UTC", "Callsign", "Band", "Mode"}

	var narrowColumns []int

	firstMyExchange := len(headers)
	for i, f := range t.myExchangeFields {
		var columnName string
		if len(f.Properties) == 1 {
			columnName = f.Short
		} else {
			columnName = "Exch"
		}
		headers = append(headers, "My "+columnName)
		if isNarrowExchangeField(f) {
			narrowColumns = append(narrowColumns, firstMyExchange+i)
		}
	}
	firstTheirExchange := len(headers)
	for i, f := range t.theirExchangeFields {
		var columnName string
		if len(f.Properties) == 1 {
			columnName = f.Short
		} else {
			columnName = "Exch"
		}
		headers = append(headers, "Th "+columnName)
		if isNarrowExchangeField(f) {
			narrowColumns = append(narrowColumns, firstTheirExchange+i)
		}
	}

	headers = append(headers, "Pts", "Mult", "D", "WM")

	t.model.Clear()
	t.model.SetHorizontalHeaderLabels(headers)

	pointsColumn := len(headers) - 4
	multiColumn := len(headers) - 3
	duplicateColumn := len(headers) - 2
	workmodeColumn := len(headers) - 1
	header := t.table.HorizontalHeader()
	header.SetSectionResizeMode2(duplicateColumn, qtlib.QHeaderView__ResizeToContents)
	header.SetSectionResizeMode2(workmodeColumn, qtlib.QHeaderView__ResizeToContents)

	SetColumnSampleWidth(t.table, 0, "00:00 ")
	SetColumnSampleWidth(t.table, 2, "00000")
	SetColumnSampleWidth(t.table, 3, "00000")
	SetColumnSampleWidth(t.table, pointsColumn, headers[pointsColumn])
	SetColumnSampleWidth(t.table, multiColumn, headers[multiColumn])
	for _, col := range narrowColumns {
		SetColumnSampleWidth(t.table, col, headers[col])
	}

	t.table.SelectionModel().OnSelectionChanged(func(selected *qtlib.QItemSelection, deselected *qtlib.QItemSelection) {
		indexes := t.table.SelectionModel().SelectedRows()
		if len(indexes) > 0 {
			row := indexes[0].Row()
			if t.controller != nil {
				t.controller.SelectRow(row)
			}
		}
	})
}

func (t *qsoTable) QSOAdded(qso core.QSO) {
	row := t.model.RowCount(qtlib.NewQModelIndex())

	items := []*qtlib.QStandardItem{
		qtlib.NewQStandardItem2(qso.Time.In(time.UTC).Format("15:04")),
		qtlib.NewQStandardItem2(qso.Callsign.String()),
		qtlib.NewQStandardItem2(qso.Band.String()),
		qtlib.NewQStandardItem2(qso.Mode.String()),
	}

	for _, value := range qso.MyExchange {
		items = append(items, qtlib.NewQStandardItem2(value))
	}

	for _, value := range qso.TheirExchange {
		items = append(items, qtlib.NewQStandardItem2(value))
	}

	items = append(items,
		qtlib.NewQStandardItem2(PointsToString(qso.Points, qso.Duplicate)),
		qtlib.NewQStandardItem2(PointsToString(qso.Multis, qso.Duplicate)),
		qtlib.NewQStandardItem2(DuplicateToCheckmark(qso.Duplicate)),
		qtlib.NewQStandardItem2(qso.Workmode.String()),
	)

	for col, item := range items {
		t.model.SetItem(row, col, item)
	}

	index := t.model.Index(row, 0, qtlib.NewQModelIndex())
	t.table.ScrollTo(index, qtlib.QAbstractItemView__EnsureVisible)
}

func (t *qsoTable) QSOsCleared() {
	ClearTableRows(t.model)
}

func (t *qsoTable) QSORowSelected(index int) {
	SelectTableRow(t.table, index, t.model)
}

func isNarrowExchangeField(f core.ExchangeField) bool {
	if len(f.Properties) != 1 {
		return false
	}
	return f.CanContainReport || f.CanContainSerial
}
