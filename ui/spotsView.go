package ui

import (
	"fmt"
	"strings"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/bandmap"
)

var _ bandmap.View = (*spotsView)(nil)

type SpotsController interface {
	SpotFilterController
	SelectEntry(core.BandmapEntryID)
	SelectMarker(core.BandmapEntryID)
	RemoveMarker(core.BandmapEntryID)
}

const (
	spotColFrequency = iota
	spotColCallsign
	spotColQualityTag
	spotColExchange
	spotColPoints
	spotColMultis
	spotColQTCs
	spotColSpotCount
	spotColSNR
	spotColAge
	spotColWeightedValue
	spotColDXCC

	spotColCount
)

type spotsView struct {
	widget     *qtlib.QWidget
	window     *spotWindow
	filterArea *spotFilterArea

	style *Style

	table *qtlib.QTableView
	model *qtlib.QStandardItemModel
	bold  *qtlib.QFont

	controller        SpotsController
	currentFrame      core.BandmapFrame
	qtcsEnabled       bool
	qtcsEnabledKnown  bool
	suppressSelection bool
}

func newSpotsView(controller SpotsController, focuser EntryFocuser, style *Style) *spotsView {
	v := &spotsView{
		controller: controller,
		style:      style,
	}

	v.bold = qtlib.NewQFont()
	v.bold.SetBold(true)

	v.widget = qtlib.NewQWidget2()
	v.widget.SetObjectName(*qtlib.NewQAnyStringView3("spotsView"))
	layout := qtlib.NewQVBoxLayout(v.widget)
	layout.SetContentsMargins(0, 0, 0, 0)

	v.filterArea = newSpotFilterArea(controller, focuser)
	layout.AddWidget3(v.filterArea.Widget(), 0, 0)

	v.buildTable()
	layout.AddWidget3(v.table.QAbstractScrollArea.QFrame.QWidget, 1, 0)

	return v
}

func (v *spotsView) SetWindow(window *spotWindow) {
	v.window = window
}

func (v *spotsView) Show() {
	if v.window == nil {
		return
	}
	v.window.Show()
}

func (v *spotsView) Hide() {
	if v.window == nil {
		return
	}
	v.window.Hide()
}

func (v *spotsView) RepaintForThemeChange() {
	v.filterArea.RepaintForThemeChange()
	v.widget.SetPalette(qtlib.QGuiApplication_Palette())
	v.widget.Update()
	repaintScrollBarsForThemeChange(v.table.QAbstractScrollArea)
}

func (v *spotsView) ShowFrame(frame core.BandmapFrame) {
	// a new filter or a new order changes the position of every row, folding does not
	filterChanged := v.currentFrame.Filter.Band != frame.Filter.Band ||
		v.currentFrame.Filter.Mode != frame.Filter.Mode ||
		v.currentFrame.Filter.SortBy != frame.Filter.SortBy ||
		v.currentFrame.Filter.Descending != frame.Filter.Descending
	bandChanged := filterChanged ||
		v.currentFrame.ActiveBand != frame.ActiveBand ||
		v.currentFrame.VisibleBand != frame.VisibleBand
	selectionChanged := v.currentFrame.SelectedEntry.ID != frame.SelectedEntry.ID

	oldFrame := v.currentFrame
	v.currentFrame = frame

	v.filterArea.ShowFrame(frame.Filter)
	v.setQTCsEnabled(frame.QTCsEnabled)

	// the table moves the selection when rows come and go, therefore those changes of the
	// selection do not reach the controller
	rowsRemoved := false
	v.withSuppressedSelection(func() {
		if bandChanged {
			v.reloadTable(frame)
		} else {
			rowsRemoved = v.applyIncrementalDiff(oldFrame, frame)
		}
	})

	// a removed row makes the table select its neighbor, therefore the selection of the
	// controller is applied again
	if selectionChanged || bandChanged || rowsRemoved {
		v.applySelection(frame)
	}
}

