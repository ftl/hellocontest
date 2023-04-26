package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"
)

const BandmapWindowID = "bandmap"

type bandmapWindow struct {
	*bandmapView

	window   *gtk.Window
	geometry *gmtry.Geometry
}

func setupBandmapWindow(geometry *gmtry.Geometry) *bandmapWindow {
	result := &bandmapWindow{
		geometry: geometry,
	}

	return result
}

func (w *bandmapWindow) RestoreVisibility() {
	visible := w.geometry.Get(BandmapWindowID).Visible
	if visible {
		w.Show()
	} else {
		w.Hide()
	}
}

func (w *bandmapWindow) Show() {
	if w.window == nil {
		builder := setupBuilder()
		w.window = getUI(builder, "bandmapWindow").(*gtk.Window)
		w.window.SetDefaultSize(200, 900)
		w.window.SetTitle("Bandmap")
		w.window.Connect("destroy", w.onDestroy)
		w.bandmapView = setupBandmapView(builder)
		connectToGeometry(w.geometry, BandmapWindowID, w.window)
	}
	w.window.ShowAll()
	w.window.Present()
}

func (w *bandmapWindow) Hide() {
	if w.window == nil {
		return
	}
	w.window.Close()
}

func (w *bandmapWindow) Visible() bool {
	if w.window == nil {
		return false
	}
	return w.window.IsVisible()
}

func (w *bandmapWindow) UseDefaultWindowGeometry() {
	if w.window == nil {
		return
	}
	w.window.Move(0, 100)
	w.window.Resize(200, 900)
}

func (w *bandmapWindow) onDestroy() {
	w.window = nil
	w.bandmapView = nil
}
