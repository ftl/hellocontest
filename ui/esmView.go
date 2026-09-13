package ui

import (
	qtlib "github.com/mappu/miqt/qt6"
)

type ESMController interface {
	SetESMEnabled(enabled bool)
}

type esmView struct {
	widget      *qtlib.QWidget
	checkbox    *qtlib.QCheckBox
	msgLabel    *qtlib.QLabel
	ignoreInput bool
	controller  ESMController
}

func newESMView() *esmView {
	v := &esmView{}
	v.widget = qtlib.NewQWidget2()
	v.widget.SetObjectName(*qtlib.NewQAnyStringView3("esm"))
	layout := qtlib.NewQHBoxLayout(v.widget)

	v.checkbox = qtlib.NewQCheckBox3("ESM")
	v.checkbox.SetFocusPolicy(qtlib.NoFocus)
	layout.AddWidget(v.checkbox.QWidget)

	v.msgLabel = qtlib.NewQLabel2()
	layout.AddWidget(v.msgLabel.QWidget)

	v.checkbox.OnStateChanged(func(state int) {
		if v.ignoreInput || v.controller == nil {
			return
		}
		v.controller.SetESMEnabled(state == int(qtlib.Checked))
	})

	return v
}

func (v *esmView) SetESMController(c ESMController) { v.controller = c }

func (v *esmView) SetESMEnabled(enabled bool) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	if enabled {
		v.checkbox.SetCheckState(qtlib.Checked)
	} else {
		v.checkbox.SetCheckState(qtlib.Unchecked)
	}
}

func (v *esmView) SetMessage(message string) {
	v.msgLabel.SetText(message)
}
