package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

const (
	COLUMN_SPACING = 5
	ROW_SPACING    = 5
	MARGIN         = 3
)

func buildHeaderLabel(grid *gtk.Grid, row int, labelText string) *gtk.Label {
	label, _ := gtk.LabelNew("")
	label.SetHAlign(gtk.ALIGN_START)
	label.SetMarginTop(MARGIN)
	label.SetMarginBottom(MARGIN)
	label.SetMarkup(fmt.Sprintf("<b>%s</b>", labelText))
	grid.Attach(label, 0, row, 2, 1)
	return label
}

func buildExplanationLabel(grid *gtk.Grid, row int, labelText string) *gtk.Label {
	label, _ := gtk.LabelNew(labelText)
	label.SetHAlign(gtk.ALIGN_START)
	label.SetMarginTop(MARGIN)
	label.SetMarginBottom(MARGIN)
	grid.Attach(label, 1, row, 1, 1)
	return label
}

func buildSeparator(grid *gtk.Grid, row int, width int) {
	separator, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	separator.SetHExpand(true)
	separator.SetMarginTop(MARGIN)
	separator.SetMarginBottom(MARGIN)
	grid.Attach(separator, 0, row, width, 1)
}

func buildLabeledCombo(grid *gtk.Grid, row int, labelText string, withEntry bool, items []string, handler any) *gtk.ComboBoxText {
	label, _ := gtk.LabelNew(labelText)
	label.SetHAlign(gtk.ALIGN_END)
	label.SetHExpand(false)
	grid.Attach(label, 0, row, 1, 1)

	var combo *gtk.ComboBoxText
	if withEntry {
		combo, _ = gtk.ComboBoxTextNewWithEntry()
	} else {
		combo, _ = gtk.ComboBoxTextNew()
	}
	combo.SetHExpand(true)
	combo.Append("", "")
	for _, item := range items {
		combo.Append(item, item)
	}
	grid.Attach(combo, 1, row, 1, 1)

	combo.Connect("changed", handler)

	return combo
}

func buildLabeledEntry(grid *gtk.Grid, row int, labelText string, handler any) *gtk.Entry {
	label, _ := gtk.LabelNew(labelText)
	label.SetHAlign(gtk.ALIGN_END)
	label.SetHExpand(false)
	grid.Attach(label, 0, row, 1, 1)

	entry, _ := gtk.EntryNew()
	entry.SetHExpand(true)
	entry.SetSizeRequest(200, 0)
	grid.Attach(entry, 1, row, 1, 1)

	if handler != nil {
		entry.Connect("changed", handler)
	} else {
		entry.SetSensitive(false)
	}

	return entry
}

func buildLabeledTextView(grid *gtk.Grid, row int, labelText string, handler any) *gtk.TextView {
	label, _ := gtk.LabelNew(labelText)
	label.SetHAlign(gtk.ALIGN_START)
	grid.Attach(label, 0, row, 1, 1)

	textView, _ := gtk.TextViewNew()

	scrolledWindow, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledWindow.Add(textView)
	scrolledWindow.SetHExpand(true)
	scrolledWindow.SetVExpand(true)
	scrolledWindow.SetSizeRequest(0, 100)
	grid.Attach(scrolledWindow, 0, row+1, 1, 1)

	buffer, _ := textView.GetBuffer()
	buffer.Connect("changed", handler)

	return textView
}

func buildCheckButton(grid *gtk.Grid, row int, labelText string, handler any) *gtk.CheckButton {
	checkButton, _ := gtk.CheckButtonNewWithLabel(labelText)
	checkButton.SetHExpand(true)
	checkButton.SetMarginTop(MARGIN)
	checkButton.SetMarginBottom(MARGIN)
	grid.Attach(checkButton, 0, row, 1, 1)

	checkButton.Connect("toggled", handler)

	return checkButton
}

func buildCheckButtonInColumn(grid *gtk.Grid, row int, col int, width int, labelText string, handler any) *gtk.CheckButton {
	checkButton, _ := gtk.CheckButtonNewWithLabel(labelText)
	checkButton.SetHExpand(true)
	checkButton.SetMarginTop(MARGIN)
	checkButton.SetMarginBottom(MARGIN)
	grid.Attach(checkButton, col, row, width, 1)

	checkButton.Connect("toggled", handler)

	return checkButton
}
