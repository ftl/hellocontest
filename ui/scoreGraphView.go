package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/rate"
)

var (
	_ rate.RateUpdatedListener = (*scoreGraphView)(nil)
)

type scoreGraphView struct {
	*dockableView
	widget *qtlib.QFrame

	graph *scoreGraph
	score core.Score
}

func newScoreGraphView(parent *qtlib.QWidget, clock core.Clock) *scoreGraphView {
	v := &scoreGraphView{}
	v.widget = qtlib.NewQFrame2()
	v.widget.SetObjectName(*qtlib.NewQAnyStringView3("scoreGraph"))
	v.widget.SetFrameShape(qtlib.QFrame__StyledPanel)
	v.widget.SetFrameShadow(qtlib.QFrame__Plain)

	layout := qtlib.NewQVBoxLayout(v.widget.QWidget)
	layout.SetContentsMargins(0, 0, 0, 0)

	v.graph = newScoreGraph(clock)
	layout.AddWidget3(v.graph.widget, 1, 0)

	v.dockableView = newDockableView(parent, v.widget.QWidget, "Score Graph", "scoreGraphDock")

	return v
}

func (v *scoreGraphView) RepaintForThemeChange() {
	v.dockableView.RepaintForThemeChange()
	v.widget.SetPalette(qtlib.QGuiApplication_Palette())
	v.widget.Update()
}

// Implements core.score.View
func (v *scoreGraphView) ShowScore(score core.Score) {
	v.score = score
	v.graph.SetGraphs(score.StackedGraphPerBand())
	v.graph.widget.Update()
}

func (v *scoreGraphView) SetGoals(points, multis int) {
	v.graph.SetGoals(points, multis)
}

func (v *scoreGraphView) SetQTCsEnabled(enabled bool) {
	// Graph doesn't need QTC info - no-op
}

// Implements core.rate.RateUpdatedListener
func (v *scoreGraphView) RateUpdated(_ core.QSORate) {
	v.graph.UpdateTimeFrame()
	v.graph.widget.Update()
}
