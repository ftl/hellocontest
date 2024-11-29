//go:build !fyne

package glade

import (
	_ "embed"
)

//go:embed contest.glade
var Assets string
