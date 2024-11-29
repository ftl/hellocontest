//go:build !fyne

package ui

import (
	"math"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/ui/style"
)

var proximityStyle = struct {
	color           style.Color
	colorExactMatch style.Color
}{
	color:           style.Black,
	colorExactMatch: style.Green,
}

type proximityIndicator struct {
	row *gtk.ListBoxRow

	proximity proximityFunc
}

type proximityFunc func(int) float64

func newProximityIndicator(row *gtk.ListBoxRow, proximity proximityFunc) *proximityIndicator {
	return &proximityIndicator{
		row:       row,
		proximity: proximity,
	}
}

func (ind *proximityIndicator) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	index := ind.row.GetIndex()
	proximity := ind.proximity(index)

	height := proximity * float64(da.GetAllocatedHeight())
	width := float64(da.GetAllocatedWidth())

	startHeight := float64(0)
	if proximity < 0 {
		startHeight = float64(da.GetAllocatedHeight())
	}
	endHeight := startHeight + height

	color := proximityStyle.color
	if math.Abs(proximity) == 1.0 {
		color = proximityStyle.colorExactMatch
	}

	cr.SetSourceRGB(color.ToRGB())
	cr.MoveTo(0, startHeight)
	cr.LineTo(width, startHeight)
	cr.LineTo(width, endHeight)
	cr.LineTo(0, endHeight)
	cr.ClosePath()
	cr.Fill()
}
