package store

import (
	"io/ioutil"
	"os"
	"testing"

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
