package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/ui/style"
)

const CallinfoWindowID = gmtry.ID("callinfo")

type callinfoWindow struct {
	*callinfoView
	controller CallinfoController

	window   *gtk.Window
	geometry *gmtry.Geometry
	style    *style.Style
}

func setupCallinfoWindow(geometry *gmtry.Geometry, style *style.Style, controller CallinfoController) *callinfoWindow {
	result := &callinfoWindow{
		controller: controller,
		geometry:   geometry,
		style:      style,
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
		w.callinfoView = setupCallinfoView(builder, w.style.ForWidget(w.window.ToWidget()), w.controller)
		w.window.Connect("style-updated", w.callinfoView.RefreshStyle)
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
