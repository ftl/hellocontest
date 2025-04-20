package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hamradio/latlon"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
	"github.com/ftl/hellocontest/ui/style"
)

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
	currentFrame core.CallinfoFrame

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

func (v *callinfoView) RefreshStyle() {
	v.style.Refresh()
	v.showCurrentFrame()
}

func attr(name, value string) string {
	return fmt.Sprintf("%s=%q", name, value)
}

func (v *callinfoView) ShowFrame(frame core.CallinfoFrame) {
	v.currentFrame = frame
	v.showCurrentFrame()
}

func (v *callinfoView) showCurrentFrame() {
	v.setBestMatchingCallsign(v.currentFrame.BestMatchOnFrequency())
	v.setDXCC(v.currentFrame.DXCCEntity, v.currentFrame.Azimuth, v.currentFrame.Distance)
	v.setValue(v.currentFrame.Points, v.currentFrame.Multis, v.currentFrame.Value)
	v.setUserInfo(v.currentFrame.UserInfo)
	v.setSupercheck(v.currentFrame.Supercheck)
	v.setPredictedExchanges(v.currentFrame.PredictedExchange)
}

func (v *callinfoView) setBestMatchingCallsign(callsign core.AnnotatedCallsign) {
	v.callsignLabel.SetMarkup(v.renderCallsign(callsign))
}

func (v *callinfoView) setDXCC(entity dxcc.Prefix, azimuth latlon.Degrees, distance latlon.Km) {
	if entity.Name == "" {
		v.dxccLabel.SetMarkup("")
		return
	}

	text := entity.Name
	if entity.PrimaryPrefix != "" {
		text += fmt.Sprintf(" (%s)", entity.PrimaryPrefix)
	}
	text += fmt.Sprintf(", %s", entity.Continent)
	if entity.ITUZone != 0 {
		text += fmt.Sprintf(", ITU %d", entity.ITUZone)
	}
	if entity.CQZone != 0 {
		text += fmt.Sprintf(", CQ %d", entity.CQZone)
	}
	if distance > 0 {
		text += fmt.Sprintf(", %s, %s", distance, azimuth)
	}

	v.dxccLabel.SetMarkup(text)
}

func (v *callinfoView) setValue(points int, multis int, value int) {
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

func (v *callinfoView) setUserInfo(value string) {
	v.userInfoLabel.SetText(value)
}

func (v *callinfoView) setSupercheck(callsigns []core.AnnotatedCallsign) {
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

func (v *callinfoView) setPredictedExchanges(exchanges []string) {
	for i := range v.predictedExchanges {
		if i < len(exchanges) {
			v.setPredictedExchange(i, exchanges[i])
		} else {
			v.setPredictedExchange(i, "")
		}
	}
}

func (v *callinfoView) setPredictedExchange(index int, text string) {
	if index < 0 || index >= len(v.predictedExchanges) {
		return
	}

	var exchangeMarkup string
	if text == "" {
		text = "-"
	} else {
		text = strings.TrimSpace(text)
	}
	exchangeText := fmt.Sprintf("<span %s>%s</span>", exchangeMarkup, text)
	v.predictedExchanges[index].SetMarkup(exchangeText)
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
