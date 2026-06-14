package ui

import (
	"fmt"
	"strings"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

// Named stylesheet snippets. Single source of truth for strings that used to
// be duplicated across the ui/qt/ package.
const (
	CentralAreaStyle         = "QWidget#centralArea { background-color: palette(alternate-base); }"
	EntryDuplicateStyle      = "QWidget#entryWidget { background-color: palette(window); border: 4px solid red; }"
	EntryEditingStyle        = "QWidget#entryWidget { background-color: palette(window); border: 4px solid palette(accent); }"
	EntryNormalStyle         = "QWidget#entryWidget { background-color: palette(window); }"
	EntryFieldStyle          = "QLineEdit { font-size: 24pt; font-family: monospace; }"
	TXIndicatorActiveStyle   = "QLabel { color: red; font-weight: bold; }"
	TXIndicatorInactiveStyle = ""
	VFOActiveStyle           = "QLabel { padding-left: 5px; padding-right: 5px; background-color: palette(highlight); color: palette(highlighted-text); border-radius: 10px; }"
	VFOInactiveStyle         = "QLabel { padding-left: 6px; padding-right: 6px; }"

	QTCPhaseActiveStyle   = "font-weight: bold; color: #1a65b1;"
	QTCPhaseInactiveStyle = "font-weight: bold;"
	QTCFieldErrorStyle    = "color: #b00020; font-weight: bold;"

	BoldSectionStyle = "font-weight: bold;"
	MutedTextStyle   = "color: palette(mid);"
)

// Band RGBs transcribed from ui/style/contest.css. Do not change without
// changing the CSS too.
var bandRGB = map[core.Band][3]int{
	core.Band160m: {0x7F, 0x00, 0x7F},
	core.Band80m:  {0x00, 0x00, 0x7F},
	core.Band40m:  {0x00, 0xFF, 0x00},
	core.Band20m:  {0xFF, 0xFF, 0x00},
	core.Band15m:  {0xFF, 0x7F, 0x00},
	core.Band10m:  {0xFF, 0x00, 0x00},
}

var (
	defaultBandRGB   = [3]int{0x80, 0x80, 0x80}
	bandBGRGB        = [3]int{0x88, 0x88, 0x88}
	lowZoneRGB       = [3]int{0xCC, 0xCC, 0xCC}
	axisRGB          = [3]int{0x20, 0x20, 0x20}
	backgroundRGB    = [3]int{0xFF, 0xFF, 0xFF}
	timeIndicatorRGB = [3]int{0x3D, 0xAE, 0xE9}
)

// Spot type RGBs transcribed from ui/style/contest.css.
var spotTypeRGB = map[core.SpotType][3]int{
	core.WorkedSpot:  {0x80, 0x80, 0x80},
	core.ManualSpot:  {0xFF, 0xFF, 0xFF},
	core.SkimmerSpot: {0xFF, 0x99, 0xFF},
	core.RBNSpot:     {0xFF, 0xFF, 0x99},
	core.ClusterSpot: {0x99, 0xFF, 0xFF},
	core.UnknownSpot: {0xFF, 0xFF, 0xFF},
}

type Style struct {
	appStylesheet string

	spotBrushes map[core.SpotType]*qtlib.QBrush
	spotFgBrush *qtlib.QBrush
}

func NewStyle() *Style {
	s := &Style{
		appStylesheet: "QDockWidget::title { padding: 4px; }",
	}
	s.spotFgBrush = qtlib.NewQBrush3(qtlib.NewQColor11(0, 0, 0, 255))
	s.spotBrushes = make(map[core.SpotType]*qtlib.QBrush, len(spotTypeRGB))
	for t, rgb := range spotTypeRGB {
		s.spotBrushes[t] = qtlib.NewQBrush3(newQColor(rgb))
	}
	return s
}

func (s *Style) Apply(app *qtlib.QApplication) {
	app.SetStyleSheet(s.appStylesheet)
}

func newQColor(rgb [3]int) *qtlib.QColor {
	return qtlib.NewQColor11(rgb[0], rgb[1], rgb[2], 255)
}

func (s *Style) BandColor(band core.Band) *qtlib.QColor {
	if rgb, ok := bandRGB[band]; ok {
		return newQColor(rgb)
	}
	return newQColor(defaultBandRGB)
}

func (s *Style) BandBGColor() *qtlib.QColor        { return newQColor(bandBGRGB) }
func (s *Style) LowZoneColor() *qtlib.QColor       { return newQColor(lowZoneRGB) }
func (s *Style) AxisColor() *qtlib.QColor          { return newQColor(axisRGB) }
func (s *Style) BackgroundColor() *qtlib.QColor    { return newQColor(backgroundRGB) }
func (s *Style) TimeIndicatorColor() *qtlib.QColor { return newQColor(timeIndicatorRGB) }

func (s *Style) SpotBrush(src core.SpotType) *qtlib.QBrush {
	if b, ok := s.spotBrushes[src]; ok {
		return b
	}
	return s.spotBrushes[core.UnknownSpot]
}

func (s *Style) SpotForegroundBrush() *qtlib.QBrush {
	return s.spotFgBrush
}

// Package-level helpers used by score graph and score table. They delegate to
// the palette defined above so both views agree on band colors.
func bandQColor(band core.Band) *qtlib.QColor {
	if rgb, ok := bandRGB[band]; ok {
		return newQColor(rgb)
	}
	return newQColor(defaultBandRGB)
}

func bandBGQColor() *qtlib.QColor {
	return newQColor(bandBGRGB)
}

func rgbQColor(rgb [3]int, alpha float32) *qtlib.QColor {
	c := qtlib.NewQColor11(rgb[0], rgb[1], rgb[2], 255)
	c.SetAlphaF(alpha)
	return c
}

func GetSpotsBandButtonStyle(active bool, visible bool, max bool) string {
	properties := []string{}
	switch {
	case active:
		properties = append(properties, "background-color: palette(highlight); color: palette(highlighted-text);")
	case visible:
		properties = append(properties, "background-color: palette(accent); color: palette(light);")
	}

	if max {
		properties = append(properties, "font-weight: bold; border: 2px solid palette(highlight); ")
	}

	return fmt.Sprintf("QPushButton { %s }", strings.Join(properties, " "))
}
