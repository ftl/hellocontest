package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

// LogbookController represents the logbook controller.
type LogbookController interface {
	GetExchangeFields() ([]core.ExchangeField, []core.ExchangeField)
	SelectRow(int)
}

type logbookView struct {
	controller LogbookController

	view *gtk.TreeView
	list *gtk.ListStore

	ignoreSelection bool

	columnUTC                int
	columnCallsign           int
	columnBand               int
	columnMode               int
	columnFirstMyExchange    int
	columnLastMyExchange     int
	columnFirstTheirExchange int
	columnLastTheirExchange  int
	columnPoints             int
	columnMultis             int
	columnDuplicate          int
	columnWorkmode           int

	columnLast int
}

func setupLogbookView(builder *gtk.Builder) *logbookView {
	result := new(logbookView)

	result.view = getUI(builder, "logView").(*gtk.TreeView)

	result.columnUTC = 0
	result.columnCallsign = 1
	result.columnBand = 2
	result.columnMode = 3
	result.columnFirstMyExchange = 4
	result.columnLastMyExchange = result.columnFirstMyExchange
	result.columnFirstTheirExchange = result.columnLastMyExchange + 1
	result.columnLastTheirExchange = result.columnFirstTheirExchange
	result.columnPoints = result.columnLastTheirExchange + 1
	result.columnMultis = result.columnPoints + 1
	result.columnDuplicate = result.columnMultis + 1
	result.columnWorkmode = result.columnMultis + 2
	result.columnLast = result.columnWorkmode

	result.view.AppendColumn(createLogColumn("UTC", result.columnUTC))
	result.view.AppendColumn(createLogColumn("Callsign", result.columnCallsign))
	result.view.AppendColumn(createLogColumn("Band", result.columnBand))
	result.view.AppendColumn(createLogColumn("Mode", result.columnMode))
	result.view.AppendColumn(createLogColumn("My Exch", result.columnFirstMyExchange))
	result.view.AppendColumn(createLogColumn("Th Exch", result.columnFirstTheirExchange))
	result.view.AppendColumn(createLogColumn("Pts", result.columnPoints))
	result.view.AppendColumn(createLogColumn("Mult", result.columnMultis))
	result.view.AppendColumn(createLogColumn("D", result.columnDuplicate))
	result.view.AppendColumn(createLogColumn("WM", result.columnWorkmode))

	result.list = createLogListStore(int(result.view.GetNColumns()))
	result.view.SetModel(result.list)

	selection, err := result.view.GetSelection()
	if err != nil {
		log.Printf("no tree selection: %v", err)
		return result
	}
	selection.Connect("changed", result.onSelectionChanged)

	return result
}

