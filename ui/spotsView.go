package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

var entrySourceStyles = map[core.SpotType]style.Class{
	core.WorkedSpot:  "worked-spot",
	core.ManualSpot:  "manual-spot",
	core.SkimmerSpot: "skimmer-spot",
	core.RBNSpot:     "rbn-spot",
	core.ClusterSpot: "cluster-spot",
}

const (
	bandClass        style.Class = "band"
	bandLabelClass   style.Class = "band-label"
	bandPointsClass  style.Class = "band-points"
	bandMultisClass  style.Class = "band-multis"
	bandActiveClass  style.Class = "band-active"
	bandVisibleClass style.Class = "band-visible"
	maxValueClass    style.Class = "max-value"

	spotRowClass               style.Class = "spot-row"
	spotFrequencyClass         style.Class = "spot-frequency"
	spotCallClass              style.Class = "spot-call"
	spotGeoInfoClass           style.Class = "spot-geo-info"
	spotPredictedExchangeClass style.Class = "spot-predicted-exchange"
	spotScoreClass             style.Class = "spot-score"
)

type SpotsController interface {
	SetVisibleBand(core.Band)
	SetActiveBand(core.Band)

	RemainingLifetime(int) float64
	EntryVisible(int) bool
	SelectEntry(int)
}

type spotsView struct {
	controller SpotsController

	bandGrid  *gtk.Grid
	entryList *gtk.ListBox
	// style     *style

	bands             []core.BandSummary
	bandsID           string
	currentFrame      core.BandmapFrame
	initialFrameShown bool
	ignoreSelection   bool
}

func setupSpotsView(builder *gtk.Builder, controller SpotsController) *spotsView {
	result := &spotsView{
		controller:        controller,
		initialFrameShown: false,
	}

	result.bandGrid = getUI(builder, "bandGrid").(*gtk.Grid)
	result.entryList = getUI(builder, "entryList").(*gtk.ListBox)
	result.entryList.SetFilterFunc(result.filterRow)
	result.entryList.Connect("row-selected", result.onRowSelected)

	return result
}

func (v *spotsView) getRemainingLifetime(index int) float64 {
	return v.controller.RemainingLifetime(index)
}

func (v *spotsView) getProximity(index int) float64 {
	if index < 0 || index >= len(v.currentFrame.Entries) {
		return 0
	}
	return v.currentFrame.Entries[index].ProximityFactor(v.currentFrame.Frequency)
}

func (v *spotsView) filterRow(row *gtk.ListBoxRow) bool {
	return v.controller.EntryVisible(row.GetIndex())
}

func (v *spotsView) ShowFrame(frame core.BandmapFrame) {
	if v == nil {
		return
	}

	runAsync(func() {
		v.currentFrame = frame
		v.setupBands(frame.Bands)
		v.updateBands(frame.Bands)
		v.entryList.SetFilterFunc(v.filterRow)

		if !v.initialFrameShown {
			v.initialFrameShown = true
			v.showEntries(frame.Entries)
		}

		v.entryList.ShowAll()
	})
}

func (v *spotsView) setupBands(bands []core.BandSummary) {
	if v == nil {
		return
	}
	bandsID := toBandsID(bands)
	if bandsID == v.bandsID {
		return
	}
	v.bands = bands
	v.bandsID = bandsID

	v.bandGrid.GetChildren().Foreach(func(c any) {
		child, ok := c.(*gtk.Widget)
		if ok {
			child.Destroy()
		}
	})
	v.bandGrid.RemoveRow(0)

	for i, band := range bands {
		label := v.newBand(band)
		label.SetHAlign(gtk.ALIGN_FILL)
		label.SetHExpand(true)
		v.bandGrid.Attach(label, i, 0, 1, 1)
	}

	v.bandGrid.ShowAll()
}

func toBandsID(bands []core.BandSummary) string {
	result := make([]byte, 0, len(bands)*4)
	for _, band := range bands {
		result = append(result, []byte(band.Band)...)
	}
	return string(result)
}

func (v *spotsView) newBand(band core.BandSummary) *gtk.Widget {
	button, _ := gtk.ButtonNew()
	button.SetHAlign(gtk.ALIGN_END)
	button.SetVAlign(gtk.ALIGN_FILL)
	button.SetHExpand(true)
	button.Connect("button-press-event", v.selectBand(band.Band))
	style.AddClass(&button.Widget, bandClass)

	grid, _ := gtk.GridNew()
	grid.SetColumnSpacing(3)
	button.Add(grid)

	label, _ := gtk.LabelNew(string(band.Band))
	label.SetHAlign(gtk.ALIGN_END)
	label.SetVAlign(gtk.ALIGN_FILL)
	label.SetHExpand(true)
	grid.Attach(label, 0, 0, 1, 2)
	style.AddClass(&label.Widget, bandLabelClass)

	points, _ := gtk.LabelNew("")
	points.SetHAlign(gtk.ALIGN_FILL)
	points.SetVAlign(gtk.ALIGN_FILL)
	points.SetHExpand(true)
	grid.Attach(points, 1, 0, 1, 1)
	style.AddClass(&points.Widget, bandPointsClass)

	multis, _ := gtk.LabelNew("")
	multis.SetHAlign(gtk.ALIGN_FILL)
	multis.SetVAlign(gtk.ALIGN_FILL)
	points.SetHExpand(true)
	grid.Attach(multis, 1, 1, 1, 1)
	style.AddClass(&multis.Widget, bandMultisClass)

	v.updateBand(button, band)

	return button.ToWidget()
}

