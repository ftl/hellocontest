package core

import (
	"fmt"
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateHandling(t *testing.T) {
	patterns := []string{"{{.MyCall}}", "{{.TheirCall}}", "{{.MyNumber}}", "{{.MyReport}}"}
	expected := []string{"DL1ABC", "DL0ZZZ", "456", "123"}
	myCall, _ := callsign.Parse("DL1ABC")
	values := KeyerValues{
		MyCall:    myCall,
		TheirCall: "DL0ZZZ",
		MyNumber:  QSONumber(456),
		MyReport:  RST("123"),
	}

	keyer, err := NewKeyer(patterns)
	require.NoError(t, err)

	for i, pattern := range patterns {
		assert.Equal(t, pattern, keyer.GetTemplate(i))
		assert.Equal(t, expected[i], keyer.GetText(i, values))

		keyer.SetTemplate(i, fmt.Sprintf("%s %d", patterns[i], i))
		assert.Equal(t, fmt.Sprintf("%s %d", expected[i], i), keyer.GetText(i, values))
	}
}
