package cfg

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/cfg"
	"github.com/ftl/hamradio/locator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoaded_MyCall(t *testing.T) {
	expected, err := callsign.Parse("dl1abc")
	require.NoError(t, err)
	config := loadFromString(t, `{"my":{"call":"dl1abc"}}`)

	assert.Equal(t, expected, config.MyCall())
}

func TestLoaded_MyLocator(t *testing.T) {
	expected, err := locator.Parse("KM12DF")
	require.NoError(t, err)
	config := loadFromString(t, `{"my":{"locator":"km12df"}}`)

	assert.Equal(t, expected, config.MyLocator())
}

func TestLoaded_EnterTheirNumber(t *testing.T) {
	config1 := loadFromString(t, `{"hellocontest":{"enter":{"theirNumber":true}}}`)
	assert.True(t, config1.EnterTheirNumber())
	config2 := loadFromString(t, `{"hellocontest":{"enter":{"theirNumber":false}}}`)
	assert.False(t, config2.EnterTheirNumber())
}

func TestLoaded_EnterTheirXchange(t *testing.T) {
	config1 := loadFromString(t, `{"hellocontest":{"enter":{"theirXchange":true}}}`)
	assert.True(t, config1.EnterTheirXchange())
	config2 := loadFromString(t, `{"hellocontest":{"enter":{"theirXchange":false}}}`)
	assert.False(t, config2.EnterTheirXchange())
}

func TestLoaded_KeyerSPPatterns(t *testing.T) {
	testCases := []struct {
		value    string
		expected []string
	}{
		{"", []string{}},
		{`"sp":[]`, []string{}},
		{`"sp":["one"]`, []string{"one"}},
	}
	for i, tC := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			config := loadFromString(t, fmt.Sprintf(`{"hellocontest":{"keyer":{%s}}}`, tC.value))
			assert.Equal(t, tC.expected, config.KeyerSPPatterns())
		})
	}
}

func loadFromString(t *testing.T, s string) *loaded {
	raw, err := cfg.Read(bytes.NewBufferString(s))
	require.NoError(t, err)
	return &loaded{configuration: raw}
}
