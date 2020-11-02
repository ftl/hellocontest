package cfg

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/locator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoaded_MyCall(t *testing.T) {
	expected, err := callsign.Parse("dl1abc")
	require.NoError(t, err)
	config := loadFromString(t, `{"mycall":"dl1abc"}`)

	assert.Equal(t, expected, config.MyCall())
}

func TestLoaded_MyLocator(t *testing.T) {
	expected, err := locator.Parse("KM12DF")
	require.NoError(t, err)
	config := loadFromString(t, `{"locator":"km12df"}`)

	assert.Equal(t, expected, config.MyLocator())
}

func TestLoaded_EnterTheirNumber(t *testing.T) {
	config1 := loadFromString(t, `{"enter_their_number":true}`)
	assert.True(t, config1.EnterTheirNumber())
	config2 := loadFromString(t, `{"enter_their_number":false}`)
	assert.False(t, config2.EnterTheirNumber())
}

func TestLoaded_EnterTheirXchange(t *testing.T) {
	config1 := loadFromString(t, `{"enter_their_xchange":true}`)
	assert.True(t, config1.EnterTheirXchange())
	config2 := loadFromString(t, `{"enter_their_xchange":false}`)
	assert.False(t, config2.EnterTheirXchange())
}

func TestLoaded_KeyerSPMacros(t *testing.T) {
	testCases := []struct {
		value    string
		expected []string
	}{
		{"", []string{}},
		{`"keyer_sp_macros":[]`, []string{}},
		{`"keyer_sp_macros":["one"]`, []string{"one"}},
	}
	for i, tC := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			config := loadFromString(t, fmt.Sprintf(`{%s}`, tC.value))
			assert.Equal(t, tC.expected, config.KeyerSPMacros())
		})
	}
}

func loadFromString(t *testing.T, s string) *LoadedConfiguration {
	var data Data
	err := json.Unmarshal([]byte(s), &data)
	require.NoError(t, err)
	return &LoadedConfiguration{data: data}
}
