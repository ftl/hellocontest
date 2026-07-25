package keyer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/mocked"
)

func TestSend(t *testing.T) {
	keyerSettings := core.KeyerSettings{
		WPM:       30,
		SPMacros:  []string{"", "", "", ""},
		RunMacros: []string{"", "", "", ""},
	}
	values := func() core.KeyerValues {
		return core.KeyerValues{
			TheirCall: "DL0ZZZ",
			MyNumber:  core.QSONumber(56),
			MyReport:  core.RST("599"),
			MyXchange: "ABC",
		}
	}
	view := new(mocked.KeyerView)
	view.On("SetKeyerController", mock.Anything)
	view.On("ShowMessage", mock.Anything)
	view.On("SetSpeed", mock.Anything)
	view.On("SetLabel", mock.Anything, mock.Anything)
	view.On("SetPattern", mock.Anything, mock.Anything)
	view.On("SetPresetNames", mock.Anything)
	cwClient := new(mocked.CWClient)
	cwClient.On("Speed", 30).Once()
	cwClient.On("Send", "DL1ABC DL0ZZZ t56 5nn ABC").Once()

	keyer := New(&testSettings{"DL1ABC"}, cwClient, keyerSettings, core.SearchPounce, nil)
	keyer.SetView(view)
	keyer.SetValues(values)
	keyer.EnterPattern(0, "{{.MyCall}} {{.TheirCall}} {{.MyNumber}} {{.MyReport}} {{.MyXchange}}")

	keyer.SendMacro(0)

	cwClient.AssertExpectations(t)
}

func TestCutDefault(t *testing.T) {
	assert.Equal(t, "t12345678n", cutDefault("0123456789"))
}

func TestCutOnly(t *testing.T) {
	assert.Equal(t, "tauv4e6gdn", cutOnly(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, "0123456789"))
}

func TestPad(t *testing.T) {
	assert.Equal(t, "0123456789", pad(10, "0123456789"))
	assert.Equal(t, "0123456789", pad(5, "0123456789"))
	assert.Equal(t, "0000000000", pad(10, ""))
	assert.Equal(t, "00000", pad(5, ""))
	assert.Equal(t, "", pad(0, ""))
}

func TestWorkmode_FocusSwitchOrderIndependence(t *testing.T) {
	keyerSettings := core.KeyerSettings{
		WPM:       25,
		SPMacros:  []string{"sp0", "", "", ""},
		RunMacros: []string{"run0", "", "", ""},
	}
	cwClient := new(mocked.CWClient)
	k := New(&testSettings{"DL1ABC"}, cwClient, keyerSettings, core.SearchPounce, nil)

	// Simulate SO2V: VFO1=Run, VFO2=S&P
	k.WorkmodeChanged(core.VFO1, core.Run)
	k.WorkmodeChanged(core.VFO2, core.SearchPounce)
	k.FocusChanged(core.VFO1)
	assert.Equal(t, core.Run, k.workmode, "VFO1 focused → Run")

	// Switch to VFO2
	k.FocusChanged(core.VFO2)
	assert.Equal(t, core.SearchPounce, k.workmode, "VFO2 focused → S&P")

	// Switch back to VFO1 — this is the regression case.
	// WorkmodeChanged may arrive before or after FocusChanged.
	// Test both orderings:

	// Order A: FocusChanged first, then WorkmodeChanged
	k.FocusChanged(core.VFO1)
	k.WorkmodeChanged(core.VFO1, core.Run)
	assert.Equal(t, core.Run, k.workmode, "Order A: VFO1 focused → Run")

	// Reset to VFO2
	k.FocusChanged(core.VFO2)
	assert.Equal(t, core.SearchPounce, k.workmode)

	// Order B: WorkmodeChanged first, then FocusChanged
	k.WorkmodeChanged(core.VFO1, core.Run)
	k.FocusChanged(core.VFO1)
	assert.Equal(t, core.Run, k.workmode, "Order B: VFO1 focused → Run")
}

type testSettings struct {
	stationCallsign string
}

func (s *testSettings) Station() core.Station {
	return core.Station{
		Callsign: core.MustParseCallsign(s.stationCallsign),
	}
}

func (s *testSettings) Contest() core.Contest {
	return core.Contest{}
}

type vfoSwitcherSpy struct{ txVFOs []core.VFOID }

func (s *vfoSwitcherSpy) SetTXVFO(vfo core.VFOID) { s.txVFOs = append(s.txVFOs, vfo) }

type transmissionStartedSpy struct{ vfos []core.VFOID }

func (s *transmissionStartedSpy) TransmissionStarted(vfo core.VFOID) { s.vfos = append(s.vfos, vfo) }

func TestSend_SwitchesTXVFOToFocusedAndAnnounces(t *testing.T) {
	keyerSettings := core.KeyerSettings{
		WPM:       25,
		SPMacros:  []string{"sp0", "", "", ""},
		RunMacros: []string{"run0", "", "", ""},
	}
	view := new(mocked.KeyerView)
	view.On("SetKeyerController", mock.Anything)
	view.On("ShowMessage", mock.Anything)
	view.On("SetSpeed", mock.Anything)
	view.On("SetLabel", mock.Anything, mock.Anything)
	view.On("SetPattern", mock.Anything, mock.Anything)
	view.On("SetPresetNames", mock.Anything)
	cwClient := new(mocked.CWClient)
	cwClient.On("Speed", mock.Anything)
	cwClient.On("Send", mock.Anything)
	vfoSw := &vfoSwitcherSpy{}
	tx := &transmissionStartedSpy{}

	k := New(&testSettings{"DL1ABC"}, cwClient, keyerSettings, core.SearchPounce, nil)
	k.SetView(view)
	k.SetVFOSwitcher(vfoSw)
	k.Notify(tx)

	// An operator send switches the TX VFO to the focused VFO and announces it.
	k.FocusChanged(core.VFO2)
	k.SendMacro(0)
	assert.Equal(t, []core.VFOID{core.VFO2}, vfoSw.txVFOs, "operator send switches TX to the focused VFO")
	assert.Equal(t, []core.VFOID{core.VFO2}, tx.vfos, "operator send announces the transmission")

	// An explicit-VFO send (parrot) targets the given VFO and does NOT announce
	// (otherwise the parrot would stop on its own CQ).
	k.SendWithWorkmodeOnVFO(core.VFO1, core.Run, 0)
	assert.Equal(t, []core.VFOID{core.VFO2, core.VFO1}, vfoSw.txVFOs, "explicit send switches TX to the given VFO")
	assert.Equal(t, []core.VFOID{core.VFO2}, tx.vfos, "explicit send must not announce")
}
