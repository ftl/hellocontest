package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type rateView struct {
	indicatorArea *gtk.DrawingArea

	indicator *rateIndicator
}

func newRateView() *rateView {
	return &rateView{
		indicator: newRateIndicator(),
	}
}

func (v *rateView) setup(builder *gtk.Builder) {
	v.indicatorArea = getUI(builder, "rateIndicatorArea").(*gtk.DrawingArea)

	v.indicatorArea.Connect("draw", v.indicator.Draw)
}

func (v *rateView) ShowRate(rate core.QSORate) {
	v.indicator.SetRate(rate)
	if v.indicatorArea != nil {
		v.indicatorArea.QueueDraw()
	}
}

func (v *rateView) SetGoals(qsos int, points int, multis int) {
	v.indicator.SetGoals(qsos, points, multis)
}
