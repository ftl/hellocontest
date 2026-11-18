package ui

import (
	"slices"

	"github.com/ftl/conval"
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core/model"
)

const colBand = 0

type scoreTable struct {
	widget *qtlib.QTableView
	model  *qtlib.QStandardItemModel

	qtcsEnabled     bool
	detailedMultis  bool
	multiProperties []conval.Property
	columns         []model.ScoreColumn
	headers         []string

	bold         *qtlib.QFont
	contestBands []model.Band
	score        model.Score
}

func newScoreTable() *scoreTable {
	t := &scoreTable{}
	t.model = qtlib.NewQStandardItemModel2(0, 0)

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

	t.bold = gc(qtlib.NewQFont())
	t.bold.SetBold(true)

	t.applyColumns()

	return t
}

func (t *scoreTable) applyColumns() {
	// model.ScoreColumns decides which columns there are, so that the table and the
	// exported summary show the same thing
	columns := model.ScoreColumns(t.qtcsEnabled, t.detailedMultis, t.multiProperties)
	headers := make([]string, len(columns)+1)
	headers[0] = "Band  "
	for i, column := range columns {
		headers[i+1] = column.Header
	}
	if slices.Equal(headers, t.headers) {
		return
	}

	t.columns = columns
	t.headers = headers

	t.model.SetColumnCount(len(headers))
	t.model.SetHorizontalHeaderLabels(headers)
	for i, header := range headers {
		SetColumnSampleWidth(t.widget, i, header)
	}
}

func (t *scoreTable) SetBands(bands []model.Band) {
	t.contestBands = bands
}

func (t *scoreTable) SetQTCsEnabled(enabled bool) {
	t.qtcsEnabled = enabled
	t.applyColumns()
	t.showRows()
}

func (t *scoreTable) SetDetailedMultis(detailed bool) {
	t.detailedMultis = detailed
	t.applyColumns()
	t.showRows()
}

func (t *scoreTable) ShowScore(score model.Score) {
	t.score = score
	// the properties come with the score, they depend on the contest and on the station
	t.multiProperties = score.MultiProperties
	t.applyColumns()
	t.showRows()
}

func (t *scoreTable) showRows() {
	// the rows must follow a change of the columns, therefore the table renders the last
	// score again
	ClearTableRows(t.model)

	for _, band := range t.bandsToShow(t.score) {
		t.appendRow(string(band), t.score.ScorePerBand[band], false)
	}
	t.appendRow("Total", t.score.Result(), true)
}

func (t *scoreTable) bandsToShow(score model.Score) []model.Band {
	contestBands := make(map[model.Band]bool, len(t.contestBands))
	for _, band := range t.contestBands {
		contestBands[band] = true
	}

	result := make([]model.Band, 0, len(model.Bands))
	for _, band := range model.Bands {
		_, scored := score.ScorePerBand[band]
		if contestBands[band] || scored {
			result = append(result, band)
		}
	}
	return result
}

func (t *scoreTable) appendRow(bandLabel string, bs model.BandScore, total bool) {
	cells := make([]*qtlib.QStandardItem, len(t.columns)+1)
	cells[colBand] = qtlib.NewQStandardItem2(bandLabel)
	for i, column := range t.columns {
		cells[i+1] = qtlib.NewQStandardItem2(column.Value(bs))
	}
	alignment := qtlib.AlignRight | qtlib.AlignVCenter
	for i, item := range cells {
		item.SetTextAlignment(alignment)
		if total {
			item.SetFont(t.bold)
			continue
		}
		if i == colBand {
			item.SetForeground(gc(qtlib.NewQBrush3(bandQColor(model.Band(bandLabel)))))
		}
	}
	rowIdx := t.model.RowCount(gc(qtlib.NewQModelIndex()))
	for col, item := range cells {
		t.model.SetItem(rowIdx, col, item)
	}
}
