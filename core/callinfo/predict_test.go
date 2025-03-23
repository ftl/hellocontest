package callinfo

import (
	"testing"

	"github.com/ftl/conval"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
	"github.com/stretchr/testify/assert"
)

func TestPredictExchange(t *testing.T) {
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
			actual := predictExchange(test.theirExchangeFields, test.dxcc, test.qsos, test.currentExchange, test.historicExchange)
			assert.Equal(t, test.expected, actual)
		})
	}
}
