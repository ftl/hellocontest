package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"
)

const CallinfoWindowID = gmtry.ID("callinfo")

type callinfoWindow struct {
	*callinfoView

	window   *gtk.Window
	geometry *gmtry.Geometry
}

func setupCallinfoWindow(geometry *gmtry.Geometry) *callinfoWindow {
	result := &callinfoWindow{
		geometry: geometry,
	}

	return result
}

func (w *callinfoWindow) RestoreVisibility() {
	visible := w.geometry.Get(CallinfoWindowID).Visible
	if visible {
		w.Show()
	} else {
		w.Hide()
	}
}

func (w *callinfoWindow) Show() {
	if w.window == nil {
		builder := setupBuilder()
		w.window = getUI(builder, "callinfoWindow").(*gtk.Window)
		w.window.SetDefaultSize(300, 500)
		w.window.SetTitle("Callsign Information")
		w.window.Connect("destroy", w.onDestroy)
		w.callinfoView = setupCallinfoView(builder)
		connectToGeometry(w.geometry, CallinfoWindowID, w.window)
	}
	w.window.ShowAll()
	w.window.Present()
}

func (w *callinfoWindow) Hide() {
	if w.window == nil {
		return
	}
	w.window.Close()
}

func (w *callinfoWindow) Visible() bool {
	if w.window == nil {
		return false
	}
	return w.window.IsVisible()
}

func (w *callinfoWindow) UseDefaultWindowGeometry() {
	if w.window == nil {
		return
	}
	w.window.Move(0, 100)
	w.window.Resize(300, 500)
}

func (w *callinfoWindow) onDestroy() {
	w.window = nil
	w.callinfoView = nil
}
