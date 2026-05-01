package ui

import (
	"fmt"
	"log"
	"strings"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/latlon"

	"github.com/ftl/hellocontest/core"
)

type callinfoView struct {
	widget *qtlib.QWidget

	callsignLabel   *qtlib.QLabel
	valueLabel      *qtlib.QLabel
	qtcStatusLabel  *qtlib.QLabel
	infoContainer   *qtlib.QWidget
	dxccLabel       *qtlib.QLabel
	userInfoLabel   *qtlib.QLabel
	supercheckLabel *qtlib.QLabel

	predictedExchangeLabels []*qtlib.QLabel

	qtcsEnabled bool
	current     core.CallinfoFrame
}

func newCallinfoView() *callinfoView {
	v := &callinfoView{}

	v.widget = qtlib.NewQWidget2()

	v.callsignLabel = qtlib.NewQLabel2()
	v.callsignLabel.SetObjectName(*qtlib.NewQAnyStringView3("predictedBestMatch"))
	v.callsignLabel.SetTextFormat(qtlib.RichText)

	v.valueLabel = qtlib.NewQLabel2()
	v.valueLabel.SetTextFormat(qtlib.RichText)
	v.valueLabel.SetObjectName(*qtlib.NewQAnyStringView3("predictedValue"))

	v.qtcStatusLabel = qtlib.NewQLabel2()
	v.qtcStatusLabel.SetTextFormat(qtlib.RichText)
	v.qtcStatusLabel.SetObjectName(*qtlib.NewQAnyStringView3("qtcStatus"))
	v.qtcStatusLabel.SetAlignment(qtlib.AlignVCenter | qtlib.AlignTrailing)

	v.supercheckLabel = qtlib.NewQLabel2()
	v.supercheckLabel.SetObjectName(*qtlib.NewQAnyStringView3("supercheckLabel"))
	v.supercheckLabel.SetTextFormat(qtlib.RichText)
	v.supercheckLabel.QWidget.SetSizePolicy2(qtlib.QSizePolicy__Ignored, qtlib.QSizePolicy__Preferred)
	v.supercheckLabel.SetMinimumWidth(0)

	v.dxccLabel = qtlib.NewQLabel2()
	v.dxccLabel.SetObjectName(*qtlib.NewQAnyStringView3("dxccLabel"))
	v.dxccLabel.QWidget.SetSizePolicy2(qtlib.QSizePolicy__Ignored, qtlib.QSizePolicy__Preferred)
	v.dxccLabel.SetMinimumWidth(0)

	v.userInfoLabel = qtlib.NewQLabel2()
	v.userInfoLabel.SetObjectName(*qtlib.NewQAnyStringView3("userInfoLabel"))
	v.userInfoLabel.QWidget.SetSizePolicy2(qtlib.QSizePolicy__Ignored, qtlib.QSizePolicy__Preferred)
	v.userInfoLabel.SetMinimumWidth(0)
	v.userInfoLabel.SetAlignment(qtlib.AlignVCenter | qtlib.AlignTrailing)

	v.infoContainer = qtlib.NewQWidget2()
	infoLayout := qtlib.NewQHBoxLayout(v.infoContainer)
	infoLayout.SetContentsMargins(0, 0, 0, 0)
	infoLayout.AddWidget(v.dxccLabel.QWidget)
	infoLayout.AddWidget(v.userInfoLabel.QWidget)

	return v
}

func (v *callinfoView) ShowFrame(frame core.CallinfoFrame) {
	v.current = frame
	v.refresh()
}

func (v *callinfoView) SetQTCsEnabled(enabled bool) {
	v.qtcsEnabled = enabled
	v.refresh()
}

func (v *callinfoView) refresh() {
	best := v.current.BestMatchOnFrequency()
	v.callsignLabel.SetText(renderAnnotatedCallsignHTML(best))
	v.dxccLabel.SetText(renderDXCC(v.current.DXCCEntity, v.current.Azimuth, v.current.Distance))
	v.valueLabel.SetText(renderValue(v.current.Points, v.current.Multis, v.current.Value))
	v.qtcStatusLabel.SetText(renderQTCStatus(v.current.SentQTCs, v.current.ReceivedQTCs, v.qtcsEnabled))
	v.userInfoLabel.SetText(v.current.UserInfo)
	v.supercheckLabel.SetText(renderSupercheckHTML(v.current.Supercheck))

	for i, lbl := range v.predictedExchangeLabels {
		if i < len(v.current.PredictedExchange) {
			lbl.SetText(strings.TrimSpace(v.current.PredictedExchange[i]))
		} else {
			lbl.SetText("-")
		}
	}
}

