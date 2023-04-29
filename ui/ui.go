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
	err = provider.LoadFromData(definition)
	if err != nil {
		log.Fatalf("Cannot parse CSS style: %v", err)
	}
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
	err := doWithStyle(widget, func(context *gtk.StyleContext) {
		context.AddClass(class)
	})
	if err != nil {
		log.Printf("Cannot add style class: %v", err)
	}
}

func removeStyleClass(widget *gtk.Widget, class string) {
	err := doWithStyle(widget, func(context *gtk.StyleContext) {
		context.RemoveClass(class)
	})
	if err != nil {
		log.Printf("Cannot remove style class: %v", err)
	}
}

func doWithStyle(widget *gtk.Widget, do func(*gtk.StyleContext)) error {
	context, err := widget.GetStyleContext()
	if err != nil {
		return err
	}
	do(context)
	return nil
}
