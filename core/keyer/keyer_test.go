package keyer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core/mocked"
	"github.com/ftl/hellocontest/core/model"
)

func TestSend(t *testing.T) {
	keyerSettings := model.KeyerSettings{
		WPM:       30,
		SPMacros:  []string{"", "", "", ""},
		RunMacros: []string{"", "", "", ""},
	}
	values := func() model.KeyerValues {
		return model.KeyerValues{
			TheirCall: "DL0ZZZ",
			MyNumber:  model.QSONumber(56),
			MyReport:  model.RST("599"),
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
	view.On("SetLastTransmission", mock.Anything)
	cwClient := new(mocked.CWClient)
	cwClient.On("Speed", 30).Once()
	cwClient.On("Send", "DL1ABC DL0ZZZ t56 5nn ABC").Once()

	keyer := New(&testSettings{"DL1ABC"}, cwClient, keyerSettings, model.SearchPounce, nil)
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
	keyerSettings := model.KeyerSettings{
		WPM:       25,
		SPMacros:  []string{"sp0", "", "", ""},
		RunMacros: []string{"run0", "", "", ""},
	}
	cwClient := new(mocked.CWClient)
	k := New(&testSettings{"DL1ABC"}, cwClient, keyerSettings, model.SearchPounce, nil)

	// Simulate SO2V: VFO1=Run, VFO2=S&P
	k.WorkmodeChanged(model.VFO1, model.Run)
	k.WorkmodeChanged(model.VFO2, model.SearchPounce)
	k.FocusedVFOChanged(model.VFO1)
	assert.Equal(t, model.Run, k.workmode, "VFO1 focused → Run")

	// Switch to VFO2
	k.FocusedVFOChanged(model.VFO2)
	assert.Equal(t, model.SearchPounce, k.workmode, "VFO2 focused → S&P")

	// Switch back to VFO1 — this is the regression case.
	// WorkmodeChanged may arrive before or after FocusChanged.
	// Test both orderings:

	// Order A: FocusedVFOChanged first, then WorkmodeChanged
	k.FocusedVFOChanged(model.VFO1)
	k.WorkmodeChanged(model.VFO1, model.Run)
	assert.Equal(t, model.Run, k.workmode, "Order A: VFO1 focused → Run")

	// Reset to VFO2
	k.FocusedVFOChanged(model.VFO2)
	assert.Equal(t, model.SearchPounce, k.workmode)

	// Order B: WorkmodeChanged first, then FocusChanged
	k.WorkmodeChanged(model.VFO1, model.Run)
	k.FocusedVFOChanged(model.VFO1)
	assert.Equal(t, model.Run, k.workmode, "Order B: VFO1 focused → Run")
}

type testSettings struct {
	stationCallsign string
}

func (s *testSettings) Station() model.Station {
	return model.Station{
		Callsign: model.MustParseCallsign(s.stationCallsign),
	}
}

func (s *testSettings) Contest() model.Contest {
	return model.Contest{}
}

type vfoSwitcherSpy struct{ txVFOs []model.VFOID }

func (s *vfoSwitcherSpy) SetTXVFO(vfo model.VFOID) { s.txVFOs = append(s.txVFOs, vfo) }

type transmissionStartedSpy struct{ vfos []model.VFOID }

func (s *transmissionStartedSpy) TransmissionStarted(vfo model.VFOID) { s.vfos = append(s.vfos, vfo) }

func TestSend_SwitchesTXVFOToFocusedAndAnnounces(t *testing.T) {
	keyerSettings := model.KeyerSettings{
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
	view.On("SetLastTransmission", mock.Anything)
	cwClient := new(mocked.CWClient)
	cwClient.On("Speed", mock.Anything)
	cwClient.On("Send", mock.Anything)
	vfoSw := &vfoSwitcherSpy{}
	tx := &transmissionStartedSpy{}

	k := New(&testSettings{"DL1ABC"}, cwClient, keyerSettings, model.SearchPounce, nil)
	k.SetView(view)
	k.SetVFOSwitcher(vfoSw)
	k.Notify(tx)

	// An operator send switches the TX VFO to the focused VFO and announces it.
	k.FocusedVFOChanged(model.VFO2)
	k.SendMacro(0)
	assert.Equal(t, []model.VFOID{model.VFO2}, vfoSw.txVFOs, "operator send switches TX to the focused VFO")
	assert.Equal(t, []model.VFOID{model.VFO2}, tx.vfos, "operator send announces the transmission")

	// An explicit-VFO send (parrot) targets the given VFO and does NOT announce
	// (otherwise the parrot would stop on its own CQ).
	k.SendWithWorkmodeOnVFO(model.VFO1, model.Run, 0)
	assert.Equal(t, []model.VFOID{model.VFO2, model.VFO1}, vfoSw.txVFOs, "explicit send switches TX to the given VFO")
	assert.Equal(t, []model.VFOID{model.VFO2}, tx.vfos, "explicit send must not announce")
}

type macroSentCall struct {
	vfo      model.VFOID
	workmode model.Workmode
	index    int
}

type macroSentSpy struct{ calls []macroSentCall }

func (s *macroSentSpy) MacroSent(vfo model.VFOID, workmode model.Workmode, index int) {
	s.calls = append(s.calls, macroSentCall{vfo: vfo, workmode: workmode, index: index})
}

func TestSend_AnnouncesTheMacroThatWentOut(t *testing.T) {
	keyerSettings := model.KeyerSettings{
		WPM:       25,
		SPMacros:  []string{"sp0", "sp1", "", ""},
		RunMacros: []string{"run0", "", "", ""},
	}
	view := new(mocked.KeyerView)
	view.On("SetKeyerController", mock.Anything)
	view.On("ShowMessage", mock.Anything)
	view.On("SetSpeed", mock.Anything)
	view.On("SetLabel", mock.Anything, mock.Anything)
	view.On("SetPattern", mock.Anything, mock.Anything)
	view.On("SetPresetNames", mock.Anything)
	view.On("SetLastTransmission", mock.Anything)
	cwClient := new(mocked.CWClient)
	cwClient.On("Speed", mock.Anything)
	cwClient.On("Send", mock.Anything)
	sent := &macroSentSpy{}

	k := New(&testSettings{"DL1ABC"}, cwClient, keyerSettings, model.SearchPounce, nil)
	k.SetView(view)
	k.SetVFOSwitcher(&vfoSwitcherSpy{})
	k.Notify(sent)

	// the operator sends with the focused VFO and the workmode of that VFO
	k.FocusedVFOChanged(model.VFO2)
	k.WorkmodeChanged(model.VFO2, model.SearchPounce)
	k.SendMacro(1)
	require.Len(t, sent.calls, 1)
	assert.Equal(t, macroSentCall{vfo: model.VFO2, workmode: model.SearchPounce, index: 1}, sent.calls[0])

	// the parrot sends on the VFO that it names, in the running workmode
	k.SendWithWorkmodeOnVFO(model.VFO1, model.Run, model.CQMacroIndex)
	require.Len(t, sent.calls, 2)
	assert.Equal(t, macroSentCall{vfo: model.VFO1, workmode: model.Run, index: model.CQMacroIndex}, sent.calls[1])
}
