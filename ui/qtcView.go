package ui

import "github.com/gotk3/gotk3/gtk"

type QTCController interface {
	// TBD

}

type qtcView struct {
	controller QTCController

	// widgets
	root *gtk.Grid
}

func newQTCView(controller QTCController) *qtcView {
	result := &qtcView{
		controller: controller,
	}

	return result
}
