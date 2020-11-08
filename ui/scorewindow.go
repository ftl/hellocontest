package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"
)

type scoreWindow struct {
	window *gtk.Window

	*scoreView
}

func setupScoreWindow(builder *gtk.Builder) *scoreWindow {
	result := new(scoreWindow)

	result.window = getUI(builder, "scoreWindow").(*gtk.Window)
	result.window.SetDefaultSize(300, 500)
	result.window.SetTitle("Score")

	result.scoreView = setupScoreView(builder)

	return result
}

func (w *scoreWindow) Show() {
	w.window.ShowAll()
}

func (w *scoreWindow) Hide() {
	w.window.Close()
}

func (w *scoreWindow) Visible() bool {
	return w.window.IsVisible()
}

func (w *scoreWindow) UseDefaultWindowGeometry() {
	w.window.Move(0, 100)
	w.window.Resize(300, 500)
}

func (w *scoreWindow) ConnectToGeometry(geometry *gmtry.Geometry) {
	connectToGeometry(geometry, "score", w.window)
}
