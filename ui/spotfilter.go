package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core/model"
)

type SpotFilterController interface {
	SetFilterBand(model.SpotFilterBand)
	SetFilterMode(model.SpotFilterMode)
	SetFilterValue(model.SpotFilterValue)
	SetFilterSort(model.SpotSortColumn, bool)
	SetFilterFolded(bool)
}

type spotFilterArea struct {
	widget  *qtlib.QWidget
	toggle  *qtlib.QToolButton
	summary *qtlib.QLabel
	body    *qtlib.QWidget

	bandCombo  *qtlib.QComboBox
	modeCombo  *qtlib.QComboBox
	valueCombo *qtlib.QComboBox
	sortCombo  *qtlib.QComboBox
	direction  *qtlib.QToolButton

	bandItems  []model.SpotFilterBand
	modeItems  []model.SpotFilterMode
	valueItems []model.SpotFilterValue
	sortItems  []model.SpotSortColumn

	controller SpotFilterController
	focuser    EntryFocuser
	updating   bool
}

func newSpotFilterArea(controller SpotFilterController, focuser EntryFocuser) *spotFilterArea {
	a := &spotFilterArea{controller: controller, focuser: focuser}

	a.widget = qtlib.NewQWidget2()
	setObjectName(a.widget, "spotFilter")
	layout := qtlib.NewQVBoxLayout(a.widget)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(2)

	layout.AddWidget(a.buildHeader())
	a.buildBody()
	layout.AddWidget(a.body)

	return a
}

func (a *spotFilterArea) buildHeader() *qtlib.QWidget {
	header := qtlib.NewQWidget2()
	layout := qtlib.NewQHBoxLayout(header)
	layout.SetContentsMargins(0, 0, 0, 0)

	a.toggle = qtlib.NewQToolButton2()
	setObjectName(a.toggle, "spotFilterToggle")
	a.toggle.SetFocusPolicy(qtlib.NoFocus)
	a.toggle.SetArrowType(qtlib.DownArrow)
	a.toggle.SetCheckable(true)
	a.toggle.SetAutoRaise(true)
	a.toggle.SetToolTip("Show or hide the filter settings")
	a.toggle.OnToggled(func(checked bool) {
		a.applyFolded(checked)
		if a.updating || a.controller == nil {
			return
		}
		a.controller.SetFilterFolded(checked)
	})
	layout.AddWidget(a.toggle.QWidget)

	a.summary = qtlib.NewQLabel2()
	layout.AddWidget3(a.summary.QFrame.QWidget, 1, 0)

	return header
}

func (a *spotFilterArea) buildBody() {
	// the items of the combo boxes arrive with the first frame, long after the first show.
	// Without the size adjust policy a combo box keeps the width it had while it was empty
	// and cuts the longest entry.
	a.body = qtlib.NewQWidget2()
	layout := qtlib.NewQHBoxLayout(a.body)
	layout.SetContentsMargins(0, 0, 0, 0)

	a.buildBandCombo(layout)
	a.buildModeCombo(layout)
	a.buildValueCombo(layout)
	a.buildSortControls(layout)

	layout.AddStretch()
}

func (a *spotFilterArea) buildBandCombo(layout *qtlib.QHBoxLayout) {
	a.bandCombo = qtlib.NewQComboBox2()
	a.bandCombo.SetSizeAdjustPolicy(qtlib.QComboBox__AdjustToContents)
	a.bandCombo.OnCurrentIndexChanged(func(index int) {
		if a.updating || a.controller == nil || index < 0 || index >= len(a.bandItems) {
			return
		}
		a.controller.SetFilterBand(a.bandItems[index])
		a.returnFocus()
	})
	layout.AddWidget(qtlib.NewQLabel3("Band").QFrame.QWidget)
	layout.AddWidget(a.bandCombo.QWidget)
}

