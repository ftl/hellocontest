package fyneui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ftl/hellocontest/core"
)

type statusBar struct {
	container *fyne.Container

	radio       *widget.Label
	keyer       *widget.Label
	dxcc        *widget.Label
	scp         *widget.Label
	callhistory *widget.Label
	mapLabel    *widget.Label
}

func setupStatusBar() *statusBar {
	result := &statusBar{
		radio:       widget.NewLabel(""),
		keyer:       widget.NewLabel(""),
		dxcc:        widget.NewLabel(""),
		scp:         widget.NewLabel(""),
		callhistory: widget.NewLabel(""),
		mapLabel:    widget.NewLabel(""),
	}
	result.container = container.New(layout.NewHBoxLayout(), result.radio, result.keyer, result.dxcc, result.scp, result.callhistory, result.mapLabel, layout.NewSpacer())

	result.updateStatus(core.RadioService, false)
	result.updateStatus(core.KeyerService, false)
	result.updateStatus(core.DXCCService, false)
	result.updateStatus(core.SCPService, false)
	result.updateStatus(core.CallHistoryService, false)
	result.updateStatus(core.MapService, false)

	return result
}

func (b *statusBar) StatusChanged(service core.Service, available bool) {
	log.Printf("service status changed: %d, %t", service, available)
	b.updateStatus(service, available)
}

func (b *statusBar) updateStatus(service core.Service, available bool) {
	label, text := b.serviceLabel(service)
	if label == nil {
		log.Printf("unknown service %d", service)
		return
	}

	label.TextStyle.Bold = available
	label.SetText(text)
}

func (b *statusBar) serviceLabel(service core.Service) (*widget.Label, string) {
	switch service {
	case core.RadioService:
		return b.radio, "Radio"
	case core.KeyerService:
		return b.keyer, "CW"
	case core.DXCCService:
		return b.dxcc, "DXCC"
	case core.SCPService:
		return b.scp, "SCP"
	case core.CallHistoryService:
		return b.callhistory, "CH"
	case core.MapService:
		return b.mapLabel, "Map"
	default:
		return nil, ""
	}
}
