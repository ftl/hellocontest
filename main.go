package main

import (
	"os"

	"github.com/ftl/hellocontest/ui"
)

var version = "development"

func main() {
	ui.Run(version, os.Args)
}
