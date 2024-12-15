package fyneui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/ftl/hellocontest/core"
	core_callinfo "github.com/ftl/hellocontest/core/callinfo"
	core_entry "github.com/ftl/hellocontest/core/entry"
)

var _ core_entry.View = (*entry)(nil)
var _ core_callinfo.View = (*entry)(nil)

type entry struct {
	container *fyne.Container

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
	predictedCallLabel       *widget.Label
	predictedExchangesParent *fyne.Container
	predictedExchanges       []fyne.CanvasObject
	predictedValueLabel      *widget.Label

	// information
	messageLabel    *widget.Label
	dxccLabel       *widget.Label
	userInfoLabel   *widget.Label
	infoLine        *fyne.Container
	supercheckLabel *widget.RichText

	// buttons
	logButton   *widget.Button
	clearButton *widget.Button
}

func setupEntry() *entry {
	result := &entry{}

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
	result.predictedCallLabel = widget.NewLabel("DL1ABC")
	result.predictedExchangesParent = container.NewHBox()
	result.predictedValueLabel = widget.NewLabel("0P 0M = 0")

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
	result.dxccLabel = widget.NewLabel("Fed. Rep. of Germany (DL), EU, ITU 28, CQ 14")
	result.userInfoLabel = widget.NewLabel("Hans, Salzgitter")
	result.infoLine = container.NewHBox(result.dxccLabel, layout.NewSpacer(), result.userInfoLabel)
	result.supercheckLabel = widget.NewRichTextWithText("DL1ABC")

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
		// TODO add event handler
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

func (e *entry) SelectText(core.EntryField, string) {
	// TODO: implement
}

func (e *entry) SetActiveField(core.EntryField) {
	// TODO: implement
}

func (e *entry) SetBand(text string) {
	// TODO: implement
}

func (e *entry) SetCallsign(string) {
	// TODO: implement
}

func (e *entry) SetDuplicateMarker(bool) {
	// TODO: implement
}

func (e *entry) SetEditingMarker(bool) {
	// TODO: implement
}

func (e *entry) SetFrequency(core.Frequency) {
	// TODO: implement
}

func (e *entry) SetMode(text string) {
	// TODO: implement
}

func (e *entry) SetMyCall(call string) {
	e.myCallLabel.SetText(call)
}

func (e *entry) SetMyExchange(int, string) {
	// TODO: implement
}

func (e *entry) SetTheirExchange(int, string) {
	// TODO: implement
}

func (e *entry) SetUTC(utc string) {
	e.utcLabel.SetText(utc)
}

func (e *entry) SetBestMatchingCallsign(callsign core.AnnotatedCallsign) {
	// TODO: implement
}

func (e *entry) SetDXCC(string, string, int, int, bool) {
	// TODO: implement
}

func (e *entry) SetPredictedExchange(index int, text string) {
	// TODO: implement
}

func (e *entry) SetSupercheck(callsigns []core.AnnotatedCallsign) {
	// TODO: implement
}

func (e *entry) SetUserInfo(string) {
	// TODO: implement
}

func (e *entry) SetValue(points int, multis int, value int) {
	// TODO: implement
}
