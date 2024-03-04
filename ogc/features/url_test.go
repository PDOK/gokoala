package features

import (
	"math/rand"
	"net/url"
	"testing"
	"time"

	"github.com/PDOK/gokoala/config"

	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom"
	"github.com/stretchr/testify/assert"
)

func Test_featureCollectionURL_parseParams(t *testing.T) {
	type fields struct {
		baseURL   url.URL
		params    url.Values
		limit     config.Limit
		dtSupport bool
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
		wantRefDate       *time.Time
		wantPropFilters   map[string]string
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "Parse no params",
			fields: fields{
				baseURL: *host,
				params:  url.Values{},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantEncodedCursor: "",
			wantLimit:         10,
			wantOutputCrs:     100000,
			wantBbox:          nil,
			wantRefDate:       nil,
			wantInputCrs:      100000,
			wantErr:           success(),
		},
		{
			name: "Parse many params",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor":   []string{"H3w"},
					"crs":      []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox":     []string{"1,2,3,4"},
					"bbox-crs": []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"limit":    []string{"10000"},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantEncodedCursor: "H3w",
			wantLimit:         20, // use max instead of supplied limit
			wantOutputCrs:     28992,
			wantBbox:          (*geom.Extent)([]float64{1, 2, 3, 4}),
			wantRefDate:       nil,
			wantInputCrs:      28992,
			wantErr:           success(),
		},
		{
			name: "Parse input crs specified, no output crs specified",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor":   []string{"H3w"},
					"bbox":     []string{"1,2,3,4"},
					"bbox-crs": []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"limit":    []string{"10000"},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantEncodedCursor: "H3w",
			wantLimit:         20, // use max instead of supplied limit
			wantOutputCrs:     100000,
			wantBbox:          (*geom.Extent)([]float64{1, 2, 3, 4}),
			wantRefDate:       nil,
			wantInputCrs:      28992,
			wantErr:           success(),
		},
		{
			name: "Parse no input crs specified, output crs specified",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor": []string{"H3w"},
					"crs":    []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox":   []string{"1,2,3,4"},
					"limit":  []string{"10000"},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantEncodedCursor: "H3w",
			wantLimit:         20, // use max instead of supplied limit
			wantOutputCrs:     28992,
			wantBbox:          (*geom.Extent)([]float64{1, 2, 3, 4}),
			wantRefDate:       nil,
			wantInputCrs:      100000,
			wantErr:           success(),
		},
		{
			name: "Parse multiple input crs specified, output crs specified",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor":     []string{"H3w"},
					"filter-crs": []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox-crs":   []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox":       []string{"1,2,3,4"},
					"limit":      []string{"10000"},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantEncodedCursor: "H3w",
			wantLimit:         20, // use max instead of supplied limit
			wantOutputCrs:     100000,
			wantBbox:          (*geom.Extent)([]float64{1, 2, 3, 4}),
			wantRefDate:       nil,
			wantInputCrs:      28992,
			wantErr:           success(),
		},
		{
			name: "Parse datetime",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"datetime": []string{time.Time{}.Format(time.RFC3339)},
				},
				limit: config.Limit{
					Default: 1,
					Max:     2,
				},
				dtSupport: true,
			},
			wantLimit:     1,
			wantOutputCrs: 100000,
			wantInputCrs:  100000,
			wantRefDate:   &time.Time{},
			wantErr:       success(),
		},
		{
			name: "Parse property filters",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"foo": []string{"baz"},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantLimit:       10,
			wantOutputCrs:   100000,
			wantInputCrs:    100000,
			wantRefDate:     nil,
			wantPropFilters: map[string]string{"foo": "baz"},
			wantErr:         success(),
		},
		{
			name: "Parse multiple property filters",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"foo": []string{"baz"},
					"bar": []string{"bazz"},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantLimit:       10,
			wantOutputCrs:   100000,
			wantInputCrs:    100000,
			wantRefDate:     nil,
			wantPropFilters: map[string]string{"foo": "baz", "bar": "bazz"},
			wantErr:         success(),
		},
		{
			name: "Fail on invalid property filters",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"non_existent": []string{"baz"},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				assert.Equalf(t, "unknown query parameter(s) found: non_existent=baz", err.Error(), "parse()")
				return false
			},
		},
		{
			name: "Fail on wildcard property filter",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"foo": []string{"baz*"},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				assert.Equalf(t, "property filter foo contains a wildcard (*), wildcard filtering is not allowed", err.Error(), "parse()")
				return false
			},
		},
		{
			name: "Fail on too large property filter",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"foo": []string{generateRandomString(propertyFilterMaxLength + 1)},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				assert.Equalf(t, "property filter foo is too large, value is limited to 512 characters", err.Error(), "parse()")
				return false
			},
		},
		{
			name: "Fail on difference in input crs",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"cursor":     []string{"H3w"},
					"filter-crs": []string{"http://www.opengis.net/def/crs/EPSG/0/28992"},
					"bbox-crs":   []string{"http://www.opengis.net/def/crs/EPSG/0/4258"},
					"bbox":       []string{"1,2,3,4"},
					"limit":      []string{"10000"},
				},
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
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
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
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
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
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
				limit: config.Limit{
					Default: 10,
					Max:     20,
				},
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				assert.Equalf(t, "limit can't be negative", err.Error(), "parse()")
				return false
			},
		},
		{
			name: "Fail on unimplemented datetime interval",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"datetime": []string{"2023-11-10T23:00:00Z/2023-11-15T23:00:00Z"},
				},
				limit: config.Limit{
					Default: 1,
					Max:     2,
				},
				dtSupport: true,
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				assert.Equalf(t, "datetime param '2023-11-10T23:00:00Z/2023-11-15T23:00:00Z' represents an interval, intervals are currently not supported", err.Error(), "parse()")
				return false
			},
		},
		{
			name: "Fail on datetime not supported by collection",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"datetime": []string{"2023-11-10T23:00:00Z/2023-11-15T23:00:00Z"},
				},
				limit: config.Limit{
					Default: 1,
					Max:     2,
				},
				dtSupport: false,
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				assert.Equalf(t, "datetime param is currently not supported for this collection", err.Error(), "parse()")
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
				limit: config.Limit{
					Default: 1,
					Max:     2,
				},
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				assert.Equalf(t, "CQL filter param is currently not supported", err.Error(), "parse()")
				return false
			},
		},
		{
			name: "Fail on unknown param",
			fields: fields{
				baseURL: *host,
				params: url.Values{
					"this_param_does_not_exist_in_openapi_spec": []string{"foobar"},
				},
				limit: config.Limit{
					Default: 1,
					Max:     2,
				},
			},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				assert.Equalf(t, "unknown query parameter(s) found: this_param_does_not_exist_in_openapi_spec=foobar", err.Error(), "parse()")
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
				configuredPropertyFilters: []config.PropertyFilter{
					{
						Name:        "foo",
						Description: "awesome foo property to filter on",
					},
					{
						Name:        "bar",
						Description: "even more awesome bar property to filter on",
					},
				},
				supportsDatetime: tt.fields.dtSupport,
			}
			gotEncodedCursor, gotLimit, gotInputCrs, gotOutputCrs, _, gotBbox, _, gotPF, err := fc.parse()
			if !tt.wantErr(t, err, "parse()") {
				return
			}
			assert.Equalf(t, tt.wantEncodedCursor, gotEncodedCursor, "parse()")
			assert.Equalf(t, tt.wantLimit, gotLimit, "parse()")
			assert.Equalf(t, tt.wantOutputCrs, gotOutputCrs.GetOrDefault(), "parse()")
			assert.Equalf(t, tt.wantBbox, gotBbox, "parse()")
			assert.Equalf(t, tt.wantInputCrs, gotInputCrs.GetOrDefault(), "parse()")
			if tt.wantPropFilters != nil {
				assert.Equalf(t, tt.wantPropFilters, gotPF, "parse()")
			}
		})
	}
}

func success() func(t assert.TestingT, err error, i ...any) bool {
	return func(_ assert.TestingT, _ error, _ ...any) bool {
		return true
	}
}

func generateRandomString(length int) string {
	const charset = "abc"
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed) //nolint:gosec  // good enough for testing

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[random.Intn(len(charset))]
	}
	return string(result)
}