func (a *spotFilterArea) buildModeCombo(layout *qtlib.QHBoxLayout) {
	a.modeCombo = qtlib.NewQComboBox2()
	a.modeCombo.SetSizeAdjustPolicy(qtlib.QComboBox__AdjustToContents)
	a.modeCombo.OnCurrentIndexChanged(func(index int) {
		if a.updating || a.controller == nil || index < 0 || index >= len(a.modeItems) {
			return
		}
		a.controller.SetFilterMode(a.modeItems[index])
		a.returnFocus()
	})
	layout.AddWidget(qtlib.NewQLabel3("Mode").QFrame.QWidget)
	layout.AddWidget(a.modeCombo.QWidget)
}

func (a *spotFilterArea) buildValueCombo(layout *qtlib.QHBoxLayout) {
	a.valueCombo = qtlib.NewQComboBox2()
	a.valueCombo.SetSizeAdjustPolicy(qtlib.QComboBox__AdjustToContents)
	a.valueCombo.OnCurrentIndexChanged(func(index int) {
		if a.updating || a.controller == nil || index < 0 || index >= len(a.valueItems) {
			return
		}
		a.controller.SetFilterValue(a.valueItems[index])
		a.returnFocus()
	})
	layout.AddWidget(qtlib.NewQLabel3("Value").QFrame.QWidget)
	layout.AddWidget(a.valueCombo.QWidget)
}

func (a *spotFilterArea) buildSortControls(layout *qtlib.QHBoxLayout) {
	a.sortCombo = qtlib.NewQComboBox2()
	a.sortCombo.SetSizeAdjustPolicy(qtlib.QComboBox__AdjustToContents)
	a.sortCombo.OnCurrentIndexChanged(func(index int) {
		if a.updating || a.controller == nil || index < 0 || index >= len(a.sortItems) {
			return
		}
		a.controller.SetFilterSort(a.sortItems[index], a.direction.IsChecked())
		a.returnFocus()
	})
	layout.AddWidget(qtlib.NewQLabel3("Sort by").QFrame.QWidget)
	layout.AddWidget(a.sortCombo.QWidget)

	a.direction = qtlib.NewQToolButton2()
	a.direction.SetFocusPolicy(qtlib.NoFocus)
	a.direction.SetCheckable(true)
	a.direction.SetAutoRaise(true)
	a.direction.SetToolTip("Sort ascending or descending")
	a.direction.OnToggled(func(checked bool) {
		a.applyDirection(checked)
		if a.updating || a.controller == nil {
			return
		}
		a.controller.SetFilterSort(a.currentSortColumn(), checked)
	})
	layout.AddWidget(a.direction.QWidget)
}

func (a *spotFilterArea) returnFocus() {
	// the user works in the central area, the filter is only a short excursion
	if a.focuser == nil {
		return
	}
	a.focuser.GotoEntryFields()
}

func (a *spotFilterArea) Widget() *qtlib.QWidget {
	return a.widget
}

func (a *spotFilterArea) ShowFrame(frame model.SpotFilterFrame) {
	a.updating = true
	defer func() { a.updating = false }()

	a.fillBands(frame.Bands)
	a.fillModes(frame.Modes)
	a.fillValues()
	a.fillSortColumns()

	a.bandCombo.SetCurrentIndex(indexOfBand(a.bandItems, frame.Band))
	a.modeCombo.SetCurrentIndex(indexOfMode(a.modeItems, frame.Mode))
	a.valueCombo.SetCurrentIndex(indexOfValue(a.valueItems, frame.Value))
	a.sortCombo.SetCurrentIndex(indexOfSortColumn(a.sortItems, frame.SortBy))

	a.direction.SetChecked(frame.Descending)
	a.applyDirection(frame.Descending)

	a.summary.SetText(frame.Description)
	a.toggle.SetChecked(frame.Folded)
	a.applyFolded(frame.Folded)
}

func (a *spotFilterArea) RepaintForThemeChange() {
	a.widget.SetPalette(qtlib.QGuiApplication_Palette())
	a.widget.Update()
}

func (a *spotFilterArea) applyFolded(folded bool) {
	a.body.SetVisible(!folded)
	a.summary.SetVisible(folded)
	if folded {
		a.toggle.SetArrowType(qtlib.RightArrow)
	} else {
		a.toggle.SetArrowType(qtlib.DownArrow)
	}
}

