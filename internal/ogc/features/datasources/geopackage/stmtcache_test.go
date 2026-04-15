package geopackage

import (
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreparedStatementCache(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "First query is a cache miss",
			query: "SELECT * FROM main.sqlite_master WHERE name = :n",
		},
		{
			name:  "Second query is a cache hit",
			query: "SELECT * FROM main.sqlite_master WHERE name = :n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCache()
			assert.NotNil(t, c)

			db, err := sqlx.Connect("sqlite3", ":memory:")
			require.NoError(t, err)

			stmt, err := c.Lookup(t.Context(), db, tt.query)
			require.NoError(t, err)
			assert.NotNil(t, stmt)

			c.Close()
		})
	}

	t.Run("Concurrent access to the cache", func(t *testing.T) {
		var wg sync.WaitGroup

		c := NewCache()
		assert.NotNil(t, c)

		db, err := sqlx.Connect("sqlite3", ":memory:")
		require.NoError(t, err)

		// Run multiple goroutines that will access the cache concurrently.
		for range 25 {
			wg.Go(func() {
				stmt1, err := c.Lookup(t.Context(), db, "SELECT * FROM main.sqlite_master WHERE name = :n")
				require.NoError(t, err)
				assert.NotNil(t, stmt1)

				stmt2, err := c.Lookup(t.Context(), db, "SELECT * FROM main.sqlite_master WHERE type = :t")
				require.NoError(t, err)
				assert.NotNil(t, stmt2)
			})
		}
		wg.Wait() // Wait for all goroutines to finish.

		assert.Equal(t, 2, c.cache.Len())
		c.Close()
	})
}
