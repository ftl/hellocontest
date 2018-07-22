package core

import (
	"fmt"
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTemplateHandling(t *testing.T) {
	patterns := []string{"{{.MyCall}}", "{{.TheirCall}}", "{{.MyNumber}}", "{{.MyReport}}"}
	expected := []string{"DL1ABC", "DL0ZZZ", "456", "123"}
	myCall, _ := callsign.Parse("DL1ABC")
	values := KeyerValues{
		MyCall:    myCall,
		TheirCall: "DL0ZZZ",
		MyNumber:  QSONumber(456),
		MyReport:  RST("123"),
	}

	keyer, err := NewKeyer(patterns, new(mockCWClient))
	require.NoError(t, err)

	for i, pattern := range patterns {
		assert.Equal(t, pattern, keyer.GetTemplate(i))
		actual, err := keyer.GetText(i, values)
		require.NoError(t, err)
		assert.Equal(t, expected[i], actual)

		keyer.SetTemplate(i, fmt.Sprintf("%s %d", patterns[i], i))
		actual, err = keyer.GetText(i, values)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%s %d", expected[i], i), actual)
	}
}

func TestSend(t *testing.T) {
	patterns := []string{"{{.MyCall}} {{.TheirCall}} {{.MyNumber}} {{.MyReport}}"}
	myCall, _ := callsign.Parse("DL1ABC")
	values := KeyerValues{
		MyCall:    myCall,
		TheirCall: "DL0ZZZ",
		MyNumber:  QSONumber(456),
		MyReport:  RST("123"),
	}
	cwClient := new(mockCWClient)
	cwClient.On("Send", mock.Anything).Once()

	keyer, err := NewKeyer(patterns, cwClient)
	require.NoError(t, err)

	keyer.Send(0, values)

	cwClient.AssertExpectations(t)
}

// Mocks

type mockCWClient struct {
	mock.Mock
}

func (m *mockCWClient) Send(text string) {
	m.Called(text)
}
