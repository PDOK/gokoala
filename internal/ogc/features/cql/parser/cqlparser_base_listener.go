// Code generated from CqlParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // CqlParser
import "github.com/antlr4-go/antlr/v4"

// BaseCqlParserListener is a complete listener for a parse tree produced by CqlParser.
type BaseCqlParserListener struct{}

var _ CqlParserListener = &BaseCqlParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseCqlParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseCqlParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseCqlParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseCqlParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterCqlFilter is called when production cqlFilter is entered.
func (s *BaseCqlParserListener) EnterCqlFilter(ctx *CqlFilterContext) {}

// ExitCqlFilter is called when production cqlFilter is exited.
func (s *BaseCqlParserListener) ExitCqlFilter(ctx *CqlFilterContext) {}

// EnterBooleanExpression is called when production booleanExpression is entered.
func (s *BaseCqlParserListener) EnterBooleanExpression(ctx *BooleanExpressionContext) {}

// ExitBooleanExpression is called when production booleanExpression is exited.
func (s *BaseCqlParserListener) ExitBooleanExpression(ctx *BooleanExpressionContext) {}

// EnterBooleanTerm is called when production booleanTerm is entered.
func (s *BaseCqlParserListener) EnterBooleanTerm(ctx *BooleanTermContext) {}

// ExitBooleanTerm is called when production booleanTerm is exited.
func (s *BaseCqlParserListener) ExitBooleanTerm(ctx *BooleanTermContext) {}

// EnterBooleanFactor is called when production booleanFactor is entered.
func (s *BaseCqlParserListener) EnterBooleanFactor(ctx *BooleanFactorContext) {}

// ExitBooleanFactor is called when production booleanFactor is exited.
func (s *BaseCqlParserListener) ExitBooleanFactor(ctx *BooleanFactorContext) {}

// EnterBooleanPrimary is called when production booleanPrimary is entered.
func (s *BaseCqlParserListener) EnterBooleanPrimary(ctx *BooleanPrimaryContext) {}

// ExitBooleanPrimary is called when production booleanPrimary is exited.
func (s *BaseCqlParserListener) ExitBooleanPrimary(ctx *BooleanPrimaryContext) {}

// EnterPredicate is called when production predicate is entered.
func (s *BaseCqlParserListener) EnterPredicate(ctx *PredicateContext) {}

// ExitPredicate is called when production predicate is exited.
func (s *BaseCqlParserListener) ExitPredicate(ctx *PredicateContext) {}

// EnterComparisonPredicate is called when production comparisonPredicate is entered.
func (s *BaseCqlParserListener) EnterComparisonPredicate(ctx *ComparisonPredicateContext) {}

// ExitComparisonPredicate is called when production comparisonPredicate is exited.
func (s *BaseCqlParserListener) ExitComparisonPredicate(ctx *ComparisonPredicateContext) {}

// EnterBinaryComparisonPredicate is called when production binaryComparisonPredicate is entered.
func (s *BaseCqlParserListener) EnterBinaryComparisonPredicate(ctx *BinaryComparisonPredicateContext) {
}

// ExitBinaryComparisonPredicate is called when production binaryComparisonPredicate is exited.
func (s *BaseCqlParserListener) ExitBinaryComparisonPredicate(ctx *BinaryComparisonPredicateContext) {
}

// EnterIsLikePredicate is called when production isLikePredicate is entered.
func (s *BaseCqlParserListener) EnterIsLikePredicate(ctx *IsLikePredicateContext) {}

// ExitIsLikePredicate is called when production isLikePredicate is exited.
func (s *BaseCqlParserListener) ExitIsLikePredicate(ctx *IsLikePredicateContext) {}

// EnterIsBetweenPredicate is called when production isBetweenPredicate is entered.
func (s *BaseCqlParserListener) EnterIsBetweenPredicate(ctx *IsBetweenPredicateContext) {}

// ExitIsBetweenPredicate is called when production isBetweenPredicate is exited.
func (s *BaseCqlParserListener) ExitIsBetweenPredicate(ctx *IsBetweenPredicateContext) {}

