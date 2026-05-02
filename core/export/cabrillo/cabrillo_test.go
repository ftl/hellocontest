package cabrillo

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/locator"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestQsoLine(t *testing.T) {
	template := template.Must(template.New("").Parse("{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyExchange}} {{.TheirCall}} {{.TheirExchange}}"))
	myCall, _ := core.ParseCallsign("AA1ZZZ")
	theirCall, _ := core.ParseCallsign("S50A")
	testCases := []struct {
		desc     string
		qso      core.QSO
		expected string
	}{
		{
			desc: "40m CW",
			qso: core.QSO{
				Callsign:      theirCall,
				Time:          time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
				Band:          core.Band40m,
				Mode:          core.ModeCW,
				MyReport:      core.RST("599"),
				MyNumber:      core.QSONumber(1),
				MyExchange:    []string{"599", "001", "ABC"},
				TheirReport:   core.RST("589"),
				TheirNumber:   core.QSONumber(4),
				TheirExchange: []string{"589", "004", "DEF"},
			},
			expected: "QSO: 7000 CW 2009-05-30 0002 AA1ZZZ 599 001 ABC S50A 589 004 DEF\n",
		},
		{
			desc: "20m SSB",
			qso: core.QSO{
				Callsign:      theirCall,
				Time:          time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
				Band:          core.Band20m,
				Mode:          core.ModeSSB,
				MyReport:      core.RST("59"),
				MyNumber:      core.QSONumber(1),
				MyExchange:    []string{"59", "001", "XXX"},
				TheirReport:   core.RST("58"),
				TheirNumber:   core.QSONumber(4),
				TheirExchange: []string{"58", "004", "YYY"},
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
	settings := &testSettings{
		stationCallsign:   "AA1ZZZ",
		stationOperator:   "AA2ZZZ",
		stationLocator:    "AA00AA",
		contestIdentifier: "HELLO-CONTEST-CABRILLO-TEST",
	}
	theirCall := core.MustParseCallsign("S50A")
	qso := core.QSO{
		Callsign:      theirCall,
		Time:          time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
		Band:          core.Band40m,
		Mode:          core.ModeCW,
		MyReport:      core.RST("599"),
		MyNumber:      core.QSONumber(1),
		MyExchange:    []string{"599", "001", "ABC"},
		TheirReport:   core.RST("589"),
		TheirNumber:   core.QSONumber(4),
		TheirExchange: []string{"589", "004", "DEF"},
	}
	qtc := core.QTC{
		Kind:          core.SentQTC,
		QSONumber:     1,
		TheirCallsign: theirCall,
		Header:        core.QTCHeader{SeriesNumber: 321, QTCCount: 1},
		Timestamp:     time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
		Frequency:     7020000,
		Band:          core.Band40m,
		Mode:          core.ModeCW,
		QTCTime:       core.QTCTime{Hour: 12, Minute: 34},
		QTCCallsign:   core.MustParseCallsign("DL1ABC"),
		QTCNumber:     5,
	}

	expected := `START-OF-LOG: 3.0
CREATED-BY: Hello Contest
CONTEST: HELLO-CONTEST-CABRILLO-TEST
CALLSIGN: AA1ZZZ
OPERATORS: AA2ZZZ
GRID-LOCATOR: AA00aa
CLAIMED-SCORE: 123
CERTIFICATE: NO
QSO: 7000 CW 2009-05-30 0002 AA1ZZZ 599 001 ABC S50A 589 004 DEF
QTC:  7020 CW 2009-05-30 0002 S50A 321/1 AA1ZZZ 1234 DL1ABC 005
END-OF-LOG:
`

	export := createCabrilloLog(settings, 123, []core.QSO{qso}, []core.QTC{qtc})
	Export(buffer, export)

	assert.Equal(t, expected, buffer.String())
}

type testSettings struct {
	stationCallsign   string
	stationOperator   string
	stationLocator    string
	contestIdentifier string
}

func (s *testSettings) Station() core.Station {
	loc, _ := locator.Parse(s.stationLocator)
	return core.Station{
		Callsign: core.MustParseCallsign(s.stationCallsign),
		Operator: core.MustParseCallsign(s.stationOperator),
		Locator:  loc,
	}
}

func (s *testSettings) Contest() core.Contest {
	return core.Contest{
		Definition: &conval.Definition{
			Identifier: conval.ContestIdentifier(s.contestIdentifier),
		},
	}
}
