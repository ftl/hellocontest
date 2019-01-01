package cabrillo

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
)

func TestQsoLine(t *testing.T) {
	template := template.Must(template.New("").Parse("{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}} {{.TheirCall}} {{.TheirReport}} {{.TheirNumber}} {{.TheirXchange}}"))
	myCall, _ := callsign.Parse("AA1ZZZ")
	theirCall, _ := callsign.Parse("S50A")
	testCases := []struct {
		desc     string
		qso      core.QSO
		expected string
	}{
		{
			desc: "40m CW",
			qso: core.QSO{
				Callsign:     theirCall,
				Time:         time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
				Band:         core.Band40m,
				Mode:         core.ModeCW,
				MyReport:     core.RST("599"),
				MyNumber:     core.QSONumber(1),
				MyXchange:    "ABC",
				TheirReport:  core.RST("589"),
				TheirNumber:  core.QSONumber(4),
				TheirXchange: "DEF",
			},
			expected: "QSO: 7000 CW 2009-05-30 0002 AA1ZZZ 599 001 ABC S50A 589 004 DEF\n",
		},
		{
			desc: "20m SSB",
			qso: core.QSO{
				Callsign:     theirCall,
				Time:         time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
				Band:         core.Band20m,
				Mode:         core.ModeSSB,
				MyReport:     core.RST("59"),
				MyNumber:     core.QSONumber(1),
				MyXchange:    "XXX",
				TheirReport:  core.RST("58"),
				TheirNumber:  core.QSONumber(4),
				TheirXchange: "YYY",
			},
			expected: "QSO: 14000 PH 2009-05-30 0002 AA1ZZZ 59 001 XXX S50A 58 004 YYY\n",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			buffer := bytes.NewBuffer([]byte{})
			err := writeQSO(buffer, template, myCall, tC.qso)
			assert.NoError(t, err)
			assert.Equal(t, tC.expected, buffer.String())
		})
	}
}

func TestExport(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})
	template := template.Must(template.New("").Parse("{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyReport}} {{.MyNumber}} {{.TheirCall}} {{.TheirReport}} {{.TheirXchange}}"))
	myCall, _ := callsign.Parse("AA1ZZZ")
	theirCall, _ := callsign.Parse("S50A")
	qso := core.QSO{
		Callsign:     theirCall,
		Time:         time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
		Band:         core.Band40m,
		Mode:         core.ModeCW,
		MyReport:     core.RST("599"),
		MyNumber:     core.QSONumber(1),
		MyXchange:    "ABC",
		TheirReport:  core.RST("589"),
		TheirNumber:  core.QSONumber(4),
		TheirXchange: "DEF",
	}

	expected := `START-OF-LOG: 3.0
CREATED-BY: Hello Contest
CALLSIGN: AA1ZZZ
QSO: 7000 CW 2009-05-30 0002 AA1ZZZ 599 001 S50A 589 DEF
END-OF-LOG:
`

	Export(buffer, template, myCall, qso)

	assert.Equal(t, expected, buffer.String())
}