// EnterIsInListPredicate is called when production isInListPredicate is entered.
func (s *BaseCqlParserListener) EnterIsInListPredicate(ctx *IsInListPredicateContext) {}

// ExitIsInListPredicate is called when production isInListPredicate is exited.
func (s *BaseCqlParserListener) ExitIsInListPredicate(ctx *IsInListPredicateContext) {}

// EnterIsNullPredicate is called when production isNullPredicate is entered.
func (s *BaseCqlParserListener) EnterIsNullPredicate(ctx *IsNullPredicateContext) {}

// ExitIsNullPredicate is called when production isNullPredicate is exited.
func (s *BaseCqlParserListener) ExitIsNullPredicate(ctx *IsNullPredicateContext) {}

// EnterIsNullOperand is called when production isNullOperand is entered.
func (s *BaseCqlParserListener) EnterIsNullOperand(ctx *IsNullOperandContext) {}

// ExitIsNullOperand is called when production isNullOperand is exited.
func (s *BaseCqlParserListener) ExitIsNullOperand(ctx *IsNullOperandContext) {}

// EnterScalarExpression is called when production scalarExpression is entered.
func (s *BaseCqlParserListener) EnterScalarExpression(ctx *ScalarExpressionContext) {}

// ExitScalarExpression is called when production scalarExpression is exited.
func (s *BaseCqlParserListener) ExitScalarExpression(ctx *ScalarExpressionContext) {}

// EnterCharacterExpression is called when production characterExpression is entered.
func (s *BaseCqlParserListener) EnterCharacterExpression(ctx *CharacterExpressionContext) {}

// ExitCharacterExpression is called when production characterExpression is exited.
func (s *BaseCqlParserListener) ExitCharacterExpression(ctx *CharacterExpressionContext) {}

// EnterPatternExpression is called when production patternExpression is entered.
func (s *BaseCqlParserListener) EnterPatternExpression(ctx *PatternExpressionContext) {}

// ExitPatternExpression is called when production patternExpression is exited.
func (s *BaseCqlParserListener) ExitPatternExpression(ctx *PatternExpressionContext) {}

// EnterCharacterClause is called when production characterClause is entered.
func (s *BaseCqlParserListener) EnterCharacterClause(ctx *CharacterClauseContext) {}

// ExitCharacterClause is called when production characterClause is exited.
func (s *BaseCqlParserListener) ExitCharacterClause(ctx *CharacterClauseContext) {}

// EnterCharacterLiteral is called when production characterLiteral is entered.
func (s *BaseCqlParserListener) EnterCharacterLiteral(ctx *CharacterLiteralContext) {}

// ExitCharacterLiteral is called when production characterLiteral is exited.
func (s *BaseCqlParserListener) ExitCharacterLiteral(ctx *CharacterLiteralContext) {}

// EnterNumericExpression is called when production numericExpression is entered.
func (s *BaseCqlParserListener) EnterNumericExpression(ctx *NumericExpressionContext) {}

// ExitNumericExpression is called when production numericExpression is exited.
func (s *BaseCqlParserListener) ExitNumericExpression(ctx *NumericExpressionContext) {}

// EnterNumericLiteral is called when production numericLiteral is entered.
func (s *BaseCqlParserListener) EnterNumericLiteral(ctx *NumericLiteralContext) {}

// ExitNumericLiteral is called when production numericLiteral is exited.
func (s *BaseCqlParserListener) ExitNumericLiteral(ctx *NumericLiteralContext) {}

// EnterBooleanLiteral is called when production booleanLiteral is entered.
func (s *BaseCqlParserListener) EnterBooleanLiteral(ctx *BooleanLiteralContext) {}

// ExitBooleanLiteral is called when production booleanLiteral is exited.
func (s *BaseCqlParserListener) ExitBooleanLiteral(ctx *BooleanLiteralContext) {}

// EnterPropertyName is called when production propertyName is entered.
func (s *BaseCqlParserListener) EnterPropertyName(ctx *PropertyNameContext) {}

// ExitPropertyName is called when production propertyName is exited.
func (s *BaseCqlParserListener) ExitPropertyName(ctx *PropertyNameContext) {}

