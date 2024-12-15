package fyneui

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/ftl/hellocontest/core"
	core_callinfo "github.com/ftl/hellocontest/core/callinfo"
	core_entry "github.com/ftl/hellocontest/core/entry"
)

var _ core_entry.View = (*entry)(nil)
var _ core_callinfo.View = (*entry)(nil)

type entry struct {
	container *fyne.Container
	canvas    func() fyne.Canvas

	// my data
	utcLabel          *widget.Label
	myCallLabel       *widget.Label
	myExchangesParent *fyne.Container
	myExchanges       []fyne.CanvasObject

	// VFO
	vfoNameLabel   *widget.Label
	frequencyLabel *widget.Label
	bandSelect     *widget.Select
	modeSelect     *widget.Select

	// their data
	theirLabel           *widget.Label
	theirCall            *widget.Entry
	theirExchangesParent *fyne.Container
	theirExchanges       []fyne.CanvasObject

	// prediction
	predictedCallLabel       *widget.RichText
	predictedExchangesParent *fyne.Container
	predictedExchanges       []fyne.CanvasObject
	predictedValueLabel      *widget.RichText

	// information
	messageLabel    *widget.Label
	dxccLabel       *widget.RichText
	userInfoLabel   *widget.Label
	infoLine        *fyne.Container
	supercheckLabel *widget.RichText

	// buttons
	logButton   *widget.Button
	clearButton *widget.Button
}

func setupEntry(canvas func() fyne.Canvas) *entry {
	result := &entry{
		canvas: canvas,
	}

	// my data row
	result.utcLabel = widget.NewLabel("00:00")
	result.myCallLabel = widget.NewLabel("DB0ABC")
	result.myExchangesParent = container.NewHBox()

	// vfo row
	result.vfoNameLabel = widget.NewLabel("VFO:")
	result.frequencyLabel = widget.NewLabel("3.500 kHz")
	result.frequencyLabel.Alignment = fyne.TextAlignTrailing
	result.bandSelect = widget.NewSelect([]string{}, result.onBandSelect)
	result.modeSelect = widget.NewSelect([]string{}, result.onModeSelect)

	// entry row: predcition
	result.predictedCallLabel = widget.NewRichText()
	result.predictedExchangesParent = container.NewHBox()
	result.predictedValueLabel = widget.NewRichText()

	// entry row: input
	result.theirLabel = widget.NewLabel("Their:")
	result.theirCall = widget.NewEntry()
	result.theirCall.PlaceHolder = "Call"
	result.theirExchangesParent = container.NewHBox()
	result.logButton = widget.NewButton("Log", result.onLog)
	result.clearButton = widget.NewButton("Clear", result.onClear)

	// entry grid
	myDataRow := container.NewHBox(result.utcLabel, result.myCallLabel, result.myExchangesParent)
	vfoRow := container.NewHBox(result.vfoNameLabel, result.frequencyLabel, result.bandSelect, result.modeSelect)
	labelColumn := container.NewVBox(layout.NewSpacer(), result.theirLabel)
	callColumn := container.NewVBox(result.predictedCallLabel, result.theirCall)
	exchangeColumn := container.NewVBox(result.predictedExchangesParent, result.theirExchangesParent)
	valueButtonColumn := container.NewVBox(result.predictedValueLabel, container.NewHBox(result.logButton, result.clearButton))
	entryRow := container.NewHBox(labelColumn, callColumn, exchangeColumn, valueButtonColumn)

	// info
	result.messageLabel = widget.NewLabel("")
	result.messageLabel.Hide()
	result.dxccLabel = widget.NewRichText()
	result.userInfoLabel = widget.NewLabel("")
	result.infoLine = container.NewHBox(result.dxccLabel, layout.NewSpacer(), result.userInfoLabel)
	result.supercheckLabel = widget.NewRichText()

	result.container = container.NewVBox(
		myDataRow,
		vfoRow,
		entryRow,
		result.messageLabel,
		result.infoLine,
		result.supercheckLabel,
	)

	return result
}

