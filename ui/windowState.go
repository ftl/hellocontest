package ui

import (
	qtlib "github.com/mappu/miqt/qt6"
)

const (
	mainWindowSettingPrefix = "ui/mainWindow"
	spotWindowSettingPrefix = "ui/spotWindow"

	settingSpotWindowVisible = "ui/spotWindowVisible"
)

func saveWindowState(settings *qtlib.QSettings, prefix string, window *qtlib.QMainWindow) {
	setSettingsBytes(settings, prefix+"Geometry", window.SaveGeometry())
	setSettingsBytes(settings, prefix+"State", window.SaveState())
}

// restoreWindowGeometry gives the window its size and its position. A window that is already
// visible keeps the position of the window manager, therefore this comes before the first show.
func restoreWindowGeometry(settings *qtlib.QSettings, prefix string, window *qtlib.QMainWindow) bool {
	v := settingsValue(settings, prefix+"Geometry")
	if !v.IsValid() {
		return false
	}
	window.RestoreGeometry(v.ToByteArray())
	return true
}

// restoreWindowLayout puts the docks of the window into their areas. QMainWindow places only the
// docks that belong to it, therefore this comes after the docks found their window.
func restoreWindowLayout(settings *qtlib.QSettings, prefix string, window *qtlib.QMainWindow) {
	v := settingsValue(settings, prefix+"State")
	if !v.IsValid() {
		return
	}
	window.RestoreState(v.ToByteArray())
}
