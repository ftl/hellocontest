package ui

import (
	qtlib "github.com/mappu/miqt/qt6"
)

type RadioMenuController interface {
	SelectRadio(name string) error
	SelectKeyer(name string) error
}

type radioMenu struct {
	menu      *qtlib.QMenu
	actions   *actions
	separator *qtlib.QAction
}

func newRadioMenu(a *actions) *radioMenu {
	return &radioMenu{actions: a}
}

func (v *radioMenu) ensureSeparator() {
	if v.separator != nil {
		return
	}
	v.separator = v.menu.AddSeparator()
}

func (v *radioMenu) AddRadio(name string) {
	v.ensureSeparator()
	if _, exists := v.actions.radioActions[name]; exists {
		return
	}
	action := v.actions.AddRadioAction(name)
	v.menu.InsertAction(v.separator, action)
}

func (v *radioMenu) AddKeyer(name string) {
	v.ensureSeparator()
	if _, exists := v.actions.keyerActions[name]; exists {
		return
	}
	action := v.actions.AddKeyerAction(name)
	v.menu.AddAction(action)
}

func (v *radioMenu) SetRadioSelected(name string) {
	v.actions.SetRadioSelected(name)
}

func (v *radioMenu) SetKeyerSelected(name string) {
	v.actions.SetKeyerSelected(name)
}
