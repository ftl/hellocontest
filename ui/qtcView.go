package ui

import (
	"strconv"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

type QTCController interface {
	// TBD
}

type qtcView struct {
	controller QTCController
	mode       core.QTCMode

	// widgets
	root           *gtk.Box
	theirCallEntry *gtk.Entry
	seriesEntry    *gtk.Entry
	qtcRows        []*qtcRow
}

func newQTCView(controller QTCController, mode core.QTCMode) *qtcView {
	result := &qtcView{
		controller: controller,
		mode:       mode,
	}

	headerGrid, _ := gtk.GridNew()
	headerGrid.SetVExpand(false)
	headerGrid.SetColumnSpacing(COLUMN_SPACING)
	headerGrid.SetRowSpacing(ROW_SPACING)
	headerGrid.SetMarginStart(MARGIN)
	headerGrid.SetMarginEnd(MARGIN)

	buildHeaderLabel(headerGrid, 0, "Header")
	result.theirCallEntry = buildLabeledEntry(headerGrid, 1, "Their Call", nil) // TODO: handler or not depends on the qtcMode
	result.theirCallEntry.SetPlaceholderText("Call")
	result.seriesEntry = buildLabeledEntry(headerGrid, 2, "Series", nil) // TODO: handler or not depends on the qtcMode
	result.seriesEntry.SetPlaceholderText("Series/QTC Count")
	sendHeaderButton, _ := gtk.ButtonNewWithLabel("Send Header")
	sendHeaderButton.SetHAlign(gtk.ALIGN_END)
	headerGrid.Attach(sendHeaderButton, 2, 2, 1, 1)

	separator, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	separator.SetVExpand(false)

	qtcGrid, _ := gtk.GridNew()
	qtcGrid.SetVExpand(true)
	qtcGrid.SetColumnSpacing(COLUMN_SPACING)
	qtcGrid.SetRowSpacing(ROW_SPACING)
	qtcGrid.SetMarginStart(MARGIN)
	qtcGrid.SetMarginEnd(MARGIN)

	buildHeaderLabel(qtcGrid, 0, "QTCs")
	buildQTCHeaderLabel(qtcGrid, 0, "Nr.")
	buildQTCHeaderLabel(qtcGrid, 1, "Time")
	buildQTCHeaderLabel(qtcGrid, 2, "Call")
	buildQTCHeaderLabel(qtcGrid, 3, "Exch.")
	buildQTCHeaderLabel(qtcGrid, 4, "Action")

	result.qtcRows = make([]*qtcRow, core.MaxQTCsPerCall)
	for i := range result.qtcRows {
		row := newQTCRow(qtcGrid, i, true)
		// TODO set input handler callback functions
		result.qtcRows[i] = row
	}

	result.root, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	result.root.Add(headerGrid)
	result.root.Add(separator)
	result.root.Add(qtcGrid)

	return result
}

func (v *qtcView) setHeader(theirCall callsign.Callsign, qtcHeader core.QTCHeader) {
	v.theirCallEntry.SetText(theirCall.String())
	v.seriesEntry.SetText(qtcHeader.String())
}

func (v *qtcView) setQTCs(qtcs []core.QTC) {
	for i, row := range v.qtcRows {
		if i >= len(qtcs) {
			row.hide()
			continue
		}
		v.setQTC(i, qtcs[i])
		row.show()
	}
}

func (v *qtcView) setQTC(index int, qtc core.QTC) {
	if index < 0 || index >= len(v.qtcRows) {
		return
	}
	row := v.qtcRows[index]
	row.timeEntry.SetText(qtc.QTCTime.String())
	row.callEntry.SetText(qtc.QTCCallsign.String())
	row.exchangeEntry.SetText(qtc.QTCNumber.String())
}

func (v *qtcView) focusHeader() {

}

func (v *qtcView) focusQTC(index int) {
	if index < 0 || index >= len(v.qtcRows) {
		return
	}

}

func buildQTCHeaderLabel(grid *gtk.Grid, column int, text string) {
	result, _ := gtk.LabelNew(text)
	result.SetHExpand(false)
	result.SetHAlign(gtk.ALIGN_CENTER)

	grid.Attach(result, column, 1, 1, 1)
}

type qtcRow struct {
	nrLabel       *gtk.Label
	timeEntry     *gtk.Entry
	callEntry     *gtk.Entry
	exchangeEntry *gtk.Entry
	okButton      *gtk.Button
	repeatButton  *gtk.Button
}

func newQTCRow(grid *gtk.Grid, index int, readOnly bool) *qtcRow {
	row := index + 2

	nr := strconv.Itoa(index + 1)
	nrLabel, _ := gtk.LabelNew(nr)
	nrLabel.SetHExpand(false)
	nrLabel.SetHAlign(gtk.ALIGN_END)
	grid.Attach(nrLabel, 0, row, 1, 1)

	timeEntry, _ := gtk.EntryNew()
	timeEntry.SetHExpand(true)
	timeEntry.SetHAlign(gtk.ALIGN_FILL)
	timeEntry.SetPlaceholderText("Time")
	timeEntry.SetSensitive(!readOnly)
	grid.Attach(timeEntry, 1, row, 1, 1)

	callEntry, _ := gtk.EntryNew()
	callEntry.SetHExpand(true)
	callEntry.SetHAlign(gtk.ALIGN_FILL)
	callEntry.SetPlaceholderText("Call")
	callEntry.SetSensitive(!readOnly)
	grid.Attach(callEntry, 2, row, 1, 1)

	exchangeEntry, _ := gtk.EntryNew()
	exchangeEntry.SetHExpand(true)
	exchangeEntry.SetHAlign(gtk.ALIGN_FILL)
	exchangeEntry.SetPlaceholderText("Exchange")
	exchangeEntry.SetSensitive(!readOnly)
	grid.Attach(exchangeEntry, 3, row, 1, 1)

	okButton, _ := gtk.ButtonNewWithLabel("R")
	okButton.SetHExpand(false)
	okButton.SetHAlign(gtk.ALIGN_END)
	grid.Attach(okButton, 4, row, 1, 1)

	repeatButton, _ := gtk.ButtonNewWithLabel("AGN")
	repeatButton.SetHExpand(false)
	repeatButton.SetHAlign(gtk.ALIGN_END)
	grid.Attach(repeatButton, 5, row, 1, 1)

	return &qtcRow{
		nrLabel:       nrLabel,
		timeEntry:     timeEntry,
		callEntry:     callEntry,
		exchangeEntry: exchangeEntry,
		okButton:      okButton,
		repeatButton:  repeatButton,
	}
}

func (r *qtcRow) hide() {
	r.nrLabel.Hide()
	r.timeEntry.Hide()
	r.callEntry.Hide()
	r.exchangeEntry.Hide()
	r.okButton.Hide()
	r.repeatButton.Hide()
}

func (r *qtcRow) show() {
	r.nrLabel.Show()
	r.timeEntry.Show()
	r.callEntry.Show()
	r.exchangeEntry.Show()
	r.okButton.Show()
	r.repeatButton.Show()
}
