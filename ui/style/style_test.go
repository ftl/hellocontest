package style

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColor_ToWeb(t *testing.T) {
	tt := []struct {
		desc     string
		value    Color
		expected string
	}{
		{
			desc:     "black",
			value:    Black,
			expected: "#000000",
		},
		{
			desc:     "white",
			value:    White,
			expected: "#ffffff",
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			actual := tc.value.ToWeb()
			assert.Equal(t, tc.expected, actual)
		})
	}
}
