package store

import (
	"os"
	"testing"
	"time"

	"github.com/ftl/hamradio/locator"
	"github.com/ftl/hellocontest/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_ReadV0File(t *testing.T) {
	fs := NewFileStore("testdata/v0.testlog")
	qsos, _, _, _, _, err := fs.ReadAll()
	require.NoError(t, err)

	assert.IsType(t, new(v0Format), fs.format)
	assert.Equal(t, 1, len(qsos))
	actual := qsos[0]
	assert.Equal(t, "DL2ABC", actual.Callsign.String())
}

func TestFileStore_EmptyFileUsesV0Format(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStore(tmpFile.Name())

	assert.IsType(t, new(v0Format), fs.format)
}

func TestFileStore_ClearUsesLatestFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStore(tmpFile.Name())
	fs.format = new(unknownFormat)
	err = fs.Clear()

	assert.NoError(t, err)
	assert.IsType(t, new(latestFormat), fs.format)
}

func TestFileStore_V1QSORoundtrip(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := &FileStore{
		filename: tmpFile.Name(),
		format:   new(v1Format),
	}
	err = fs.format.Clear(&pbReadWriter{writer: tmpFile})
	require.NoError(t, err)

	qso := model.QSO{
		Callsign:      model.MustParseCallsign("DL1ABC"),
		Time:          time.Unix(123, 0),
		Frequency:     3535000,
		Band:          model.Band80m,
		Mode:          model.ModeCW,
		MyReport:      "599",
		MyNumber:      1,
		MyExchange:    []string{"599", "001", "mx"},
		TheirReport:   "579",
		TheirNumber:   2,
		TheirExchange: []string{"579", "002", "tx"},
		LogTimestamp:  time.Unix(456, 0),
	}
	err = fs.WriteQSO(qso)
	require.NoError(t, err)

	qsos, station, contest, keyer, qtcs, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 1, len(qsos))
	assert.Equal(t, qso, qsos[0])
	assert.Nil(t, station)
	assert.Nil(t, contest)
	assert.Nil(t, keyer)
	assert.Empty(t, qtcs)
}

func TestFileStore_V1StationRoundtrip(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := &FileStore{
		filename: tmpFile.Name(),
		format:   new(v1Format),
	}
	err = fs.format.Clear(&pbReadWriter{writer: tmpFile})
	require.NoError(t, err)

	loc, err := locator.Parse("AA00AA")
	station1 := model.Station{
		Callsign: model.MustParseCallsign("DL0ABC"),
		Operator: model.MustParseCallsign("DL1ABC"),
		Locator:  loc,
		Bandplan: model.DefaultBandplanID,
	}
	station2 := model.Station{
		Callsign: model.MustParseCallsign("DL0ABC"),
		Operator: model.MustParseCallsign("DL2ABC"),
		Locator:  loc,
		Bandplan: model.DefaultBandplanID,
	}
	err = fs.WriteStation(station1)
	err = fs.WriteStation(station2)
	require.NoError(t, err)

	qsos, station, contest, keyer, qtcs, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, qsos)
	assert.Equal(t, station2, *station)
	assert.Nil(t, contest)
	assert.Nil(t, keyer)
	assert.Empty(t, qtcs)
}

func TestFileStore_V1ContestRoundtrip(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := &FileStore{
		filename: tmpFile.Name(),
		format:   new(v1Format),
	}
	err = fs.format.Clear(&pbReadWriter{writer: tmpFile})
	require.NoError(t, err)

	contest1 := model.Contest{
		Name: "ONE",
	}
	contest2 := model.Contest{
		Name: "TWO",
	}
	err = fs.WriteContest(contest1)
	err = fs.WriteContest(contest2)
	require.NoError(t, err)

	qsos, station, contest, keyer, qtcs, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, qsos)
	assert.Nil(t, station)
	assert.Equal(t, contest2, *contest)
	assert.Nil(t, keyer)
	assert.Empty(t, qtcs)
}

func TestFileStore_V1KeyerRoundtrip(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := &FileStore{
		filename: tmpFile.Name(),
		format:   new(v1Format),
	}
	err = fs.format.Clear(&pbReadWriter{writer: tmpFile})
	require.NoError(t, err)

	keyer1 := model.KeyerSettings{
		WPM: 25,
	}
	keyer2 := model.KeyerSettings{
		WPM: 35,
	}
	err = fs.WriteKeyer(keyer1)
	err = fs.WriteKeyer(keyer2)
	require.NoError(t, err)

	qsos, station, contest, keyer, qtcs, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, qsos)
	assert.Nil(t, station)
	assert.Nil(t, contest)
	assert.Equal(t, keyer2, *keyer)
	assert.Empty(t, qtcs)
}

func TestFileStore_V1QTCRoundtrip(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := &FileStore{
		filename: tmpFile.Name(),
		format:   new(v1Format),
	}
	err = fs.format.Clear(&pbReadWriter{writer: tmpFile})
	require.NoError(t, err)

	qtc := model.QTC{
		Timestamp: time.Unix(123, 0),
		Frequency: 3535000,
		Band:      model.Band80m,
		Mode:      model.ModeCW,

		Kind:          model.SentQTC,
		QSONumber:     45,
		TheirCallsign: model.MustParseCallsign("DL1ABC"),
		Header:        model.QTCHeader{SeriesNumber: 7, QTCCount: 10},

		QTCTime:     model.QTCTime{Hour: 12, Minute: 34},
		QTCCallsign: model.MustParseCallsign("DL2ABC"),
		QTCNumber:   56,
	}
	err = fs.WriteQTC(qtc)
	require.NoError(t, err)

	qsos, station, contest, keyer, qtcs, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, qsos)
	assert.Nil(t, station)
	assert.Nil(t, contest)
	assert.Nil(t, keyer)
	assert.Equal(t, 1, len(qtcs))
	assert.Equal(t, qtc, qtcs[0])
}
