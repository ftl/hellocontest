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

const useCurvedGraph = false

const timeIndicatorColorName = "hellocontest-timeindicator"

type scoreGraphStyle struct {
	colorProvider

	backgroundColor style.Color
	axisColor       style.Color
	lowZoneColor    style.Color
	timeFrameColor  style.Color
	areaAlpha       float64
	borderAlpha     float64

	fontSize float64
}

func (s *scoreGraphStyle) Refresh() {
	s.backgroundColor = s.colorProvider.BackgroundColor()
	s.axisColor = s.colorProvider.ColorByName(axisColorName)
	s.lowZoneColor = s.colorProvider.ColorByName(lowZoneColorName)
	s.timeFrameColor = s.colorProvider.ColorByName(timeIndicatorColorName)
}

type scoreGraph struct {
	clock          core.Clock
	graphs         []core.BandGraph
	maxPoints      int
	maxMultis      int
	pointsGoal     int
	multisGoal     int
	timeFrameIndex int

	pointsBinGoal float64
	multisBinGoal float64

	style          *scoreGraphStyle
	useCurvedGraph bool
}

func newScoreGraph(colors colorProvider, clock core.Clock) *scoreGraph {
	style := &scoreGraphStyle{
		colorProvider: colors,
		areaAlpha:     0.4,
		borderAlpha:   0.8,
		fontSize:      15,
	}
	style.Refresh()

	result := &scoreGraph{
		clock:          clock,
		graphs:         nil,
		pointsGoal:     60,
		multisGoal:     60,
		style:          style,
		useCurvedGraph: useCurvedGraph,
	}

	result.updateBinGoals()

	return result
}

