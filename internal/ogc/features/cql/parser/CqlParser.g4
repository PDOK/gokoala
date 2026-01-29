/*
 * ------------------
 * Note: This file is based on https://github.com/ldproxy/xtraplatform-spatial/blob/482b607f6709389fcd43ebea7dd0434389b8011b/
 * xtraplatform-cql/src/main/antlr/de/ii/xtraplatform/cql/infra/CqlParser.g4
 *
 * Keep the following license header intact:
 * ------------------
 *
 * Copyright interactive instruments GmbH
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */
parser grammar CqlParser;
options { tokenVocab=CqlLexer; }

/*
#=============================================================================#
# A CQL2 filter is a logically connected expression of one or more predicates.
# Predicates include scalar or comparison predicates, spatial predicates or
# temporal predicates.
#
# Arithmetic expressions are not implemented yet.
#=============================================================================#
*/

cqlFilter : booleanExpression EOF;
booleanExpression : booleanTerm (OR booleanTerm)*;
booleanTerm : booleanFactor (AND booleanFactor)*;
booleanFactor : (NOT)? booleanPrimary;
booleanPrimary : predicate
               | booleanLiteral
               | LEFTPAREN booleanExpression RIGHTPAREN
               | function;

/*
#=============================================================================#
# Nested filters are an extension to CQL2
#=============================================================================#
*/

//nestedCqlFilter: {isNotInsideNestedFilter($ctx)}? booleanExpression;

/*
#=============================================================================#
#  CQL2 supports scalar, spatial, temporal and array predicates.
#=============================================================================#
*/

predicate : comparisonPredicate
          | spatialPredicate
          | temporalPredicate
          | arrayPredicate;

/*
#=============================================================================#
# A comparison predicate evaluates if two scalar expression statisfy the
# specified comparison operator.  The comparion operators includes an operator
# to evaluate pattern matching expressions (LIKE), a range evaluation operator
# and an operator to test if a scalar expression is NULL or not.
#=============================================================================#
*/

comparisonPredicate : binaryComparisonPredicate
                    | isLikePredicate
                    | isBetweenPredicate
                    | isInListPredicate
                    | isNullPredicate;

binaryComparisonPredicate : scalarExpression ComparisonOperator scalarExpression;

isLikePredicate :  characterExpression (NOT)? LIKE patternExpression;

isBetweenPredicate : numericExpression (NOT)? BETWEEN numericExpression AND numericExpression;

isInListPredicate : scalarExpression (NOT)? IN LEFTPAREN scalarExpression (COMMA scalarExpression)* RIGHTPAREN;

isNullPredicate : isNullOperand IS (NOT)? NULL;

isNullOperand : characterClause
              | numericLiteral
              | booleanLiteral
              | instantInstance
              | geometryLiteral
              | propertyName
              | function;
/*
#=============================================================================#
# Scalar expressions
#=============================================================================#
*/

scalarExpression : characterClause
                 | numericLiteral
                 | instantInstance
                 /*
                 | arithmeticExpression
                 */
                 | booleanLiteral
                 | propertyName
                 | function;

characterExpression : characterClause
                    | propertyName
                    | function;

patternExpression : characterLiteral
                  | CASEI LEFTPAREN patternExpression RIGHTPAREN
                  | ACCENTI LEFTPAREN patternExpression RIGHTPAREN
                  // UPPER() and LOWER() are extensions to CQL2
                  | LOWER LEFTPAREN patternExpression RIGHTPAREN
                  | UPPER LEFTPAREN patternExpression RIGHTPAREN;

characterClause : characterLiteral
                | CASEI LEFTPAREN characterExpression RIGHTPAREN
                | ACCENTI LEFTPAREN characterExpression RIGHTPAREN
                // UPPER() and LOWER() are extensions to CQL2
                | LOWER LEFTPAREN characterExpression RIGHTPAREN
                | UPPER LEFTPAREN characterExpression RIGHTPAREN;

characterLiteral : CharacterStringLiteral;

numericExpression : numericLiteral
                  | propertyName
                  | function
                  /*
                  | arithmeticExpression
                  */;

numericLiteral : NumericLiteral;

booleanLiteral : BooleanLiteral;

// Support for compound property names is a CQL2 extension
// Support for nested filters is a CQL2 extension
//propertyName : (Identifier (LEFTSQUAREBRACKET nestedCqlFilter RIGHTSQUAREBRACKET)? PERIOD)* Identifier;
propertyName: Identifier;

/*
#=============================================================================#
# A spatial predicate evaluates if two spatial expressions satisfy the
# condition implied by a standardized spatial comparison function.  If the
# conditions of the spatial comparison function are met, the function returns
# a Boolean value of true.  Otherwise the function returns false.
#=============================================================================#
*/

spatialPredicate :  SpatialFunction LEFTPAREN geomExpression COMMA geomExpression RIGHTPAREN;

/*
# A geometric expression is a property name of a geometry-valued property,
# a geometric literal (expressed as WKT) or a function that returns a
# geometric value.
*/

geomExpression : spatialInstance
               | propertyName
               | function;

