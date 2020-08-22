package ui

import (
	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gtk"
)

type callinfoView struct {
	controller core.CallinfoController
}

func setupCallinfoView(builder *gtk.Builder) *callinfoView {
	result := new(callinfoView)

	return result
}

func (v *callinfoView) SetCallinfoController(controller core.CallinfoController) {
	v.controller = controller
}
