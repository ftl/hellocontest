package hamdial

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ftl/hamdial"
	"github.com/ftl/hellocontest/core"
)

const (
	longPressDuration = 500 * time.Millisecond
)

type DialActions interface {
	GotoNextSpotUp()               // 1
	ActivateBestMatchOnFrequency() // 2
	GotoNextSpotDown()             // 3
	MarkInBandmap()                // 4
	FocusVFO1()                    // 5
	// 6
	FocusVFO2() // 7

	ShiftFrequency(core.Frequency)
}

// ActiveChangedListener is notified when the connection state of the dial changes.
type ActiveChangedListener interface {
	DialActiveChanged(active bool)
}

type ActiveChangedListenerFunc func(active bool)

func (f ActiveChangedListenerFunc) DialActiveChanged(active bool) {
	f(active)
}

type Controller struct {
	actions     DialActions
	asyncRunner core.AsyncRunner
	listeners   []any

	dial     *hamdial.Dial
	stopDial context.CancelFunc

	detent bool

	button1Pressed      bool
	button1PressedSince time.Time
}

func New(actions DialActions, asyncRunner core.AsyncRunner) *Controller {
	return &Controller{
		actions:     actions,
		asyncRunner: asyncRunner,
	}
}

func (c *Controller) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Controller) emitActiveChanged() {
	active := c.Active()
	c.asyncRunner(func() {
		for _, listener := range c.listeners {
			if activeChangedListener, ok := listener.(ActiveChangedListener); ok {
				activeChangedListener.DialActiveChanged(active)
			}
		}
	})
}

func (c *Controller) Active() bool {
	return c.dial != nil
}

func (c *Controller) SetActive(active bool) error {
	if active == c.Active() {
		return nil
	}

	var err error
	if active {
		err = c.enable()
	} else {
		err = c.disable()
	}
	c.emitActiveChanged()
	return err
}

func (c *Controller) enable() error {
	log.Printf("enable HamDial")
	dial, err := hamdial.OpenDial("")
	if err != nil {
		log.Printf("cannot open HamDial: %v", err)
		return fmt.Errorf("Cannot open HamDial")
	}
	dial.Notify(c)

	dial.SetHapticFeedback(false)
	c.detent = false

	ctx, cancel := context.WithCancel(context.Background())

	c.dial = dial
	c.stopDial = cancel

	go func() {
		c.dial.Run(ctx)
	}()

	return nil
}

func (c *Controller) disable() error {
	log.Printf("disable HamDial")
	c.stopDial()
	err := c.dial.Close()
	c.dial = nil
	c.stopDial = nil
	if err != nil {
		log.Printf("cannot close HamDial: %v", err)
	}
	return nil
}

func (c *Controller) ButtonPressed(button hamdial.Button) {
	switch button {
	case hamdial.Button1:
		c.button1Pressed = true
		c.button1PressedSince = time.Now()
	case hamdial.Button2:
		c.actions.ActivateBestMatchOnFrequency()
	case hamdial.Button3:
		c.actions.GotoNextSpotUp()
	case hamdial.Button4:
		c.actions.MarkInBandmap()
	case hamdial.Button5:
		c.actions.FocusVFO1()
	case hamdial.Button6:
	// TODO: observe in VFO2
	case hamdial.Button7:
		c.actions.FocusVFO2()
	}
}

func (c *Controller) ButtonReleased(button hamdial.Button) {
	switch button {
	case hamdial.Button1:
		if !c.spotMode() {
			c.actions.GotoNextSpotDown()
		}
		c.button1Pressed = false
		c.updateDetent()
	default:
		// no actions when other buttons are released
	}
}

func (c *Controller) DialTurned(direction hamdial.Direction) {
	c.updateDetent()
	if c.spotMode() {
		switch direction {
		case hamdial.Clockwise:
			c.actions.GotoNextSpotUp()
		case hamdial.CounterClockwise:
			c.actions.GotoNextSpotDown()
		}
	} else {
		delta := core.Frequency(direction) * 50 // TODO: calculate the delta based on the turn rate
		c.actions.ShiftFrequency(delta)
	}
}

func (c *Controller) Disconnected() {
	log.Print("dial disconnected")
	c.dial = nil
	c.stopDial = nil
	c.button1Pressed = false
	c.emitActiveChanged()
}

func (c *Controller) spotMode() bool {
	return c.button1Pressed && time.Since(c.button1PressedSince) >= longPressDuration
}

func (c *Controller) updateDetent() {
	if c.dial == nil {
		return
	}
	spotMode := c.spotMode()
	if c.detent == spotMode {
		return
	}
	c.detent = spotMode
	c.dial.SetHapticFeedback(c.detent)
}