/*
#=============================================================================#
# Definition of GEOMETRIC literals
#
# NOTE: This is basically BNF that define WKT encoding; it would be nice
#       to instead reference some normative BNF for WKT.
#=============================================================================#
*/

spatialInstance : geometryLiteral
                | geometryCollection
                | bbox;

geometryLiteral : point
                | linestring
                | polygon
                | multiPoint
                | multiLinestring
                | multiPolygon;

point : POINT LEFTPAREN coordinate RIGHTPAREN;

linestring : LINESTRING linestringDef;

linestringDef: LEFTPAREN coordinate (COMMA coordinate)* RIGHTPAREN;

polygon : POLYGON polygonDef;

polygonDef : LEFTPAREN linestringDef (COMMA linestringDef)* RIGHTPAREN;

multiPoint : MULTIPOINT LEFTPAREN coordinate (COMMA coordinate)* RIGHTPAREN;

multiLinestring : MULTILINESTRING LEFTPAREN linestringDef (COMMA linestringDef)* RIGHTPAREN;

multiPolygon : MULTIPOLYGON LEFTPAREN polygonDef (COMMA polygonDef)* RIGHTPAREN;

geometryCollection : GEOMETRYCOLLECTION LEFTPAREN geometryLiteral (COMMA geometryLiteral)* RIGHTPAREN;

bbox: BBOX LEFTPAREN westBoundLon COMMA southBoundLat COMMA (minElev COMMA)? eastBoundLon  COMMA northBoundLat (COMMA maxElev)? RIGHTPAREN;

coordinate : xCoord yCoord (zCoord)?;

xCoord : NumericLiteral;

yCoord : NumericLiteral;

zCoord : NumericLiteral;

westBoundLon : NumericLiteral;

eastBoundLon : NumericLiteral;

northBoundLat : NumericLiteral;

southBoundLat : NumericLiteral;

minElev : NumericLiteral;

maxElev : NumericLiteral;

/*
#=============================================================================#
# A temporal predicate evaluates if two temporal expressions satisfy the
# specified temporal operator.
#=============================================================================#
*/

temporalPredicate : TemporalFunction LEFTPAREN temporalExpression COMMA temporalExpression RIGHTPAREN;

temporalExpression : temporalClause
                   | propertyName
                   | function;

temporalClause: instantInstance | interval;

instantInstance: DATE LEFTPAREN DateString RIGHTPAREN
               | TIMESTAMP LEFTPAREN TimestampString RIGHTPAREN
               | NOW LEFTPAREN RIGHTPAREN; // NOW() is a CQL2 extension

interval: INTERVAL LEFTPAREN intervalParameter COMMA intervalParameter RIGHTPAREN;

intervalParameter: propertyName
                 | DateString
                 | TimestampString
                 | NOW LEFTPAREN RIGHTPAREN // NOW() is a CQL2 extension
                 | DotDotString
                 | function;

/*
#=============================================================================#
# An array predicate evaluates if two array expressions statisfy the
# specified comparison operator.  The comparion operators include equality,
# not equal, less than, greater than, less than or equal, greater than or equal,
# superset, subset and overlap operators.
#=============================================================================#
*/

arrayPredicate: ArrayFunction LEFTPAREN arrayExpression COMMA arrayExpression RIGHTPAREN;

arrayExpression: propertyName
               | arrayClause
               | function;

arrayClause: LEFTPAREN RIGHTPAREN
           | LEFTPAREN arrayElement ( COMMA arrayElement )* RIGHTPAREN;

arrayElement: characterClause
            | numericLiteral
            | booleanLiteral
            | temporalClause
            | arrayClause
            | propertyName
            | function
            /*
            | arithmeticExpression
            */;

/*
#=============================================================================#
# Definition of a FUNCTION
#=============================================================================#
*/

function : Identifier argumentList;

argumentList : LEFTPAREN (positionalArgument)?  RIGHTPAREN;

positionalArgument : argument ( COMMA argument )*;

argument : characterClause
         | numericLiteral
         | booleanLiteral
         | geometryLiteral
         | temporalClause
         | arrayClause
         | propertyName
         | function
         /*
         | arithmeticExpression
         */;

/*
#=============================================================================#
# An arithemtic expression is an expression composed of an arithmetic
# operand (a property name, a number or a function that returns a number),
# an arithmetic operators (+,-,*,/) and another arithmetic operand.
#=============================================================================#
*/
/* Unsupported for now

arithmeticExpression : arithmeticTerm (arithmeticOperatorPlusMinus arithmeticTerm)?;

arithmeticOperatorPlusMinus : PLUS | MINUS;

arithmeticTerm : powerTerm (arithmeticOperatorMultDiv powerTerm)?;

arithmeticOperatorMultDiv : ASTERISK | SOLIDUS | PERCENT | DIV;

powerTerm : arithmeticFactor (CARET arithmeticFactor)?;

arithmeticFactor : LEFTPAREN arithmeticExpression RIGHTPAREN
                 | (MINUS)? arithmeticOperand;

arithmeticOperand : numericLiteral
                  | propertyName
                  | function;
*/