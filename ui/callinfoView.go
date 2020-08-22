package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

type callinfoView struct {
	controller core.CallinfoController

	callsignLabel  *gtk.Label
	dxccLabel      *gtk.Label
	continentLabel *gtk.Label
	ituLabel       *gtk.Label
	cqLabel        *gtk.Label
	arrlLabel      *gtk.Label
}

func setupCallinfoView(builder *gtk.Builder) *callinfoView {
	result := new(callinfoView)

	result.callsignLabel = getUI(builder, "callsignLabel").(*gtk.Label)
	result.dxccLabel = getUI(builder, "dxccLabel").(*gtk.Label)
	result.continentLabel = getUI(builder, "continentLabel").(*gtk.Label)
	result.ituLabel = getUI(builder, "ituLabel").(*gtk.Label)
	result.cqLabel = getUI(builder, "cqLabel").(*gtk.Label)
	result.arrlLabel = getUI(builder, "arrlLabel").(*gtk.Label)

	return result
}

func (v *callinfoView) SetCallinfoController(controller core.CallinfoController) {
	v.controller = controller
}

func (v *callinfoView) SetCallsign(callsign string) {
	normalized := strings.ToUpper(strings.TrimSpace(callsign))
	if normalized != "" {
		v.callsignLabel.SetText(normalized)
	} else {
		v.callsignLabel.SetText("-")
	}
}

func (v *callinfoView) SetDXCC(name, continent string, itu, cq int, arrlCompliant bool) {
	v.dxccLabel.SetText(name)
	v.continentLabel.SetText(continent)
	var ituText string
	if itu == 0 {
		ituText = "-"
	} else {
		ituText = strconv.Itoa(itu)
	}
	v.ituLabel.SetText(fmt.Sprintf("ITU: %s", ituText))
	var cqText string
	if cq == 0 {
		cqText = "-"
	} else {
		cqText = strconv.Itoa(cq)
	}
	v.cqLabel.SetText(fmt.Sprintf("CQ: %s", cqText))
	var arrlText string
	if arrlCompliant {
		arrlText = "compliant"
	} else {
		arrlText = "not compl."
	}
	v.arrlLabel.SetText(fmt.Sprintf("ARRL: %s", arrlText))
}
