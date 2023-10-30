package postgis

import (
	"context"
	"log"

	"github.com/PDOK/gokoala/ogc/features/datasources"
	"github.com/PDOK/gokoala/ogc/features/domain"
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

func (pg PostGIS) GetFeatures(_ context.Context, _ string, _ datasources.FeatureOptions) (*domain.FeatureCollection, domain.Cursor, error) {
	log.Fatal("PostGIS support is not implemented yet, this just serves to demonstrate that we can support multiple datastores")
	return &domain.FeatureCollection{},
		domain.Cursor{},
		nil
}

func (pg PostGIS) GetFeature(_ context.Context, _ string, _ int64) (*domain.Feature, error) {
	log.Fatal("PostGIS support is not implemented yet, this just serves to demonstrate that we can support multiple datastores")
	return nil, nil //nolint:nilnil
}
