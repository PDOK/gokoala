package cql

import (
	"fmt"
	"strings"

	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/PDOK/gokoala/internal/ogc/features/cql/parser"
)

const (
	bboxKeyword  = "BBOX"
	pointKeyword = "POINT"
)

// CQL-to-Spatial function mapping, both for SpatiaLite and Postgres.
var spatialFunctions = map[string]string{
	"S_INTERSECTS": "ST_Intersects",
	"S_DISJOINT":   "ST_Disjoint",
	"S_TOUCHES":    "ST_Touches",
	"S_WITHIN":     "ST_Within",
	"S_OVERLAPS":   "ST_Overlaps",
	"S_CROSSES":    "ST_Crosses",
	"S_CONTAINS":   "ST_Contains",
	"S_EQUALS":     "ST_Equals",
}

// ExitCoordinate Spatial coordinate
func (cl *CommonListener) ExitCoordinate(ctx *parser.CoordinateContext) {
	y := ctx.YCoord().GetText()
	x := ctx.XCoord().GetText()
	coordinate := x + " " + y

	if ctx.ZCoord() != nil {
		coordinate += " " + ctx.ZCoord().GetText()
	}
	cl.stack.Push(coordinate)
}

// ExitPoint Handle POINT Well-Known Text (WKT) literal
func (cl *CommonListener) ExitPoint(ctx *parser.PointContext) {
	cl.currentWktType = ctx.POINT().GetText()
	coordinate := cl.stack.Pop()
	cl.stack.Push(fmt.Sprintf("%s(%s)", cl.currentWktType, coordinate))
}

// ExitLinestring Handle LINESTRING Well-Known Text (WKT) literal
func (cl *CommonListener) ExitLinestring(ctx *parser.LinestringContext) {
	cl.currentWktType = ctx.LINESTRING().GetText()
	coordinates := cl.stack.Pop()
	cl.stack.Push(cl.currentWktType + coordinates)
}

// ExitLinestringDef Handle LINESTRING coordinates
func (cl *CommonListener) ExitLinestringDef(ctx *parser.LinestringDefContext) {
	count := len(ctx.AllCoordinate())
	coordinates := cl.stack.PopMany(count)
	cl.stack.Push("(" + strings.Join(coordinates, ", ") + ")")
}

// ExitPolygon Handle POLYGON Well-Known Text (WKT) literal
func (cl *CommonListener) ExitPolygon(ctx *parser.PolygonContext) {
	cl.currentWktType = ctx.POLYGON().GetText()
	coordinates := cl.stack.Pop()
	cl.stack.Push(cl.currentWktType + coordinates)
}

// ExitPolygonDef Handle POLYGON coordinates
func (cl *CommonListener) ExitPolygonDef(ctx *parser.PolygonDefContext) {
	count := len(ctx.AllLinestringDef())
	coordinates := cl.stack.PopMany(count)
	cl.stack.Push("(" + strings.Join(coordinates, ", ") + ")")
}

// ExitMultiPoint Handle MULTIPOINT Well-Known Text (WKT) literal
func (cl *CommonListener) ExitMultiPoint(ctx *parser.MultiPointContext) {
	count := len(ctx.AllMultiPointDef())
	coordinates := cl.stack.PopMany(count)
	cl.currentWktType = ctx.MULTIPOINT().GetText()
	cl.stack.Push(fmt.Sprintf("%s(%s)", cl.currentWktType, strings.Join(coordinates, ", ")))
}

// ExitMultiPointDef Handle MULTIPOINT coordinates
func (cl *CommonListener) ExitMultiPointDef(ctx *parser.MultiPointDefContext) {
	if ctx.LEFTPAREN() != nil {
		// handle alternative notation for MULTIPOINT. The one with extra parentheses.
		coordinate := cl.stack.Pop()
		cl.stack.Push("(" + coordinate + ")")
	}
}

// ExitMultiLinestring Handle MULTILINESTRING Well-Known Text (WKT) literal
func (cl *CommonListener) ExitMultiLinestring(ctx *parser.MultiLinestringContext) {
	count := len(ctx.AllLinestringDef())
	coordinates := cl.stack.PopMany(count)
	cl.currentWktType = ctx.MULTILINESTRING().GetText()
	cl.stack.Push(fmt.Sprintf("%s(%s)", cl.currentWktType, strings.Join(coordinates, ", ")))
}

// ExitMultiPolygon Handle MULTIPOLYGON Well-Known Text (WKT) literal
func (cl *CommonListener) ExitMultiPolygon(ctx *parser.MultiPolygonContext) {
	count := len(ctx.AllPolygonDef())
	coordinates := cl.stack.PopMany(count)
	cl.currentWktType = ctx.MULTIPOLYGON().GetText()
	cl.stack.Push(fmt.Sprintf("%s(%s)", cl.currentWktType, strings.Join(coordinates, ", ")))
}

// ExitGeometryCollection Handle GEOMETRYCOLLECTION Well-Known Text (WKT) literal
func (cl *CommonListener) ExitGeometryCollection(ctx *parser.GeometryCollectionContext) {
	count := len(ctx.AllGeometryLiteral())
	literals := cl.stack.PopMany(count)
	cl.currentWktType = ctx.GEOMETRYCOLLECTION().GetText()
	cl.stack.Push(fmt.Sprintf("%s(%s)", cl.currentWktType, strings.Join(literals, ", ")))
}

func (cl *CommonListener) isSpatialFilterAllowed(cqlFunction string) bool {
	isBboxOrPoint := func() bool {
		return strings.ToUpper(cl.currentWktType) == bboxKeyword || strings.ToUpper(cl.currentWktType) == pointKeyword
	}

	if cl.collectionType != geospatial.Features {
		cl.errorListener.Errorf("spatial filtering using '%s' is not allowed for this collection since it does not "+
			"contain geospatial items (features), only non-geospatial items (attributes)", cqlFunction)
		return false
	}
	if !cl.cqlConfig.IsBasicSpatialFunctionsEnabled() {
		cl.errorListener.Error(errSpatialOperatorsNotEnabled)
		return false
	}
	if !cl.cqlConfig.IsSpatialFunctionsEnabled() && strings.ToUpper(cqlFunction) != "S_INTERSECTS" {
		cl.errorListener.Errorf("spatial operator '%s' is not enabled for this collection, only S_INTERSECTS is allowed", cqlFunction)
		return false
	}
	if !cl.cqlConfig.IsBasicSpatialFunctionsPlusEnabled() && !cl.cqlConfig.IsSpatialFunctionsEnabled() && !isBboxOrPoint() {
		cl.errorListener.Errorf("geometry type '%s' is not allowed, only %s and %s are "+
			"allowed with basic spatial filtering", cl.currentWktType, pointKeyword, bboxKeyword)
		return false
	}
	return true
}
