package ui

import (
	"time"
	"unsafe"

	qtlib "github.com/mappu/miqt/qt6"
)

const doubleStopThreshold = 250 * time.Millisecond

type StopKeyController interface {
	Stop()
	DoubleStop()
}

type stopKeyHandler struct {
	controller StopKeyController
	keyDown    bool
	lastUp     time.Time
}

func newStopKeyHandler(widget *qtlib.QWidget) *stopKeyHandler {
	h := &stopKeyHandler{
		keyDown: false,
		lastUp:  time.Now(),
	}

	// Create a directly-constructed object to use as event filter
	filterObj := qtlib.NewQObject()
	filterObj.OnEventFilter(func(super func(watched *qtlib.QObject, event *qtlib.QEvent) bool, watched *qtlib.QObject, event *qtlib.QEvent) bool {
		eventType := event.Type()

		// QEvent::KeyPress = 6, QEvent::KeyRelease = 7
		if eventType != 6 && eventType != 7 {
			return super(watched, event)
		}

		// Cast the QEvent to QKeyEvent to access key information
		// This works because the actual runtime object is a QKeyEvent subclass
		keyEvent := (*qtlib.QKeyEvent)(unsafe.Pointer(event))

		if eventType == 6 { // KeyPress
			if keyEvent.Key() == int(qtlib.Key_Escape) {
				if !h.onKeyPress() {
					return super(watched, event)
				}
				return true
			}
		} else if eventType == 7 { // KeyRelease
			if keyEvent.Key() == int(qtlib.Key_Escape) {
				if !h.onKeyRelease() {
					return super(watched, event)
				}
				return true
			}
		}

		return super(watched, event)
	})

	widget.InstallEventFilter(filterObj)

	return h
}

func (h *stopKeyHandler) onKeyPress() bool {
	if h.controller == nil {
		return false
	}

	if h.keyDown {
		return true
	}
	h.keyDown = true
	h.controller.Stop()
	return true
}

func (h *stopKeyHandler) onKeyRelease() bool {
	if h.controller == nil {
		return false
	}

	h.keyDown = false
	now := time.Now()
	duration := now.Sub(h.lastUp)
	if duration < doubleStopThreshold {
		h.controller.DoubleStop()
	}
	h.lastUp = now
	return true
}

func (h *stopKeyHandler) SetStopKeyController(controller StopKeyController) {
	h.controller = controller
}
