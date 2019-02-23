package geometry

import (
	"fmt"
	"io"
)

type ID string

type Window struct {
	ID        ID
	X         int
	Y         int
	Width     int
	Height    int
	Maximized bool
}

func (w *Window) String() string {
	return fmt.Sprintf("Window %s: (%d, %d) (%d x %d) %t", w.ID, w.X, w.Y, w.Width, w.Height, w.Maximized)
}

func (w *Window) Apply(a Applyable) {
	a.Move(w.X, w.Y)
	a.Resize(w.Width, w.Height)
	if w.Maximized {
		a.Maximize()
	}
}

func (w *Window) Observe(o Observable) func() {
	return func() {
		w.Maximized = o.IsMaximized()
		if !w.Maximized {
			w.X, w.Y = o.GetPosition()
			w.Width, w.Height = o.GetSize()
		}
		fmt.Println(w)
	}
}

type Applyable interface {
	Move(x, y int)
	Resize(width, height int)
	Maximize()
}

type Observable interface {
	GetPosition() (x, y int)
	GetSize() (width, height int)
	IsMaximized() bool
}

type Connectable interface {
	Applyable
	Observable
}

type Windows map[ID]*Window

func NewWindows() Windows {
	return make(map[ID]*Window)
}

func LoadWindows(w io.Writer) (Windows, error) {
	return NewWindows(), nil
}

func (w Windows) Store(writer io.Writer) error {
	return nil
}

func (w Windows) Connect(c Connectable, id ID) func() {
	g, ok := w[id]
	if !ok {
		g = &Window{
			ID: id,
		}
		w[id] = g
	}

	g.Apply(c)
	return g.Observe(c)
}
