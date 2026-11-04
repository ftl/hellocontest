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
	headers  []string
	delegate *qtlib.QStyledItemDelegate

	controller   BandMatrixController
	focuser      EntryFocuser
	currentFrame core.BandMatrixFrame
}

func newBandMatrix(focuser EntryFocuser) *bandMatrix {
	t := &bandMatrix{focuser: focuser}
	t.model = qtlib.NewQStandardItemModel2(0, matrixColCount)
	t.headers = []string{"Band  ", "VFO1", "VFO2", "S   ", "P   ", "M   "}
	t.model.SetHorizontalHeaderLabels(t.headers)

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

	for i, header := range t.headers {
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
		if index.DataWithRole(roleMaxValue).ToBool() {
			// the style paints a selected item with the selection background and the matching
			// text color, State_Active selects the colors of an active window
			option.SetState(option.State() | qtlib.QStyle__State_Selected | qtlib.QStyle__State_Active)
		}
		super(painter, option, index)
	})
	t.widget.SetItemDelegate(t.delegate.QAbstractItemDelegate)
}

func (t *bandMatrix) installFocusedVFOHeader() {
	t.header.OnPaintSection(func(super func(painter *qtlib.QPainter, rect *qtlib.QRect, logicalIndex int), painter *qtlib.QPainter, rect *qtlib.QRect, logicalIndex int) {
		if logicalIndex != t.columnOfVFO(t.currentFrame.FocusedVFO) {
			super(painter, rect, logicalIndex)
			return
		}
		t.drawSelectedSection(painter, rect, logicalIndex)
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
		default:
			return
		}
		t.returnFocus()
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
		default:
			return
		}
		t.returnFocus()
	})
}

func (t *bandMatrix) returnFocus() {
	// the user works in the central area, the band matrix is only a short excursion
	if t.focuser == nil {
		return
	}
	t.focuser.GotoEntryFields()
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
			item.SetForeground(gc(qtlib.NewQBrush3(bandQColor(summary.Band))))
		case matrixColVFO1, matrixColVFO2:
			item.SetTextAlignment(qtlib.AlignCenter)
		default:
			item.SetTextAlignment(qtlib.AlignRight | qtlib.AlignVCenter)
		}
	}

	row := t.model.RowCount(gc(qtlib.NewQModelIndex()))
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
	item.SetData(gc(qtlib.NewQVariant8(true)), roleMaxValue)
}

func (t *bandMatrix) drawSelectedSection(painter *qtlib.QPainter, rect *qtlib.QRect, logicalIndex int) {
	// QHeaderView offers no style option to mark a section as selected, therefore the
	// section is painted here completely instead of calling the default implementation
	if logicalIndex < 0 || logicalIndex >= len(t.headers) {
		return
	}

	painter.Save()
	defer painter.Restore()

	palette := qtlib.QGuiApplication_Palette()
	painter.FillRect3(rect, palette.Highlight())
	painter.SetPen(palette.HighlightedText().Color())

	margin := t.widget.Style().PixelMetric(qtlib.QStyle__PM_HeaderMargin, nil, t.header.QWidget)
	textRect := gc(qtlib.NewQRect4(rect.X()+margin, rect.Y(), rect.Width()-2*margin, rect.Height()))
	painter.DrawText6(textRect, int(qtlib.AlignLeft|qtlib.AlignVCenter), t.headers[logicalIndex])
}
