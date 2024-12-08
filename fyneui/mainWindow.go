package fyneui

import (
	"fmt"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

type mainWindow struct {
	window fyne.Window
}

func setupMainWindow(window fyne.Window, qsoList *qsoList, statusBar *statusBar) *mainWindow {
	result := &mainWindow{
		window: window,
	}
	window.SetMaster()

	root := container.NewBorder(
		nil,                 // top
		statusBar.container, // bottom
		nil,                 // left
		nil,                 // right
		qsoList.container,   // center
	)
	window.SetContent(root)

	return result
}

func (w *mainWindow) Show() {
	w.window.Show()
}

func (w *mainWindow) ShowFilename(filename string) {
	w.window.SetTitle(fmt.Sprintf("Hello Contest %s", filepath.Base(filename)))
}

func (w *mainWindow) BringToFront() {
	w.window.RequestFocus()
}

func (w *mainWindow) UseDefaultWindowGeometry() {
	w.window.Resize(fyne.NewSize(570, 700))
	w.window.CenterOnScreen()
}

func (w *mainWindow) SelectOpenFile(callback func(string, error), title string, dir string, extensions ...string) {
	dirURI, err := storage.ListerForURI(storage.NewFileURI(dir))
	if err != nil {
		callback("", err)
		return
	}
	log.Printf("OPEN FILE in %s with extensions %v", dir, extensions)

	dialogCallback := func(r fyne.URIReadCloser, err error) {
		defer func() {
			if r != nil {
				r.Close()
			}
		}()
		if err != nil {
			callback("", err)
			return
		}
		if r == nil {
			callback("", nil)
			return
		}
		filename := r.URI().Path()
		log.Printf("file selected to open: %s", filename)
		callback(filename, nil)
	}

	fileDialog := dialog.NewFileOpen(dialogCallback, w.window)
	fileDialog.SetView(dialog.ListView)
	fileDialog.Resize(fyne.NewSize(1000, 600))
	// fileDialog.SetTitleText(title) // TODO: activate with fyne 2.6
	fileDialog.SetConfirmText("Open")
	fileDialog.SetDismissText("Cancel")
	fileDialog.SetLocation(dirURI)
	if len(extensions) > 0 {
		filterExtensions := make([]string, len(extensions), len(extensions))
		for i, extension := range extensions {
			filterExtensions[i] = "." + extension
		}
		fileDialog.SetFilter(storage.NewExtensionFileFilter(filterExtensions))
	}
	fileDialog.Show()
}

func (w *mainWindow) SelectSaveFile(callback func(filename string, err error), title string, dir string, proposedFilename string, extensions ...string) {
	dirURI, err := storage.ListerForURI(storage.NewFileURI(dir))
	if err != nil {
		callback("", err)
		return
	}
	log.Printf("SAVE FILE %s in %s with extensions %v", proposedFilename, dir, extensions)

	dialogCallback := func(w fyne.URIWriteCloser, err error) {
		defer func() {
			if w != nil {
				w.Close()
			}
		}()
		if err != nil {
			callback("", err)
			return
		}
		if w == nil {
			callback("", nil)
			return
		}
		filename := w.URI().Path()
		log.Printf("file selected to save: %s", filename)
		callback(filename, nil)
	}

	fileDialog := dialog.NewFileSave(dialogCallback, w.window)
	fileDialog.SetView(dialog.ListView)
	fileDialog.Resize(fyne.NewSize(1000, 600))
	// fileDialog.SetTitleText(title) // TODO: activate with fyne 2.6
	fileDialog.SetConfirmText("Save")
	fileDialog.SetDismissText("Cancel")
	fileDialog.SetLocation(dirURI)
	if len(extensions) > 0 {
		filterExtensions := make([]string, len(extensions), len(extensions))
		for i, extension := range extensions {
			filterExtensions[i] = "." + extension
		}
		fileDialog.SetFilter(storage.NewExtensionFileFilter(filterExtensions))
	}
	fileDialog.Show()
}

func (w *mainWindow) ShowInfoDialog(title string, format string, a ...any) {
	dialog.ShowInformation(
		title,
		fmt.Sprintf(format, a...),
		w.window,
	)
}

func (w *mainWindow) ShowErrorDialog(format string, a ...any) {
	err := fmt.Errorf(format, a...)
	log.Println(err)
	dialog.ShowError(
		err,
		w.window,
	)
}
