package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ftl/hellocontest/core/model"
)

// Band parses a string into a HF band value
func Band(s string) (model.Band, error) {
	for _, band := range model.Bands {
		if string(band) == s {
			return band, nil
		}
	}
	return model.NoBand, fmt.Errorf("%q is not a supported HF band", s)
}

// Mode parses a string into a HF Mode value
func Mode(s string) (model.Mode, error) {
	for _, mode := range model.Modes {
		if string(mode) == s {
			return mode, nil
		}
	}
	return model.NoMode, fmt.Errorf("%q is not a supported mode", s)
}

var parseRSTExpression = regexp.MustCompile("\\b[1-5]([1-9]([1-9])?)?\\b")

// RST parses the given string for a report and returns the parsed RST value.
func RST(s string) (model.RST, error) {
	normalized := strings.TrimSpace(s)
	length := len(normalized)
	if length == 0 {
		return model.RST(""), fmt.Errorf("The report in RST notation must not be empty")
	}
	if length > 3 {
		return model.RST(""), fmt.Errorf("%q is not a valid report in RST notation", s)
	}
	if !parseRSTExpression.MatchString(normalized) {
		return model.RST(""), fmt.Errorf("%q is not a valid report in RST notation", s)
	}
	return model.RST(normalized), nil
}
