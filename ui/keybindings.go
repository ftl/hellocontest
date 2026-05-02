package ui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	qtlib "github.com/mappu/miqt/qt6"
)

// ActionInfo holds the display information for a configurable action.
type ActionInfo struct {
	ID              string
	Name            string
	Tooltip         string
	DefaultShortcut string
}

// PrintKeybindingsTable writes a Markdown table of all actions sorted by ID to stdout.
// It initializes Qt internally and does not require a running application.
func PrintKeybindingsTable() {
	qApp := qtlib.NewQApplication(os.Args)
	defer qApp.Delete()
	parent := qtlib.NewQWidget2()
	defer parent.Delete()

	a := newActions(parent, nil, map[string]string{})

	infos := make([]ActionInfo, len(a.allInfos))
	copy(infos, a.allInfos)
	sort.Slice(infos, func(i, j int) bool { return infos[i].ID < infos[j].ID })

	idW, nameW, shortcutW := len("Action ID"), len("Name"), len("Default Shortcut")
	displayName := func(info ActionInfo) string {
		if info.Tooltip != "" {
			return info.Tooltip
		}
		return info.Name
	}

	for _, info := range infos {
		if len(info.ID) > idW {
			idW = len(info.ID)
		}
		if n := len(displayName(info)); n > nameW {
			nameW = n
		}
		if len(info.DefaultShortcut) > shortcutW {
			shortcutW = len(info.DefaultShortcut)
		}
	}

	row := func(id, name, shortcut string) {
		fmt.Printf("| %-*s | %-*s | %-*s |\n", idW, id, nameW, name, shortcutW, shortcut)
	}

	row("Action ID", "Name", "Default Shortcut")
	fmt.Printf("|-%s-|-%s-|-%s-|\n",
		strings.Repeat("-", idW),
		strings.Repeat("-", nameW),
		strings.Repeat("-", shortcutW))
	for _, info := range infos {
		row(info.ID, displayName(info), info.DefaultShortcut)
	}
}
