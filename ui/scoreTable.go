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
	scoreColumnQTCs
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
	tableColumns []*gtk.TreeViewColumn

	score       core.Score
	qtcsEnabled bool
}

func newScoreTable(colors colorProvider) *scoreTable {
	result := &scoreTable{
		colors: colors,
	}

	result.tableContent = createScoreListStore(scoreColumnCount)

	result.tableColumns = []*gtk.TreeViewColumn{
		createScoreBandColumn("Band", scoreColumnBand, colors),
		createScoreColumn("QSOs", scoreColumnQSOs),
		createScoreColumn("QTCs", scoreColumnQTCs),
		createScoreColumn("Dupes", scoreColumnDupes),
		createScoreColumn("Points", scoreColumnPoints),
		createScoreColumn("P/Q", scoreColumnPointsPerQSOs),
		createScoreColumn("Mult", scoreColumnMultis),
		createScoreColumn("Q/M", scoreColumnQSOsPerMulti),
		createScoreColumn("Result", scoreColumnResult),
	}

	result.table, _ = gtk.TreeViewNew()
	result.table.SetHExpand(true)
	result.table.SetVExpand(false)
	result.table.SetHAlign(gtk.ALIGN_CENTER)
	result.table.SetVAlign(gtk.ALIGN_FILL)
	result.table.SetCanFocus(false)
	result.table.SetModel(result.tableContent)
	result.table.AppendColumn(result.tableColumns[scoreColumnBand])
	result.table.AppendColumn(result.tableColumns[scoreColumnQSOs])
	result.table.AppendColumn(result.tableColumns[scoreColumnDupes])
	result.table.AppendColumn(result.tableColumns[scoreColumnPoints])
	result.table.AppendColumn(result.tableColumns[scoreColumnPointsPerQSOs])
	result.table.AppendColumn(result.tableColumns[scoreColumnMultis])
	result.table.AppendColumn(result.tableColumns[scoreColumnQSOsPerMulti])
	result.table.AppendColumn(result.tableColumns[scoreColumnResult])
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

func (t *scoreTable) SetQTCsEnabled(enabled bool) {
	if t.qtcsEnabled == enabled {
		return
	}
	t.qtcsEnabled = enabled

	if enabled {
		t.table.InsertColumn(t.tableColumns[scoreColumnQTCs], scoreColumnQTCs)
	} else {
		t.table.RemoveColumn(t.tableColumns[scoreColumnQTCs])
	}
}

func (t *scoreTable) ShowScore(score core.Score) {
	t.score = score
	t.showScoreInTable(score)
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
		scoreColumnQTCs,
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
		fmt.Sprintf(styler("%d"), score.QTCs),
		fmt.Sprintf(styler("%d"), score.Duplicates),
		fmt.Sprintf(styler("%d"), score.Points),
		fmt.Sprintf(styler("%4.1f"), score.PointsPerQSO()),
		fmt.Sprintf(styler("%d"), score.Multis),
		fmt.Sprintf(styler("%4.1f"), score.QSOsPerMulti()),
		fmt.Sprintf(styler("%d"), score.Result()),
	}

	if t.colors != nil {
		columns = append(columns, scoreColumnForeground, scoreColumnBackground)
		values = append(values, bandColor(t.colors, band).ToWeb(), bandBackgroundColor(t.colors).ToWeb())
	}

	return t.tableContent.Set(row, columns, values)
}

func (t *scoreTable) refreshTableStyle() {
	t.showScoreInTable(t.score)
}
