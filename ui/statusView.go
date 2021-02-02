package ui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type statusView struct {
	tciLabel    *gtk.Label
	hamlibLabel *gtk.Label
	cwLabel     *gtk.Label
	dxccLabel   *gtk.Label
	scpLabel    *gtk.Label
}

const (
	availableStyle   = "foreground='black'"
	unavailableStyle = "foreground='lightgray'"
)

func setupStatusView(builder *gtk.Builder) *statusView {
	result := new(statusView)

	result.tciLabel = getUI(builder, "tciStatusLabel").(*gtk.Label)
	result.hamlibLabel = getUI(builder, "hamlibStatusLabel").(*gtk.Label)
	result.cwLabel = getUI(builder, "cwStatusLabel").(*gtk.Label)
	result.dxccLabel = getUI(builder, "dxccStatusLabel").(*gtk.Label)
	result.scpLabel = getUI(builder, "scpStatusLabel").(*gtk.Label)

	setStyledText(result.tciLabel, unavailableStyle, "TCI")
	setStyledText(result.hamlibLabel, unavailableStyle, "Hamlib")
	setStyledText(result.cwLabel, unavailableStyle, "CW")
	setStyledText(result.dxccLabel, unavailableStyle, "DXCC")
	setStyledText(result.scpLabel, unavailableStyle, "SCP")

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
	case core.TCIService:
		return v.tciLabel, "TCI"
	case core.HamlibService:
		return v.hamlibLabel, "Hamlib"
	case core.CWDaemonService:
		return v.cwLabel, "CW"
	case core.DXCCService:
		return v.dxccLabel, "DXCC"
	case core.SCPService:
		return v.scpLabel, "SCP"
	default:
		return nil, ""
	}
}

func setStyledText(label *gtk.Label, style, text string) {
	label.SetMarkup(fmt.Sprintf(`<span %s>%s</span>`, style, text))
}
