package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type ShortcutID string

const (
	OpenShortcut         ShortcutID = "open"
	OpenSettingsShortcut ShortcutID = "open_settings"
	QuitShortcut         ShortcutID = "quit"
)

type ShortcutConsumer interface {
	AddShortcut(fyne.Shortcut, func(shortcut fyne.Shortcut))
}

type Shortcut struct {
	desktop.CustomShortcut
	ID     ShortcutID
	Action func()
}

func (s *Shortcut) ShortcutName() string {
	return s.CustomShortcut.ShortcutName()
}

func (s *Shortcut) AddTo(consumer ShortcutConsumer) {
	consumer.AddShortcut(s, func(shortcut fyne.Shortcut) {
		s.Action()
	})
}

type ShortcutController interface {
	Open()
	OpenSettings()
	Quit()
}

type Shortcuts struct {
	shortcuts map[ShortcutID]*Shortcut
}

func setupShortcuts(controller ShortcutController) *Shortcuts {
	return &Shortcuts{
		shortcuts: map[ShortcutID]*Shortcut{
			OpenShortcut: {
				ID:             OpenShortcut,
				Action:         controller.Open,
				CustomShortcut: desktop.CustomShortcut{Modifier: fyne.KeyModifierControl, KeyName: fyne.KeyO},
			},
			OpenSettingsShortcut: {
				ID:             OpenSettingsShortcut,
				Action:         controller.OpenSettings,
				CustomShortcut: desktop.CustomShortcut{Modifier: fyne.KeyModifierControl, KeyName: fyne.KeyPeriod},
			},
			QuitShortcut: {
				ID:             QuitShortcut,
				Action:         controller.Quit,
				CustomShortcut: desktop.CustomShortcut{Modifier: fyne.KeyModifierControl, KeyName: fyne.KeyQ},
			},
		},
	}
}

func (s *Shortcuts) Get(id ShortcutID) fyne.Shortcut {
	return s.shortcuts[id]
}

func (s *Shortcuts) AddTo(consumer ShortcutConsumer) {
	for _, shortcut := range s.shortcuts {
		shortcut.AddTo(consumer)
	}
}
