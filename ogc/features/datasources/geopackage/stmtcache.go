package geopackage

import (
	"context"
	"log"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/jmoiron/sqlx"
)

var preparedStmtCacheSize = 15

// PreparedStatementCache is thread safe
type PreparedStatementCache struct {
	cache *lru.Cache[string, *sqlx.NamedStmt]
}

// NewCache creates a new PreparedStatementCache that will evict least-recently used (LRU) statements.
func NewCache() *PreparedStatementCache {
	cache, _ := lru.NewWithEvict[string, *sqlx.NamedStmt](preparedStmtCacheSize,
		func(_ string, stmt *sqlx.NamedStmt) {
			if stmt != nil {
				_ = stmt.Close()
			}
		})

	return &PreparedStatementCache{cache: cache}
}

// Lookup gets a prepared statement from the cache for the given query, or creates a new one and adds it to the cache
func (c *PreparedStatementCache) Lookup(ctx context.Context, db *sqlx.DB, query string) (*sqlx.NamedStmt, error) {
	cachedStmt, ok := c.cache.Get(query)
	if !ok {
		stmt, err := db.PrepareNamedContext(ctx, query)
		if err != nil {
			return nil, err
		}
		c.cache.Add(query, stmt)
		return stmt, nil
	}
	return cachedStmt, nil
}

// Close purges the cache, and closes remaining prepared statements
func (c *PreparedStatementCache) Close() {
	log.Printf("closing %d prepared statements", c.cache.Len())
	c.cache.Purge()
}