func (v *spotsView) updateBand(button *gtk.Button, band core.BandSummary) {
	child, _ := button.GetChild()
	grid := child.(*gtk.Grid)

	child, _ = grid.GetChildAt(1, 0)
	points := child.(*gtk.Label)
	points.SetText(fmt.Sprintf("%dP", band.Points))

	child, _ = grid.GetChildAt(1, 1)
	multis := child.(*gtk.Label)
	multis.SetText(fmt.Sprintf("%dM", band.Multis))

	if band.MaxPoints {
		style.AddClass(&points.Widget, maxValueClass)
	} else {
		style.RemoveClass(&points.Widget, maxValueClass)
	}
	if band.MaxMultis {
		style.AddClass(&multis.Widget, maxValueClass)
	} else {
		style.RemoveClass(&multis.Widget, maxValueClass)
	}
	if band.Active {
		style.AddClass(&button.Widget, bandActiveClass)
	} else {
		style.RemoveClass(&button.Widget, bandActiveClass)
	}
	if band.Visible {
		style.AddClass(&button.Widget, bandVisibleClass)
	} else {
		style.RemoveClass(&button.Widget, bandVisibleClass)
	}
}

func (v *spotsView) updateBands(bands []core.BandSummary) {
	for i, band := range bands {
		child, _ := v.bandGrid.GetChildAt(i, 0)
		button, ok := child.(*gtk.Button)
		if ok {
			v.updateBand(button, band)
		}
	}
}

func (v *spotsView) showEntries(entries []core.BandmapEntry) {
	children := v.entryList.GetChildren()
	children.Foreach(func(child any) {
		w := child.(gtk.IWidget)
		w.ToWidget().Destroy()

	})

	for _, entry := range entries {
		w := v.newListEntry(entry)
		if w != nil {
			v.entryList.Add(w)
		}
	}
}

func (v *spotsView) newListEntry(entry core.BandmapEntry) *gtk.Widget {
	root, _ := gtk.ListBoxRowNew()
	root.SetHExpand(true)

	layout, _ := gtk.GridNew()
	layout.SetHExpand(true)
	style.AddClass(&layout.Widget, spotRowClass)
	style.AddClass(&layout.Widget, entrySourceStyles[entry.Source])
	root.Add(layout)

	proximityIndicator := newProximityIndicator(root, v.getProximity)
	proximityIndicatorArea, _ := gtk.DrawingAreaNew()
	proximityIndicatorArea.SetHExpand(false)
	proximityIndicatorArea.SetHAlign(gtk.ALIGN_FILL)
	proximityIndicatorArea.SetVAlign(gtk.ALIGN_FILL)
	proximityIndicatorArea.SetSizeRequest(10, -1)
	proximityIndicatorArea.Connect("draw", proximityIndicator.Draw)
	layout.Attach(proximityIndicatorArea, 0, 0, 1, 5)

	frequency, _ := gtk.LabelNew("")
	frequency.SetHExpand(true)
	frequency.SetHAlign(gtk.ALIGN_START)
	style.AddClass(&frequency.Widget, spotFrequencyClass)
	layout.Attach(frequency, 1, 0, 1, 1)

	call, _ := gtk.LabelNew(entry.Call.String())
	call.SetHExpand(true)
	call.SetHAlign(gtk.ALIGN_START)
	style.AddClass(&call.Widget, spotCallClass)
	layout.Attach(call, 1, 1, 1, 2)

	var geoInfoText string
	if entry.Info.PrimaryPrefix != "" {
		geoInfoText = fmt.Sprintf("%s (%s), %s, ITU %d, CQ %d", entry.Info.DXCCName, entry.Info.PrimaryPrefix, entry.Info.Continent, entry.Info.ITUZone, entry.Info.CQZone)
	}
	geoInfo, _ := gtk.LabelNew(geoInfoText)
	geoInfo.SetHExpand(true)
	geoInfo.SetHAlign(gtk.ALIGN_START)
	style.AddClass(&geoInfo.Widget, spotGeoInfoClass)
	layout.Attach(geoInfo, 1, 3, 1, 1)

	predictedExchange, _ := gtk.LabelNew(entry.Info.ExchangeText)
	predictedExchange.SetHExpand(true)
	predictedExchange.SetHAlign(gtk.ALIGN_END)
	// v.style.applyTo(&predictedExchange.Widget)
	style.AddClass(&predictedExchange.Widget, spotPredictedExchangeClass)
	layout.Attach(predictedExchange, 2, 0, 1, 2)

	score, _ := gtk.LabelNew("")
	score.SetHExpand(true)
	score.SetHAlign(gtk.ALIGN_END)
	// v.style.applyTo(&score.Widget)
	style.AddClass(&score.Widget, spotScoreClass)
	layout.Attach(score, 2, 2, 1, 2)

	lifetimeIndicator := newLifetimeIndicator(root, v.getRemainingLifetime)
	lifetimeIndicatorArea, _ := gtk.DrawingAreaNew()
	lifetimeIndicatorArea.SetHExpand(true)
	lifetimeIndicatorArea.SetHAlign(gtk.ALIGN_FILL)
	lifetimeIndicatorArea.SetVAlign(gtk.ALIGN_FILL)
	lifetimeIndicatorArea.SetSizeRequest(-1, 10)
	lifetimeIndicatorArea.Connect("draw", lifetimeIndicator.Draw)
	layout.Attach(lifetimeIndicatorArea, 1, 4, 3, 1)

	updateListEntry(root, entry)

	return root.ToWidget()
}

