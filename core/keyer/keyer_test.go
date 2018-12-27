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
	myCall, _ := callsign.Parse("DL1ABC")
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
	cwClient := new(mocked.CWClient)
	cwClient.On("Send", "DL1ABC DL0ZZZ t56 5nn ABC").Once()
	cwClient.On("IsConnected").Return(true)

	keyer := NewController(cwClient, myCall, values)
	keyer.SetView(view)
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
