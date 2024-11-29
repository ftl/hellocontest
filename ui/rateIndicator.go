//go:build !fyne

package ui

import (
	"fmt"
	"math"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

var rateColors = style.ColorMap{
	style.NewRGB(1, 0, 0),
	style.NewRGB(1, 0.6, 0.2),
	style.NewRGB(0, 0.8, 0),
}

var timeColors = style.ColorMap{
	style.NewRGB(1, 0, 0),
	style.NewRGB(1, 0.6, 0.2),
	style.NewRGB(0, 0.8, 0),
	style.NewRGB(0, 0.8, 0),
	style.NewRGB(0, 0.8, 0),
	style.NewRGB(0, 0.8, 0),
}

const angleRotation = (3.0 / 2.0) * math.Pi

type rateStyle struct {
	colorProvider

	backgroundColor    style.Color
	fontColor          style.Color
	fontSize           float64
	axisColor          style.Color
	axisMargin         float64
	areaAlpha          float64
	borderAlpha        float64
	lowZoneColor       style.Color
	timeIndicatorWidth float64
}

func (s *rateStyle) Refresh() {
	s.backgroundColor = s.colorProvider.BackgroundColor()
	s.fontColor = s.colorProvider.ForegroundColor()
	s.axisColor = s.colorProvider.ColorByName(axisColorName)
	s.lowZoneColor = s.colorProvider.ColorByName(lowZoneColorName)
}

type rateIndicator struct {
	qAxis         *rateAxis
	pAxis         *rateAxis
	mAxis         *rateAxis
	timeIndicator *timeIndicator
	style         *rateStyle
}

func newRateIndicator(colors colorProvider) *rateIndicator {
	style := &rateStyle{
		colorProvider: colors,

		fontSize:           15,
		axisMargin:         15,
		areaAlpha:          0.4,
		borderAlpha:        0.8,
		timeIndicatorWidth: 10,
	}
	style.Refresh()

	return &rateIndicator{
		qAxis:         newRateAxis("Q/h", 0, style),
		pAxis:         newRateAxis("P/h", 120, style),
		mAxis:         newRateAxis("M/h", 240, style),
		timeIndicator: newTimeIndicator(style),
		style:         style,
	}
}

func (ind *rateIndicator) RefreshStyle() {
	ind.style.Refresh()
}

func (ind *rateIndicator) SetRate(rate core.QSORate) {
	ind.qAxis.SetValues(float64(rate.Last5MinRate), float64(rate.LastHourRate))
	ind.pAxis.SetValues(float64(rate.Last5MinPoints), float64(rate.LastHourPoints))
	ind.mAxis.SetValues(float64(rate.Last5MinMultis), float64(rate.LastHourMultis))
	ind.timeIndicator.SetCurrentTime(rate.SinceLastQSO, rate.SinceLastQSOFormatted())
}

func (ind *rateIndicator) SetGoals(qsos int, points int, multis int) {
	ind.qAxis.SetGoal(float64(qsos))
	ind.pAxis.SetGoal(float64(points))
	ind.mAxis.SetGoal(float64(multis))
	ind.timeIndicator.SetGoal(float64(qsos))
}

func (ind *rateIndicator) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	ind.qAxis.PrepareGeometry(da, cr)
	ind.pAxis.PrepareGeometry(da, cr)
	ind.mAxis.PrepareGeometry(da, cr)

	ind.fillBackground(cr)

	cr.SetSourceRGBA(ind.style.lowZoneColor.WithAlpha(ind.style.areaAlpha))
	cr.MoveTo(ind.qAxis.goalPoint.x, ind.qAxis.goalPoint.y)
	cr.LineTo(ind.pAxis.goalPoint.x, ind.pAxis.goalPoint.y)
	cr.LineTo(ind.mAxis.goalPoint.x, ind.mAxis.goalPoint.y)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGBA(ind.style.lowZoneColor.WithAlpha(ind.style.borderAlpha))
	cr.MoveTo(ind.qAxis.goalPoint.x, ind.qAxis.goalPoint.y)
	cr.LineTo(ind.pAxis.goalPoint.x, ind.pAxis.goalPoint.y)
	cr.LineTo(ind.mAxis.goalPoint.x, ind.mAxis.goalPoint.y)
	cr.ClosePath()
	cr.Stroke()

	overallAchievment := (ind.qAxis.achievement + ind.pAxis.achievement + ind.mAxis.achievement) / 3
	cr.SetSourceRGBA(rateColors.ToRGBA(overallAchievment, ind.style.areaAlpha))
	cr.MoveTo(ind.qAxis.value1Point.x, ind.qAxis.value1Point.y)
	cr.LineTo(ind.pAxis.value1Point.x, ind.pAxis.value1Point.y)
	cr.LineTo(ind.mAxis.value1Point.x, ind.mAxis.value1Point.y)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGBA(rateColors.ToRGBA(overallAchievment, ind.style.borderAlpha))
	cr.MoveTo(ind.qAxis.value1Point.x, ind.qAxis.value1Point.y)
	cr.LineTo(ind.pAxis.value1Point.x, ind.pAxis.value1Point.y)
	cr.LineTo(ind.mAxis.value1Point.x, ind.mAxis.value1Point.y)
	cr.ClosePath()
	cr.Stroke()

	ind.qAxis.Draw(da, cr)
	ind.pAxis.Draw(da, cr)
	ind.mAxis.Draw(da, cr)
	ind.timeIndicator.Draw(da, cr)
}

