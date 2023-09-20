package datasources

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-spatial/geom"
)

const nrOfFakeFeatures = 1000
const cursorColumnName = "cursor"

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

func (fdb FakeDB) GetFeatures(_ string, cursor string, limit int) (*domain.FeatureCollection, domain.Cursor) {
	var low int
	if cursor == "" {
		low = 0
	} else {
		low, _ = strconv.Atoi(cursor)
	}

	high := low + limit
	if high > len(fdb.featureCollection.Features) {
		high = len(fdb.featureCollection.Features)
	}

	page := fdb.featureCollection.Features[low:high]
	return &domain.FeatureCollection{
			Features: page,
		},
		domain.NewCursor(page, cursorColumnName)
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

			// we use an explicit cursor column in our fake data to keep things simple
			cursorColumnName: i,
		}

		feature := domain.Feature{}
		feature.ID = gofakeit.Numerify(fmt.Sprintf("%d#######", i))
		feature.Geometry.Geometry = geom.Point{address.Longitude, address.Latitude}
		feature.Properties = props

		feats = append(feats, &feature)
	}

	// the column that will be used as a cursor must be ordered
	sort.Slice(feats, func(i, j int) bool {
		return feats[i].Properties[cursorColumnName].(int) < feats[j].Properties[cursorColumnName].(int)
	})

	fc := domain.FeatureCollection{}
	fc.Features = feats
	return &fc
}
