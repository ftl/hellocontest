package ui

import (
	"fmt"
	"log"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

func setupSpotsTableView(v *spotsView, builder *gtk.Builder, controller SpotsController) {
	v.table = getUI(builder, "entryTable").(*gtk.TreeView)

	v.columnFrequency = 0
	v.columnCallsign = 1

	v.table.AppendColumn(createColumn("Frequency", v.columnFrequency))
	v.table.AppendColumn(createColumn("Callsign", v.columnCallsign))

	v.tableContent = createListStore(int(v.table.GetNColumns()))

	filter, err := v.tableContent.FilterNew(nil)
	if err != nil {
		log.Printf("No table filter: %v", err)
		v.table.SetModel(v.tableContent)
		return
	}
	filter.SetVisibleFunc(v.filterTableRow)

	v.tableFilter = filter
	v.table.SetModel(v.tableFilter)

	selection, err := v.table.GetSelection()
	if err != nil {
		log.Printf("no tree selection: %v", err)
		return
	}
	selection.Connect("changed", v.onTableSelectionChanged)
}

func (v *spotsView) filterTableRow(model *gtk.TreeModel, iter *gtk.TreeIter) bool {
	if v.controller == nil {
		log.Print("filterTableRow: no controller")
		return false
	}
	if model == &v.tableContent.TreeModel {
		log.Printf("filtering using the content model")
	}
	if model == &v.tableFilter.TreeModel {
		log.Printf("filtering using the filter model")
	}

	path, err := model.GetPath(iter)
	if err != nil {
		log.Printf("filterTableRow: unable to get path for iter %v: %v", iter, err)
		return false
	}

	index := path.GetIndices()[0]
	return v.controller.EntryVisible(index)
}

func (v *spotsView) fillEntryToTableRow(row *gtk.TreeIter, entry core.BandmapEntry) error {
	err := v.tableContent.Set(row,
		[]int{
			v.columnFrequency,
			v.columnCallsign,
		},
		[]any{
			entry.Frequency.String(),
			entry.Call.String(),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (v *spotsView) showInitialFrameInTable(frame core.BandmapFrame) {
	log.Printf("show inital frame in table")
	v.tableContent.Clear()

	for _, entry := range frame.Entries {
		newRow := v.tableContent.Append()
		err := v.fillEntryToTableRow(newRow, entry)
		if err != nil {
			log.Printf("Cannot add entry to spots table row %v: %v", entry, err)
		}
	}
}

func (v *spotsView) addTableEntry(entry core.BandmapEntry) {
	newRow := v.tableContent.Insert(entry.Index)
	err := v.fillEntryToTableRow(newRow, entry)
	if err != nil {
		log.Printf("Cannot insert entry into spots table row %v: %v", entry, err)
	}
}

func (v *spotsView) updateTableEntry(entry core.BandmapEntry) {
	row := v.tableRowByIndex(entry.Index)
	if row == nil {
		return
	}

	err := v.fillEntryToTableRow(row, entry)
	if err != nil {
		log.Printf("Cannot update entry in spots table row %v: %v", entry, err)
	}
}

func (v *spotsView) removeTableEntry(entry core.BandmapEntry) {
	row := v.tableRowByIndex(entry.Index)
	if row == nil {
		return
	}

	v.tableContent.Remove(row)
}

func (v *spotsView) selectTableEntry(entry core.BandmapEntry) {
	log.Printf("select entry in table: %s %d", entry.Call, entry.Index)
	if !v.controller.EntryVisible(entry.Index) {
		log.Printf("invisible entry not selected")
		return
	}

	row, err := v.tableContent.GetIterFromString(fmt.Sprintf("%d", entry.Index))
	if err != nil {
		log.Printf("cannot find table row with index %d", entry.Index)
		return
	}

	path, err := v.tableContent.GetPath(row)
	if err != nil {
		log.Printf("no table path found for index %d: %v", entry.Index, err)
		return
	}

	filteredPath := v.tableFilter.ConvertChildPathToPath(path)
	log.Printf("filtered path: %v", filteredPath)

	column := v.table.GetColumn(1)
	v.table.SetCursorOnCell(filteredPath, column, nil, false)
	v.table.ScrollToCell(filteredPath, column, false, 0, 0)

	selection, _ := v.table.GetSelection()
	selection.UnselectAll()
}

func (v *spotsView) refreshTable() {
	runAsync(func() {
		v.ignoreSelection = true
		defer func() {
			v.ignoreSelection = false
		}()
		log.Printf("refresh table")
		selection, _ := v.table.GetSelection()
		selection.UnselectAll()

		v.tableFilter.Refilter()
	})
}

func (v *spotsView) tableRowByIndex(index int) *gtk.TreeIter {
	result, err := v.tableContent.GetIterFromString(fmt.Sprintf("%d", index))
	if err != nil {
		log.Printf("Cannot find table row with index %d", index)
		return nil
	}
	return result
}

func (v *spotsView) onTableSelectionChanged(selection *gtk.TreeSelection) bool {
	if v.ignoreSelection {
		log.Printf("table selection change ignored")
		return true
	}

	log.Printf("table selection changed entered")
	defer func() {
		log.Printf("table selection changed left")
	}()

	if v.controller == nil {
		return true
	}

	log.Printf("getting selected rows")
	rows := selection.GetSelectedRows(v.tableFilter)
	if rows.Length() != 1 {
		return true
	}

	log.Printf("extract the filtered row")
	filteredPath := rows.NthData(0).(*gtk.TreePath)
	log.Printf("convert path to child path %v", filteredPath)
	path := v.tableFilter.ConvertPathToChildPath(filteredPath)
	log.Printf("getting the index %v", path)
	index := path.GetIndices()[0]
	log.Printf("selecting the entry %d", index)
	v.controller.SelectEntry(index)

	return true
}
