package ui

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"unsafe"

	qtlib "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
)

type CaptureMode int

const (
	CaptureAuto CaptureMode = iota
	CaptureMainWindow
	CaptureActiveWindow
	CaptureWidget
	CaptureRect
	CaptureWidgetUnion
	CaptureActivePopup
)

type Marker struct {
	Number     int
	X, Y       int
	WidgetName string
	DX, DY     int
	// a frame marks a whole widget instead of pointing at a spot with a number
	Frame   bool
	W, H    int
	Padding int
}

type Rect struct{ X, Y, W, H int }

type ScreenshotRequest struct {
	Name        string
	Mode        CaptureMode
	WidgetName  string
	WidgetNames []string
	ParentName  string
	Rect        Rect
	Padding     int
}

type Pixmap any

type Screenshotter interface {
	Capture(req ScreenshotRequest) (Pixmap, error)
	Annotate(p Pixmap, markers []Marker) error
	Save(p Pixmap, path string) error
	ShowMenu(name string) error
	HideMenu(name string) error
	HighlightMenuAction(menuName, actionTitle string) error
	SetWidgetVisible(objectName string, visible bool) error
	HideDocks() error
	ResizeMainWindow(width, height int) error
	ResizeWidget(objectName string, width, height int) error
	SetChecked(objectName string, checked bool) error
	ResizeDocks(objectNames []string, sizes []int, horizontal bool) error
	MoveDock(objectName string, area string) error
	SplitDocks(firstObjectName, secondObjectName string, horizontal bool) error
	RejectDialog(objectName string) error
	RestoreWindowState() error
}

type capturedImage struct {
	pm               *qtlib.QPixmap
	root             *qtlib.QWidget
	originX, originY int
}

type screenshotter struct{ app *application }

func newScreenshotter(a *application) *screenshotter { return &screenshotter{app: a} }

func runOnMainSync(work func() error) error {
	ch := make(chan error, 1)
	mainthread.Start(func() { ch <- work() })
	return <-ch
}

const (
	markerDiameter            = 36
	markerLineWidth           = 3.0
	markerFontSize            = 14
	markerR, markerG, markerB = 220, 40, 40
	markerFrameRadius         = 4.0
)

