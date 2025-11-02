package ui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type qtcDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller QTCController
	view       *qtcView
	logButton  *gtk.Button

	// data fields
	activePhase core.QTCWorkflowPhase
	activeQTC   int
}

func setupQTCDialog(parent gtk.IWidget, controller QTCController) *qtcDialog {
	result := &qtcDialog{
		parent:     parent,
		controller: controller,
	}
	return result
}

func (d *qtcDialog) onDestroy() {
	d.dialog = nil
	d.view = nil
}

func (d *qtcDialog) QuestionQTCCount(max int) (int, bool) {
	// TODO: implement modal dialog
	return 10, true
}

func (d *qtcDialog) Show(qtcMode core.QTCMode, qtcSeries core.QTCSeries) {
	d.view = newQTCView(d.controller, qtcMode)

	// setup the dialog
	dialog, _ := gtk.DialogNew()
	d.dialog = dialog
	d.dialog.SetDefaultSize(400, 400)
	d.dialog.SetTransientFor(nil)
	d.dialog.SetPosition(gtk.WIN_POS_CENTER)
	d.dialog.Connect("destroy", d.onDestroy)
	d.dialog.SetTitle("QTC")
	d.dialog.SetDefaultResponse(gtk.RESPONSE_OK)
	d.dialog.SetModal(true)
	contentArea, _ := d.dialog.GetContentArea()
	contentArea.Add(d.view.root)
	d.logButton, _ = d.dialog.AddButton("Log", gtk.RESPONSE_OK)
	// TODO: add a check before closing the dialog
	d.dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)
	d.dialog.ShowAll()

	// put the QTC series data into the view's widgets
	d.view.setHeader(qtcSeries.TheirCallsign(), qtcSeries.Header)
	d.view.setQTCs(qtcSeries.QTCs)
	d.focusActivePhase()

	defer log.Println("QTC Dialog closed")

	// run the dialog until it is closed by the controller (d.dialog == nil)
	for {
		response := d.dialog.Run()
		switch response {
		case gtk.RESPONSE_OK:
			d.controller.CompleteQTCSeries()
		default:
			d.controller.AbortQTCSeries()
		}
		if d.dialog == nil {
			return
		}
	}
}

func (d *qtcDialog) UpdateQTC(index int, qtc core.QTC) {
	if d.view == nil {
		return
	}
	d.view.setQTC(index, qtc)
}

func (d *qtcDialog) Close() {
	if d.dialog == nil {
		return
	}

	d.dialog.Close()
	d.dialog.Destroy()
	d.dialog = nil
	d.view = nil
}

func (d *qtcDialog) SetActivePhase(phase core.QTCWorkflowPhase) {
	d.activePhase = phase
	d.focusActivePhase()
}

func (d *qtcDialog) focusActivePhase() {
	if d.view == nil {
		return
	}
	switch d.activePhase {
	case core.QTCStart:
		d.view.focusStart()
	case core.QTCExchangeHeader:
		d.view.focusHeader()
	case core.QTCExchangeData:
		d.view.focusData()
	case core.QTCFinish:
		d.view.focusNone()
		d.logButton.GrabFocus()
	}
}

func (d *qtcDialog) SetActiveQTC(index int) {
	d.activeQTC = index
	d.focusActiveQTC()
}

func (d *qtcDialog) focusActiveQTC() {
	if d.view == nil {
		return
	}
	d.view.focusQTC(d.activeQTC)
}
