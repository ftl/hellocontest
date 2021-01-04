package store

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_ReadV0File(t *testing.T) {
	fs := NewFileStore("testdata/v0.testlog")
	qsos, err := fs.ReadAllQSOs()
	require.NoError(t, err)

	assert.IsType(t, new(v0Format), fs.format)
	assert.Equal(t, 1, len(qsos))
	actual := qsos[0]
	assert.Equal(t, "DL2ABC", actual.Callsign.String())
}

func TestFileStore_EmptyFileUsesLatestFormat(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), t.Name())
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStore(tmpFile.Name())

	assert.IsType(t, new(latestFormat), fs.format)
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
		Callsign:     callsign.MustParse("DL1ABC"),
		Time:         time.Unix(123, 0),
		Frequency:    3535000,
		Band:         core.Band80m,
		Mode:         core.ModeCW,
		MyReport:     "599",
		MyNumber:     1,
		MyXchange:    "mx",
		TheirReport:  "579",
		TheirNumber:  2,
		TheirXchange: "tx",
		LogTimestamp: time.Unix(456, 0),
	}
	err = fs.WriteQSO(qso)
	require.NoError(t, err)

	qsos, _, _, err := fs.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 1, len(qsos))
	assert.Equal(t, qso, qsos[0])
}