func (e *entry) onLog() {
	// TODO implement
}

func (e *entry) onClear() {
	// TODO implement
}

func (e *entry) onBandSelect(bandLabel string) {
	// TODO implement
}

func (e *entry) onModeSelect(modeLabel string) {
	// TODO implement
}

func (e *entry) SetMyExchangeFields(fields []core.ExchangeField) {
	e.setupExchangeEntry(fields, e.myExchangesParent, &e.myExchanges)
}

func (e *entry) SetTheirExchangeFields(fields []core.ExchangeField) {
	e.setupExchangeEntry(fields, e.theirExchangesParent, &e.theirExchanges)
}

func (e *entry) setupExchangeEntry(fields []core.ExchangeField, parent *fyne.Container, entries *[]fyne.CanvasObject) {
	parent.RemoveAll()

	*entries = make([]fyne.CanvasObject, len(fields))
	for i, field := range fields {
		entry := widget.NewEntry()
		entry.SetPlaceHolder(field.Short)
		entry.Resize(fyne.NewSize(200, 0))
		(*entries)[i] = entry
		parent.Add(entry)
		// TODO add event handler
	}
}

func (e *entry) SetPredictedExchangeFields(fields []core.ExchangeField) {
	e.setupExchangeLabels(fields, e.predictedExchangesParent, &e.predictedExchanges)
}

func (e *entry) setupExchangeLabels(fields []core.ExchangeField, parent *fyne.Container, labels *[]fyne.CanvasObject) {
	parent.RemoveAll()

	*labels = make([]fyne.CanvasObject, len(fields))
	for i := range fields {
		label := widget.NewLabel("")
		(*labels)[i] = label
		parent.Add(label)
	}
}

func (e *entry) ShowMessage(args ...any) {
	e.messageLabel.SetText(fmt.Sprint(args...))
	e.messageLabel.Show()
	e.infoLine.Hide()
}

func (e *entry) ClearMessage() {
	e.messageLabel.SetText("")
	e.messageLabel.Hide()
	e.infoLine.Show()
}

func (e *entry) SelectText(field core.EntryField, s string) {
	entry := e.fieldToEntry(field)
	if entry == nil {
		return
	}
	text := entry.Text
	index := strings.Index(strings.ToUpper(text), strings.ToUpper(s))
	if index == -1 {
		return
	}
	// TODO: select text in the Entry widget
}

func (e *entry) SetActiveField(field core.EntryField) {
	w := e.fieldToWidget(field)
	if f, ok := w.(fyne.Focusable); ok {
		e.canvas().Focus(f)
	}
}

func (e *entry) SetBand(bandLabel string) {
	e.bandSelect.Selected = bandLabel
}

func (e *entry) SetMode(modeLabel string) {
	e.modeSelect.Selected = modeLabel
}

func (e *entry) SetCallsign(callsign string) {
	e.theirCall.SetText(callsign)
}

func (e *entry) SetDuplicateMarker(bool) {
	// TODO: implement
}

func (e *entry) SetEditingMarker(bool) {
	// TODO: implement
}

func (e *entry) SetFrequency(frequency core.Frequency) {
	e.frequencyLabel.Text = frequency.LabelString()
}

func (e *entry) SetMyCall(call string) {
	e.myCallLabel.SetText(call)
}

func (e *entry) SetMyExchange(index int, text string) {
	i := index - 1
	if i < 0 || i >= len(e.myExchanges) {
		return
	}
	e.myExchanges[i].(*widget.Entry).SetText(text)
}

func (e *entry) SetTheirExchange(index int, text string) {
	i := index - 1
	if i < 0 || i >= len(e.theirExchanges) {
		return
	}
	e.theirExchanges[i].(*widget.Entry).SetText(text)
}

