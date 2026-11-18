package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFrequencyJump(t *testing.T) {
	tt := []struct {
		desc string
		from Frequency
		to   Frequency
		want bool
	}{
		{desc: "no change", from: 14020000, to: 14020000},
		{desc: "tuning around the frequency", from: 14020000, to: 14020250},
		{desc: "tuning away", from: 14020000, to: 14020251, want: true},
		{desc: "tuning down and away", from: 14020000, to: 14019749, want: true},
		{desc: "another band", from: 14020000, to: 7030000, want: true},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.want, IsFrequencyJump(tc.from, tc.to))
		})
	}
}

func TestParseRST(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		valid    bool
		expected RST
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
			actual, err := ParseRST(tC.value)
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
