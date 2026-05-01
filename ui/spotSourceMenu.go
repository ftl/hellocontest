package ui

import (
	qtlib "github.com/mappu/miqt/qt6"
)

type SpotSourceMenuController interface {
	SetSpotSourceEnabled(name string, enabled bool)
}

type spotSourceMenu struct {
	menu    *qtlib.QMenu
	actions *actions
}

func newSpotSourceMenu(a *actions) *spotSourceMenu {
	return &spotSourceMenu{actions: a}
}

func (v *spotSourceMenu) AddSpotSourceEntry(name string) {
	if _, exists := v.actions.spotSourceActions[name]; exists {
		return
	}
	action := v.actions.AddSpotSourceAction(name)
	v.menu.AddAction(action)
}

func (v *spotSourceMenu) SetSpotSourceEnabled(name string, enabled bool) {
	v.actions.SetSpotSourceEnabled(name, enabled)
}
