package model

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseBand parses a string into a HF band value
func ParseBand(s string) (Band, error) {
	for _, band := range Bands {
		if string(band) == s {
			return band, nil
		}
	}
	return NoBand, fmt.Errorf("%q is not a supported HF band", s)
}

// ParseMode parses a string into a HF Mode value
func ParseMode(s string) (Mode, error) {
	for _, mode := range Modes {
		if string(mode) == s {
			return mode, nil
		}
	}
	return NoMode, fmt.Errorf("%q is not a supported mode", s)
}

var parseRSTExpression = regexp.MustCompile("\\b[1-5]([1-9]([1-9])?)?\\b")

// ParseRST parses the given string for a report and returns the parsed RST value.
func ParseRST(s string) (RST, error) {
	normalized := strings.TrimSpace(s)
	length := len(normalized)
	if length == 0 {
		return RST(""), fmt.Errorf("The report in RST notation must not be empty")
	}
	if length > 3 {
		return RST(""), fmt.Errorf("%q is not a valid report in RST notation", s)
	}
	if !parseRSTExpression.MatchString(normalized) {
		return RST(""), fmt.Errorf("%q is not a valid report in RST notation", s)
	}
	return RST(normalized), nil
}
