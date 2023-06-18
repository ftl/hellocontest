package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type NewContestController interface {
	SelectContestIdentifier(string)
	EnterContestName(string)
	ChooseContestFilename()
}

type newContestView struct {
	parent     *gtk.Dialog
	controller NewContestController

	ignoreChangedEvent bool

	identifierCombo *gtk.ComboBoxText
	nameEntry       *gtk.Entry
	filenameLabel   *gtk.Label
	filenameChooser *gtk.Button
	cancelButton    *gtk.Button
	okButton        *gtk.Button
}

func setupNewContestView(builder *gtk.Builder, parent *gtk.Dialog, controller NewContestController, ids []string, texts []string) *newContestView {
	result := new(newContestView)
	result.parent = parent
	result.controller = controller

	result.identifierCombo = getUI(builder, "newContestIdentifierCombo").(*gtk.ComboBoxText)
	result.nameEntry = getUI(builder, "newContestNameEntry").(*gtk.Entry)
	result.filenameLabel = getUI(builder, "newContestFilenameLabel").(*gtk.Label)
	result.filenameChooser = getUI(builder, "newContestChooseFileButton").(*gtk.Button)
	result.cancelButton = getUI(builder, "cancelNewContestButton").(*gtk.Button)
	result.okButton = getUI(builder, "createNewContestButton").(*gtk.Button)

	result.cancelButton.Connect("clicked", result.onCancelClicked)
	result.okButton.Connect("clicked", result.onOKClicked)
	result.okButton.SetSensitive(false)
	result.okButton.SetCanDefault(true)
	result.parent.SetDefault(&result.okButton.Widget)

	for i, value := range ids {
		result.identifierCombo.Append(value, texts[i])
	}
	result.identifierCombo.SetActive(0)

	result.identifierCombo.Connect("changed", result.onIdentifierChanged)
	result.nameEntry.Connect("changed", result.onNameChanged)
	result.filenameChooser.Connect("clicked", result.onChooseFilenameClicked)

	return result
}

func (v *newContestView) onCancelClicked() {
	v.parent.Response(gtk.RESPONSE_CANCEL)
}

func (v *newContestView) onOKClicked() {
	v.parent.Response(gtk.RESPONSE_OK)
}

func (v *newContestView) onIdentifierChanged() {
	v.controller.SelectContestIdentifier(v.identifierCombo.GetActiveID())
}

func (v *newContestView) onNameChanged() {
	text, _ := v.nameEntry.GetText()
	v.controller.EnterContestName(text)
}

func (v *newContestView) onChooseFilenameClicked() {
	v.controller.ChooseContestFilename()
}

func (v *newContestView) doIgnoreChanges(f func()) {
	if v == nil {
		return
	}

	v.ignoreChangedEvent = true
	defer func() {
		v.ignoreChangedEvent = false
	}()
	f()
}

func (v *newContestView) SelectContestIdentifier(value string) {
	v.doIgnoreChanges(func() {
		v.identifierCombo.SetActiveID(value)
	})
}

func (v *newContestView) SetContestName(value string) {
	v.doIgnoreChanges(func() {
		v.nameEntry.SetText(value)
	})
}

func (v *newContestView) SetContestFilename(value string) {
	v.doIgnoreChanges(func() {
		v.filenameLabel.SetText(value)
	})
}

func (v *newContestView) SetDataComplete(value bool) {
	v.doIgnoreChanges(func() {
		v.okButton.SetSensitive(value)
	})
}
