package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/logbook"
)

var (
	_ logbook.QSOAddedListener       = (*qsoTableView)(nil)
	_ logbook.QSOsClearedListener    = (*qsoTableView)(nil)
	_ logbook.QSORowSelectedListener = (*qsoTableView)(nil)
)

type qsoTableView struct {
	*dockableView
	widget *qtlib.QWidget

	table *qsoTable
}

func newQSOTableView(parent *qtlib.QWidget, controller QSOListController) *qsoTableView {
	v := &qsoTableView{}
	v.widget = qtlib.NewQWidget2()

	layout := qtlib.NewQVBoxLayout(v.widget)
	layout.SetContentsMargins(0, 0, 0, 0)

	v.table = newQSOTable()
	v.table.SetQSOListController(controller)
	layout.AddWidget3(v.table.table.QWidget, 1, 0)

	v.dockableView = newDockableView(parent, v.widget, "QSOs", "qsoTableDock")

	return v
}

func (v *qsoTableView) RepaintForThemeChange() {
	v.dockableView.RepaintForThemeChange()
	v.widget.SetPalette(qtlib.QGuiApplication_Palette())
	v.widget.Update()
	repaintScrollBarsForThemeChange(v.table.table.QAbstractScrollArea)
}

func (v *qsoTableView) QSOAdded(qso core.QSO) {
	v.table.QSOAdded(qso)
}

func (v *qsoTableView) QSOsCleared() {
	v.table.QSOsCleared()
}

func (v *qsoTableView) QSORowSelected(row int) {
	v.table.QSORowSelected(row)
}
