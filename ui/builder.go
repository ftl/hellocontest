package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

func buildExplanationLabel(grid *gtk.Grid, row int, labelText string) {
	label, _ := gtk.LabelNew(labelText)
	label.SetHAlign(gtk.ALIGN_START)
	label.SetMarginTop(5)
	label.SetMarginBottom(5)
	grid.Attach(label, 0, row, 2, 1)
}

func buildSeparator(grid *gtk.Grid, row int) {
	separator, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	separator.SetHExpand(true)
	separator.SetMarginTop(5)
	separator.SetMarginBottom(5)
	grid.Attach(separator, 0, row, 2, 1)
}

func buildLabeledCombo(grid *gtk.Grid, row int, labelText string, items []string, handler any) *gtk.ComboBoxText {
	label, _ := gtk.LabelNew(labelText)
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, row, 1, 1)

	combo, _ := gtk.ComboBoxTextNew()
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
	grid.Attach(entry, 1, row, 1, 1)

	entry.Connect("changed", handler)

	return entry
}

func buildCheckButton(grid *gtk.Grid, row int, labelText string, handler any) *gtk.CheckButton {
	checkButton, _ := gtk.CheckButtonNewWithLabel(labelText)
	grid.Attach(checkButton, 0, row, 2, 1)

	checkButton.Connect("toggled", handler)

	return checkButton
}
