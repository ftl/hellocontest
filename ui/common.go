package ui

import (
	"fmt"
	"strings"
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

func makeFilterableCombo() *qtlib.QComboBox {
	combo := qtlib.NewQComboBox2()
	combo.SetEditable(true)
	completer := qtlib.NewQCompleter2(combo.Model())
	completer.SetCompletionMode(qtlib.QCompleter__PopupCompletion)
	completer.SetFilterMode(qtlib.MatchContains)
	completer.SetCaseSensitivity(qtlib.CaseInsensitive)
	combo.SetCompleter(completer)
	return combo
}

func setupBandCombo(combo *qtlib.QComboBox) {
	combo.Clear()
	for _, band := range core.Bands {
		combo.AddItem(band.String())
	}
	combo.SetCurrentIndex(0)
}

func setupModeCombo(combo *qtlib.QComboBox) {
	combo.Clear()
	for _, mode := range core.Modes {
		combo.AddItem(mode.String())
	}
	combo.SetCurrentIndex(0)
}

func SetColumnSampleWidth(widget *qtlib.QTableView, column int, sample string) {
	width := widget.FontMetrics().HorizontalAdvance(sample) + 6
	widget.HorizontalHeader().SetSectionResizeMode2(column, qtlib.QHeaderView__Interactive)
	widget.SetColumnWidth(column, width)
}

func ClearTableRows(model *qtlib.QStandardItemModel) {
	rowCount := model.RowCount(qtlib.NewQModelIndex())
	if rowCount > 0 {
		model.RemoveRows(0, rowCount, qtlib.NewQModelIndex())
	}
}

func SelectTableRow(view *qtlib.QTableView, index int, model *qtlib.QStandardItemModel) {
	if index < 0 || index >= model.RowCount(qtlib.NewQModelIndex()) {
		return
	}
	view.SelectRow(index)
	modelIndex := model.Index(index, 0, qtlib.NewQModelIndex())
	view.ScrollTo(modelIndex, qtlib.QAbstractItemView__EnsureVisible)
}

func ConfigureReadOnlyTable(view *qtlib.QTableView) {
	view.SetEditTriggers(qtlib.QAbstractItemView__NoEditTriggers)
	view.SetSelectionMode(qtlib.QAbstractItemView__SingleSelection)
	view.SetSelectionBehavior(qtlib.QAbstractItemView__SelectRows)
	view.SetAlternatingRowColors(true)
	view.VerticalHeader().SetVisible(false)
	view.HorizontalHeader().SetHighlightSections(false)
	view.HorizontalHeader().SetDefaultAlignment(qtlib.AlignLeft | qtlib.AlignVCenter)
}

func PointsToString(points int, duplicate bool) string {
	if duplicate {
		return fmt.Sprintf("(%d)", points)
	}
	return fmt.Sprintf("%d", points)
}

func DuplicateToCheckmark(duplicate bool) string {
	if duplicate {
		return "✓"
	}
	return ""
}

func FormatQTCCount(sent, received int) string {
	switch {
	case sent > 0 && received > 0:
		return fmt.Sprintf("%dS %dR", sent, received)
	case sent > 0:
		return fmt.Sprintf("%dS", sent)
	case received > 0:
		return fmt.Sprintf("%dR", received)
	default:
		return "0"
	}
}

func FormatSpotAge(lastHeard time.Time) (string, bool) {
	s := time.Since(lastHeard).Truncate(time.Minute).String()
	if s == "0s" {
		return "< 1m", true
	}
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s, false
}

func FormatSpotMark(entry core.BandmapEntry, frame core.BandmapFrame) string {
	switch {
	case entry.ID == frame.HighestValueEntry.ID:
		return "H"
	case entry.ID == frame.SelectedEntry.ID:
		return ">"
	case entry.OnFrequency(frame.Frequency):
		return "|"
	case entry.ID == frame.NearestEntry.ID:
		return "N"
	default:
		return ""
	}
}

func FormatSpotFrequency(f core.Frequency) string {
	return fmt.Sprintf("%.2f kHz", f/1000)
}

func KindToString(kind core.QTCKind) string {
	switch kind {
	case core.ReceivedQTC:
		return "R"
	case core.SentQTC:
		return "S"
	default:
		return "?"
	}
}

type dockableView struct {
	dock *qtlib.QDockWidget
}

func newDockableView(parent, root *qtlib.QWidget, title, objectName string) *dockableView {
	dock := qtlib.NewQDockWidget4(title, parent)
	dock.SetWidget(root)
	dock.SetObjectName(*qtlib.NewQAnyStringView3(objectName))
	dock.SetAllowedAreas(qtlib.AllDockWidgetAreas)

	return &dockableView{
		dock: dock,
	}
}

func (v *dockableView) RepaintForThemeChange() {
	v.dock.SetPalette(qtlib.QGuiApplication_Palette())
	v.dock.Update()
}

func (v *dockableView) Dock() *qtlib.QDockWidget {
	return v.dock
}

func (v *dockableView) Show() {
	if v.dock == nil {
		return
	}
	v.dock.SetVisible(true)
	v.dock.Raise()
}

func (v *dockableView) Hide() {
	if v.dock == nil {
		return
	}
	v.dock.SetVisible(false)
}

func repaintScrollBarsForThemeChange(area *qtlib.QAbstractScrollArea) {
	style := qtlib.QApplication_Style()
	vBar := area.VerticalScrollBar()
	style.Unpolish(vBar.QWidget)
	style.Polish(vBar.QWidget)
	vBar.QWidget.Update()
	hBar := area.HorizontalScrollBar()
	style.Unpolish(hBar.QWidget)
	style.Polish(hBar.QWidget)
	hBar.QWidget.Update()
}
