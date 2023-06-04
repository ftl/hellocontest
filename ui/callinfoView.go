package ui

import (
	"fmt"
	"strings"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

type CallinfoController interface {
	Refresh()
}

type callinfoStyle struct {
	colorProvider

	backgroundColor style.Color
	fontColor       style.Color
	fontSize        float64
	duplicateFG     style.Color
	duplicateBG     style.Color
	workedFG        style.Color
	workedBG        style.Color
	worthlessFG     style.Color
	worthlessBG     style.Color
}

func (s *callinfoStyle) Refresh() {
	s.backgroundColor = s.colorProvider.BackgroundColor()
	s.fontColor = s.colorProvider.ForegroundColor()
	s.duplicateFG = s.colorProvider.ColorByName(duplicateFGColorName)
	s.duplicateBG = s.colorProvider.ColorByName(duplicateBGColorName)
	s.workedFG = s.colorProvider.ColorByName(workedFGColorName)
	s.workedBG = s.colorProvider.ColorByName(workedBGColorName)
	s.worthlessFG = s.colorProvider.ColorByName(worthlessFGColorName)
	s.worthlessBG = s.colorProvider.ColorByName(worthlessBGColorName)
}

type callinfoView struct {
	controller CallinfoController

	callsignLabel   *gtk.Label
	exchangeLabel   *gtk.Label
	dxccLabel       *gtk.Label
	valueLabel      *gtk.Label
	userInfoLabel   *gtk.Label
	supercheckLabel *gtk.Label

	style *callinfoStyle
}

func setupCallinfoView(builder *gtk.Builder, colors colorProvider, controller CallinfoController) *callinfoView {
	result := &callinfoView{
		controller: controller,
		style:      &callinfoStyle{colorProvider: colors},
	}
	result.style.Refresh()

	result.callsignLabel = getUI(builder, "callsignLabel").(*gtk.Label)
	result.exchangeLabel = getUI(builder, "xchangeLabel").(*gtk.Label)
	result.dxccLabel = getUI(builder, "dxccLabel").(*gtk.Label)
	result.valueLabel = getUI(builder, "valueLabel").(*gtk.Label)
	result.userInfoLabel = getUI(builder, "userInfoLabel").(*gtk.Label)
	result.supercheckLabel = getUI(builder, "supercheckLabel").(*gtk.Label)

	return result
}

func (c *callinfoView) SetCallinfoController(controller CallinfoController) {
	c.controller = controller
}

func (v *callinfoView) RefreshStyle() {
	v.style.Refresh()
	if v.controller != nil {
		v.controller.Refresh()
	}
}

func attr(name, value string) string {
	return fmt.Sprintf("%s=%q", name, value)
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
		attributes = append(attributes,
			attr("background", v.style.duplicateBG.ToWeb()),
			attr("foreground", v.style.duplicateFG.ToWeb()),
		)
	} else if worked {
		attributes = append(attributes,
			attr("background", v.style.workedBG.ToWeb()),
			attr("foreground", v.style.workedFG.ToWeb()),
		)
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
		text += ", <span foreground='red' font-weight='heavy'>not ARRL compliant</span>"
	}

	v.dxccLabel.SetMarkup(text)
}

func (v *callinfoView) SetValue(points, multis int) {
	if v == nil {
		return
	}

	// see https://docs.gtk.org/Pango/pango_markup.html for reference
	var pointsMarkup string
	switch {
	case points < 1:
		pointsMarkup = attr("foreground", "silver")
	case points > 1:
		pointsMarkup = attr("font-weight", "heavy")
	}

	var multisMarkup string
	switch {
	case multis < 1:
		multisMarkup = attr("foreground", "silver")
	case multis > 1:
		multisMarkup = attr("font-weight", "heavy")
	}

	valueText := fmt.Sprintf("<span %s>%d points</span> / <span %s>%d multis</span>", pointsMarkup, points, multisMarkup, multis)
	v.valueLabel.SetMarkup(valueText)
}

func (v *callinfoView) SetExchange(exchange string) {
	var exchangeMarkup string
	if exchange == "" {
		exchange = "-"
	} else {
		exchange = strings.ToUpper(strings.TrimSpace(exchange))
	}
	exchangeText := fmt.Sprintf("<span %s>%s</span>", exchangeMarkup, exchange)
	v.exchangeLabel.SetMarkup(exchangeText)
}

func (v *callinfoView) SetUserInfo(value string) {
	if v == nil {
		return
	}
	v.userInfoLabel.SetText(value)
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
			attributes = append(attributes,
				attr("background", v.style.duplicateBG.ToWeb()),
				attr("foreground", v.style.duplicateFG.ToWeb()),
			)

		case callsign.Worked:
			attributes = append(attributes,
				attr("background", v.style.workedBG.ToWeb()),
				attr("foreground", v.style.workedFG.ToWeb()),
			)
		case (callsign.Points == 0) && (callsign.Multis == 0):
			attributes = append(attributes,
				attr("background", v.style.worthlessBG.ToWeb()),
				attr("foreground", v.style.worthlessFG.ToWeb()),
			)
		case callsign.Multis > 0:
			attributes = append(attributes,
				attr("background", v.style.backgroundColor.ToWeb()),
				attr("foreground", v.style.fontColor.ToWeb()),
				attr("font-weight", "heavy"),
			)
		default:
			attributes = append(attributes,
				attr("background", v.style.backgroundColor.ToWeb()),
				attr("foreground", v.style.fontColor.ToWeb()),
			)
		}

		hasPredictedExchange := strings.Join(callsign.PredictedExchange, "") != ""
		if callsign.ExactMatch || hasPredictedExchange {
			attributes = append(attributes, attr("font-size", "x-large"))
		}
		if hasPredictedExchange {
			attributes = append(attributes, attr("font-style", "italic"))
		}
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
				partAttributeString = "underline='single'"
				partString = part.Value
			case core.FalseFriend:
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
