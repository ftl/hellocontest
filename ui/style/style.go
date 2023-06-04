package style

import (
	_ "embed"
	"log"
	"math"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type Class string

//go:embed contest.css
var css string

type Color struct{ R, G, B float64 }

var (
	Black = Color{}
	White = Color{1, 1, 1}
	Red   = Color{1, 0, 0}
	Green = Color{0, 1, 0}
	Blue  = Color{0, 0, 1}
)

func (c Color) ToRGB() (r, g, b float64) {
	return c.R, c.G, c.B
}

func (c Color) ToRGBA(alpha float64) (r, g, b, a float64) {
	return c.R, c.G, c.B, alpha
}

type ColorMap []Color

func (c ColorMap) ToRGB(f float64) (r, g, b float64) {
	f = math.Abs(f)
	adaptedHeat := float64(f) * float64(len(c)-1)
	colorIndex := int(adaptedHeat)
	lower := c[int(math.Min(float64(colorIndex), float64(len(c)-1)))]
	upper := c[int(math.Min(float64(colorIndex+1), float64(len(c)-1)))]
	p := adaptedHeat - float64(colorIndex)
	r = (1-p)*lower.R + p*upper.R
	g = (1-p)*lower.G + p*upper.G
	b = (1-p)*lower.B + p*upper.B
	return
}

func (c ColorMap) ToRGBA(f float64, alpha float64) (r, g, b, a float64) {
	r, g, b = c.ToRGB(f)
	a = alpha
	return
}

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

func (s *Style) FindColor(name string) Color {
	return Black
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