// EnterSpatialPredicate is called when production spatialPredicate is entered.
func (s *BaseCqlParserListener) EnterSpatialPredicate(ctx *SpatialPredicateContext) {}

// ExitSpatialPredicate is called when production spatialPredicate is exited.
func (s *BaseCqlParserListener) ExitSpatialPredicate(ctx *SpatialPredicateContext) {}

// EnterGeomExpression is called when production geomExpression is entered.
func (s *BaseCqlParserListener) EnterGeomExpression(ctx *GeomExpressionContext) {}

// ExitGeomExpression is called when production geomExpression is exited.
func (s *BaseCqlParserListener) ExitGeomExpression(ctx *GeomExpressionContext) {}

// EnterSpatialInstance is called when production spatialInstance is entered.
func (s *BaseCqlParserListener) EnterSpatialInstance(ctx *SpatialInstanceContext) {}

// ExitSpatialInstance is called when production spatialInstance is exited.
func (s *BaseCqlParserListener) ExitSpatialInstance(ctx *SpatialInstanceContext) {}

// EnterGeometryLiteral is called when production geometryLiteral is entered.
func (s *BaseCqlParserListener) EnterGeometryLiteral(ctx *GeometryLiteralContext) {}

// ExitGeometryLiteral is called when production geometryLiteral is exited.
func (s *BaseCqlParserListener) ExitGeometryLiteral(ctx *GeometryLiteralContext) {}

// EnterPoint is called when production point is entered.
func (s *BaseCqlParserListener) EnterPoint(ctx *PointContext) {}

// ExitPoint is called when production point is exited.
func (s *BaseCqlParserListener) ExitPoint(ctx *PointContext) {}

// EnterLinestring is called when production linestring is entered.
func (s *BaseCqlParserListener) EnterLinestring(ctx *LinestringContext) {}

// ExitLinestring is called when production linestring is exited.
func (s *BaseCqlParserListener) ExitLinestring(ctx *LinestringContext) {}

// EnterLinestringDef is called when production linestringDef is entered.
func (s *BaseCqlParserListener) EnterLinestringDef(ctx *LinestringDefContext) {}

// ExitLinestringDef is called when production linestringDef is exited.
func (s *BaseCqlParserListener) ExitLinestringDef(ctx *LinestringDefContext) {}

// EnterPolygon is called when production polygon is entered.
func (s *BaseCqlParserListener) EnterPolygon(ctx *PolygonContext) {}

// ExitPolygon is called when production polygon is exited.
func (s *BaseCqlParserListener) ExitPolygon(ctx *PolygonContext) {}

// EnterPolygonDef is called when production polygonDef is entered.
func (s *BaseCqlParserListener) EnterPolygonDef(ctx *PolygonDefContext) {}

// ExitPolygonDef is called when production polygonDef is exited.
func (s *BaseCqlParserListener) ExitPolygonDef(ctx *PolygonDefContext) {}

// EnterMultiPoint is called when production multiPoint is entered.
func (s *BaseCqlParserListener) EnterMultiPoint(ctx *MultiPointContext) {}

// ExitMultiPoint is called when production multiPoint is exited.
func (s *BaseCqlParserListener) ExitMultiPoint(ctx *MultiPointContext) {}

// EnterMultiLinestring is called when production multiLinestring is entered.
func (s *BaseCqlParserListener) EnterMultiLinestring(ctx *MultiLinestringContext) {}

// ExitMultiLinestring is called when production multiLinestring is exited.
func (s *BaseCqlParserListener) ExitMultiLinestring(ctx *MultiLinestringContext) {}

// EnterMultiPolygon is called when production multiPolygon is entered.
func (s *BaseCqlParserListener) EnterMultiPolygon(ctx *MultiPolygonContext) {}

// ExitMultiPolygon is called when production multiPolygon is exited.
func (s *BaseCqlParserListener) ExitMultiPolygon(ctx *MultiPolygonContext) {}

// EnterGeometryCollection is called when production geometryCollection is entered.
func (s *BaseCqlParserListener) EnterGeometryCollection(ctx *GeometryCollectionContext) {}

