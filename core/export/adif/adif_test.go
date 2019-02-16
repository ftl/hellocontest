package adif

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			expected: "<QSO_DATE:8>20090530<TIME_ON:4>0002<TIME_OFF:4>0002<CALL:4>S50A<BAND:3>40m<MODE:2>CW<RST_SENT:3>599<RST_RCVD:3>589<COMMENT:15>001 ABC 004 DEF<EOR>\n",
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
			expected: "<QSO_DATE:8>20090530<TIME_ON:4>0002<TIME_OFF:4>0002<CALL:4>S50A<BAND:3>20m<MODE:3>SSB<RST_SENT:2>59<RST_RCVD:2>58<COMMENT:15>001 XXX 004 YYY<EOR>\n",
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

	expected := `Generated by Hello Contest
<adif_ver:5>3.0.9
<programid:11>HelloContest
<EOH>
<QSO_DATE:8>20090530<TIME_ON:4>0002<TIME_OFF:4>0002<CALL:4>S50A<BAND:3>40m<MODE:2>CW<RST_SENT:3>599<RST_RCVD:3>589<COMMENT:15>001 ABC 004 DEF<EOR>
`

	Export(buffer, qso)

	assert.Equal(t, expected, buffer.String())
}