package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
TEST LIST:

startup
- should start with a default filename
- should start with an empty log
- should update the log view
- should connect the entry view to the new log

new
- should ask for a filename
- should overwrite, if the file already exists
- should update the log view
- should connect the entry view to the new log

open
- should ask for a filename
- should load the log data from file
- should update the log view
- should append new QSOs to the selected file
- should connect the entry view to the loaded log

save as
- should ask for a filename
- should clear the new file
- should overwrite, if the file already exists
- should write all existing QSOs from the log to the file
- should append new QSOs to the selected file

*/

func TestMarkerNumber(t *testing.T) {
	tt := []struct {
		desc    string
		params  map[string]string
		want    int
		wantErr bool
	}{
		{desc: "a number", params: map[string]string{"number": "3"}, want: 3},
		{desc: "without the parameter", params: map[string]string{}, wantErr: true},
		{desc: "an empty value", params: map[string]string{"number": ""}, wantErr: true},
		{desc: "no number", params: map[string]string{"number": "nonsense"}, wantErr: true},
		{desc: "zero", params: map[string]string{"number": "0"}, wantErr: true},
		{desc: "a negative number", params: map[string]string{"number": "-1"}, wantErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			number, err := markerNumber(tc.params)

			if tc.wantErr {
				assert.Error(t, err, "a marker number starts with 1")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, number)
		})
	}
}
