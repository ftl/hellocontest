package ui

import (
	"github.com/ftl/gmtry"
	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/ui/style"
)

const SpotsWindowID = "spots"

type spotsWindow struct {
	spotsView *spotsView

	window      *gtk.Window
	geometry    *gmtry.Geometry
	style       *style.Style
	acceptFocus bool

	controller SpotsController
}

func setupSpotsWindow(geometry *gmtry.Geometry, style *style.Style, controller SpotsController) *spotsWindow {
	result := &spotsWindow{
		geometry:   geometry,
		style:      style,
		controller: controller,
	}

	return result
}

func (w *spotsWindow) RestoreVisibility() {
	visible := w.geometry.Get(SpotsWindowID).Visible
	if visible {
		w.Show()
	} else {
		w.Hide()
	}
}

func (w *spotsWindow) Show() {
	if w.window == nil {
		builder := setupBuilder()
		w.window = getUI(builder, "spotsWindow").(*gtk.Window)
		w.window.SetDefaultSize(400, 900)
		w.window.SetTitle("Spots")
		w.window.SetAcceptFocus(w.acceptFocus)
		w.window.Connect("destroy", w.onDestroy)
		w.spotsView = setupSpotsView(builder, w.style.ForWidget(w.window.ToWidget()), w.controller)
		connectToGeometry(w.geometry, SpotsWindowID, w.window)
	}
	w.window.ShowAll()
	w.window.Present()
}

func (w *spotsWindow) Hide() {
	if w.window == nil {
		return
	}
	w.window.Close()
}

func (w *spotsWindow) Visible() bool {
	if w.window == nil {
		return false
	}
	return w.window.IsVisible()
}

func (w *spotsWindow) UseDefaultWindowGeometry() {
	if w.window == nil {
		return
	}
	w.window.Move(0, 100)
	w.window.Resize(200, 900)
}

func (w *spotsWindow) onDestroy() {
	w.window = nil
	w.spotsView = nil
}

func (w *spotsWindow) SetAcceptFocus(acceptFocus bool) {
	w.acceptFocus = true
	if w.window == nil {
		return
	}
	w.window.SetAcceptFocus(w.acceptFocus)
}

func (w *spotsWindow) ShowFrame(frame core.BandmapFrame) {
	if w.spotsView != nil {
		w.spotsView.ShowFrame(frame)
	}
}

func (w *spotsWindow) EntryAdded(entry core.BandmapEntry) {
	if w.spotsView != nil {
		w.spotsView.EntryAdded(entry)
	}
}

func (w *spotsWindow) EntryUpdated(entry core.BandmapEntry) {
	if w.spotsView != nil {
		w.spotsView.EntryUpdated(entry)
	}
}

func (w *spotsWindow) EntryRemoved(entry core.BandmapEntry) {
	if w.spotsView != nil {
		w.spotsView.EntryRemoved(entry)
	}
}

func (w *spotsWindow) EntrySelected(entry core.BandmapEntry) {
	if w.spotsView != nil {
		w.spotsView.EntrySelected(entry)
	}
}

func (w *spotsWindow) RevealEntry(entry core.BandmapEntry) {
	if w.spotsView != nil {
		w.spotsView.RevealEntry(entry)
	}
}
