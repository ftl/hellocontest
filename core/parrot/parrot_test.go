package parrot

import (
	"testing"
	"time"

	"github.com/ftl/hellocontest/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type keyerSpy struct {
	sentVFO      model.VFOID
	sentWorkmode model.Workmode
	sentIndex    int
	sendCalled   bool
	stopped      bool
}

func (k *keyerSpy) SendWithWorkmodeOnVFO(vfo model.VFOID, workmode model.Workmode, index int) {
	k.sentVFO = vfo
	k.sentWorkmode = workmode
	k.sentIndex = index
	k.sendCalled = true
}

func (k *keyerSpy) Stop() {
	k.stopped = true
}

func (k *keyerSpy) reset() {
	k.sentVFO = 0
	k.sentWorkmode = model.UnknownWorkmode
	k.sentIndex = -1
	k.sendCalled = false
	k.stopped = false
}

type workmodeSpy struct {
	workmode model.Workmode
}

func (w *workmodeSpy) SetWorkmode(workmode model.Workmode) {
	w.workmode = workmode
}

func newTestParrot() (*Parrot, *keyerSpy) {
	keyer := &keyerSpy{}
	wm := &workmodeSpy{}
	p := New(wm, keyer, func(f func()) { f() })
	p.SetInterval(50 * time.Millisecond)
	return p, keyer
}

func TestParrot_CQ_UseRunWorkmode(t *testing.T) {
	p, keyer := newTestParrot()

	p.Start()
	defer p.Stop()

	// Wait for first CQ tick (tickInterval = 1s, first CQ after ~1 tick)
	require.Eventually(t, func() bool {
		return keyer.sendCalled
	}, 3*time.Second, 50*time.Millisecond, "Parrot must send CQ")

	assert.Equal(t, model.Run, keyer.sentWorkmode,
		"Parrot must always use Run workmode for CQ")
	assert.Equal(t, CQMessageIndex, keyer.sentIndex,
		"Parrot must send CQ macro index 0")
	assert.Equal(t, model.VFO1, keyer.sentVFO,
		"Parrot must transmit CQ on VFO1")
}

func TestParrot_TransmissionStarted_VFO2_StopsParrot(t *testing.T) {
	p, keyer := newTestParrot()

	p.Start()

	// VFO2 transmission must interrupt parrot
	keyer.reset()
	p.TransmissionStarted(model.VFO2)

	assert.True(t, keyer.stopped, "VFO2 transmission must stop keyer")
}

func TestParrot_TransmissionStarted_VFO1_StopsKeyer(t *testing.T) {
	p, keyer := newTestParrot()

	p.Start()

	// Any manual transmission stops the parrot, incl. VFO1 (e.g. pressing '?')
	keyer.reset()
	p.TransmissionStarted(model.VFO1)

	assert.True(t, keyer.stopped, "VFO1 transmission must stop keyer")
}

func TestParrot_CallsignEntered_VFO1_StopsKeyer(t *testing.T) {
	p, keyer := newTestParrot()

	p.Start()

	keyer.reset()
	p.CallsignEntered(model.VFO1, "DL1ABC")

	assert.True(t, keyer.stopped, "VFO1 callsign must stop keyer")
}

func TestParrot_CallsignEntered_VFO2_Ignored(t *testing.T) {
	p, keyer := newTestParrot()

	p.Start()

	keyer.reset()
	p.CallsignEntered(model.VFO2, "DL1ABC")

	assert.False(t, keyer.stopped, "VFO2 callsign must not stop keyer")
}

func TestParrot_WorkmodeChanged_VFO1_NonRun_StopsKeyer(t *testing.T) {
	p, keyer := newTestParrot()

	p.Start()

	keyer.reset()
	p.WorkmodeChanged(model.VFO1, model.SearchPounce)

	assert.True(t, keyer.stopped, "VFO1 workmode S&P must stop keyer")
}

func TestParrot_WorkmodeChanged_VFO2_Ignored(t *testing.T) {
	p, keyer := newTestParrot()

	p.Start()

	keyer.reset()
	p.WorkmodeChanged(model.VFO2, model.SearchPounce)

	assert.False(t, keyer.stopped, "VFO2 workmode change must not stop keyer")
}
