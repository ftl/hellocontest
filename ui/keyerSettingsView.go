package ui

import (
	"fmt"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

const keyerMacroCount = 4

// KeyerSettingsController is the callback surface the keyer settings dialog uses.
type KeyerSettingsController interface {
	EnterLabel(core.Workmode, int, string)
	EnterMacro(core.Workmode, int, string)
	SelectPreset(string)
	EnterParrotIntervalSeconds(int)
	Save()
}

type keyerSettingsView struct {
	dialog *qtlib.QDialog

	root *qtlib.QWidget

	// labels[workmode-1][index], macros[workmode-1][index]
	labels [2][keyerMacroCount]*qtlib.QLineEdit
	macros [2][keyerMacroCount]*qtlib.QLineEdit

	presetCombo *qtlib.QComboBox
	parrotSpin  *qtlib.QSpinBox

	messageLabel *qtlib.QLabel
	closeBtn     *qtlib.QPushButton

	controller         KeyerSettingsController
	ignoreChangedEvent bool
}

func (v *keyerSettingsView) doIgnore(f func()) {
	v.ignoreChangedEvent = true
	defer func() { v.ignoreChangedEvent = false }()
	f()
}

func newKeyerSettingsView(dialog *qtlib.QDialog, controller KeyerSettingsController) *keyerSettingsView {
	v := &keyerSettingsView{dialog: dialog, controller: controller}

	v.root = qtlib.NewQWidget2()
	root := qtlib.NewQVBoxLayout(v.root)

	templatesLabel := qtlib.NewQLabel2()
	templatesLabel.SetTextFormat(qtlib.RichText)
	templatesLabel.SetOpenExternalLinks(true)
	templatesLabel.SetText(`<b>Templates:</b>
<table>
<tr><td>{{.MyCall}}</td><td>the station callsign</td></tr>
<tr><td>{{.MyReport}}</td><td>my report</td></tr>
<tr><td>{{.MyNumber}}</td><td>my QSO number (t=0, n=9)</td></tr>
<tr><td>{{.MyXchange}}</td><td>all exchange fields concatenated, except report and number</td></tr>
<tr><td>{{.MyExchange}}</td><td>all exchange fields concatenated, including report and number</td></tr>
<tr><td>{{.LastNumber}}</td><td>the QSO number of the last logged QSO (t=0, n=9)</td></tr>
<tr><td>{{.TheirCall}}</td><td>their callsign</td></tr>
</table>
For more details see <a href="https://github.com/ftl/hellocontest/wiki/Main-Window#cw-macros">the wiki</a>.
`)
	root.AddWidget(templatesLabel.QWidget)

	tabs := qtlib.NewQTabWidget2()
	tabs.AddTab(v.buildMacroTab(core.SearchPounce), "Search && Pounce")
	tabs.AddTab(v.buildMacroTab(core.Run), "Run")
	root.AddWidget(tabs.QWidget)

	root.AddWidget(v.buildPresetGroup().QWidget)
	root.AddWidget(v.buildParrotGroup().QWidget)

	v.messageLabel = qtlib.NewQLabel3("")
	root.AddWidget(v.messageLabel.QWidget)

	root.AddWidget(v.buildButtonRow())

	return v
}

func (v *keyerSettingsView) buildMacroTab(workmode core.Workmode) *qtlib.QWidget {
	page := qtlib.NewQWidget2()
	grid := qtlib.NewQGridLayout(page)

	// Header row
	grid.AddWidget2(qtlib.NewQLabel3("Key").QWidget, 0, 0)
	grid.AddWidget2(qtlib.NewQLabel3("Label").QWidget, 0, 1)
	grid.AddWidget2(qtlib.NewQLabel3("Macro").QWidget, 0, 2)

	idx := int(workmode) - 1
	for i := 0; i < keyerMacroCount; i++ {
		row := i + 1
		slot := i

		keyLabel := qtlib.NewQLabel3(fmt.Sprintf("F%d:", i+1))
		grid.AddWidget2(keyLabel.QWidget, row, 0)

		labelEdit := qtlib.NewQLineEdit2()
		labelEdit.OnTextChanged(func(text string) {
			if v.ignoreChangedEvent {
				return
			}
			v.controller.EnterLabel(workmode, slot, text)
		})
		labelEdit.OnFocusOutEvent(func(super func(ev *qtlib.QFocusEvent), ev *qtlib.QFocusEvent) {
			super(ev)
			v.controller.Save()
		})
		grid.AddWidget2(labelEdit.QWidget, row, 1)
		v.labels[idx][slot] = labelEdit

		macroEdit := qtlib.NewQLineEdit2()
		macroEdit.OnTextChanged(func(text string) {
			if v.ignoreChangedEvent {
				return
			}
			v.controller.EnterMacro(workmode, slot, text)
		})
		macroEdit.OnFocusOutEvent(func(super func(ev *qtlib.QFocusEvent), ev *qtlib.QFocusEvent) {
			super(ev)
			v.controller.Save()
		})
		grid.AddWidget2(macroEdit.QWidget, row, 2)
		v.macros[idx][slot] = macroEdit
	}

	return page
}

func (v *keyerSettingsView) buildPresetGroup() *qtlib.QGroupBox {
	box := qtlib.NewQGroupBox3("Preset")
	form := qtlib.NewQFormLayout(box.QWidget)

	v.presetCombo = qtlib.NewQComboBox2()
	v.presetCombo.OnCurrentTextChanged(func(text string) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.SelectPreset(text)
	})
	form.AddRow3("Load preset:", v.presetCombo.QWidget)

	return box
}

