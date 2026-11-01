package hamlib

import (
	"testing"

	"github.com/ftl/hamradio/bandplan"
	hl "github.com/ftl/hl-go"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestSetFrequencyCoalescesCommands(t *testing.T) {
	c := New("localhost:4532", bandplan.IARURegion1, "", "", nil)
	c.loopRunning.Store(true)

	c.SetFrequency(core.VFO1, 3500000)
	c.SetFrequency(core.VFO1, 3500100)
	c.SetFrequency(core.VFO1, 3500200)

	assert.Len(t, c.do, 1, "only one frequency command must be queued")
	assert.Equal(t, core.Frequency(3500200), c.takePendingFrequency(core.VFO1), "the queued command must use the most recent frequency")
}

func TestSetFrequencyQueuesAgainAfterTheCommandWasTaken(t *testing.T) {
	c := New("localhost:4532", bandplan.IARURegion1, "", "", nil)
	c.loopRunning.Store(true)

	c.SetFrequency(core.VFO1, 3500000)
	<-c.do
	c.takePendingFrequency(core.VFO1)
	c.SetFrequency(core.VFO1, 3500100)

	assert.Len(t, c.do, 1)
	assert.Equal(t, core.Frequency(3500100), c.takePendingFrequency(core.VFO1))
}

func TestSetFrequencyCoalescesPerVFO(t *testing.T) {
	c := New("localhost:4532", bandplan.IARURegion1, "", "", nil)
	c.loopRunning.Store(true)

	c.SetFrequency(core.VFO1, 3500000)
	c.SetFrequency(core.VFO2, 7000000)

	assert.Len(t, c.do, 2)
}

func TestSetFrequencyDoesNotBlockTheQueueIfTheLoopIsNotRunning(t *testing.T) {
	c := New("localhost:4532", bandplan.IARURegion1, "", "", nil)

	c.SetFrequency(core.VFO1, 3500000)
	c.loopRunning.Store(true)
	c.SetFrequency(core.VFO1, 3500100)

	assert.Len(t, c.do, 1, "the command must be queued once the loop is running")
}

func TestEmitUsesTheAsyncRunner(t *testing.T) {
	testCases := []struct {
		desc string
		emit func(c *Client)
	}{
		{
			// the client must know both VFOs, otherwise it ignores the change
			desc: "current VFO changed",
			emit: func(c *Client) { c.emitCurrentVFOChanged(hl.VFOA, hl.VFOB) },
		},
		{
			desc: "TX VFO changed",
			emit: func(c *Client) { c.emitTXVFOChanged(core.VFO1) },
		},
		{
			desc: "change notifications",
			emit: func(c *Client) {
				c.emitChangeNotifications(core.VFO1, vfoState{frequency: 3500000}, vfoState{frequency: 3500100})
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			runCount := 0
			c := New("localhost:4532", bandplan.IARURegion1, string(hl.VFOA), string(hl.VFOB), func(f func()) {
				runCount++
				f()
			})

			tc.emit(c)

			assert.Equal(t, 1, runCount, "the listeners must be notified through the async runner")
		})
	}
}
