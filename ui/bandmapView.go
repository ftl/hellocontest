package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

var entrySourceStyles = map[core.SpotType]string{
	core.WorkedSpot:  "workedSpot",
	core.ManualSpot:  "manualSpot",
	core.SkimmerSpot: "skimmerSpot",
	core.RBNSpot:     "rbnSpot",
	core.ClusterSpot: "clusterSpot",
}

type BandmapController interface {
	SetVisibleBand(core.Band)
	SetActiveBand(core.Band)

	EntryVisible(int) bool
}

type bandmapView struct {
	controller BandmapController

	bandGrid  *gtk.Grid
	entryList *gtk.ListBox
	style     *style

	bands             []core.BandSummary
	bandsID           string
	currentFrame      core.BandmapFrame
	initialFrameShown bool
}

func setupBandmapView(builder *gtk.Builder, controller BandmapController) *bandmapView {
	result := &bandmapView{
		controller:        controller,
		initialFrameShown: false,
	}

	result.bandGrid = getUI(builder, "bandGrid").(*gtk.Grid)
	result.entryList = getUI(builder, "entryList").(*gtk.ListBox)
	result.entryList.SetFilterFunc(result.filterRow)

	result.style = newStyle(`
	.row{
		margin: 3px;
		padding: 3px;
		border: 2px solid black;
		color: black;
	}
	.workedSpot{
		background-color: rgba(128, 128, 128, 255);
	}
	.manualSpot{
		background-color: rgba(255, 255, 255, 255);
	}
	.skimmerSpot{
		background-color: rgba(255, 153, 255, 255);
	}
	.rbnSpot{
		background-color: rgba(255, 255, 153, 255);
	}
	.clusterSpot{
		background-color: rgba(153, 255, 255, 255);
	}
	.frequency{
		font-size: small;
	}
	.call {
		font-size: x-large;
		font-weight: bold;
	}
	.geoInfo{
		font-size: small;
	}
	.predictedExchange{
		font-size: x-large;
	}
	.score{
		font-size: x-large;
	}

	.band{
		margin: 3px;
		padding: 3px;
	}
	.bandActive{
		border-top: 4px solid blue;
	}
	.bandVisible{
		border-bottom: 4px solid black;
	}
	.maxValue{
		background-color: red;
		color: white;
		border: 1px solid red;
		border-radius: 3px;
		font-weight: bold;
	}
	.bandLabel{
		font-size: x-large;
	}
	.bandPoints{
		font-size: medium;
	}
	.bandMultis{
		font-size: medium;
	}
	`)
	result.style.applyTo(&result.entryList.Widget)

	return result
}

func (v *bandmapView) getLifetime(index int) float64 {
	if index < 0 || index >= len(v.currentFrame.Entries) {
		return 0
	}
	return v.currentFrame.Entries[index].Lifetime
}

func (v *bandmapView) getProximity(index int) float64 {
	if index < 0 || index >= len(v.currentFrame.Entries) {
		return 0
	}
	return v.currentFrame.Entries[index].ProximityFactor(v.currentFrame.Frequency)
}

func (v *bandmapView) filterRow(row *gtk.ListBoxRow) bool {
	return v.controller.EntryVisible(row.GetIndex())
}

