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
	settings.SetValue(
		*qtlib.NewQAnyStringView3(prefix + "Geometry"),
		qtlib.NewQVariant12(window.SaveGeometry()),
	)
	settings.SetValue(
		*qtlib.NewQAnyStringView3(prefix + "State"),
		qtlib.NewQVariant12(window.SaveState()),
	)
}

func restoreWindowState(settings *qtlib.QSettings, prefix string, window *qtlib.QMainWindow) bool {
	geometryFound := false
	if v := settings.Value(*qtlib.NewQAnyStringView3(prefix + "Geometry"), qtlib.NewQVariant()); v.IsValid() {
		window.RestoreGeometry(v.ToByteArray())
		geometryFound = true
	}
	if v := settings.Value(*qtlib.NewQAnyStringView3(prefix + "State"), qtlib.NewQVariant()); v.IsValid() {
		window.RestoreState(v.ToByteArray())
	}
	return geometryFound
}