func (g *scoreGraph) RefreshStyle() {
	g.style.Refresh()
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

func (g *scoreGraph) SetGoals(points int, multis int) {
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

func (g *scoreGraph) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	// preparations
	valueCount := 0
	if len(g.graphs) > 0 {
		valueCount = len(g.graphs[0].DataPoints)
	}
	layout := g.calculateLayout(da, cr, valueCount)

	g.fillBackground(cr)
	g.drawLowZone(cr, layout)

	// the graph
	for i := len(g.graphs) - 1; i >= 0; i-- {
		graph := g.graphs[i]
		color := bandColor(g.style, graph.Band)
		cr.SetSourceRGB(color.ToRGB())

		if useCurvedGraph {
			g.drawDataPointsCurved(cr, layout, graph.DataPoints)
		} else {
			g.drawDataPointsRectangular(cr, layout, graph.DataPoints)
		}
	}

	g.drawTimeDivisions(cr, layout)
	g.drawTimeIndicator(cr, layout)
	g.drawZeroLine(cr, layout)
}

func (g *scoreGraph) calculateLayout(da *gtk.DrawingArea, cr *cairo.Context, valueCount int) graphLayout {
	result := graphLayout{
		width:             float64(da.GetAllocatedWidth()),
		height:            float64(da.GetAllocatedHeight()),
		marginY:           10.0,
		axisLineWidth:     1.0,
		divisionLineWidth: .5,
	}

	cr.SetFontSize(g.style.fontSize)
	result.leftLegendWidth = cr.TextExtents("00:00").Width + 2.0
	result.timeIndicatorHeight = cr.TextExtents("Hg").Height + 2.0
	graphWidth := result.width - result.leftLegendWidth

	result.zeroY = (result.height - result.timeIndicatorHeight) / 2.0
	result.maxHeight = result.zeroY - result.marginY
	result.pointsLowZoneHeight = math.Min(result.maxHeight/2.0, (result.maxHeight/float64(g.maxPoints))*g.pointsBinGoal)
	result.multisLowZoneHeight = math.Min(result.maxHeight/2.0, (result.maxHeight/float64(g.maxMultis))*g.multisBinGoal)
	if valueCount > 0 {
		result.binWidth = graphWidth / float64(valueCount)
	} else {
		result.binWidth = graphWidth
	}

	const divCount = 8
	if len(result.divX) != divCount {
		result.divX = make([]float64, divCount-1)
	}
	divWidth := graphWidth / float64(divCount)
	for i := range result.divX {
		result.divX[i] = float64(i+1) * divWidth
	}

	return result
}

func (g *scoreGraph) fillBackground(cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(g.style.backgroundColor.ToRGB())
	cr.Paint()
}

func (g *scoreGraph) drawZeroLine(cr *cairo.Context, layout graphLayout) {
	// the line
	cr.SetSourceRGB(g.style.axisColor.ToRGB())
	cr.SetLineWidth(layout.axisLineWidth)
	cr.MoveTo(layout.leftLegendWidth, layout.zeroY)
	cr.LineTo(layout.width, layout.zeroY)
	cr.Stroke()

	// the legend
	g.drawYLegendAt(cr, layout, layout.zeroY, "0")
}

func (g *scoreGraph) drawYLegendAt(cr *cairo.Context, layout graphLayout, y float64, text string) {
	textExtents := cr.TextExtents(text)
	left := layout.leftLegendWidth - textExtents.Width - 2.0
	bottom := y + textExtents.Height/2.0
	cr.SetSourceRGB(g.style.axisColor.ToRGB())
	cr.SetFontSize(g.style.fontSize)
	cr.MoveTo(left, bottom)
	cr.ShowText(text)
}

func (g *scoreGraph) drawTimeDivisions(cr *cairo.Context, layout graphLayout) {
	cr.SetSourceRGB(g.style.axisColor.ToRGB())
	cr.SetLineWidth(layout.divisionLineWidth)
	cr.SetFontSize(g.style.fontSize)

	// the zero line
	cr.MoveTo(layout.leftLegendWidth, layout.zeroY-layout.maxHeight)
	cr.LineTo(layout.leftLegendWidth, layout.zeroY+layout.maxHeight)
	cr.Stroke()

	// the vertical divisions
	for _, x := range layout.divX {
		cr.MoveTo(x+layout.leftLegendWidth, layout.zeroY-layout.maxHeight)
		cr.LineTo(x+layout.leftLegendWidth, layout.zeroY+layout.maxHeight)
		cr.Stroke()
	}
}

func (g *scoreGraph) drawTimeIndicator(cr *cairo.Context, layout graphLayout) {
	now := g.clock.Now()

	var elapsedTime time.Duration
	var elapsedTimePercent float64
	if g.timeFrameIndex >= 0 && len(g.graphs) > 0 {
		elapsedTime = g.graphs[0].ElapsedTime(now)
		elapsedTimePercent = g.graphs[0].ElapsedTimePercent(now)
	} else {
		elapsedTime = 0
		elapsedTimePercent = 0.0
	}

	// the time bar
	left := layout.leftLegendWidth
	right := left + (layout.width-left)*elapsedTimePercent
	bottom := layout.height - layout.marginY
	top := bottom - layout.timeIndicatorHeight

	cr.SetSourceRGBA(g.style.timeFrameColor.ToRGBA())
	cr.MoveTo(left, top)
	cr.LineTo(right, top)
	cr.LineTo(right, bottom)
	cr.LineTo(left, bottom)
	cr.ClosePath()
	cr.Fill()

	// the elapsed time
	elapsedTimeText := formatDuration(elapsedTime)

	cr.SetSourceRGB(g.style.axisColor.ToRGB())
	cr.SetFontSize(g.style.fontSize)
	cr.MoveTo(1, layout.height-layout.marginY-1)
	cr.ShowText(elapsedTimeText)

	// the time legend
	for i, x := range layout.divX {
		if i%2 == 1 && len(g.graphs) > 0 {
			percent := float64(i+1) / float64(len(layout.divX)+1)
			text := formatDuration(g.graphs[0].PercentAsDuration(percent))
			textExtents := cr.TextExtents(text)
			cr.MoveTo(x+layout.leftLegendWidth-textExtents.Width/2.0, layout.zeroY+layout.maxHeight+textExtents.Height+2)
			cr.ShowText(text)
		}
	}

	// the old box
	if g.timeFrameIndex >= 0 {
		startX := float64(g.timeFrameIndex)*layout.binWidth + layout.leftLegendWidth
		endX := float64(g.timeFrameIndex+1)*layout.binWidth + layout.leftLegendWidth

		cr.SetSourceRGBA(g.style.timeFrameColor.ToRGBA())
		cr.SetLineWidth(layout.divisionLineWidth)
		cr.MoveTo(startX, layout.zeroY-layout.maxHeight)
		cr.LineTo(endX, layout.zeroY-layout.maxHeight)
		cr.LineTo(endX, layout.zeroY+layout.maxHeight)
		cr.LineTo(startX, layout.zeroY+layout.maxHeight)
		cr.ClosePath()
		cr.Stroke()
	}
}

func (g *scoreGraph) drawLowZone(cr *cairo.Context, layout graphLayout) {
	cr.SetSourceRGBA(g.style.lowZoneColor.WithAlpha(g.style.areaAlpha))
	cr.MoveTo(layout.leftLegendWidth, layout.zeroY-layout.pointsLowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY-layout.pointsLowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY+layout.multisLowZoneHeight)
	cr.LineTo(layout.leftLegendWidth, layout.zeroY+layout.multisLowZoneHeight)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGBA(g.style.lowZoneColor.WithAlpha(g.style.borderAlpha))
	cr.SetLineWidth(layout.divisionLineWidth)
	cr.MoveTo(layout.leftLegendWidth, layout.zeroY-layout.pointsLowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY-layout.pointsLowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY+layout.multisLowZoneHeight)
	cr.LineTo(layout.leftLegendWidth, layout.zeroY+layout.multisLowZoneHeight)
	cr.ClosePath()
	cr.Stroke()

	// the legend
	g.drawYLegendAt(cr, layout, layout.zeroY-layout.pointsLowZoneHeight, fmt.Sprintf("%d", int(g.pointsGoal)))
	g.drawYLegendAt(cr, layout, layout.zeroY+layout.multisLowZoneHeight, fmt.Sprintf("%d", int(g.multisGoal)))
}

func (g *scoreGraph) drawDataPointsRectangular(cr *cairo.Context, layout graphLayout, datapoints []core.BandScore) {
	valueCount := len(datapoints)

	cr.MoveTo(0, layout.zeroY)

	var valueScaling float64
	if g.pointsBinGoal > 0 {
		valueScaling = layout.pointsLowZoneHeight / g.pointsBinGoal
	} else {
		valueScaling = 0
	}
	for i := range valueCount {
		startX := float64(i)*layout.binWidth + layout.leftLegendWidth
		endX := float64(i+1)*layout.binWidth + layout.leftLegendWidth
		value := float64(datapoints[i].Points)
		y := layout.zeroY - math.Min(value*valueScaling, layout.maxHeight)
		cr.LineTo(startX, y)
		cr.LineTo(endX, y)
		if i == valueCount-1 {
			cr.LineTo(endX, layout.zeroY)
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
		cr.LineTo(startX, y)
		cr.LineTo(endX, y)
		if i == valueCount-1 {
			cr.LineTo(endX, layout.zeroY)
		}
		if i == 0 {
			cr.LineTo(endX, layout.zeroY)
		}
	}
	cr.ClosePath()
	cr.Fill()
}

func (g *scoreGraph) drawDataPointsCurved(cr *cairo.Context, layout graphLayout, datapoints []core.BandScore) {
	valueCount := len(datapoints)

	cr.MoveTo(0, layout.zeroY)

	var valueScaling float64
	if g.pointsBinGoal > 0 {
		valueScaling = layout.pointsLowZoneHeight / g.pointsBinGoal
	} else {
		valueScaling = 0
	}
	lastY := layout.zeroY
	for i := range valueCount {
		startX := float64(i)*layout.binWidth + layout.leftLegendWidth
		centerX := startX + layout.binWidth/2.0
		endX := float64(i+1)*layout.binWidth + layout.leftLegendWidth
		value := float64(datapoints[i].Points)
		y := layout.zeroY - math.Min(value*valueScaling, layout.maxHeight)
		if i == 0 {
			cr.LineTo(startX, y)
			cr.LineTo(centerX, y)
		} else {
			cr.CurveTo(startX, lastY, startX, y, centerX, y)
		}
		if i == valueCount-1 {
			cr.LineTo(endX, y)
			cr.LineTo(endX, layout.zeroY)
		}

		lastY = y
	}

	if g.multisBinGoal > 0 {
		valueScaling = layout.multisLowZoneHeight / g.multisBinGoal
	} else {
		valueScaling = 0
	}
	valueScaling = layout.multisLowZoneHeight / g.multisBinGoal
	lastY = layout.zeroY
	for i := valueCount - 1; i >= 0; i-- {
		startX := float64(i+1)*layout.binWidth + layout.leftLegendWidth
		centerX := startX - layout.binWidth/2.0
		endX := float64(i)*layout.binWidth + layout.leftLegendWidth
		value := float64(datapoints[i].Multis)
		y := layout.zeroY + math.Min(value*valueScaling, layout.maxHeight)
		if i == valueCount-1 {
			cr.LineTo(startX, y)
			cr.LineTo(centerX, y)
		} else {
			cr.CurveTo(startX, lastY, startX, y, centerX, y)
		}
		if i == 0 {
			cr.LineTo(endX, y)
			cr.LineTo(endX, layout.zeroY)
		}
		lastY = y
	}
	cr.ClosePath()
	cr.Fill()
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(d.Hours()), int(d.Minutes())%60)
}
