package ui

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	spotColumnFrequency = iota
	spotColumnCallsign
	spotColumnPredictedExchange
	spotColumnPoints
	spotColumnMultis
	spotColumnSpotCount
	spotColumnAge
	spotColumnWeightedValue
	spotColumnDXCC

	spotColumnForeground
	spotColumnBackground

	spotColumnCount
)

func setupSpotsTableView(v *spotsView, builder *gtk.Builder, controller SpotsController) {
	v.table = getUI(builder, "entryTable").(*gtk.TreeView)

	v.table.AppendColumn(createSpotMarkupColumn("Frequency", spotColumnFrequency))
	v.table.AppendColumn(createSpotTextColumn("Callsign", spotColumnCallsign))
	v.table.AppendColumn(createSpotTextColumn("Exchange", spotColumnPredictedExchange))
	v.table.AppendColumn(createSpotMarkupColumn("Pts", spotColumnPoints))
	v.table.AppendColumn(createSpotMarkupColumn("Mult", spotColumnMultis))
	v.table.AppendColumn(createSpotTextColumn("Spots", spotColumnSpotCount))
	v.table.AppendColumn(createSpotMarkupColumn("Age", spotColumnAge))
	v.table.AppendColumn(createSpotMarkupColumn("Value", spotColumnWeightedValue))
	v.table.AppendColumn(createSpotTextColumn("DXCC", spotColumnDXCC))

	v.tableContent = createSpotListStore(spotColumnCount)

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

func createSpotTextColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatalf("Cannot create text cell renderer for column %s: %v", title, err)
	}
	cellRenderer.SetProperty("foreground-set", true)
	cellRenderer.SetProperty("background-set", true)

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)
	if err != nil {
		log.Fatalf("Cannot create column %s: %v", title, err)
	}
	column.AddAttribute(cellRenderer, "foreground", spotColumnForeground)
	column.AddAttribute(cellRenderer, "background", spotColumnBackground)
	return column
}

func createSpotMarkupColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatalf("Cannot create text cell renderer for column %s: %v", title, err)
	}
	cellRenderer.SetProperty("foreground-set", true)
	cellRenderer.SetProperty("background-set", true)

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "markup", id)
	if err != nil {
		log.Fatalf("Cannot create column %s: %v", title, err)
	}
	column.AddAttribute(cellRenderer, "foreground", spotColumnForeground)
	column.AddAttribute(cellRenderer, "background", spotColumnBackground)
	return column
}

func createSpotProgressColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererProgressNew()
	if err != nil {
		log.Fatalf("Cannot create progress cell renderer for column %s: %v", title, err)
	}

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "value", id)
	if err != nil {
		log.Fatalf("Cannot create column %s: %v", title, err)
	}
	return column
}

func createSpotListStore(columnCount int) *gtk.ListStore {
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

func (v *spotsView) fillEntryToTableRow(row *gtk.TreeIter, entry core.BandmapEntry) error {
	foregroundColor, backgroundColor := v.getEntryColor(entry)

	return v.tableContent.Set(row,
		[]int{
			spotColumnFrequency,
			spotColumnCallsign,
			spotColumnPredictedExchange,
			spotColumnPoints,
			spotColumnMultis,
			spotColumnSpotCount,
			spotColumnAge,
			spotColumnWeightedValue,
			spotColumnDXCC,
			spotColumnForeground,
			spotColumnBackground,
		},
		[]any{
			formatSpotFrequency(entry.Frequency, entry.ProximityFactor(v.currentFrame.Frequency), entry.OnFrequency(v.currentFrame.Frequency)),
			entry.Call.String(),
			entry.Info.ExchangeText,
			formatPoints(entry.Info.Points, entry.Info.Duplicate, 1),
			formatPoints(entry.Info.Multis, entry.Info.Duplicate, 0),
			fmt.Sprintf("%d", entry.SpotCount),
			formatSpotAge(entry.LastHeard),
			fmt.Sprintf("%.1f", entry.Info.WeightedValue),
			v.getDXCCInformation(entry),
			foregroundColor.ToWeb(),
			backgroundColor.ToWeb(),
		},
	)
}

func formatSpotFrequency(frequency core.Frequency, proximity float64, onFrequency bool) string {
	size := 100 + math.Abs(proximity)*30
	result := fmt.Sprintf("<span size=\"%.0f%%\">%.2f kHz</span>", size, frequency/1000)
	if onFrequency {
		return fmt.Sprintf("<b>%s</b>", result)
	}

	return result
}

func formatPoints(value int, duplicate bool, threshold int) string {
	result := pointsToString(value, duplicate)
	if value > threshold && !duplicate {
		return fmt.Sprintf("<b>%s</b>", result)
	}
	return result
}

func formatSpotAge(lastHeard time.Time) string {
	result := time.Since(lastHeard).Truncate(time.Minute).String()
	if result == "0s" {
		return "<b>&lt; 1m</b>"
	}
	if strings.HasSuffix(result, "m0s") {
		result = result[:len(result)-2]
	}
	if strings.HasSuffix(result, "h0m") {
		result = result[:len(result)-2]
	}

	return result
}

func (v *spotsView) getEntryColor(entry core.BandmapEntry) (foreground, background style.Color) {
	foreground = v.colors.ColorByName("hellocontest-spot-fg")
	backgroundName := fmt.Sprintf("hellocontest-%s-bg", entrySourceStyles[entry.Source])
	background = v.colors.ColorByName(backgroundName)

	return foreground, background
}

func (v *spotsView) updateHighlightedColumns(entry core.BandmapEntry) error {
	row := v.tableRowByIndex(entry.Index)
	if row == nil {
		return fmt.Errorf("cannot reset frequency label for row with index %d", entry.Index)
	}

	return v.tableContent.Set(row,
		[]int{
			spotColumnFrequency,
			spotColumnAge,
			spotColumnWeightedValue,
		},
		[]any{
			formatSpotFrequency(entry.Frequency, entry.ProximityFactor(v.currentFrame.Frequency), entry.OnFrequency(v.currentFrame.Frequency)),
			formatSpotAge(entry.LastHeard),
			fmt.Sprintf("%.1f", entry.Info.WeightedValue),
		},
	)
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

func (v *spotsView) showInitialFrameInTable(frame core.BandmapFrame) {
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

func (v *spotsView) revealTableEntry(entry core.BandmapEntry) {
	if !v.controller.EntryVisible(entry.Index) {
		log.Printf("invisible entry #%d %s on %s not selected", entry.Index, entry.Call, entry.Band)
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

	column := v.table.GetColumn(1)
	v.table.ScrollToCell(filteredPath, column, false, 0, 0)
}

func (v *spotsView) refreshTable() {
	if !v.ignoreSelection {
		v.ignoreSelection = true
		defer func() {
			v.ignoreSelection = false
		}()
	}

	v.tableFilter.Refilter()
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
	index, selected := v.getSelectedIndex(selection)
	if !selected {
		return true
	}
	selection.UnselectAll()

	if v.ignoreSelection {
		log.Printf("table selection change ignored")
		return true
	}

	if v.controller == nil {
		return true
	}

	v.controller.SelectEntry(index)

	return true
}

func (v *spotsView) getSelectedIndex(selection *gtk.TreeSelection) (int, bool) {
	rows := selection.GetSelectedRows(v.tableFilter)
	if rows.Length() != 1 {
		return 0, false
	}

	filteredPath := rows.NthData(0).(*gtk.TreePath)
	path := v.tableFilter.ConvertPathToChildPath(filteredPath)
	return path.GetIndices()[0], true
}
