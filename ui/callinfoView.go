package ui

import (
	"fmt"
	"strings"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type callinfoView struct {
	callsignLabel   *gtk.Label
	xchangeLabel    *gtk.Label
	dxccLabel       *gtk.Label
	valueLabel      *gtk.Label
	supercheckLabel *gtk.Label
}

func setupCallinfoView(builder *gtk.Builder) *callinfoView {
	result := new(callinfoView)

	result.callsignLabel = getUI(builder, "callsignLabel").(*gtk.Label)
	result.xchangeLabel = getUI(builder, "xchangeLabel").(*gtk.Label)
	result.dxccLabel = getUI(builder, "dxccLabel").(*gtk.Label)
	result.valueLabel = getUI(builder, "valueLabel").(*gtk.Label)
	result.supercheckLabel = getUI(builder, "supercheckLabel").(*gtk.Label)

	return result
}

func (v *callinfoView) SetCallsign(callsign string, worked, duplicate bool) {
	if v == nil {
		return
	}

	normalized := strings.ToUpper(strings.TrimSpace(callsign))
	if normalized == "" {
		v.callsignLabel.SetMarkup("-")
		return
	}

	// see https://docs.gtk.org/Pango/pango_markup.html for reference
	attributes := make([]string, 0, 1)
	if duplicate {
		attributes = append(attributes, "background='red' foreground='white'")
	} else if worked {
		attributes = append(attributes, "foreground='orange'")
	}
	attributeString := strings.Join(attributes, " ")

	renderedCallsign := fmt.Sprintf("<span %s>%s</span>", attributeString, normalized)
	v.callsignLabel.SetMarkup(renderedCallsign)
}

func (v *callinfoView) SetDXCC(name, continent string, itu, cq int, arrlCompliant bool) {
	if v == nil {
		return
	}

	if name == "" {
		v.dxccLabel.SetMarkup("")
		return
	}

	text := fmt.Sprintf("%s, %s", name, continent)
	if itu != 0 {
		text += fmt.Sprintf(", ITU %d", itu)
	}
	if cq != 0 {
		text += fmt.Sprintf(", CQ %d", cq)
	}
	if name != "" && !arrlCompliant {
		text += fmt.Sprintf(", <span foreground='red' font-weight='heavy'>not ARRL compliant</span>")
	}

	v.dxccLabel.SetMarkup(text)
}

func (v *callinfoView) SetValue(points, multis int, xchange string) {
	if v == nil {
		return
	}

	// see https://docs.gtk.org/Pango/pango_markup.html for reference
	var pointsMarkup string
	switch {
	case points < 1:
		pointsMarkup = "foreground='silver'"
	case points > 1:
		pointsMarkup = "font-weight='heavy'"
	}

	var multisMarkup string
	switch {
	case multis < 1:
		multisMarkup = "foreground='silver'"
	case multis > 1:
		multisMarkup = "font-weight='heavy'"
	}

	var xchangeMarkup string
	switch {
	case multis < 1:
		xchangeMarkup = "foreground='silver'"
	case multis > 1:
		xchangeMarkup = "font-weight='heavy'"
	}

	valueText := fmt.Sprintf("<span %s>%d points</span> / <span %s>%d multis</span>", pointsMarkup, points, multisMarkup, multis)
	v.valueLabel.SetMarkup(valueText)

	if xchange == "" {
		xchange = "-"
	} else {
		xchange = strings.ToUpper(strings.TrimSpace(xchange))
	}
	xchangeText := fmt.Sprintf("<span %s>%s</span>", xchangeMarkup, xchange)
	v.xchangeLabel.SetMarkup(xchangeText)
}

func (v *callinfoView) SetSupercheck(callsigns []core.AnnotatedCallsign) {
	if v == nil {
		return
	}

	var text string
	for _, callsign := range callsigns {
		// see https://docs.gtk.org/Pango/pango_markup.html for reference
		attributes := make([]string, 0, 3)
		switch {
		case callsign.Duplicate:
			attributes = append(attributes, "foreground='red'")
		case callsign.Worked:
			attributes = append(attributes, "foreground='orange'")
		case (callsign.Points == 0) && (callsign.Multis == 0):
			attributes = append(attributes, "foreground='silver'")
		case callsign.Multis > 0:
			attributes = append(attributes, "font-weight='heavy'")
		}
		if callsign.ExactMatch {
			attributes = append(attributes, "font-size='large'")
		}
		attributes = append(attributes, "background='white'")
		attributeString := strings.Join(attributes, " ")

		var renderedCallsign string
		for _, part := range callsign.Assembly {
			var partAttributeString string
			var partString string
			switch part.OP {
			case core.Matching:
				partAttributeString = ""
				partString = part.Value
			case core.Insert:
				partAttributeString = "underline='single'"
				partString = part.Value
			case core.Delete:
				partAttributeString = ""
				partString = "|"
			case core.Substitute:
				partAttributeString = "underline='double'"
				partString = part.Value
			}
			renderedCallsign += fmt.Sprintf("<span %s>%s</span>", strings.Join([]string{attributeString, partAttributeString}, " "), partString)
		}

		if text != "" {
			text += "   "
		}
		text += renderedCallsign
	}
	v.supercheckLabel.SetMarkup(text)
}
