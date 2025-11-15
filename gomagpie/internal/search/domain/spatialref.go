package domain

const (
	CrsURIPrefix     = "http://www.opengis.net/def/crs/"
	UndefinedSRID    = 0
	WGS84SRIDPostgis = 4326 // Use the same SRID as used during ETL
	WGS84CodeOGC     = "CRS84"
)

// SRID Spatial Reference System Identifier: a unique value to unambiguously identify a spatial coordinate system.
// For example '28992' in https://www.opengis.net/def/crs/EPSG/0/28992
type SRID int
