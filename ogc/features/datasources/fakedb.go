package datasources

import (
	"strconv"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

const nrOfFakeFeatures = 1500

type FakeDB struct {
	featureCollection *geojson.FeatureCollection
}

func NewFakeDB() *FakeDB {
	return &FakeDB{
		featureCollection: generateFakeFeatureCollection(),
	}
}

func (FakeDB) Close() {
	// noop
}

func (fdb FakeDB) GetFeatures(_ string) *geojson.FeatureCollection {
	return fdb.featureCollection
}

func (fdb FakeDB) GetFeature(_ string, featureID string) *geojson.Feature {
	fid, _ := strconv.Atoi(featureID)
	return fdb.featureCollection.Features[fid]
}

func generateFakeFeatureCollection() *geojson.FeatureCollection {
	var features []*geojson.Feature
	for i := 0; i < nrOfFakeFeatures; i++ {
		var props = map[string]interface{}{
			"streetnane": gofakeit.Address().Street,
			"city":       gofakeit.Address().City,
			"year":       gofakeit.Year(),
			"floorsize":  gofakeit.Number(10, 300),
			"purpose":    gofakeit.BuzzWord(),
		}

		geom := orb.LineString{{1, 2}, {3, 4}, {5, 6}}
		feature := geojson.NewFeature(geom)
		feature.ID = i
		feature.Properties = props

		features = append(features, feature)
	}
	fc := geojson.NewFeatureCollection()
	fc.Features = features
	return fc
}
