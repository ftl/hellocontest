package callinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core/dxcc"
	"github.com/ftl/hellocontest/core/model"
)

func TestCollector_addCallsign(t *testing.T) {
	tests := []struct {
		name     string
		input    model.Callinfo
		expected model.Callinfo
		failure  bool
	}{
		{
			name:     "empty input",
			input:    model.Callinfo{},
			expected: model.Callinfo{},
			failure:  true,
		},
		{
			name:     "valid callsign",
			input:    model.Callinfo{Input: "DL1ABC"},
			expected: model.Callinfo{Input: "DL1ABC", Call: model.MustParseCallsign("DL1ABC"), CallValid: true},
		},
		{
			name:     "incomplete callsign",
			input:    model.Callinfo{Input: "DL"},
			expected: model.Callinfo{Input: "DL"},
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
			actual := c.GetInfoForInput(test.input, model.Band10m, model.ModeCW, []string{})
			assert.Equal(t, test.expected, actual.Input)
		})
	}
}

func TestCollector_addDXCC(t *testing.T) {
	dxccFinder, err := dxcc.NewFromFile("testdata/cty.dat")
	require.NoError(t, err)

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
			info := model.Callinfo{Input: test.input}
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

type dupeCheckerStub struct {
	workedQSOs []model.QSO
	duplicate  bool
}

func (s *dupeCheckerStub) FindWorkedQSOs(model.Callsign, model.Band, model.Mode) ([]model.QSO, bool) {
	return s.workedQSOs, s.duplicate
}

func (s *dupeCheckerStub) Find(string) ([]model.AnnotatedCallsign, error) {
	return nil, nil
}

type valuerStub struct {
	points int
	multis int
}

func (s *valuerStub) Value(model.Callsign, dxcc.Prefix, model.Band, model.Mode, []string) (int, int, map[conval.Property]string) {
	return s.points, s.multis, nil
}

func TestCollector_UpdateValue_RefreshesTheDuplicateState(t *testing.T) {
	dupes := &dupeCheckerStub{}
	collector := NewCollector(new(dxcc.Finder), nil, nil, dupes, &valuerStub{points: 2}, nil)
	info := model.Callinfo{Call: model.MustParseCallsign("DL1ABC")}

	require.True(t, collector.UpdateValue(&info, model.Band80m, model.ModeCW))
	require.False(t, info.Duplicate, "the station was not worked yet")

	// the QSO is logged
	dupes.duplicate = true
	dupes.workedQSOs = []model.QSO{{Callsign: model.MustParseCallsign("DL1ABC")}}

	require.True(t, collector.UpdateValue(&info, model.Band80m, model.ModeCW))

	assert.True(t, info.Duplicate, "the update of the value sees the logged QSO")
	assert.True(t, info.Worked)
}
