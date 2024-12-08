package fyneui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/ftl/hellocontest/core"
)

const (
	exchangeAt            = 4
	myExchangeTemplate    = "My %s"
	theirExchangeTemplate = "Th %s"
	exchangeColumnWidth   = 100
)

var columnHeaders = []string{"UTC", "Callsign", "Band", "Mode", "Pts", "Mult", "D"}
var columnWidth = []float32{55, 100, 55, 50, 35, 40, 10}

type LogbookController interface {
	GetExchangeFields() ([]core.ExchangeField, []core.ExchangeField)
	SelectRow(int)
}

type qsoList struct {
	container  *fyne.Container
	table      *widget.Table
	controller LogbookController

	myExchangeFields    []core.ExchangeField
	theirExchangeFields []core.ExchangeField

	headerRow []string
	entryRows [][]string
}

func setupQSOList() *qsoList {
	result := &qsoList{
		headerRow: columnHeaders,
		entryRows: [][]string{},
	}

	result.table = widget.NewTable(result.tableSize, result.createTableCell, result.updateValueCell)
	result.table.ShowHeaderRow = true
	result.table.CreateHeader = result.createTableCell
	result.table.UpdateHeader = result.updateHeaderCell

	result.container = container.New(layout.NewStackLayout(), result.table)

	return result
}

func (l *qsoList) SetLogbookController(controller LogbookController) {
	l.controller = controller
	l.ExchangeFieldsChanged(l.controller.GetExchangeFields())
}

func (l *qsoList) updateHeaderRow() {
	exchangeLength := len(l.myExchangeFields) + len(l.theirExchangeFields)
	firstTheirExchange := exchangeAt + len(l.myExchangeFields)
	length := len(columnHeaders) + exchangeLength

	headerRow := make([]string, length)
	for i := range headerRow {
		if i < exchangeAt {
			headerRow[i] = columnHeaders[i]
		} else if i >= exchangeAt+exchangeLength {
			headerRow[i] = columnHeaders[i-exchangeLength]
		} else if i < firstTheirExchange {
			headerRow[i] = fmt.Sprintf(myExchangeTemplate, exchangeColumnName(l.myExchangeFields[i-exchangeAt]))
		} else {
			headerRow[i] = fmt.Sprintf(theirExchangeTemplate, exchangeColumnName(l.theirExchangeFields[i-firstTheirExchange]))
		}
	}

	l.headerRow = headerRow
}

func exchangeColumnName(field core.ExchangeField) string {
	if len(field.Properties) == 1 {
		return field.Short
	}
	return "Exch"
}

func (l *qsoList) tableSize() (int, int) {
	return len(l.entryRows), len(l.headerRow)
}

func (l *qsoList) createTableCell() fyne.CanvasObject {
	return widget.NewLabel("")
}

func (l *qsoList) updateHeaderCell(id widget.TableCellID, cell fyne.CanvasObject) {
	label := cell.(*widget.Label)
	label.TextStyle.Bold = true
	label.SetText(l.columnHeaderText(id.Col))

	l.table.SetColumnWidth(id.Col, l.columnWidth(id.Col))
}

func (l *qsoList) columnHeaderText(column int) string {
	if column < 0 || column >= len(l.headerRow) {
		return ""
	}

	return l.headerRow[column]
}

func (l *qsoList) columnWidth(column int) float32 {
	if column < exchangeAt {
		return columnWidth[column]
	} else if column >= len(l.headerRow)-exchangeAt {
		return columnWidth[column-exchangeAt]
	} else {
		return exchangeColumnWidth
	}
}

func (l *qsoList) updateValueCell(id widget.TableCellID, cell fyne.CanvasObject) {
	label := cell.(*widget.Label)
	label.TextStyle.Bold = false
	label.SetText(l.valueCellText(id.Row, id.Col))
}

func (l *qsoList) valueCellText(row int, column int) string {
	if row < 0 || row >= len(l.entryRows) {
		return ""
	}
	entryRow := l.entryRows[row]
	if column < 0 || column >= len(entryRow) {
		return ""
	}

	return entryRow[column]
}

func (l *qsoList) QSOsCleared() {
	l.entryRows = [][]string{}
}

func (l *qsoList) QSOAdded(qso core.QSO) {
	entryRow := l.qsoToRow(qso)
	l.entryRows = append(l.entryRows, entryRow)
}

func (l *qsoList) qsoToRow(qso core.QSO) []string {
	length := len(l.headerRow)
	firstTheirExchange := exchangeAt + len(l.myExchangeFields)
	result := make([]string, length)
	result[0] = qso.Time.In(time.UTC).Format("15:04")
	result[1] = qso.Callsign.String()
	result[2] = qso.Band.String()
	result[3] = qso.Mode.String()
	result[length-3] = pointsToString(qso.Points, qso.Duplicate)
	result[length-2] = pointsToString(qso.Multis, qso.Duplicate)
	result[length-1] = boolToCheckmark(qso.Duplicate)

	for i, value := range qso.MyExchange {
		result[i+exchangeAt] = value
	}

	for i, value := range qso.TheirExchange {
		result[i+firstTheirExchange] = value
	}

	return result
}

func pointsToString(points int, duplicate bool) string {
	if duplicate {
		return fmt.Sprintf("(%d)", points)
	}
	return fmt.Sprintf("%d", points)
}

func boolToCheckmark(value bool) string {
	if value {
		return "✓"
	}
	return ""
}

func (l *qsoList) forRow(row int, f func(widget.TableCellID)) {
	for col := range l.headerRow {
		id := widget.TableCellID{Row: row, Col: col}
		f(id)
	}
}

func (l *qsoList) RowSelected(row int) {
	l.table.Select(widget.TableCellID{Row: row, Col: 0})
}

func (l *qsoList) ExchangeFieldsChanged(myExchangeFields []core.ExchangeField, theirExchangeFields []core.ExchangeField) {
	l.myExchangeFields = myExchangeFields
	l.theirExchangeFields = theirExchangeFields
	l.updateHeaderRow()
}
