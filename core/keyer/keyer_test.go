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
	keyerSettings := core.KeyerSettings{
		WPM:       25,
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
	cwClient.On("Send", "DL1ABC DL0ZZZ t56 5nn ABC").Once()

	keyer := New(&testSettings{"DL1ABC"}, cwClient, keyerSettings, core.SearchPounce, nil)
	keyer.SetView(view)
	keyer.SetValues(values)
	keyer.EnterPattern(0, "{{.MyCall}} {{.TheirCall}} {{.MyNumber}} {{.MyReport}} {{.MyXchange}}")

	keyer.Send(0)

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
