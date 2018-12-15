package cfg

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/cfg"
	"github.com/ftl/hamradio/locator"
	"github.com/ftl/hellocontest/core"
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

func TestLoaded_MyExchanger(t *testing.T) {
	testCases := []struct {
		value    string
		expected core.Exchanger
	}{
		{"", core.MyNumber},
		{"Number", core.MyNumber},
		{"NUMBER", core.MyNumber},
		{"Xchange", core.MyXchange},
		{"XCHANGE", core.MyXchange},
		{"both", core.MyNumberAndXchange},
		{"Both", core.MyNumberAndXchange},
		{"none", core.NoExchange},
		{"NONE", core.NoExchange},
	}
	for i, tC := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			config := loadFromString(t, fmt.Sprintf(`{"hellocontest":{"exchange":{"my":"%s"}}}`, tC.value))
			qso := core.QSO{MyNumber: core.QSONumber(123), MyXchange: "ABC"}
			assert.Equal(t, tC.expected(qso), config.MyExchanger()(qso))
		})
	}
}

func TestLoaded_TheirExchanger(t *testing.T) {
	testCases := []struct {
		value    string
		expected core.Exchanger
	}{
		{"", core.TheirNumber},
		{"Number", core.TheirNumber},
		{"NUMBER", core.TheirNumber},
		{"Xchange", core.TheirXchange},
		{"XCHANGE", core.TheirXchange},
		{"both", core.TheirNumberAndXchange},
		{"Both", core.TheirNumberAndXchange},
		{"none", core.NoExchange},
		{"NONE", core.NoExchange},
	}
	for i, tC := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			config := loadFromString(t, fmt.Sprintf(`{"hellocontest":{"exchange":{"their":"%s"}}}`, tC.value))
			qso := core.QSO{TheirNumber: core.QSONumber(123), TheirXchange: "ABC"}
			assert.Equal(t, tC.expected(qso), config.TheirExchanger()(qso))
		})
	}
}

func loadFromString(t *testing.T, s string) *loaded {
	raw, err := cfg.Read(bytes.NewBufferString(s))
	require.NoError(t, err)
	return &loaded{configuration: raw}
}
