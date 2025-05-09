package postgis

import (
	neturl "net/url"
	"testing"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/stretchr/testify/assert"
)

// PostGIS !!! Placeholder implementation, for future reference !!!
func TestPostGIS(t *testing.T) {
	pg := PostGIS{}
	url, _ := neturl.Parse("http://example.com")
	p := domain.NewProfile(domain.RelAsLink, *url, []string{})

	t.Run("GetFeatureIDs", func(t *testing.T) {
		ids, cursors, err := pg.GetFeatureIDs(t.Context(), "", datasources.FeaturesCriteria{})
		assert.NoError(t, err)
		assert.Empty(t, ids)
		assert.NotNil(t, cursors)
	})

	t.Run("GetFeaturesByID", func(t *testing.T) {
		fc, err := pg.GetFeaturesByID(t.Context(), "", nil, p)
		assert.NoError(t, err)
		assert.NotNil(t, fc)
	})

	t.Run("GetFeatures", func(t *testing.T) {
		fc, cursors, err := pg.GetFeatures(t.Context(), "", datasources.FeaturesCriteria{}, p)
		assert.NoError(t, err)
		assert.Nil(t, fc)
		assert.NotNil(t, cursors)
	})

	t.Run("GetFeature", func(t *testing.T) {
		f, err := pg.GetFeature(t.Context(), "", 0, p)
		assert.NoError(t, err)
		assert.Nil(t, f)
	})

	t.Run("GetSchema", func(t *testing.T) {
		schema, err := pg.GetSchema("")
		assert.NoError(t, err)
		assert.Nil(t, schema)
	})
}
