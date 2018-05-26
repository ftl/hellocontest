package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type mainWindow struct {
	window      *gtk.ApplicationWindow
	callsign    *gtk.Entry
	theirReport *gtk.Entry
	theirNumber *gtk.Entry
	myReport    *gtk.Entry
	myNumber    *gtk.Entry
	logButton   *gtk.Button
	resetButton *gtk.Button
	errorLabel  *gtk.Label

	qsoList *gtk.ListStore

	entry core.EntryController
	log   core.Log
}

func setupMainWindow(builder *gtk.Builder, application *gtk.Application) *mainWindow {
	result := new(mainWindow)

	result.window = getUI(builder, "mainWindow").(*gtk.ApplicationWindow)
	result.window.SetApplication(application)
	result.window.SetDefaultSize(500, 500)

	result.callsign = getUI(builder, "callsignEntry").(*gtk.Entry)
	result.theirReport = getUI(builder, "theirReportEntry").(*gtk.Entry)
	result.theirNumber = getUI(builder, "theirNumberEntry").(*gtk.Entry)
	result.myReport = getUI(builder, "myReportEntry").(*gtk.Entry)
	result.myNumber = getUI(builder, "myNumberEntry").(*gtk.Entry)
	result.logButton = getUI(builder, "logButton").(*gtk.Button)
	result.resetButton = getUI(builder, "resetButton").(*gtk.Button)
	result.errorLabel = getUI(builder, "errorLabel").(*gtk.Label)

	result.addEntryTraversal(result.callsign)
	result.addEntryTraversal(result.theirReport)
	result.addEntryTraversal(result.theirNumber)
	result.addEntryTraversal(result.myReport)
	result.addEntryTraversal(result.myNumber)

	result.qsoList = setupQsoView(getUI(builder, "qsoView").(*gtk.TreeView))

	return result
}

const (
	qsoColumnUTC int = iota
	qsoColumnCallsign
	qsoColumnBand
	qsoColumnMyReport
	qsoColumnMyNumber
	qsoColumnTheirReport
	qsoColumnTheirNumber
)

func setupQsoView(qsoView *gtk.TreeView) *gtk.ListStore {
	qsoView.AppendColumn(createColumn("UTC", qsoColumnUTC))
	qsoView.AppendColumn(createColumn("Callsign", qsoColumnCallsign))
	qsoView.AppendColumn(createColumn("Band", qsoColumnBand))
	qsoView.AppendColumn(createColumn("My RST", qsoColumnMyReport))
	qsoView.AppendColumn(createColumn("My #", qsoColumnMyNumber))
	qsoView.AppendColumn(createColumn("Th RST", qsoColumnTheirReport))
	qsoView.AppendColumn(createColumn("Th #", qsoColumnTheirNumber))

	qsoList, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		log.Fatalf("Cannot create QSO list store: %v", err)
	}
	qsoView.SetModel(qsoList)
	return qsoList
}

func createColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatalf("Cannot create text cell renderer for column %s: %v", title, err)
	}

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)
	if err != nil {
		log.Fatalf("Cannot create column %s: %v", title, err)
	}
	return column
}

func (w *mainWindow) addEntryTraversal(widget *gtk.Entry) {
	widget.Connect("key_press_event", w.onEntryKeyPress)
	widget.Connect("focus_in_event", w.onEntryFocusIn)
	widget.Connect("focus_out_event", w.onEntryFocusOut)
}

func (w *mainWindow) onEntryKeyPress(widget interface{}, event *gdk.Event) bool {
	keyEvent := gdk.EventKeyNewFromEvent(event)
	log.Printf("key pressed in: %[1]T %[1]v", widget)
	switch keyEvent.KeyVal() {
	case gdk.KEY_Tab:
		log.Print("TAB pressed")
		w.entry.GotoNextField()
		return true
	case gdk.KEY_space:
		log.Print("Space pressed")
		w.entry.GotoNextField()
		return true
	case gdk.KEY_Return:
		log.Print("Return pressed")
		w.entry.Log()
		return true
	default:
		log.Printf("Key pressed: %q", keyEvent.KeyVal())
		return false
	}
}

