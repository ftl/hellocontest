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

func setupNewScoreView(builder *gtk.Builder, style *style.Style) *scoreView {
	result := &scoreView{}

	result.tableLabel = getUI(builder, "tableLabel").(*gtk.Label)

	result.graphArea = getUI(builder, "scoreGraphArea").(*gtk.DrawingArea)
	result.graph = newScoreGraph(style.ForWidget(result.graphArea.ToWidget()))
	result.graphArea.Connect("draw", result.graph.Draw)
	result.graphArea.Connect("style-updated", result.graph.RefreshStyle)

	return result
}

func (v *scoreView) ShowScore(score core.Score) {
	if v == nil {
		return
	}
	v.graph.SetGraphs(score.StackedGraphPerBand())

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
