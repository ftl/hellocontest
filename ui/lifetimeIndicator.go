package ui

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

var lifetimeStyle = struct {
	color color
}{
	color: color{0, 0, 0},
}

type lifetimeIndicator struct {
	row *gtk.ListBoxRow

	lifetime lifetimeFunc
}

type lifetimeFunc func(int) float64

func newLifetimeIndicator(row *gtk.ListBoxRow, lifetime lifetimeFunc) *lifetimeIndicator {
	return &lifetimeIndicator{
		row:      row,
		lifetime: lifetime,
	}
}

func (ind *lifetimeIndicator) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	index := ind.row.GetIndex()
	lifetime := ind.lifetime(index)

	height := float64(da.GetAllocatedHeight())
	width := lifetime * float64(da.GetAllocatedWidth())

	cr.SetSourceRGB(lifetimeStyle.color.toRGB())
	cr.MoveTo(0, 0)
	cr.LineTo(width, 0)
	cr.LineTo(width, height)
	cr.LineTo(0, height)
	cr.ClosePath()
	cr.Fill()
}
