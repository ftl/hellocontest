package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/ftl/hellocontest/script"
	"github.com/ftl/hellocontest/ui"
)

var version = "development"

//go:embed sponsors.txt
var sponsors string

func main() {
	var startupScript ui.Script = nil
	args := os.Args

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Println(version)
			os.Exit(0)
		case "sponsors":
			fmt.Printf("This release of Hello Contest is sponsored by:\n%s\nThank you all for your great support!\n73, Florian\n", sponsors)
			os.Exit(0)
		case "screenshots":
			startupScript = script.ScreenshotsScript
			args = args[0:1]
		}
	}

	ui.Run(version, sponsors, startupScript, args)
}
