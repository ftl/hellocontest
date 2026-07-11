package ui

import (
	"fmt"

	qtlib "github.com/mappu/miqt/qt6"
)

type KeyerController interface {
	SendMacro(int)
	Stop()
	EnterSpeed(int)
	Save()
	OpenKeyerSettings()
}

type keyerView struct {
	widget     *qtlib.QWidget
	buttons    [4]*qtlib.QPushButton
	stopButton *qtlib.QPushButton
	macrosBtn  *qtlib.QPushButton
	speedSpin  *qtlib.QSpinBox

	cqLabelText  string
	parrotActive bool
	ignoreInput  bool
	controller   KeyerController
}

func newKeyerView() *keyerView {
	v := &keyerView{}
	v.widget = qtlib.NewQWidget2()
	v.widget.SetObjectName(*qtlib.NewQAnyStringView3("keyerButtons"))
	layout := qtlib.NewQGridLayout(v.widget)
	layout.SetContentsMargins(4, 5, 4, 5)
	layout.SetSpacing(5)

	v.stopButton = qtlib.NewQPushButton3("ESC: Stop")
	layout.AddWidget2(v.stopButton.QWidget, 0, 0)
	v.stopButton.OnClicked(func() {
		if v.controller != nil {
			v.controller.Stop()
		}
	})

	for i := range 4 {
		idx := i
		btn := qtlib.NewQPushButton3(fmt.Sprintf("F%d", i+1))
		v.buttons[i] = btn
		layout.AddWidget2(btn.QWidget, 0, i+1)
		btn.OnClicked(func() {
			if v.controller == nil {
				return
			}
			v.controller.SendMacro(idx)
		})
	}

	v.macrosBtn = qtlib.NewQPushButton3("Macros...")
	v.macrosBtn.SetObjectName(*qtlib.NewQAnyStringView3("keyerSettingsButton"))
	layout.AddWidget2(v.macrosBtn.QWidget, 1, 0)
	v.macrosBtn.OnClicked(func() {
		if v.controller != nil {
			v.controller.OpenKeyerSettings()
		}
	})

	speedLabel := qtlib.NewQLabel3("Speed:")
	speedLabel.SetAlignment(qtlib.AlignTrailing | qtlib.AlignVCenter)
	layout.AddWidget2(speedLabel.QWidget, 1, 3)

	v.speedSpin = qtlib.NewQSpinBox2()
	v.speedSpin.SetObjectName(*qtlib.NewQAnyStringView3("keyerSpeed"))
	v.speedSpin.SetMinimum(5)
	v.speedSpin.SetMaximum(60)
	v.speedSpin.SetSuffix(" WPM")
	layout.AddWidget2(v.speedSpin.QWidget, 1, 4)
	v.speedSpin.OnValueChanged(func(val int) {
		if v.ignoreInput || v.controller == nil {
			return
		}
		v.controller.EnterSpeed(val)
	})
	v.speedSpin.OnFocusOutEvent(func(super func(event *qtlib.QFocusEvent), event *qtlib.QFocusEvent) {
		super(event)
		if v.controller != nil {
			v.controller.Save()
		}
	})

	return v
}

func (v *keyerView) SetKeyerController(c KeyerController) { v.controller = c }

func (v *keyerView) SetLabel(index int, text string) {
	if index < 0 || index >= 4 {
		return
	}
	if index == 0 {
		v.cqLabelText = text
	}
	v.buttons[index].SetText(v.buildLabel(index, text))
}

func (v *keyerView) buildLabel(index int, text string) string {
	decoration := fmt.Sprintf("F%d", index+1)
	if index == 0 && v.parrotActive {
		decoration = "🦜"
	}
	if text == "" {
		return decoration
	}
	return fmt.Sprintf("%s: %s", decoration, text)
}

func (v *keyerView) SetPattern(index int, text string) {
	if index < 0 || index >= 4 {
		return
	}
	v.buttons[index].SetToolTip(fmt.Sprintf("F%d: %s", index+1, text))
}

func (v *keyerView) SetSpeed(wpm int) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	v.speedSpin.SetValue(wpm)
}

func (v *keyerView) SetParrotActive(active bool) {
	v.parrotActive = active
	v.buttons[0].SetText(v.buildLabel(0, v.cqLabelText))
}
