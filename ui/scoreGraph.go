package ui

import (
	"math"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

type scoreGraph struct {
	graph          core.BandGraph
	pointsGoal     int
	multisGoal     int
	timeFrameIndex int

	pointsBinGoal float64
	multisBinGoal float64
}

func newScoreGraph() *scoreGraph {
	result := &scoreGraph{
		graph:      core.BandGraph{},
		pointsGoal: 60,
		multisGoal: 60,
	}
	result.updateBinGoals()
	return result
}

func (g *scoreGraph) SetGraph(graph core.BandGraph) {
	g.graph = graph
	g.updateBinGoals()
	g.UpdateTimeFrame()
}

func (g *scoreGraph) SetGoals(points int, multis int) {
	g.pointsGoal = points
	g.multisGoal = multis
	g.updateBinGoals()
}

func (g *scoreGraph) updateBinGoals() {
	g.pointsBinGoal = g.graph.ScaleHourlyGoalToBin(g.pointsGoal)
	g.multisBinGoal = g.graph.ScaleHourlyGoalToBin(g.multisGoal)
}

func (g *scoreGraph) UpdateTimeFrame() {
	g.timeFrameIndex = g.graph.Bindex(time.Now())
}

type graphLayout struct {
	width         float64
	height        float64
	marginY       float64
	zeroY         float64
	maxHeight     float64
	lowZoneHeight float64
	binWidth      float64
}

func (g *scoreGraph) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	// preparations

	valueCount := len(g.graph.DataPoints)
	layout := g.calculateLayout(da, valueCount)

	// the background
	g.fillBackground(cr)

	// the zone
	cr.SetSourceRGBA(rateStyle.lowZoneColor.toRGBA(rateStyle.areaAlpha))
	cr.MoveTo(0, layout.zeroY-layout.lowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY-layout.lowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY+layout.lowZoneHeight)
	cr.LineTo(0, layout.zeroY+layout.lowZoneHeight)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGBA(rateStyle.lowZoneColor.toRGBA(rateStyle.borderAlpha))
	cr.MoveTo(0, layout.zeroY-layout.lowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY-layout.lowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY+layout.lowZoneHeight)
	cr.LineTo(0, layout.zeroY+layout.lowZoneHeight)
	cr.ClosePath()
	cr.Stroke()

	// the graph
	// TODO: draw the stacked band graph
	g.drawDataPoints(cr, layout, g.graph.DataPoints)

	// the time frame
	if g.timeFrameIndex >= 0 && valueCount > 1 {
		startX := float64(g.timeFrameIndex) * layout.binWidth
		endX := float64(g.timeFrameIndex+1) * layout.binWidth
		cr.SetSourceRGBA(rateStyle.timeFrameColor.toRGBA(rateStyle.timeFrameAlpha)) // TODO calculate the achievment of the current time frame and use the corresponding color
		cr.MoveTo(startX, layout.zeroY-layout.maxHeight)
		cr.LineTo(endX, layout.zeroY-layout.maxHeight)
		cr.LineTo(endX, layout.zeroY+layout.maxHeight)
		cr.LineTo(startX, layout.zeroY+layout.maxHeight)
		cr.ClosePath()
		cr.Stroke()
	}

	// the zero line
	cr.SetSourceRGB(rateStyle.axisColor.toRGB())
	cr.MoveTo(0, layout.zeroY)
	cr.LineTo(layout.width, layout.zeroY)
	cr.Stroke()
}

func (g *scoreGraph) calculateLayout(da *gtk.DrawingArea, valueCount int) graphLayout {
	result := graphLayout{
		width:   float64(da.GetAllocatedWidth()),
		height:  float64(da.GetAllocatedHeight()),
		marginY: 5.0,
	}

	result.zeroY = result.height / 2.0
	result.maxHeight = result.zeroY - result.marginY
	result.lowZoneHeight = result.maxHeight / 2.0
	result.binWidth = result.width / float64(valueCount)

	return result
}

func (g *scoreGraph) fillBackground(cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(rateStyle.backgroundColor.toRGB())
	cr.Paint()
}

func (g *scoreGraph) drawDataPoints(cr *cairo.Context, layout graphLayout, datapoints []core.BandScore) {
	valueCount := len(datapoints)

	cr.SetSourceRGB(rateStyle.scoreGraphColor.toRGB())
	cr.MoveTo(0, layout.zeroY)

	valueScaling := layout.lowZoneHeight / g.pointsBinGoal
	lastY := layout.zeroY
	for i := 0; i < valueCount; i++ {
		startX := float64(i) * layout.binWidth
		centerX := startX + layout.binWidth/2.0
		endX := float64(i+1) * layout.binWidth
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

	valueScaling = layout.lowZoneHeight / g.multisBinGoal
	lastY = layout.zeroY
	for i := valueCount - 1; i >= 0; i-- {
		startX := float64(i+1) * layout.binWidth
		centerX := startX - layout.binWidth/2.0
		endX := float64(i) * layout.binWidth
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
