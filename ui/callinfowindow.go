package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"
)

type callinfoWindow struct {
	window *gtk.Window

	*callinfoView
}

func setupCallinfoWindow(builder *gtk.Builder) *callinfoWindow {
	result := new(callinfoWindow)

	result.window = getUI(builder, "callinfoWindow").(*gtk.Window)
	result.window.SetDefaultSize(300, 500)
	result.window.SetTitle("Callsign Information")

	result.callinfoView = setupCallinfoView(builder)

	return result
}

func (w *callinfoWindow) Show() {
	w.window.ShowAll()
}

func (w *callinfoWindow) Hide() {
	w.window.Close()
}

func (w *callinfoWindow) Visible() bool {
	return w.window.IsVisible()
}

func (w *callinfoWindow) UseDefaultWindowGeometry() {
	w.window.Move(0, 100)
	w.window.Resize(300, 500)
}

func (w *callinfoWindow) ConnectToGeometry(geometry *gmtry.Geometry) {
	connectToGeometry(geometry, "callinfo", w.window)
}
