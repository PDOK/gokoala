package postgres

import (
	"context"
	"fmt"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxgeom "github.com/twpayne/pgx-geom"
	pgxuuid "github.com/vgarvardt/pgx-google-uuid/v5"
)

var postgresExtensions = []string{"postgis", "unaccent"}

// InitConnectionPool initializes a connection pool for the given connection string and runs setup queries.
func InitConnectionPool(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	pool, err := newReadOnlyConnectionPool(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	// only for setup purposes, we want client requests to use the read-only connection pool!
	conn, err := newWriteableConnection(ctx, connectionString)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	// create extensions if they don't exist
	for _, ext := range postgresExtensions {
		_, err = conn.Exec(ctx, `create extension if not exists `+ext+`;`)
		if err != nil {
			return nil, fmt.Errorf("error creating %s extension: %w", ext, err)
		}
	}

	// create collations (to support ACCENTI/CASEI cql operators) if they don't exist
	collations := map[string]string{
		common.IgnoreCaseCollation:          "und-u-ks-level2",
		common.IgnoreAccentCollation:        "und-u-ks-level1-kc-true",
		common.IgnoreAccentAndCaseCollation: "und-u-ks-level1",
	}
	for collation, locale := range collations {
		_, err = conn.Exec(ctx, fmt.Sprintf(
			`create collation if not exists %s (provider = icu, locale = '%s', deterministic = false);`,
			collation, locale))
		if err != nil {
			return nil, err
		}
	}

	return pool, nil
}

// newReadOnlyConnectionPool creates a connection pool for the given connection string with read-only connections.
func newReadOnlyConnectionPool(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// enable SQL logging when appropriate environment variable (LOG_SQL=true) is set
	if sl := NewSQLLogFromEnv(); sl.LogSQL {
		pgxConfig.ConnConfig.Tracer = sl.Tracer
	}

	// set connection to read-only for safety since we (should) never write to Postgres.
	pgxConfig.ConnConfig.RuntimeParams["default_transaction_read_only"] = "on"

	pgxConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// add support for github.com/google/uuid <-> PostGIS conversions
		pgxuuid.Register(conn.TypeMap())
		// add support for Go <-> PostGIS conversions
		return pgxgeom.Register(ctx, conn)
	}

	return pgxpool.NewWithConfig(ctx, pgxConfig)
}

// newWriteableConnection only use this for setup purposes!
func newWriteableConnection(ctx context.Context, connectionString string) (*pgx.Conn, error) {
	pgxConfig, err := pgx.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// enable SQL logging when appropriate environment variable (LOG_SQL=true) is set
	if sl := NewSQLLogFromEnv(); sl.LogSQL {
		pgxConfig.Tracer = sl.Tracer
	}

	return pgx.ConnectConfig(ctx, pgxConfig)
}
