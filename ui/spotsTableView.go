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
	spotColumnMark = iota
	spotColumnFrequency
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

	v.table.AppendColumn(createSpotMarkupColumn("", spotColumnMark))
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
	v.table.SetModel(v.tableContent)

	v.tableSelectionActive = false
	selection, err := v.table.GetSelection()
	if err != nil {
		log.Printf("no tree selection: %v", err)
		return
	}
	selection.SetMode(gtk.SELECTION_SINGLE)
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
			spotColumnMark,
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
			formatSpotMark(entry, v.currentFrame),
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

func formatSpotMark(entry core.BandmapEntry, frame core.BandmapFrame) string {
	if entry.ID == frame.HighestValueEntry.ID {
		return "H"
	}
	if entry.ID == frame.SelectedEntry.ID {
		return ">"
	}
	if entry.OnFrequency(frame.Frequency) {
		return "|"
	}
	if entry.ID == frame.NearestEntry.ID {
		return "N"
	}
	return ""
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

func (v *spotsView) getTablePathForEntry(entry core.BandmapEntry) (*gtk.TreePath, bool) {
	index, found := v.currentFrame.IndexOf(entry.ID)
	if !found {
		log.Printf("cannot find index for entry with ID %d", entry.ID)
		return nil, false
	}

	row, err := v.tableContent.GetIterFromString(fmt.Sprintf("%d", index))
	if err != nil {
		log.Printf("cannot find table row with ID %d", entry.ID)
		return nil, false
	}

	path, err := v.tableContent.GetPath(row)
	if err != nil {
		log.Printf("no table path found for index with ID %d: %v", entry.ID, err)
		return nil, false
	}

	return path, true
}

func (v *spotsView) revealTableEntry(entry core.BandmapEntry) {
	path, found := v.getTablePathForEntry(entry)
	if !found {
		return
	}

	column := v.table.GetColumn(1)
	v.table.ScrollToCell(path, column, false, 0, 0)
}

func (v *spotsView) setSelectedTableEntry(entry core.BandmapEntry) {
	path, found := v.getTablePathForEntry(entry)
	if !found {
		return
	}

	selection, err := v.table.GetSelection()
	if err != nil {
		log.Printf("no table selection available: %v", err)
	}
	selection.SelectPath(path)
	column := v.table.GetColumn(1)
	v.table.ScrollToCell(path, column, false, 0, 0)
}

func (v *spotsView) clearSelection() {
	selection, err := v.table.GetSelection()
	if err != nil {
		log.Printf("no table selection available: %v", err)
	}
	selection.UnselectAll()
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
	rows := selection.GetSelectedRows(v.tableContent)
	if rows.Length() != 1 {
		return core.BandmapEntry{}, false
	}

	path := rows.NthData(0).(*gtk.TreePath)
	index := path.GetIndices()[0]
	if index < 0 || index >= len(v.currentFrame.Entries) {
		return core.BandmapEntry{}, false
	}
	entry := v.currentFrame.Entries[index]
	return entry, true
}
