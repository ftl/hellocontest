package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

const RateWindowID = "rate"

type rateWindow struct {
	window   *gtk.Window
	geometry *gmtry.Geometry
	style    *style.Style

	rate       core.QSORate
	qsosGoal   int
	pointsGoal int
	multisGoal int

	*rateView
}

func setupRateWindow(geometry *gmtry.Geometry, style *style.Style) *rateWindow {
	result := &rateWindow{
		geometry: geometry,
		style:    style,
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
		w.rateView = setupNewRateView(builder, w.style.ForWidget(w.window.ToWidget()))
		w.rateView.SetGoals(w.qsosGoal, w.pointsGoal, w.multisGoal)
		w.rateView.ShowRate(w.rate)
		connectToGeometry(w.geometry, RateWindowID, w.window)
	}
	w.window.ShowAll()
	w.window.Present()
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

func (w *rateWindow) ShowRate(rate core.QSORate) {
	w.rate = rate
	if w.rateView != nil {
		w.rateView.ShowRate(rate)
	}
}

func (w *rateWindow) SetGoals(qsos int, points int, multis int) {
	w.qsosGoal = qsos
	w.pointsGoal = points
	w.multisGoal = multis
	if w.rateView != nil {
		w.rateView.SetGoals(qsos, points, multis)
	}
}
