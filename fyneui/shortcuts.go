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

	WorkmodeSearchPounceShortcut ShortcutID = "workmode_sp"
	WorkmodeRunShortcut          ShortcutID = "workmode_run"

	SendMacroF1Shortcut ShortcutID = "macro_f1"
	SendMacroF2Shortcut ShortcutID = "macro_f2"
	SendMacroF3Shortcut ShortcutID = "macro_f3"
	SendMacroF4Shortcut ShortcutID = "macro_f4"
	StopTXShortcut      ShortcutID = "stop_tx"
)

type ShortcutConsumer interface {
	AddShortcut(fyne.Shortcut, func(shortcut fyne.Shortcut))
	SetOnTypedKey(func(*fyne.KeyEvent))
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

	SwitchToSPWorkmode()
	SwitchToRunWorkmode()
}

type Shortcuts struct {
	controller      ShortcutController
	keyerController KeyerController

	shortcuts map[ShortcutID]*Shortcut
}

func setupShortcuts(controller ShortcutController, keyerController KeyerController) *Shortcuts {
	return &Shortcuts{
		controller:      controller,
		keyerController: keyerController,
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
			WorkmodeSearchPounceShortcut: {
				ID:             WorkmodeSearchPounceShortcut,
				Action:         controller.SwitchToSPWorkmode,
				CustomShortcut: desktop.CustomShortcut{Modifier: fyne.KeyModifierControl, KeyName: fyne.KeyS},
			},
			WorkmodeRunShortcut: {
				ID:             WorkmodeRunShortcut,
				Action:         controller.SwitchToRunWorkmode,
				CustomShortcut: desktop.CustomShortcut{Modifier: fyne.KeyModifierControl, KeyName: fyne.KeyR},
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
	consumer.SetOnTypedKey(s.onTypedKey)
}

func (s *Shortcuts) onTypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyF1:
		s.keyerController.Send(0)
	case fyne.KeyF2:
		s.keyerController.Send(1)
	case fyne.KeyF3:
		s.keyerController.Send(2)
	case fyne.KeyF4:
		s.keyerController.Send(3)
	case fyne.KeyEscape:
		s.keyerController.Stop()
	}
}