func (v *keyerSettingsView) buildParrotGroup() *qtlib.QGroupBox {
	box := qtlib.NewQGroupBox3("CQ Parrot")
	form := qtlib.NewQFormLayout(box.QWidget)

	v.parrotSpin = qtlib.NewQSpinBox2()
	v.parrotSpin.SetMinimum(1)
	v.parrotSpin.SetMaximum(600)
	v.parrotSpin.SetSuffix(" s")
	v.parrotSpin.OnValueChanged(func(val int) {
		if v.ignoreChangedEvent {
			return
		}
		v.controller.EnterParrotIntervalSeconds(val)
	})
	v.parrotSpin.OnFocusOutEvent(func(super func(ev *qtlib.QFocusEvent), ev *qtlib.QFocusEvent) {
		super(ev)
		v.controller.Save()
	})
	form.AddRow3("Repeat interval:", v.parrotSpin.QWidget)

	return box
}

func (v *keyerSettingsView) buildButtonRow() *qtlib.QWidget {
	line := qtlib.NewQWidget2()
	hbox := qtlib.NewQHBoxLayout(line)
	hbox.SetContentsMargins(0, 0, 0, 0)
	hbox.AddStretch()

	v.closeBtn = qtlib.NewQPushButton3("Close")
	v.closeBtn.OnClicked(func() {
		v.controller.Save()
		v.dialog.Reject()
	})
	hbox.AddWidget(v.closeBtn.QWidget)

	return line
}

// keyer.SettingsView implementation

func (v *keyerSettingsView) ShowMessage(args ...any) {
	v.messageLabel.SetText(fmt.Sprint(args...))
}

func (v *keyerSettingsView) ClearMessage() {
	v.messageLabel.SetText("")
}

func (v *keyerSettingsView) SetLabel(workmode core.Workmode, index int, text string) {
	idx := int(workmode) - 1
	if idx < 0 || idx >= len(v.labels) || index < 0 || index >= keyerMacroCount {
		return
	}
	v.doIgnore(func() { v.labels[idx][index].SetText(text) })
}

func (v *keyerSettingsView) SetMacro(workmode core.Workmode, index int, text string) {
	idx := int(workmode) - 1
	if idx < 0 || idx >= len(v.macros) || index < 0 || index >= keyerMacroCount {
		return
	}
	v.doIgnore(func() { v.macros[idx][index].SetText(text) })
}

func (v *keyerSettingsView) SetPresetNames(names []string) {
	v.doIgnore(func() {
		v.presetCombo.Clear()
		v.presetCombo.AddItem("")
		for _, n := range names {
			v.presetCombo.AddItem(n)
		}
	})
}

func (v *keyerSettingsView) SetPreset(name string) {
	v.doIgnore(func() {
		idx := v.presetCombo.FindText(name)
		if idx < 0 {
			idx = 0
		}
		v.presetCombo.SetCurrentIndex(idx)
	})
}

func (v *keyerSettingsView) SetParrotIntervalSeconds(interval int) {
	if interval <= 0 {
		interval = 1
	}
	v.doIgnore(func() { v.parrotSpin.SetValue(interval) })
}
