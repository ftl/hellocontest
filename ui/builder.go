package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

func buildHeaderLabel(grid *gtk.Grid, row int, labelText string) *gtk.Label {
	label, _ := gtk.LabelNew("")
	label.SetHAlign(gtk.ALIGN_START)
	label.SetMarginTop(5)
	label.SetMarginBottom(5)
	label.SetMarkup(fmt.Sprintf("<b>%s</b>", labelText))
	grid.Attach(label, 0, row, 2, 1)
	return label
}

func buildExplanationLabel(grid *gtk.Grid, row int, labelText string) *gtk.Label {
	label, _ := gtk.LabelNew(labelText)
	label.SetHAlign(gtk.ALIGN_START)
	label.SetMarginTop(5)
	label.SetMarginBottom(5)
	grid.Attach(label, 1, row, 1, 1)
	return label
}

func buildSeparator(grid *gtk.Grid, row int) {
	separator, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	separator.SetHExpand(true)
	separator.SetMarginTop(5)
	separator.SetMarginBottom(5)
	grid.Attach(separator, 0, row, 2, 1)
}

func buildLabeledCombo(grid *gtk.Grid, row int, labelText string, withEntry bool, items []string, handler any) *gtk.ComboBoxText {
	label, _ := gtk.LabelNew(labelText)
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, row, 1, 1)

	var combo *gtk.ComboBoxText
	if withEntry {
		combo, _ = gtk.ComboBoxTextNewWithEntry()
	} else {
		combo, _ = gtk.ComboBoxTextNew()
	}
	combo.SetHExpand(true)
	combo.RemoveAll()
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
	grid.Attach(label, 0, row, 1, 1)

	entry, _ := gtk.EntryNew()
	entry.SetHExpand(true)
	grid.Attach(entry, 1, row, 1, 1)

	entry.Connect("changed", handler)

	return entry
}

func buildCheckButton(grid *gtk.Grid, row int, labelText string, handler any) *gtk.CheckButton {
	checkButton, _ := gtk.CheckButtonNewWithLabel(labelText)
	checkButton.SetHExpand(true)
	grid.Attach(checkButton, 0, row, 2, 1)

	checkButton.Connect("toggled", handler)

	return checkButton
}
