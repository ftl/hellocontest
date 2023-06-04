package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

type rateView struct {
	indicatorArea *gtk.DrawingArea

	indicator *rateIndicator
}

func setupNewRateView(builder *gtk.Builder, style *style.Style) *rateView {
	result := &rateView{}

	result.indicatorArea = getUI(builder, "rateIndicatorArea").(*gtk.DrawingArea)
	result.indicator = newRateIndicator(style.ForWidget(result.indicatorArea.ToWidget()))
	result.indicatorArea.Connect("draw", result.indicator.Draw)
	result.indicatorArea.Connect("style-updated", result.indicator.RefreshStyle)

	return result
}

func (v *rateView) ShowRate(rate core.QSORate) {
	if v == nil {
		return
	}
	v.indicator.SetRate(rate)
	if v.indicatorArea != nil {
		v.indicatorArea.QueueDraw()
	}
}

func (v *rateView) SetGoals(qsos int, points int, multis int) {
	if v == nil {
		return
	}
	v.indicator.SetGoals(qsos, points, multis)
}
