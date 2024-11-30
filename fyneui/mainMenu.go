package fyneui

import "fyne.io/fyne/v2"

type MainMenuController interface {
	Quit()
}

type mainMenu struct {
	controller MainMenuController

	fileQuit *fyne.MenuItem
}

func setupMainMenu(mainWindow fyne.Window, controller MainMenuController) *mainMenu {
	result := &mainMenu{
		controller: controller,
	}

	result.fileQuit = fyne.NewMenuItem("Quit", result.onFileQuit)
	result.fileQuit.IsQuit = true

	fileMenu := fyne.NewMenu("File", result.fileQuit)
	mainMenu := fyne.NewMainMenu(fileMenu)
	mainWindow.SetMainMenu(mainMenu)

	return result
}

func (m *mainMenu) onFileQuit() {
	m.controller.Quit()
}
