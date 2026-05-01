package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/rate"
)

var _ rate.View = (*rateView)(nil)

type rateView struct {
	*dockableView
	widget *qtlib.QFrame

	indicator *rateIndicator
}

func newRateView(parent *qtlib.QWidget) *rateView {
	v := &rateView{
		indicator: newRateIndicator(),
	}

	v.widget = qtlib.NewQFrame2()
	v.widget.SetObjectName(*qtlib.NewQAnyStringView3("rate"))
	v.widget.SetFrameShape(qtlib.QFrame__StyledPanel)
	v.widget.SetFrameShadow(qtlib.QFrame__Plain)

	layout := qtlib.NewQVBoxLayout(v.widget.QWidget)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.AddWidget3(v.indicator.widget, 1, 0)

	v.dockableView = newDockableView(parent, v.widget.QWidget, "QSO Rate", "rateDock")

	return v
}

func (v *rateView) RepaintForThemeChange() {
	v.dockableView.RepaintForThemeChange()
	v.widget.SetPalette(qtlib.QGuiApplication_Palette())
	v.widget.Update()
}

func (v *rateView) ShowRate(rate core.QSORate) {
	v.indicator.SetRate(rate)
	v.indicator.widget.Update()
}

func (v *rateView) SetGoals(qsos int, points int, multis int) {
	v.indicator.SetGoals(qsos, points, multis)
}
