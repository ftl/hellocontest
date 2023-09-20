package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

const ScoreWindowID = "score"

type scoreWindow struct {
	scoreView *scoreView

	window   *gtk.Window
	geometry *gmtry.Geometry
	style    *style.Style

	score      core.Score
	rate       core.QSORate
	pointsGoal int
	multisGoal int
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
		w.scoreView.SetGoals(w.pointsGoal, w.multisGoal)
		w.scoreView.ShowScore(w.score)
		w.scoreView.RateUpdated(w.rate)
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
