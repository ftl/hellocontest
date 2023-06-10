package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type statusView struct {
	radioLabel       *gtk.Label
	keyerLabel       *gtk.Label
	dxccLabel        *gtk.Label
	scpLabel         *gtk.Label
	callHistoryLabel *gtk.Label
}

const (
	availableStyle   = "foreground='black'"
	unavailableStyle = "foreground='lightgray'"
)

func setupStatusView(builder *gtk.Builder) *statusView {
	result := new(statusView)

	result.radioLabel = getUI(builder, "radioStatusLabel").(*gtk.Label)
	result.keyerLabel = getUI(builder, "keyerStatusLabel").(*gtk.Label)
	result.dxccLabel = getUI(builder, "dxccStatusLabel").(*gtk.Label)
	result.scpLabel = getUI(builder, "scpStatusLabel").(*gtk.Label)
	result.callHistoryLabel = getUI(builder, "callHistoryStatusLabel").(*gtk.Label)

	setStyledText(result.radioLabel, unavailableStyle, "Radio")
	setStyledText(result.keyerLabel, unavailableStyle, "CW")
	setStyledText(result.dxccLabel, unavailableStyle, "DXCC")
	setStyledText(result.scpLabel, unavailableStyle, "SCP")
	setStyledText(result.callHistoryLabel, unavailableStyle, "CH")

	return result
}

func (v *statusView) StatusChanged(service core.Service, available bool) {
	log.Printf("service status changed: %d, %t", service, available)
	label, text := v.serviceLabel(service)
	if label == nil {
		log.Printf("unknown service %d", service)
		return
	}

	var style string
	if available {
		style = availableStyle
	} else {
		style = unavailableStyle
	}
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
	default:
		return nil, ""
	}
}

func setStyledText(label *gtk.Label, style, text string) {
	label.SetMarkup(fmt.Sprintf(`<span %s>%s</span>`, style, text))
}
