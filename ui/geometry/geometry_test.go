package geometry

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundtrip(t *testing.T) {
	initials := NewWindows()
	main := ID("main")
	initialMain := initials.Get(main)
	initialMain.SetPosition(1, 2)
	initialMain.SetSize(3, 4)
	initialMain.SetMaximized(true)
	dialog := ID("dialog")
	initialDialog := initials.Get(dialog)
	initialDialog.SetPosition(10, 20)
	initialDialog.SetSize(300, 400)
	initialDialog.SetMaximized(false)

	buffer := bytes.NewBuffer([]byte{})
	initials.Store(buffer)

	loaded, err := LoadWindows(buffer)
	require.NoError(t, err)

	assert.Equal(t, initialMain, loaded.Get(main))
	assert.Equal(t, initialDialog, loaded.Get(dialog))
}

func TestRoundtripWithFile(t *testing.T) {
	initials := NewWindows()
	main := ID("main")
	initialMain := initials.Get(main)
	initialMain.SetPosition(1, 2)
	initialMain.SetSize(3, 4)
	initialMain.SetMaximized(true)
	dialog := ID("dialog")
	initialDialog := initials.Get(dialog)
	initialDialog.SetPosition(10, 20)
	initialDialog.SetSize(300, 400)
	initialDialog.SetMaximized(false)

	writeFile, err := ioutil.TempFile("", "TestRoundtripWithFile")
	require.NoError(t, err)
	defer writeFile.Close()

	initials.Store(writeFile)

	readFile, err := os.Open(writeFile.Name())
	require.NoError(t, err)
	defer readFile.Close()

	loaded, err := LoadWindows(readFile)
	require.NoError(t, err)

	assert.Equal(t, initialMain, loaded.Get(main))
	assert.Equal(t, initialDialog, loaded.Get(dialog))

	assert.Fail(t, writeFile.Name())
}
