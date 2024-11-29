//go:build !fyne

package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type statusView struct {
	colors colorProvider

	radioLabel       *gtk.Label
	keyerLabel       *gtk.Label
	dxccLabel        *gtk.Label
	scpLabel         *gtk.Label
	callHistoryLabel *gtk.Label
	mapLabel         *gtk.Label
}

const (
	availableColor   = "theme_fg_color"
	unavailableColor = "unfocused_insensitive_color"
)

func setupStatusView(builder *gtk.Builder, colors colorProvider) *statusView {
	result := &statusView{
		colors: colors,
	}

	result.radioLabel = getUI(builder, "radioStatusLabel").(*gtk.Label)
	result.keyerLabel = getUI(builder, "keyerStatusLabel").(*gtk.Label)
	result.dxccLabel = getUI(builder, "dxccStatusLabel").(*gtk.Label)
	result.scpLabel = getUI(builder, "scpStatusLabel").(*gtk.Label)
	result.callHistoryLabel = getUI(builder, "callHistoryStatusLabel").(*gtk.Label)
	result.mapLabel = getUI(builder, "mapStatusLabel").(*gtk.Label)

	style := result.indicatorStyle(false)
	setStyledText(result.radioLabel, style, "Radio")
	setStyledText(result.keyerLabel, style, "CW")
	setStyledText(result.dxccLabel, style, "DXCC")
	setStyledText(result.scpLabel, style, "SCP")
	setStyledText(result.callHistoryLabel, style, "CH")
	setStyledText(result.mapLabel, style, "Map")

	return result
}

func (v *statusView) indicatorStyle(available bool) string {
	var color string
	if available {
		color = availableColor
	} else {
		color = unavailableColor
	}
	return fmt.Sprintf("foreground='%s'", v.colors.ColorByName(color).ToWeb())
}

func (v *statusView) StatusChanged(service core.Service, available bool) {
	log.Printf("service status changed: %d, %t", service, available)
	label, text := v.serviceLabel(service)
	if label == nil {
		log.Printf("unknown service %d", service)
		return
	}

	style := v.indicatorStyle(available)
	setStyledText(label, style, text)
}

func (v *statusView) serviceLabel(service core.Service) (*gtk.Label, string) {
	switch service {
	case core.RadioService:
		return v.radioLabel, "Radio"
	case core.KeyerService:
		return v.keyerLabel, "CW"
	case core.DXCCService:
		return v.dxccLabel, "DXCC"
	case core.SCPService:
		return v.scpLabel, "SCP"
	case core.CallHistoryService:
		return v.callHistoryLabel, "CH"
	case core.MapService:
		return v.mapLabel, "Map"
	default:
		return nil, ""
	}
}

func setStyledText(label *gtk.Label, style, text string) {
	label.SetMarkup(fmt.Sprintf(`<span %s>%s</span>`, style, text))
}
