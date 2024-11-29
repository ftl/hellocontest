//go:build fyne

package main

import "github.com/ftl/hellocontest/fyneui"

func init() {
	run = fyneui.Run
}
