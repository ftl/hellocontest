package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
)

type WorkmodeController interface {
	SetWorkmode(core.Workmode)
}

type workmodeView struct {
	widget                 *qtlib.QWidget
	searchPounceRadio      *qtlib.QRadioButton
	runRadio               *qtlib.QRadioButton
	operationModeHintLabel *qtlib.QLabel
	controller             WorkmodeController
	ignoreInput            bool
}

func newWorkmodeView() *workmodeView {
	v := &workmodeView{}
	v.widget = qtlib.NewQWidget2()
	v.widget.SetObjectName(*qtlib.NewQAnyStringView3("workmode"))
	layout := qtlib.NewQHBoxLayout(v.widget)

	workmodeLabel := qtlib.NewQLabel3("Workmode:")
	v.searchPounceRadio = qtlib.NewQRadioButton3("S&&P")
	v.searchPounceRadio.SetFocusPolicy(qtlib.NoFocus)
	v.runRadio = qtlib.NewQRadioButton3("Run")
	v.runRadio.SetFocusPolicy(qtlib.NoFocus)
	v.operationModeHintLabel = qtlib.NewQLabel2()

	layout.AddWidget(workmodeLabel.QWidget)
	layout.AddWidget(v.searchPounceRadio.QWidget)
	layout.AddWidget(v.runRadio.QWidget)
	layout.AddWidget(v.operationModeHintLabel.QWidget)

	v.searchPounceRadio.OnToggled(func(checked bool) {
		if v.ignoreInput || !checked || v.controller == nil {
			return
		}
		v.controller.SetWorkmode(core.SearchPounce)
	})
	v.runRadio.OnToggled(func(checked bool) {
		if v.ignoreInput || !checked || v.controller == nil {
			return
		}
		v.controller.SetWorkmode(core.Run)
	})

	return v
}

func (v *workmodeView) SetWorkmodeController(c WorkmodeController) { v.controller = c }

func (v *workmodeView) SetWorkmode(wm core.Workmode) {
	v.ignoreInput = true
	defer func() { v.ignoreInput = false }()
	switch wm {
	case core.SearchPounce:
		v.searchPounceRadio.SetChecked(true)
	case core.Run:
		v.runRadio.SetChecked(true)
	}
}

func (v *workmodeView) SetOperationModeHint(hint string) {
	v.operationModeHintLabel.SetText(hint)
}
