package entry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallsignRequest(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "?"},
		{"DL1ABC", "DL1ABC?"},
		{"DL", "DL?"},
		{"G/DL1", "G/DL1?"},
		{"DL.ABC", "DL?"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual := callsignRequest(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}