func (ind *rateIndicator) fillBackground(cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(ind.style.backgroundColor.ToRGB())
	cr.Paint()
}

type rateAxis struct {
	style *rateStyle

	value1      float64
	value2      float64
	goalValue   float64
	maxValue    float64
	achievement float64
	unit        string

	angle  float64
	margin float64

	labelRect   rect
	axisLine    rect
	value1Point point
	value2Point point
	goalPoint   point
}

func newRateAxis(unit string, angle float64, style *rateStyle) *rateAxis {
	result := &rateAxis{
		style:     style,
		goalValue: 0,
		maxValue:  0,
		unit:      unit,
		angle:     degreesToRadians(angle) + angleRotation,
		margin:    style.axisMargin,
	}
	result.updateMaxValue()
	result.updateAchievement()
	return result
}

func (a *rateAxis) SetValues(value1, value2 float64) {
	a.value1 = value1
	a.value2 = value2
	a.updateAchievement()
}

func (a *rateAxis) updateAchievement() {
	if a.goalValue == 0 {
		a.achievement = 1
	} else {
		a.achievement = a.value1 / a.goalValue
	}
}

func (a *rateAxis) SetGoal(goal float64) {
	a.goalValue = goal
	a.updateMaxValue()
	a.updateAchievement()
}

func (a *rateAxis) updateMaxValue() {
	a.maxValue = 1.5 * a.goalValue
}

func (a *rateAxis) LabelText() string {
	return fmt.Sprintf("%2.0f %s", a.value1, a.unit)
}

func (a *rateAxis) PrepareGeometry(da *gtk.DrawingArea, cr *cairo.Context) {
	center := point{
		x: float64(da.GetAllocatedWidth()) / 2,
		y: float64(da.GetAllocatedHeight()) / 2,
	}

	axisLength := math.Min(float64(da.GetAllocatedWidth()), float64(da.GetAllocatedHeight()))/2 - a.margin

	end := polar{radius: axisLength, radians: a.angle}.toPoint().translate(center.x, center.y)
	a.axisLine = rect{
		top:    center.y,
		left:   center.x,
		bottom: end.y,
		right:  end.x,
	}

	if a.goalValue == 0 {
		a.value1Point = center
		a.value2Point = center
		a.goalPoint = center
	} else {
		a.value1Point = polar{
			radius:  math.Min((a.value1/a.maxValue)*axisLength, axisLength),
			radians: a.angle,
		}.toPoint().translate(center.x, center.y)

		a.value2Point = polar{
			radius:  math.Min((a.value2/a.maxValue)*axisLength, axisLength),
			radians: a.angle,
		}.toPoint().translate(center.x, center.y)

		a.goalPoint = polar{
			radius:  math.Min((a.goalValue/a.maxValue)*axisLength, axisLength),
			radians: a.angle,
		}.toPoint().translate(center.x, center.y)
	}

	cr.SetFontSize(a.style.fontSize)
	labelExtents := cr.TextExtents(a.LabelText())
	switch {
	case end.y < center.y:
		a.labelRect.top = (center.y - labelExtents.Height) / 2.0
		a.labelRect.left = a.goalPoint.x + 10.0
	case end.x < center.x:
		a.labelRect.top = center.y - labelExtents.Height/2.0
		a.labelRect.left = center.x - axisLength
	case end.x > center.x:
		a.labelRect.top = center.y - labelExtents.Height/2.0
		a.labelRect.left = center.x + axisLength - labelExtents.Width
	}
	a.labelRect.bottom = a.labelRect.top + labelExtents.Height
	a.labelRect.right = a.labelRect.left + labelExtents.Width
}

