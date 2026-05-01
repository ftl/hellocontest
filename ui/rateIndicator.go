package ui

import (
	"fmt"
	"math"
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

const (
	rateFontSize           = 15
	rateAxisMargin         = 15.0
	rateAreaAlpha          = 0.4
	rateBorderAlpha        = 0.8
	rateAxisLineWidth      = 1.0
	rateValueDotRadius     = 5.0
	rateValueLineWidth     = 5.0
	rateTimeLineWidth      = 10.0
	rateAngleRotation      = (3.0 / 2.0) * math.Pi
	rateIndicatorMinWidth  = 160
	rateIndicatorMinHeight = 160
)

var (
	// rateGradient: red → orange → green
	rateGradient = [][3]int{
		{0xFF, 0x00, 0x00},
		{0xFF, 0x99, 0x33},
		{0x00, 0xCC, 0x00},
	}

	// timeGradient: first 3 steps red→orange→green, last 3 stay green
	timeGradient = [][3]int{
		{0xFF, 0x00, 0x00},
		{0xFF, 0x99, 0x33},
		{0x00, 0xCC, 0x00},
		{0x00, 0xCC, 0x00},
		{0x00, 0xCC, 0x00},
		{0x00, 0xCC, 0x00},
	}
)

type point struct {
	x, y float64
}

func (p point) translate(dx, dy float64) point {
	return point{x: p.x + dx, y: p.y + dy}
}

type polar struct {
	radius  float64
	radians float64
}

func (p polar) toPoint() point {
	return point{
		x: p.radius * math.Cos(p.radians),
		y: p.radius * math.Sin(p.radians),
	}
}

func degreesToRadians(d float64) float64 {
	return d / 180.0 * math.Pi
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// interpRGB linearly interpolates between color stops based on t [0, 1]
func interpRGB(stops [][3]int, t float64) [3]int {
	if len(stops) == 0 {
		return [3]int{0, 0, 0}
	}
	if len(stops) == 1 {
		return stops[0]
	}
	t = clamp01(t)
	// Which segment?
	segmentSize := 1.0 / float64(len(stops)-1)
	segment := int(t / segmentSize)
	if segment >= len(stops)-1 {
		return stops[len(stops)-1]
	}
	// Interpolation within segment
	tLocal := (t - float64(segment)*segmentSize) / segmentSize
	start := stops[segment]
	end := stops[segment+1]
	return [3]int{
		start[0] + int(float64(end[0]-start[0])*tLocal),
		start[1] + int(float64(end[1]-start[1])*tLocal),
		start[2] + int(float64(end[2]-start[2])*tLocal),
	}
}

type rateAxis struct {
	unit        string
	angle       float64
	value1      float64
	value2      float64
	goalValue   float64
	maxValue    float64
	achievement float64

	labelPoint  point
	axisEndPt   point
	value1Point point
	value2Point point
	goalPoint   point
}

func newRateAxis(unit string, angleDegrees float64) *rateAxis {
	return &rateAxis{
		unit:        unit,
		angle:       degreesToRadians(angleDegrees) + rateAngleRotation,
		goalValue:   0,
		maxValue:    0,
		achievement: 1,
	}
}

func (a *rateAxis) SetValues(value1, value2 float64) {
	a.value1 = value1
	a.value2 = value2
	a.updateAchievement()
}

func (a *rateAxis) SetGoal(goal float64) {
	a.goalValue = goal
	a.updateMaxValue()
	a.updateAchievement()
}

func (a *rateAxis) updateMaxValue() {
	a.maxValue = 1.5 * a.goalValue
}

func (a *rateAxis) updateAchievement() {
	if a.goalValue == 0 {
		a.achievement = 1
	} else {
		a.achievement = a.value1 / a.goalValue
	}
}

func (a *rateAxis) LabelText() string {
	return fmt.Sprintf("%2.0f %s", a.value1, a.unit)
}

type timeIndicator struct {
	goalValue   float64
	currentTime time.Duration
	labelText   string
	goalSeconds float64
	achievement float64
}

func newTimeIndicator() *timeIndicator {
	return &timeIndicator{
		goalValue: 0,
	}
}

func (ti *timeIndicator) SetGoal(goal float64) {
	ti.goalValue = goal
	ti.updateGoalSeconds()
}

func (ti *timeIndicator) updateGoalSeconds() {
	if ti.goalValue == 0 {
		ti.goalSeconds = 0
	} else {
		ti.goalSeconds = time.Hour.Seconds() / ti.goalValue
	}
}

func (ti *timeIndicator) SetCurrentTime(dur time.Duration, text string) {
	ti.currentTime = dur
	ti.labelText = text
	ti.updateAchievement()
}

func (ti *timeIndicator) updateAchievement() {
	if ti.goalSeconds == 0 {
		ti.achievement = 1
	} else {
		ti.achievement = 1 - math.Min(1, ti.currentTime.Seconds()/ti.goalSeconds)
	}
}

type rateIndicator struct {
	widget *qtlib.QWidget

	qAxis         *rateAxis
	pAxis         *rateAxis
	mAxis         *rateAxis
	timeIndicator *timeIndicator
}

func newRateIndicator() *rateIndicator {
	ri := &rateIndicator{
		qAxis:         newRateAxis("Q/h", 0),
		pAxis:         newRateAxis("P/h", 120),
		mAxis:         newRateAxis("M/h", 240),
		timeIndicator: newTimeIndicator(),
	}
	ri.widget = qtlib.NewQWidget2()
	ri.widget.SetMinimumWidth(rateIndicatorMinWidth)
	ri.widget.SetMinimumHeight(rateIndicatorMinHeight)
	ri.widget.OnPaintEvent(func(super func(ev *qtlib.QPaintEvent), ev *qtlib.QPaintEvent) {
		super(ev)
		ri.paint()
	})
	return ri
}

func (ri *rateIndicator) SetRate(rate core.QSORate) {
	ri.qAxis.SetValues(float64(rate.Last5MinRate), float64(rate.LastHourRate))
	ri.pAxis.SetValues(float64(rate.Last5MinPoints), float64(rate.LastHourPoints))
	ri.mAxis.SetValues(float64(rate.Last5MinMultis), float64(rate.LastHourMultis))
	ri.timeIndicator.SetCurrentTime(rate.SinceLastQSO, rate.SinceLastQSOFormatted())
}

func (ri *rateIndicator) SetGoals(qsos, points, multis int) {
	ri.qAxis.SetGoal(float64(qsos))
	ri.pAxis.SetGoal(float64(points))
	ri.mAxis.SetGoal(float64(multis))
	ri.timeIndicator.SetGoal(float64(qsos))
}

func (ri *rateIndicator) paint() {
	painter := qtlib.NewQPainter2(ri.widget.QPaintDevice)
	defer painter.End()
	painter.SetRenderHint(qtlib.QPainter__Antialiasing)

	// Get colors from widget palette for theme support
	palette := qtlib.QGuiApplication_Palette()
	bgColor := palette.ColorWithCr(qtlib.QPalette__Window)
	fgColor := palette.ColorWithCr(qtlib.QPalette__WindowText)

	bgRGB := [3]int{bgColor.Red(), bgColor.Green(), bgColor.Blue()}
	fgRGB := [3]int{fgColor.Red(), fgColor.Green(), fgColor.Blue()}

	// Derive axis color as mid-tone between bg and fg
	axisRGB := [3]int{
		(bgRGB[0] + fgRGB[0]) / 2,
		(bgRGB[1] + fgRGB[1]) / 2,
		(bgRGB[2] + fgRGB[2]) / 2,
	}

	// Low zone (goal triangle border) is a lighter version of axis color
	lowZoneRGB := [3]int{
		(bgRGB[0] + axisRGB[0]) / 2,
		(bgRGB[1] + axisRGB[1]) / 2,
		(bgRGB[2] + axisRGB[2]) / 2,
	}

	w := float64(ri.widget.Width())
	h := float64(ri.widget.Height())
	center := point{x: w / 2, y: h / 2}
	axisLength := math.Min(w, h)/2 - rateAxisMargin

	// Prepare geometry for all axes
	ri.prepareGeometry(painter, center, axisLength)

	// Fill background with theme color
	ri.fillBackgroundWithColor(painter, w, h, bgRGB)

	// Draw goal triangle with theme colors
	ri.drawTriangle(painter, [3]point{ri.qAxis.goalPoint, ri.pAxis.goalPoint, ri.mAxis.goalPoint}, lowZoneRGB, rateAreaAlpha, rateBorderAlpha)

	// Draw achievement triangle
	overallAchievement := (ri.qAxis.achievement + ri.pAxis.achievement + ri.mAxis.achievement) / 3
	achievementColor := interpRGB(rateGradient, overallAchievement)
	ri.drawTriangle(painter, [3]point{ri.qAxis.value1Point, ri.pAxis.value1Point, ri.mAxis.value1Point}, achievementColor, rateAreaAlpha, rateBorderAlpha)

	// Set up font for text
	font := qtlib.NewQFont()
	font.SetPixelSize(rateFontSize)
	painter.SetFont(font)

	// Draw per-axis overlays with theme colors
	ri.drawAxis(painter, ri.qAxis, center, fgRGB, axisRGB)
	ri.drawAxis(painter, ri.pAxis, center, fgRGB, axisRGB)
	ri.drawAxis(painter, ri.mAxis, center, fgRGB, axisRGB)

	// Draw time indicator with theme colors
	ri.drawTimeIndicator(painter, center, axisLength, fgRGB)
}

func (ri *rateIndicator) prepareGeometry(painter *qtlib.QPainter, center point, axisLength float64) {
	axes := []*rateAxis{ri.qAxis, ri.pAxis, ri.mAxis}
	fm := painter.FontMetrics()

	for _, axis := range axes {
		if axis.goalValue == 0 {
			axis.value1Point = center
			axis.value2Point = center
			axis.goalPoint = center
			axis.axisEndPt = center
		} else {
			axis.value1Point = polar{
				radius:  math.Min((axis.value1/axis.maxValue)*axisLength, axisLength),
				radians: axis.angle,
			}.toPoint().translate(center.x, center.y)

			axis.value2Point = polar{
				radius:  math.Min((axis.value2/axis.maxValue)*axisLength, axisLength),
				radians: axis.angle,
			}.toPoint().translate(center.x, center.y)

			axis.goalPoint = polar{
				radius:  math.Min((axis.goalValue/axis.maxValue)*axisLength, axisLength),
				radians: axis.angle,
			}.toPoint().translate(center.x, center.y)

			axis.axisEndPt = polar{radius: axisLength, radians: axis.angle}.toPoint().translate(center.x, center.y)
		}

		// Label positioning based on axis direction
		labelText := axis.LabelText()
		textWidth := float64(fm.HorizontalAdvance(labelText))
		textHeight := float64(fm.Height())

		switch {
		case axis.axisEndPt.y < center.y: // Top half (Q/h at 12 o'clock)
			// Label positioned to the right of the goal point
			// Text top edge aligned with goal point (top tip of gray triangle)
			// Note: drawAxis adds textHeight for baseline, so we subtract here
			axis.labelPoint = point{
				x: axis.goalPoint.x + 10.0,
				y: axis.goalPoint.y - textHeight,
			}
		case axis.axisEndPt.x < center.x: // Left half (M/h)
			axis.labelPoint = point{
				x: center.x - axisLength,
				y: center.y - textHeight/2.0,
			}
		case axis.axisEndPt.x > center.x: // Right half (P/h)
			axis.labelPoint = point{
				x: center.x + axisLength - textWidth,
				y: center.y - textHeight/2.0,
			}
		}
	}
}

func (ri *rateIndicator) fillBackground(painter *qtlib.QPainter, w, h float64) {
	color := newQColor(backgroundRGB)
	brush := qtlib.NewQBrush3(color)
	painter.FillRect2(0, 0, int(w), int(h), brush)
}

func (ri *rateIndicator) fillBackgroundWithColor(painter *qtlib.QPainter, w, h float64, bgRGB [3]int) {
	color := qtlib.NewQColor11(bgRGB[0], bgRGB[1], bgRGB[2], 255)
	brush := qtlib.NewQBrush3(color)
	painter.FillRect2(0, 0, int(w), int(h), brush)
}

func (ri *rateIndicator) drawTriangle(painter *qtlib.QPainter, points [3]point, rgb [3]int, fillAlpha, borderAlpha float32) {
	// Fill
	fillColor := qtlib.NewQColor11(rgb[0], rgb[1], rgb[2], 255)
	fillColor.SetAlphaF(fillAlpha)
	fillBrush := qtlib.NewQBrush3(fillColor)

	path := qtlib.NewQPainterPath()
	path.MoveTo2(points[0].x, points[0].y)
	path.LineTo2(points[1].x, points[1].y)
	path.LineTo2(points[2].x, points[2].y)
	path.CloseSubpath()
	painter.FillPath(path, fillBrush)

	// Border
	borderColor := qtlib.NewQColor11(rgb[0], rgb[1], rgb[2], 255)
	borderColor.SetAlphaF(borderAlpha)
	pen := qtlib.NewQPen()
	pen.SetColor(borderColor)
	painter.SetPenWithPen(pen)
	painter.SetBrush(qtlib.NewQBrush())
	painter.DrawPath(path)
}

func (ri *rateIndicator) drawAxis(painter *qtlib.QPainter, axis *rateAxis, center point, fgRGB [3]int, axisRGB [3]int) {
	// Axis line
	axisPen := qtlib.NewQPen()
	axisPen.SetColor(qtlib.NewQColor11(axisRGB[0], axisRGB[1], axisRGB[2], 255))
	axisPen.SetWidthF(rateAxisLineWidth)
	painter.SetPenWithPen(axisPen)
	painter.DrawLine(qtlib.NewQLineF3(center.x, center.y, axis.axisEndPt.x, axis.axisEndPt.y))

	// Label text
	fontPen := qtlib.NewQPen()
	fontPen.SetColor(qtlib.NewQColor11(fgRGB[0], fgRGB[1], fgRGB[2], 255))
	painter.SetPenWithPen(fontPen)
	fm := painter.FontMetrics()
	painter.DrawText(qtlib.NewQPointF3(axis.labelPoint.x, axis.labelPoint.y+float64(fm.Height())), axis.LabelText())

	// Filled dot at value1Point (5-min rate)
	dotColor := interpRGB(rateGradient, axis.achievement)
	dotQColor := qtlib.NewQColor11(dotColor[0], dotColor[1], dotColor[2], 255)
	dotBrush := qtlib.NewQBrush3(dotQColor)
	painter.SetBrush(dotBrush)
	painter.SetPenWithPen(qtlib.NewQPen())

	// Draw dot as filled circle using path with arc
	// Draw filled circle as an ellipse path
	dotPath := qtlib.NewQPainterPath()
	rectF := qtlib.NewQRectF4(axis.value1Point.x-rateValueDotRadius, axis.value1Point.y-rateValueDotRadius, rateValueDotRadius*2, rateValueDotRadius*2)
	dotPath.ArcMoveTo(rectF, 0.0)
	dotPath.ArcTo(rectF, 0.0, 360.0)
	dotPath.CloseSubpath()
	painter.FillPath(dotPath, dotBrush)

	// Line from value2Point to value1Point (last-hour to 5-min)
	linePen := qtlib.NewQPen()
	linePen.SetColor(dotQColor)
	linePen.SetWidthF(rateValueLineWidth)
	painter.SetPenWithPen(linePen)
	painter.DrawLine(qtlib.NewQLineF3(axis.value2Point.x, axis.value2Point.y, axis.value1Point.x, axis.value1Point.y))
}

func (ri *rateIndicator) drawTimeIndicator(painter *qtlib.QPainter, center point, axisLength float64, fgRGB [3]int) {
	w := float64(ri.widget.Width())
	h := float64(ri.widget.Height())
	radius := math.Min(w, h)/2 - rateTimeLineWidth/2

	// Arc color from gradient
	arcColor := interpRGB(timeGradient, ri.timeIndicator.achievement)
	arcQColor := qtlib.NewQColor11(arcColor[0], arcColor[1], arcColor[2], 255)

	// Draw arc
	arcPen := qtlib.NewQPen()
	arcPen.SetColor(arcQColor)
	arcPen.SetWidthF(rateTimeLineWidth)
	painter.SetPenWithPen(arcPen)
	painter.SetBrush(qtlib.NewQBrush())

	// Cairo uses angleRotation (3π/2) starting from 12 o'clock, spanning clockwise
	// Qt's ArcTo uses degrees: 0° = 3 o'clock, counter-clockwise positive
	// Conversion: Cairo 12 o'clock (3π/2 rad) → Qt 90°
	// Cairo clockwise → Qt negative sweep
	span := (1 - ri.timeIndicator.achievement) * 2 * math.Pi
	spanDeg := -span * 180 / math.Pi

	rect := qtlib.NewQRectF4(center.x-radius, center.y-radius, radius*2, radius*2)
	path := qtlib.NewQPainterPath()
	path.ArcMoveTo(rect, 90.0)
	path.ArcTo(rect, 90.0, spanDeg)
	painter.DrawPath(path)

	// Label at 6 o'clock
	fm := painter.FontMetrics()
	labelWidth := float64(fm.HorizontalAdvance(ri.timeIndicator.labelText))
	labelHeight := float64(fm.Height())
	labelX := center.x - labelWidth/2
	labelY := center.y + radius/2 + labelHeight

	fontPen := qtlib.NewQPen()
	fontPen.SetColor(qtlib.NewQColor11(fgRGB[0], fgRGB[1], fgRGB[2], 255))
	painter.SetPenWithPen(fontPen)
	painter.DrawText(qtlib.NewQPointF3(labelX, labelY), ri.timeIndicator.labelText)
}
