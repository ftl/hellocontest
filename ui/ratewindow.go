package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"
)

const RateWindowID = "rate"

type rateWindow struct {
	window   *gtk.Window
	geometry *gmtry.Geometry

	*rateView
}

func setupRateWindow(geometry *gmtry.Geometry) *rateWindow {
	result := &rateWindow{
		geometry: geometry,
	}

	return result
}

func (w *rateWindow) RestoreVisibility() {
	visible := w.geometry.Get(RateWindowID).Visible
	if visible {
		w.Show()
	} else {
		w.Hide()
	}
}

func (w *rateWindow) Show() {
	if w.window == nil {
		builder := setupBuilder()
		w.window = getUI(builder, "rateWindow").(*gtk.Window)
		w.window.SetDefaultSize(300, 500)
		w.window.SetTitle("QSO Rate")
		w.window.Connect("destroy", w.onDestroy)
		w.rateView = setupRateView(builder)
		connectToGeometry(w.geometry, RateWindowID, w.window)
	}
	w.window.ShowAll()
}

func (w *rateWindow) Hide() {
	if w.window == nil {
		return
	}
	w.window.Close()
}

func (w *rateWindow) Visible() bool {
	if w.window == nil {
		return false
	}
	return w.window.IsVisible()
}

func (w *rateWindow) UseDefaultWindowGeometry() {
	if w.window == nil {
		return
	}
	w.window.Move(0, 100)
	w.window.Resize(300, 500)
}

func (w *rateWindow) onDestroy() {
	w.window = nil
	w.rateView = nil
}
