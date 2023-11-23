package features

import (
	"net/url"
	"testing"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom"
	"github.com/stretchr/testify/assert"
)

func Test_featureCollectionURL_parseParams(t *testing.T) {
	type fields struct {
		baseURL url.URL
		params  url.Values
		limit   engine.Limit
	}
	host, _ := url.Parse("http://ogc.example")
	tests := []struct {
		name              string
		fields            fields
		wantEncodedCursor domain.EncodedCursor
		wantLimit         int
		wantCrs           int
		wantBbox          *geom.Extent
		wantBboxCrs       int
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "Parse no params",
			fields: fields{
				baseURL: *host,
				params:  url.Values{},
				limit: engine.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantEncodedCursor: "",
			wantLimit:         10,
			wantCrs:           4326,
			wantBbox:          nil,
			wantBboxCrs:       4326,
			wantErr:           success(),
		},
		{
			name: "Parse many params",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor":   []string{"H3w%3D"},
					"crs":      []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox":     []string{"1,2,3,4"},
					"bbox-crs": []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"limit":    []string{"10000"},
				},
				limit: engine.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantEncodedCursor: "H3w%3D",
			wantLimit:         20, // use max instead of supplied limit
			wantCrs:           28992,
			wantBbox:          (*geom.Extent)([]float64{1, 2, 3, 4}),
			wantBboxCrs:       28992,
			wantErr:           success(),
		},
		{
			name: "Fail on wrong crs",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"crs":      []string{"EPSG:28992"},
					"bbox-crs": []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
				},
				limit: engine.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equalf(t, "crs param should start with http://www.opengis.net/def/crs/, got: EPSG:28992", err.Error(), "parseParams()")
				return false
			},
		},
		{
			name: "Fail on wrong bbox",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"bbox": []string{"1,2,3,4,5,6"},
				},
				limit: engine.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equalf(t, "bbox should contain exactly 4 values separated by commas: minx,miny,maxx,maxy", err.Error(), "parseParams()")
				return false
			},
		},
		{
			name: "Fail on wrong limit",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"limit": []string{"-200"},
				},
				limit: engine.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equalf(t, "limit can't be negative", err.Error(), "parseParams()")
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := featureCollectionURL{
				baseURL: tt.fields.baseURL,
				params:  tt.fields.params,
				limit:   tt.fields.limit,
			}
			gotEncodedCursor, gotLimit, gotCrs, gotBbox, gotBboxCrs, err := fc.parseParams()
			if !tt.wantErr(t, err, "parseParams()") {
				return
			}
			assert.Equalf(t, tt.wantEncodedCursor, gotEncodedCursor, "parseParams()")
			assert.Equalf(t, tt.wantLimit, gotLimit, "parseParams()")
			assert.Equalf(t, tt.wantCrs, gotCrs, "parseParams()")
			assert.Equalf(t, tt.wantBbox, gotBbox, "parseParams()")
			assert.Equalf(t, tt.wantBboxCrs, gotBboxCrs, "parseParams()")
		})
	}
}

func success() func(t assert.TestingT, err error, i ...interface{}) bool {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		return true
	}
}
