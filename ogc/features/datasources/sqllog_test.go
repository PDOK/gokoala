package datasources

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLLog(t *testing.T) {
	logSQL = true
	t.Cleanup(func() { logSQL = false })

	var capturedLogOutput bytes.Buffer
	log.SetOutput(&capturedLogOutput)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	s := &SQLLog{}

	ctx, err := s.Before(context.Background(), "SELECT * FROM test WHERE id = ?", 123)
	assert.NoError(t, err)

	_, err = s.After(ctx, "SELECT * FROM test WHERE id = ?", 123)
	assert.NoError(t, err)

	assert.Contains(t, capturedLogOutput.String(), "SQL:\nSELECT * FROM test WHERE id = 123\n--- SQL query took")
}
