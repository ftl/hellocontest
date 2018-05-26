package ui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
)

func getUI(builder *gtk.Builder, name string) interface{} {
	obj, err := builder.GetObject(name)
	if err != nil {
		log.Fatalf("Cannot get UI object %s: %v", name, err)
	}
	return obj
}
