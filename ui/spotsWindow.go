package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"
)

const SpotsWindowID = "spots"

type spotsWindow struct {
	*spotsView

	window   *gtk.Window
	geometry *gmtry.Geometry

	controller SpotsController
}

func setupSpotsWindow(geometry *gmtry.Geometry, controller SpotsController) *spotsWindow {
	result := &spotsWindow{
		geometry:   geometry,
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
		w.window.Connect("destroy", w.onDestroy)
		w.spotsView = setupSpotsView(builder, w.controller)
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