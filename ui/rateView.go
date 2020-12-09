package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type rateView struct {
	tableLabel *gtk.Label
}

func setupRateView(builder *gtk.Builder) *rateView {
	result := new(rateView)

	result.tableLabel = getUI(builder, "rateTableLabel").(*gtk.Label)

	return result
}

func (v *rateView) ShowRate(rate core.QSORate) {
	if v == nil {
		return
	}

	text := `<span allow_breaks='true' font_family='monospace'>last 60min: %3d Q/h
last  5min: %3d Q/h
last QSO: %9s
</span>`

	renderedRate := fmt.Sprintf(text, rate.LastHourRate, rate.Last5MinRate, rate.SinceLastQSOFormatted())
	v.tableLabel.SetMarkup(renderedRate)
}