func (a *spotFilterArea) applyDirection(descending bool) {
	if descending {
		a.direction.SetText("↓")
	} else {
		a.direction.SetText("↑")
	}
}

func (a *spotFilterArea) currentSortColumn() model.SpotSortColumn {
	index := a.sortCombo.CurrentIndex()
	if index < 0 || index >= len(a.sortItems) {
		return model.SortSpotsByFrequency
	}
	return a.sortItems[index]
}

func (a *spotFilterArea) fillBands(bands []model.Band) {
	items := make([]model.SpotFilterBand, 0, len(bands)+5)
	labels := make([]string, 0, len(bands)+5)
	for _, kind := range specialSpotFilterKinds {
		items = append(items, model.SpotFilterBand{Kind: kind})
		labels = append(labels, spotFilterKindLabel(kind))
	}
	for _, band := range bands {
		items = append(items, model.FixedSpotFilterBand(band))
		labels = append(labels, string(band))
	}
	if len(items) == len(a.bandItems) {
		return
	}
	a.bandItems = items
	a.bandCombo.Clear()
	a.bandCombo.AddItems(labels)
}

func (a *spotFilterArea) fillModes(modes []model.Mode) {
	items := make([]model.SpotFilterMode, 0, len(modes)+5)
	labels := make([]string, 0, len(modes)+5)
	for _, kind := range specialSpotFilterKinds {
		items = append(items, model.SpotFilterMode{Kind: kind})
		labels = append(labels, spotFilterKindLabel(kind))
	}
	for _, mode := range modes {
		items = append(items, model.FixedSpotFilterMode(mode))
		labels = append(labels, string(mode))
	}
	if len(items) == len(a.modeItems) {
		return
	}
	a.modeItems = items
	a.modeCombo.Clear()
	a.modeCombo.AddItems(labels)
}

func (a *spotFilterArea) fillValues() {
	if len(a.valueItems) == len(model.SpotFilterValues) {
		return
	}
	a.valueItems = model.SpotFilterValues
	labels := make([]string, 0, len(model.SpotFilterValues))
	for _, value := range model.SpotFilterValues {
		labels = append(labels, value.Label())
	}
	a.valueCombo.Clear()
	a.valueCombo.AddItems(labels)
}

func (a *spotFilterArea) fillSortColumns() {
	if len(a.sortItems) == len(model.SpotSortColumns) {
		return
	}
	a.sortItems = model.SpotSortColumns
	labels := make([]string, 0, len(model.SpotSortColumns))
	for _, column := range model.SpotSortColumns {
		labels = append(labels, column.Label())
	}
	a.sortCombo.Clear()
	a.sortCombo.AddItems(labels)
}

var specialSpotFilterKinds = []model.SpotFilterKind{
	model.SpotFilterAll,
	model.SpotFilterVFO1,
	model.SpotFilterVFO2,
	model.SpotFilterFocused,
	model.SpotFilterContest,
}

func spotFilterKindLabel(kind model.SpotFilterKind) string {
	switch kind {
	case model.SpotFilterVFO1:
		return "VFO1"
	case model.SpotFilterVFO2:
		return "VFO2"
	case model.SpotFilterFocused:
		return "Focused"
	case model.SpotFilterContest:
		return "Contest"
	default:
		return "All"
	}
}

func indexOfBand(items []model.SpotFilterBand, band model.SpotFilterBand) int {
	for i, item := range items {
		if item == band {
			return i
		}
	}
	return 0
}

func indexOfMode(items []model.SpotFilterMode, mode model.SpotFilterMode) int {
	for i, item := range items {
		if item == mode {
			return i
		}
	}
	return 0
}

func indexOfValue(items []model.SpotFilterValue, value model.SpotFilterValue) int {
	for i, item := range items {
		if item == value {
			return i
		}
	}
	return 0
}

func indexOfSortColumn(items []model.SpotSortColumn, column model.SpotSortColumn) int {
	for i, item := range items {
		if item == column {
			return i
		}
	}
	return 0
}
