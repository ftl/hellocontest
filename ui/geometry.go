//go:build !fyne

package ui

import "math"

type rect struct {
	top, left, bottom, right float64
}

func (r rect) width() float64 {
	return math.Abs(r.left - r.right)
}

func (r rect) height() float64 {
	return math.Abs(r.top - r.bottom)
}

func (r rect) contains(p point) bool {
	return r.left <= p.x && r.right >= p.x && r.top <= p.y && r.bottom >= p.y
}

func (r rect) toX(fraction float64) float64 {
	return r.left + r.width()*fraction
}

func (r rect) toY(fraction float64) float64 {
	return r.bottom - r.height()*fraction
}

func (r rect) translate(deltax, deltay float64) rect {
	return rect{
		top:    r.top + deltay,
		left:   r.left + deltax,
		bottom: r.bottom + deltay,
		right:  r.right + deltax,
	}
}

type point struct {
	x, y float64
}

func (p point) toPolar() polar {
	var radians float64
	if p.x == 0 {
		radians = 0
	} else {
		radians = math.Atan(p.y / p.x)
	}
	return polar{
		radius:  math.Sqrt((p.x * p.x) + (p.y * p.y)),
		radians: radians,
	}
}

func (p point) translate(deltaX, deltaY float64) point {
	return point{
		x: p.x + deltaX,
		y: p.y + deltaY,
	}
}

type polar struct {
	radius, radians float64
}

func (p polar) toPoint() point {
	return point{
		x: p.radius * math.Cos(p.radians),
		y: p.radius * math.Sin(p.radians),
	}
}

func radiansToDegrees(r float64) float64 {
	const halfcircle = 180.0 / math.Pi
	return r * halfcircle
}

func degreesToRadians(d float64) float64 {
	const halfcircle = 180 / math.Pi
	return d / halfcircle
}
