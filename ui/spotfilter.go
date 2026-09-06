package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

type SpotFilterController interface {
	SetFilterBand(core.SpotFilterBand)
	SetFilterMode(core.SpotFilterMode)
	SetFilterSort(core.SpotSortColumn, bool)
	SetFilterFolded(bool)
}

type spotFilterArea struct {
	widget  *qtlib.QWidget
	toggle  *qtlib.QToolButton
	summary *qtlib.QLabel
	body    *qtlib.QWidget

	bandCombo *qtlib.QComboBox
	modeCombo *qtlib.QComboBox
	sortCombo *qtlib.QComboBox
	direction *qtlib.QToolButton

	bandItems []core.SpotFilterBand
	modeItems []core.SpotFilterMode
	sortItems []core.SpotSortColumn

	controller SpotFilterController
	focuser    EntryFocuser
	updating   bool
}

func newSpotFilterArea(controller SpotFilterController, focuser EntryFocuser) *spotFilterArea {
	a := &spotFilterArea{controller: controller, focuser: focuser}

	a.widget = qtlib.NewQWidget2()
	a.widget.SetObjectName(*qtlib.NewQAnyStringView3("spotFilter"))
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
	a.body = qtlib.NewQWidget2()
	layout := qtlib.NewQHBoxLayout(a.body)
	layout.SetContentsMargins(0, 0, 0, 0)

	a.buildBandCombo(layout)
	a.buildModeCombo(layout)
	a.buildSortControls(layout)

	layout.AddStretch()
}

func (a *spotFilterArea) buildBandCombo(layout *qtlib.QHBoxLayout) {
	a.bandCombo = qtlib.NewQComboBox2()
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

func (a *spotFilterArea) buildSortControls(layout *qtlib.QHBoxLayout) {
	a.sortCombo = qtlib.NewQComboBox2()
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

func (a *spotFilterArea) ShowFrame(frame core.SpotFilterFrame) {
	a.updating = true
	defer func() { a.updating = false }()

	a.fillBands(frame.Bands)
	a.fillModes(frame.Modes)
	a.fillSortColumns()

	a.bandCombo.SetCurrentIndex(indexOfBand(a.bandItems, frame.Band))
	a.modeCombo.SetCurrentIndex(indexOfMode(a.modeItems, frame.Mode))
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

func (a *spotFilterArea) currentSortColumn() core.SpotSortColumn {
	index := a.sortCombo.CurrentIndex()
	if index < 0 || index >= len(a.sortItems) {
		return core.SortSpotsByFrequency
	}
	return a.sortItems[index]
}

func (a *spotFilterArea) fillBands(bands []core.Band) {
	items := make([]core.SpotFilterBand, 0, len(bands)+5)
	labels := make([]string, 0, len(bands)+5)
	for _, kind := range specialSpotFilterKinds {
		items = append(items, core.SpotFilterBand{Kind: kind})
		labels = append(labels, spotFilterKindLabel(kind))
	}
	for _, band := range bands {
		items = append(items, core.FixedSpotFilterBand(band))
		labels = append(labels, string(band))
	}
	if len(items) == len(a.bandItems) {
		return
	}
	a.bandItems = items
	a.bandCombo.Clear()
	a.bandCombo.AddItems(labels)
}

func (a *spotFilterArea) fillModes(modes []core.Mode) {
	items := make([]core.SpotFilterMode, 0, len(modes)+5)
	labels := make([]string, 0, len(modes)+5)
	for _, kind := range specialSpotFilterKinds {
		items = append(items, core.SpotFilterMode{Kind: kind})
		labels = append(labels, spotFilterKindLabel(kind))
	}
	for _, mode := range modes {
		items = append(items, core.FixedSpotFilterMode(mode))
		labels = append(labels, string(mode))
	}
	if len(items) == len(a.modeItems) {
		return
	}
	a.modeItems = items
	a.modeCombo.Clear()
	a.modeCombo.AddItems(labels)
}

func (a *spotFilterArea) fillSortColumns() {
	if len(a.sortItems) == len(core.SpotSortColumns) {
		return
	}
	a.sortItems = core.SpotSortColumns
	labels := make([]string, 0, len(core.SpotSortColumns))
	for _, column := range core.SpotSortColumns {
		labels = append(labels, column.Label())
	}
	a.sortCombo.Clear()
	a.sortCombo.AddItems(labels)
}

var specialSpotFilterKinds = []core.SpotFilterKind{
	core.SpotFilterAll,
	core.SpotFilterVFO1,
	core.SpotFilterVFO2,
	core.SpotFilterFocused,
	core.SpotFilterContest,
}

func spotFilterKindLabel(kind core.SpotFilterKind) string {
	switch kind {
	case core.SpotFilterVFO1:
		return "VFO1"
	case core.SpotFilterVFO2:
		return "VFO2"
	case core.SpotFilterFocused:
		return "Focused"
	case core.SpotFilterContest:
		return "Contest"
	default:
		return "All"
	}
}

func indexOfBand(items []core.SpotFilterBand, band core.SpotFilterBand) int {
	for i, item := range items {
		if item == band {
			return i
		}
	}
	return 0
}

func indexOfMode(items []core.SpotFilterMode, mode core.SpotFilterMode) int {
	for i, item := range items {
		if item == mode {
			return i
		}
	}
	return 0
}

func indexOfSortColumn(items []core.SpotSortColumn, column core.SpotSortColumn) int {
	for i, item := range items {
		if item == column {
			return i
		}
	}
	return 0
}
