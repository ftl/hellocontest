package style

import (
	_ "embed"
	"fmt"
	"log"
	"math"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type Class string

//go:embed contest.css
var css string

type Color struct{ R, G, B, A float64 }

var (
	Black = Color{}
	White = Color{1, 1, 1, 1}
	Red   = Color{1, 0, 0, 1}
	Green = Color{0, 1, 0, 1}
	Blue  = Color{0, 0, 1, 1}
)

func NewRGB(r, g, b float64) Color {
	return NewRGBA(r, g, b, 1)
}

func NewRGBA(r, g, b, a float64) Color {
	return Color{r, g, b, a}
}

func colorFromGDK(rgba *gdk.RGBA) Color {
	return Color{rgba.GetRed(), rgba.GetGreen(), rgba.GetBlue(), rgba.GetAlpha()}
}

func (c Color) ToRGB() (r, g, b float64) {
	return c.R, c.G, c.B
}

func (c Color) ToRGBA() (r, g, b, a float64) {
	return c.R, c.G, c.B, c.A
}

func (c Color) ToWeb() string {
	return fmt.Sprintf("#%02x%02x%02x", toByte(c.R), toByte(c.G), toByte(c.B))
}

func toByte(f float64) byte {
	return byte(f * 255.0)
}

func (c Color) WithAlpha(alpha float64) (r, g, b, a float64) {
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

func (s *Style) ForWidget(widget *gtk.Widget) *WidgetStyle {
	return &WidgetStyle{
		style:  s,
		widget: widget,
	}
}

func (s *Style) AddToWidget(widget *gtk.Widget) {
	context, err := widget.GetStyleContext()
	if err != nil {
		log.Printf("AddToWidget: cannot get style context: %v", err)
		return
	}
	context.AddProvider(s.provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func (s *Style) AddToScreen(screen *gdk.Screen) {
	gtk.AddProviderForScreen(screen, s.provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func (s *Style) HasColor(widget *gtk.Widget, name string) bool {
	context, err := widget.GetStyleContext()
	if err != nil {
		log.Printf("HasColor: cannot get style context: %v", err)
		return false
	}
	_, found := context.LookupColor(name)
	return found
}

func (s *Style) ColorByName(widget *gtk.Widget, name string) Color {
	context, err := widget.GetStyleContext()
	if err != nil {
		log.Printf("FindColor: cannot get style context: %v", err)
		return Black
	}
	rgba, found := context.LookupColor(name)
	if !found {
		log.Printf("FindColor: cannot find color %s", name)
		return Red
	}
	return colorFromGDK(rgba)
}

func (s *Style) BackgroundColor(widget *gtk.Widget) Color {
	return s.ColorByName(widget, "theme_bg_color")
}

func (s *Style) ForegroundColor(widget *gtk.Widget) Color {
	return s.ColorByName(widget, "theme_fg_color")
}

func (s *Style) TextColor(widget *gtk.Widget) Color {
	return s.ColorByName(widget, "theme_text_color")
}

type WidgetStyle struct {
	style  *Style
	widget *gtk.Widget
}

func (s *WidgetStyle) HasColor(name string) bool {
	return s.style.HasColor(s.widget, name)
}

func (s *WidgetStyle) ColorByName(name string) Color {
	return s.style.ColorByName(s.widget, name)
}

func (s *WidgetStyle) BackgroundColor() Color {
	return s.style.BackgroundColor(s.widget)
}

func (s *WidgetStyle) ForegroundColor() Color {
	return s.style.ForegroundColor(s.widget)
}

func (s *WidgetStyle) TextColor() Color {
	return s.style.TextColor(s.widget)
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
