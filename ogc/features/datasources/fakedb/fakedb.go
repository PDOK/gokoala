package fakedb

import (
	"sort"

	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-spatial/geom"
)

const nrOfFakeFeatures = 10000

// FakeDB fake/mock datasource used for prototyping/testing/demos/etc.
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

func (fdb FakeDB) GetFeatures(_ string, cursor int64, limit int) (*domain.FeatureCollection, domain.Cursor, error) {
	low := cursor
	high := low + int64(limit)

	last := high > int64(len(fdb.featureCollection.Features))
	if last {
		high = int64(len(fdb.featureCollection.Features))
	}
	if high < 0 {
		high = 0
	}

	page := fdb.featureCollection.Features[low:high]
	return &domain.FeatureCollection{
			NumberReturned: len(page),
			Features:       page,
		},
		domain.NewCursor(page, limit, last),
		nil
}

func (fdb FakeDB) GetFeature(_ string, featureID int64) (*domain.Feature, error) {
	for _, feat := range fdb.featureCollection.Features {
		if feat.ID == featureID {
			return feat, nil
		}
	}
	return nil, nil //nolint:nilnil
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

		feature := domain.Feature{}
		feature.ID = int64(i)
		feature.Geometry.Geometry = geom.Point{address.Longitude, address.Latitude}
		feature.Properties = props

		feats = append(feats, &feature)
	}

	// the collection must be ordered by the cursor column
	sort.Slice(feats, func(i, j int) bool {
		return feats[i].ID < feats[j].ID
	})

	fc := domain.FeatureCollection{}
	fc.Features = feats
	return &fc
}
