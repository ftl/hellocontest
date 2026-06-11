package ui

import (
	"fmt"
	"strings"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/latlon"

	"github.com/ftl/hellocontest/core"
)

type callinfoVFOWidgets struct {
	callsignLabel           *qtlib.QLabel
	valueLabel              *qtlib.QLabel
	qtcStatusLabel          *qtlib.QLabel
	infoContainer           *qtlib.QWidget
	dxccLabel               *qtlib.QLabel
	userInfoLabel           *qtlib.QLabel
	supercheckLabel         *qtlib.QLabel
	predictedExchangeLabels []*qtlib.QLabel
}

type callinfoView struct {
	vfo [core.VFOCount]callinfoVFOWidgets

	qtcsEnabled bool
	vfo2Enabled bool

	current [core.VFOCount]core.CallinfoFrame
}

func newCallinfoVFOWidgets(prefix string) callinfoVFOWidgets {
	w := callinfoVFOWidgets{}

	w.callsignLabel = qtlib.NewQLabel2()
	w.callsignLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "predictedBestMatch"))
	w.callsignLabel.SetTextFormat(qtlib.RichText)

	w.valueLabel = qtlib.NewQLabel2()
	w.valueLabel.SetTextFormat(qtlib.RichText)
	w.valueLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "predictedValue"))

	w.qtcStatusLabel = qtlib.NewQLabel2()
	w.qtcStatusLabel.SetTextFormat(qtlib.RichText)
	w.qtcStatusLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "qtcStatus"))
	w.qtcStatusLabel.SetAlignment(qtlib.AlignVCenter | qtlib.AlignTrailing)

	w.supercheckLabel = qtlib.NewQLabel2()
	w.supercheckLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "supercheckLabel"))
	w.supercheckLabel.SetTextFormat(qtlib.RichText)
	w.supercheckLabel.QWidget.SetSizePolicy2(qtlib.QSizePolicy__Ignored, qtlib.QSizePolicy__Preferred)
	w.supercheckLabel.SetMinimumWidth(0)

	w.dxccLabel = qtlib.NewQLabel2()
	w.dxccLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "dxccLabel"))
	w.dxccLabel.QWidget.SetSizePolicy2(qtlib.QSizePolicy__Ignored, qtlib.QSizePolicy__Preferred)
	w.dxccLabel.SetMinimumWidth(0)

	w.userInfoLabel = qtlib.NewQLabel2()
	w.userInfoLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "userInfoLabel"))
	w.userInfoLabel.QWidget.SetSizePolicy2(qtlib.QSizePolicy__Ignored, qtlib.QSizePolicy__Preferred)
	w.userInfoLabel.SetMinimumWidth(0)
	w.userInfoLabel.SetAlignment(qtlib.AlignVCenter | qtlib.AlignTrailing)

	w.infoContainer = qtlib.NewQWidget2()
	infoLayout := qtlib.NewQHBoxLayout(w.infoContainer)
	infoLayout.SetContentsMargins(0, 0, 0, 0)
	infoLayout.AddWidget(w.dxccLabel.QWidget)
	infoLayout.AddWidget(w.userInfoLabel.QWidget)

	return w
}

func newCallinfoView() *callinfoView {
	v := &callinfoView{}

	v.vfo[core.VFO1] = newCallinfoVFOWidgets("vfo1")
	v.vfo[core.VFO2] = newCallinfoVFOWidgets("vfo2")

	// VFO2 widgets initially hidden
	v.SetVFOEnabled(core.VFO2, false)

	return v
}

func (v *callinfoView) ShowFrame(vfo core.VFOID, frame core.CallinfoFrame) {
	v.current[vfo] = frame
	v.refreshVFO(vfo)
}

func (v *callinfoView) SetQTCsEnabled(enabled bool) {
	v.qtcsEnabled = enabled
	for vfo := range core.VFOCount {
		v.refreshVFO(vfo)
	}
}

func (v *callinfoView) refreshVFO(vfo core.VFOID) {
	w := &v.vfo[vfo]
	cur := &v.current[vfo]

	best := cur.BestMatchOnFrequency()
	w.callsignLabel.SetText(renderAnnotatedCallsignHTML(best))
	w.dxccLabel.SetText(renderDXCC(cur.DXCCEntity, cur.Azimuth, cur.Distance))
	w.valueLabel.SetText(renderValue(cur.Points, cur.Multis, cur.Value))
	w.qtcStatusLabel.SetText(renderQTCStatus(cur.SentQTCs, cur.ReceivedQTCs, v.qtcsEnabled))
	w.userInfoLabel.SetText(cur.UserInfo)
	w.supercheckLabel.SetText(renderSupercheckHTML(cur.Supercheck))

	for i, lbl := range w.predictedExchangeLabels {
		if i < len(cur.PredictedExchange) {
			lbl.SetText(strings.TrimSpace(cur.PredictedExchange[i]))
		} else {
			lbl.SetText("-")
		}
	}
}

func (v *callinfoView) SetPredictedExchangeFields(fields []core.ExchangeField) {
	for vfo := range core.VFOCount {
		w := &v.vfo[vfo]
		for _, label := range w.predictedExchangeLabels {
			label.Delete()
		}
		w.predictedExchangeLabels = nil

		prefix := "vfo1"
		if vfo == core.VFO2 {
			prefix = "vfo2"
		}
		for i := range fields {
			valueLabel := qtlib.NewQLabel2()
			if i == 0 {
				valueLabel.SetObjectName(*qtlib.NewQAnyStringView3(prefix + "predictedExchange"))
			}
			if vfo == core.VFO2 {
				valueLabel.SetVisible(v.vfo2Enabled)
			}
			w.predictedExchangeLabels = append(w.predictedExchangeLabels, valueLabel)
		}
	}
}

func (v *callinfoView) SetVFOEnabled(vfo core.VFOID, enabled bool) {
	if vfo == core.VFO1 {
		return
	}
	v.vfo2Enabled = enabled
	w := &v.vfo[core.VFO2]
	w.callsignLabel.SetVisible(enabled)
	w.valueLabel.SetVisible(enabled)
	w.qtcStatusLabel.SetVisible(enabled)
	w.infoContainer.SetVisible(enabled)
	w.supercheckLabel.SetVisible(enabled)
	for _, lbl := range w.predictedExchangeLabels {
		lbl.SetVisible(enabled)
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

	return text
}
