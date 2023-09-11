package datasources

import "github.com/paulmach/orb/geojson"

type FakeDB struct {
}

func NewFakeDB() *FakeDB {
	return &FakeDB{}
}

func (FakeDB) GetFeatures(collection string) geojson.FeatureCollection {
	panic("implement me")
}

func (FakeDB) GetFeature(collection string, featureID string) geojson.Feature {
	panic("implement me")
}
