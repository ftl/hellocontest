package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

type scoreTableView struct {
	*dockableView
	widget *qtlib.QWidget

	table *scoreTable
	score core.Score
}

func newScoreTableView(parent *qtlib.QWidget) *scoreTableView {
	v := &scoreTableView{}
	v.widget = qtlib.NewQWidget2()
	v.widget.SetObjectName(*qtlib.NewQAnyStringView3("scoreTable"))

	layout := qtlib.NewQVBoxLayout(v.widget)
	layout.SetContentsMargins(0, 0, 0, 0)

	v.table = newScoreTable()
	layout.AddWidget3(v.table.widget.QWidget, 1, 0)

	v.dockableView = newDockableView(parent, v.widget, "Score Table", "scoreTableDock")

	return v
}

func (v *scoreTableView) RepaintForThemeChange() {
	v.dockableView.RepaintForThemeChange()
	v.widget.SetPalette(qtlib.QGuiApplication_Palette())
	v.widget.Update()
	repaintScrollBarsForThemeChange(v.table.widget.QAbstractScrollArea)
}

// Implements core.score.View
func (v *scoreTableView) ShowScore(score core.Score) {
	v.score = score
	v.table.ShowScore(score)
}

func (v *scoreTableView) SetGoals(points, multis int) {
	// Table doesn't display goals - no-op
}

func (v *scoreTableView) SetQTCsEnabled(enabled bool) {
	v.table.SetQTCsEnabled(enabled)
}
