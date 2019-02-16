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

func newStyle(style string) *gtk.CssProvider {
	provider, err := gtk.CssProviderNew()
	if err != nil {
		log.Fatalf("Cannot create CSS provider: %v", err)
	}
	provider.LoadFromData(style)
	return provider
}

func addStyleProvider(widget *gtk.Widget, styleProvider *gtk.CssProvider) {
	context, err := widget.GetStyleContext()
	if err != nil {
		log.Printf("Cannot get style context: %v", err)
		return
	}
	context.AddProvider(styleProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func addStyleClass(widget *gtk.Widget, class string) {
	doWithStyle(widget, func(style *gtk.StyleContext) {
		style.AddClass(class)
	})
}

func removeStyleClass(widget *gtk.Widget, class string) {
	doWithStyle(widget, func(style *gtk.StyleContext) {
		style.RemoveClass(class)
	})
}

func doWithStyle(widget *gtk.Widget, do func(*gtk.StyleContext)) error {
	style, err := widget.GetStyleContext()
	if err != nil {
		return err
	}
	do(style)
	return nil
}
