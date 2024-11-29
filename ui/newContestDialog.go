//go:build !fyne

package ui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
)

type newContestDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller         NewContestController
	contestIdentifiers []string
	contestLabels      []string

	*newContestView
}

func setupNewContestDialog(parent gtk.IWidget, controller NewContestController) *newContestDialog {
	result := &newContestDialog{
		parent:     parent,
		controller: controller,
	}
	return result
}

func (d *newContestDialog) onDestroy() {
	d.dialog = nil
	d.newContestView = nil
}

func (d *newContestDialog) Show() bool {
	if d.dialog != nil {
		log.Print("new contest dialog already open")
		return false
	}

	builder := setupBuilder()
	d.dialog = getUI(builder, "newContestDialog").(*gtk.Dialog)
	d.dialog.SetTransientFor(nil)
	d.dialog.SetPosition(gtk.WIN_POS_CENTER)
	d.dialog.Connect("destroy", d.onDestroy)
	d.newContestView = setupNewContestView(builder, d.dialog, d.controller, d.contestIdentifiers, d.contestLabels)

	result := d.dialog.Run() == gtk.RESPONSE_OK
	d.dialog.Close()
	d.dialog.Destroy()
	d.dialog = nil

	return result
}

func (d *newContestDialog) SetContestIdentifiers(ids []string, labels []string) {
	d.contestIdentifiers = ids
	d.contestLabels = labels
}
