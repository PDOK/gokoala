package domain

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	UndefinedSRID    = 0
	WGS84SRID        = 100000 // We use the SRID for CRS84 (WGS84) as defined in the GeoPackage, instead of EPSG:4326 (due to axis order). In time, we may need to read this value dynamically from the GeoPackage.
	WGS84SRIDPostgis = 4326   // PostGIS knows one SRID for WGS84

	CrsURIPrefix = "http://www.opengis.net/def/crs/"
	WGS84CodeOGC = "CRS84"
	WGS84CrsURI  = CrsURIPrefix + "OGC/1.3/" + WGS84CodeOGC
	EPSGPrefix   = "EPSG:"
	OGCPrefix    = "OGC:"
)

// AxisOrder the order of axis for a certain CRS.
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

func (s SRID) ToContentCrs() ContentCrs {
	return ContentCrs(fmt.Sprintf("%sEPSG/0/%d", CrsURIPrefix, s))
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

// ParseCrsURIToSRID parses an HTTP OGC CRS identifier.
func ParseCrsURIToSRID(identifier string) (SRID, error) {
	identifier = strings.TrimSpace(identifier)
	if !strings.HasPrefix(identifier, CrsURIPrefix) {
		return UndefinedSRID, fmt.Errorf("should start with %s, got: %s", CrsURIPrefix, identifier)
	}
	crsCode := identifier[strings.LastIndex(identifier, "/")+1:]
	if crsCode == WGS84CodeOGC {
		// CRS84 uses WGS84 with longitude/latitude axis order, unlike EPSG:4326.
		return WGS84SRID, nil
	}
	value, err := strconv.Atoi(crsCode)
	if err != nil {
		return UndefinedSRID, fmt.Errorf("expected numerical CRS code, received: %s", crsCode)
	}
	return SRID(value), nil
}

// HasSafeCURIESyntax checks if CRS identifier looks like a safe CURIE (enclosed in brackets).
func HasSafeCURIESyntax(identifier string) bool {
	return strings.HasPrefix(identifier, "[") && strings.HasSuffix(identifier, "]")
}

// ParseSafeCURIEToSRID parses bracketed EPSG identifiers or [OGC:CRS84].
func ParseSafeCURIEToSRID(identifier string) (SRID, error) {
	identifier = strings.TrimSpace(identifier)
	if !HasSafeCURIESyntax(identifier) {
		return UndefinedSRID, fmt.Errorf("safe CURIE must be enclosed in brackets, got %q", identifier)
	}
	value := identifier[1 : len(identifier)-1]
	if value == OGCPrefix+WGS84CodeOGC {
		return WGS84SRID, nil
	}
	code, found := strings.CutPrefix(value, EPSGPrefix)
	if !found || code == "" || strings.ContainsAny(code, "[]") ||
		strings.ContainsFunc(code, func(r rune) bool { return r < '0' || r > '9' }) {
		return UndefinedSRID, fmt.Errorf("expected [EPSG:positive-integer] or [OGC:CRS84], got %q", identifier)
	}
	srid, err := EpsgToSrid(value)
	if err != nil {
		return UndefinedSRID, err
	}
	if srid <= 0 {
		return UndefinedSRID, fmt.Errorf("CRS code must be positive, got %q", identifier)
	}
	return srid, nil
}

// ContentCrs the coordinate reference system (represented as a URI) of the content/output to return.
type ContentCrs string

// ToLink returns link target conforming to RFC 8288.
func (c ContentCrs) ToLink() string {
	return fmt.Sprintf("<%s>", c)
}

func (c ContentCrs) IsWGS84() bool {
	return string(c) == WGS84CrsURI
}
