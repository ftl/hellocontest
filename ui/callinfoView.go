package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type callinfoView struct {
	callsignLabel   *gtk.Label
	dxccLabel       *gtk.Label
	continentLabel  *gtk.Label
	ituLabel        *gtk.Label
	cqLabel         *gtk.Label
	arrlLabel       *gtk.Label
	supercheckLabel *gtk.Label
}

func setupCallinfoView(builder *gtk.Builder) *callinfoView {
	result := new(callinfoView)

	result.callsignLabel = getUI(builder, "callsignLabel").(*gtk.Label)
	result.dxccLabel = getUI(builder, "dxccLabel").(*gtk.Label)
	result.continentLabel = getUI(builder, "continentLabel").(*gtk.Label)
	result.ituLabel = getUI(builder, "ituLabel").(*gtk.Label)
	result.cqLabel = getUI(builder, "cqLabel").(*gtk.Label)
	result.arrlLabel = getUI(builder, "arrlLabel").(*gtk.Label)
	result.supercheckLabel = getUI(builder, "supercheckLabel").(*gtk.Label)

	return result
}

func (v *callinfoView) SetCallsign(callsign string, worked, duplicate bool) {
	normalized := strings.ToUpper(strings.TrimSpace(callsign))
	if normalized == "" {
		v.callsignLabel.SetMarkup("-")
		return
	}

	// see https://developer.gnome.org/pango/stable/pango-Markup.html for reference
	attributes := make([]string, 0)
	if duplicate {
		attributes = append(attributes, "background='red' foreground='white'")
	} else if worked {
		attributes = append(attributes, "foreground='blue'")
	}
	attributeString := strings.Join(attributes, " ")

	renderedCallsign := fmt.Sprintf("<span %s>%s</span>", attributeString, normalized)
	v.callsignLabel.SetMarkup(renderedCallsign)
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

func (v *callinfoView) SetSupercheck(callsigns []core.AnnotatedCallsign) {
	text := ""
	for _, callsign := range callsigns {
		// see https://developer.gnome.org/pango/stable/pango-Markup.html for reference
		attributes := make([]string, 0)
		if callsign.Duplicate {
			attributes = append(attributes, "foreground='red'")
		} else if callsign.Worked {
			attributes = append(attributes, "foreground='blue'")
		}
		if callsign.ExactMatch {
			attributes = append(attributes, "font-weight='heavy' font-size='large'")
		}
		attributeString := strings.Join(attributes, " ")

		renderedCallsign := fmt.Sprintf("<span %s>%s</span>", attributeString, callsign.Callsign)

		if text != "" {
			text += ", "
		}
		text += renderedCallsign
	}
	v.supercheckLabel.SetMarkup(text)
}
