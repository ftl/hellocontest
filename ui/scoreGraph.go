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

func (g *scoreGraph) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	// preparations

	// TODO extract graph to separate type and use parameters for width, height, marginY
	width := float64(da.GetAllocatedWidth())
	height := float64(da.GetAllocatedHeight())
	marginY := 5.0

	zeroY := height / 2
	maxHeight := zeroY - marginY
	lowZoneHeight := maxHeight / 2

	valueCount := len(g.graph.DataPoints)
	binWidth := width / float64(valueCount)

	// the background
	g.fillBackground(cr)

	// the zone
	cr.SetSourceRGBA(rateStyle.lowZoneColor.toRGBA(rateStyle.areaAlpha))
	cr.MoveTo(0, zeroY-lowZoneHeight)
	cr.LineTo(width, zeroY-lowZoneHeight)
	cr.LineTo(width, zeroY+lowZoneHeight)
	cr.LineTo(0, zeroY+lowZoneHeight)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGBA(rateStyle.lowZoneColor.toRGBA(rateStyle.borderAlpha))
	cr.MoveTo(0, zeroY-lowZoneHeight)
	cr.LineTo(width, zeroY-lowZoneHeight)
	cr.LineTo(width, zeroY+lowZoneHeight)
	cr.LineTo(0, zeroY+lowZoneHeight)
	cr.ClosePath()
	cr.Stroke()

	// the graph

	// cr.SetSourceRGBA(0, 0, 1, 0.6) // TODO calculate the overall achievement and use the corresponding color
	cr.SetSourceRGB(rateStyle.scoreGraphColor.toRGB()) // TODO calculate the overall achievement and use the corresponding color
	cr.MoveTo(0, zeroY)

	valueScaling := lowZoneHeight / g.pointsBinGoal
	lastY := zeroY
	for i := 0; i < valueCount; i++ {
		startX := float64(i) * binWidth
		centerX := startX + binWidth/2.0
		endX := float64(i+1) * binWidth
		value := float64(g.graph.DataPoints[i].Points)
		y := zeroY - math.Min(value*valueScaling, maxHeight)
		if i == 0 {
			cr.LineTo(startX, y)
			cr.LineTo(centerX, y)
		} else {
			cr.CurveTo(startX, lastY, startX, y, centerX, y)
		}
		if i == valueCount-1 {
			cr.LineTo(endX, y)
			cr.LineTo(endX, zeroY)
		}
		lastY = y
	}

	valueScaling = lowZoneHeight / g.multisBinGoal
	lastY = zeroY
	for i := valueCount - 1; i >= 0; i-- {
		startX := float64(i+1) * binWidth
		centerX := startX - binWidth/2.0
		endX := float64(i) * binWidth
		value := float64(g.graph.DataPoints[i].Multis)
		y := zeroY + math.Min(value*valueScaling, maxHeight)
		if i == valueCount-1 {
			cr.LineTo(startX, y)
			cr.LineTo(centerX, y)
		} else {
			cr.CurveTo(startX, lastY, startX, y, centerX, y)
		}
		if i == 0 {
			cr.LineTo(endX, y)
			cr.LineTo(endX, zeroY)
		}
		lastY = y
	}
	cr.ClosePath()
	cr.Fill()

	// the time frame
	if g.timeFrameIndex >= 0 && valueCount > 1 {
		startX := float64(g.timeFrameIndex) * binWidth
		endX := float64(g.timeFrameIndex+1) * binWidth
		cr.SetSourceRGBA(rateStyle.timeFrameColor.toRGBA(rateStyle.timeFrameAlpha)) // TODO calculate the achievment of the current time frame and use the corresponding color
		cr.MoveTo(startX, zeroY-maxHeight)
		cr.LineTo(endX, zeroY-maxHeight)
		cr.LineTo(endX, zeroY+maxHeight)
		cr.LineTo(startX, zeroY+maxHeight)
		cr.ClosePath()
		cr.Stroke()
	}

	// the zero line
	cr.SetSourceRGB(rateStyle.axisColor.toRGB())
	cr.MoveTo(0, zeroY)
	cr.LineTo(width, zeroY)
	cr.Stroke()
}

func (g *scoreGraph) fillBackground(cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(rateStyle.backgroundColor.toRGB())
	cr.Paint()
}
