package ui

import (
	"fmt"
	"html"
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
	dxccLabel               *qtlib.QLabel
	userInfoLabel           *qtlib.QLabel
	supercheckLabel         *qtlib.QLabel
	predictedExchangeLabels []*qtlib.QLabel
}

type serialClaim struct {
	serial    core.QSONumber
	committed bool
}

// callinfoView handles interaction and content related to core/callinfo. It owns all related widgets.
type callinfoView struct {
	vfo [core.VFOCount]callinfoVFOWidgets

	qtcsEnabled bool

	area *entryArea

	serialFieldIndex int
	serialClaims     [core.VFOCount]serialClaim

	current [core.VFOCount]core.CallinfoFrame
}

func newCallinfoVFOWidgets() callinfoVFOWidgets {
	w := callinfoVFOWidgets{}

	w.callsignLabel = qtlib.NewQLabel2()
	w.valueLabel = qtlib.NewQLabel2()
	w.qtcStatusLabel = qtlib.NewQLabel2()
	w.supercheckLabel = qtlib.NewQLabel2()
	w.dxccLabel = qtlib.NewQLabel2()
	w.userInfoLabel = qtlib.NewQLabel2()

	return w
}

func newCallinfoView() *callinfoView {
	v := &callinfoView{
		serialFieldIndex: -1,
	}

	v.vfo[core.VFO1] = newCallinfoVFOWidgets()
	v.vfo[core.VFO2] = newCallinfoVFOWidgets()

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
	w.callsignLabel.SetText(renderCallsignOnFrequencyHTML(best, cur.MarkerOnFrequency))
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
	v.refreshSerialClaim(vfo)
}

func (v *callinfoView) SetSerialClaim(vfo core.VFOID, serial core.QSONumber, committed bool) {
	v.serialClaims[vfo] = serialClaim{serial: serial, committed: committed}
	v.refreshSerialClaim(vfo)
}

func (v *callinfoView) refreshSerialClaim(vfo core.VFOID) {
	w := &v.vfo[vfo]
	if v.serialFieldIndex < 0 || v.serialFieldIndex >= len(w.predictedExchangeLabels) {
		return
	}
	w.predictedExchangeLabels[v.serialFieldIndex].SetText(renderSerialClaim(v.serialClaims[vfo]))
}

func (v *callinfoView) SetPredictedExchangeFields(fields []core.ExchangeField) {
	v.serialFieldIndex = serialFieldIndex(fields)
	for vfo := range core.VFOCount {
		w := &v.vfo[vfo]
		for _, label := range w.predictedExchangeLabels {
			label.Delete()
		}
		w.predictedExchangeLabels = nil

		for range fields {
			w.predictedExchangeLabels = append(w.predictedExchangeLabels, qtlib.NewQLabel2())
		}
		v.refreshSerialClaim(vfo)
	}
}

func serialFieldIndex(fields []core.ExchangeField) int {
	for i, field := range fields {
		if field.CanContainSerial {
			return i
		}
	}
	return -1
}

func renderSerialClaim(claim serialClaim) string {
	switch {
	case claim.serial == 0:
		return ""
	case claim.committed:
		return fmt.Sprintf("<b>%s committed</b>", claim.serial.String())
	default:
		return fmt.Sprintf("%s claimed", claim.serial.String())
	}
}

func (v *callinfoView) SetVFOEnabled(vfo core.VFOID, enabled bool) {
	if vfo == core.VFO1 {
		return
	}
	v.area.setVFO2Enabled(enabled)
}

const (
	markerPrefix         = "\U0001F782"
	supercheckSpacerCell = "<td>&nbsp;&nbsp;&nbsp;</td>"
)

func renderCallsignOnFrequencyHTML(cs core.AnnotatedCallsign, markerOnFrequency string) string {
	callsign := renderAnnotatedCallsignHTML(cs)
	if markerOnFrequency == "" {
		return callsign
	}

	marker := markerPrefix + " " + html.EscapeString(markerOnFrequency)
	if callsign == "" {
		return marker
	}
	return marker + " | " + callsign
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
	return fmt.Sprintf(`<table border="0" cellspacing="0" cellpadding="0"><tr>%s</tr></table>`, strings.Join(entries, supercheckSpacerCell))
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
