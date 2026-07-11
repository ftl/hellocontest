package parrot

import (
	"testing"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type keyerSpy struct {
	sentWorkmode core.Workmode
	sentIndex    int
	sendCalled   bool
	stopped      bool
}

func (k *keyerSpy) SendWithWorkmode(workmode core.Workmode, index int) {
	k.sentWorkmode = workmode
	k.sentIndex = index
	k.sendCalled = true
}

func (k *keyerSpy) Stop() {
	k.stopped = true
}

func (k *keyerSpy) reset() {
	k.sentWorkmode = core.UnknownWorkmode
	k.sentIndex = -1
	k.sendCalled = false
	k.stopped = false
}

type workmodeSpy struct {
	workmode core.Workmode
}

func (w *workmodeSpy) SetWorkmode(workmode core.Workmode) {
	w.workmode = workmode
}

type vfoSwitcherSpy struct {
	txVFO    core.VFOID
	txCalled bool
}

func (v *vfoSwitcherSpy) SetTXVFO(vfo core.VFOID) {
	v.txVFO = vfo
	v.txCalled = true
}

func (v *vfoSwitcherSpy) reset() {
	v.txVFO = 0
	v.txCalled = false
}

func newTestParrot() (*Parrot, *keyerSpy, *vfoSwitcherSpy) {
	keyer := &keyerSpy{}
	wm := &workmodeSpy{}
	vfoSw := &vfoSwitcherSpy{}
	p := New(wm, keyer, vfoSw, func(f func()) { f() })
	p.SetInterval(50 * time.Millisecond)
	return p, keyer, vfoSw
}

func TestParrot_CQ_UseRunWorkmode(t *testing.T) {
	p, keyer, vfoSw := newTestParrot()

	p.Start()
	defer p.Stop()

	// Wait for first CQ tick (tickInterval = 1s, first CQ after ~1 tick)
	require.Eventually(t, func() bool {
		return keyer.sendCalled
	}, 3*time.Second, 50*time.Millisecond, "Parrot must send CQ")

	assert.Equal(t, core.Run, keyer.sentWorkmode,
		"Parrot must always use Run workmode for CQ")
	assert.Equal(t, CQMessageIndex, keyer.sentIndex,
		"Parrot must send CQ macro index 0")
	assert.True(t, vfoSw.txCalled, "Parrot must call SetTXVFO")
	assert.Equal(t, core.VFO1, vfoSw.txVFO,
		"Parrot must set TX VFO to VFO1")
}

func TestParrot_TransmissionStarted_VFO2_StopsParrot(t *testing.T) {
	p, keyer, _ := newTestParrot()

	p.Start()

	// VFO2 transmission must interrupt parrot
	keyer.reset()
	p.TransmissionStarted(core.VFO2)

	assert.True(t, keyer.stopped, "VFO2 transmission must stop keyer")
}

func TestParrot_TransmissionStarted_VFO1_StopsKeyer(t *testing.T) {
	p, keyer, _ := newTestParrot()

	p.Start()

	// Any manual transmission stops the parrot, incl. VFO1 (e.g. pressing '?')
	keyer.reset()
	p.TransmissionStarted(core.VFO1)

	assert.True(t, keyer.stopped, "VFO1 transmission must stop keyer")
}

func TestParrot_CallsignEntered_VFO1_StopsKeyer(t *testing.T) {
	p, keyer, _ := newTestParrot()

	p.Start()

	keyer.reset()
	p.CallsignEntered(core.VFO1, "DL1ABC")

	assert.True(t, keyer.stopped, "VFO1 callsign must stop keyer")
}

func TestParrot_CallsignEntered_VFO2_Ignored(t *testing.T) {
	p, keyer, _ := newTestParrot()

	p.Start()

	keyer.reset()
	p.CallsignEntered(core.VFO2, "DL1ABC")

	assert.False(t, keyer.stopped, "VFO2 callsign must not stop keyer")
}

func TestParrot_WorkmodeChanged_VFO1_NonRun_StopsKeyer(t *testing.T) {
	p, keyer, _ := newTestParrot()

	p.Start()

	keyer.reset()
	p.WorkmodeChanged(core.VFO1, core.SearchPounce)

	assert.True(t, keyer.stopped, "VFO1 workmode S&P must stop keyer")
}

func TestParrot_WorkmodeChanged_VFO2_Ignored(t *testing.T) {
	p, keyer, _ := newTestParrot()

	p.Start()

	keyer.reset()
	p.WorkmodeChanged(core.VFO2, core.SearchPounce)

	assert.False(t, keyer.stopped, "VFO2 workmode change must not stop keyer")
}
