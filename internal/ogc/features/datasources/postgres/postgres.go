package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxgeom "github.com/twpayne/pgx-geom"
)

type Postgres struct {
	db           *pgxpool.Pool
	queryTimeout time.Duration
}

func NewPostgres(_ config.GeoSpatialCollections, postgresConfig config.Postgres, transformOnTheFly bool) (*Postgres, error) {
	if !transformOnTheFly {
		return nil, errors.New("ahead-of-time reprojected features are currently not supported for PostgreSQL, " +
			"reprojection/transformation is always applied")
	}

	pgxConfig, err := pgxpool.ParseConfig(postgresConfig.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}
	// add support for Go <-> PostGIS conversions
	pgxConfig.AfterConnect = pgxgeom.Register

	db, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &Postgres{
		db,
		postgresConfig.QueryTimeout.Duration,
	}, nil
}

func (Postgres) Close() {
	// noop
}

func (pg Postgres) GetFeatureIDs(_ context.Context, _ string, _ datasources.FeaturesCriteria) ([]int64, domain.Cursors, error) {
	log.Println("Postgres support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return []int64{}, domain.Cursors{}, nil
}

func (pg Postgres) GetFeaturesByID(_ context.Context, _ string, _ []int64, _ domain.AxisOrder, _ domain.Profile) (*domain.FeatureCollection, error) {
	log.Println("Postgres support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return &domain.FeatureCollection{}, nil
}

func (pg Postgres) GetFeatures(_ context.Context, _ string, _ datasources.FeaturesCriteria, _ domain.AxisOrder, _ domain.Profile) (*domain.FeatureCollection, domain.Cursors, error) {
	log.Println("Postgres support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return nil, domain.Cursors{}, nil
}

func (pg Postgres) GetFeature(_ context.Context, _ string, _ any, _ domain.AxisOrder, _ domain.Profile) (*domain.Feature, error) {
	log.Println("Postgres support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return nil, nil
}

func (pg Postgres) GetSchema(_ string) (*domain.Schema, error) {
	log.Println("Postgres support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return nil, nil
}

func (pg Postgres) GetPropertyFiltersWithAllowedValues(_ string) datasources.PropertyFiltersWithAllowedValues {
	log.Println("Postgres support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return nil
}
