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

const (
	callinfoDuplicateClass style.Class = "callinfo-duplicate"
	callinfoWorkedClass    style.Class = "callinfo-worked"
	callinfoWorthlessClass style.Class = "callinfo-worthless"
	callinfoMultiClass     style.Class = "callinfo-multi"
)

type callinfoView struct {
	controller CallinfoController

	rootGrid        *gtk.Grid
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

	result.rootGrid = getUI(builder, "callinfoGrid").(*gtk.Grid)
	result.callsignLabel = getUI(builder, "callsignLabel").(*gtk.Label)
	result.exchangeLabel = getUI(builder, "xchangeLabel").(*gtk.Label)
	result.dxccLabel = getUI(builder, "dxccLabel").(*gtk.Label)
	result.valueLabel = getUI(builder, "valueLabel").(*gtk.Label)
	result.userInfoLabel = getUI(builder, "userInfoLabel").(*gtk.Label)
	result.supercheckLabel = getUI(builder, "supercheckLabel").(*gtk.Label)

	return result
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
	normalized := strings.ToUpper(strings.TrimSpace(callsign))
	if normalized == "" {
		normalized = "-"
	}
	v.callsignLabel.SetText(normalized)

	style.RemoveClass(&v.callsignLabel.Widget, callinfoDuplicateClass)
	style.RemoveClass(&v.callsignLabel.Widget, callinfoWorkedClass)
	style.RemoveClass(&v.callsignLabel.Widget, callinfoWorthlessClass)

	if duplicate {
		style.AddClass(&v.callsignLabel.Widget, callinfoDuplicateClass)
	} else if worked {
		style.AddClass(&v.callsignLabel.Widget, callinfoWorkedClass)
	}
}

func (v *callinfoView) SetDXCC(dxccName, continent string, itu, cq int, arrlCompliant bool) {
	if dxccName == "" {
		v.dxccLabel.SetMarkup("")
		return
	}

	text := fmt.Sprintf("%s, %s", dxccName, continent)
	if itu != 0 {
		text += fmt.Sprintf(", ITU %d", itu)
	}
	if cq != 0 {
		text += fmt.Sprintf(", CQ %d", cq)
	}
	if dxccName != "" && !arrlCompliant {
		text += ", <span foreground='red' font-weight='heavy'>not ARRL compliant</span>"
	}

	v.dxccLabel.SetMarkup(text)
}

func (v *callinfoView) SetValue(points, multis int) {
	style.RemoveClass(&v.valueLabel.Widget, callinfoWorthlessClass)
	style.RemoveClass(&v.valueLabel.Widget, callinfoMultiClass)

	switch {
	case points < 1 && multis < 1:
		style.AddClass(&v.valueLabel.Widget, callinfoWorthlessClass)
	case multis > 0:
		style.AddClass(&v.valueLabel.Widget, callinfoMultiClass)
	}

	v.valueLabel.SetText(fmt.Sprintf("%dP %dM", points, multis))
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
	v.userInfoLabel.SetText(value)
}

func (v *callinfoView) SetSupercheck(callsigns []core.AnnotatedCallsign) {
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
