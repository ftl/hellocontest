package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hellocontest/core"
)

// DXCCFinder returns a list of matching prefixes for the given string and indicates if there was a match at all.
type DXCCFinder interface {
	Find(string) (dxcc.Prefix, bool)
}

// Export writes the given QSOs to the given writer in the CSV format.
// The header is very limited and needs to be completed manually after the log was written.
func Export(w io.Writer, mycall callsign.Callsign, qsos ...core.QSO) error {
	csvWriter := csv.NewWriter(w)
	for _, qso := range qsos {
		if err := writeQSO(csvWriter, mycall, qso); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return nil
}

func writeQSO(w *csv.Writer, mycall callsign.Callsign, qso core.QSO) error {
	myCallIndex := 5
	theirCallIndex := myCallIndex + 1 + len(qso.MyExchange)
	dxccPrefixIndex := theirCallIndex + 1 + len(qso.TheirExchange)
	values := make([]string, 13+len(qso.MyExchange)+len(qso.TheirExchange))
	values[0] = qso.Band.String()
	values[1] = fmt.Sprintf("%5.3f", float64(qso.Frequency/1000000.0))
	values[2] = qso.Mode.String()
	values[3] = qso.Time.In(time.UTC).Format("2006-01-02")
	values[4] = qso.Time.In(time.UTC).Format("1504")
	values[5] = mycall.String()
	for i, value := range qso.MyExchange {
		values[myCallIndex+1+i] = value
	}
	values[theirCallIndex] = qso.Callsign.String()
	for i, value := range qso.TheirExchange {
		values[theirCallIndex+1+i] = value
	}
	values[dxccPrefixIndex] = qso.DXCC.PrimaryPrefix
	values[dxccPrefixIndex+1] = qso.DXCC.Continent
	values[dxccPrefixIndex+2] = fmt.Sprintf("%d", qso.DXCC.ITUZone)
	values[dxccPrefixIndex+3] = fmt.Sprintf("%d", qso.DXCC.CQZone)
	values[dxccPrefixIndex+4] = strconv.Itoa(qso.Points)
	values[dxccPrefixIndex+5] = ""

	return w.Write(values)
}
