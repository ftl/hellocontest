package hamdial

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/ftl/hamdial"
	"github.com/ftl/hellocontest/core"
)

const (
	longPressDuration = 500 * time.Millisecond

	// tuning acceleration: the frequency delta per detent grows with the turn rate
	baseTuningStep      = core.Frequency(10)     // delta per detent when turning slowly
	maxTuningStep       = core.Frequency(500)    // delta per detent when turning at full speed
	tuningStepIncrement = core.Frequency(10)     // the delta is rounded to a multiple of this value
	slowTurnInterval    = 200 * time.Millisecond // detents further apart than this are considered slow
	tuningAcceleration  = 2.0                    // exponent of the acceleration curve, > 1 means faster than linear
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

	lastTurn          time.Time
	lastTurnDirection hamdial.Direction
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
		c.actions.ShiftFrequency(c.turnDelta(direction, time.Now()))
	}
}

func (c *Controller) Disconnected() {
	log.Print("dial disconnected")
	c.dial = nil
	c.stopDial = nil
	c.button1Pressed = false
	c.emitActiveChanged()
}

// turnDelta returns the frequency delta for one detent, accelerated by the turn rate:
// the shorter the interval since the preceding detent, the larger the delta.
func (c *Controller) turnDelta(direction hamdial.Direction, now time.Time) core.Frequency {
	interval := now.Sub(c.lastTurn)
	sameDirection := direction == c.lastTurnDirection
	c.lastTurn = now
	c.lastTurnDirection = direction

	step := baseTuningStep
	if sameDirection && interval > 0 && interval < slowTurnInterval {
		factor := math.Pow(float64(slowTurnInterval)/float64(interval), tuningAcceleration)
		step = core.Frequency(math.Round(float64(baseTuningStep)*factor/float64(tuningStepIncrement))) * tuningStepIncrement
		step = min(step, maxTuningStep)
	}

	return core.Frequency(direction) * step
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
