package ui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

const (
	axisColorName    = "hellocontest-graph-axis"
	lowZoneColorName = "hellocontest-lowzone"
	defaultBandColor = "hellocontest-band-default"

	duplicateFGColorName = "hellocontest-duplicate-fg"
	duplicateBGColorName = "hellocontest-duplicate-bg"
	workedFGColorName    = "hellocontest-spot-fg"
	workedBGColorName    = "hellocontest-worked-spot-bg"
	worthlessFGColorName = "insensitive_fg_color"
	worthlessBGColorName = "insensitive_bg_color"
)

type fieldID string

func bandColor(colors colorProvider, band core.Band) style.Color {
	bandColorName := "hellocontest-band" + string(band)
	if !colors.HasColor(bandColorName) {
		return colors.ColorByName(defaultBandColor)
	}
	return colors.ColorByName(bandColorName)
}

type colorProvider interface {
	HasColor(name string) bool
	ColorByName(name string) style.Color
	BackgroundColor() style.Color
	ForegroundColor() style.Color
	TextColor() style.Color
}

func getUI(builder *gtk.Builder, name string) any {
	obj, err := builder.GetObject(name)
	if err != nil {
		log.Fatalf("Cannot get UI object %s: %v", name, err)
	}
	return obj
}
