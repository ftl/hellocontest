package model

import (
	"strings"
	"testing"

	"github.com/ftl/conval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummary_ScoreTable(t *testing.T) {
	score := NewScore()
	score.MultiProperties = []conval.Property{conval.CQZoneProperty, conval.DXCCEntityProperty}
	score.ScorePerBand[Band80m] = BandScore{
		QSOs: 42, Duplicates: 1, Points: 126, Multis: 25,
		MultisPerProperty: map[conval.Property]int{conval.DXCCEntityProperty: 18, conval.CQZoneProperty: 7},
	}

	tt := []struct {
		desc           string
		detailedMultis bool
		properties     []conval.Property
		expected       []string
	}{
		{
			desc: "sum mode",
			expected: []string{
				"Band QSOs Dupes Points P/Q Mult Q/M Result",
				"80m    42     1    126 3.0   25 1.7   3150",
				"Tot    42     1    126 3.0   25 1.7   3150",
			},
		},
		{
			desc:           "detailed mode",
			detailedMultis: true,
			expected: []string{
				"Band QSOs Dupes Points P/Q cq_zone dxcc_entity Q/M Result",
				"80m    42     1    126 3.0       7          18 1.7   3150",
				"Tot    42     1    126 3.0       7          18 1.7   3150",
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			summary := Summary{Score: score, DetailedMultis: tc.detailedMultis}

			actual := strings.Split(strings.TrimSpace(summary.ScoreTable()), "\n")

			require.Len(t, actual, 5)
			assert.Equal(t, tc.expected[0], actual[0], "header")
			assert.Equal(t, tc.expected[1], actual[2], "band")
			assert.Equal(t, tc.expected[2], actual[4], "total")
		})
	}
}

func TestSummary_ScoreTable_WithoutMultipliers(t *testing.T) {
	score := NewScore()
	score.ScorePerBand[Band80m] = BandScore{QSOs: 42, Points: 126}
	summary := Summary{Score: score}

	actual := strings.Split(strings.TrimSpace(summary.ScoreTable()), "\n")

	assert.Equal(t, "Band QSOs Dupes Points P/Q Result", actual[0], "neither a multiplier column nor Q/M")
}
