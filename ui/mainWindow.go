//go:build !fyne

package ui

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gotk3/gotk3/gtk"
	"github.com/pkg/errors"

	"github.com/ftl/gmtry"
	"github.com/ftl/hellocontest/ui/style"
)

type mainWindow struct {
	window *gtk.ApplicationWindow

	*mainMenu
	*radioMenu
	*spotSourceMenu
	*logbookView
	*entryView
	*workmodeView
	*keyerView
	*statusView
	*callinfoView
	*stopKeyHandler
}

func setupMainWindow(builder *gtk.Builder, application *gtk.Application, style *style.Style, setAcceptFocus AcceptFocusFunc) *mainWindow {
	result := new(mainWindow)

	result.window = getUI(builder, "mainWindow").(*gtk.ApplicationWindow)
	result.window.SetApplication(application)
	result.window.SetDefaultSize(569, 700)

	result.mainMenu = setupMainMenu(builder, setAcceptFocus)
	result.radioMenu = setupRadioMenu(builder)
	result.spotSourceMenu = setupSpotSourceMenu(builder)
	result.logbookView = setupLogbookView(builder)
	result.entryView = setupEntryView(builder)
	result.workmodeView = setupWorkmodeView(builder)
	result.keyerView = setupKeyerView(builder)
	result.statusView = setupStatusView(builder, style.ForWidget(result.window.ToWidget()))
	result.callinfoView = setupCallinfoView(builder, style.ForWidget(result.window.ToWidget()))
	result.stopKeyHandler = setupStopKeyHandler(&result.window.Widget)

	result.window.Connect("style-updated", result.callinfoView.RefreshStyle)

	return result
}

func (w *mainWindow) Show() {
	w.window.ShowAll()
}

func (w *mainWindow) ShowFilename(filename string) {
	w.window.SetTitle(fmt.Sprintf("Hello Contest %s", filepath.Base(filename)))
}

func (w *mainWindow) UseDefaultWindowGeometry() {
	w.window.Move(300, 100)
	w.window.Window.Resize(569, 700)
}

func (w *mainWindow) ConnectToGeometry(geometry *gmtry.Geometry) {
	connectToGeometry(geometry, "main", &w.window.Window)
}

func (w *mainWindow) BringToFront() {
	w.window.Present()
}

func (w *mainWindow) SelectOpenFile(callback func(string, error), title string, dir string, extensions ...string) {
	dlg, err := gtk.FileChooserDialogNewWith1Button(title, &w.window.Window, gtk.FILE_CHOOSER_ACTION_OPEN, "Open", gtk.RESPONSE_ACCEPT)
	if err != nil {
		errors.Wrap(err, "cannot create a file selection dialog to open a file")
		return
	}
	defer dlg.Destroy()

	log.Printf("OPEN FILE in %s with extensions %v", dir, extensions)

	dlg.SetTransientFor(nil)
	dlg.SetCurrentFolder(dir)

	if len(extensions) > 0 {
		filter, err := gtk.FileFilterNew()
		if err != nil {
			callback("", errors.Wrap(err, "cannot create a file selection dialog to open a file"))
			return
		}
		for _, extension := range extensions {
			filter.AddPattern("*." + extension)
		}
		dlg.SetFilter(filter)
	}

	result := dlg.Run()
	if result != gtk.RESPONSE_ACCEPT {
		callback("", nil)
		return
	}

	callback(dlg.GetFilename(), nil)
}

func (w *mainWindow) SelectSaveFile(callback func(string, error), title string, dir string, filename string, patterns ...string) {
	dlg, err := gtk.FileChooserDialogNewWith1Button(title, &w.window.Window, gtk.FILE_CHOOSER_ACTION_SAVE, "Save", gtk.RESPONSE_ACCEPT)
	if err != nil {
		errors.Wrap(err, "cannot create a file selection dialog to save a file")
		return
	}
	defer dlg.Destroy()

	log.Printf("SAVE FILE in %s", dir)

	dlg.SetTransientFor(nil)
	dlg.SetDoOverwriteConfirmation(true)
	dlg.SetCurrentFolder(dir)
	dlg.SetCurrentName(filename)

	if len(patterns) > 0 {
		filter, err := gtk.FileFilterNew()
		if err != nil {
			errors.Wrap(err, "cannot create a file selection dialog to save a file")
			return
		}
		for _, pattern := range patterns {
			filter.AddPattern(pattern)
		}
		dlg.SetFilter(filter)
	}

	result := dlg.Run()
	if result != gtk.RESPONSE_ACCEPT {
		return
	}

	callback(dlg.GetFilename(), nil)
}

func (w *mainWindow) ShowInfoDialog(title string, format string, a ...any) {
	dlg := gtk.MessageDialogNew(w.window, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, format, a...)
	defer dlg.Destroy()
	dlg.SetTitle(title)
	dlg.SetTransientFor(nil)
	dlg.Run()
}

func (w *mainWindow) ShowErrorDialog(format string, a ...any) {
	dlg := gtk.MessageDialogNew(w.window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, format, a...)
	defer dlg.Destroy()
	dlg.SetTransientFor(nil)
	dlg.Run()
}
