package adif

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestData(t *testing.T) {
	testCases := []struct {
		field    string
		datatype string
		data     string
		expected string
	}{
		{"CALL", "", "DB0ABC", "<CALL:6>DB0ABC"},
		{"CALL", "datatype", "DB0ABC", "<CALL:6:datatype>DB0ABC"},
	}
	for i, tC := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			buffer := bytes.NewBuffer([]byte{})
			err := data(buffer, tC.field, tC.datatype, tC.data)
			require.NoError(t, err)
			assert.Equal(t, tC.expected, buffer.String())
		})
	}
}

func TestRecord(t *testing.T) {
	theirCall, _ := callsign.Parse("S50A")
	testCases := []struct {
		desc     string
		qso      core.QSO
		expected string
	}{
		{
			desc: "80m RTTY with Frequency",
			qso: core.QSO{
				Callsign:      theirCall,
				Time:          time.Date(2009, time.May, 30, 0, 2, 0, 0, time.UTC),
				Frequency:     3550000,
				Band:          core.Band80m,
				Mode:          core.ModeRTTY,
				MyReport:      core.RST("599"),
				MyNumber:      core.QSONumber(1),
				MyExchange:    []string{"599", "001", "ABC"},
				TheirReport:   core.RST("589"),
				TheirNumber:   core.QSONumber(4),
				TheirExchange: []string{"589", "004", "DEF"},
			},
			expected: "<QSO_DATE:8>20090530<TIME_ON:4>0002<TIME_OFF:4>0002<CALL:4>S50A<FREQ:5>3.550<BAND:3>80m<MODE:4>RTTY<RST_SENT:3>599<RST_RCVD:3>589<COMMENT:23>599 001 ABC 589 004 DEF<EOR>\n",
		},
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
			expected: "<QSO_DATE:8>20090530<TIME_ON:4>0002<TIME_OFF:4>0002<CALL:4>S50A<FREQ:5>7.000<BAND:3>40m<MODE:2>CW<RST_SENT:3>599<RST_RCVD:3>589<COMMENT:23>599 001 ABC 589 004 DEF<EOR>\n",
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
			expected: "<QSO_DATE:8>20090530<TIME_ON:4>0002<TIME_OFF:4>0002<CALL:4>S50A<FREQ:6>14.000<BAND:3>20m<MODE:3>SSB<RST_SENT:2>59<RST_RCVD:2>58<COMMENT:21>59 001 XXX 58 004 YYY<EOR>\n",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			buffer := bytes.NewBuffer([]byte{})
			err := record(buffer, tC.qso)
			require.NoError(t, err)
			assert.Equal(t, tC.expected, buffer.String())
		})
	}
}

func TestExport(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})
	theirCall, _ := callsign.Parse("S50A")
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

	expected := `Generated by Hello Contest
<adif_ver:5>3.0.9
<programid:11>HelloContest
<EOH>
<QSO_DATE:8>20090530<TIME_ON:4>0002<TIME_OFF:4>0002<CALL:4>S50A<FREQ:5>7.000<BAND:3>40m<MODE:2>CW<RST_SENT:3>599<RST_RCVD:3>589<COMMENT:23>599 001 ABC 589 004 DEF<EOR>
`

	Export(buffer, qso)

	assert.Equal(t, expected, buffer.String())
}
