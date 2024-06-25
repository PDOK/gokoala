package postgis

import (
	"context"
	"log"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

// PostGIS !!! Placeholder implementation, for future reference !!!
type PostGIS struct {
}

func NewPostGIS() *PostGIS {
	return &PostGIS{}
}

func (PostGIS) Close() {
	// noop
}

func (pg PostGIS) GetFeatureIDs(_ context.Context, _ string, _ datasources.FeaturesCriteria) ([]int64, domain.Cursors, error) {
	log.Println("PostGIS support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return []int64{}, domain.Cursors{}, nil
}

func (pg PostGIS) GetFeaturesByID(_ context.Context, _ string, _ []int64) (*domain.FeatureCollection, error) {
	log.Println("PostGIS support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return &domain.FeatureCollection{}, nil
}

func (pg PostGIS) GetFeatures(_ context.Context, _ string, _ datasources.FeaturesCriteria) (*domain.FeatureCollection, domain.Cursors, error) {
	log.Println("PostGIS support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return nil, domain.Cursors{}, nil
}

func (pg PostGIS) GetFeature(_ context.Context, _ string, _ any) (*domain.Feature, error) {
	log.Println("PostGIS support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return nil, nil
}

func (pg PostGIS) GetFeatureTableMetadata(_ string) (datasources.FeatureTableMetadata, error) {
	log.Println("PostGIS support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return nil, nil
}

func (pg PostGIS) GetPropertyFiltersWithAllowedValues(_ string) datasources.PropertyFiltersWithAllowedValues {
	log.Println("PostGIS support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return nil
}
