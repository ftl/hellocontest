package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core/newcontest"
)

var _ newcontest.View = (*newContestDialog)(nil)

// NewContestController is the callback surface the dialog uses.
type NewContestController interface {
	SelectContestIdentifier(string)
	EnterContestName(string)
}

type newContestDialog struct {
	parent     *qtlib.QMainWindow
	controller NewContestController

	dialog       *qtlib.QDialog
	contestCombo *qtlib.QComboBox
	nameEdit     *qtlib.QLineEdit
	okButton     *qtlib.QPushButton

	ids                []string
	texts              []string
	ignoreChangedEvent bool
}

func newNewContestDialog(parent *qtlib.QMainWindow, controller NewContestController) *newContestDialog {
	return &newContestDialog{parent: parent, controller: controller}
}

func (d *newContestDialog) Show() bool {
	d.dialog = qtlib.NewQDialog(d.parent.QWidget)
	d.dialog.SetWindowTitle("New Contest")
	d.dialog.SetModal(true)
	d.dialog.SetWindowFlags(
		qtlib.Window |
			qtlib.CustomizeWindowHint |
			qtlib.WindowTitleHint |
			qtlib.WindowCloseButtonHint,
	)

	root := qtlib.NewQVBoxLayout(d.dialog.QWidget)

	form := qtlib.NewQGridLayout2()
	d.contestCombo = makeFilterableCombo()
	d.nameEdit = qtlib.NewQLineEdit2()

	form.AddWidget2(qtlib.NewQLabel3("Contest:").QWidget, 0, 0)
	form.AddWidget2(d.contestCombo.QWidget, 0, 1)
	form.AddWidget2(qtlib.NewQLabel3("Name:").QWidget, 1, 0)
	form.AddWidget2(d.nameEdit.QWidget, 1, 1)
	root.AddLayout(form.QLayout)

	buttons := qtlib.NewQDialogButtonBox4(
		qtlib.QDialogButtonBox__Ok | qtlib.QDialogButtonBox__Cancel,
	)
	d.okButton = buttons.Button(qtlib.QDialogButtonBox__Ok)
	d.okButton.SetText("Create New Contest")
	d.okButton.SetEnabled(false)
	buttons.OnAccepted(func() { d.dialog.Accept() })
	buttons.OnRejected(func() { d.dialog.Reject() })
	root.AddWidget(buttons.QWidget)

	// Populate combo with texts received prior to Show (the controller pushes
	// SetContestIdentifiers before calling Show).
	d.ignoreChangedEvent = true
	for _, t := range d.texts {
		d.contestCombo.AddItem(t)
	}
	d.ignoreChangedEvent = false

	d.contestCombo.OnActivated(func(index int) {
		if d.ignoreChangedEvent || index < 0 || index >= len(d.ids) {
			return
		}
		d.controller.SelectContestIdentifier(d.ids[index])
	})
	d.nameEdit.OnTextChanged(func(text string) {
		if d.ignoreChangedEvent {
			return
		}
		d.controller.EnterContestName(text)
	})

	accepted := d.dialog.Exec() == int(qtlib.QDialog__Accepted)
	d.dialog.DeleteLater()
	d.dialog = nil
	d.contestCombo = nil
	d.nameEdit = nil
	d.okButton = nil
	return accepted
}

func (d *newContestDialog) SetContestIdentifiers(ids []string, texts []string) {
	// Keep ids for lookup; display texts in the combo (ids and texts are parallel).
	d.ids = append(d.ids[:0], ids...)
	d.texts = append(d.texts[:0], texts...)
	if d.contestCombo == nil {
		return
	}
	d.ignoreChangedEvent = true
	defer func() { d.ignoreChangedEvent = false }()
	d.contestCombo.Clear()
	for _, t := range texts {
		d.contestCombo.AddItem(t)
	}
}

func (d *newContestDialog) SelectContestIdentifier(value string) {
	if d.contestCombo == nil {
		return
	}
	for i, id := range d.ids {
		if id == value {
			d.ignoreChangedEvent = true
			d.contestCombo.SetCurrentIndex(i)
			d.ignoreChangedEvent = false
			return
		}
	}
}

func (d *newContestDialog) SetContestName(value string) {
	if d.nameEdit == nil {
		return
	}
	d.ignoreChangedEvent = true
	defer func() { d.ignoreChangedEvent = false }()
	d.nameEdit.SetText(value)
}

func (d *newContestDialog) SetDataComplete(complete bool) {
	if d.okButton == nil {
		return
	}
	d.okButton.SetEnabled(complete)
}
