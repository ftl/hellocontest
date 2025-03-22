package callinfo

import (
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
)

func TestCollector_addCallsign(t *testing.T) {
	tests := []struct {
		name     string
		input    core.Callinfo
		expected core.Callinfo
		failure  bool
	}{
		{
			name:     "empty input",
			input:    core.Callinfo{},
			expected: core.Callinfo{},
			failure:  true,
		},
		{
			name:     "valid callsign",
			input:    core.Callinfo{Input: "DL1ABC"},
			expected: core.Callinfo{Input: "DL1ABC", Call: callsign.MustParse("DL1ABC"), CallValid: true},
		},
		{
			name:     "incomplete callsign",
			input:    core.Callinfo{Input: "DL"},
			expected: core.Callinfo{Input: "DL"},
			failure:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Collector{}
			success := c.addCallsign(&test.input)
			assert.NotEqual(t, test.failure, success)
			assert.Equal(t, test.expected, test.input)
		})
	}
}

func TestCollector_GetInfoForInput_normalizesInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "", expected: ""},
		{input: " dl1abc", expected: "DL1ABC"},
		{input: "dl1abc ", expected: "DL1ABC"},
		{input: " dl1abc ", expected: "DL1ABC"},
		{input: " DL1abc ", expected: "DL1ABC"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			c := &Collector{}
			actual := c.GetInfoForInput(test.input, core.Band10m, core.ModeCW, []string{})
			assert.Equal(t, test.expected, actual.Input)
		})
	}
}
