package ui

import (
	"fmt"
	"strings"
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/bandmap"
)

var _ bandmap.View = (*spotsView)(nil)

type SpotsController interface {
	SetVisibleBand(core.Band)
	SetActiveBand(core.Band)
	SelectEntry(core.BandmapEntryID)
}

const (
	spotColMark = iota
	spotColFrequency
	spotColCallsign
	spotColQualityTag
	spotColExchange
	spotColPoints
	spotColMultis
	spotColQTCs
	spotColSpotCount
	spotColAge
	spotColWeightedValue
	spotColDXCC

	spotColCount
)

type spotsView struct {
	*dockableView
	widget *qtlib.QWidget

	style *Style

	bandGrid    *qtlib.QWidget
	bandLayout  *qtlib.QHBoxLayout
	bandButtons map[core.Band]*qtlib.QPushButton
	bandsID     string

	table *qtlib.QTableView
	model *qtlib.QStandardItemModel
	bold  *qtlib.QFont

	controller        SpotsController
	currentFrame      core.BandmapFrame
	qtcsEnabled       bool
	qtcsEnabledKnown  bool
	suppressSelection bool
}

func newSpotsView(parent *qtlib.QWidget, controller SpotsController, style *Style) *spotsView {
	v := &spotsView{
		controller:  controller,
		style:       style,
		bandButtons: make(map[core.Band]*qtlib.QPushButton),
	}

	v.bold = qtlib.NewQFont()
	v.bold.SetBold(true)

	v.widget = qtlib.NewQWidget2()
	layout := qtlib.NewQVBoxLayout(v.widget)
	layout.SetContentsMargins(0, 0, 0, 0)

	v.bandGrid = qtlib.NewQWidget2()
	v.bandLayout = qtlib.NewQHBoxLayout(v.bandGrid)
	v.bandLayout.SetContentsMargins(0, 0, 0, 0)
	layout.AddWidget3(v.bandGrid, 0, 0)

	v.buildTable()
	layout.AddWidget3(v.table.QAbstractScrollArea.QFrame.QWidget, 1, 0)

	v.dockableView = newDockableView(parent, v.widget, "Spots", "spotsDock")

	return v
}

func (v *spotsView) RepaintForThemeChange() {
	v.dockableView.RepaintForThemeChange()
	v.widget.SetPalette(qtlib.QGuiApplication_Palette())
	v.widget.Update()
	repaintScrollBarsForThemeChange(v.table.QAbstractScrollArea)
}

func (v *spotsView) ShowFrame(frame core.BandmapFrame) {
	bandChanged := v.currentFrame.ActiveBand != frame.ActiveBand ||
		v.currentFrame.VisibleBand != frame.VisibleBand
	selectionChanged := v.currentFrame.SelectedEntry.ID != frame.SelectedEntry.ID

	oldFrame := v.currentFrame
	v.currentFrame = frame

	v.setupBands(frame.Bands)
	v.updateBands(frame.Bands)
	v.setQTCsEnabled(frame.QTCsEnabled)

	if bandChanged {
		v.reloadTable(frame)
	} else {
		v.applyIncrementalDiff(oldFrame, frame)
	}

	if selectionChanged || bandChanged {
		v.applySelection(frame)
	}
}

