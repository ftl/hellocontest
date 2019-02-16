package ui

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/ftl/hellocontest/core"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/pkg/errors"
)

type mainWindow struct {
	window *gtk.ApplicationWindow

	*mainMenu
	*entryView
	*keyer

	qsoView *gtk.TreeView
	qsoList *gtk.ListStore

	log   core.Log
	entry core.EntryController
}

func setupMainWindow(builder *gtk.Builder, application *gtk.Application) *mainWindow {
	result := new(mainWindow)

	result.window = getUI(builder, "mainWindow").(*gtk.ApplicationWindow)
	result.window.SetApplication(application)
	result.window.SetDefaultSize(500, 500)

	result.mainMenu = setupMainMenu(builder)
	result.entryView = setupEntryView(builder)
	result.keyer = setupKeyer(builder)

	result.qsoView = getUI(builder, "qsoView").(*gtk.TreeView)
	result.qsoList = setupQsoView(result.qsoView)

	return result
}

const (
	qsoColumnUTC int = iota
	qsoColumnCallsign
	qsoColumnBand
	qsoColumnMode
	qsoColumnMyReport
	qsoColumnMyNumber
	qsoColumnMyXchange
	qsoColumnTheirReport
	qsoColumnTheirNumber
	qsoColumnTheirXchange
)

func setupQsoView(qsoView *gtk.TreeView) *gtk.ListStore {
	qsoView.AppendColumn(createColumn("UTC", qsoColumnUTC))
	qsoView.AppendColumn(createColumn("Callsign", qsoColumnCallsign))
	qsoView.AppendColumn(createColumn("Band", qsoColumnBand))
	qsoView.AppendColumn(createColumn("Mode", qsoColumnMode))
	qsoView.AppendColumn(createColumn("My RST", qsoColumnMyReport))
	qsoView.AppendColumn(createColumn("My #", qsoColumnMyNumber))
	qsoView.AppendColumn(createColumn("My XChg", qsoColumnMyXchange))
	qsoView.AppendColumn(createColumn("Th RST", qsoColumnTheirReport))
	qsoView.AppendColumn(createColumn("Th #", qsoColumnTheirNumber))
	qsoView.AppendColumn(createColumn("Th XChg", qsoColumnTheirXchange))

	qsoList, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
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

func (w *mainWindow) SetLog(log core.Log) {
	w.log = log
}

func (w *mainWindow) UpdateAllRows(qsos []core.QSO) {
	w.qsoList.Clear()
	for _, qso := range qsos {
		w.RowAdded(qso)
	}
}

func (w *mainWindow) RowAdded(qso core.QSO) {
	newRow := w.qsoList.Append()
	err := w.qsoList.Set(newRow,
		[]int{
			qsoColumnUTC,
			qsoColumnCallsign,
			qsoColumnBand,
			qsoColumnMode,
			qsoColumnMyReport,
			qsoColumnMyNumber,
			qsoColumnMyXchange,
			qsoColumnTheirReport,
			qsoColumnTheirNumber,
			qsoColumnTheirXchange,
		},
		[]interface{}{
			qso.Time.In(time.UTC).Format("15:04"),
			qso.Callsign.String(),
			qso.Band.String(),
			qso.Mode.String(),
			qso.MyReport.String(),
			qso.MyNumber.String(),
			qso.MyXchange,
			qso.TheirReport.String(),
			qso.TheirNumber.String(),
			qso.TheirXchange,
		})
	if err != nil {
		log.Printf("Cannot add QSO row %s: %v", qso.String(), err)
	}
	path, err := w.qsoList.GetPath(newRow)
	if err != nil {
		log.Printf("Cannot get path for list item: %s", err)
	}
	w.qsoView.SetCursorOnCell(path, w.qsoView.GetColumn(1), nil, false)
}

func (w *mainWindow) Show() {
	w.window.ShowAll()
}

func (w *mainWindow) ShowFilename(filename string) {
	w.window.SetTitle(fmt.Sprintf("Hello Contest %s", filepath.Base(filename)))
}

func (w *mainWindow) SelectOpenFile(title string, patterns ...string) (string, bool, error) {
	dlg, err := gtk.FileChooserDialogNewWith1Button(title, &w.window.Window, gtk.FILE_CHOOSER_ACTION_OPEN, "Open", gtk.RESPONSE_ACCEPT)
	if err != nil {
		errors.Wrap(err, "cannot create a file selection dialog to open a file")
	}
	defer dlg.Destroy()

	if len(patterns) > 0 {
		filter, err := gtk.FileFilterNew()
		if err != nil {
			return "", false, errors.Wrap(err, "cannot create a file selection dialog to open a file")
		}
		for _, pattern := range patterns {
			filter.AddPattern(pattern)
		}
		dlg.SetFilter(filter)
	}

	result := dlg.Run()
	if result != int(gtk.RESPONSE_ACCEPT) {
		return "", false, nil
	}

	return dlg.GetFilename(), true, nil
}

func (w *mainWindow) SelectSaveFile(title string, patterns ...string) (string, bool, error) {
	dlg, err := gtk.FileChooserDialogNewWith1Button(title, &w.window.Window, gtk.FILE_CHOOSER_ACTION_SAVE, "Save", gtk.RESPONSE_ACCEPT)
	if err != nil {
		return "", false, errors.Wrap(err, "cannot create a file selection dialog to save a file")
	}
	defer dlg.Destroy()

	dlg.SetDoOverwriteConfirmation(true)

	if len(patterns) > 0 {
		filter, err := gtk.FileFilterNew()
		if err != nil {
			return "", false, errors.Wrap(err, "cannot create a file selection dialog to save a file")
		}
		for _, pattern := range patterns {
			filter.AddPattern(pattern)
		}
		dlg.SetFilter(filter)
	}

	result := dlg.Run()
	if result != int(gtk.RESPONSE_ACCEPT) {
		return "", false, nil
	}

	return dlg.GetFilename(), true, nil
}

func (w *mainWindow) ShowInfoDialog(format string, a ...interface{}) {
	dlg := gtk.MessageDialogNew(w.window, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, format, a...)
	defer dlg.Destroy()
	dlg.Run()
}

func (w *mainWindow) ShowErrorDialog(format string, a ...interface{}) {
	dlg := gtk.MessageDialogNew(w.window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, format, a...)
	defer dlg.Destroy()
	dlg.Run()
}
