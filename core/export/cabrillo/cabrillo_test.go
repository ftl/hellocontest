package cabrillo

import (
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
)

func TestQsoLine(t *testing.T) {
	myCall, _ := callsign.Parse("AA1ZZZ")
	theirCall, _ := callsign.Parse("S50A")
	testCases := []struct {
		desc          string
		qso           core.QSO
		myExchange    core.Exchanger
		theirExchange core.Exchanger
		expected      string
	}{
		{
			desc: "40m CW MyNumber TheirXchange",
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
			myExchange:    core.MyNumber,
			theirExchange: core.TheirXchange,
			expected:      "QSO: 7000 CW 2009-05-30 0002 AA1ZZZ 599 001 S50A 589 DEF",
		},
		{
			desc: "20m SSB MyXchange TheirNumber",
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
			myExchange:    core.MyXchange,
			theirExchange: core.TheirNumber,
			expected:      "QSO: 14000 PH 2009-05-30 0002 AA1ZZZ 59 XXX S50A 58 004",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual := qsoLine(myCall, tC.myExchange, tC.theirExchange, tC.qso)
			assert.Equal(t, tC.expected, actual)
		})
	}
}
