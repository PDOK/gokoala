package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		srid     SRID
		expected int
	}{
		{"Positive SRID", SRID(28992), 28992},
		{"Zero SRID", SRID(0), WGS84SRID},
		{"Negative SRID", SRID(-1), WGS84SRID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.srid.GetOrDefault())
		})
	}
}

func TestEpsgToSrid(t *testing.T) {
	tests := []struct {
		name        string
		srs         string
		expected    SRID
		expectError bool
	}{
		{"Valid EPSG", "EPSG:28992", SRID(28992), false},
		{"Invalid prefix", "INVALID:28992", SRID(-1), true},
		{"Non-numeric EPSG code", "EPSG:ABC", SRID(-1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EpsgToSrid(tt.srs)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseCrsURIToSRID(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       SRID
		wantErr    bool
	}{
		{name: "EPSG URI", identifier: CrsURIPrefix + "EPSG/0/28992", want: 28992},
		{name: "CRS84", identifier: WGS84CrsURI, want: WGS84SRID},
		{name: "EPSG4326 remains distinct", identifier: CrsURIPrefix + "EPSG/0/4326", want: 4326},
		{name: "whitespace", identifier: " \t" + CrsURIPrefix + "EPSG/0/28992\n", want: 28992},
		{name: "legacy permissive path", identifier: CrsURIPrefix + "other/version/28992", want: 28992},
		{name: "legacy zero code", identifier: CrsURIPrefix + "EPSG/0/0", want: UndefinedSRID},
		{name: "legacy negative code", identifier: CrsURIPrefix + "EPSG/0/-1", want: -1},
		{name: "empty", wantErr: true},
		{name: "invalid prefix", identifier: "http://example.com/EPSG/0/28992", wantErr: true},
		{name: "nonnumeric code", identifier: CrsURIPrefix + "EPSG/0/abc", wantErr: true},
		{name: "HTTPS", identifier: "https://www.opengis.net/def/crs/EPSG/0/28992", wantErr: true},
		{name: "safe CURIE", identifier: "[EPSG:28992]", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCrsURIToSRID(tt.identifier)
			if tt.wantErr {
				require.Error(t, err)
				assert.Zero(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseSafeCURIEToSRID(t *testing.T) {
	tests := []struct {
		identifier string
		want       SRID
		wantErr    bool
	}{
		{identifier: "[EPSG:28992]", want: 28992},
		{identifier: "[EPSG:4326]", want: 4326},
		{identifier: "[OGC:CRS84]", want: WGS84SRID},
		{identifier: " \t[EPSG:28992]\n", want: 28992},
		{identifier: "", wantErr: true},
		{identifier: "EPSG:28992", wantErr: true},
		{identifier: "[EPSG:28992", wantErr: true},
		{identifier: "EPSG:28992]", wantErr: true},
		{identifier: "[[EPSG:28992]]", wantErr: true},
		{identifier: "[]", wantErr: true},
		{identifier: "[EPSG:]", wantErr: true},
		{identifier: "[EPSG:0]", wantErr: true},
		{identifier: "[EPSG:-1]", wantErr: true},
		{identifier: "[EPSG:+1]", wantErr: true},
		{identifier: "[EPSG:1.5]", wantErr: true},
		{identifier: "[EPSG:abc]", wantErr: true},
		{identifier: "[EPSG: 28992]", wantErr: true},
		{identifier: "[EPSG:999999999999999999999999999999]", wantErr: true},
		{identifier: "[epsg:28992]", wantErr: true},
		{identifier: "[OTHER:28992]", wantErr: true},
		{identifier: "[OGC:4326]", wantErr: true},
		{identifier: WGS84CrsURI, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.identifier, func(t *testing.T) {
			got, err := ParseSafeCURIEToSRID(tt.identifier)
			if tt.wantErr {
				require.Error(t, err)
				assert.Zero(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
