package keyer

import (
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/mocked"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSend(t *testing.T) {
	keyerSettings := core.Keyer{
		SPMacros:  []string{"", "", "", ""},
		RunMacros: []string{"", "", "", ""},
		WPM:       25,
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
	view.On("SetPattern", mock.Anything, mock.Anything)
	view.On("SetPresetNames", mock.Anything)
	cwClient := new(mocked.CWClient)
	cwClient.On("Send", "DL1ABC DL0ZZZ t56 5nn ABC").Once()
	cwClient.On("IsConnected").Return(true)

	keyer := New(&testSettings{"DL1ABC"}, cwClient, keyerSettings, core.SearchPounce, nil)
	keyer.SetView(view)
	keyer.SetValues(values)
	keyer.EnterPattern(0, "{{.MyCall}} {{.TheirCall}} {{.MyNumber}} {{.MyReport}} {{.MyXchange}}")

	keyer.Send(0)

	cwClient.AssertExpectations(t)
}

func TestSoftcut(t *testing.T) {
	assert.Equal(t, "t12345678n", softcut("0123456789"))
}

func TestCut(t *testing.T) {
	assert.Equal(t, "tauv4e6gdn", cut("0123456789"))
}

type testSettings struct {
	stationCallsign string
}

func (s *testSettings) Station() core.Station {
	return core.Station{
		Callsign: callsign.MustParse(s.stationCallsign),
	}
}

func (s *testSettings) Contest() core.Contest {
	return core.Contest{}
}
