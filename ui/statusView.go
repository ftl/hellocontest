package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core/model"
)

var statusServiceOrder = []struct {
	svc  model.Service
	name string
}{
	{model.RadioService, "Radio"},
	{model.KeyerService, "CW"},
	{model.DXCCService, "DXCC"},
	{model.SCPService, "SCP"},
	{model.CallHistoryService, "CH"},
	{model.MapService, "Map"},
	{model.RemoteService, "RMT"},
}

type statusView struct {
	statusBar *qtlib.QStatusBar
	labels    map[model.Service]*qtlib.QLabel
}

func newStatusView() *statusView {
	v := &statusView{labels: make(map[model.Service]*qtlib.QLabel)}
	v.statusBar = qtlib.NewQStatusBar2()
	setObjectName(v.statusBar, "statusBar")
	for _, s := range statusServiceOrder {
		lbl := qtlib.NewQLabel3(s.name)
		v.applyStyle(lbl, false)
		v.statusBar.AddWidget(lbl.QWidget)
		v.labels[s.svc] = lbl
	}
	return v
}

func (v *statusView) StatusChanged(service model.Service, available bool) {
	lbl, ok := v.labels[service]
	if !ok {
		return
	}
	v.applyStyle(lbl, available)
}

func (v *statusView) applyStyle(lbl *qtlib.QLabel, available bool) {
	if available {
		lbl.SetStyleSheet("")
	} else {
		lbl.SetStyleSheet(MutedTextStyle)
	}
}