func (w *mainWindow) onEntryFocusIn(entry *gtk.Entry, event *gdk.Event) bool {
	entryField := w.entryToField(entry)
	log.Printf("found active field %d", entryField)
	w.entry.SetActiveField(entryField)
	return false
}

func (w *mainWindow) onEntryFocusOut(entry *gtk.Entry, event *gdk.Event) bool {
	entry.SelectRegion(0, 0)
	return false
}

func (w *mainWindow) Show() {
	w.window.ShowAll()
}

func (w *mainWindow) SetEntryController(entry core.EntryController) {
	w.entry = entry
}

func (w *mainWindow) GetCallsign() string {
	text, err := w.callsign.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (w *mainWindow) SetCallsign(text string) {
	w.callsign.SetText(text)
}

func (w *mainWindow) GetTheirReport() string {
	text, err := w.theirReport.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (w *mainWindow) SetTheirReport(text string) {
	w.theirReport.SetText(text)
}

func (w *mainWindow) GetTheirNumber() string {
	text, err := w.theirNumber.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (w *mainWindow) SetTheirNumber(text string) {
	w.theirNumber.SetText(text)
}

func (w *mainWindow) GetMyReport() string {
	text, err := w.myReport.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (w *mainWindow) SetMyReport(text string) {
	w.myReport.SetText(text)
}

func (w *mainWindow) GetMyNumber() string {
	text, err := w.myNumber.GetText()
	if err != nil {
		log.Printf("Error getting text: %v", err)
	}
	return text
}

func (w *mainWindow) SetMyNumber(text string) {
	w.myNumber.SetText(text)
}

func (w *mainWindow) SetActiveField(field core.EntryField) {
	log.Printf("window.SetActiveField: %d", field)
	entry := w.fieldToEntry(field)
	log.Printf("found entry %v", entry)
	entry.GrabFocus()
}

func (w *mainWindow) fieldToEntry(field core.EntryField) *gtk.Entry {
	switch field {
	case core.CallsignField:
		return w.callsign
	case core.TheirReportField:
		return w.theirReport
	case core.TheirNumberField:
		return w.theirNumber
	case core.MyReportField:
		return w.myReport
	case core.MyNumberField:
		return w.myNumber
	case core.OtherField:
		return w.callsign
	default:
		log.Fatalf("Unknown entry field %d", field)
	}
	panic("this is never reached")
}

func (w *mainWindow) entryToField(entry *gtk.Entry) core.EntryField {
	name, _ := entry.GetName()
	switch name {
	case "callsignEntry":
		return core.CallsignField
	case "theirReportEntry":
		return core.TheirReportField
	case "theirNumberEntry":
		return core.TheirNumberField
	case "myReportEntry":
		return core.MyReportField
	case "myNumberEntry":
		return core.MyNumberField
	default:
		return core.OtherField
	}
}

func (w *mainWindow) SetDuplicateMarker(bool) {}
func (w *mainWindow) ShowError(err error) {
	w.errorLabel.SetText(err.Error())
}
func (w *mainWindow) ClearError() {
	w.errorLabel.SetText("")
}

func (w *mainWindow) SetLog(log core.Log) {
	w.log = log
}

func (w *mainWindow) UpdateAllRows(qsos []core.QSO) {

}

func (w *mainWindow) RowAdded(qso core.QSO) {
	newRow := w.qsoList.Append()
	err := w.qsoList.Set(newRow,
		[]int{
			qsoColumnUTC,
			qsoColumnCallsign,
			qsoColumnBand,
			qsoColumnMyReport,
			qsoColumnMyNumber,
			qsoColumnTheirReport,
			qsoColumnTheirNumber,
		},
		[]interface{}{
			qso.Time.In(time.UTC).Format("15:04"),
			qso.Callsign.String(),
			qso.Band.String(),
			qso.MyReport.String(),
			fmt.Sprintf("%03d", qso.MyNumber),
			qso.TheirReport.String(),
			fmt.Sprintf("%03d", qso.TheirNumber),
		})
	if err != nil {
		log.Printf("Cannot add QSO row %s: %v", qso.String(), err)
	}
}
