package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

const (
	columnUTC = iota
	columnCallsign
	columnBand
	columnMode
	columnKind
	columnHeader
	columnQTCTime
	columnQTCCallsign
	columnQTCExchange

	columnLast
)

type QTCListController interface {
	SelectRow(int)
}

type qtcListView struct {
	controller QTCListController

	view *gtk.TreeView
	list *gtk.ListStore
}

func setupQTCListView(builder *gtk.Builder) *qtcListView {
	result := new(qtcListView)

	result.view = getUI(builder, "qtcList").(*gtk.TreeView)

	result.view.AppendColumn(createLogColumn("UTC", columnUTC))
	result.view.AppendColumn(createLogColumn("Callsign", columnCallsign))
	result.view.AppendColumn(createLogColumn("Band", columnBand))
	result.view.AppendColumn(createLogColumn("Mode", columnMode))
	result.view.AppendColumn(createLogColumn("K", columnKind))
	result.view.AppendColumn(createLogColumn("Hdr", columnHeader))
	result.view.AppendColumn(createLogColumn("QTC Time", columnQTCTime))
	result.view.AppendColumn(createLogColumn("QTC Call", columnQTCCallsign))
	result.view.AppendColumn(createLogColumn("QTC Exch.", columnQTCExchange))

	result.list = createQTCListStore(int(result.view.GetNColumns()))
	result.view.SetModel(result.list)

	selection, err := result.view.GetSelection()
	if err != nil {
		log.Printf("no tree selection: %v", err)
		return result
	}
	selection.Connect("changed", result.onSelectionChanged)

	return result
}

func (v *qtcListView) SetQTCListController(controller QTCListController) {
	v.controller = controller
}

func (v *qtcListView) SetVisible(visible bool) {
	v.view.SetVisible(visible)
}

func (v *qtcListView) QTCsCleared() {
	v.list.Clear()
}

func (v *qtcListView) QTCAdded(qtc core.QTC) {
	newRow := v.list.Append()
	err := v.fillQTCToRow(newRow, qtc)
	if err != nil {
		log.Printf("Cannot fill new QTC data into row %s: %v", qtc.String(), err)
	}
}

func (v *qtcListView) fillQTCToRow(row *gtk.TreeIter, qtc core.QTC) error {
	return v.list.Set(row,
		[]int{
			columnUTC,
			columnCallsign,
			columnBand,
			columnMode,
			columnKind,
			columnHeader,
			columnQTCTime,
			columnQTCCallsign,
			columnQTCExchange,
		},
		[]any{
			qtc.Timestamp.In(time.UTC).Format("15:04"),
			qtc.TheirCallsign.String(),
			qtc.Band.String(),
			qtc.Mode.String(),
			kindToString(qtc.Kind),
			qtc.Header.String(),
			qtc.QTCTime.String(),
			qtc.QTCCallsign.String(),
			qtc.QTCNumber.String(),
		})
}

func kindToString(kind core.QTCKind) string {
	switch kind {
	case core.ReceivedQTC:
		return "R"
	case core.SentQTC:
		return "S"
	default:
		return "?"
	}
}

func (v *qtcListView) QTCUpdated(index int, _, qtc core.QTC) {
	row, err := v.list.GetIterFromString(fmt.Sprintf("%d", index))
	if err != nil {
		log.Printf("cannot get iter: %v", err)
		return
	}

	err = v.fillQTCToRow(row, qtc)
	if err != nil {
		log.Printf("Cannot fill changed QSO data into row %s: %v", qtc.String(), err)
	}
}

func (v *qtcListView) QTCRowSelected(index int) {
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

func (v *qtcListView) onSelectionChanged(selection *gtk.TreeSelection) bool {
	model, _ := v.view.GetModel()
	rows := selection.GetSelectedRows(model)
	if rows.Length() != 1 {
		return false
	}

	row := rows.NthData(0).(*gtk.TreePath)
	index := row.GetIndices()[0]
	if v.controller != nil {
		v.controller.SelectRow(index)
	}
	return true
}
