package ui

import (
	qtlib "github.com/mappu/miqt/qt6"
)

const (
	defaultSpotWindowWidth  = 600
	defaultSpotWindowHeight = 600
)

type spotWindow struct {
	window *qtlib.QMainWindow
}

func newSpotWindow(central *qtlib.QWidget) *spotWindow {
	w := &spotWindow{
		window: qtlib.NewQMainWindow2(),
	}
	w.window.SetWindowTitle("Hello Contest - Spots")
	w.window.SetObjectName(*qtlib.NewQAnyStringView3("spotWindow"))
	w.window.SetDockOptions(
		qtlib.QMainWindow__AllowNestedDocks |
			qtlib.QMainWindow__AllowTabbedDocks |
			qtlib.QMainWindow__AnimatedDocks |
			qtlib.QMainWindow__GroupedDragging,
	)
	w.window.SetCorner(qtlib.TopLeftCorner, qtlib.LeftDockWidgetArea)
	w.window.SetCorner(qtlib.BottomLeftCorner, qtlib.LeftDockWidgetArea)
	w.window.SetCorner(qtlib.TopRightCorner, qtlib.RightDockWidgetArea)
	w.window.SetCorner(qtlib.BottomRightCorner, qtlib.RightDockWidgetArea)
	// closing the window only hides it, the dock layout survives, and the Window menu
	// can bring it back. Vetoing the close event instead would also veto the
	// application's quit, which closes all top-level windows.
	w.window.SetCentralWidget(central)

	return w
}

func (w *spotWindow) Show() {
	w.window.SetVisible(true)
	w.window.Raise()
	w.window.ActivateWindow()
}

func (w *spotWindow) Hide() {
	w.window.SetVisible(false)
}

func (w *spotWindow) Visible() bool {
	return w.window.IsVisible()
}

func (w *spotWindow) RepaintForThemeChange() {
	w.window.SetPalette(qtlib.QGuiApplication_Palette())
	w.window.Update()
}

func (w *spotWindow) Close() {
	w.window.Close()
}
