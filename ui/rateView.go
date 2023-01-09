package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type rateView struct {
	indicatorArea *gtk.DrawingArea

	indicator *rateIndicator
}

func setupRateView(builder *gtk.Builder) *rateView {
	result := new(rateView)
	result.indicatorArea = getUI(builder, "rateIndicatorArea").(*gtk.DrawingArea)
	result.indicator = newRateIndicator()

	result.indicatorArea.Connect("draw", result.indicator.Draw)

	return result
}

func (v *rateView) ShowRate(rate core.QSORate) {
	if v == nil {
		return
	}

	v.indicator.SetRate(rate)
	v.indicatorArea.QueueDraw()
}
