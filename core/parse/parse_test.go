package parse

import (
	"testing"

	"github.com/ftl/hellocontest/core"
)

func TestRST(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		valid    bool
		expected core.RST
	}{
		{"valid CW report", "599", true, "599"},
		{"valid SSB report", "59", true, "59"},
		{"valid FM repeater report", "5", true, "5"},
		{"with whitespace", " 599 ", true, "599"},
		{"empty string is invalid", "", false, ""},
		{"single digit out of range", "6", false, ""},
		{"double digit out of range", "40", false, ""},
		{"trible digit out of range", "480", false, ""},
		{"invalid characters", "a-b", false, ""},
		{"too long", "1234", false, ""},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual, err := RST(tC.value)
			if err != nil && tC.valid {
				t.Errorf("expected to be valid, but got error %v", err)
			}
			if err == nil && !tC.valid {
				t.Errorf("%q should not be parsed successfully", tC.value)
			}
			if tC.valid && actual != tC.expected {
				t.Errorf("%q: expected %v but got %v", tC.value, tC.expected, actual)
			}
		})
	}
}
