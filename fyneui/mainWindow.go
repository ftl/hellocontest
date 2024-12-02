package fyneui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
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

func (w *mainWindow) SelectOpenFile(title string, dir string, patterns ...string) (string, bool, error) {
	return "", false, nil
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

func (w *mainWindow) ShowErrorDialog(string, ...interface{}) {

}
