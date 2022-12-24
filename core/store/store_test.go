package store

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/locator"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_ReadV0File(t *testing.T) {
	fs := NewFileStore("testdata/v0.testlog")
	qsos, _, _, _, err := fs.ReadAll()
	require.NoError(t, err)

	assert.IsType(t, new(v0Format), fs.format)
	assert.Equal(t, 1, len(qsos))
	actual := qsos[0]
	assert.Equal(t, "DL2ABC", actual.Callsign.String())
}

func TestFileStore_EmptyFileUsesV0Format(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStore(tmpFile.Name())

	assert.IsType(t, new(v0Format), fs.format)
}

func TestFileStore_ClearUsesLatestFormat(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStore(tmpFile.Name())
	fs.format = new(unknownFormat)
	err = fs.Clear()

	assert.NoError(t, err)
	assert.IsType(t, new(latestFormat), fs.format)
}

func TestFileStore_V1QSORoundtrip(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := &FileStore{
		filename: tmpFile.Name(),
		format:   new(v1Format),
	}
	err = fs.format.Clear(&pbReadWriter{writer: tmpFile})
	require.NoError(t, err)

	qso := core.QSO{
		Callsign:      callsign.MustParse("DL1ABC"),
		Time:          time.Unix(123, 0),
		Frequency:     3535000,
		Band:          core.Band80m,
		Mode:          core.ModeCW,
		MyReport:      "599",
		MyNumber:      1,
		MyExchange:    []string{"599", "1", "mx"},
		TheirReport:   "579",
		TheirNumber:   2,
		TheirExchange: []string{"579", "2", "tx"},
		LogTimestamp:  time.Unix(456, 0),
	}
	err = fs.WriteQSO(qso)
	require.NoError(t, err)

	qsos, station, contest, keyer, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 1, len(qsos))
	assert.Equal(t, qso, qsos[0])
	assert.Nil(t, station)
	assert.Nil(t, contest)
	assert.Nil(t, keyer)
}

func TestFileStore_V1StationRoundtrip(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := &FileStore{
		filename: tmpFile.Name(),
		format:   new(v1Format),
	}
	err = fs.format.Clear(&pbReadWriter{writer: tmpFile})
	require.NoError(t, err)

	loc, err := locator.Parse("AA00AA")
	station1 := core.Station{
		Callsign: callsign.MustParse("DL0ABC"),
		Operator: callsign.MustParse("DL1ABC"),
		Locator:  loc,
	}
	station2 := core.Station{
		Callsign: callsign.MustParse("DL0ABC"),
		Operator: callsign.MustParse("DL2ABC"),
		Locator:  loc,
	}
	err = fs.WriteStation(station1)
	err = fs.WriteStation(station2)
	require.NoError(t, err)

	qsos, station, contest, keyer, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, qsos)
	assert.Equal(t, station2, *station)
	assert.Nil(t, contest)
	assert.Nil(t, keyer)
}

func TestFileStore_V1ContestRoundtrip(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := &FileStore{
		filename: tmpFile.Name(),
		format:   new(v1Format),
	}
	err = fs.format.Clear(&pbReadWriter{writer: tmpFile})
	require.NoError(t, err)

	contest1 := core.Contest{
		Name: "ONE",
	}
	contest2 := core.Contest{
		Name: "TWO",
	}
	err = fs.WriteContest(contest1)
	err = fs.WriteContest(contest2)
	require.NoError(t, err)

	qsos, station, contest, keyer, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, qsos)
	assert.Nil(t, station)
	assert.Equal(t, contest2, *contest)
	assert.Nil(t, keyer)
}

func TestFileStore_V1KeyerRoundtrip(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := &FileStore{
		filename: tmpFile.Name(),
		format:   new(v1Format),
	}
	err = fs.format.Clear(&pbReadWriter{writer: tmpFile})
	require.NoError(t, err)

	keyer1 := core.Keyer{
		WPM: 25,
	}
	keyer2 := core.Keyer{
		WPM: 35,
	}
	err = fs.WriteKeyer(keyer1)
	err = fs.WriteKeyer(keyer2)
	require.NoError(t, err)

	qsos, station, contest, keyer, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, qsos)
	assert.Nil(t, station)
	assert.Nil(t, contest)
	assert.Equal(t, keyer2, *keyer)
}
