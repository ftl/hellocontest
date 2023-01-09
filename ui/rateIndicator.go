package ui

import (
	"fmt"
	"math"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

var rateColors = colorMap{
	{1, 0, 0}, {1, 0.6, 0.2}, {0, 0.8, 0},
}

var timeColors = colorMap{
	{1, 0, 0}, {1, 0.6, 0.2}, {0, 0.8, 0}, {0, 0.8, 0}, {0, 0.8, 0}, {0, 0.8, 0},
}

const angleRotation = (3.0 / 2.0) * math.Pi

var rateStyle = struct {
	fontColor          color
	fontSize           float64
	axisColor          color
	axisMargin         float64
	timeIndicatorWidth float64
}{
	fontColor:          color{0.4, 0.4, 0.4},
	fontSize:           15,
	axisColor:          color{0.4, 0.4, 0.4},
	axisMargin:         15,
	timeIndicatorWidth: 10,
}

type rateIndicator struct {
	qAxis         *rateAxis
	pAxis         *rateAxis
	mAxis         *rateAxis
	timeIndicator *timeIndicator
}

func newRateIndicator() *rateIndicator {
	// TODO add parameters
	qTarget := 48.0
	pTarget := 60.0
	mTarget := 24.0
	return &rateIndicator{
		qAxis:         newRateAxis(qTarget, "Q/h", 0, rateStyle.axisMargin),
		pAxis:         newRateAxis(pTarget, "P/h", 120, rateStyle.axisMargin),
		mAxis:         newRateAxis(mTarget, "M/h", 240, rateStyle.axisMargin),
		timeIndicator: newTimeIndicator(qTarget, rateStyle.timeIndicatorWidth),
	}
}

func (ind *rateIndicator) SetRate(rate core.QSORate) {
	ind.qAxis.SetValues(float64(rate.Last5MinRate), float64(rate.LastHourRate))
	ind.pAxis.SetValues(float64(rate.Last5MinPoints), float64(rate.LastHourPoints))
	ind.mAxis.SetValues(float64(rate.Last5MinMultis), float64(rate.LastHourMultis))
	ind.timeIndicator.SetCurrentTime(rate.SinceLastQSO, rate.SinceLastQSOFormatted())
}

func (ind *rateIndicator) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	ind.qAxis.PrepareGeometry(da, cr)
	ind.pAxis.PrepareGeometry(da, cr)
	ind.mAxis.PrepareGeometry(da, cr)

	ind.fillBackground(cr)

	cr.SetSourceRGBA(0.8, 0.8, 0.8, 0.4)
	cr.MoveTo(ind.qAxis.targetPoint.x, ind.qAxis.targetPoint.y)
	cr.LineTo(ind.pAxis.targetPoint.x, ind.pAxis.targetPoint.y)
	cr.LineTo(ind.mAxis.targetPoint.x, ind.mAxis.targetPoint.y)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGBA(0.8, 0.8, 0.8, 0.8)
	cr.MoveTo(ind.qAxis.targetPoint.x, ind.qAxis.targetPoint.y)
	cr.LineTo(ind.pAxis.targetPoint.x, ind.pAxis.targetPoint.y)
	cr.LineTo(ind.mAxis.targetPoint.x, ind.mAxis.targetPoint.y)
	cr.ClosePath()
	cr.Stroke()

	overallAchievment := (ind.qAxis.achievement + ind.pAxis.achievement + ind.mAxis.achievement) / 3
	cr.SetSourceRGBA(rateColors.toRGBA(overallAchievment, 0.4))
	cr.MoveTo(ind.qAxis.value1Point.x, ind.qAxis.value1Point.y)
	cr.LineTo(ind.pAxis.value1Point.x, ind.pAxis.value1Point.y)
	cr.LineTo(ind.mAxis.value1Point.x, ind.mAxis.value1Point.y)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGBA(rateColors.toRGBA(overallAchievment, 0.8))
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

	cr.SetSourceRGB(1, 1, 1)
	cr.Paint()
}

type rateAxis struct {
	value1      float64
	value2      float64
	targetValue float64
	maxValue    float64
	achievement float64
	unit        string

	angle  float64
	margin float64

	labelRect   rect
	axisLine    rect
	value1Point point
	value2Point point
	targetPoint point
}

