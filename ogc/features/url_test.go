package features

import (
	"net/url"
	"testing"
	"time"

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
		wantOutputCrs     int
		wantBbox          *geom.Extent
		wantInputCrs      int
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
			wantOutputCrs:     100000,
			wantBbox:          nil,
			wantInputCrs:      100000,
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
			wantOutputCrs:     28992,
			wantBbox:          (*geom.Extent)([]float64{1, 2, 3, 4}),
			wantInputCrs:      28992,
			wantErr:           success(),
		},
		{
			name: "Parse input crs specified, no output crs specified",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor":   []string{"H3w%3D"},
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
			wantOutputCrs:     100000,
			wantBbox:          (*geom.Extent)([]float64{1, 2, 3, 4}),
			wantInputCrs:      28992,
			wantErr:           success(),
		},
		{
			name: "Parse no input crs specified, output crs specified",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor": []string{"H3w%3D"},
					"crs":    []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox":   []string{"1,2,3,4"},
					"limit":  []string{"10000"},
				},
				limit: engine.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantEncodedCursor: "H3w%3D",
			wantLimit:         20, // use max instead of supplied limit
			wantOutputCrs:     28992,
			wantBbox:          (*geom.Extent)([]float64{1, 2, 3, 4}),
			wantInputCrs:      100000,
			wantErr:           success(),
		},
		{
			name: "Parse multiple input crs specified, output crs specified",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor":     []string{"H3w%3D"},
					"filter-crs": []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox-crs":   []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox":       []string{"1,2,3,4"},
					"limit":      []string{"10000"},
				},
				limit: engine.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantEncodedCursor: "H3w%3D",
			wantLimit:         20, // use max instead of supplied limit
			wantOutputCrs:     100000,
			wantBbox:          (*geom.Extent)([]float64{1, 2, 3, 4}),
			wantInputCrs:      28992,
			wantErr:           success(),
		},
		{
			name: "Fail on difference in input crs",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor":     []string{"H3w%3D"},
					"filter-crs": []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox-crs":   []string{"http://www.opengis.net/def/crs/EPSG/0/4258"},
					"bbox":       []string{"1,2,3,4"},
					"limit":      []string{"10000"},
				},
				limit: engine.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				assert.Equalf(t, "bbox-crs and filter-crs need to be equal. Can't use more than one CRS as input, but input and output CRS may differ", err.Error(), "parse()")
				return false
			},
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
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				assert.Equalf(t, "crs param should start with http://www.opengis.net/def/crs/, got: EPSG:28992", err.Error(), "parse()")
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
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				assert.Equalf(t, "bbox should contain exactly 4 values separated by commas: minx,miny,maxx,maxy", err.Error(), "parse()")
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
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				assert.Equalf(t, "limit can't be negative", err.Error(), "parse()")
				return false
			},
		},
		{
			name: "Fail on unimplemented datetime",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"datetime": []string{time.Now().String()},
				},
				limit: engine.Limit{
					Default: 1,
					Max:     2,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				assert.Equalf(t, "datetime param is currently not supported", err.Error(), "parse()")
				return false
			},
		},
		{
			name: "Fail on unimplemented filter",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"filter": []string{"some CQL expression"},
				},
				limit: engine.Limit{
					Default: 1,
					Max:     2,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				assert.Equalf(t, "CQL filter param is currently not supported", err.Error(), "parse()")
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
			gotEncodedCursor, gotLimit, gotInputCrs, gotOutputCrs, _, gotBbox, _, err := fc.parse()
			if !tt.wantErr(t, err, "parse()") {
				return
			}
			assert.Equalf(t, tt.wantEncodedCursor, gotEncodedCursor, "parse()")
			assert.Equalf(t, tt.wantLimit, gotLimit, "parse()")
			assert.Equalf(t, tt.wantOutputCrs, gotOutputCrs.GetOrDefault(), "parse()")
			assert.Equalf(t, tt.wantBbox, gotBbox, "parse()")
			assert.Equalf(t, tt.wantInputCrs, gotInputCrs.GetOrDefault(), "parse()")
		})
	}
}

func success() func(t assert.TestingT, err error, i ...any) bool {
	return func(t assert.TestingT, err error, i ...any) bool {
		return true
	}
}
