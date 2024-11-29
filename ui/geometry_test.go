//go:build !fyne

package ui

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDegreesToRadians(t *testing.T) {
	tt := []struct {
		value    float64
		expected float64
	}{
		{value: 0, expected: 0},
		{value: 60, expected: math.Pi / 3},
		{value: 90, expected: math.Pi / 2},
		{value: 120, expected: 2 * math.Pi / 3},
		{value: 180, expected: math.Pi},
		{value: 240, expected: 4 * math.Pi / 3},
		{value: 270, expected: 3 * math.Pi / 2},
		{value: 360, expected: 2 * math.Pi},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			actual := degreesToRadians(tc.value)
			assert.InDelta(t, tc.expected, actual, 1e-10)
		})
	}
}
