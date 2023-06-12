package ui

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type RadioMenuController interface {
	SelectRadio(name string)
	SelectKeyer(name string)
}

type radioMenu struct {
	controller RadioMenuController

	parentMenu *gtk.Menu
	separator  *gtk.SeparatorMenuItem
	radioGroup *glib.SList
	keyerGroup *glib.SList

	radioItems map[string]*gtk.RadioMenuItem
	keyerItems map[string]*gtk.RadioMenuItem
}

func setupRadioMenu(builder *gtk.Builder) *radioMenu {
	result := new(radioMenu)
	result.radioItems = make(map[string]*gtk.RadioMenuItem)
	result.keyerItems = make(map[string]*gtk.RadioMenuItem)

	parentItem := getUI(builder, "menuRadio").(*gtk.MenuItem)
	parentMenu, _ := parentItem.GetSubmenu()
	result.parentMenu = parentMenu.(*gtk.Menu)
	result.separator = getUI(builder, "menuRadioSeparator").(*gtk.SeparatorMenuItem)

	return result
}

func (m *radioMenu) SetRadioMenuController(controller RadioMenuController) {
	m.controller = controller
}

func (m *radioMenu) AddRadio(name string) {
	_, ok := m.radioItems[name]
	if ok {
		return
	}

	radioItem, _ := gtk.RadioMenuItemNewWithLabel(m.radioGroup, name)
	radioGroup, _ := radioItem.GetGroup()
	if m.radioGroup == nil {
		m.radioGroup = radioGroup
	}
	radioItem.Connect("toggled", m.onRadioSelected(name, radioItem))

	m.radioItems[name] = radioItem
	m.parentMenu.Insert(radioItem, len(m.radioItems)-1)
}

func (m *radioMenu) AddKeyer(name string) {
	_, ok := m.keyerItems[name]
	if ok {
		return
	}

	keyerItem, _ := gtk.RadioMenuItemNewWithLabel(m.keyerGroup, name)
	keyerGroup, _ := keyerItem.GetGroup()
	if m.keyerGroup == nil {
		m.keyerGroup = keyerGroup
	}
	keyerItem.Connect("toggled", m.onKeyerSelected(name, keyerItem))

	m.keyerItems[name] = keyerItem
	m.parentMenu.Add(keyerItem)
}

func (m *radioMenu) SetRadioSelected(name string) {
	if name == "" {
		for _, item := range m.radioItems {
			item.SetActive(false)
		}
		return
	}

	item, ok := m.radioItems[name]
	if !ok {
		return
	}
	item.SetActive(true)
}

func (m *radioMenu) SetKeyerSelected(name string) {
	if name == "" {
		for _, item := range m.keyerItems {
			item.SetActive(false)
		}
		return
	}

	item, ok := m.keyerItems[name]
	if !ok {
		return
	}
	item.SetActive(true)
}

func (m *radioMenu) onRadioSelected(name string, item *gtk.RadioMenuItem) func() {
	return func() {
		if m.controller == nil {
			return
		}
		if !item.GetActive() {
			return
		}
		m.controller.SelectRadio(name)
	}
}

func (m *radioMenu) onKeyerSelected(name string, item *gtk.RadioMenuItem) func() {
	return func() {
		if m.controller == nil {
			return
		}
		if !item.GetActive() {
			return
		}
		m.controller.SelectKeyer(name)
	}
}
