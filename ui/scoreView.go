package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

const (
	scoreColumnBand int = iota
	scoreColumnQSOs
	scoreColumnDupes
	scoreColumnPoints
	scoreColumnPointsPerQSOs
	scoreColumnMultis
	scoreColumnQSOsPerMulti
	scoreColumnResult

	scoreColumnForeground
	scoreColumnBackground

	scoreColumnCount
)

const totalBandName = "Total"

type scoreView struct {
	colors colorProvider

	rootGrid     *gtk.Grid
	graphArea    *gtk.DrawingArea
	table        *gtk.TreeView
	tableContent *gtk.ListStore

	graph *scoreGraph

	score core.Score
}

func setupNewScoreView(colors colorProvider, clock core.Clock) *scoreView {
	result := &scoreView{
		colors: colors,
	}

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

	result.tableContent = createScoreListStore(scoreColumnCount)

	result.table, _ = gtk.TreeViewNew()
	result.table.SetHExpand(true)
	result.table.SetVExpand(false)
	result.table.SetHAlign(gtk.ALIGN_CENTER)
	result.table.SetVAlign(gtk.ALIGN_FILL)
	result.table.SetCanFocus(false)
	result.table.SetModel(result.tableContent)
	result.table.AppendColumn(createScoreBandColumn("Band", scoreColumnBand))
	result.table.AppendColumn(createScoreColumn("QSOs", scoreColumnQSOs))
	result.table.AppendColumn(createScoreColumn("Dupes", scoreColumnDupes))
	result.table.AppendColumn(createScoreColumn("Points", scoreColumnPoints))
	result.table.AppendColumn(createScoreColumn("P/Q", scoreColumnPointsPerQSOs))
	result.table.AppendColumn(createScoreColumn("Mult", scoreColumnMultis))
	result.table.AppendColumn(createScoreColumn("Q/M", scoreColumnQSOsPerMulti))
	result.table.AppendColumn(createScoreColumn("Result", scoreColumnResult))
	result.table.Connect("style-updated", result.refreshTableStyle)

	result.rootGrid.Attach(result.graphArea, 0, 0, 1, 1)
	result.rootGrid.Attach(result.table, 0, 1, 1, 1)

	return result
}

func createScoreListStore(columnCount int) *gtk.ListStore {
	types := make([]glib.Type, columnCount)
	for i := range types {
		types[i] = glib.TYPE_STRING
	}
	result, err := gtk.ListStoreNew(types...)
	if err != nil {
		log.Fatalf("Cannot create list store: %v", err)
	}
	return result
}

func createScoreBandColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatalf("Cannot create cell renderer for band column: %v", err)
	}
	cellRenderer.SetProperty("xalign", 1.0) // align text to the right
	cellRenderer.SetProperty("foreground-set", true)
	cellRenderer.SetProperty("background-set", true)

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "markup", id)
	if err != nil {
		log.Fatalf("Cannot create column %s: %v", title, err)
	}
	column.AddAttribute(cellRenderer, "foreground", scoreColumnForeground)
	column.AddAttribute(cellRenderer, "background", scoreColumnBackground)

	return column
}

func createScoreColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatalf("Cannot create cell renderer for column %s: %v", title, err)
	}
	cellRenderer.SetProperty("xalign", 1.0) // align text to the right

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "markup", id)
	if err != nil {
		log.Fatalf("Cannot create column %s: %v", title, err)
	}

	return column
}

func (v *scoreView) fillBandScoreToTableRow(row *gtk.TreeIter, band core.Band, score core.BandScore) error {
	styler := func(s string) string {
		result := s
		if band == totalBandName {
			result = fmt.Sprintf("<b>%s</b>", result)
		}
		return result
	}

	return v.tableContent.Set(row,
		[]int{
			scoreColumnBand,
			scoreColumnQSOs,
			scoreColumnDupes,
			scoreColumnPoints,
			scoreColumnPointsPerQSOs,
			scoreColumnMultis,
			scoreColumnQSOsPerMulti,
			scoreColumnResult,
			scoreColumnForeground,
			scoreColumnBackground,
		},
		[]any{
			styler(string(band)),
			fmt.Sprintf(styler("%d"), score.QSOs),
			fmt.Sprintf(styler("%d"), score.Duplicates),
			fmt.Sprintf(styler("%d"), score.Points),
			fmt.Sprintf(styler("%4.1f"), score.PointsPerQSO()),
			fmt.Sprintf(styler("%d"), score.Multis),
			fmt.Sprintf(styler("%4.1f"), score.QSOsPerMulti()),
			fmt.Sprintf(styler("%d"), score.Result()),
			bandColor(v.colors, band).ToWeb(),
			bandBackgroundColor(v.colors),
		},
	)
}

func bandBackgroundColor(colors colorProvider) string {
	if !colors.HasColor("hellocontest-graph-bg") {
		return colors.BackgroundColor().ToWeb()
	}
	return colors.ColorByName("hellocontest-graph-bg").ToWeb()
}

func (v *scoreView) refreshTableStyle() {
	v.showScoreInTable(v.score)
}

func (v *scoreView) ShowScore(score core.Score) {
	v.score = score

	v.graph.SetGraphs(score.StackedGraphPerBand())
	v.showScoreInTable(score)

	v.graphArea.QueueDraw()
}

func (v *scoreView) showScoreInTable(score core.Score) {
	v.tableContent.Clear()
	for _, band := range core.Bands {
		bandScore, ok := score.ScorePerBand[band]
		if !ok {
			continue
		}
		row := v.tableContent.Append()
		err := v.fillBandScoreToTableRow(row, band, bandScore)
		if err != nil {
			log.Printf("Cannot add entry to band score for band %s: %v", band, err)
		}
	}
	row := v.tableContent.Append()
	err := v.fillBandScoreToTableRow(row, totalBandName, score.Result())
	if err != nil {
		log.Printf("Cannot add entry to band score for total score: %v", err)
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
