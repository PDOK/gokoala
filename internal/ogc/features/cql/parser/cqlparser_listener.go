// Code generated from CqlParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // CqlParser
import "github.com/antlr4-go/antlr/v4"

// CqlParserListener is a complete listener for a parse tree produced by CqlParser.
type CqlParserListener interface {
	antlr.ParseTreeListener

	// EnterCqlFilter is called when entering the cqlFilter production.
	EnterCqlFilter(c *CqlFilterContext)

	// EnterBooleanExpression is called when entering the booleanExpression production.
	EnterBooleanExpression(c *BooleanExpressionContext)

	// EnterBooleanTerm is called when entering the booleanTerm production.
	EnterBooleanTerm(c *BooleanTermContext)

	// EnterBooleanFactor is called when entering the booleanFactor production.
	EnterBooleanFactor(c *BooleanFactorContext)

	// EnterBooleanPrimary is called when entering the booleanPrimary production.
	EnterBooleanPrimary(c *BooleanPrimaryContext)

	// EnterPredicate is called when entering the predicate production.
	EnterPredicate(c *PredicateContext)

	// EnterComparisonPredicate is called when entering the comparisonPredicate production.
	EnterComparisonPredicate(c *ComparisonPredicateContext)

	// EnterBinaryComparisonPredicate is called when entering the binaryComparisonPredicate production.
	EnterBinaryComparisonPredicate(c *BinaryComparisonPredicateContext)

	// EnterIsLikePredicate is called when entering the isLikePredicate production.
	EnterIsLikePredicate(c *IsLikePredicateContext)

	// EnterIsBetweenPredicate is called when entering the isBetweenPredicate production.
	EnterIsBetweenPredicate(c *IsBetweenPredicateContext)

	// EnterIsInListPredicate is called when entering the isInListPredicate production.
	EnterIsInListPredicate(c *IsInListPredicateContext)

	// EnterIsNullPredicate is called when entering the isNullPredicate production.
	EnterIsNullPredicate(c *IsNullPredicateContext)

	// EnterIsNullOperand is called when entering the isNullOperand production.
	EnterIsNullOperand(c *IsNullOperandContext)

	// EnterScalarExpression is called when entering the scalarExpression production.
	EnterScalarExpression(c *ScalarExpressionContext)

	// EnterCharacterExpression is called when entering the characterExpression production.
	EnterCharacterExpression(c *CharacterExpressionContext)

	// EnterPatternExpression is called when entering the patternExpression production.
	EnterPatternExpression(c *PatternExpressionContext)

	// EnterCharacterClause is called when entering the characterClause production.
	EnterCharacterClause(c *CharacterClauseContext)

	// EnterCharacterLiteral is called when entering the characterLiteral production.
	EnterCharacterLiteral(c *CharacterLiteralContext)

	// EnterNumericExpression is called when entering the numericExpression production.
	EnterNumericExpression(c *NumericExpressionContext)

	// EnterNumericLiteral is called when entering the numericLiteral production.
	EnterNumericLiteral(c *NumericLiteralContext)

	// EnterBooleanLiteral is called when entering the booleanLiteral production.
	EnterBooleanLiteral(c *BooleanLiteralContext)

	// EnterPropertyName is called when entering the propertyName production.
	EnterPropertyName(c *PropertyNameContext)

	// EnterSpatialPredicate is called when entering the spatialPredicate production.
	EnterSpatialPredicate(c *SpatialPredicateContext)

	// EnterGeomExpression is called when entering the geomExpression production.
	EnterGeomExpression(c *GeomExpressionContext)

	// EnterSpatialInstance is called when entering the spatialInstance production.
	EnterSpatialInstance(c *SpatialInstanceContext)

	// EnterGeometryLiteral is called when entering the geometryLiteral production.
	EnterGeometryLiteral(c *GeometryLiteralContext)

	// EnterPoint is called when entering the point production.
	EnterPoint(c *PointContext)

	// EnterLinestring is called when entering the linestring production.
	EnterLinestring(c *LinestringContext)

	// EnterLinestringDef is called when entering the linestringDef production.
	EnterLinestringDef(c *LinestringDefContext)

	// EnterPolygon is called when entering the polygon production.
	EnterPolygon(c *PolygonContext)

	// EnterPolygonDef is called when entering the polygonDef production.
	EnterPolygonDef(c *PolygonDefContext)

	// EnterMultiPoint is called when entering the multiPoint production.
	EnterMultiPoint(c *MultiPointContext)

	// EnterMultiLinestring is called when entering the multiLinestring production.
	EnterMultiLinestring(c *MultiLinestringContext)

	// EnterMultiPolygon is called when entering the multiPolygon production.
	EnterMultiPolygon(c *MultiPolygonContext)

	// EnterGeometryCollection is called when entering the geometryCollection production.
	EnterGeometryCollection(c *GeometryCollectionContext)

	// EnterBbox is called when entering the bbox production.
	EnterBbox(c *BboxContext)

	// EnterCoordinate is called when entering the coordinate production.
	EnterCoordinate(c *CoordinateContext)

	// EnterXCoord is called when entering the xCoord production.
	EnterXCoord(c *XCoordContext)

	// EnterYCoord is called when entering the yCoord production.
	EnterYCoord(c *YCoordContext)

	// EnterZCoord is called when entering the zCoord production.
	EnterZCoord(c *ZCoordContext)

	// EnterWestBoundLon is called when entering the westBoundLon production.
	EnterWestBoundLon(c *WestBoundLonContext)

	// EnterEastBoundLon is called when entering the eastBoundLon production.
	EnterEastBoundLon(c *EastBoundLonContext)

	// EnterNorthBoundLat is called when entering the northBoundLat production.
	EnterNorthBoundLat(c *NorthBoundLatContext)

	// EnterSouthBoundLat is called when entering the southBoundLat production.
	EnterSouthBoundLat(c *SouthBoundLatContext)

	// EnterMinElev is called when entering the minElev production.
	EnterMinElev(c *MinElevContext)

	// EnterMaxElev is called when entering the maxElev production.
	EnterMaxElev(c *MaxElevContext)

	// EnterTemporalPredicate is called when entering the temporalPredicate production.
	EnterTemporalPredicate(c *TemporalPredicateContext)

	// EnterTemporalExpression is called when entering the temporalExpression production.
	EnterTemporalExpression(c *TemporalExpressionContext)

	// EnterTemporalClause is called when entering the temporalClause production.
	EnterTemporalClause(c *TemporalClauseContext)

	// EnterInstantInstance is called when entering the instantInstance production.
	EnterInstantInstance(c *InstantInstanceContext)

	// EnterInterval is called when entering the interval production.
	EnterInterval(c *IntervalContext)

	// EnterIntervalParameter is called when entering the intervalParameter production.
	EnterIntervalParameter(c *IntervalParameterContext)

	// EnterArrayPredicate is called when entering the arrayPredicate production.
	EnterArrayPredicate(c *ArrayPredicateContext)

	// EnterArrayExpression is called when entering the arrayExpression production.
	EnterArrayExpression(c *ArrayExpressionContext)

	// EnterArrayClause is called when entering the arrayClause production.
	EnterArrayClause(c *ArrayClauseContext)

	// EnterArrayElement is called when entering the arrayElement production.
	EnterArrayElement(c *ArrayElementContext)

	// EnterFunction is called when entering the function production.
	EnterFunction(c *FunctionContext)

	// EnterArgumentList is called when entering the argumentList production.
	EnterArgumentList(c *ArgumentListContext)

	// EnterPositionalArgument is called when entering the positionalArgument production.
	EnterPositionalArgument(c *PositionalArgumentContext)

	// EnterArgument is called when entering the argument production.
	EnterArgument(c *ArgumentContext)

	// ExitCqlFilter is called when exiting the cqlFilter production.
	ExitCqlFilter(c *CqlFilterContext)

	// ExitBooleanExpression is called when exiting the booleanExpression production.
	ExitBooleanExpression(c *BooleanExpressionContext)

	// ExitBooleanTerm is called when exiting the booleanTerm production.
	ExitBooleanTerm(c *BooleanTermContext)

	// ExitBooleanFactor is called when exiting the booleanFactor production.
	ExitBooleanFactor(c *BooleanFactorContext)

	// ExitBooleanPrimary is called when exiting the booleanPrimary production.
	ExitBooleanPrimary(c *BooleanPrimaryContext)

	// ExitPredicate is called when exiting the predicate production.
	ExitPredicate(c *PredicateContext)

	// ExitComparisonPredicate is called when exiting the comparisonPredicate production.
	ExitComparisonPredicate(c *ComparisonPredicateContext)

	// ExitBinaryComparisonPredicate is called when exiting the binaryComparisonPredicate production.
	ExitBinaryComparisonPredicate(c *BinaryComparisonPredicateContext)

	// ExitIsLikePredicate is called when exiting the isLikePredicate production.
	ExitIsLikePredicate(c *IsLikePredicateContext)

	// ExitIsBetweenPredicate is called when exiting the isBetweenPredicate production.
	ExitIsBetweenPredicate(c *IsBetweenPredicateContext)

	// ExitIsInListPredicate is called when exiting the isInListPredicate production.
	ExitIsInListPredicate(c *IsInListPredicateContext)

	// ExitIsNullPredicate is called when exiting the isNullPredicate production.
	ExitIsNullPredicate(c *IsNullPredicateContext)

	// ExitIsNullOperand is called when exiting the isNullOperand production.
	ExitIsNullOperand(c *IsNullOperandContext)

	// ExitScalarExpression is called when exiting the scalarExpression production.
	ExitScalarExpression(c *ScalarExpressionContext)

	// ExitCharacterExpression is called when exiting the characterExpression production.
	ExitCharacterExpression(c *CharacterExpressionContext)

	// ExitPatternExpression is called when exiting the patternExpression production.
	ExitPatternExpression(c *PatternExpressionContext)

	// ExitCharacterClause is called when exiting the characterClause production.
	ExitCharacterClause(c *CharacterClauseContext)

	// ExitCharacterLiteral is called when exiting the characterLiteral production.
	ExitCharacterLiteral(c *CharacterLiteralContext)

	// ExitNumericExpression is called when exiting the numericExpression production.
	ExitNumericExpression(c *NumericExpressionContext)

	// ExitNumericLiteral is called when exiting the numericLiteral production.
	ExitNumericLiteral(c *NumericLiteralContext)

	// ExitBooleanLiteral is called when exiting the booleanLiteral production.
	ExitBooleanLiteral(c *BooleanLiteralContext)

	// ExitPropertyName is called when exiting the propertyName production.
	ExitPropertyName(c *PropertyNameContext)

	// ExitSpatialPredicate is called when exiting the spatialPredicate production.
	ExitSpatialPredicate(c *SpatialPredicateContext)

	// ExitGeomExpression is called when exiting the geomExpression production.
	ExitGeomExpression(c *GeomExpressionContext)

	// ExitSpatialInstance is called when exiting the spatialInstance production.
	ExitSpatialInstance(c *SpatialInstanceContext)

	// ExitGeometryLiteral is called when exiting the geometryLiteral production.
	ExitGeometryLiteral(c *GeometryLiteralContext)

	// ExitPoint is called when exiting the point production.
	ExitPoint(c *PointContext)

	// ExitLinestring is called when exiting the linestring production.
	ExitLinestring(c *LinestringContext)

	// ExitLinestringDef is called when exiting the linestringDef production.
	ExitLinestringDef(c *LinestringDefContext)

	// ExitPolygon is called when exiting the polygon production.
	ExitPolygon(c *PolygonContext)

	// ExitPolygonDef is called when exiting the polygonDef production.
	ExitPolygonDef(c *PolygonDefContext)

	// ExitMultiPoint is called when exiting the multiPoint production.
	ExitMultiPoint(c *MultiPointContext)

	// ExitMultiLinestring is called when exiting the multiLinestring production.
	ExitMultiLinestring(c *MultiLinestringContext)

	// ExitMultiPolygon is called when exiting the multiPolygon production.
	ExitMultiPolygon(c *MultiPolygonContext)

	// ExitGeometryCollection is called when exiting the geometryCollection production.
	ExitGeometryCollection(c *GeometryCollectionContext)

	// ExitBbox is called when exiting the bbox production.
	ExitBbox(c *BboxContext)

	// ExitCoordinate is called when exiting the coordinate production.
	ExitCoordinate(c *CoordinateContext)

	// ExitXCoord is called when exiting the xCoord production.
	ExitXCoord(c *XCoordContext)

	// ExitYCoord is called when exiting the yCoord production.
	ExitYCoord(c *YCoordContext)

	// ExitZCoord is called when exiting the zCoord production.
	ExitZCoord(c *ZCoordContext)

	// ExitWestBoundLon is called when exiting the westBoundLon production.
	ExitWestBoundLon(c *WestBoundLonContext)

	// ExitEastBoundLon is called when exiting the eastBoundLon production.
	ExitEastBoundLon(c *EastBoundLonContext)

	// ExitNorthBoundLat is called when exiting the northBoundLat production.
	ExitNorthBoundLat(c *NorthBoundLatContext)

	// ExitSouthBoundLat is called when exiting the southBoundLat production.
	ExitSouthBoundLat(c *SouthBoundLatContext)

	// ExitMinElev is called when exiting the minElev production.
	ExitMinElev(c *MinElevContext)

	// ExitMaxElev is called when exiting the maxElev production.
	ExitMaxElev(c *MaxElevContext)

	// ExitTemporalPredicate is called when exiting the temporalPredicate production.
	ExitTemporalPredicate(c *TemporalPredicateContext)

	// ExitTemporalExpression is called when exiting the temporalExpression production.
	ExitTemporalExpression(c *TemporalExpressionContext)

	// ExitTemporalClause is called when exiting the temporalClause production.
	ExitTemporalClause(c *TemporalClauseContext)

	// ExitInstantInstance is called when exiting the instantInstance production.
	ExitInstantInstance(c *InstantInstanceContext)

	// ExitInterval is called when exiting the interval production.
	ExitInterval(c *IntervalContext)

	// ExitIntervalParameter is called when exiting the intervalParameter production.
	ExitIntervalParameter(c *IntervalParameterContext)

	// ExitArrayPredicate is called when exiting the arrayPredicate production.
	ExitArrayPredicate(c *ArrayPredicateContext)

	// ExitArrayExpression is called when exiting the arrayExpression production.
	ExitArrayExpression(c *ArrayExpressionContext)

	// ExitArrayClause is called when exiting the arrayClause production.
	ExitArrayClause(c *ArrayClauseContext)

	// ExitArrayElement is called when exiting the arrayElement production.
	ExitArrayElement(c *ArrayElementContext)

	// ExitFunction is called when exiting the function production.
	ExitFunction(c *FunctionContext)

	// ExitArgumentList is called when exiting the argumentList production.
	ExitArgumentList(c *ArgumentListContext)

	// ExitPositionalArgument is called when exiting the positionalArgument production.
	ExitPositionalArgument(c *PositionalArgumentContext)

	// ExitArgument is called when exiting the argument production.
	ExitArgument(c *ArgumentContext)
}