func (v *spotsView) buildTable() {
	v.model = qtlib.NewQStandardItemModel2(0, spotColCount)
	v.model.SetHorizontalHeaderLabels([]string{
		"", "Frequency", "Callsign", "T", "Exchange",
		"Pts", "Mult", "QTCs", "Spots", "Age", "Value", "DXCC",
	})

	v.table = qtlib.NewQTableView2()
	v.table.SetModel(v.model.QAbstractItemModel)
	ConfigureReadOnlyTable(v.table)
	v.table.SetSortingEnabled(false)
	v.table.HorizontalHeader().SetStretchLastSection(true)

	SetColumnSampleWidth(v.table, spotColMark, "W")
	SetColumnSampleWidth(v.table, spotColFrequency, "000000.00 kHz")
	SetColumnSampleWidth(v.table, spotColCallsign, "WW0WWW/p")
	SetColumnSampleWidth(v.table, spotColQualityTag, "W")
	SetColumnSampleWidth(v.table, spotColExchange, "Exchange")
	SetColumnSampleWidth(v.table, spotColPoints, "Pts")
	SetColumnSampleWidth(v.table, spotColMultis, "Mult")
	SetColumnSampleWidth(v.table, spotColQTCs, "QTCs")
	SetColumnSampleWidth(v.table, spotColSpotCount, "Spots")
	SetColumnSampleWidth(v.table, spotColAge, "< 00m")
	SetColumnSampleWidth(v.table, spotColWeightedValue, "0000.0")

	v.table.SetColumnHidden(spotColQTCs, true)

	v.table.SelectionModel().OnSelectionChanged(func(selected, deselected *qtlib.QItemSelection) {
		if v.suppressSelection {
			v.suppressSelection = false
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
		if row < 0 || row >= len(v.currentFrame.Entries) {
			return
		}
		v.controller.SelectEntry(v.currentFrame.Entries[row].ID)
	})
}

// --- Band grid -------------------------------------------------------------

func (v *spotsView) setupBands(bands []core.BandSummary) {
	id := toBandsID(bands)
	if id == v.bandsID {
		return
	}
	v.bandsID = id

	for band, btn := range v.bandButtons {
		v.bandLayout.RemoveWidget(btn.QWidget)
		btn.QWidget.DeleteLater()
		delete(v.bandButtons, band)
	}

	for _, bs := range bands {
		v.createBandButton(bs.Band)
	}
}

func toBandsID(bands []core.BandSummary) string {
	var b strings.Builder
	for _, bs := range bands {
		b.WriteString(string(bs.Band))
	}
	return b.String()
}

func (v *spotsView) createBandButton(band core.Band) {
	btn := qtlib.NewQPushButton2()
	v.bandLayout.AddWidget(btn.QWidget)
	v.bandButtons[band] = btn

	state := &bandClickState{
		timer: qtlib.NewQTimer2(btn.QWidget.QObject),
	}
	state.timer.SetSingleShot(true)
	state.timer.OnTimeout(func() {
		state.lastClickAt = time.Time{}
		if v.controller != nil {
			v.controller.SetVisibleBand(band)
		}
	})

	btn.OnClicked(func() {
		if v.controller == nil {
			return
		}
		interval := time.Duration(qtlib.QApplication_DoubleClickInterval()) * time.Millisecond
		if interval <= 0 {
			interval = 400 * time.Millisecond
		}
		now := time.Now()
		if !state.lastClickAt.IsZero() && now.Sub(state.lastClickAt) < interval {
			state.timer.Stop()
			state.lastClickAt = time.Time{}
			v.controller.SetActiveBand(band)
			return
		}
		state.lastClickAt = now
		state.timer.Stop()
		state.timer.Start(int(interval / time.Millisecond))
	})
}

type bandClickState struct {
	lastClickAt time.Time
	timer       *qtlib.QTimer
}

func (v *spotsView) updateBands(bands []core.BandSummary) {
	for _, bs := range bands {
		btn, ok := v.bandButtons[bs.Band]
		if !ok {
			continue
		}
		v.styleBandButton(btn, bs)
	}
}

func (v *spotsView) styleBandButton(btn *qtlib.QPushButton, bs core.BandSummary) {
	btn.SetText(fmt.Sprintf("%s\n%dP  %dM", bs.Band, bs.Points, bs.Multis()))
	btn.QWidget.SetStyleSheet(GetSpotsBandButtonStyle(bs.Active, bs.Visible, bs.MaxPoints || bs.MaxMultis))
}

// --- Table updates ---------------------------------------------------------

func (v *spotsView) applyIncrementalDiff(oldFrame, newFrame core.BandmapFrame) {
	visited := make(map[core.BandmapEntryID]bool, len(newFrame.Entries))
	var toInsert []core.BandmapEntryID

	for _, entry := range newFrame.Entries {
		visited[entry.ID] = true
		if oldIdx, existed := oldFrame.IndexOf(entry.ID); existed {
			v.updateRow(oldIdx, entry)
		} else {
			toInsert = append(toInsert, entry.ID)
		}
	}

	for i := len(oldFrame.Entries) - 1; i >= 0; i-- {
		if !visited[oldFrame.Entries[i].ID] {
			v.model.RemoveRows(i, 1, qtlib.NewQModelIndex())
		}
	}

	for _, id := range toInsert {
		idx, ok := newFrame.IndexOf(id)
		if !ok {
			continue
		}
		v.insertRow(idx, newFrame.Entries[idx])
	}
}

func (v *spotsView) reloadTable(frame core.BandmapFrame) {
	ClearTableRows(v.model)
	for i, entry := range frame.Entries {
		v.insertRow(i, entry)
	}
}

func (v *spotsView) insertRow(idx int, entry core.BandmapEntry) {
	v.model.InsertRow(idx, v.buildRow(entry))
}

func (v *spotsView) updateRow(idx int, entry core.BandmapEntry) {
	items := v.buildRow(entry)
	for col, item := range items {
		v.model.SetItem(idx, col, item)
	}
}

func (v *spotsView) buildRow(entry core.BandmapEntry) []*qtlib.QStandardItem {
	ageText, ageBold := FormatSpotAge(entry.LastHeard)
	bg := v.style.SpotBrush(entry.Source)
	fg := v.style.SpotForegroundBrush()

	type cell struct {
		text  string
		bold  bool
		align qtlib.AlignmentFlag
	}
	const alignRight = qtlib.AlignRight | qtlib.AlignVCenter
	cells := [spotColCount]cell{
		spotColMark:          {FormatSpotMark(entry, v.currentFrame), false, 0},
		spotColFrequency:     {FormatSpotFrequency(entry.Frequency), false, 0},
		spotColCallsign:      {entry.Call.String(), false, 0},
		spotColQualityTag:    {entry.Quality.Tag(), false, 0},
		spotColExchange:      {strings.Join(entry.Info.PredictedExchange, " "), false, 0},
		spotColPoints:        {PointsToString(entry.Info.Points, entry.Info.Duplicate), entry.Info.Points > 1 && !entry.Info.Duplicate, alignRight},
		spotColMultis:        {PointsToString(entry.Info.Multis, entry.Info.Duplicate), entry.Info.Multis > 0 && !entry.Info.Duplicate, alignRight},
		spotColQTCs:          {FormatQTCCount(entry.Info.SentQTCs, entry.Info.ReceivedQTCs), false, alignRight},
		spotColSpotCount:     {fmt.Sprintf("%d", entry.SpotCount), false, alignRight},
		spotColAge:           {ageText, ageBold, alignRight},
		spotColWeightedValue: {fmt.Sprintf("%.1f", entry.Info.WeightedValue), false, alignRight},
		spotColDXCC:          {getDXCCInformation(entry), false, 0},
	}

	items := make([]*qtlib.QStandardItem, spotColCount)
	for i, c := range cells {
		item := qtlib.NewQStandardItem2(c.text)
		item.SetForeground(fg)
		item.SetBackground(bg)
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
	v.suppressSelection = true
	defer func() { v.suppressSelection = false }()

	sel := v.table.SelectionModel()
	if frame.SelectedEntry.ID == core.NoEntryID {
		sel.ClearSelection()
		return
	}
	idx, ok := frame.IndexOf(frame.SelectedEntry.ID)
	if !ok {
		sel.ClearSelection()
		return
	}
	modelIdx := v.model.Index(idx, 0, qtlib.NewQModelIndex())
	sel.Select(modelIdx, qtlib.QItemSelectionModel__ClearAndSelect|qtlib.QItemSelectionModel__Rows)
	v.table.ScrollTo(modelIdx, qtlib.QAbstractItemView__EnsureVisible)
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
