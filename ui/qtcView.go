package ui

import (
	"strconv"

	"github.com/gotk3/gotk3/gtk"

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

	for i := range core.MaxQTCsPerCall {
		buildQTCLine(qtcGrid, i, true, nil, nil) // TODO: readOnly and the handlers depend on qtcMode
		// TODO: store the entry widgets
	}

	result.root, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	result.root.Add(headerGrid)
	result.root.Add(separator)
	result.root.Add(qtcGrid)

	return result
}

func buildQTCHeaderLabel(grid *gtk.Grid, column int, text string) {
	result, _ := gtk.LabelNew(text)
	result.SetHExpand(false)
	result.SetHAlign(gtk.ALIGN_CENTER)

	grid.Attach(result, column, 1, 1, 1)
}

func buildQTCLine(grid *gtk.Grid, index int, readOnly bool, okHandler, repeatHandler any) (*gtk.Entry, *gtk.Entry, *gtk.Entry) {
	nrLabel, _ := gtk.LabelNew(strconv.Itoa(index + 1))
	nrLabel.SetHExpand(false)
	nrLabel.SetHAlign(gtk.ALIGN_END)
	grid.Attach(nrLabel, 0, index+2, 1, 1)

	timeEntry, _ := gtk.EntryNew()
	timeEntry.SetHExpand(true)
	timeEntry.SetHAlign(gtk.ALIGN_FILL)
	timeEntry.SetPlaceholderText("Time")
	timeEntry.SetSensitive(!readOnly)
	grid.Attach(timeEntry, 1, index+2, 1, 1)

	callEntry, _ := gtk.EntryNew()
	callEntry.SetHExpand(true)
	callEntry.SetHAlign(gtk.ALIGN_FILL)
	callEntry.SetPlaceholderText("Call")
	callEntry.SetSensitive(!readOnly)
	grid.Attach(callEntry, 2, index+2, 1, 1)

	exchangeEntry, _ := gtk.EntryNew()
	exchangeEntry.SetHExpand(true)
	exchangeEntry.SetHAlign(gtk.ALIGN_FILL)
	exchangeEntry.SetPlaceholderText("Exchange")
	exchangeEntry.SetSensitive(!readOnly)
	grid.Attach(exchangeEntry, 3, index+2, 1, 1)

	okButton, _ := gtk.ButtonNewWithLabel("R")
	okButton.SetHExpand(false)
	okButton.SetHAlign(gtk.ALIGN_END)
	if okHandler != nil {
		okButton.Connect("pressed", okHandler)
	}
	grid.Attach(okButton, 4, index+2, 1, 1)

	repeatButton, _ := gtk.ButtonNewWithLabel("AGN")
	repeatButton.SetHExpand(false)
	repeatButton.SetHAlign(gtk.ALIGN_END)
	if repeatHandler != nil {
		repeatButton.Connect("pressed", repeatHandler)
	}
	grid.Attach(repeatButton, 5, index+2, 1, 1)

	return timeEntry, callEntry, exchangeEntry
}
