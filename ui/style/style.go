package style

import (
	_ "embed"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type Class string

//go:embed contest.css
var css string

type Style struct {
	provider *gtk.CssProvider
}

func New() *Style {
	provider, err := gtk.CssProviderNew()
	if err != nil {
		log.Fatalf("Cannot create CSS provider: %v", err)
	}
	err = provider.LoadFromData(css)
	if err != nil {
		log.Fatalf("Cannot parse CSS style: %v", err)
	}
	return &Style{provider: provider}
}

func (s *Style) AddToWidget(widget *gtk.Widget) {
	context, err := widget.GetStyleContext()
	if err != nil {
		log.Printf("Cannot get style context: %v", err)
		return
	}
	context.AddProvider(s.provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func (s *Style) AddToScreen(screen *gdk.Screen) {
	gtk.AddProviderForScreen(screen, s.provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func AddClass(widget *gtk.Widget, class Class) {
	err := doWithStyle(widget, func(context *gtk.StyleContext) {
		context.AddClass(string(class))
	})
	if err != nil {
		log.Printf("Cannot add style class: %v", err)
	}
}

func RemoveClass(widget *gtk.Widget, class Class) {
	err := doWithStyle(widget, func(context *gtk.StyleContext) {
		context.RemoveClass(string(class))
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