func (s *screenshotter) Capture(req ScreenshotRequest) (Pixmap, error) {
	var img *capturedImage
	err := runOnMainSync(func() error {
		switch req.Mode {
		case CaptureRect, CaptureWidgetUnion:
			parent := s.app.window.QWidget
			if req.ParentName != "" {
				p := findWidgetByObjectName(req.ParentName)
				if p == nil {
					return fmt.Errorf("parent widget %q not found", req.ParentName)
				}
				parent = p
			}
			pmFull := parent.Grab()
			if pmFull == nil {
				return errors.New("Grab returned nil")
			}
			parentGlobal := parent.MapToGlobalWithQPoint(gc(qtlib.NewQPoint2(0, 0)))
			parentSize := parent.Size()
			parentW, parentH := parentSize.Width(), parentSize.Height()

			var x0, y0, x1, y1 int
			if req.Mode == CaptureRect {
				x0 = req.Rect.X
				y0 = req.Rect.Y
				x1 = req.Rect.X + req.Rect.W
				y1 = req.Rect.Y + req.Rect.H
			} else {
				if len(req.WidgetNames) == 0 {
					return errors.New("CaptureWidgetUnion: WidgetNames empty")
				}
				havePoint := false
				for _, name := range req.WidgetNames {
					w := findWidgetByObjectName(name)
					if w == nil {
						return fmt.Errorf("widget %q not found", name)
					}
					tlGlobal := w.MapToGlobalWithQPoint(gc(qtlib.NewQPoint2(0, 0)))
					sz := w.Size()
					lx := tlGlobal.X() - parentGlobal.X()
					ly := tlGlobal.Y() - parentGlobal.Y()
					rx := lx + sz.Width()
					ry := ly + sz.Height()
					if !havePoint {
						x0, y0, x1, y1 = lx, ly, rx, ry
						havePoint = true
					} else {
						if lx < x0 {
							x0 = lx
						}
						if ly < y0 {
							y0 = ly
						}
						if rx > x1 {
							x1 = rx
						}
						if ry > y1 {
							y1 = ry
						}
					}
				}
			}

			if req.Padding != 0 {
				x0 -= req.Padding
				y0 -= req.Padding
				x1 += req.Padding
				y1 += req.Padding
			}
			if x0 < 0 {
				x0 = 0
			}
			if y0 < 0 {
				y0 = 0
			}
			if x1 > parentW {
				x1 = parentW
			}
			if y1 > parentH {
				y1 = parentH
			}
			if x1 <= x0 || y1 <= y0 {
				return fmt.Errorf("empty crop rect after clamp: (%d,%d)-(%d,%d)", x0, y0, x1, y1)
			}

			ratio := pmFull.DevicePixelRatio()
			dx := int(float64(x0) * ratio)
			dy := int(float64(y0) * ratio)
			dw := int(float64(x1-x0) * ratio)
			dh := int(float64(y1-y0) * ratio)
			pmCrop := pmFull.Copy(dx, dy, dw, dh)
			if pmCrop == nil {
				return errors.New("QPixmap.Copy returned nil")
			}
			pmCrop.SetDevicePixelRatio(ratio)
			img = &capturedImage{pm: pmCrop, root: parent, originX: x0, originY: y0}
			return nil
		}

		var w *qtlib.QWidget
		switch req.Mode {
		case CaptureWidget:
			w = findWidgetByObjectName(req.WidgetName)
			if w == nil {
				return fmt.Errorf("widget %q not found", req.WidgetName)
			}
		case CaptureMainWindow:
			w = s.app.window.QWidget
		case CaptureActiveWindow:
			w = qtlib.QApplication_ActiveWindow()
			if w == nil {
				w = s.app.window.QWidget
			}
		case CaptureActivePopup:
			w = qtlib.QApplication_ActivePopupWidget()
			if w == nil {
				return errors.New("no active popup widget")
			}
		default:
			if m := qtlib.QApplication_ActiveModalWidget(); m != nil {
				w = m
			} else {
				w = s.app.window.QWidget
			}
		}
		pm := w.Grab()
		if pm == nil {
			return errors.New("Grab returned nil")
		}
		img = &capturedImage{pm: pm, root: w}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (s *screenshotter) ShowMenu(name string) error {
	return runOnMainSync(func() error {
		menu := s.app.findMenu(name)
		if menu == nil {
			return fmt.Errorf("menu %q not found", name)
		}
		bar := s.app.mainMenu.menuBar
		pos := bar.MapToGlobalWithQPoint(gc(qtlib.NewQPoint2(0, bar.Height())))
		menu.Popup(pos)
		return nil
	})
}

func (s *screenshotter) HideMenu(name string) error {
	return runOnMainSync(func() error {
		menu := s.app.findMenu(name)
		if menu == nil {
			return fmt.Errorf("menu %q not found", name)
		}
		menu.Hide()
		return nil
	})
}

func (s *screenshotter) HighlightMenuAction(menuName, actionTitle string) error {
	return runOnMainSync(func() error {
		menu := s.app.findMenu(menuName)
		if menu == nil {
			return fmt.Errorf("menu %q not found", menuName)
		}
		want := stripMnemonics(actionTitle)
		for _, a := range menu.Actions() {
			if stripMnemonics(a.Text()) == want {
				menu.SetActiveAction(a)
				return nil
			}
		}
		return fmt.Errorf("action %q not found in menu %q", actionTitle, menuName)
	})
}

func (s *screenshotter) SetWidgetVisible(objectName string, visible bool) error {
	return runOnMainSync(func() error {
		w := findWidgetByObjectName(objectName)
		if w == nil {
			return fmt.Errorf("widget %q not found", objectName)
		}
		if visible {
			// a dock stays invisible if the window that contains it is hidden
			w.Window().SetVisible(true)
		}
		w.SetVisible(visible)
		if visible {
			w.Raise()
		}
		return nil
	})
}

// RestoreWindowState puts the saved layout back. Qt does not restore the size of a dock
// that is shown again after it was hidden, therefore the saved state decides how the main
// window looks with its views.
func (s *screenshotter) RestoreWindowState() error {
	return runOnMainSync(func() error {
		if s.app.startupScript == nil {
			return errors.New("no script, no saved window state")
		}
		s.app.restoreWindowStateFromString(s.app.startupScript.WindowState())
		return nil
	})
}

// RejectDialog ends a modal dialog that runs its own event loop, where hiding the widget
// would leave the loop running.
func (s *screenshotter) RejectDialog(objectName string) error {
	return runOnMainSync(func() error {
		w := findWidgetByObjectName(objectName)
		if w == nil {
			return fmt.Errorf("dialog %q not found", objectName)
		}
		dialog := qtlib.UnsafeNewQDialog(unsafe.Pointer(w.UnsafePointer()))
		dialog.Reject()
		return nil
	})
}

func dockArea(name string) (qtlib.DockWidgetArea, error) {
	switch name {
	case "top":
		return qtlib.TopDockWidgetArea, nil
	case "bottom":
		return qtlib.BottomDockWidgetArea, nil
	case "left":
		return qtlib.LeftDockWidgetArea, nil
	case "right":
		return qtlib.RightDockWidgetArea, nil
	}
	return 0, fmt.Errorf("unknown dock area %q", name)
}

// MoveDock puts a dock into another area of the main window, which decides whether the
// view is wide and flat or narrow and tall
func (s *screenshotter) MoveDock(objectName string, area string) error {
	return runOnMainSync(func() error {
		dock := s.app.dockManager.find(objectName)
		if dock == nil {
			return fmt.Errorf("dock %q not found", objectName)
		}
		dockWidgetArea, err := dockArea(area)
		if err != nil {
			return err
		}
		s.app.window.AddDockWidget(dockWidgetArea, dock)
		return nil
	})
}

// SplitDocks puts the second dock beside the first one, in the area of the first
func (s *screenshotter) SplitDocks(firstObjectName, secondObjectName string, horizontal bool) error {
	return runOnMainSync(func() error {
		first := s.app.dockManager.find(firstObjectName)
		second := s.app.dockManager.find(secondObjectName)
		if first == nil || second == nil {
			return fmt.Errorf("dock %q or %q not found", firstObjectName, secondObjectName)
		}
		orientation := qtlib.Vertical
		if horizontal {
			orientation = qtlib.Horizontal
		}
		s.app.window.SplitDockWidget(first, second, orientation)
		return nil
	})
}

// ResizeDocks gives the docks the size they need, because Qt does not restore the size of
// a dock that is shown again after it was hidden. Horizontal sizes are widths, vertical
// sizes are heights.
func (s *screenshotter) ResizeDocks(objectNames []string, sizes []int, horizontal bool) error {
	return runOnMainSync(func() error {
		if len(objectNames) != len(sizes) {
			return fmt.Errorf("%d docks, but %d sizes", len(objectNames), len(sizes))
		}
		docks := make([]*qtlib.QDockWidget, 0, len(objectNames))
		for _, name := range objectNames {
			dock := s.app.dockManager.find(name)
			if dock == nil {
				return fmt.Errorf("dock %q not found", name)
			}
			docks = append(docks, dock)
		}
		orientation := qtlib.Vertical
		if horizontal {
			orientation = qtlib.Horizontal
		}
		s.app.window.ResizeDocks(docks, sizes, orientation)
		return nil
	})
}

// SetChecked toggles a named checkable widget as the operator would, for the states that
// the model keeps to itself because normally the widget tells the model, not the other
// way round
func (s *screenshotter) SetChecked(objectName string, checked bool) error {
	return runOnMainSync(func() error {
		w := findWidgetByObjectName(objectName)
		if w == nil {
			return fmt.Errorf("widget %q not found", objectName)
		}
		button := qtlib.UnsafeNewQAbstractButton(unsafe.Pointer(w.UnsafePointer()))
		button.SetChecked(checked)
		return nil
	})
}

// ResizeWidget sizes any named widget, for the windows beside the main window
func (s *screenshotter) ResizeWidget(objectName string, width, height int) error {
	return runOnMainSync(func() error {
		w := findWidgetByObjectName(objectName)
		if w == nil {
			return fmt.Errorf("widget %q not found", objectName)
		}
		w.Resize(width, height)
		return nil
	})
}

func (s *screenshotter) ResizeMainWindow(width, height int) error {
	return runOnMainSync(func() error {
		s.app.window.Resize(width, height)
		return nil
	})
}

func (s *screenshotter) HideDocks() error {
	return runOnMainSync(func() error {
		for _, dock := range s.app.dockManager.docks {
			dock.dock.SetVisible(false)
		}
		return nil
	})
}

func stripMnemonics(s string) string {
	return strings.ReplaceAll(s, "&", "")
}

func findWidgetByObjectName(name string) *qtlib.QWidget {
	var match *qtlib.QWidget
	hits := 0
	for _, w := range qtlib.QApplication_AllWidgets() {
		if w.ObjectName() == name {
			if match == nil {
				match = w
			}
			hits++
		}
	}
	if hits > 1 {
		log.Printf("ObjectName %q matched %d widgets; using first", name, hits)
	}
	return match
}

func (s *screenshotter) Annotate(p Pixmap, markers []Marker) error {
	img, ok := p.(*capturedImage)
	if !ok {
		return errors.New("invalid Pixmap handle")
	}
	return runOnMainSync(func() error {
		resolved := make([]Marker, 0, len(markers))
		rootGlobal := img.root.MapToGlobalWithQPoint(gc(qtlib.NewQPoint2(0, 0)))
		for _, m := range markers {
			if m.WidgetName != "" {
				w := findWidgetByObjectName(m.WidgetName)
				if w == nil {
					return fmt.Errorf("marker widget %q not found", m.WidgetName)
				}
				g := w.MapToGlobalWithQPoint(gc(qtlib.NewQPoint2(m.DX, m.DY)))
				m.X = g.X() - rootGlobal.X() - img.originX
				m.Y = g.Y() - rootGlobal.Y() - img.originY
				if m.Frame {
					size := w.Size()
					m.X -= m.Padding
					m.Y -= m.Padding
					m.W = size.Width() + 2*m.Padding
					m.H = size.Height() + 2*m.Padding
				}
			}
			resolved = append(resolved, m)
		}

		painter := qtlib.NewQPainter2(img.pm.QPaintDevice)
		defer painter.Delete()
		defer painter.End()
		painter.SetRenderHint2(qtlib.QPainter__Antialiasing, true)
		painter.SetRenderHint2(qtlib.QPainter__TextAntialiasing, true)

		circleColor := gc(qtlib.NewQColor11(markerR, markerG, markerB, 230))
		textColor := gc(qtlib.NewQColor11(255, 255, 255, 255))
		outlineColor := gc(qtlib.NewQColor11(255, 255, 255, 255))
		outlinePen := gc(qtlib.NewQPen3(outlineColor))
		outlinePen.SetWidthF(markerLineWidth)
		fillBrush := gc(qtlib.NewQBrush3(circleColor))
		textPen := gc(qtlib.NewQPen3(textColor))
		font := gc(qtlib.NewQFont7("Sans", markerFontSize, int(qtlib.QFont__Bold)))

		framePen := gc(qtlib.NewQPen3(circleColor))
		framePen.SetWidthF(markerLineWidth)

		for _, m := range resolved {
			if m.Frame {
				painter.SetPenWithPen(framePen)
				painter.SetBrush(gc(qtlib.NewQBrush()))
				painter.DrawRoundedRect2(m.X, m.Y, m.W, m.H, markerFrameRadius, markerFrameRadius)
				continue
			}
			painter.SetPenWithPen(outlinePen)
			painter.SetBrush(fillBrush)
			painter.DrawEllipse2(
				m.X-markerDiameter/2, m.Y-markerDiameter/2,
				markerDiameter, markerDiameter,
			)
			painter.SetPenWithPen(textPen)
			painter.SetFont(font)
			rect := gc(qtlib.NewQRect4(
				m.X-markerDiameter/2, m.Y-markerDiameter/2,
				markerDiameter, markerDiameter,
			))
			painter.DrawText6(rect, int(qtlib.AlignCenter), fmt.Sprintf("%d", m.Number))
		}
		return nil
	})
}

func (s *screenshotter) Save(p Pixmap, path string) error {
	img, ok := p.(*capturedImage)
	if !ok {
		return errors.New("invalid Pixmap handle")
	}
	return runOnMainSync(func() error {
		if !img.pm.Save(path) {
			return fmt.Errorf("save %s failed", path)
		}
		return nil
	})
}
