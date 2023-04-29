package ui

import (
	"log"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

type bandmapView struct {
	entryList *gtk.ListBox
	style     *style

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
	.exchangePrediction{
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

	geoInfoText := "DL, EU, ITU 28, CQ 14, 8°"
	geoInfo, _ := gtk.LabelNew(geoInfoText)
	geoInfo.SetHExpand(true)
	geoInfo.SetHAlign(gtk.ALIGN_START)
	v.style.applyTo(&geoInfo.Widget)
	addStyleClass(&geoInfo.Widget, "geoInfo")
	layout.Attach(geoInfo, 1, 2, 1, 1)

	exchangePredictionText := "Hans DL"
	exchangePrediction, _ := gtk.LabelNew(exchangePredictionText)
	exchangePrediction.SetHExpand(true)
	exchangePrediction.SetHAlign(gtk.ALIGN_END)
	exchangePrediction.SetVAlign(gtk.ALIGN_START)
	v.style.applyTo(&exchangePrediction.Widget)
	addStyleClass(&exchangePrediction.Widget, "exchangePrediction")
	layout.Attach(exchangePrediction, 2, 1, 1, 1)

	scoreText := "1 P, 0 M"
	score, _ := gtk.LabelNew(scoreText)
	score.SetHExpand(true)
	score.SetHAlign(gtk.ALIGN_END)
	v.style.applyTo(&score.Widget)
	addStyleClass(&score.Widget, "score")
	layout.Attach(score, 2, 2, 1, 1)

	ageIndicator, _ := gtk.DrawingAreaNew()
	ageIndicator.SetHExpand(true)
	ageIndicator.SetHAlign(gtk.ALIGN_FILL)
	ageIndicator.SetVAlign(gtk.ALIGN_FILL)
	layout.Attach(ageIndicator, 0, 3, 4, 1)

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
		v.entryList.ShowAll()
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
