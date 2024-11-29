//go:build !fyne

package main

import "github.com/ftl/hellocontest/ui"

func init() {
	run = ui.Run
}
