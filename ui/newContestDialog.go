package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type newContestDialog struct {
	dialog *gtk.Dialog

	controller         NewContestController
	contestIdentifiers []string
	contestLabels      []string

	*newContestView
}

func setupNewContestDialog(controller NewContestController) *newContestDialog {
	result := &newContestDialog{
		controller: controller,
	}
	return result
}

func (d *newContestDialog) onDestroy() {
	d.dialog = nil
	d.newContestView = nil
}

func (d *newContestDialog) Show() bool {
	if d.dialog == nil {
		builder := setupBuilder()
		d.dialog = getUI(builder, "newContestDialog").(*gtk.Dialog)
		d.dialog.Connect("destroy", d.onDestroy)
		d.newContestView = setupNewContestView(builder, d.dialog, d.controller, d.contestIdentifiers, d.contestLabels)
	}
	d.dialog.ShowAll()
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
