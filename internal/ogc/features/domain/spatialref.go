package domain

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	UndefinedSRID    = 0
	WGS84SRID        = 100000 // We use the SRID for CRS84 (WGS84) as defined in the GeoPackage, instead of EPSG:4326 (due to axis order). In time, we may need to read this value dynamically from the GeoPackage.
	WGS84SRIDPostgis = 4326

	CrsURIPrefix = "http://www.opengis.net/def/crs/"
	WGS84CodeOGC = "CRS84"
	WGS84CrsURI  = CrsURIPrefix + "OGC/1.3/" + WGS84CodeOGC
	EPSGPrefix   = "EPSG:"
)

// AxisOrder the order of axis for a certain CRS
type AxisOrder int

const (
	AxisOrderXY AxisOrder = iota
	AxisOrderYX
)

// SRID Spatial Reference System Identifier: a unique value to unambiguously identify a spatial coordinate system.
// For example '28992' in https://www.opengis.net/def/crs/EPSG/0/28992
type SRID int

func (s SRID) GetOrDefault() int {
	val := int(s)
	if val <= 0 {
		return WGS84SRID
	}
	return val
}

func EpsgToSrid(srs string) (SRID, error) {
	srsCode, found := strings.CutPrefix(srs, EPSGPrefix)
	if !found {
		return -1, fmt.Errorf("expected SRS to start with '%s', got %s", EPSGPrefix, srs)
	}
	srid, err := strconv.Atoi(srsCode)
	if err != nil {
		return -1, fmt.Errorf("expected EPSG code to have numeric value, got %s", srsCode)
	}
	return SRID(srid), nil
}

// ContentCrs the coordinate reference system (represented as a URI) of the content/output to return.
type ContentCrs string

// ToLink returns link target conforming to RFC 8288
func (c ContentCrs) ToLink() string {
	return fmt.Sprintf("<%s>", c)
}

func (c ContentCrs) IsWGS84() bool {
	return string(c) == WGS84CrsURI
}