func (v *bandmapView) ShowFrame(frame core.BandmapFrame) {
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

func (v *bandmapView) setupBands(bands []core.BandSummary) {
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

func (v *bandmapView) newBand(band core.BandSummary) *gtk.Widget {
	button, _ := gtk.ButtonNew()
	button.SetName("band")
	button.SetHAlign(gtk.ALIGN_END)
	button.SetVAlign(gtk.ALIGN_FILL)
	button.SetHExpand(true)
	v.style.applyTo(&button.Widget)
	addStyleClass(&button.Widget, "band")
	button.Connect("button-press-event", v.selectBand(band.Band))

	grid, _ := gtk.GridNew()
	grid.SetColumnSpacing(3)
	button.Add(grid)

	label, _ := gtk.LabelNew(string(band.Band))
	label.SetHAlign(gtk.ALIGN_END)
	label.SetVAlign(gtk.ALIGN_FILL)
	label.SetHExpand(true)
	v.style.applyTo(&label.Widget)
	addStyleClass(&label.Widget, "bandLabel")
	grid.Attach(label, 0, 0, 1, 2)

	points, _ := gtk.LabelNew("")
	points.SetHAlign(gtk.ALIGN_FILL)
	points.SetVAlign(gtk.ALIGN_FILL)
	points.SetHExpand(true)
	v.style.applyTo(&points.Widget)
	addStyleClass(&points.Widget, "bandPoints")
	grid.Attach(points, 1, 0, 1, 1)

	multis, _ := gtk.LabelNew("")
	multis.SetHAlign(gtk.ALIGN_FILL)
	multis.SetVAlign(gtk.ALIGN_FILL)
	points.SetHExpand(true)
	v.style.applyTo(&multis.Widget)
	addStyleClass(&multis.Widget, "bandPoints")
	grid.Attach(multis, 1, 1, 1, 1)

	v.updateBand(button, band)

	return button.ToWidget()
}

func (v *bandmapView) updateBand(button *gtk.Button, band core.BandSummary) {
	child, _ := button.GetChild()
	grid := child.(*gtk.Grid)

	child, _ = grid.GetChildAt(1, 0)
	points := child.(*gtk.Label)
	points.SetText(fmt.Sprintf("%dP", band.Points))

	child, _ = grid.GetChildAt(1, 1)
	multis := child.(*gtk.Label)
	multis.SetText(fmt.Sprintf("%dM", band.Multis))

	if band.MaxPoints {
		addStyleClass(&points.Widget, "maxValue")
	} else {
		removeStyleClass(&points.Widget, "maxValue")
	}
	if band.MaxMultis {
		addStyleClass(&multis.Widget, "maxValue")
	} else {
		removeStyleClass(&multis.Widget, "maxValue")
	}
	if band.Active {
		addStyleClass(&button.Widget, "bandActive")
	} else {
		removeStyleClass(&button.Widget, "bandActive")
	}
	if band.Visible {
		addStyleClass(&button.Widget, "bandVisible")
	} else {
		removeStyleClass(&button.Widget, "bandVisible")
	}
}

func (v *bandmapView) updateBands(bands []core.BandSummary) {
	for i, band := range bands {
		child, _ := v.bandGrid.GetChildAt(i, 0)
		button, ok := child.(*gtk.Button)
		if ok {
			v.updateBand(button, band)
		}
	}
}

func (v *bandmapView) showEntries(entries []core.BandmapEntry) {
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

func (v *bandmapView) newListEntry(entry core.BandmapEntry) *gtk.Widget {
	root, _ := gtk.ListBoxRowNew()
	root.SetHExpand(true)

	layout, _ := gtk.GridNew()
	layout.SetHExpand(true)
	v.style.applyTo(&layout.Widget)
	addStyleClass(&layout.Widget, "row")
	addStyleClass(&layout.Widget, entrySourceStyles[entry.Source])
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
	v.style.applyTo(&frequency.Widget)
	addStyleClass(&frequency.Widget, "frequency")
	layout.Attach(frequency, 1, 0, 1, 1)

	call, _ := gtk.LabelNew(entry.Call.String())
	call.SetHExpand(true)
	call.SetHAlign(gtk.ALIGN_START)
	v.style.applyTo(&call.Widget)
	addStyleClass(&call.Widget, "call")
	layout.Attach(call, 1, 1, 1, 2)

	var geoInfoText string
	if entry.Info.PrimaryPrefix != "" {
		geoInfoText = fmt.Sprintf("%s (%s), %s, ITU %d, CQ %d", entry.Info.DXCCName, entry.Info.PrimaryPrefix, entry.Info.Continent, entry.Info.ITUZone, entry.Info.CQZone)
	}
	geoInfo, _ := gtk.LabelNew(geoInfoText)
	geoInfo.SetHExpand(true)
	geoInfo.SetHAlign(gtk.ALIGN_START)
	v.style.applyTo(&geoInfo.Widget)
	addStyleClass(&geoInfo.Widget, "geoInfo")
	layout.Attach(geoInfo, 1, 3, 1, 1)

	predictedExchange, _ := gtk.LabelNew(entry.Info.ExchangeText)
	predictedExchange.SetHExpand(true)
	predictedExchange.SetHAlign(gtk.ALIGN_END)
	v.style.applyTo(&predictedExchange.Widget)
	addStyleClass(&predictedExchange.Widget, "predictedExchange")
	layout.Attach(predictedExchange, 2, 0, 1, 2)

	score, _ := gtk.LabelNew("")
	score.SetHExpand(true)
	score.SetHAlign(gtk.ALIGN_END)
	v.style.applyTo(&score.Widget)
	addStyleClass(&score.Widget, "score")
	layout.Attach(score, 2, 2, 1, 2)

	lifetimeIndicator := newLifetimeIndicator(root, v.getLifetime)
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
		removeStyleClass(&layout.Widget, class)
	}
	addStyleClass(&layout.Widget, entrySourceStyles[entry.Source])

	child, _ = layout.GetChildAt(1, 0)
	frequency := child.(*gtk.Label)
	frequency.SetText(entry.Frequency.String())

	child, _ = layout.GetChildAt(2, 2)
	score := child.(*gtk.Label)
	score.SetText(fmt.Sprintf("%dP %dM", entry.Info.Points, entry.Info.Multis))
}

func (v *bandmapView) EntryAdded(entry core.BandmapEntry) {
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

func (v *bandmapView) EntryUpdated(entry core.BandmapEntry) {
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

func (v *bandmapView) EntryRemoved(entry core.BandmapEntry) {
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

func (v *bandmapView) selectBand(band core.Band) func(*gtk.Button, *gdk.Event) {
	return func(button *gtk.Button, event *gdk.Event) {
		buttonEvent := gdk.EventButtonNewFromEvent(event)
		switch buttonEvent.Type() {
		case gdk.EVENT_BUTTON_PRESS:
			log.Printf("select %s as visible band: %v", band, button)
			v.controller.SetVisibleBand(band)
		case gdk.EVENT_DOUBLE_BUTTON_PRESS:
			log.Printf("select %s as active band: %v", band, button)
			v.controller.SetActiveBand(band)
		}
	}
}
