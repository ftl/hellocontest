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

func setupNewScoreView(builder *gtk.Builder) *scoreView {
	result := &scoreView{
		graph: newScoreGraph(),
	}

	result.tableLabel = getUI(builder, "tableLabel").(*gtk.Label)
	result.graphArea = getUI(builder, "scoreGraphArea").(*gtk.DrawingArea)

	result.graphArea.Connect("draw", result.graph.Draw)

	return result
}

func (v *scoreView) ShowScore(score core.Score) {
	if v == nil {
		return
	}
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
	if v == nil {
		return
	}
	v.graph.SetGoals(points, multis)
}

func (v *scoreView) RateUpdated(rate core.QSORate) {
	if v == nil {
		return
	}
	v.graph.UpdateTimeFrame()

	if v.graphArea != nil {
		v.graphArea.QueueDraw()
	}
}
