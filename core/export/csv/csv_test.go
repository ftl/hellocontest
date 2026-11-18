package csv

import (
	"bytes"
	"testing"
	"time"

	"github.com/ftl/hellocontest/core/model"
	"github.com/stretchr/testify/assert"
)

func TestExportCSV(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})

	err := Export(buffer, model.MustParseCallsign("DL0ABC"), model.QSO{
		Callsign:      model.MustParseCallsign("DL1ABC"),
		Time:          time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
		Band:          model.Band40m,
		Mode:          model.ModeCW,
		MyReport:      model.RST("599"),
		MyNumber:      model.QSONumber(1),
		MyExchange:    []string{"599", "001", "ABC"},
		TheirReport:   model.RST("589"),
		TheirNumber:   model.QSONumber(4),
		TheirExchange: []string{"589", "004", "DEF"},
	})
	expected := "40m,0.000,CW,2009-05-30,0002,DL0ABC,599,001,ABC,DL1ABC,589,004,DEF,,,0,0,0,\n"

	assert.NoError(t, err)
	assert.Equal(t, expected, buffer.String())
}
