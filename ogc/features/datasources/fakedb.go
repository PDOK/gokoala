package datasources

import (
	"fmt"

	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-spatial/geom"
)

const nrOfFakeFeatures = 100

type FakeDB struct {
	featureCollection *domain.FeatureCollection
}

func NewFakeDB() *FakeDB {
	return &FakeDB{
		featureCollection: generateFakeFeatureCollection(),
	}
}

func (FakeDB) Close() {
	// noop
}

func (fdb FakeDB) GetFeatures(_ string) *domain.FeatureCollection {
	return fdb.featureCollection
}

func (fdb FakeDB) GetFeature(_ string, featureID string) *domain.Feature {
	for _, feat := range fdb.featureCollection.Features {
		if feat.ID == featureID {
			return feat
		}
	}
	return nil
}

func generateFakeFeatureCollection() *domain.FeatureCollection {
	var feats []*domain.Feature
	for i := 0; i < nrOfFakeFeatures; i++ {
		address := gofakeit.Address()
		var props = map[string]interface{}{
			"streetname": address.Street,
			"city":       address.City,
			"year":       gofakeit.Year(),
			"floorsize":  gofakeit.Number(10, 300),
			"purpose":    gofakeit.Blurb(),
		}

		geometry := geom.Point{address.Longitude, address.Latitude}
		feature := domain.Feature{}
		feature.ID = gofakeit.Numerify(fmt.Sprintf("%d#######", i))
		feature.Geometry.Geometry = geometry
		feature.Properties = props

		feats = append(feats, &feature)
	}
	fc := domain.FeatureCollection{}
	fc.Features = feats
	return &fc
}
