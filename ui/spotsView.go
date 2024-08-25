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
)

type SpotsController interface {
	SetVisibleBand(core.Band)
	SetActiveBand(core.Band)

	SelectEntry(core.BandmapEntryID)
}

type spotsView struct {
	controller SpotsController
	colors     colorProvider

	bandGrid *gtk.Grid

	table        *gtk.TreeView
	tableContent *gtk.ListStore

	bands                []core.BandSummary
	bandsID              string
	currentFrame         core.BandmapFrame
	initialFrameShown    bool
	tableSelectionActive bool
}

func setupSpotsView(builder *gtk.Builder, colors colorProvider, controller SpotsController) *spotsView {
	result := &spotsView{
		controller:        controller,
		colors:            colors,
		initialFrameShown: false,
	}

	result.bandGrid = getUI(builder, "bandGrid").(*gtk.Grid)

	setupSpotsTableView(result, builder, controller)

	return result
}

func (v *spotsView) ShowFrame(frame core.BandmapFrame) {
	runAsync(func() {
		// frequencyChanged := v.currentFrame.Frequency != frame.Frequency
		bandChanged := (v.currentFrame.ActiveBand != frame.ActiveBand) || (v.currentFrame.VisibleBand != frame.VisibleBand)
		selectionChanged := v.currentFrame.SelectedEntry.ID != frame.SelectedEntry.ID
		isEntrySelected := frame.SelectedEntry.ID != core.NoEntryID

		oldFrame := v.currentFrame
		v.currentFrame = frame
		v.setupBands(frame.Bands)
		v.updateBands(frame.Bands)

		if bandChanged {
			v.showFrameInTable(frame)
		} else {
			v.updateTableFromFrame(oldFrame, frame)
		}

		if selectionChanged || bandChanged {
			if isEntrySelected {
				v.setSelectedTableEntry(frame.SelectedEntry)
			} else {
				v.clearSelection()
			}
		}

		// if (bandChanged || frequencyChanged) && v.currentFrame.RevealNearestEntry {
		// 	v.revealTableEntry(v.currentFrame.NearestEntry)
		// }
	})
}

func (v *spotsView) updateTableFromFrame(oldFrame core.BandmapFrame, newFrame core.BandmapFrame) {
	visitedEntries := make(map[core.BandmapEntryID]bool, len(newFrame.Entries))
	newEntries := make([]core.BandmapEntryID, 0, len(newFrame.Entries))
	for _, entry := range newFrame.Entries {
		visitedEntries[entry.ID] = true
		oldIndex, existed := oldFrame.IndexOf(entry.ID)
		if existed {
			v.updateTableEntry(oldIndex, entry)
		} else {
			newEntries = append(newEntries, entry.ID)
		}
	}

	for i := len(oldFrame.Entries) - 1; i >= 0; i-- {
		entry := oldFrame.Entries[i]
		visited := visitedEntries[entry.ID]
		if !visited {
			v.removeTableEntry(i)
		}
	}

	for _, id := range newEntries {
		index, found := newFrame.IndexOf(id)
		if !found {
			log.Printf("cannot find new entry with ID %d", id)
			continue
		}
		v.addTableEntry(index, newFrame.Entries[index])
	}
}

func (v *spotsView) setupBands(bands []core.BandSummary) {
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
	multis.SetText(fmt.Sprintf("%dM", band.Multis()))

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
