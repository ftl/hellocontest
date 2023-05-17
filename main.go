package main

import (
	"fmt"
	"os"

	"github.com/ftl/hellocontest/ui"
)

var version = "development"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println(version)
		os.Exit(0)
	}

	ui.Run(version, os.Args)
}