// ExitGeometryCollection is called when production geometryCollection is exited.
func (s *BaseCqlParserListener) ExitGeometryCollection(ctx *GeometryCollectionContext) {}

// EnterBbox is called when production bbox is entered.
func (s *BaseCqlParserListener) EnterBbox(ctx *BboxContext) {}

// ExitBbox is called when production bbox is exited.
func (s *BaseCqlParserListener) ExitBbox(ctx *BboxContext) {}

// EnterCoordinate is called when production coordinate is entered.
func (s *BaseCqlParserListener) EnterCoordinate(ctx *CoordinateContext) {}

// ExitCoordinate is called when production coordinate is exited.
func (s *BaseCqlParserListener) ExitCoordinate(ctx *CoordinateContext) {}

// EnterXCoord is called when production xCoord is entered.
func (s *BaseCqlParserListener) EnterXCoord(ctx *XCoordContext) {}

// ExitXCoord is called when production xCoord is exited.
func (s *BaseCqlParserListener) ExitXCoord(ctx *XCoordContext) {}

// EnterYCoord is called when production yCoord is entered.
func (s *BaseCqlParserListener) EnterYCoord(ctx *YCoordContext) {}

// ExitYCoord is called when production yCoord is exited.
func (s *BaseCqlParserListener) ExitYCoord(ctx *YCoordContext) {}

// EnterZCoord is called when production zCoord is entered.
func (s *BaseCqlParserListener) EnterZCoord(ctx *ZCoordContext) {}

// ExitZCoord is called when production zCoord is exited.
func (s *BaseCqlParserListener) ExitZCoord(ctx *ZCoordContext) {}

// EnterWestBoundLon is called when production westBoundLon is entered.
func (s *BaseCqlParserListener) EnterWestBoundLon(ctx *WestBoundLonContext) {}

// ExitWestBoundLon is called when production westBoundLon is exited.
func (s *BaseCqlParserListener) ExitWestBoundLon(ctx *WestBoundLonContext) {}

// EnterEastBoundLon is called when production eastBoundLon is entered.
func (s *BaseCqlParserListener) EnterEastBoundLon(ctx *EastBoundLonContext) {}

// ExitEastBoundLon is called when production eastBoundLon is exited.
func (s *BaseCqlParserListener) ExitEastBoundLon(ctx *EastBoundLonContext) {}

// EnterNorthBoundLat is called when production northBoundLat is entered.
func (s *BaseCqlParserListener) EnterNorthBoundLat(ctx *NorthBoundLatContext) {}

// ExitNorthBoundLat is called when production northBoundLat is exited.
func (s *BaseCqlParserListener) ExitNorthBoundLat(ctx *NorthBoundLatContext) {}

// EnterSouthBoundLat is called when production southBoundLat is entered.
func (s *BaseCqlParserListener) EnterSouthBoundLat(ctx *SouthBoundLatContext) {}

// ExitSouthBoundLat is called when production southBoundLat is exited.
func (s *BaseCqlParserListener) ExitSouthBoundLat(ctx *SouthBoundLatContext) {}

// EnterMinElev is called when production minElev is entered.
func (s *BaseCqlParserListener) EnterMinElev(ctx *MinElevContext) {}

// ExitMinElev is called when production minElev is exited.
func (s *BaseCqlParserListener) ExitMinElev(ctx *MinElevContext) {}

// EnterMaxElev is called when production maxElev is entered.
func (s *BaseCqlParserListener) EnterMaxElev(ctx *MaxElevContext) {}

// ExitMaxElev is called when production maxElev is exited.
func (s *BaseCqlParserListener) ExitMaxElev(ctx *MaxElevContext) {}

// EnterTemporalPredicate is called when production temporalPredicate is entered.
func (s *BaseCqlParserListener) EnterTemporalPredicate(ctx *TemporalPredicateContext) {}

// ExitTemporalPredicate is called when production temporalPredicate is exited.
func (s *BaseCqlParserListener) ExitTemporalPredicate(ctx *TemporalPredicateContext) {}

