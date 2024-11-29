//go:build !fyne

package ui

import (
	"math"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

type scoreGraphStyle struct {
	colorProvider

	backgroundColor style.Color
	axisColor       style.Color
	lowZoneColor    style.Color
	timeFrameColor  style.Color
	areaAlpha       float64
	borderAlpha     float64
}

func (s *scoreGraphStyle) Refresh() {
	s.backgroundColor = s.colorProvider.BackgroundColor()
	s.axisColor = s.colorProvider.ColorByName(axisColorName)
	s.lowZoneColor = s.colorProvider.ColorByName(lowZoneColorName)
	s.timeFrameColor = s.colorProvider.ColorByName(timeIndicatorColorName)
}

type scoreGraph struct {
	graphs         []core.BandGraph
	maxPoints      int
	maxMultis      int
	pointsGoal     int
	multisGoal     int
	timeFrameIndex int

	pointsBinGoal float64
	multisBinGoal float64

	style *scoreGraphStyle
}

const timeIndicatorColorName = "hellocontest-timeindicator"

func newScoreGraph(colors colorProvider) *scoreGraph {
	style := &scoreGraphStyle{
		colorProvider: colors,
		areaAlpha:     0.4,
		borderAlpha:   0.8,
	}
	style.Refresh()

	result := &scoreGraph{
		graphs:     nil,
		pointsGoal: 60,
		multisGoal: 60,
		style:      style,
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
	g.timeFrameIndex = g.graphs[0].Bindex(time.Now()) // TODO: use the central clock!!!
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
}

func (g *scoreGraph) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	// preparations

	valueCount := 0
	if len(g.graphs) > 0 {
		valueCount = len(g.graphs[0].DataPoints)
	}
	layout := g.calculateLayout(da, valueCount)

	// the background
	g.fillBackground(cr)

	// the zone
	cr.SetSourceRGBA(g.style.lowZoneColor.WithAlpha(g.style.areaAlpha))
	cr.MoveTo(0, layout.zeroY-layout.pointsLowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY-layout.pointsLowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY+layout.multisLowZoneHeight)
	cr.LineTo(0, layout.zeroY+layout.multisLowZoneHeight)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGBA(g.style.lowZoneColor.WithAlpha(g.style.borderAlpha))
	cr.MoveTo(0, layout.zeroY-layout.pointsLowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY-layout.pointsLowZoneHeight)
	cr.LineTo(layout.width, layout.zeroY+layout.multisLowZoneHeight)
	cr.LineTo(0, layout.zeroY+layout.multisLowZoneHeight)
	cr.ClosePath()
	cr.Stroke()

	// the graph
	for i := len(g.graphs) - 1; i >= 0; i-- {
		graph := g.graphs[i]
		color := bandColor(g.style, graph.Band)
		cr.SetSourceRGB(color.ToRGB())

		g.drawDataPointsRectangular(cr, layout, graph.DataPoints)
	}

	// the time frame
	if g.timeFrameIndex >= 0 && valueCount > 1 {
		startX := float64(g.timeFrameIndex) * layout.binWidth
		endX := float64(g.timeFrameIndex+1) * layout.binWidth
		cr.SetSourceRGBA(g.style.timeFrameColor.ToRGBA())
		cr.MoveTo(startX, layout.zeroY-layout.maxHeight)
		cr.LineTo(endX, layout.zeroY-layout.maxHeight)
		cr.LineTo(endX, layout.zeroY+layout.maxHeight)
		cr.LineTo(startX, layout.zeroY+layout.maxHeight)
		cr.ClosePath()
		cr.Stroke()
	}

	// the zero line
	cr.SetSourceRGB(g.style.axisColor.ToRGB())
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
	result.pointsLowZoneHeight = math.Min(result.maxHeight/2.0, (result.maxHeight/float64(g.maxPoints))*g.pointsBinGoal)
	result.multisLowZoneHeight = math.Min(result.maxHeight/2.0, (result.maxHeight/float64(g.maxMultis))*g.multisBinGoal)
	if valueCount > 0 {
		result.binWidth = result.width / float64(valueCount)
	} else {
		result.binWidth = result.width
	}

	return result
}

func (g *scoreGraph) fillBackground(cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(g.style.backgroundColor.ToRGB())
	cr.Paint()
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
	for i := 0; i < valueCount; i++ {
		startX := float64(i) * layout.binWidth
		endX := float64(i+1) * layout.binWidth
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
		startX := float64(i+1) * layout.binWidth
		endX := float64(i) * layout.binWidth
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

	if g.multisBinGoal > 0 {
		valueScaling = layout.multisLowZoneHeight / g.multisBinGoal
	} else {
		valueScaling = 0
	}
	valueScaling = layout.multisLowZoneHeight / g.multisBinGoal
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
