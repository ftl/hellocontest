package ui

import (
	"fmt"
	"log"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
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

type scoreTable struct {
	colors colorProvider

	table        *gtk.TreeView
	tableContent *gtk.ListStore

	score core.Score
}

func newScoreTable(colors colorProvider) *scoreTable {
	result := &scoreTable{
		colors: colors,
	}

	result.tableContent = createScoreListStore(scoreColumnCount)

	result.table, _ = gtk.TreeViewNew()
	result.table.SetHExpand(true)
	result.table.SetVExpand(false)
	result.table.SetHAlign(gtk.ALIGN_CENTER)
	result.table.SetVAlign(gtk.ALIGN_FILL)
	result.table.SetCanFocus(false)
	result.table.SetModel(result.tableContent)
	result.table.AppendColumn(createScoreBandColumn("Band", scoreColumnBand, colors))
	result.table.AppendColumn(createScoreColumn("QSOs", scoreColumnQSOs))
	result.table.AppendColumn(createScoreColumn("Dupes", scoreColumnDupes))
	result.table.AppendColumn(createScoreColumn("Points", scoreColumnPoints))
	result.table.AppendColumn(createScoreColumn("P/Q", scoreColumnPointsPerQSOs))
	result.table.AppendColumn(createScoreColumn("Mult", scoreColumnMultis))
	result.table.AppendColumn(createScoreColumn("Q/M", scoreColumnQSOsPerMulti))
	result.table.AppendColumn(createScoreColumn("Result", scoreColumnResult))
	result.table.Connect("style-updated", result.refreshTableStyle)

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

func createScoreBandColumn(title string, id int, colors colorProvider) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatalf("Cannot create cell renderer for band column: %v", err)
	}
	cellRenderer.SetProperty("xalign", 1.0) // align text to the right

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "markup", id)
	if err != nil {
		log.Fatalf("Cannot create column %s: %v", title, err)
	}

	if colors != nil {
		cellRenderer.SetProperty("foreground-set", true)
		cellRenderer.SetProperty("background-set", true)
		column.AddAttribute(cellRenderer, "foreground", scoreColumnForeground)
		column.AddAttribute(cellRenderer, "background", scoreColumnBackground)
	}

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

func (t *scoreTable) Table() *gtk.TreeView {
	return t.table
}

func (t *scoreTable) ShowScore(score core.Score) {
	t.score = score
}

func (t *scoreTable) showScoreInTable(score core.Score) {
	t.tableContent.Clear()
	for _, band := range core.Bands {
		bandScore, ok := score.ScorePerBand[band]
		if !ok {
			continue
		}
		row := t.tableContent.Append()
		err := t.fillBandScoreToTableRow(row, band, bandScore)
		if err != nil {
			log.Printf("Cannot add entry to band score for band %s: %v", band, err)
		}
	}
	row := t.tableContent.Append()
	err := t.fillBandScoreToTableRow(row, totalBandName, score.Result())
	if err != nil {
		log.Printf("Cannot add entry to band score for total score: %v", err)
	}
}
func (t *scoreTable) fillBandScoreToTableRow(row *gtk.TreeIter, band core.Band, score core.BandScore) error {
	styler := func(s string) string {
		result := s
		if band == totalBandName {
			result = fmt.Sprintf("<b>%s</b>", result)
		}
		return result
	}

	columns := []int{
		scoreColumnBand,
		scoreColumnQSOs,
		scoreColumnDupes,
		scoreColumnPoints,
		scoreColumnPointsPerQSOs,
		scoreColumnMultis,
		scoreColumnQSOsPerMulti,
		scoreColumnResult,
	}

	values := []any{
		styler(string(band)),
		fmt.Sprintf(styler("%d"), score.QSOs),
		fmt.Sprintf(styler("%d"), score.Duplicates),
		fmt.Sprintf(styler("%d"), score.Points),
		fmt.Sprintf(styler("%4.1f"), score.PointsPerQSO()),
		fmt.Sprintf(styler("%d"), score.Multis),
		fmt.Sprintf(styler("%4.1f"), score.QSOsPerMulti()),
		fmt.Sprintf(styler("%d"), score.Result()),
	}

	if t.colors != nil {
		columns = append(columns, scoreColumnForeground, scoreColumnBackground)
		values = append(values, bandColor(t.colors, band).ToWeb(), bandBackgroundColor(t.colors))
	}

	return t.tableContent.Set(row, columns, values)
}

func (t *scoreTable) refreshTableStyle() {
	t.showScoreInTable(t.score)
}