func newRateAxis(target float64, unit string, angle float64, margin float64) *rateAxis {
	return &rateAxis{
		targetValue: target,
		maxValue:    1.5 * target,
		unit:        unit,
		angle:       degreesToRadians(angle) + angleRotation,
		margin:      margin,
	}
}

func (a *rateAxis) SetValues(value1, value2 float64) {
	a.value1 = value1
	a.value2 = value2
	a.updateAchievement()
}

func (a *rateAxis) updateAchievement() {
	if a.targetValue == 0 {
		a.achievement = 0
	} else {
		a.achievement = a.value1 / a.targetValue
	}
}

func (a *rateAxis) SetTarget(target float64) {
	a.targetValue = target
	a.updateAchievement()
}

func (a *rateAxis) SetMax(max float64) {
	a.maxValue = max
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

	a.value1Point = polar{
		radius:  math.Min((a.value1/a.maxValue)*axisLength, axisLength),
		radians: a.angle,
	}.toPoint().translate(center.x, center.y)

	a.value2Point = polar{
		radius:  math.Min((a.value2/a.maxValue)*axisLength, axisLength),
		radians: a.angle,
	}.toPoint().translate(center.x, center.y)

	a.targetPoint = polar{
		radius:  math.Min((a.targetValue/a.maxValue)*axisLength, axisLength),
		radians: a.angle,
	}.toPoint().translate(center.x, center.y)

	cr.SetFontSize(rateStyle.fontSize)
	labelExtents := cr.TextExtents(a.LabelText())
	switch {
	case end.y < center.y:
		a.labelRect.top = a.targetPoint.y - labelExtents.Height/2.0
		a.labelRect.left = a.targetPoint.x + 10.0
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

	cr.SetSourceRGB(rateStyle.axisColor.toRGB())
	cr.MoveTo(a.axisLine.left, a.axisLine.top)
	cr.LineTo(a.axisLine.right, a.axisLine.bottom)
	cr.Stroke()

	cr.SetSourceRGB(rateStyle.fontColor.toRGB())
	cr.SetFontSize(rateStyle.fontSize)
	cr.MoveTo(a.labelRect.left, a.labelRect.bottom)
	cr.ShowText(a.LabelText())

	cr.SetSourceRGB(rateColors.toRGB(a.achievement))
	cr.Arc(a.value1Point.x, a.value1Point.y, 5, 0, 2*math.Pi)
	cr.Fill()
	cr.Stroke()

	cr.SetLineWidth(5)
	cr.MoveTo(a.value2Point.x, a.value2Point.y)
	cr.LineTo(a.value1Point.x, a.value1Point.y)
	cr.Stroke()
}

type timeIndicator struct {
	targetValue float64
	currentTime time.Duration

	lineWidth float64

	targetSeconds float64
	achievement   float64
	labelText     string
}

func newTimeIndicator(target float64, lineWidth float64) *timeIndicator {
	result := &timeIndicator{
		targetValue: target,
		lineWidth:   lineWidth,
	}
	result.updateTargetTime()
	result.updateAchievement()
	return result
}

func (ind *timeIndicator) SetTarget(target float64) {
	ind.targetValue = target
	ind.updateTargetTime()
}

func (ind *timeIndicator) updateTargetTime() {
	if ind.targetValue == 0 {
		ind.targetSeconds = 0
	} else {
		ind.targetSeconds = time.Hour.Seconds() / ind.targetValue
	}
}

func (ind *timeIndicator) SetCurrentTime(currentTime time.Duration, currentText string) {
	ind.currentTime = currentTime
	ind.labelText = currentText
	ind.updateAchievement()
}

func (ind *timeIndicator) updateAchievement() {
	if ind.targetValue == 0 {
		ind.achievement = 0
	} else {
		ind.achievement = 1 - math.Min(1, ind.currentTime.Seconds()/ind.targetSeconds)
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

	cr.SetSourceRGB(timeColors.toRGB(ind.achievement))
	cr.SetLineWidth(ind.lineWidth)
	cr.Arc(center.x, center.y, radius, angleRotation, angle+angleRotation)
	cr.Stroke()

	cr.SetSourceRGB(rateStyle.fontColor.toRGB())
	cr.SetFontSize(rateStyle.fontSize)
	labelExtents := cr.TextExtents(ind.labelText)
	label := point{
		x: center.x - labelExtents.Width/2.0,
		y: center.y + (radius+labelExtents.Height)/2.0,
	}
	cr.MoveTo(label.x, label.y)
	cr.ShowText(ind.labelText)
}
