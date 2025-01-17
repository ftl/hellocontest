package ui

import (
	"fmt"
	"log"
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

	callsignLabel            *gtk.Label
	dxccLabel                *gtk.Label
	valueLabel               *gtk.Label
	userInfoLabel            *gtk.Label
	supercheckLabel          *gtk.Label
	predictedExchangesParent *gtk.Grid
	predictedExchanges       []*gtk.Label

	style *callinfoStyle
}

func setupCallinfoView(builder *gtk.Builder, colors colorProvider) *callinfoView {
	result := &callinfoView{
		style: &callinfoStyle{colorProvider: colors},
	}
	result.style.Refresh()

	result.callsignLabel = getUI(builder, "bestMatchCallsign").(*gtk.Label)
	result.dxccLabel = getUI(builder, "callsignDXCCLabel").(*gtk.Label)
	result.valueLabel = getUI(builder, "predictedValueLabel").(*gtk.Label)
	result.userInfoLabel = getUI(builder, "callsignUserInfoLabel").(*gtk.Label)
	result.supercheckLabel = getUI(builder, "callsignSupercheckLabel").(*gtk.Label)
	result.predictedExchangesParent = getUI(builder, "predictedExchangesGrid").(*gtk.Grid)

	return result
}

func (v *callinfoView) SetCallinfoController(controller CallinfoController) {
	v.controller = controller
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

func (v *callinfoView) SetBestMatchingCallsign(callsign core.AnnotatedCallsign) {
	v.callsignLabel.SetMarkup(v.renderCallsign(callsign))
}

func (v *callinfoView) SetDXCC(dxccName, continent string, itu, cq int) {
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

	v.dxccLabel.SetMarkup(text)
}

func (v *callinfoView) SetValue(points int, multis int, value int) {
	style.RemoveClass(&v.valueLabel.Widget, callinfoWorthlessClass)
	style.RemoveClass(&v.valueLabel.Widget, callinfoMultiClass)

	switch {
	case points < 1 && multis < 1:
		style.AddClass(&v.valueLabel.Widget, callinfoWorthlessClass)
	case multis > 0:
		style.AddClass(&v.valueLabel.Widget, callinfoMultiClass)
	}

	v.valueLabel.SetText(fmt.Sprintf("%dP x %dM = %d", points, multis, value))
}

func (v *callinfoView) SetUserInfo(value string) {
	v.userInfoLabel.SetText(value)
}

func (v *callinfoView) SetSupercheck(callsigns []core.AnnotatedCallsign) {
	var text string
	for i, callsign := range callsigns {
		if text != "" {
			text += "   "
		}
		if i < 9 {
			text += fmt.Sprintf("(%d) ", i+1)
		}

		text += v.renderCallsign(callsign)
	}
	v.supercheckLabel.SetMarkup(text)
}

func (v *callinfoView) renderCallsign(callsign core.AnnotatedCallsign) string {
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
	if hasPredictedExchange {
		attributes = append(attributes, attr("font-style", "italic"))
	}
	attributeString := strings.Join(attributes, " ")

	renderedCallsign := ""
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

	if callsign.OnFrequency {
		renderedCallsign = fmt.Sprintf("[%s]", renderedCallsign)
	}

	return renderedCallsign
}

func (v *callinfoView) SetPredictedExchange(index int, text string) {
	i := index - 1
	if i < 0 || i >= len(v.predictedExchanges) {
		return
	}

	var exchangeMarkup string
	if text == "" {
		text = "-"
	} else {
		text = strings.TrimSpace(text)
	}
	exchangeText := fmt.Sprintf("<span %s>%s</span>", exchangeMarkup, text)
	v.predictedExchanges[i].SetMarkup(exchangeText)
}

func (v *callinfoView) SetPredictedExchangeFields(fields []core.ExchangeField) {
	v.setExchangeFields(fields, v.predictedExchangesParent, &v.predictedExchanges)
}

func (v *callinfoView) setExchangeFields(fields []core.ExchangeField, parent *gtk.Grid, labels *[]*gtk.Label) {
	for _, label := range *labels {
		label.Destroy()
		parent.RemoveColumn(0)
	}

	*labels = make([]*gtk.Label, len(fields))
	for i, field := range fields {
		label, err := gtk.LabelNew("")
		if err != nil {
			log.Printf("cannot create entry for %s: %v", field.Field, err)
			break
		}
		label.SetName(string(field.Field))
		label.SetTooltipText(field.Short) // TODO use field.Hint
		label.SetHExpand(true)
		label.SetHAlign(gtk.ALIGN_FILL)
		label.SetXAlign(0)

		(*labels)[i] = label
		parent.Add(label)
	}
	parent.ShowAll()
}
