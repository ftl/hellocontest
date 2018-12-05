package keyer

import (
	"fmt"
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/mocked"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTemplateHandling(t *testing.T) {
	patterns := []string{"{{.MyCall}}", "{{.TheirCall}}", "{{.MyNumber}}", "{{.MyReport}}"}
	expected := []string{"DL1ABC", "DL0ZZZ", "456", "123"}
	myCall, _ := callsign.Parse("DL1ABC")
	values := func() core.KeyerValues {
		return core.KeyerValues{
			MyCall:    myCall,
			TheirCall: "DL0ZZZ",
			MyNumber:  core.QSONumber(456),
			MyReport:  core.RST("123"),
		}
	}

	keyer, err := New(patterns, new(mocked.CWClient), values)
	require.NoError(t, err)

	for i, pattern := range patterns {
		assert.Equal(t, pattern, keyer.GetTemplate(i))
		actual, err := keyer.GetText(i)
		require.NoError(t, err)
		assert.Equal(t, expected[i], actual)

		keyer.SetTemplate(i, fmt.Sprintf("%s %d", patterns[i], i))
		actual, err = keyer.GetText(i)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%s %d", expected[i], i), actual)
	}
}

func TestSend(t *testing.T) {
	patterns := []string{"{{.MyCall}} {{.TheirCall}} {{.MyNumber}} {{.MyReport}}"}
	myCall, _ := callsign.Parse("DL1ABC")
	values := func() core.KeyerValues {
		return core.KeyerValues{
			MyCall:    myCall,
			TheirCall: "DL0ZZZ",
			MyNumber:  core.QSONumber(456),
			MyReport:  core.RST("123"),
		}
	}
	cwClient := new(mocked.CWClient)
	cwClient.On("Send", mock.Anything).Once()

	keyer, err := New(patterns, cwClient, values)
	require.NoError(t, err)

	keyer.Send(0)

	cwClient.AssertExpectations(t)
}

// Mocks
