package cabrillo

import (
	"testing"

	"github.com/ftl/cabrillo"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/conval"
)

func TestConvertOverlay(t *testing.T) {
	tt := []struct {
		value    conval.Overlay
		expected cabrillo.CategoryOverlay
	}{
		{"", ""},
		{"classic", "CLASSIC"},
		{"tb_wires", "TB-WIRES"},
		{"wire_only", "WIRE-ONLY"},
		{"rookie", "ROOKIE"},
		{"youth", "YOUTH"},
		{"yl", "YL"},
		{"yn", "YN"},
		{"teen", "TEEN"},
		{"newcomer", "NEWCOMER"},
		{"dxpedition", "DXPEDITION"},
		{"single_element", "SINGLE-ELEMENT"},
		{"12_hour", "12-HOUR"},
		{"novice_tech", "NOVICE-TECH"},
		{"over_50", "OVER-50"},
	}
	for _, tc := range tt {
		t.Run(string(tc.value), func(t *testing.T) {
			assert.Equal(t, tc.expected, convertOverlay(tc.value))
		})
	}
}
