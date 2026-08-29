package hamlib

import (
	"testing"

	"github.com/ftl/hamradio/bandplan"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestSetFrequencyCoalescesCommands(t *testing.T) {
	c := New("localhost:4532", bandplan.IARURegion1, "", "")
	c.loopRunning.Store(true)

	c.SetFrequency(core.VFO1, 3500000)
	c.SetFrequency(core.VFO1, 3500100)
	c.SetFrequency(core.VFO1, 3500200)

	assert.Len(t, c.do, 1, "only one frequency command must be queued")
	assert.Equal(t, core.Frequency(3500200), c.takePendingFrequency(core.VFO1), "the queued command must use the most recent frequency")
}

func TestSetFrequencyQueuesAgainAfterTheCommandWasTaken(t *testing.T) {
	c := New("localhost:4532", bandplan.IARURegion1, "", "")
	c.loopRunning.Store(true)

	c.SetFrequency(core.VFO1, 3500000)
	<-c.do
	c.takePendingFrequency(core.VFO1)
	c.SetFrequency(core.VFO1, 3500100)

	assert.Len(t, c.do, 1)
	assert.Equal(t, core.Frequency(3500100), c.takePendingFrequency(core.VFO1))
}

func TestSetFrequencyCoalescesPerVFO(t *testing.T) {
	c := New("localhost:4532", bandplan.IARURegion1, "", "")
	c.loopRunning.Store(true)

	c.SetFrequency(core.VFO1, 3500000)
	c.SetFrequency(core.VFO2, 7000000)

	assert.Len(t, c.do, 2)
}

func TestSetFrequencyDoesNotBlockTheQueueIfTheLoopIsNotRunning(t *testing.T) {
	c := New("localhost:4532", bandplan.IARURegion1, "", "")

	c.SetFrequency(core.VFO1, 3500000)
	c.loopRunning.Store(true)
	c.SetFrequency(core.VFO1, 3500100)

	assert.Len(t, c.do, 1, "the command must be queued once the loop is running")
}
