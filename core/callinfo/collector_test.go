package callinfo

import (
	"testing"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
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

func TestCollector_addDXCC(t *testing.T) {
	dxccFinder := dxcc.New()
	dxccFinder.WaitUntilAvailable(10 * time.Second)
	require.True(t, dxccFinder.Available(), "DXCC database not available")

	tests := []struct {
		input              string
		failure            bool
		expectedDXCCPrefix string
	}{
		{input: "", failure: true},
		{input: "D", failure: true},
		{input: "DL", expectedDXCCPrefix: "DL"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			c := &Collector{
				dxcc: dxccFinder,
			}
			info := core.Callinfo{Input: test.input}
			success := c.addDXCC(&info)
			assert.NotEqual(t, test.failure, success, "success")
			if test.failure {
				assert.Equal(t, dxcc.Prefix{}, info.DXCCEntity, "DXCC entity")
			} else {
				assert.Equal(t, test.expectedDXCCPrefix, info.DXCCEntity.Prefix, "DXCC prefix")
			}
		})
	}
}

func TestCollector_predictExchange(t *testing.T) {
	rstProperty := core.ExchangeField{
		Field:            "theirExchange_report",
		Properties:       conval.ExchangeField{conval.RSTProperty},
		CanContainReport: true,
	}
	nameProperty := core.ExchangeField{
		Field:            "theirExchange_name",
		Properties:       conval.ExchangeField{conval.NameProperty},
		CanContainReport: true,
	}

	tests := []struct {
		name                string
		theirExchangeFields []core.ExchangeField
		dxcc                dxcc.Prefix
		qsos                []core.QSO
		currentExchange     []string
		historicExchange    []string
		expected            []string
	}{
		{
			name:                "only report, the entry field must be initialized with the default report",
			theirExchangeFields: []core.ExchangeField{rstProperty},
			currentExchange:     []string{"59"},
			historicExchange:    []string{""},
			expected:            []string{"59"},
		},
		{
			name:                "name, empty",
			theirExchangeFields: []core.ExchangeField{nameProperty},
			currentExchange:     []string{""},
			historicExchange:    []string{""},
			expected:            []string{""},
		},
		{
			name:                "name, only current",
			theirExchangeFields: []core.ExchangeField{nameProperty},
			currentExchange:     []string{"Flo"},
			historicExchange:    []string{""},
			expected:            []string{"Flo"},
		},
		{
			name:                "name, with history",
			theirExchangeFields: []core.ExchangeField{nameProperty},
			currentExchange:     []string{""},
			historicExchange:    []string{"Flo"},
			expected:            []string{"Flo"},
		},
		{
			name:                "name, history over current",
			theirExchangeFields: []core.ExchangeField{nameProperty},
			currentExchange:     []string{"Hans"},
			historicExchange:    []string{"Flo"},
			expected:            []string{"Flo"},
		},
		{
			name:                "name, qso over history",
			theirExchangeFields: []core.ExchangeField{nameProperty},
			qsos:                []core.QSO{{TheirExchange: []string{"Steve"}}},
			currentExchange:     []string{""},
			historicExchange:    []string{"Flo"},
			expected:            []string{"Steve"},
		},
		{
			name:                "name, history over unclear qso",
			theirExchangeFields: []core.ExchangeField{nameProperty},
			qsos:                []core.QSO{{TheirExchange: []string{"Steve"}}, {TheirExchange: []string{"Bud"}}},
			currentExchange:     []string{""},
			historicExchange:    []string{"Flo"},
			expected:            []string{"Flo"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Collector{
				theirExchangeFields: test.theirExchangeFields,
			}
			actual := c.predictExchange(test.dxcc, test.qsos, test.currentExchange, test.historicExchange)
			assert.Equal(t, test.expected, actual)
		})
	}
}
