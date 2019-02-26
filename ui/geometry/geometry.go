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

func (w *Window) SetPosition(x, y int) {
	if w.Maximized {
		return
	}
	w.X = x
	w.Y = y
}

func (w *Window) SetSize(width, height int) {
	if w.Maximized {
		return
	}
	w.Width = width
	w.Height = height
}

func (w *Window) SetMaximized(maximized bool) {
	w.Maximized = maximized
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

func (w Windows) Get(id ID) *Window {
	g, ok := w[id]
	if !ok {
		g = &Window{
			ID: id,
		}
		w[id] = g
	}
	return g
}
