package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type entry struct {
	container *fyne.Container

	// my data
	utcLabel          *widget.Label
	myCallLabel       *widget.Label
	myExchangesParent *fyne.Container
	myExchanges       []*widget.Entry

	// VFO
	vfoNameLabel   *widget.Label
	frequencyLabel *widget.Label
	bandSelect     *widget.Select
	modeSelect     *widget.Select

	// their data
	theirLabel           *widget.Label
	theirCall            *widget.Entry
	theirExchangesParent *fyne.Container
	theirExchanges       []*widget.Entry

	// prediction
	predictedCallLabel       *widget.Label
	predictedExchangesParent *fyne.Container
	predictedExchanges       []*widget.Label
	predictedValueLabel      *widget.Label

	// information
	messageLabel    *widget.Label
	dxccLabel       *widget.Label
	userInfoLabel   *widget.Label
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
	result.theirExchangesParent = container.NewHBox()
	result.logButton = widget.NewButton("Log", result.onLog)
	result.clearButton = widget.NewButton("Clear", result.onClear)

	// entry grid
	myDataRow := container.NewHBox(result.utcLabel, result.myCallLabel, result.myExchangesParent, layout.NewSpacer())
	vfoRow := container.NewHBox(result.vfoNameLabel, result.frequencyLabel, result.bandSelect, result.modeSelect, layout.NewSpacer())
	labelColumn := container.NewVBox(layout.NewSpacer(), result.theirLabel)
	callColumn := container.NewVBox(result.predictedCallLabel, result.theirCall)
	exchangeColumn := container.NewVBox(result.predictedExchangesParent, result.theirExchangesParent)
	valueButtonColumn := container.NewVBox(result.predictedValueLabel, container.NewHBox(result.logButton, result.clearButton))
	entryRow := container.NewHBox(labelColumn, callColumn, exchangeColumn, valueButtonColumn)

	// info
	result.messageLabel = widget.NewLabel("Message")
	result.dxccLabel = widget.NewLabel("Fed. Rep. of Germany (DL), EU, ITU 28, CQ 14")
	result.userInfoLabel = widget.NewLabel("Hans, Salzgitter")
	infoLine := container.NewHBox(result.dxccLabel, layout.NewSpacer(), result.userInfoLabel)
	result.supercheckLabel = widget.NewRichTextWithText("DL1ABC")

	result.container = container.NewVBox(
		myDataRow,
		vfoRow,
		entryRow,
		result.messageLabel,
		infoLine,
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
