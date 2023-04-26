package ui

import (
	"fmt"
	"log"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

type bandmapView struct {
	entryList *gtk.ListBox

	initialFrameShown bool
}

func setupBandmapView(builder *gtk.Builder) *bandmapView {
	result := &bandmapView{
		initialFrameShown: true,
	}

	result.entryList = getUI(builder, "entryList").(*gtk.ListBox)

	// TODO connect signals

	return result
}

func (v *bandmapView) ShowFrame(frame core.BandmapFrame) {
	if v == nil {
		return
	}
	// if v.initialFrameShown {
	// 	return
	// }
	// v.initialFrameShown = true

	runAsync(func() {
		children := v.entryList.GetChildren()
		log.Printf("new frame: %d entries vs old children count: %d", len(frame.Entries), children.Length())
		if true {
			return
		}

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

func (v *bandmapView) newListEntry(entry core.BandmapEntry) *gtk.Widget {
	text := fmt.Sprintf("%s:%s", entry.Call, entry.Frequency)
	result, err := gtk.LabelNew(text)
	if err != nil {
		return nil
	}
	return result.ToWidget()
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
