package ui

import (
	"fmt"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

const (
	matrixColBand = iota
	matrixColVFO1
	matrixColVFO2
	matrixColSpots
	matrixColPoints
	matrixColMultis

	matrixColCount
)

const (
	roleMaxValue = int(qtlib.UserRole) + 1

	vfoMarker = "●"

	markerRadius = 5.0
)

type BandMatrixController interface {
	SelectBand(core.VFOID, core.Band)
	ActivateBand(core.Band)
	FocusVFO(core.VFOID)
}

type bandMatrix struct {
	widget   *qtlib.QTableView
	model    *qtlib.QStandardItemModel
	header   *qtlib.QHeaderView
	delegate *qtlib.QStyledItemDelegate

	controller   BandMatrixController
	currentFrame core.BandMatrixFrame
}

func newBandMatrix() *bandMatrix {
	t := &bandMatrix{}
	t.model = qtlib.NewQStandardItemModel2(0, matrixColCount)
	headers := []string{"Band  ", "VFO1", "VFO2", "S   ", "P   ", "M   "}
	t.model.SetHorizontalHeaderLabels(headers)

	t.widget = qtlib.NewQTableView2()
	t.widget.SetModel(t.model.QAbstractItemModel)
	t.widget.SetEditTriggers(qtlib.QAbstractItemView__NoEditTriggers)
	t.widget.SetSelectionMode(qtlib.QAbstractItemView__NoSelection)
	t.widget.SetFocusPolicy(qtlib.NoFocus)
	t.widget.VerticalHeader().SetVisible(false)

	// miqt can only override the virtual methods of a directly constructed type,
	// therefore the horizontal header is created here instead of using the one that
	// the table view creates itself
	t.header = qtlib.NewQHeaderView2(qtlib.Horizontal, t.widget.QAbstractScrollArea.QFrame.QWidget)
	t.widget.SetHorizontalHeader(t.header)
	t.header.SetStretchLastSection(false)
	t.header.SetHighlightSections(false)
	t.header.SetDefaultAlignment(qtlib.AlignLeft | qtlib.AlignVCenter)
	t.header.SetSectionsClickable(true)

	for i, header := range headers {
		SetColumnSampleWidth(t.widget, i, header)
	}

	t.installMaxValueDelegate()
	t.installFocusedVFOHeader()
	t.connectCellClicks()
	t.connectHeaderClicks()

	return t
}

func (t *bandMatrix) SetController(controller BandMatrixController) {
	t.controller = controller
}

func (t *bandMatrix) installMaxValueDelegate() {
	t.delegate = qtlib.NewQStyledItemDelegate()
	t.delegate.OnPaint(func(super func(painter *qtlib.QPainter, option *qtlib.QStyleOptionViewItem, index *qtlib.QModelIndex), painter *qtlib.QPainter, option *qtlib.QStyleOptionViewItem, index *qtlib.QModelIndex) {
		super(painter, option, index)
		if !index.DataWithRole(roleMaxValue).ToBool() {
			return
		}
		drawSelectionMarker(painter, option.Rect(), option.Palette())
	})
	t.widget.SetItemDelegate(t.delegate.QAbstractItemDelegate)
}

func (t *bandMatrix) installFocusedVFOHeader() {
	t.header.OnPaintSection(func(super func(painter *qtlib.QPainter, rect *qtlib.QRect, logicalIndex int), painter *qtlib.QPainter, rect *qtlib.QRect, logicalIndex int) {
		super(painter, rect, logicalIndex)
		if logicalIndex != t.columnOfVFO(t.currentFrame.FocusedVFO) {
			return
		}
		drawSelectionMarker(painter, rect, qtlib.QGuiApplication_Palette())
	})
}

func (t *bandMatrix) connectCellClicks() {
	t.widget.OnClicked(func(index *qtlib.QModelIndex) {
		if t.controller == nil {
			return
		}
		band, ok := t.bandOfRow(index.Row())
		if !ok {
			return
		}
		switch index.Column() {
		case matrixColVFO1:
			t.controller.SelectBand(core.VFO1, band)
		case matrixColVFO2:
			t.controller.SelectBand(core.VFO2, band)
		case matrixColSpots, matrixColPoints, matrixColMultis:
			t.controller.ActivateBand(band)
		}
	})
}

func (t *bandMatrix) connectHeaderClicks() {
	t.header.OnSectionClicked(func(logicalIndex int) {
		if t.controller == nil {
			return
		}
		switch logicalIndex {
		case matrixColVFO1:
			t.controller.FocusVFO(core.VFO1)
		case matrixColVFO2:
			t.controller.FocusVFO(core.VFO2)
		}
	})
}

func (t *bandMatrix) ShowFrame(frame core.BandMatrixFrame) {
	t.currentFrame = frame
	t.widget.SetColumnHidden(matrixColVFO2, !frame.VFO2Available)

	ClearTableRows(t.model)
	for _, summary := range frame.Bands {
		t.appendRow(summary, frame)
	}

	t.header.Viewport().Update()
}

func (t *bandMatrix) appendRow(summary core.BandSummary, frame core.BandMatrixFrame) {
	cells := make([]*qtlib.QStandardItem, matrixColCount)
	cells[matrixColBand] = qtlib.NewQStandardItem2(string(summary.Band))
	cells[matrixColVFO1] = qtlib.NewQStandardItem2(markerFor(summary.Band, frame.VFOBands[core.VFO1]))
	cells[matrixColVFO2] = qtlib.NewQStandardItem2(markerFor(summary.Band, frame.VFOBands[core.VFO2]))
	cells[matrixColSpots] = qtlib.NewQStandardItem2(fmt.Sprintf("%d", summary.SpotCount))
	cells[matrixColPoints] = qtlib.NewQStandardItem2(fmt.Sprintf("%d", summary.Points))
	cells[matrixColMultis] = qtlib.NewQStandardItem2(fmt.Sprintf("%d", summary.Multis()))

	markMaxValue(cells[matrixColSpots], summary.MaxSpots)
	markMaxValue(cells[matrixColPoints], summary.MaxPoints)
	markMaxValue(cells[matrixColMultis], summary.MaxMultis)

	for column, item := range cells {
		switch column {
		case matrixColBand:
			item.SetTextAlignment(qtlib.AlignLeft | qtlib.AlignVCenter)
			item.SetForeground(qtlib.NewQBrush3(bandQColor(summary.Band)))
			item.SetBackground(qtlib.NewQBrush3(bandBGQColor()))
		case matrixColVFO1, matrixColVFO2:
			item.SetTextAlignment(qtlib.AlignCenter)
		default:
			item.SetTextAlignment(qtlib.AlignRight | qtlib.AlignVCenter)
		}
	}

	row := t.model.RowCount(qtlib.NewQModelIndex())
	for column, item := range cells {
		t.model.SetItem(row, column, item)
	}
}

func (t *bandMatrix) bandOfRow(row int) (core.Band, bool) {
	if row < 0 || row >= len(t.currentFrame.Bands) {
		return core.NoBand, false
	}
	return t.currentFrame.Bands[row].Band, true
}

func (t *bandMatrix) columnOfVFO(vfo core.VFOID) int {
	if vfo == core.VFO2 {
		return matrixColVFO2
	}
	return matrixColVFO1
}

func markerFor(band, vfoBand core.Band) string {
	if band != core.NoBand && band == vfoBand {
		return vfoMarker
	}
	return ""
}

func markMaxValue(item *qtlib.QStandardItem, max bool) {
	if !max {
		return
	}
	item.SetData(qtlib.NewQVariant8(true), roleMaxValue)
}

func drawSelectionMarker(painter *qtlib.QPainter, rect *qtlib.QRect, palette *qtlib.QPalette) {
	painter.Save()
	defer painter.Restore()

	painter.SetRenderHint(qtlib.QPainter__Antialiasing)
	painter.SetPen(palette.Highlight().Color())
	painter.SetBrushWithStyle(qtlib.NoBrush)
	painter.DrawRoundedRect2(rect.X()+1, rect.Y()+1, rect.Width()-3, rect.Height()-3, markerRadius, markerRadius)
}