// EnterTemporalExpression is called when production temporalExpression is entered.
func (s *BaseCqlParserListener) EnterTemporalExpression(ctx *TemporalExpressionContext) {}

// ExitTemporalExpression is called when production temporalExpression is exited.
func (s *BaseCqlParserListener) ExitTemporalExpression(ctx *TemporalExpressionContext) {}

// EnterTemporalClause is called when production temporalClause is entered.
func (s *BaseCqlParserListener) EnterTemporalClause(ctx *TemporalClauseContext) {}

// ExitTemporalClause is called when production temporalClause is exited.
func (s *BaseCqlParserListener) ExitTemporalClause(ctx *TemporalClauseContext) {}

// EnterInstantInstance is called when production instantInstance is entered.
func (s *BaseCqlParserListener) EnterInstantInstance(ctx *InstantInstanceContext) {}

// ExitInstantInstance is called when production instantInstance is exited.
func (s *BaseCqlParserListener) ExitInstantInstance(ctx *InstantInstanceContext) {}

// EnterInterval is called when production interval is entered.
func (s *BaseCqlParserListener) EnterInterval(ctx *IntervalContext) {}

// ExitInterval is called when production interval is exited.
func (s *BaseCqlParserListener) ExitInterval(ctx *IntervalContext) {}

// EnterIntervalParameter is called when production intervalParameter is entered.
func (s *BaseCqlParserListener) EnterIntervalParameter(ctx *IntervalParameterContext) {}

// ExitIntervalParameter is called when production intervalParameter is exited.
func (s *BaseCqlParserListener) ExitIntervalParameter(ctx *IntervalParameterContext) {}

// EnterArrayPredicate is called when production arrayPredicate is entered.
func (s *BaseCqlParserListener) EnterArrayPredicate(ctx *ArrayPredicateContext) {}

// ExitArrayPredicate is called when production arrayPredicate is exited.
func (s *BaseCqlParserListener) ExitArrayPredicate(ctx *ArrayPredicateContext) {}

// EnterArrayExpression is called when production arrayExpression is entered.
func (s *BaseCqlParserListener) EnterArrayExpression(ctx *ArrayExpressionContext) {}

// ExitArrayExpression is called when production arrayExpression is exited.
func (s *BaseCqlParserListener) ExitArrayExpression(ctx *ArrayExpressionContext) {}

// EnterArrayClause is called when production arrayClause is entered.
func (s *BaseCqlParserListener) EnterArrayClause(ctx *ArrayClauseContext) {}

// ExitArrayClause is called when production arrayClause is exited.
func (s *BaseCqlParserListener) ExitArrayClause(ctx *ArrayClauseContext) {}

// EnterArrayElement is called when production arrayElement is entered.
func (s *BaseCqlParserListener) EnterArrayElement(ctx *ArrayElementContext) {}

// ExitArrayElement is called when production arrayElement is exited.
func (s *BaseCqlParserListener) ExitArrayElement(ctx *ArrayElementContext) {}

// EnterFunction is called when production function is entered.
func (s *BaseCqlParserListener) EnterFunction(ctx *FunctionContext) {}

// ExitFunction is called when production function is exited.
func (s *BaseCqlParserListener) ExitFunction(ctx *FunctionContext) {}

// EnterArgumentList is called when production argumentList is entered.
func (s *BaseCqlParserListener) EnterArgumentList(ctx *ArgumentListContext) {}

// ExitArgumentList is called when production argumentList is exited.
func (s *BaseCqlParserListener) ExitArgumentList(ctx *ArgumentListContext) {}

// EnterPositionalArgument is called when production positionalArgument is entered.
func (s *BaseCqlParserListener) EnterPositionalArgument(ctx *PositionalArgumentContext) {}

// ExitPositionalArgument is called when production positionalArgument is exited.
func (s *BaseCqlParserListener) ExitPositionalArgument(ctx *PositionalArgumentContext) {}

// EnterArgument is called when production argument is entered.
func (s *BaseCqlParserListener) EnterArgument(ctx *ArgumentContext) {}

// ExitArgument is called when production argument is exited.
func (s *BaseCqlParserListener) ExitArgument(ctx *ArgumentContext) {}
