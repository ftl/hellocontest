package ui

import (
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
	// TODO implement
}
