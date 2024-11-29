//go:build !fyne

package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

type SpotSourceMenuController interface {
	SetSpotSourceEnabled(string, bool)
}

type spotSourceMenu struct {
	controller SpotSourceMenuController

	parentMenu *gtk.Menu

	items map[string]*gtk.CheckMenuItem
}

func setupSpotSourceMenu(builder *gtk.Builder) *spotSourceMenu {
	result := new(spotSourceMenu)
	result.items = make(map[string]*gtk.CheckMenuItem)

	parentItem := getUI(builder, "menuBandmap").(*gtk.MenuItem)
	parentMenu, _ := parentItem.GetSubmenu()
	result.parentMenu = parentMenu.(*gtk.Menu)

	return result
}

func (m *spotSourceMenu) SetSpotSourceMenuController(controller SpotSourceMenuController) {
	m.controller = controller
}

func (m *spotSourceMenu) AddSpotSourceEntry(name string) {
	_, ok := m.items[name]
	if ok {
		return
	}

	checkItem, _ := gtk.CheckMenuItemNewWithLabel(fmt.Sprintf("Use %s", name))
	checkItem.Connect("toggled", m.onEnableSpotSource(name, checkItem))

	m.items[name] = checkItem
	m.parentMenu.Add(checkItem)
}

func (m *spotSourceMenu) onEnableSpotSource(name string, item *gtk.CheckMenuItem) func() {
	return func() {
		if m.controller == nil {
			return
		}
		m.controller.SetSpotSourceEnabled(name, item.GetActive())
	}
}

func (m *spotSourceMenu) SetSpotSourceEnabled(name string, enabled bool) {
	item, ok := m.items[name]
	if !ok {
		return
	}
	item.SetActive(enabled)
}
