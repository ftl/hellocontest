package fyneui

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func Run(version string, sponsors string, args []string) {
	app := app.New()
	window := app.NewWindow("Hello Contest")

	window.SetContent(widget.NewLabel("Hello Contest"))
	window.ShowAndRun()
}
