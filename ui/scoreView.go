package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

type scoreView struct {
	rootGrid   *gtk.Grid
	tableLabel *gtk.Label
	graphArea  *gtk.DrawingArea

	graph *scoreGraph
}

func setupNewScoreView(parent gtk.IWidget, colors colorProvider, clock core.Clock) *scoreView {
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
	result.graphArea.SetHExpand(true)
	result.graphArea.SetVExpand(true)
	result.graphArea.SetHAlign(gtk.ALIGN_FILL)
	result.graphArea.SetVAlign(gtk.ALIGN_FILL)
	result.graphArea.SetCanFocus(false)
	result.graphArea.Connect("draw", result.graph.Draw)
	result.graphArea.Connect("style-updated", result.graph.RefreshStyle)

	result.tableLabel, _ = gtk.LabelNew("")
	result.tableLabel.SetHExpand(true)
	result.tableLabel.SetVExpand(false)
	result.tableLabel.SetHAlign(gtk.ALIGN_FILL)
	result.tableLabel.SetVAlign(gtk.ALIGN_FILL)
	result.tableLabel.SetTrackVisitedLinks(false)
	result.tableLabel.SetCanFocus(false)
	style.AddClass(result.tableLabel.ToWidget(), "score_table")

	result.rootGrid.Attach(result.graphArea, 0, 0, 1, 1)
	result.rootGrid.Attach(result.tableLabel, 0, 1, 1, 1)

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
