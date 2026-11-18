package callinfo

import (
	"testing"

	"github.com/ftl/conval"
	"github.com/ftl/hellocontest/core/dxcc"
	"github.com/ftl/hellocontest/core/model"
	"github.com/stretchr/testify/assert"
)

func TestPredictExchange(t *testing.T) {
	rstProperty := model.ExchangeField{
		Field:            "theirExchange_report",
		Properties:       conval.ExchangeField{conval.RSTProperty},
		CanContainReport: true,
	}
	nameProperty := model.ExchangeField{
		Field:            "theirExchange_name",
		Properties:       conval.ExchangeField{conval.NameProperty},
		CanContainReport: true,
	}
	memberNoMemberField := model.ExchangeField{
		Field:      "theirExchange_member",
		Properties: conval.ExchangeField{conval.MemberNumberProperty, conval.NoMemberProperty},
	}
	noMemberMemberField := model.ExchangeField{
		Field:      "theirExchange_member",
		Properties: conval.ExchangeField{conval.NoMemberProperty, conval.MemberNumberProperty},
	}
	memberOnlyField := model.ExchangeField{
		Field:      "theirExchange_member",
		Properties: conval.ExchangeField{conval.MemberNumberProperty},
	}

	tests := []struct {
		name                string
		theirExchangeFields []model.ExchangeField
		dxcc                dxcc.Prefix
		qsos                []model.QSO
		currentExchange     []string
		historicExchange    []string
		historyAvailable    []bool
		expected            []string
	}{
		{
			name:                "only report, the entry field must be initialized with the default report",
			theirExchangeFields: []model.ExchangeField{rstProperty},
			currentExchange:     []string{"59"},
			historicExchange:    []string{""},
			expected:            []string{"59"},
		},
		{
			name:                "name, empty",
			theirExchangeFields: []model.ExchangeField{nameProperty},
			currentExchange:     []string{""},
			historicExchange:    []string{""},
			expected:            []string{""},
		},
		{
			name:                "name, only current",
			theirExchangeFields: []model.ExchangeField{nameProperty},
			currentExchange:     []string{"Flo"},
			historicExchange:    []string{""},
			expected:            []string{"Flo"},
		},
		{
			name:                "name, with history",
			theirExchangeFields: []model.ExchangeField{nameProperty},
			currentExchange:     []string{""},
			historicExchange:    []string{"Flo"},
			expected:            []string{"Flo"},
		},
		{
			name:                "name, history over current",
			theirExchangeFields: []model.ExchangeField{nameProperty},
			currentExchange:     []string{"Hans"},
			historicExchange:    []string{"Flo"},
			expected:            []string{"Flo"},
		},
		{
			name:                "name, qso over history",
			theirExchangeFields: []model.ExchangeField{nameProperty},
			qsos:                []model.QSO{{TheirExchange: []string{"Steve"}}},
			currentExchange:     []string{""},
			historicExchange:    []string{"Flo"},
			expected:            []string{"Steve"},
		},
		{
			name:                "name, history over unclear qso",
			theirExchangeFields: []model.ExchangeField{nameProperty},
			qsos:                []model.QSO{{TheirExchange: []string{"Steve"}}, {TheirExchange: []string{"Bud"}}},
			currentExchange:     []string{""},
			historicExchange:    []string{"Flo"},
			expected:            []string{"Flo"},
		},
		{
			name:                "member, number found in history",
			theirExchangeFields: []model.ExchangeField{memberNoMemberField},
			currentExchange:     []string{""},
			historicExchange:    []string{"1234"},
			historyAvailable:    []bool{true},
			expected:            []string{"1234"},
		},
		{
			name:                "member, callsign found but member number blank, predicts nm",
			theirExchangeFields: []model.ExchangeField{memberNoMemberField},
			currentExchange:     []string{""},
			historicExchange:    []string{""},
			historyAvailable:    []bool{true},
			expected:            []string{"nm"},
		},
		{
			name:                "member, callsign not found but history available, predicts nm",
			theirExchangeFields: []model.ExchangeField{memberNoMemberField},
			currentExchange:     []string{""},
			historicExchange:    nil,
			historyAvailable:    []bool{true},
			expected:            []string{"nm"},
		},
		{
			name:                "member, order reversed, predicts nm",
			theirExchangeFields: []model.ExchangeField{noMemberMemberField},
			currentExchange:     []string{""},
			historicExchange:    []string{""},
			historyAvailable:    []bool{true},
			expected:            []string{"nm"},
		},
		{
			name:                "member, history not available, stays empty",
			theirExchangeFields: []model.ExchangeField{memberNoMemberField},
			currentExchange:     []string{""},
			historicExchange:    []string{""},
			historyAvailable:    []bool{false},
			expected:            []string{""},
		},
		{
			name:                "member, field unmapped in history, stays empty",
			theirExchangeFields: []model.ExchangeField{memberNoMemberField},
			currentExchange:     []string{""},
			historicExchange:    []string{""},
			historyAvailable:    nil,
			expected:            []string{""},
		},
		{
			name:                "member only, no nm property, stays empty",
			theirExchangeFields: []model.ExchangeField{memberOnlyField},
			currentExchange:     []string{""},
			historicExchange:    []string{""},
			historyAvailable:    []bool{true},
			expected:            []string{""},
		},
		{
			name:                "member, worked qso number over nm",
			theirExchangeFields: []model.ExchangeField{memberNoMemberField},
			qsos:                []model.QSO{{TheirExchange: []string{"5678"}}},
			currentExchange:     []string{""},
			historicExchange:    []string{""},
			historyAvailable:    []bool{true},
			expected:            []string{"5678"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := predictExchange(test.theirExchangeFields, test.dxcc, test.qsos, test.currentExchange, test.historicExchange, test.historyAvailable)
			assert.Equal(t, test.expected, actual)
		})
	}
}
