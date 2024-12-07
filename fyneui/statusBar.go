package fyneui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ftl/hellocontest/core"
)

type statusBar struct {
	container *fyne.Container

	radio       *widget.RichText
	keyer       *widget.RichText
	dxcc        *widget.RichText
	scp         *widget.RichText
	callhistory *widget.RichText
	mapLabel    *widget.RichText
}

func setupStatusBar() *statusBar {
	result := &statusBar{
		radio:       widget.NewRichText(&widget.TextSegment{}),
		keyer:       widget.NewRichText(&widget.TextSegment{}),
		dxcc:        widget.NewRichText(&widget.TextSegment{}),
		scp:         widget.NewRichText(&widget.TextSegment{}),
		callhistory: widget.NewRichText(&widget.TextSegment{}),
		mapLabel:    widget.NewRichText(&widget.TextSegment{}),
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
	indicator, text := b.serviceIndicator(service)
	if indicator == nil {
		log.Printf("unknown service %d", service)
	}

	var color fyne.ThemeColorName
	if available {
		color = theme.ColorNameForeground
	} else {
		color = theme.ColorNameDisabled
	}

	segment := indicator.Segments[0].(*widget.TextSegment)
	segment.Text = text
	segment.Style.ColorName = color
	segment.Style.TextStyle.Bold = available
	indicator.Refresh()
}

func (b *statusBar) serviceIndicator(service core.Service) (*widget.RichText, string) {
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
