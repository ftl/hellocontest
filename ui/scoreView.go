package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type scoreView struct {
	tableLabel *gtk.Label
	graphArea  *gtk.DrawingArea

	graph *scoreGraph
}

func newScoreView() *scoreView {
	return &scoreView{
		graph: newScoreGraph(),
	}
}

func (v *scoreView) setup(builder *gtk.Builder) {
	v.tableLabel = getUI(builder, "tableLabel").(*gtk.Label)
	v.graphArea = getUI(builder, "scoreGraphArea").(*gtk.DrawingArea)

	v.graphArea.Connect("draw", v.graph.Draw)
}

func (v *scoreView) ShowScore(score core.Score) {
	v.graph.SetGraph(score.GraphPerBand[core.NoBand])

	if v.tableLabel != nil {
		renderedScore := fmt.Sprintf("<span allow_breaks='true' font_family='monospace'>%s</span>", score)
		v.tableLabel.SetMarkup(renderedScore)
	}

	if v.graphArea != nil {
		v.graphArea.QueueDraw()
	}
}

func (v *scoreView) SetGoals(points int, multis int) {
	v.graph.SetGoals(points, multis)
}

func (v *scoreView) RateUpdated(rate core.QSORate) {
	v.graph.UpdateTimeFrame()

	if v.graphArea != nil {
		v.graphArea.QueueDraw()
	}
}
