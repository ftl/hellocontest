package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_ReadV0(t *testing.T) {
	fs := NewFileStore("testdata/v0.testlog")
	qsos, err := fs.ReadAllQSOs()
	require.NoError(t, err)

	assert.Equal(t, 1, len(qsos))
	actual := qsos[0]
	assert.Equal(t, "DL2ABC", actual.Callsign.String())
}