func (v *spotsView) buildTable() {
	v.model = qtlib.NewQStandardItemModel2(0, spotColCount)
	v.model.SetHorizontalHeaderLabels([]string{
		"Frequency", "Callsign", "T", "Exchange",
		"Pts", "Mult", "QTCs", "Spots", "SNR", "Age", "Value", "DXCC",
	})

	v.table = qtlib.NewQTableView2()
	v.table.SetModel(v.model.QAbstractItemModel)
	ConfigureReadOnlyTable(v.table)
	v.table.SetSortingEnabled(false)
	v.table.HorizontalHeader().SetStretchLastSection(true)

	SetColumnSampleWidth(v.table, spotColFrequency, "000000.00 kHz")
	SetColumnSampleWidth(v.table, spotColCallsign, "WW0WWW/p")
	SetColumnSampleWidth(v.table, spotColQualityTag, "W")
	SetColumnSampleWidth(v.table, spotColExchange, "Exchange")
	SetColumnSampleWidth(v.table, spotColPoints, "Pts")
	SetColumnSampleWidth(v.table, spotColMultis, "Mult")
	SetColumnSampleWidth(v.table, spotColQTCs, "QTCs")
	SetColumnSampleWidth(v.table, spotColSpotCount, "Spots")
	SetColumnSampleWidth(v.table, spotColSNR, "-00")
	SetColumnSampleWidth(v.table, spotColAge, "< 00m")
	SetColumnSampleWidth(v.table, spotColWeightedValue, "0000.0")

	v.table.SetColumnHidden(spotColQTCs, true)

	// a right click opens the context menu, it selects no row: a selection tunes a VFO and
	// brings the main window to the front, which fights with the open menu
	v.table.OnMousePressEvent(func(super func(event *qtlib.QMouseEvent), event *qtlib.QMouseEvent) {
		if event.Button() == qtlib.RightButton {
			return
		}
		super(event)
	})

	v.table.SetContextMenuPolicy(qtlib.CustomContextMenu)
	v.table.OnCustomContextMenuRequested(func(pos *qtlib.QPoint) {
		v.showContextMenu(pos)
	})

	v.table.SelectionModel().OnSelectionChanged(func(selected, deselected *qtlib.QItemSelection) {
		if v.suppressSelection {
			return
		}
		if v.controller == nil {
			return
		}
		indexes := selected.Indexes()
		if len(indexes) == 0 {
			return
		}
		row := indexes[0].Row()
		if row < 0 || row >= len(v.currentFrame.Rows) {
			return
		}
		if v.currentFrame.Rows[row].Kind == core.MarkerRow {
			v.controller.SelectMarker(v.currentFrame.Rows[row].Marker.ID)
			return
		}
		v.controller.SelectEntry(v.currentFrame.Rows[row].Entry.ID)
	})
}

func (v *spotsView) showContextMenu(pos *qtlib.QPoint) {
	if v.controller == nil {
		return
	}
	// the scroll area forwards the position of the viewport, therefore the position needs
	// no translation
	row := v.table.IndexAt(pos).Row()
	if row < 0 || row >= len(v.currentFrame.Rows) {
		return
	}
	if v.currentFrame.Rows[row].Kind != core.MarkerRow {
		return
	}
	marker := v.currentFrame.Rows[row].Marker
	if marker.Kind == core.CQMarker {
		return
	}

	menu := qtlib.NewQMenu2()
	deleteAction := menu.AddActionWithText("Delete Marker")
	// the menu holds the mouse grab and runs its own event loop, therefore the deletion
	// happens after the menu closed
	chosen := menu.ExecWithPos(v.table.Viewport().MapToGlobalWithQPoint(pos))
	deleteChosen := chosen != nil && chosen.UnsafePointer() == deleteAction.UnsafePointer()
	menu.Delete()

	if deleteChosen {
		v.controller.RemoveMarker(marker.ID)
	}
}

// --- Table updates ---------------------------------------------------------

func (v *spotsView) applyIncrementalDiff(oldFrame, newFrame core.BandmapFrame) bool {
	visited := make(map[core.BandmapEntryID]bool, len(newFrame.Rows))
	var toInsert []core.BandmapEntryID

	for _, row := range newFrame.Rows {
		visited[row.ID()] = true
		if oldIdx, existed := oldFrame.IndexOf(row.ID()); existed {
			v.updateRow(oldIdx, row)
		} else {
			toInsert = append(toInsert, row.ID())
		}
	}

	rowsRemoved := false
	for i := len(oldFrame.Rows) - 1; i >= 0; i-- {
		if !visited[oldFrame.Rows[i].ID()] {
			v.model.RemoveRows(i, 1, qtlib.NewQModelIndex())
			rowsRemoved = true
		}
	}

	for _, id := range toInsert {
		idx, ok := newFrame.IndexOf(id)
		if !ok {
			continue
		}
		v.insertRow(idx, newFrame.Rows[idx])
	}

	return rowsRemoved
}

func (v *spotsView) reloadTable(frame core.BandmapFrame) {
	ClearTableRows(v.model)
	for i, row := range frame.Rows {
		v.insertRow(i, row)
	}
}

func (v *spotsView) insertRow(idx int, row core.BandmapRow) {
	v.model.InsertRow(idx, v.buildRow(row))
}

func (v *spotsView) updateRow(idx int, row core.BandmapRow) {
	items := v.buildRow(row)
	for col, item := range items {
		v.model.SetItem(idx, col, item)
	}
}

