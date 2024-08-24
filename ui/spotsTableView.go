package ui

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	spotColumnFrequency = iota
	spotColumnCallsign
	spotColumnQualityTag
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
	v.table.Connect("button-press-event", v.activateTableSelection)

	v.table.AppendColumn(createSpotMarkupColumn("Frequency", spotColumnFrequency))
	v.table.AppendColumn(createSpotMarkupColumn("Callsign", spotColumnCallsign))
	v.table.AppendColumn(createSpotTextColumn("T", spotColumnQualityTag))
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

	v.tableSelectionActive = false
	selection, err := v.table.GetSelection()
	if err != nil {
		log.Printf("no tree selection: %v", err)
		return
	}
	selection.SetMode(gtk.SELECTION_NONE)
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
			spotColumnQualityTag,
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
			formatSpotFrequency(entry.Frequency),
			formatSpotCall(entry.Call),
			entry.Quality.Tag(),
			entry.Info.ExchangeText,
			formatPoints(entry.Info.Points, entry.Info.Duplicate, 1),
			formatPoints(entry.Info.Multis, entry.Info.Duplicate, 0),
			fmt.Sprintf("%d", entry.SpotCount),
			formatSpotAge(entry.LastHeard),
			fmt.Sprintf("%.1f", entry.Info.WeightedValue),
			getDXCCInformation(entry),
			foregroundColor.ToWeb(),
			backgroundColor.ToWeb(),
		},
	)
}

func formatSpotFrequency(frequency core.Frequency) string {
	return fmt.Sprintf("%.2f kHz", frequency/1000)
}

func formatSpotCall(call callsign.Callsign) string {
	return call.String()
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

func getDXCCInformation(entry core.BandmapEntry) string {
	if entry.Info.PrimaryPrefix == "" {
		return ""
	}
	return fmt.Sprintf("%s (%s), %s, ITU %d, CQ %d", entry.Info.DXCCName, entry.Info.PrimaryPrefix, entry.Info.Continent, entry.Info.ITUZone, entry.Info.CQZone)
}

func (v *spotsView) getEntryColor(entry core.BandmapEntry) (foreground, background style.Color) {
	foreground = v.colors.ColorByName("hellocontest-spot-fg")
	backgroundName := fmt.Sprintf("hellocontest-%s-bg", entrySourceStyles[entry.Source])
	background = v.colors.ColorByName(backgroundName)

	return foreground, background
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
	if index < 0 || index >= len(v.currentFrame.Entries) {
		log.Printf("filterTableRow: row index out of bounds: %d", index)
		return false

	}

	entry := v.currentFrame.Entries[index]
	return v.controller.EntryVisible(entry.ID)
}

func (v *spotsView) showFrameInTable(frame core.BandmapFrame) {
	v.tableContent.Clear()

	for _, entry := range frame.Entries {
		newRow := v.tableContent.Append()
		err := v.fillEntryToTableRow(newRow, entry)
		if err != nil {
			log.Printf("Cannot add entry to spots table row %v: %v", entry, err)
		}
	}
}

func (v *spotsView) revealTableEntry(entry core.BandmapEntry) {
	if !v.controller.EntryVisible(entry.ID) {
		log.Printf("invisible entry #%d %s on %s not selected", entry.ID, entry.Call, entry.Band)
		return
	}

	index := -1
	for i, e := range v.currentFrame.Entries {
		if e.ID == entry.ID {
			index = i
			break
		}
	}
	if index == -1 {
		log.Printf("cannot find index for entry with ID %d", entry.ID)
		return
	}

	row, err := v.tableContent.GetIterFromString(fmt.Sprintf("%d", index))
	if err != nil {
		log.Printf("cannot find table row with ID %d", entry.ID)
		return
	}

	path, err := v.tableContent.GetPath(row)
	if err != nil {
		log.Printf("no table path found for index with ID %d: %v", entry.ID, err)
		return
	}
	filteredPath := v.tableFilter.ConvertChildPathToPath(path)

	column := v.table.GetColumn(1)
	v.table.ScrollToCell(filteredPath, column, false, 0, 0)
}

func (v *spotsView) activateTableSelection(_ *gtk.TreeView, event *gdk.Event) {
	buttonEvent := gdk.EventButtonNewFromEvent(event)
	if buttonEvent.Button() != gdk.BUTTON_PRIMARY {
		return
	}

	if buttonEvent.Type() == gdk.EVENT_BUTTON_PRESS {
		v.tableSelectionActive = true
		selection, _ := v.table.GetSelection()
		selection.SetMode(gtk.SELECTION_SINGLE)
	}
}

func (v *spotsView) onTableSelectionChanged(selection *gtk.TreeSelection) bool {
	entry, selected := v.getSelectedEntry(selection)

	if !v.tableSelectionActive {
		log.Printf("table selection change ignored")
		return true
	}
	v.tableSelectionActive = false
	selection.UnselectAll()
	selection.SetMode(gtk.SELECTION_NONE)

	if !selected {
		return true
	}

	if v.controller == nil {
		return true
	}

	v.controller.SelectEntry(entry.ID)

	return true
}

func (v *spotsView) getSelectedEntry(selection *gtk.TreeSelection) (core.BandmapEntry, bool) {
	rows := selection.GetSelectedRows(v.tableFilter)
	if rows.Length() != 1 {
		return core.BandmapEntry{}, false
	}

	filteredPath := rows.NthData(0).(*gtk.TreePath)
	path := v.tableFilter.ConvertPathToChildPath(filteredPath)
	index := path.GetIndices()[0]
	if index < 0 || index >= len(v.currentFrame.Entries) {
		return core.BandmapEntry{}, false
	}
	entry := v.currentFrame.Entries[index]
	return entry, true
}
