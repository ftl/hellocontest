package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

type scoreView struct {
	tableLabel *gtk.Label
	graphArea  *gtk.DrawingArea

	graph *scoreGraph
}

func setupNewScoreView(builder *gtk.Builder, colors colorProvider) *scoreView {
	result := &scoreView{}

	result.tableLabel = getUI(builder, "tableLabel").(*gtk.Label)
	style.AddClass(result.tableLabel.ToWidget(), "score_table")

	result.graph = newScoreGraph(colors)
	result.graphArea = getUI(builder, "scoreGraphArea").(*gtk.DrawingArea)
	result.graphArea.Connect("draw", result.graph.Draw)
	result.graphArea.Connect("style-updated", result.graph.RefreshStyle)

	return result
}

func (v *scoreView) ShowScore(score core.Score) {
	v.graph.SetGraphs(score.StackedGraphPerBand())

	renderedScore := fmt.Sprintf("<span allow_breaks='true' font_family='monospace'>%s</span>", score)
	v.tableLabel.SetMarkup(renderedScore)

	v.graphArea.QueueDraw()
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