func createLogColumn(title string, id int) *gtk.TreeViewColumn {
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

func createLogListStore(columnCount int) *gtk.ListStore {
	types := make([]glib.Type, columnCount)
	for i := range types {
		types[i] = glib.TYPE_STRING
	}
	result, err := gtk.ListStoreNew(types...)
	if err != nil {
		log.Fatalf("Cannot create list store: %v", err)
	}
	return result
}

func (v *logbookView) SetLogbookController(controller LogbookController) {
	v.controller = controller
	v.ExchangeFieldsChanged(v.controller.GetExchangeFields())
}

func (v *logbookView) ExchangeFieldsChanged(myExchangeFields []core.ExchangeField, theirExchangeFields []core.ExchangeField) {
	columnCount := int(v.view.GetNColumns())
	for i := v.columnFirstMyExchange; i < columnCount; i++ {
		column := v.view.GetColumn(v.columnFirstMyExchange)
		v.view.RemoveColumn(column)
	}

	v.columnLastMyExchange = v.columnFirstMyExchange + len(myExchangeFields) - 1
	v.columnFirstTheirExchange = v.columnLastMyExchange + 1
	v.columnLastTheirExchange = v.columnFirstTheirExchange + len(theirExchangeFields) - 1
	v.columnPoints = v.columnLastTheirExchange + 1
	v.columnMultis = v.columnPoints + 1
	v.columnDuplicate = v.columnMultis + 1
	v.columnWorkmode = v.columnMultis + 2
	v.columnLast = v.columnWorkmode

	for i := v.columnFirstMyExchange; i <= v.columnLastMyExchange; i++ {
		field := myExchangeFields[i-v.columnFirstMyExchange]
		var columnName string
		if len(field.Properties) == 1 {
			columnName = field.Short
		} else {
			columnName = "Exch"
		}
		v.view.AppendColumn(createLogColumn("My "+columnName, i))
	}

	for i := v.columnFirstTheirExchange; i <= v.columnLastTheirExchange; i++ {
		field := theirExchangeFields[i-v.columnFirstTheirExchange]
		var columnName string
		if len(field.Properties) == 1 {
			columnName = field.Short
		} else {
			columnName = "Exch"
		}
		v.view.AppendColumn(createLogColumn("Th "+columnName, i))
	}

	v.view.AppendColumn(createLogColumn("Pts", v.columnPoints))
	v.view.AppendColumn(createLogColumn("Mult", v.columnMultis))
	v.view.AppendColumn(createLogColumn("D", v.columnDuplicate))
	v.view.AppendColumn(createLogColumn("WM", v.columnWorkmode))

	v.list = createLogListStore(int(v.view.GetNColumns()))
	v.view.SetModel(v.list)
}

func (v *logbookView) QSOsCleared() {
	v.list.Clear()
}

func (v *logbookView) QSOAdded(qso core.QSO) {
	newRow := v.list.Append()
	err := v.fillQSOToRow(newRow, qso)
	if err != nil {
		log.Printf("Cannot fill new QSO data into row %s: %v", qso.String(), err)
	}
}
func (v *logbookView) fillQSOToRow(row *gtk.TreeIter, qso core.QSO) error {
	err := v.list.Set(row,
		[]int{
			v.columnUTC,
			v.columnCallsign,
			v.columnBand,
			v.columnMode,
			v.columnPoints,
			v.columnMultis,
			v.columnDuplicate,
			v.columnWorkmode,
		},
		[]any{
			qso.Time.In(time.UTC).Format("15:04"),
			qso.Callsign.String(),
			qso.Band.String(),
			qso.Mode.String(),
			pointsToString(qso.Points, qso.Duplicate),
			pointsToString(qso.Multis, qso.Duplicate),
			boolToCheckmark(qso.Duplicate),
			qso.Workmode.String(),
		})
	if err != nil {
		return err
	}

	for i, value := range qso.MyExchange {
		err := v.list.SetValue(row, i+v.columnFirstMyExchange, value)
		if err != nil {
			return err
		}
	}

	for i, value := range qso.TheirExchange {
		err := v.list.SetValue(row, i+v.columnFirstTheirExchange, value)
		if err != nil {
			return err
		}
	}

	return nil
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

func (v *logbookView) QSOInserted(index int, qso core.QSO) {
	// insertion is currently not supported as it does not happen in practice
	log.Printf("qso %d inserted at %d", qso.MyNumber, index)
}

func (v *logbookView) QSOUpdated(index int, _, qso core.QSO) {
	row, err := v.list.GetIterFromString(fmt.Sprintf("%d", index))
	if err != nil {
		log.Printf("cannot get iter: %v", err)
		return
	}

	err = v.fillQSOToRow(row, qso)
	if err != nil {
		log.Printf("Cannot fill changed QSO data into row %s: %v", qso.String(), err)
	}
}

func (v *logbookView) RowSelected(index int) {
	row, err := v.list.GetIterFromString(fmt.Sprintf("%d", index))
	if err != nil {
		log.Printf("cannot get iter: %v", err)
		return
	}
	path, err := v.list.GetPath(row)
	if err != nil {
		log.Printf("Cannot get path for list item: %v", err)
		return
	}
	v.view.SetCursorOnCell(path, v.view.GetColumn(1), nil, false)
	v.view.ScrollToCell(path, v.view.GetColumn(1), false, 0, 0)
}

func (v *logbookView) onSelectionChanged(selection *gtk.TreeSelection) bool {
	if v.ignoreSelection {
		return false
	}
	log.Print("selection changed")

	model, _ := v.view.GetModel()
	rows := selection.GetSelectedRows(model)
	if rows.Length() == 1 {
		row := rows.NthData(0).(*gtk.TreePath)
		index := row.GetIndices()[0]
		if v.controller != nil {
			v.controller.SelectRow(index)
		}
	}
	return true
}