func (v *callinfoView) SetPredictedExchangeFields(fields []core.ExchangeField) {
	for _, label := range v.predictedExchangeLabels {
		label.Delete()
	}
	v.predictedExchangeLabels = nil

	for i := range fields {
		valueLabel := qtlib.NewQLabel2()
		if i == 0 {
			valueLabel.SetObjectName(*qtlib.NewQAnyStringView3("predictedExchange"))
		}
		v.predictedExchangeLabels = append(v.predictedExchangeLabels, valueLabel)
	}
}

func renderAnnotatedCallsignHTML(cs core.AnnotatedCallsign) string {
	var text strings.Builder
	for _, part := range cs.Assembly {
		switch part.OP {
		case core.Matching:
			text.WriteString(part.Value)
		case core.Insert:
			text.WriteString("<i>")
			text.WriteString(part.Value)
			text.WriteString("</i>")
		case core.Delete:
			text.WriteString("<s>")
			text.WriteString(part.Value)
			text.WriteString("</s>")
		case core.Substitute:
			text.WriteString("<u>")
			text.WriteString(part.Value)
			text.WriteString("</u>")
		case core.FalseFriend:
			text.WriteString("<b>")
			text.WriteString(part.Value)
			text.WriteString("</b>")
		}
	}

	result := text.String()
	switch {
	case cs.Duplicate:
		result = fmt.Sprintf(`<span style="color: red;">%s</span>`, result)
	case cs.Worked:
		result = fmt.Sprintf(`<span style="color: palette(accent);">%s</span>`, result)
	}
	return result
}

func renderDXCC(entity dxcc.Prefix, azimuth latlon.Degrees, distance latlon.Km) string {
	if entity.Name == "" {
		return ""
	}
	return fmt.Sprintf("%s (%s), %s, ITU %d, CQ %d, %.0f km, %.0f°",
		entity.Name, entity.PrimaryPrefix, entity.Continent,
		entity.ITUZone, entity.CQZone, float64(distance), float64(azimuth))
}

func renderSupercheckHTML(matches []core.AnnotatedCallsign) string {
	if len(matches) == 0 {
		return ""
	}

	var entries []string
	for i, m := range matches {
		var parts []string
		entryStyle := ""
		if i < 9 {
			tag := fmt.Sprintf(`<span style="color: palette(highlighted-text); background-color: palette(highlight);">&emsp;%d&emsp;</span>`, i+1)
			parts = append(parts, tag)
			entryStyle = "border: 1px solid palette(highlight); border-radius: 3px;"
		}
		parts = append(parts, renderAnnotatedCallsignHTML(m)+"&emsp;")

		entry := fmt.Sprintf(`<td style="%s">%s</td>`, entryStyle, strings.Join(parts, "&nbsp;"))

		entries = append(entries, entry)
	}
	return fmt.Sprintf(`<table border="0" cellspacing="10" cellpadding="0"><tr>%s</tr></table>`, strings.Join(entries, "  "))
}

func renderValue(points int, multis int, value int) string {
	text := fmt.Sprintf("%dP x %dM = %d", points, multis, value)

	switch {
	case points < 1 && multis < 1:
		text = fmt.Sprintf("<s>%s</s>", text)
	case multis > 0:
		text = fmt.Sprintf("<b>%s</b>", text)
	}

	return text
}

func renderQTCStatus(sentQTCs int, receivedQTCs int, enabled bool) string {
	if !enabled {
		log.Printf("QTCs DISABLED")
		return ""
	}

	var text string
	switch {
	case sentQTCs > 0 && receivedQTCs > 0:
		text = fmt.Sprintf("QTCs: %dS %dR", sentQTCs, receivedQTCs)
	case sentQTCs > 0:
		text = fmt.Sprintf("QTCs: %dS", sentQTCs)
	case receivedQTCs > 0:
		text = fmt.Sprintf("QTCs: %dR", receivedQTCs)
	default:
		text = "QTCs: 0"
	}

	log.Println(text)
	return text
}
