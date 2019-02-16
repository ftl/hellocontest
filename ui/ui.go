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

func newStyle(definition string) *style {
	provider, err := gtk.CssProviderNew()
	if err != nil {
		log.Fatalf("Cannot create CSS provider: %v", err)
	}
	provider.LoadFromData(definition)
	return &style{provider: provider}
}

type style struct {
	provider *gtk.CssProvider
}

func (s *style) applyTo(widget *gtk.Widget) {
	context, err := widget.GetStyleContext()
	if err != nil {
		log.Printf("Cannot get style context: %v", err)
		return
	}
	context.AddProvider(s.provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func addStyleClass(widget *gtk.Widget, class string) {
	doWithStyle(widget, func(context *gtk.StyleContext) {
		context.AddClass(class)
	})
}

func removeStyleClass(widget *gtk.Widget, class string) {
	doWithStyle(widget, func(context *gtk.StyleContext) {
		context.RemoveClass(class)
	})
}

func doWithStyle(widget *gtk.Widget, do func(*gtk.StyleContext)) error {
	context, err := widget.GetStyleContext()
	if err != nil {
		return err
	}
	do(context)
	return nil
}
