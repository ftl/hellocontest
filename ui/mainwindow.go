package ui

import (
	"fmt"
	"path/filepath"

	"github.com/ftl/hellocontest/ui/geometry"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/pkg/errors"
)

var geometryID = geometry.ID("main")

type mainWindow struct {
	window *gtk.ApplicationWindow

	*mainMenu
	*logView
	*entryView
	*keyerView
}

func setupMainWindow(builder *gtk.Builder, application *gtk.Application, windowGeometry geometry.Windows) *mainWindow {
	result := new(mainWindow)

	result.window = getUI(builder, "mainWindow").(*gtk.ApplicationWindow)
	result.window.SetApplication(application)
	result.window.SetDefaultSize(500, 500)

	geo := windowGeometry.Get(geometryID)
	result.window.Move(geo.X, geo.Y)
	result.window.Resize(geo.Width, geo.Height)
	if geo.Maximized {
		result.window.Maximize()
	}

	result.window.Connect("configure-event", func(_ interface{}, event *gdk.Event) {
		e := gdk.EventConfigureNewFromEvent(event)
		w := windowGeometry.Get(geometryID)
		w.SetPosition(e.X(), e.Y())
		w.SetSize(e.Width(), e.Height())
	})
	result.window.Connect("window-state-event", func(_ interface{}, event *gdk.Event) {
		e := gdk.EventWindowStateNewFromEvent(event)
		if e.ChangedMask()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED {
			windowGeometry.Get(geometryID).SetMaximized(e.NewWindowState()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED)
		}
	})

	result.mainMenu = setupMainMenu(builder)
	result.logView = setupLogView(builder)
	result.entryView = setupEntryView(builder)
	result.keyerView = setupKeyerView(builder)

	return result
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
	if result != gtk.RESPONSE_ACCEPT {
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
	if result != gtk.RESPONSE_ACCEPT {
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
