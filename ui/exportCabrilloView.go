package ui

import "github.com/gotk3/gotk3/gtk"

type ExportCabrilloController interface {
	SetOpenAfterExport(bool)
}

type exportCabrilloView struct {
	controller ExportCabrilloController

	openAfterExportCheckButton *gtk.CheckButton
}

func (v *exportCabrilloView) setup(controller ExportCabrilloController) {
	v.controller = controller
	v.openAfterExportCheckButton.Connect("toggled", v.onOpenAfterExportToggled)
}

func (v *exportCabrilloView) onOpenAfterExportToggled() {
	v.controller.SetOpenAfterExport(v.openAfterExportCheckButton.GetActive())
}
