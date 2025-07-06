package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

const ScoreWindowID = "score"

type scoreWindow struct {
	clock     core.Clock
	scoreView *scoreView

	window      *gtk.Window
	geometry    *gmtry.Geometry
	style       *style.Style
	acceptFocus bool

	score      core.Score
	rate       core.QSORate
	pointsGoal int
	multisGoal int
}

func setupScoreWindow(geometry *gmtry.Geometry, style *style.Style, clock core.Clock) *scoreWindow {
	result := &scoreWindow{
		clock:    clock,
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
		w.window, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		w.window.SetDefaultSize(300, 500)
		w.window.SetTitle("Score")
		w.window.SetCanFocus(false)
		w.window.SetAcceptFocus(w.acceptFocus)
		w.window.Connect("destroy", w.onDestroy)
		w.scoreView = setupNewScoreView(w.style.ForWidget(w.window.ToWidget()), w.clock)
		w.scoreView.SetGoals(w.pointsGoal, w.multisGoal)
		w.scoreView.ShowScore(w.score)
		w.scoreView.RateUpdated(w.rate)

		w.window.Add(w.scoreView.rootGrid.ToWidget())
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

func (w *scoreWindow) SetAcceptFocus(acceptFocus bool) {
	w.acceptFocus = acceptFocus
	if w.window == nil {
		return
	}
	w.window.SetAcceptFocus(w.acceptFocus)
}

func (w *scoreWindow) ShowScore(score core.Score) {
	w.score = score
	if w.scoreView != nil {
		w.scoreView.ShowScore(score)
	}
}

func (w *scoreWindow) SetGoals(points int, multis int) {
	w.pointsGoal = points
	w.multisGoal = multis
	if w.scoreView != nil {
		w.scoreView.SetGoals(points, multis)
	}
}

func (w *scoreWindow) RateUpdated(rate core.QSORate) {
	w.rate = rate
	if w.scoreView != nil {
		w.scoreView.RateUpdated(rate)
	}
}
