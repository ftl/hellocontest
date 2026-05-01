package ui

import (
	"fmt"
	"math"
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

const (
	scoreGraphFontSize    = 15
	scoreGraphAreaAlpha   = 0.4
	scoreGraphBorderAlpha = 0.8
	scoreGraphMarginY     = 10.0
	scoreGraphAxisWidth   = 1.0
	scoreGraphDivWidth    = 0.5
	scoreGraphDivCount    = 8
	scoreGraphDefaultGoal = 60
	scoreGraphMinWidth    = 320
	scoreGraphMinHeight   = 160
)

type scoreGraph struct {
	widget *qtlib.QWidget

	clock          core.Clock
	graphs         []core.BandGraph
	maxPoints      int
	maxMultis      int
	pointsGoal     int
	multisGoal     int
	pointsBinGoal  float64
	multisBinGoal  float64
	timeFrameIndex int
}

func newScoreGraph(clock core.Clock) *scoreGraph {
	g := &scoreGraph{
		clock:      clock,
		pointsGoal: scoreGraphDefaultGoal,
		multisGoal: scoreGraphDefaultGoal,
	}
	g.widget = qtlib.NewQWidget2()
	g.widget.SetMinimumWidth(scoreGraphMinWidth)
	g.widget.SetMinimumHeight(scoreGraphMinHeight)
	g.updateBinGoals()

	g.widget.OnPaintEvent(func(super func(ev *qtlib.QPaintEvent), ev *qtlib.QPaintEvent) {
		super(ev)
		g.paint()
	})
	return g
}

func (g *scoreGraph) SetGraphs(graphs []core.BandGraph) {
	g.graphs = graphs
	g.maxPoints = 0
	g.maxMultis = 0
	for _, graph := range graphs {
		if g.maxPoints < graph.Max.Points {
			g.maxPoints = graph.Max.Points
		}
		if g.maxMultis < graph.Max.Multis {
			g.maxMultis = graph.Max.Multis
		}
	}
	g.updateBinGoals()
	g.UpdateTimeFrame()
}

func (g *scoreGraph) SetGoals(points, multis int) {
	g.pointsGoal = points
	g.multisGoal = multis
	g.updateBinGoals()
}

func (g *scoreGraph) updateBinGoals() {
	if len(g.graphs) == 0 {
		g.pointsBinGoal = float64(g.pointsGoal)
		g.multisBinGoal = float64(g.multisGoal)
		return
	}
	g.pointsBinGoal = g.graphs[0].ScaleHourlyGoalToBin(g.pointsGoal)
	g.multisBinGoal = g.graphs[0].ScaleHourlyGoalToBin(g.multisGoal)
}

func (g *scoreGraph) UpdateTimeFrame() {
	if len(g.graphs) == 0 {
		g.timeFrameIndex = -1
		return
	}
	g.timeFrameIndex = g.graphs[0].Bindex(g.clock.Now())
}

type graphLayout struct {
	width               float64
	height              float64
	marginY             float64
	zeroY               float64
	maxHeight           float64
	pointsLowZoneHeight float64
	multisLowZoneHeight float64
	binWidth            float64
	divX                []float64
	axisLineWidth       float64
	divisionLineWidth   float64
	leftLegendWidth     float64
	timeIndicatorHeight float64
}

func (g *scoreGraph) calculateLayout(painter *qtlib.QPainter, valueCount int) graphLayout {
	result := graphLayout{
		width:             float64(g.widget.Width()),
		height:            float64(g.widget.Height()),
		marginY:           scoreGraphMarginY,
		axisLineWidth:     scoreGraphAxisWidth,
		divisionLineWidth: scoreGraphDivWidth,
	}

	fm := painter.FontMetrics()
	result.leftLegendWidth = float64(fm.HorizontalAdvance("00:00")) + 2.0
	result.timeIndicatorHeight = float64(fm.Height()) + 2.0
	graphWidth := result.width - result.leftLegendWidth

	result.zeroY = (result.height - result.timeIndicatorHeight) / 2.0
	result.maxHeight = result.zeroY - result.marginY
	if g.maxPoints > 0 {
		result.pointsLowZoneHeight = math.Min(result.maxHeight/2.0, (result.maxHeight/float64(g.maxPoints))*g.pointsBinGoal)
	} else {
		result.pointsLowZoneHeight = result.maxHeight / 2.0
	}
	if g.maxMultis > 0 {
		result.multisLowZoneHeight = math.Min(result.maxHeight/2.0, (result.maxHeight/float64(g.maxMultis))*g.multisBinGoal)
	} else {
		result.multisLowZoneHeight = result.maxHeight / 2.0
	}
	if valueCount > 0 {
		result.binWidth = graphWidth / float64(valueCount)
	} else {
		result.binWidth = graphWidth
	}

	divCount := scoreGraphDivCount
	result.divX = make([]float64, divCount-1)
	divWidth := graphWidth / float64(divCount)
	for i := range result.divX {
		result.divX[i] = float64(i+1) * divWidth
	}

	return result
}

func (g *scoreGraph) paint() {
	painter := qtlib.NewQPainter2(g.widget.QPaintDevice)
	defer painter.End()
	painter.SetRenderHint(qtlib.QPainter__Antialiasing)

	// Get colors from widget palette for theme support
	palette := qtlib.QGuiApplication_Palette()
	bgColor := palette.ColorWithCr(qtlib.QPalette__Window)
	fgColor := palette.ColorWithCr(qtlib.QPalette__WindowText)
	selectionColor := palette.ColorWithCr(qtlib.QPalette__Highlight)

	bgRGB := [3]int{bgColor.Red(), bgColor.Green(), bgColor.Blue()}
	fgRGB := [3]int{fgColor.Red(), fgColor.Green(), fgColor.Blue()}
	selectionRGB := [3]int{selectionColor.Red(), selectionColor.Green(), selectionColor.Blue()}

	// Derive low zone color as mid-tone between bg and fg
	lowZoneRGB := [3]int{
		(bgRGB[0] + fgRGB[0]) / 2,
		(bgRGB[1] + fgRGB[1]) / 2,
		(bgRGB[2] + fgRGB[2]) / 2,
	}

	font := qtlib.NewQFont()
	font.SetPixelSize(scoreGraphFontSize)
	painter.SetFont(font)

	valueCount := 0
	if len(g.graphs) > 0 {
		valueCount = len(g.graphs[0].DataPoints)
	}
	layout := g.calculateLayout(painter, valueCount)

	g.fillBackground(painter, layout, bgRGB)
	g.drawLowZone(painter, layout, lowZoneRGB, fgRGB)

	for i := len(g.graphs) - 1; i >= 0; i-- {
		g.drawBand(painter, layout, g.graphs[i])
	}

	g.drawTimeDivisions(painter, layout, fgRGB)
	g.drawTimeIndicator(painter, layout, fgRGB, selectionRGB)
	g.drawZeroLine(painter, layout, fgRGB)
}

func (g *scoreGraph) fillBackground(painter *qtlib.QPainter, layout graphLayout, bgRGB [3]int) {
	brush := qtlib.NewQBrush3(qtlib.NewQColor11(bgRGB[0], bgRGB[1], bgRGB[2], 255))
	painter.FillRect2(0, 0, int(layout.width), int(layout.height), brush)
}

func (g *scoreGraph) drawZeroLine(painter *qtlib.QPainter, layout graphLayout, fgRGB [3]int) {
	pen := qtlib.NewQPen()
	pen.SetColor(qtlib.NewQColor11(fgRGB[0], fgRGB[1], fgRGB[2], 255))
	pen.SetWidthF(layout.axisLineWidth)
	painter.SetPenWithPen(pen)

	painter.DrawLine(qtlib.NewQLineF3(layout.leftLegendWidth, layout.zeroY, layout.width, layout.zeroY))

	g.drawYLegendAt(painter, layout, layout.zeroY, "0", fgRGB)
}

func (g *scoreGraph) drawYLegendAt(painter *qtlib.QPainter, layout graphLayout, y float64, text string, fgRGB [3]int) {
	fm := painter.FontMetrics()
	textWidth := float64(fm.HorizontalAdvance(text))
	textHeight := float64(fm.Height())
	left := layout.leftLegendWidth - textWidth - 2.0
	bottom := y + textHeight/2.0

	pen := qtlib.NewQPen()
	pen.SetColor(qtlib.NewQColor11(fgRGB[0], fgRGB[1], fgRGB[2], 255))
	painter.SetPenWithPen(pen)
	painter.DrawText(qtlib.NewQPointF3(left, bottom), text)
}

func (g *scoreGraph) drawTimeDivisions(painter *qtlib.QPainter, layout graphLayout, fgRGB [3]int) {
	pen := qtlib.NewQPen()
	pen.SetColor(qtlib.NewQColor11(fgRGB[0], fgRGB[1], fgRGB[2], 255))
	pen.SetWidthF(layout.divisionLineWidth)
	painter.SetPenWithPen(pen)

	painter.DrawLine(qtlib.NewQLineF3(
		layout.leftLegendWidth, layout.zeroY-layout.maxHeight,
		layout.leftLegendWidth, layout.zeroY+layout.maxHeight,
	))

	for _, x := range layout.divX {
		painter.DrawLine(qtlib.NewQLineF3(
			x+layout.leftLegendWidth, layout.zeroY-layout.maxHeight,
			x+layout.leftLegendWidth, layout.zeroY+layout.maxHeight,
		))
	}
}

func (g *scoreGraph) drawTimeIndicator(painter *qtlib.QPainter, layout graphLayout, fgRGB [3]int, selectionRGB [3]int) {
	now := g.clock.Now()

	var elapsedTime time.Duration
	var elapsedTimePercent float64
	if g.timeFrameIndex >= 0 && len(g.graphs) > 0 {
		elapsedTime = g.graphs[0].ElapsedTime(now)
		elapsedTimePercent = g.graphs[0].ElapsedTimePercent(now)
	}

	left := layout.leftLegendWidth
	right := left + (layout.width-left)*elapsedTimePercent
	bottom := layout.height - layout.marginY
	top := bottom - layout.timeIndicatorHeight

	if right > left {
		indicatorColor := qtlib.NewQColor11(selectionRGB[0], selectionRGB[1], selectionRGB[2], 255)
		indicatorColor.SetAlphaF(scoreGraphBorderAlpha)
		brush := qtlib.NewQBrush3(indicatorColor)
		painter.FillRect2(int(left), int(top), int(right-left), int(bottom-top), brush)
	}

	pen := qtlib.NewQPen()
	pen.SetColor(qtlib.NewQColor11(fgRGB[0], fgRGB[1], fgRGB[2], 255))
	painter.SetPenWithPen(pen)

	elapsedTimeText := formatQtDuration(elapsedTime)
	fm := painter.FontMetrics()
	textHeight := float64(fm.Ascent())

	// Center text vertically in the progress bar
	centerY := layout.height - 1.0 - textHeight
	painter.DrawText(qtlib.NewQPointF3(1, centerY), elapsedTimeText)

	for i, x := range layout.divX {
		if i%2 == 1 && len(g.graphs) > 0 {
			percent := float64(i+1) / float64(len(layout.divX)+1)
			text := formatQtDuration(g.graphs[0].PercentAsDuration(percent))
			textWidth := float64(fm.HorizontalAdvance(text))
			painter.DrawText(qtlib.NewQPointF3(
				x+layout.leftLegendWidth-textWidth/2.0,
				centerY,
			), text)
		}
	}

	if g.timeFrameIndex >= 0 {
		startX := float64(g.timeFrameIndex)*layout.binWidth + layout.leftLegendWidth
		endX := float64(g.timeFrameIndex+1)*layout.binWidth + layout.leftLegendWidth

		framePen := qtlib.NewQPen()
		frameColor := qtlib.NewQColor11(selectionRGB[0], selectionRGB[1], selectionRGB[2], 255)
		frameColor.SetAlphaF(scoreGraphBorderAlpha)
		framePen.SetColor(frameColor)
		framePen.SetWidthF(layout.divisionLineWidth)
		painter.SetPenWithPen(framePen)
		painter.SetBrush(qtlib.NewQBrush())

		path := qtlib.NewQPainterPath()
		path.MoveTo2(startX, layout.zeroY-layout.maxHeight)
		path.LineTo2(endX, layout.zeroY-layout.maxHeight)
		path.LineTo2(endX, layout.zeroY+layout.maxHeight)
		path.LineTo2(startX, layout.zeroY+layout.maxHeight)
		path.CloseSubpath()
		painter.DrawPath(path)
	}
}

func (g *scoreGraph) drawLowZone(painter *qtlib.QPainter, layout graphLayout, lowZoneRGB [3]int, fgRGB [3]int) {
	fillColor := qtlib.NewQColor11(lowZoneRGB[0], lowZoneRGB[1], lowZoneRGB[2], 255)
	fillColor.SetAlphaF(scoreGraphAreaAlpha)
	fillBrush := qtlib.NewQBrush3(fillColor)

	top := layout.zeroY - layout.pointsLowZoneHeight
	bottom := layout.zeroY + layout.multisLowZoneHeight
	painter.FillRect2(int(layout.leftLegendWidth), int(top), int(layout.width-layout.leftLegendWidth), int(bottom-top), fillBrush)

	borderColor := qtlib.NewQColor11(lowZoneRGB[0], lowZoneRGB[1], lowZoneRGB[2], 255)
	borderColor.SetAlphaF(scoreGraphBorderAlpha)
	pen := qtlib.NewQPen()
	pen.SetColor(borderColor)
	pen.SetWidthF(layout.divisionLineWidth)
	painter.SetPenWithPen(pen)
	painter.SetBrush(qtlib.NewQBrush())

	path := qtlib.NewQPainterPath()
	path.MoveTo2(layout.leftLegendWidth, top)
	path.LineTo2(layout.width, top)
	path.LineTo2(layout.width, bottom)
	path.LineTo2(layout.leftLegendWidth, bottom)
	path.CloseSubpath()
	painter.DrawPath(path)

	g.drawYLegendAt(painter, layout, layout.zeroY-layout.pointsLowZoneHeight, fmt.Sprintf("%dP", g.pointsGoal), fgRGB)
	g.drawYLegendAt(painter, layout, layout.zeroY+layout.multisLowZoneHeight, fmt.Sprintf("%dM", g.multisGoal), fgRGB)
}

func (g *scoreGraph) drawBand(painter *qtlib.QPainter, layout graphLayout, graph core.BandGraph) {
	datapoints := graph.DataPoints
	valueCount := len(datapoints)
	if valueCount == 0 {
		return
	}

	fill := bandQColor(graph.Band)
	fill.SetAlphaF(scoreGraphAreaAlpha)
	brush := qtlib.NewQBrush3(fill)

	path := qtlib.NewQPainterPath()
	path.MoveTo2(layout.leftLegendWidth, layout.zeroY)

	var valueScaling float64
	if g.pointsBinGoal > 0 {
		valueScaling = layout.pointsLowZoneHeight / g.pointsBinGoal
	}
	for i := 0; i < valueCount; i++ {
		startX := float64(i)*layout.binWidth + layout.leftLegendWidth
		endX := float64(i+1)*layout.binWidth + layout.leftLegendWidth
		value := float64(datapoints[i].Points)
		y := layout.zeroY - math.Min(value*valueScaling, layout.maxHeight)
		path.LineTo2(startX, y)
		path.LineTo2(endX, y)
		if i == valueCount-1 {
			path.LineTo2(endX, layout.zeroY)
		}
	}

	if g.multisBinGoal > 0 {
		valueScaling = layout.multisLowZoneHeight / g.multisBinGoal
	} else {
		valueScaling = 0
	}
	for i := valueCount - 1; i >= 0; i-- {
		startX := float64(i+1)*layout.binWidth + layout.leftLegendWidth
		endX := float64(i)*layout.binWidth + layout.leftLegendWidth
		value := float64(datapoints[i].Multis)
		y := layout.zeroY + math.Min(value*valueScaling, layout.maxHeight)
		path.LineTo2(startX, y)
		path.LineTo2(endX, y)
		if i == valueCount-1 {
			path.LineTo2(endX, layout.zeroY)
		}
		if i == 0 {
			path.LineTo2(endX, layout.zeroY)
		}
	}
	path.CloseSubpath()

	painter.FillPath(path, brush)
}

func formatQtDuration(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(d.Hours()), int(d.Minutes())%60)
}
