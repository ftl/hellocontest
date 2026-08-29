package hamdial

import (
	"testing"
	"time"

	"github.com/ftl/hamdial"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
)

func TestTurnDelta(t *testing.T) {
	now := time.Now()
	tt := []struct {
		desc      string
		interval  time.Duration
		direction hamdial.Direction
		expected  core.Frequency
	}{
		{desc: "first detent", interval: 0, direction: hamdial.Clockwise, expected: baseTuningStep},
		{desc: "slow turn", interval: slowTurnInterval, direction: hamdial.Clockwise, expected: baseTuningStep},
		{desc: "very slow turn", interval: 2 * slowTurnInterval, direction: hamdial.Clockwise, expected: baseTuningStep},
		{desc: "half interval", interval: slowTurnInterval / 2, direction: hamdial.Clockwise, expected: 40},
		{desc: "fifth interval", interval: slowTurnInterval / 5, direction: hamdial.Clockwise, expected: 250},
		{desc: "very fast turn", interval: time.Millisecond, direction: hamdial.Clockwise, expected: maxTuningStep},
		{desc: "counter clockwise", interval: slowTurnInterval / 2, direction: hamdial.CounterClockwise, expected: -40},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			controller := &Controller{lastTurn: now.Add(-tc.interval), lastTurnDirection: tc.direction}
			assert.Equal(t, tc.expected, controller.turnDelta(tc.direction, now))
		})
	}
}

func TestTurnDeltaResetsOnDirectionChange(t *testing.T) {
	now := time.Now()
	controller := &Controller{lastTurn: now.Add(-slowTurnInterval / 5), lastTurnDirection: hamdial.Clockwise}

	assert.Equal(t, -baseTuningStep, controller.turnDelta(hamdial.CounterClockwise, now))
}

func TestTurnDeltaGrowsFasterThanLinear(t *testing.T) {
	now := time.Now()
	deltaAt := func(interval time.Duration) core.Frequency {
		controller := &Controller{lastTurn: now.Add(-interval), lastTurnDirection: hamdial.Clockwise}
		return controller.turnDelta(hamdial.Clockwise, now)
	}

	// halving the interval must more than double the delta
	assert.Greater(t, deltaAt(slowTurnInterval/4), 2*deltaAt(slowTurnInterval/2))
}

// queuedRunner collects the functions to run on the main thread, so that a busy
// main thread can be simulated.
type queuedRunner struct {
	queue []func()
}

func (r *queuedRunner) Run(f func()) {
	r.queue = append(r.queue, f)
}

func (r *queuedRunner) RunAll() {
	queue := r.queue
	r.queue = nil
	for _, f := range queue {
		f()
	}
}

type shiftRecorder struct {
	DialActions
	shifts []core.Frequency
}

func (a *shiftRecorder) ShiftFrequency(delta core.Frequency) {
	a.shifts = append(a.shifts, delta)
}

func TestDialTurnedCoalescesShiftsWhileTheMainThreadIsBusy(t *testing.T) {
	actions := new(shiftRecorder)
	runner := new(queuedRunner)
	controller := New(actions, runner.Run)

	for range 5 {
		controller.shiftFrequency(100)
	}
	assert.Empty(t, actions.shifts, "nothing must happen before the main thread is free")

	runner.RunAll()
	assert.Len(t, actions.shifts, 1, "all turns must be applied as one single shift")
	assert.Equal(t, core.Frequency(500), actions.shifts[0])

	runner.RunAll()
	assert.Len(t, actions.shifts, 1, "the accumulated delta must not be applied twice")
}

func TestDialTurnedSchedulesAgainAfterTheShiftWasApplied(t *testing.T) {
	actions := new(shiftRecorder)
	runner := new(queuedRunner)
	controller := New(actions, runner.Run)

	controller.shiftFrequency(100)
	runner.RunAll()
	controller.shiftFrequency(100)
	runner.RunAll()

	assert.Len(t, actions.shifts, 2)
}