func (v *spotsView) buildRow(row core.BandmapRow) []*qtlib.QStandardItem {
	if row.Kind == core.MarkerRow {
		return v.buildMarkerRow(row.Marker)
	}
	return v.buildSpotRow(row.Entry)
}

func (v *spotsView) buildMarkerRow(marker core.BandmapMarker) []*qtlib.QStandardItem {
	ageText, _ := FormatSpotAge(marker.CreatedAt)
	cells := [spotColCount]string{
		spotColFrequency: FormatSpotFrequency(marker.Frequency),
		spotColCallsign:  marker.Text,
		spotColAge:       ageText,
	}
	background, foreground := v.style.MarkerBrushes(marker.Kind)

	items := make([]*qtlib.QStandardItem, spotColCount)
	for i, text := range cells {
		item := qtlib.NewQStandardItem2(text)
		item.SetBackground(background)
		item.SetForeground(foreground)
		items[i] = item
	}
	return items
}

func (v *spotsView) buildSpotRow(entry core.BandmapEntry) []*qtlib.QStandardItem {
	ageText, ageBold := FormatSpotAge(entry.LastHeard)
	worked := entry.Source == core.WorkedSpot

	type cell struct {
		text  string
		bold  bool
		align qtlib.AlignmentFlag
	}
	const alignRight = qtlib.AlignRight | qtlib.AlignVCenter
	cells := [spotColCount]cell{
		spotColFrequency:     {FormatSpotFrequency(entry.Frequency), false, 0},
		spotColCallsign:      {entry.Call.String(), false, 0},
		spotColQualityTag:    {entry.Quality.Tag(), false, 0},
		spotColExchange:      {strings.Join(entry.Info.PredictedExchange, " "), false, 0},
		spotColPoints:        {PointsToString(entry.Info.Points, entry.Info.Duplicate), entry.Info.Points > 1 && !entry.Info.Duplicate, alignRight},
		spotColMultis:        {PointsToString(entry.Info.Multis, entry.Info.Duplicate), entry.Info.Multis > 0 && !entry.Info.Duplicate, alignRight},
		spotColQTCs:          {FormatQTCCount(entry.Info.SentQTCs, entry.Info.ReceivedQTCs), false, alignRight},
		spotColSpotCount:     {fmt.Sprintf("%d", entry.SpotCount), false, alignRight},
		spotColSNR:           {FormatSpotSNR(entry.SNR), false, alignRight},
		spotColAge:           {ageText, ageBold, alignRight},
		spotColWeightedValue: {fmt.Sprintf("%.1f", entry.Info.WeightedValue), false, alignRight},
		spotColDXCC:          {getDXCCInformation(entry), false, 0},
	}

	items := make([]*qtlib.QStandardItem, spotColCount)
	for i, c := range cells {
		item := qtlib.NewQStandardItem2(c.text)
		if worked {
			item.SetForeground(v.style.WorkedSpotBrush())
		}
		if c.align != 0 {
			item.SetTextAlignment(c.align)
		}
		if c.bold {
			item.SetFont(v.bold)
		}
		items[i] = item
	}
	return items
}

func (v *spotsView) applySelection(frame core.BandmapFrame) {
	v.withSuppressedSelection(func() {
		sel := v.table.SelectionModel()
		idx, ok := frame.IndexOf(frame.SelectedEntry.ID)
		if frame.SelectedEntry.ID == core.NoEntryID || !ok {
			sel.ClearSelection()
			return
		}
		modelIdx := v.model.Index(idx, 0, qtlib.NewQModelIndex())
		sel.Select(modelIdx, qtlib.QItemSelectionModel__ClearAndSelect|qtlib.QItemSelectionModel__Rows)
		v.table.ScrollTo(modelIdx, qtlib.QAbstractItemView__EnsureVisible)
	})
}

func (v *spotsView) withSuppressedSelection(f func()) {
	previous := v.suppressSelection
	v.suppressSelection = true
	defer func() { v.suppressSelection = previous }()
	f()
}

func (v *spotsView) setQTCsEnabled(enabled bool) {
	if v.qtcsEnabledKnown && v.qtcsEnabled == enabled {
		return
	}
	v.qtcsEnabled = enabled
	v.qtcsEnabledKnown = true
	v.table.SetColumnHidden(spotColQTCs, !enabled)
}

// --- Formatters ------------------------------------------------------------

func getDXCCInformation(e core.BandmapEntry) string {
	if e.Info.DXCCEntity.PrimaryPrefix == "" {
		return ""
	}
	return fmt.Sprintf("%s (%s), %s, ITU %d, CQ %d",
		e.Info.DXCCEntity.Name, e.Info.DXCCEntity.PrimaryPrefix,
		e.Info.DXCCEntity.Continent,
		e.Info.DXCCEntity.ITUZone, e.Info.DXCCEntity.CQZone)
}
