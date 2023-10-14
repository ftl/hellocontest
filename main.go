package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/ftl/hellocontest/ui"
)

var version = "development"

//go:embed sponsors.txt
var sponsors string

func main() {
	if len(os.Args) > 1 {
		switch {
		case os.Args[1] == "version":
			fmt.Println(version)
			os.Exit(0)
		case os.Args[1] == "sponsors":
			fmt.Printf("This release of Hello Contest is sponsored by %s\n", sponsors)
			os.Exit(0)
		}
	}

	ui.Run(version, sponsors, os.Args)
}
