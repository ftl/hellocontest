package ui

import (
	qtlib "github.com/mappu/miqt/qt6"
)

type dockTarget int

const (
	dockTargetMain dockTarget = iota
	dockTargetSpots
)

const (
	dockTargetSettingPrefix = "ui/dockTarget/"

	dockTargetValueMain  = "main"
	dockTargetValueSpots = "spots"
)

type dockable interface {
	Dock() *qtlib.QDockWidget
	Name() string
}

type managedDock struct {
	dock        *qtlib.QDockWidget
	name        string
	defaultArea qtlib.DockWidgetArea
	target      dockTarget
}

type dockManager struct {
	main  *qtlib.QMainWindow
	spots *spotWindow
	docks []*managedDock
}

func newDockManager(main *qtlib.QMainWindow, spots *spotWindow) *dockManager {
	return &dockManager{
		main:  main,
		spots: spots,
	}
}

func (m *dockManager) Add(view dockable, area qtlib.DockWidgetArea) {
	dock := &managedDock{
		dock:        view.Dock(),
		name:        view.Name(),
		defaultArea: area,
		target:      dockTargetMain,
	}
	m.docks = append(m.docks, dock)
	m.main.AddDockWidget(area, dock.dock)

	dock.dock.SetContextMenuPolicy(qtlib.CustomContextMenu)
	dock.dock.OnCustomContextMenuRequested(func(pos *qtlib.QPoint) {
		m.showContextMenu(dock, pos)
	})
}

func (m *dockManager) showContextMenu(dock *managedDock, pos *qtlib.QPoint) {
	target, text := dockTargetSpots, "Move to Spot Window"
	if dock.target == dockTargetSpots {
		target, text = dockTargetMain, "Move to Main Window"
	}

	menu := qtlib.NewQMenu2()
	defer menu.Delete()
	action := menu.AddActionWithText(text)
	action.OnTriggered(func() {
		m.MoveTo(dock, target)
	})
	menu.ExecWithPos(dock.dock.MapToGlobalWithQPoint(pos))
}

func (m *dockManager) MoveTo(dock *managedDock, target dockTarget) {
	if dock.target == target {
		return
	}

	m.windowFor(dock.target).RemoveDockWidget(dock.dock)
	m.windowFor(target).AddDockWidget(dock.defaultArea, dock.dock)
	dock.target = target

	if target == dockTargetSpots {
		m.spots.Show()
	}
	dock.dock.SetVisible(true)
	dock.dock.Raise()
}

func (m *dockManager) Save(settings *qtlib.QSettings) {
	for _, dock := range m.docks {
		value := dockTargetValueMain
		if dock.target == dockTargetSpots {
			value = dockTargetValueSpots
		}
		settings.SetValue(
			*qtlib.NewQAnyStringView3(dockTargetSettingPrefix + dock.name),
			qtlib.NewQVariant14(value),
		)
	}
}

func (m *dockManager) Restore(settings *qtlib.QSettings) {
	// this must run before the windows restore their state, because
	// QMainWindow.restoreState only positions the docks that belong to that window
	for _, dock := range m.docks {
		value := settings.Value(*qtlib.NewQAnyStringView3(dockTargetSettingPrefix + dock.name), qtlib.NewQVariant())
		if !value.IsValid() {
			continue
		}
		if value.ToString() == dockTargetValueSpots {
			m.MoveTo(dock, dockTargetSpots)
		} else {
			m.MoveTo(dock, dockTargetMain)
		}
	}
}

func (m *dockManager) windowFor(target dockTarget) *qtlib.QMainWindow {
	if target == dockTargetSpots {
		return m.spots.window
	}
	return m.main
}
