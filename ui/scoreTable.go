package ui

import (
	"fmt"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

const (
	colBand = iota
	colQSOs
	colQTCs
	colDupes
	colPoints
	colPointsPerQSO
	colMultis
	colQSOsPerMulti
	colResult

	colCount
)

type scoreTable struct {
	widget *qtlib.QTableView
	model  *qtlib.QStandardItemModel

	qtcsEnabled bool
	bold        *qtlib.QFont
}

func newScoreTable() *scoreTable {
	t := &scoreTable{}
	t.model = qtlib.NewQStandardItemModel2(0, colCount)
	headers := []string{"Band  ", "QSOs", "QTCs", "Dupes", "Points", "P/Q  ", "Mult", "Q/M", "Result"}
	t.model.SetHorizontalHeaderLabels(headers)

	t.widget = qtlib.NewQTableView2()
	t.widget.SetModel(t.model.QAbstractItemModel)

	// configure the score table individually instead of using ConfigureReadOnlyTable
	t.widget.SetEditTriggers(qtlib.QAbstractItemView__NoEditTriggers)
	t.widget.SetSelectionMode(qtlib.QAbstractItemView__NoSelection)
	t.widget.SetFocusPolicy(qtlib.NoFocus)
	t.widget.VerticalHeader().SetVisible(false)
	t.widget.HorizontalHeader().SetStretchLastSection(false)
	t.widget.HorizontalHeader().SetHighlightSections(false)
	t.widget.HorizontalHeader().SetDefaultAlignment(qtlib.AlignLeft | qtlib.AlignVCenter)

	for i, header := range headers {
		SetColumnSampleWidth(t.widget, i, header)
	}

	t.bold = qtlib.NewQFont()
	t.bold.SetBold(true)

	t.widget.SetColumnHidden(colQTCs, true)

	return t
}

func (t *scoreTable) SetQTCsEnabled(enabled bool) {
	t.qtcsEnabled = enabled
	t.widget.SetColumnHidden(colQTCs, !enabled)
}

func (t *scoreTable) ShowScore(score core.Score) {
	ClearTableRows(t.model)

	for _, band := range core.Bands {
		bs, ok := score.ScorePerBand[band]
		if !ok {
			continue
		}
		t.appendRow(string(band), bs, false)
	}
	t.appendRow("Total", score.Result(), true)
}

func (t *scoreTable) appendRow(bandLabel string, bs core.BandScore, total bool) {
	cells := []*qtlib.QStandardItem{
		qtlib.NewQStandardItem2(bandLabel),
		qtlib.NewQStandardItem2(fmt.Sprintf("%d", bs.QSOs)),
		qtlib.NewQStandardItem2(fmt.Sprintf("%d", bs.QTCs)),
		qtlib.NewQStandardItem2(fmt.Sprintf("%d", bs.Duplicates)),
		qtlib.NewQStandardItem2(fmt.Sprintf("%d", bs.Points)),
		qtlib.NewQStandardItem2(fmt.Sprintf("%4.1f", bs.PointsPerQSO())),
		qtlib.NewQStandardItem2(fmt.Sprintf("%d", bs.Multis)),
		qtlib.NewQStandardItem2(fmt.Sprintf("%4.1f", bs.QSOsPerMulti())),
		qtlib.NewQStandardItem2(fmt.Sprintf("%d", bs.Result())),
	}
	alignment := qtlib.AlignRight | qtlib.AlignVCenter
	for i, item := range cells {
		item.SetTextAlignment(alignment)
		if total {
			item.SetFont(t.bold)
			continue
		}
		if i == colBand {
			item.SetForeground(qtlib.NewQBrush3(bandQColor(core.Band(bandLabel))))
			item.SetBackground(qtlib.NewQBrush3(bandBGQColor()))
		}
	}
	rowIdx := t.model.RowCount(qtlib.NewQModelIndex())
	for col, item := range cells {
		t.model.SetItem(rowIdx, col, item)
	}
}
