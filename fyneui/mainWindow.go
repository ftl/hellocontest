package fyneui

import (
	"fmt"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type mainWindow struct {
	window fyne.Window
}

func setupMainWindow(window fyne.Window) *mainWindow {
	result := &mainWindow{
		window: window,
	}
	window.SetMaster()
	window.SetContent(widget.NewLabel("Hello Contest"))

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

	dialogCallback := func(reader fyne.URIReadCloser, err error) {
		defer func() {
			if reader != nil {
				reader.Close()
			}
		}()
		if err != nil {
			callback("", err)
			return
		}
		if reader == nil {
			callback("", nil)
			return
		}
		filename := reader.URI().Path()
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

func (w *mainWindow) SelectSaveFile(title string, dir string, filename string, patterns ...string) (string, bool, error) {
	return "", false, nil
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
