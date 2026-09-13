package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

type bandMatrixView struct {
	*dockableView
	widget *qtlib.QWidget

	matrix *bandMatrix
}

func newBandMatrixView(parent *qtlib.QWidget, controller BandMatrixController, focuser EntryFocuser) *bandMatrixView {
	v := &bandMatrixView{}
	v.widget = qtlib.NewQWidget2()
	v.widget.SetObjectName(*qtlib.NewQAnyStringView3("bandMatrix"))

	layout := qtlib.NewQVBoxLayout(v.widget)
	layout.SetContentsMargins(0, 0, 0, 0)

	v.matrix = newBandMatrix(focuser)
	v.matrix.SetController(controller)
	layout.AddWidget3(v.matrix.widget.QAbstractScrollArea.QFrame.QWidget, 1, 0)

	v.dockableView = newDockableView(parent, v.widget, "Band Matrix", "bandMatrixDock")

	return v
}

func (v *bandMatrixView) RepaintForThemeChange() {
	v.dockableView.RepaintForThemeChange()
	v.widget.SetPalette(qtlib.QGuiApplication_Palette())
	v.widget.Update()
	repaintScrollBarsForThemeChange(v.matrix.widget.QAbstractScrollArea)
}

func (v *bandMatrixView) ShowFrame(frame core.BandMatrixFrame) {
	v.matrix.ShowFrame(frame)
}