func (e *entry) SetPredictedExchange(index int, text string) {
	i := index - 1
	if i < 0 || i >= len(e.predictedExchanges) {
		return
	}
	if text == "" {
		text = "-"
	} else {
		text = strings.TrimSpace(text)
	}
	e.predictedExchanges[i].(*widget.Label).SetText(text)
}

func (e *entry) SetUTC(utc string) {
	e.utcLabel.SetText(utc)
}

func (e *entry) SetBestMatchingCallsign(callsign core.AnnotatedCallsign) {
	e.predictedCallLabel.Segments = e.renderCallsign(callsign)
}

func (e *entry) SetDXCC(dxccName, continent string, itu, cq int, arrlCompliant bool) {
	text := fmt.Sprintf("%s, %s", dxccName, continent)
	if itu != 0 {
		text += fmt.Sprintf(", ITU %d", itu)
	}
	if cq != 0 {
		text += fmt.Sprintf(", CQ %d", cq)
	}
	segments := []widget.RichTextSegment{
		&widget.TextSegment{Text: text},
	}
	if dxccName != "" && !arrlCompliant {
		segments = append(segments, &widget.TextSegment{
			Text: ", not ARRL compliant",
			Style: widget.RichTextStyle{
				ColorName: theme.ColorRed,
				TextStyle: fyne.TextStyle{
					Bold: true,
				},
			},
		})
	}
	e.dxccLabel.Segments = segments
}

func (e *entry) SetSupercheck(callsigns []core.AnnotatedCallsign) {
	var segments []widget.RichTextSegment
	for i, callsign := range callsigns {
		if len(segments) > 0 {
			segments = append(segments, &widget.TextSegment{Text: "|"})
		}
		if i < 9 {
			text := &widget.TextSegment{
				Text: fmt.Sprintf("(%d) ", i+1),
			}
			segments = append(segments, text)
		}
		segments = append(segments, e.renderCallsign(callsign)...)
	}
	e.supercheckLabel.Segments = segments
}

func (e *entry) SetUserInfo(text string) {
	e.userInfoLabel.SetText(text)
}

func (e *entry) SetValue(points int, multis int, value int) {
	segment := &widget.TextSegment{
		Text: fmt.Sprintf("%dP x %dM = %d", points, multis, value),
	}
	switch {
	case points < 1 && multis < 1:
		segment.Style.ColorName = theme.ColorNameDisabled
	case multis > 0:
		segment.Style.TextStyle.Bold = true
	}
	e.predictedValueLabel.Segments = []widget.RichTextSegment{segment}
}

func (e *entry) renderCallsign(callsign core.AnnotatedCallsign) []widget.RichTextSegment {
	// TODO: visualize the annotations
	return []widget.RichTextSegment{
		&widget.TextSegment{
			Text: callsign.Callsign.String(),
		},
	}
}

func (e *entry) fieldToWidget(field core.EntryField) fyne.CanvasObject {
	switch field {
	case core.CallsignField:
		return e.theirCall
	case core.BandField:
		return e.bandSelect
	case core.ModeField:
		return e.modeSelect
	case core.OtherField:
		return e.theirCall
	}
	switch {
	case field.IsMyExchange():
		i := field.ExchangeIndex() - 1
		return e.myExchanges[i]
	case field.IsTheirExchange():
		i := field.ExchangeIndex() - 1
		return e.theirExchanges[i]
	default:
		log.Fatalf("Unknown entry field %s", field)
	}
	panic("this is never reached")
}

func (e *entry) fieldToEntry(field core.EntryField) *widget.Entry {
	switch field {
	case core.CallsignField:
		return e.theirCall
	case core.OtherField:
		return e.theirCall
	}
	switch {
	case field.IsMyExchange():
		i := field.ExchangeIndex() - 1
		return e.myExchanges[i].(*widget.Entry)
	case field.IsTheirExchange():
		i := field.ExchangeIndex() - 1
		return e.theirExchanges[i].(*widget.Entry)
	}
	return nil
}
