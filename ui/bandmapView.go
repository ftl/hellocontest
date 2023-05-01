package ui

import (
	"fmt"
	"log"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

type bandmapView struct {
	entryList *gtk.ListBox
	style     *style

	currentFrame      core.BandmapFrame
	initialFrameShown bool
}

func setupBandmapView(builder *gtk.Builder) *bandmapView {
	result := &bandmapView{
		initialFrameShown: false,
	}

	result.entryList = getUI(builder, "entryList").(*gtk.ListBox)

	// TODO connect signals

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
		font-size: xx-large;
	}
	.predictedExchange{
		font-size: large;
	}
	.geoInfo{
		font-size: small;
	}
	.score{
		font-size: small;
	}
	`)
	result.style.applyTo(&result.entryList.Widget)

	return result
}

func (v *bandmapView) ShowFrame(frame core.BandmapFrame) {
	if v == nil {
		return
	}
	v.currentFrame = frame

	if v.initialFrameShown {
		return
	}
	v.initialFrameShown = true

	runAsync(func() {
		children := v.entryList.GetChildren()

		children.Foreach(func(child any) {
			w := child.(gtk.IWidget)
			w.ToWidget().Destroy()

		})

		for _, entry := range frame.Entries {
			w := v.newListEntry(entry)
			if w != nil {
				log.Printf("Entry: %s", entry.Call.String())
				v.entryList.Add(w)
			}
		}
		log.Printf("new children count: %d", v.entryList.GetChildren().Length())
		v.entryList.ShowAll()
	})
}

func (v *bandmapView) getLifetime(index int) float64 {
	if index < 0 || index >= len(v.currentFrame.Entries) {
		return 0
	}
	return v.currentFrame.Entries[index].Lifetime
}

var sourceStyles = map[core.SpotType]string{
	core.WorkedSpot:  "workedSpot",
	core.ManualSpot:  "manualSpot",
	core.SkimmerSpot: "skimmerSpot",
	core.RBNSpot:     "rbnSpot",
	core.ClusterSpot: "clusterSpot",
}

func (v *bandmapView) newListEntry(entry core.BandmapEntry) *gtk.Widget {
	root, _ := gtk.ListBoxRowNew()
	root.SetHExpand(true)

	layout, _ := gtk.GridNew()
	layout.SetHExpand(true)
	v.style.applyTo(&layout.Widget)
	addStyleClass(&layout.Widget, "row")
	addStyleClass(&layout.Widget, sourceStyles[entry.Source])
	root.Add(layout)

	proximityIndicator, _ := gtk.LabelNew("|")
	layout.Attach(proximityIndicator, 0, 0, 1, 3)

	frequency, _ := gtk.LabelNew(entry.Frequency.String())
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
	layout.Attach(call, 1, 1, 1, 1)

	var geoInfoText string
	if entry.Info.PrimaryPrefix != "" {
		geoInfoText = fmt.Sprintf("%s (%s), %s, ITU %d, CQ %d", entry.Info.DXCCName, entry.Info.PrimaryPrefix, entry.Info.Continent, entry.Info.ITUZone, entry.Info.CQZone)
	}
	geoInfo, _ := gtk.LabelNew(geoInfoText)
	geoInfo.SetHExpand(true)
	geoInfo.SetHAlign(gtk.ALIGN_START)
	v.style.applyTo(&geoInfo.Widget)
	addStyleClass(&geoInfo.Widget, "geoInfo")
	layout.Attach(geoInfo, 1, 2, 1, 1)

	predictedExchange, _ := gtk.LabelNew(entry.Info.ExchangeText)
	predictedExchange.SetHExpand(true)
	predictedExchange.SetHAlign(gtk.ALIGN_END)
	predictedExchange.SetVAlign(gtk.ALIGN_START)
	v.style.applyTo(&predictedExchange.Widget)
	addStyleClass(&predictedExchange.Widget, "predictedExchange")
	layout.Attach(predictedExchange, 2, 1, 1, 1)

	score, _ := gtk.LabelNew("")
	score.SetHExpand(true)
	score.SetHAlign(gtk.ALIGN_END)
	v.style.applyTo(&score.Widget)
	addStyleClass(&score.Widget, "score")
	layout.Attach(score, 2, 2, 1, 1)
	setScore(root, entry.Info.Points, entry.Info.Multis)

	lifetimeIndicator := newLifetimeIndicator(root, v.getLifetime)

	lifetimeIndicatorArea, _ := gtk.DrawingAreaNew()
	lifetimeIndicatorArea.SetHExpand(true)
	lifetimeIndicatorArea.SetHAlign(gtk.ALIGN_FILL)
	lifetimeIndicatorArea.SetVAlign(gtk.ALIGN_FILL)
	lifetimeIndicatorArea.SetSizeRequest(-1, 10)
	lifetimeIndicatorArea.Connect("draw", lifetimeIndicator.Draw)
	layout.Attach(lifetimeIndicatorArea, 0, 3, 4, 1)

	return root.ToWidget()
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
		log.Printf("New Entry @ %d: %s", entry.Index, entry.Call.String())
		v.entryList.Insert(w, entry.Index)
		v.entryList.QueueDraw()
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
		setScore(row, entry.Info.Points, entry.Info.Multis)
		setLifetime(row, entry.Lifetime)
		row.ShowAll()
		log.Printf("Updated Entry @ %d: %s", entry.Index, entry.Call.String())
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
		v.entryList.ShowAll()
		log.Printf("Removed Entry @ %d: %s", entry.Index, entry.Call.String())
	})
}

func setScore(row *gtk.ListBoxRow, points, multis int) {
	child, _ := row.GetChild()
	layout := child.(*gtk.Grid)

	child, _ = layout.GetChildAt(2, 2)
	score := child.(*gtk.Label)
	score.SetText(fmt.Sprintf("%dP %dM", points, multis))
}

func setLifetime(row *gtk.ListBoxRow, lifetime float64) {
	child, _ := row.GetChild()
	layout := child.(*gtk.Grid)

	child, _ = layout.GetChildAt(0, 3)
	lifetimeIndicator, _ := child.(*gtk.DrawingArea)
	lifetimeIndicator.QueueDraw()
}