func updateListEntry(row *gtk.ListBoxRow, entry core.BandmapEntry) {
	child, _ := row.GetChild()
	layout := child.(*gtk.Grid)
	for _, class := range entrySourceStyles {
		style.RemoveClass(&layout.Widget, class)
	}
	style.AddClass(&layout.Widget, entrySourceStyles[entry.Source])

	child, _ = layout.GetChildAt(1, 0)
	frequency := child.(*gtk.Label)
	frequency.SetText(entry.Frequency.String())

	child, _ = layout.GetChildAt(2, 2)
	score := child.(*gtk.Label)
	score.SetText(fmt.Sprintf("%dP %dM", entry.Info.Points, entry.Info.Multis))
}

func (v *spotsView) EntryAdded(entry core.BandmapEntry) {
	if v == nil {
		return
	}
	runAsync(func() {
		w := v.newListEntry(entry)
		if w == nil {
			return
		}
		v.entryList.Insert(w, entry.Index)
	})
}

func (v *spotsView) EntryUpdated(entry core.BandmapEntry) {
	if v == nil {
		return
	}
	runAsync(func() {
		row := v.entryList.GetRowAtIndex(entry.Index)
		if row == nil {
			return
		}
		updateListEntry(row, entry)
	})
}

func (v *spotsView) EntryRemoved(entry core.BandmapEntry) {
	if v == nil {
		return
	}
	runAsync(func() {
		row := v.entryList.GetRowAtIndex(entry.Index)
		if row == nil {
			return
		}
		row.ToWidget().Destroy()
	})
}

func (v *spotsView) onRowSelected(listBox *gtk.ListBox, row *gtk.ListBoxRow) {
	v.ignoreSelection = true
	defer func() {
		v.ignoreSelection = false
	}()

	v.controller.SelectEntry(row.GetIndex())
}

func (v *spotsView) EntrySelected(entry core.BandmapEntry) {
	if v.ignoreSelection {
		return
	}
	v.RevealEntry(entry)
}

func (v *spotsView) RevealEntry(entry core.BandmapEntry) {
	if v == nil {
		return
	}
	runAsync(func() {
		row := v.entryList.GetRowAtIndex(entry.Index)
		if row == nil {
			return
		}

		_, y, err := row.TranslateCoordinates(v.entryList, 0, 0)
		if err != nil {
			log.Printf("cannot translate list row box coordinates: %v", err)
			return
		}

		adj := v.entryList.GetAdjustment()
		if adj == nil {
			return
		}

		_, rowHeight := row.GetPreferredHeight()
		adj.SetValue(float64(y) - (adj.GetPageSize()-float64(rowHeight))/2)
	})
}

func (v *spotsView) selectBand(band core.Band) func(*gtk.Button, *gdk.Event) {
	return func(button *gtk.Button, event *gdk.Event) {
		buttonEvent := gdk.EventButtonNewFromEvent(event)
		if buttonEvent.Button() != gdk.BUTTON_PRIMARY {
			return
		}

		switch buttonEvent.Type() {
		case gdk.EVENT_BUTTON_PRESS:
			v.controller.SetVisibleBand(band)
		case gdk.EVENT_DOUBLE_BUTTON_PRESS:
			v.controller.SetActiveBand(band)
		}
	}
}
