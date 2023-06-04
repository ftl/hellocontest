package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/ui/style"
)

const ScoreWindowID = "score"

type scoreWindow struct {
	window   *gtk.Window
	geometry *gmtry.Geometry
	style    *style.Style

	*scoreView
}

func setupScoreWindow(geometry *gmtry.Geometry, style *style.Style) *scoreWindow {
	result := &scoreWindow{
		geometry: geometry,
		style:    style,
	}

	return result
}

func (w *scoreWindow) RestoreVisibility() {
	visible := w.geometry.Get(ScoreWindowID).Visible
	if visible {
		w.Show()
	} else {
		w.Hide()
	}
}

func (w *scoreWindow) Show() {
	if w.window == nil {
		builder := setupBuilder()
		w.window = getUI(builder, "scoreWindow").(*gtk.Window)
		w.window.SetDefaultSize(300, 500)
		w.window.SetTitle("Score")
		w.window.Connect("destroy", w.onDestroy)
		w.scoreView = setupNewScoreView(builder, w.style.ForWidget(w.window.ToWidget()))
		connectToGeometry(w.geometry, ScoreWindowID, w.window)
	}
	w.window.ShowAll()
	w.window.Present()
}

func (w *scoreWindow) Hide() {
	if w.window == nil {
		return
	}
	w.window.Close()
}

func (w *scoreWindow) Visible() bool {
	if w.window == nil {
		return false
	}
	return w.window.IsVisible()
}

func (w *scoreWindow) UseDefaultWindowGeometry() {
	if w.window == nil {
		return
	}
	w.window.Move(0, 100)
	w.window.Resize(300, 500)
}

func (w *scoreWindow) onDestroy() {
	w.window = nil
	w.scoreView = nil
}
