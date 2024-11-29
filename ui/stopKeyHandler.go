//go:build !fyne

package ui

import (
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const doubleStopThreshold = 250 * time.Millisecond

type StopKeyController interface {
	Stop()
	DoubleStop()
}

type stopKeyHandler struct {
	controller StopKeyController

	keyDown bool
	lastUp  time.Time
}

func setupStopKeyHandler(w *gtk.Widget) *stopKeyHandler {
	result := new(stopKeyHandler)

	w.Connect("key_press_event", result.onKeyPress)
	w.Connect("key_release_event", result.onKeyRelease)

	return result
}

func (h *stopKeyHandler) onKeyPress(_ interface{}, event *gdk.Event) bool {
	if h.controller == nil {
		return false
	}

	keyEvent := gdk.EventKeyNewFromEvent(event)
	switch keyEvent.KeyVal() {
	case gdk.KEY_Escape:
		if h.keyDown {
			return true
		}
		h.keyDown = true
		h.controller.Stop()
		return true
	default:
		return false
	}
}

func (h *stopKeyHandler) onKeyRelease(_ interface{}, event *gdk.Event) bool {
	if h.controller == nil {
		return false
	}

	keyEvent := gdk.EventKeyNewFromEvent(event)
	switch keyEvent.KeyVal() {
	case gdk.KEY_Escape:
		h.keyDown = false
		now := time.Now()
		duration := now.Sub(h.lastUp)
		if duration < doubleStopThreshold {
			h.controller.DoubleStop()
		}
		h.lastUp = now
		return true
	default:
		return false
	}
}

func (h *stopKeyHandler) SetStopKeyController(controller StopKeyController) {
	h.controller = controller
}
