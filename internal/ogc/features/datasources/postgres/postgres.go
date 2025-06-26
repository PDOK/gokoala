package postgres

import (
	"context"
	"log"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

// Postgres !!! Placeholder implementation, for future reference !!!
type Postgres struct {
}

func NewPostgres() *Postgres {
	return &Postgres{}
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