func (a *rateAxis) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(a.style.axisColor.ToRGB())
	cr.MoveTo(a.axisLine.left, a.axisLine.top)
	cr.LineTo(a.axisLine.right, a.axisLine.bottom)
	cr.Stroke()

	cr.SetSourceRGB(a.style.fontColor.ToRGB())
	cr.SetFontSize(a.style.fontSize)
	cr.MoveTo(a.labelRect.left, a.labelRect.bottom)
	cr.ShowText(a.LabelText())

	cr.SetSourceRGB(rateColors.ToRGB(a.achievement))
	cr.Arc(a.value1Point.x, a.value1Point.y, 5, 0, 2*math.Pi)
	cr.Fill()
	cr.Stroke()

	cr.SetLineWidth(5)
	cr.MoveTo(a.value2Point.x, a.value2Point.y)
	cr.LineTo(a.value1Point.x, a.value1Point.y)
	cr.Stroke()
}

type timeIndicator struct {
	style *rateStyle

	goalValue   float64
	currentTime time.Duration

	lineWidth float64

	goalSeconds float64
	achievement float64
	labelText   string
}

func newTimeIndicator(style *rateStyle) *timeIndicator {
	result := &timeIndicator{
		style:     style,
		goalValue: 0,
		lineWidth: style.timeIndicatorWidth,
	}
	result.updateGoalTime()
	result.updateAchievement()
	return result
}

func (ind *timeIndicator) SetGoal(goal float64) {
	ind.goalValue = goal
	ind.updateGoalTime()
}

func (ind *timeIndicator) updateGoalTime() {
	if ind.goalValue == 0 {
		ind.goalSeconds = 0
	} else {
		ind.goalSeconds = time.Hour.Seconds() / ind.goalValue
	}
}

func (ind *timeIndicator) SetCurrentTime(currentTime time.Duration, currentText string) {
	ind.currentTime = currentTime
	ind.labelText = currentText
	ind.updateAchievement()
}

func (ind *timeIndicator) updateAchievement() {
	if ind.goalSeconds == 0 {
		ind.achievement = 1
	} else {
		ind.achievement = 1 - math.Min(1, ind.currentTime.Seconds()/ind.goalSeconds)
	}
}

func (ind *timeIndicator) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	center := point{
		x: float64(da.GetAllocatedWidth()) / 2,
		y: float64(da.GetAllocatedHeight()) / 2,
	}
	radius := (math.Min(float64(da.GetAllocatedWidth()), float64(da.GetAllocatedHeight())) / 2) - (ind.lineWidth / 2)
	angle := (1 - ind.achievement) * 2 * math.Pi

	cr.SetSourceRGB(timeColors.ToRGB(ind.achievement))
	cr.SetLineWidth(ind.lineWidth)
	cr.Arc(center.x, center.y, radius, angleRotation, angle+angleRotation)
	cr.Stroke()

	cr.SetSourceRGB(ind.style.fontColor.ToRGB())
	cr.SetFontSize(ind.style.fontSize)
	labelExtents := cr.TextExtents(ind.labelText)
	label := point{
		x: center.x - labelExtents.Width/2.0,
		y: center.y + (radius+labelExtents.Height)/2.0,
	}
	cr.MoveTo(label.x, label.y)
	cr.ShowText(ind.labelText)
}
