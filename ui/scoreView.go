package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type scoreView struct {
	rootGrid  *gtk.Grid
	graphArea *gtk.DrawingArea
	table     *scoreTable

	graph *scoreGraph

	score core.Score
}

func setupNewScoreView(colors colorProvider, clock core.Clock) *scoreView {
	result := &scoreView{}

	result.rootGrid, _ = gtk.GridNew()
	result.rootGrid.InsertColumn(0)
	result.rootGrid.InsertRow(0)
	result.rootGrid.InsertRow(0)
	result.rootGrid.SetColumnSpacing(5)
	result.rootGrid.SetRowSpacing(5)
	result.rootGrid.SetCanFocus(false)

	result.graph = newScoreGraph(colors, clock)
	result.graphArea, _ = gtk.DrawingAreaNew()
	result.graphArea.SetSizeRequest(400, 250)
	result.graphArea.SetHExpand(true)
	result.graphArea.SetVExpand(false)
	result.graphArea.SetHAlign(gtk.ALIGN_FILL)
	result.graphArea.SetVAlign(gtk.ALIGN_FILL)
	result.graphArea.SetCanFocus(false)
	result.graphArea.Connect("draw", result.graph.Draw)
	result.graphArea.Connect("style-updated", result.graph.RefreshStyle)

	result.table = newScoreTable(colors)
	result.table.Table().SetVAlign(gtk.ALIGN_START)

	result.rootGrid.Attach(result.graphArea, 0, 0, 1, 1)
	result.rootGrid.Attach(result.table.Table(), 0, 1, 1, 1)

	return result
}

func (v *scoreView) ShowScore(score core.Score) {
	v.score = score

	v.graph.SetGraphs(score.StackedGraphPerBand())
	v.table.ShowScore(score)

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
