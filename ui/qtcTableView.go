package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/logbook"
)

var (
	_ logbook.QTCsEnabledListener    = (*qtcTableView)(nil)
	_ logbook.QTCAddedListener       = (*qtcTableView)(nil)
	_ logbook.QTCsClearedListener    = (*qtcTableView)(nil)
	_ logbook.QTCRowSelectedListener = (*qtcTableView)(nil)
)

type qtcTableView struct {
	*dockableView
	widget *qtlib.QWidget

	table *qtcTable
}

func newQTCTableView(parent *qtlib.QWidget) *qtcTableView {
	v := &qtcTableView{}
	v.widget = qtlib.NewQWidget2()

	layout := qtlib.NewQVBoxLayout(v.widget)
	layout.SetContentsMargins(0, 0, 0, 0)

	v.table = newQTCTable()
	layout.AddWidget3(v.table.widget.QWidget, 1, 0)

	v.dockableView = newDockableView(parent, v.widget, "QTCs", "qtcTableDock")

	return v
}

func (v *qtcTableView) RepaintForThemeChange() {
	v.dockableView.RepaintForThemeChange()
	v.widget.SetPalette(qtlib.QGuiApplication_Palette())
	v.widget.Update()
	repaintScrollBarsForThemeChange(v.table.widget.QAbstractScrollArea)
}

func (v *qtcTableView) SetQTCsEnabled(enabled bool) {
	if v.dock == nil {
		return
	}
	if v.dock.IsVisible() == enabled {
		return
	}

	if enabled {
		v.Show()
	} else {
		v.Hide()
	}
}

func (v *qtcTableView) QTCAdded(qtc core.QTC) {
	v.table.QTCAdded(qtc)
}

func (v *qtcTableView) QTCsCleared() {
	v.table.QTCsCleared()
}

func (v *qtcTableView) QTCRowSelected(index int) {
	v.table.QTCRowSelected(index)
}
