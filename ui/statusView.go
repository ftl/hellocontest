package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

var statusServiceOrder = []struct {
	svc  core.Service
	name string
}{
	{core.RadioService, "Radio"},
	{core.KeyerService, "CW"},
	{core.DXCCService, "DXCC"},
	{core.SCPService, "SCP"},
	{core.CallHistoryService, "CH"},
	{core.MapService, "Map"},
}

type statusView struct {
	statusBar *qtlib.QStatusBar
	labels    map[core.Service]*qtlib.QLabel
}

func newStatusView() *statusView {
	v := &statusView{labels: make(map[core.Service]*qtlib.QLabel)}
	v.statusBar = qtlib.NewQStatusBar2()
	v.statusBar.SetObjectName(*qtlib.NewQAnyStringView3("statusBar"))
	for _, s := range statusServiceOrder {
		lbl := qtlib.NewQLabel3(s.name)
		v.applyStyle(lbl, false)
		v.statusBar.AddWidget(lbl.QWidget)
		v.labels[s.svc] = lbl
	}
	return v
}

func (v *statusView) StatusChanged(service core.Service, available bool) {
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
