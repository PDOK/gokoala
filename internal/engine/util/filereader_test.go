package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "Test read gzip file",
			filePath: "../testdata/readfile-gzipped.txt",
			wantErr:  false,
		},
		{
			name:     "Test read plain file",
			filePath: "../testdata/readfile-plain.txt",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadFile(tt.filePath)
			assert.Equal(t, "foobar", got)
		})
	}
}
