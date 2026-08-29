package hamdial

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
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

	shiftLock      sync.Mutex
	pendingDelta   core.Frequency
	shiftScheduled bool
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

// runAction runs an action on the main thread. All events of the dial are
// handled on the dial's own goroutine, but the actions modify the state of the
// application and must therefore run on the main thread.
func (c *Controller) runAction(action func()) {
	c.asyncRunner(action)
}

func (c *Controller) emitActiveChanged() {
	active := c.Active()
	c.asyncRunner(func() {
		c.notifyActiveChanged(active)
	})
}

func (c *Controller) notifyActiveChanged(active bool) {
	for _, listener := range c.listeners {
		if activeChangedListener, ok := listener.(ActiveChangedListener); ok {
			activeChangedListener.DialActiveChanged(active)
		}
	}
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
		c.runAction(c.actions.ActivateBestMatchOnFrequency)
	case hamdial.Button3:
		c.runAction(c.actions.GotoNextSpotUp)
	case hamdial.Button4:
		c.runAction(c.actions.MarkInBandmap)
	case hamdial.Button5:
		c.runAction(c.actions.FocusVFO1)
	case hamdial.Button6:
	// TODO: observe in VFO2
	case hamdial.Button7:
		c.runAction(c.actions.FocusVFO2)
	}
}

func (c *Controller) ButtonReleased(button hamdial.Button) {
	switch button {
	case hamdial.Button1:
		if !c.spotMode() {
			c.runAction(c.actions.GotoNextSpotDown)
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
			c.runAction(c.actions.GotoNextSpotUp)
		case hamdial.CounterClockwise:
			c.runAction(c.actions.GotoNextSpotDown)
		}
	} else {
		// the delta is calculated here, on the dial's goroutine, because it depends on the time between the turn events
		c.shiftFrequency(c.turnDelta(direction, time.Now()))
	}
}

// shiftFrequency accumulates the given delta and schedules its application on
// the main thread, if that is not already pending.
func (c *Controller) shiftFrequency(delta core.Frequency) {
	c.shiftLock.Lock()
	c.pendingDelta += delta
	alreadyScheduled := c.shiftScheduled
	c.shiftScheduled = true
	c.shiftLock.Unlock()

	if alreadyScheduled {
		return
	}
	c.runAction(c.applyPendingShift)
}

// applyPendingShift applies all deltas that accumulated since the last run as
// one single shift. Turning the dial faster than the rig can follow would
// otherwise pile up shifts that keep changing the frequency long after the dial
// stopped.
func (c *Controller) applyPendingShift() {
	c.shiftLock.Lock()
	delta := c.pendingDelta
	c.pendingDelta = 0
	c.shiftScheduled = false
	c.shiftLock.Unlock()

	if delta != 0 {
		c.actions.ShiftFrequency(delta)
	}
}

func (c *Controller) Disconnected() {
	log.Print("dial disconnected")
	// the state is modified on the main thread, where enable and disable modify it, too
	c.runAction(func() {
		c.dial = nil
		c.stopDial = nil
		c.button1Pressed = false
		c.notifyActiveChanged(c.Active())
	})
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
	dial := c.dial
	if dial == nil {
		return
	}
	spotMode := c.spotMode()
	if c.detent == spotMode {
		return
	}
	c.detent = spotMode
	dial.SetHapticFeedback(c.detent)
}
