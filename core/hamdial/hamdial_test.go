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
