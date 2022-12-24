package csv

import (
	"bytes"
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
)

func TestExportCSV(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})

	err := Export(buffer, callsign.MustParse("DL0ABC"), core.QSO{
		Callsign:      callsign.MustParse("DL1ABC"),
		Time:          time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
		Band:          core.Band40m,
		Mode:          core.ModeCW,
		MyReport:      core.RST("599"),
		MyNumber:      core.QSONumber(1),
		MyExchange:    []string{"599", "001", "ABC"},
		TheirReport:   core.RST("589"),
		TheirNumber:   core.QSONumber(4),
		TheirExchange: []string{"589", "004", "DEF"},
	})
	expected := "40m,0.000,CW,2009-05-30,0002,DL0ABC,599,001,ABC,DL1ABC,589,004,DEF,,,0,0,0,\n"

	assert.NoError(t, err)
	assert.Equal(t, expected, buffer.String())
}
