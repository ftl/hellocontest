package ui

import (
	"fmt"
	"log"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

type bandmapView struct {
	entryList *gtk.ListBox
}

func newBandmapView() *bandmapView {
	return &bandmapView{}
}

func (v *bandmapView) setup(builder *gtk.Builder) {
	v.entryList = getUI(builder, "entryList").(*gtk.ListBox)

	// TODO connect signals
}

func (v *bandmapView) ShowFrame(frame core.BandmapFrame) {
	if v.entryList == nil {
		return
	}

	runAsync(func() {
		children := v.entryList.GetChildren()
		log.Printf("old children count: %d", children.Length())
		children.Foreach(func(child any) {
			w := child.(gtk.IWidget)
			v.entryList.Remove(w)
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
