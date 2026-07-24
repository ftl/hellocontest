package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core/clock"
)

var _ clock.View = (*clockView)(nil)

// clockFillFraction leaves a small margin so the digits don't touch the edges.
const clockFillFraction = 0.9

// clockBasePixelSize is the reference size used to measure the text before
// scaling it to fill the widget.
const clockBasePixelSize = 100

type clockView struct {
	*dockableView
	widget *qtlib.QWidget

	text string
}

func newClockView(parent *qtlib.QWidget) *clockView {
	v := &clockView{
		text: "00:00:00",
	}

	v.widget = qtlib.NewQWidget2()
	v.widget.SetObjectName(*qtlib.NewQAnyStringView3("clock"))
	v.widget.SetMinimumWidth(120)
	v.widget.SetMinimumHeight(40)
	v.widget.OnPaintEvent(func(super func(ev *qtlib.QPaintEvent), ev *qtlib.QPaintEvent) {
		super(ev)
		v.paint()
	})

	v.dockableView = newDockableView(parent, v.widget, "UTC", "clockDock")

	return v
}

func (v *clockView) RepaintForThemeChange() {
	v.dockableView.RepaintForThemeChange()
	v.widget.SetPalette(qtlib.QGuiApplication_Palette())
	v.widget.Update()
}

func (v *clockView) SetTime(text string) {
	v.text = text
	v.widget.Update()
}

func (v *clockView) paint() {
	painter := qtlib.NewQPainter2(v.widget.QPaintDevice)
	defer painter.End()
	painter.SetRenderHint(qtlib.QPainter__Antialiasing)

	fgColor := qtlib.QGuiApplication_Palette().ColorWithCr(qtlib.QPalette__WindowText)
	painter.SetPen(fgColor)

	font := qtlib.NewQFont()
	font.SetFamily("monospace")
	font.SetStyleHint(qtlib.QFont__TypeWriter)
	font.SetFixedPitch(true)
	font.SetBold(true)

	// Measure at a fixed base size, then scale to fill both dimensions.
	font.SetPixelSize(clockBasePixelSize)
	painter.SetFont(font)
	fm := painter.FontMetrics()
	baseW := float64(fm.HorizontalAdvance(v.text))
	baseH := float64(fm.Height())
	if baseW <= 0 || baseH <= 0 {
		return
	}

	usableW := float64(v.widget.Width()) * clockFillFraction
	usableH := float64(v.widget.Height()) * clockFillFraction
	scale := min(usableW/baseW, usableH/baseH)
	size := int(clockBasePixelSize * scale)
	if size < 1 {
		size = 1
	}

	font.SetPixelSize(size)
	painter.SetFont(font)
	fm = painter.FontMetrics()

	textW := fm.HorizontalAdvance(v.text)
	x := (v.widget.Width() - textW) / 2
	// DrawText's y is the baseline; center the text box vertically.
	y := (v.widget.Height()-fm.Height())/2 + fm.Ascent()
	painter.DrawText(qtlib.NewQPointF3(float64(x), float64(y)), v.text)
}
