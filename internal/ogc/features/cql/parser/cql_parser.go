// Code generated from CqlParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // CqlParser
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type CqlParser struct {
	*antlr.BaseParser
}

var CqlParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func cqlparserParserInit() {
	staticData := &CqlParserParserStaticData
	staticData.LiteralNames = []string{
		"", "", "'<'", "'='", "'>'", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "'$'", "'_'", "'\"'", "", "'('", "')'", "'['", "']'", "'*'",
		"'+'", "','", "'^'", "'-'", "'.'", "'/'", "':'", "'%'", "", "", "",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "", "", "", "", "''''",
	}
	staticData.SymbolicNames = []string{
		"", "ComparisonOperator", "LT", "EQ", "GT", "NEQ", "GTEQ", "LTEQ", "BooleanLiteral",
		"AND", "OR", "NOT", "LIKE", "BETWEEN", "IS", "NULL", "SpatialFunction",
		"TemporalFunction", "ArrayFunction", "IN", "POINT", "LINESTRING", "POLYGON",
		"MULTIPOINT", "MULTILINESTRING", "MULTIPOLYGON", "GEOMETRYCOLLECTION",
		"BBOX", "CASEI", "ACCENTI", "LOWER", "UPPER", "NumericLiteral", "DIGIT",
		"DOLLAR", "UNDERSCORE", "DOUBLEQUOTE", "QUOTE", "LEFTPAREN", "RIGHTPAREN",
		"LEFTSQUAREBRACKET", "RIGHTSQUAREBRACKET", "ASTERISK", "PLUS", "COMMA",
		"CARET", "MINUS", "PERIOD", "SOLIDUS", "COLON", "PERCENT", "DIV", "ALPHA",
		"IdentifierStart", "IdentifierPart", "UnsignedNumericLiteral", "SignedNumericLiteral",
		"ExactNumericLiteral", "ApproximateNumericLiteral", "Mantissa", "Exponent",
		"SignedInteger", "UnsignedInteger", "Sign", "DATE", "TIMESTAMP", "INTERVAL",
		"DateString", "TimestampString", "DotDotString", "Instant", "FullDate",
		"DateYear", "DateMonth", "DateDay", "UtcTime", "TimeHour", "TimeMinute",
		"TimeSecond", "Identifier", "IdentifierBare", "WS", "CharacterStringLiteral",
		"QuotedQuote",
	}
	staticData.RuleNames = []string{
		"cqlFilter", "booleanExpression", "booleanTerm", "booleanFactor", "booleanPrimary",
		"predicate", "comparisonPredicate", "binaryComparisonPredicate", "isLikePredicate",
		"isBetweenPredicate", "isInListPredicate", "isNullPredicate", "isNullOperand",
		"scalarExpression", "characterExpression", "patternExpression", "characterClause",
		"characterLiteral", "numericExpression", "numericLiteral", "booleanLiteral",
		"propertyName", "spatialPredicate", "geomExpression", "spatialInstance",
		"geometryLiteral", "point", "linestring", "linestringDef", "polygon",
		"polygonDef", "multiPoint", "multiPointDef", "multiLinestring", "multiPolygon",
		"geometryCollection", "bbox", "coordinate", "xCoord", "yCoord", "zCoord",
		"westBoundLon", "eastBoundLon", "northBoundLat", "southBoundLat", "minElev",
		"maxElev", "temporalPredicate", "temporalExpression", "temporalClause",
		"instantInstance", "interval", "intervalParameter", "arrayPredicate",
		"arrayExpression", "arrayClause", "arrayElement", "function", "argumentList",
		"positionalArgument", "argument",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 83, 536, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2, 31, 7,
		31, 2, 32, 7, 32, 2, 33, 7, 33, 2, 34, 7, 34, 2, 35, 7, 35, 2, 36, 7, 36,
		2, 37, 7, 37, 2, 38, 7, 38, 2, 39, 7, 39, 2, 40, 7, 40, 2, 41, 7, 41, 2,
		42, 7, 42, 2, 43, 7, 43, 2, 44, 7, 44, 2, 45, 7, 45, 2, 46, 7, 46, 2, 47,
		7, 47, 2, 48, 7, 48, 2, 49, 7, 49, 2, 50, 7, 50, 2, 51, 7, 51, 2, 52, 7,
		52, 2, 53, 7, 53, 2, 54, 7, 54, 2, 55, 7, 55, 2, 56, 7, 56, 2, 57, 7, 57,
		2, 58, 7, 58, 2, 59, 7, 59, 2, 60, 7, 60, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1,
		1, 1, 5, 1, 129, 8, 1, 10, 1, 12, 1, 132, 9, 1, 1, 2, 1, 2, 1, 2, 5, 2,
		137, 8, 2, 10, 2, 12, 2, 140, 9, 2, 1, 3, 3, 3, 143, 8, 3, 1, 3, 1, 3,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 3, 4, 154, 8, 4, 1, 5, 1, 5,
		1, 5, 1, 5, 3, 5, 160, 8, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 3, 6, 167, 8,
		6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 3, 8, 175, 8, 8, 1, 8, 1, 8, 1,
		8, 1, 9, 1, 9, 3, 9, 182, 8, 9, 1, 9, 1, 9, 1, 9, 1, 9, 1, 9, 1, 10, 1,
		10, 3, 10, 191, 8, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 5, 10, 198, 8,
		10, 10, 10, 12, 10, 201, 9, 10, 1, 10, 1, 10, 1, 11, 1, 11, 1, 11, 3, 11,
		208, 8, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1,
		12, 3, 12, 219, 8, 12, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 3, 13,
		227, 8, 13, 1, 14, 1, 14, 1, 14, 3, 14, 232, 8, 14, 1, 15, 1, 15, 1, 15,
		1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 3, 15, 245, 8,
		15, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16,
		1, 16, 3, 16, 258, 8, 16, 1, 17, 1, 17, 1, 18, 1, 18, 1, 18, 3, 18, 265,
		8, 18, 1, 19, 1, 19, 1, 20, 1, 20, 1, 21, 1, 21, 1, 22, 1, 22, 1, 22, 1,
		22, 1, 22, 1, 22, 1, 22, 1, 23, 1, 23, 1, 23, 3, 23, 283, 8, 23, 1, 24,
		1, 24, 1, 24, 3, 24, 288, 8, 24, 1, 25, 1, 25, 1, 25, 1, 25, 1, 25, 1,
		25, 3, 25, 296, 8, 25, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 27, 1, 27,
		1, 27, 1, 28, 1, 28, 1, 28, 1, 28, 5, 28, 310, 8, 28, 10, 28, 12, 28, 313,
		9, 28, 1, 28, 1, 28, 1, 29, 1, 29, 1, 29, 1, 30, 1, 30, 1, 30, 1, 30, 5,
		30, 324, 8, 30, 10, 30, 12, 30, 327, 9, 30, 1, 30, 1, 30, 1, 31, 1, 31,
		1, 31, 1, 31, 1, 31, 5, 31, 336, 8, 31, 10, 31, 12, 31, 339, 9, 31, 1,
		31, 1, 31, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32, 3, 32, 348, 8, 32, 1, 33,
		1, 33, 1, 33, 1, 33, 1, 33, 5, 33, 355, 8, 33, 10, 33, 12, 33, 358, 9,
		33, 1, 33, 1, 33, 1, 34, 1, 34, 1, 34, 1, 34, 1, 34, 5, 34, 367, 8, 34,
		10, 34, 12, 34, 370, 9, 34, 1, 34, 1, 34, 1, 35, 1, 35, 1, 35, 1, 35, 1,
		35, 5, 35, 379, 8, 35, 10, 35, 12, 35, 382, 9, 35, 1, 35, 1, 35, 1, 36,
		1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 3, 36, 395, 8,
		36, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 3, 36, 402, 8, 36, 1, 36, 1, 36,
		1, 37, 1, 37, 1, 37, 3, 37, 409, 8, 37, 1, 38, 1, 38, 1, 39, 1, 39, 1,
		40, 1, 40, 1, 41, 1, 41, 1, 42, 1, 42, 1, 43, 1, 43, 1, 44, 1, 44, 1, 45,
		1, 45, 1, 46, 1, 46, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 1,
		48, 1, 48, 1, 48, 3, 48, 439, 8, 48, 1, 49, 1, 49, 3, 49, 443, 8, 49, 1,
		50, 1, 50, 1, 50, 1, 50, 1, 50, 1, 50, 1, 50, 1, 50, 1, 50, 1, 50, 3, 50,
		455, 8, 50, 1, 51, 1, 51, 1, 51, 1, 51, 1, 51, 1, 51, 1, 51, 1, 52, 1,
		52, 1, 52, 1, 52, 1, 52, 1, 52, 1, 52, 3, 52, 471, 8, 52, 1, 53, 1, 53,
		1, 53, 1, 53, 1, 53, 1, 53, 1, 53, 1, 54, 1, 54, 1, 54, 3, 54, 483, 8,
		54, 1, 55, 1, 55, 1, 55, 1, 55, 1, 55, 1, 55, 5, 55, 491, 8, 55, 10, 55,
		12, 55, 494, 9, 55, 1, 55, 1, 55, 3, 55, 498, 8, 55, 1, 56, 1, 56, 1, 56,
		1, 56, 1, 56, 1, 56, 1, 56, 3, 56, 507, 8, 56, 1, 57, 1, 57, 1, 57, 1,
		58, 1, 58, 3, 58, 514, 8, 58, 1, 58, 1, 58, 1, 59, 1, 59, 1, 59, 5, 59,
		521, 8, 59, 10, 59, 12, 59, 524, 9, 59, 1, 60, 1, 60, 1, 60, 1, 60, 1,
		60, 1, 60, 1, 60, 1, 60, 3, 60, 534, 8, 60, 1, 60, 0, 0, 61, 0, 2, 4, 6,
		8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42,
		44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 66, 68, 70, 72, 74, 76, 78,
		80, 82, 84, 86, 88, 90, 92, 94, 96, 98, 100, 102, 104, 106, 108, 110, 112,
		114, 116, 118, 120, 0, 0, 559, 0, 122, 1, 0, 0, 0, 2, 125, 1, 0, 0, 0,
		4, 133, 1, 0, 0, 0, 6, 142, 1, 0, 0, 0, 8, 153, 1, 0, 0, 0, 10, 159, 1,
		0, 0, 0, 12, 166, 1, 0, 0, 0, 14, 168, 1, 0, 0, 0, 16, 172, 1, 0, 0, 0,
		18, 179, 1, 0, 0, 0, 20, 188, 1, 0, 0, 0, 22, 204, 1, 0, 0, 0, 24, 218,
		1, 0, 0, 0, 26, 226, 1, 0, 0, 0, 28, 231, 1, 0, 0, 0, 30, 244, 1, 0, 0,
		0, 32, 257, 1, 0, 0, 0, 34, 259, 1, 0, 0, 0, 36, 264, 1, 0, 0, 0, 38, 266,
		1, 0, 0, 0, 40, 268, 1, 0, 0, 0, 42, 270, 1, 0, 0, 0, 44, 272, 1, 0, 0,
		0, 46, 282, 1, 0, 0, 0, 48, 287, 1, 0, 0, 0, 50, 295, 1, 0, 0, 0, 52, 297,
		1, 0, 0, 0, 54, 302, 1, 0, 0, 0, 56, 305, 1, 0, 0, 0, 58, 316, 1, 0, 0,
		0, 60, 319, 1, 0, 0, 0, 62, 330, 1, 0, 0, 0, 64, 347, 1, 0, 0, 0, 66, 349,
		1, 0, 0, 0, 68, 361, 1, 0, 0, 0, 70, 373, 1, 0, 0, 0, 72, 385, 1, 0, 0,
		0, 74, 405, 1, 0, 0, 0, 76, 410, 1, 0, 0, 0, 78, 412, 1, 0, 0, 0, 80, 414,
		1, 0, 0, 0, 82, 416, 1, 0, 0, 0, 84, 418, 1, 0, 0, 0, 86, 420, 1, 0, 0,
		0, 88, 422, 1, 0, 0, 0, 90, 424, 1, 0, 0, 0, 92, 426, 1, 0, 0, 0, 94, 428,
		1, 0, 0, 0, 96, 438, 1, 0, 0, 0, 98, 442, 1, 0, 0, 0, 100, 454, 1, 0, 0,
		0, 102, 456, 1, 0, 0, 0, 104, 470, 1, 0, 0, 0, 106, 472, 1, 0, 0, 0, 108,
		482, 1, 0, 0, 0, 110, 497, 1, 0, 0, 0, 112, 506, 1, 0, 0, 0, 114, 508,
		1, 0, 0, 0, 116, 511, 1, 0, 0, 0, 118, 517, 1, 0, 0, 0, 120, 533, 1, 0,
		0, 0, 122, 123, 3, 2, 1, 0, 123, 124, 5, 0, 0, 1, 124, 1, 1, 0, 0, 0, 125,
		130, 3, 4, 2, 0, 126, 127, 5, 10, 0, 0, 127, 129, 3, 4, 2, 0, 128, 126,
		1, 0, 0, 0, 129, 132, 1, 0, 0, 0, 130, 128, 1, 0, 0, 0, 130, 131, 1, 0,
		0, 0, 131, 3, 1, 0, 0, 0, 132, 130, 1, 0, 0, 0, 133, 138, 3, 6, 3, 0, 134,
		135, 5, 9, 0, 0, 135, 137, 3, 6, 3, 0, 136, 134, 1, 0, 0, 0, 137, 140,
		1, 0, 0, 0, 138, 136, 1, 0, 0, 0, 138, 139, 1, 0, 0, 0, 139, 5, 1, 0, 0,
		0, 140, 138, 1, 0, 0, 0, 141, 143, 5, 11, 0, 0, 142, 141, 1, 0, 0, 0, 142,
		143, 1, 0, 0, 0, 143, 144, 1, 0, 0, 0, 144, 145, 3, 8, 4, 0, 145, 7, 1,
		0, 0, 0, 146, 154, 3, 10, 5, 0, 147, 154, 3, 40, 20, 0, 148, 149, 5, 38,
		0, 0, 149, 150, 3, 2, 1, 0, 150, 151, 5, 39, 0, 0, 151, 154, 1, 0, 0, 0,
		152, 154, 3, 114, 57, 0, 153, 146, 1, 0, 0, 0, 153, 147, 1, 0, 0, 0, 153,
		148, 1, 0, 0, 0, 153, 152, 1, 0, 0, 0, 154, 9, 1, 0, 0, 0, 155, 160, 3,
		12, 6, 0, 156, 160, 3, 44, 22, 0, 157, 160, 3, 94, 47, 0, 158, 160, 3,
		106, 53, 0, 159, 155, 1, 0, 0, 0, 159, 156, 1, 0, 0, 0, 159, 157, 1, 0,
		0, 0, 159, 158, 1, 0, 0, 0, 160, 11, 1, 0, 0, 0, 161, 167, 3, 14, 7, 0,
		162, 167, 3, 16, 8, 0, 163, 167, 3, 18, 9, 0, 164, 167, 3, 20, 10, 0, 165,
		167, 3, 22, 11, 0, 166, 161, 1, 0, 0, 0, 166, 162, 1, 0, 0, 0, 166, 163,
		1, 0, 0, 0, 166, 164, 1, 0, 0, 0, 166, 165, 1, 0, 0, 0, 167, 13, 1, 0,
		0, 0, 168, 169, 3, 26, 13, 0, 169, 170, 5, 1, 0, 0, 170, 171, 3, 26, 13,
		0, 171, 15, 1, 0, 0, 0, 172, 174, 3, 28, 14, 0, 173, 175, 5, 11, 0, 0,
		174, 173, 1, 0, 0, 0, 174, 175, 1, 0, 0, 0, 175, 176, 1, 0, 0, 0, 176,
		177, 5, 12, 0, 0, 177, 178, 3, 30, 15, 0, 178, 17, 1, 0, 0, 0, 179, 181,
		3, 36, 18, 0, 180, 182, 5, 11, 0, 0, 181, 180, 1, 0, 0, 0, 181, 182, 1,
		0, 0, 0, 182, 183, 1, 0, 0, 0, 183, 184, 5, 13, 0, 0, 184, 185, 3, 36,
		18, 0, 185, 186, 5, 9, 0, 0, 186, 187, 3, 36, 18, 0, 187, 19, 1, 0, 0,
		0, 188, 190, 3, 26, 13, 0, 189, 191, 5, 11, 0, 0, 190, 189, 1, 0, 0, 0,
		190, 191, 1, 0, 0, 0, 191, 192, 1, 0, 0, 0, 192, 193, 5, 19, 0, 0, 193,
		194, 5, 38, 0, 0, 194, 199, 3, 26, 13, 0, 195, 196, 5, 44, 0, 0, 196, 198,
		3, 26, 13, 0, 197, 195, 1, 0, 0, 0, 198, 201, 1, 0, 0, 0, 199, 197, 1,
		0, 0, 0, 199, 200, 1, 0, 0, 0, 200, 202, 1, 0, 0, 0, 201, 199, 1, 0, 0,
		0, 202, 203, 5, 39, 0, 0, 203, 21, 1, 0, 0, 0, 204, 205, 3, 24, 12, 0,
		205, 207, 5, 14, 0, 0, 206, 208, 5, 11, 0, 0, 207, 206, 1, 0, 0, 0, 207,
		208, 1, 0, 0, 0, 208, 209, 1, 0, 0, 0, 209, 210, 5, 15, 0, 0, 210, 23,
		1, 0, 0, 0, 211, 219, 3, 32, 16, 0, 212, 219, 3, 38, 19, 0, 213, 219, 3,
		40, 20, 0, 214, 219, 3, 100, 50, 0, 215, 219, 3, 50, 25, 0, 216, 219, 3,
		42, 21, 0, 217, 219, 3, 114, 57, 0, 218, 211, 1, 0, 0, 0, 218, 212, 1,
		0, 0, 0, 218, 213, 1, 0, 0, 0, 218, 214, 1, 0, 0, 0, 218, 215, 1, 0, 0,
		0, 218, 216, 1, 0, 0, 0, 218, 217, 1, 0, 0, 0, 219, 25, 1, 0, 0, 0, 220,
		227, 3, 32, 16, 0, 221, 227, 3, 38, 19, 0, 222, 227, 3, 100, 50, 0, 223,
		227, 3, 40, 20, 0, 224, 227, 3, 42, 21, 0, 225, 227, 3, 114, 57, 0, 226,
		220, 1, 0, 0, 0, 226, 221, 1, 0, 0, 0, 226, 222, 1, 0, 0, 0, 226, 223,
		1, 0, 0, 0, 226, 224, 1, 0, 0, 0, 226, 225, 1, 0, 0, 0, 227, 27, 1, 0,
		0, 0, 228, 232, 3, 32, 16, 0, 229, 232, 3, 42, 21, 0, 230, 232, 3, 114,
		57, 0, 231, 228, 1, 0, 0, 0, 231, 229, 1, 0, 0, 0, 231, 230, 1, 0, 0, 0,
		232, 29, 1, 0, 0, 0, 233, 245, 3, 34, 17, 0, 234, 235, 5, 28, 0, 0, 235,
		236, 5, 38, 0, 0, 236, 237, 3, 30, 15, 0, 237, 238, 5, 39, 0, 0, 238, 245,
		1, 0, 0, 0, 239, 240, 5, 29, 0, 0, 240, 241, 5, 38, 0, 0, 241, 242, 3,
		30, 15, 0, 242, 243, 5, 39, 0, 0, 243, 245, 1, 0, 0, 0, 244, 233, 1, 0,
		0, 0, 244, 234, 1, 0, 0, 0, 244, 239, 1, 0, 0, 0, 245, 31, 1, 0, 0, 0,
		246, 258, 3, 34, 17, 0, 247, 248, 5, 28, 0, 0, 248, 249, 5, 38, 0, 0, 249,
		250, 3, 28, 14, 0, 250, 251, 5, 39, 0, 0, 251, 258, 1, 0, 0, 0, 252, 253,
		5, 29, 0, 0, 253, 254, 5, 38, 0, 0, 254, 255, 3, 28, 14, 0, 255, 256, 5,
		39, 0, 0, 256, 258, 1, 0, 0, 0, 257, 246, 1, 0, 0, 0, 257, 247, 1, 0, 0,
		0, 257, 252, 1, 0, 0, 0, 258, 33, 1, 0, 0, 0, 259, 260, 5, 82, 0, 0, 260,
		35, 1, 0, 0, 0, 261, 265, 3, 38, 19, 0, 262, 265, 3, 42, 21, 0, 263, 265,
		3, 114, 57, 0, 264, 261, 1, 0, 0, 0, 264, 262, 1, 0, 0, 0, 264, 263, 1,
		0, 0, 0, 265, 37, 1, 0, 0, 0, 266, 267, 5, 32, 0, 0, 267, 39, 1, 0, 0,
		0, 268, 269, 5, 8, 0, 0, 269, 41, 1, 0, 0, 0, 270, 271, 5, 79, 0, 0, 271,
		43, 1, 0, 0, 0, 272, 273, 5, 16, 0, 0, 273, 274, 5, 38, 0, 0, 274, 275,
		3, 46, 23, 0, 275, 276, 5, 44, 0, 0, 276, 277, 3, 46, 23, 0, 277, 278,
		5, 39, 0, 0, 278, 45, 1, 0, 0, 0, 279, 283, 3, 48, 24, 0, 280, 283, 3,
		42, 21, 0, 281, 283, 3, 114, 57, 0, 282, 279, 1, 0, 0, 0, 282, 280, 1,
		0, 0, 0, 282, 281, 1, 0, 0, 0, 283, 47, 1, 0, 0, 0, 284, 288, 3, 50, 25,
		0, 285, 288, 3, 70, 35, 0, 286, 288, 3, 72, 36, 0, 287, 284, 1, 0, 0, 0,
		287, 285, 1, 0, 0, 0, 287, 286, 1, 0, 0, 0, 288, 49, 1, 0, 0, 0, 289, 296,
		3, 52, 26, 0, 290, 296, 3, 54, 27, 0, 291, 296, 3, 58, 29, 0, 292, 296,
		3, 62, 31, 0, 293, 296, 3, 66, 33, 0, 294, 296, 3, 68, 34, 0, 295, 289,
		1, 0, 0, 0, 295, 290, 1, 0, 0, 0, 295, 291, 1, 0, 0, 0, 295, 292, 1, 0,
		0, 0, 295, 293, 1, 0, 0, 0, 295, 294, 1, 0, 0, 0, 296, 51, 1, 0, 0, 0,
		297, 298, 5, 20, 0, 0, 298, 299, 5, 38, 0, 0, 299, 300, 3, 74, 37, 0, 300,
		301, 5, 39, 0, 0, 301, 53, 1, 0, 0, 0, 302, 303, 5, 21, 0, 0, 303, 304,
		3, 56, 28, 0, 304, 55, 1, 0, 0, 0, 305, 306, 5, 38, 0, 0, 306, 311, 3,
		74, 37, 0, 307, 308, 5, 44, 0, 0, 308, 310, 3, 74, 37, 0, 309, 307, 1,
		0, 0, 0, 310, 313, 1, 0, 0, 0, 311, 309, 1, 0, 0, 0, 311, 312, 1, 0, 0,
		0, 312, 314, 1, 0, 0, 0, 313, 311, 1, 0, 0, 0, 314, 315, 5, 39, 0, 0, 315,
		57, 1, 0, 0, 0, 316, 317, 5, 22, 0, 0, 317, 318, 3, 60, 30, 0, 318, 59,
		1, 0, 0, 0, 319, 320, 5, 38, 0, 0, 320, 325, 3, 56, 28, 0, 321, 322, 5,
		44, 0, 0, 322, 324, 3, 56, 28, 0, 323, 321, 1, 0, 0, 0, 324, 327, 1, 0,
		0, 0, 325, 323, 1, 0, 0, 0, 325, 326, 1, 0, 0, 0, 326, 328, 1, 0, 0, 0,
		327, 325, 1, 0, 0, 0, 328, 329, 5, 39, 0, 0, 329, 61, 1, 0, 0, 0, 330,
		331, 5, 23, 0, 0, 331, 332, 5, 38, 0, 0, 332, 337, 3, 64, 32, 0, 333, 334,
		5, 44, 0, 0, 334, 336, 3, 64, 32, 0, 335, 333, 1, 0, 0, 0, 336, 339, 1,
		0, 0, 0, 337, 335, 1, 0, 0, 0, 337, 338, 1, 0, 0, 0, 338, 340, 1, 0, 0,
		0, 339, 337, 1, 0, 0, 0, 340, 341, 5, 39, 0, 0, 341, 63, 1, 0, 0, 0, 342,
		343, 5, 38, 0, 0, 343, 344, 3, 74, 37, 0, 344, 345, 5, 39, 0, 0, 345, 348,
		1, 0, 0, 0, 346, 348, 3, 74, 37, 0, 347, 342, 1, 0, 0, 0, 347, 346, 1,
		0, 0, 0, 348, 65, 1, 0, 0, 0, 349, 350, 5, 24, 0, 0, 350, 351, 5, 38, 0,
		0, 351, 356, 3, 56, 28, 0, 352, 353, 5, 44, 0, 0, 353, 355, 3, 56, 28,
		0, 354, 352, 1, 0, 0, 0, 355, 358, 1, 0, 0, 0, 356, 354, 1, 0, 0, 0, 356,
		357, 1, 0, 0, 0, 357, 359, 1, 0, 0, 0, 358, 356, 1, 0, 0, 0, 359, 360,
		5, 39, 0, 0, 360, 67, 1, 0, 0, 0, 361, 362, 5, 25, 0, 0, 362, 363, 5, 38,
		0, 0, 363, 368, 3, 60, 30, 0, 364, 365, 5, 44, 0, 0, 365, 367, 3, 60, 30,
		0, 366, 364, 1, 0, 0, 0, 367, 370, 1, 0, 0, 0, 368, 366, 1, 0, 0, 0, 368,
		369, 1, 0, 0, 0, 369, 371, 1, 0, 0, 0, 370, 368, 1, 0, 0, 0, 371, 372,
		5, 39, 0, 0, 372, 69, 1, 0, 0, 0, 373, 374, 5, 26, 0, 0, 374, 375, 5, 38,
		0, 0, 375, 380, 3, 50, 25, 0, 376, 377, 5, 44, 0, 0, 377, 379, 3, 50, 25,
		0, 378, 376, 1, 0, 0, 0, 379, 382, 1, 0, 0, 0, 380, 378, 1, 0, 0, 0, 380,
		381, 1, 0, 0, 0, 381, 383, 1, 0, 0, 0, 382, 380, 1, 0, 0, 0, 383, 384,
		5, 39, 0, 0, 384, 71, 1, 0, 0, 0, 385, 386, 5, 27, 0, 0, 386, 387, 5, 38,
		0, 0, 387, 388, 3, 82, 41, 0, 388, 389, 5, 44, 0, 0, 389, 390, 3, 88, 44,
		0, 390, 394, 5, 44, 0, 0, 391, 392, 3, 90, 45, 0, 392, 393, 5, 44, 0, 0,
		393, 395, 1, 0, 0, 0, 394, 391, 1, 0, 0, 0, 394, 395, 1, 0, 0, 0, 395,
		396, 1, 0, 0, 0, 396, 397, 3, 84, 42, 0, 397, 398, 5, 44, 0, 0, 398, 401,
		3, 86, 43, 0, 399, 400, 5, 44, 0, 0, 400, 402, 3, 92, 46, 0, 401, 399,
		1, 0, 0, 0, 401, 402, 1, 0, 0, 0, 402, 403, 1, 0, 0, 0, 403, 404, 5, 39,
		0, 0, 404, 73, 1, 0, 0, 0, 405, 406, 3, 76, 38, 0, 406, 408, 3, 78, 39,
		0, 407, 409, 3, 80, 40, 0, 408, 407, 1, 0, 0, 0, 408, 409, 1, 0, 0, 0,
		409, 75, 1, 0, 0, 0, 410, 411, 5, 32, 0, 0, 411, 77, 1, 0, 0, 0, 412, 413,
		5, 32, 0, 0, 413, 79, 1, 0, 0, 0, 414, 415, 5, 32, 0, 0, 415, 81, 1, 0,
		0, 0, 416, 417, 5, 32, 0, 0, 417, 83, 1, 0, 0, 0, 418, 419, 5, 32, 0, 0,
		419, 85, 1, 0, 0, 0, 420, 421, 5, 32, 0, 0, 421, 87, 1, 0, 0, 0, 422, 423,
		5, 32, 0, 0, 423, 89, 1, 0, 0, 0, 424, 425, 5, 32, 0, 0, 425, 91, 1, 0,
		0, 0, 426, 427, 5, 32, 0, 0, 427, 93, 1, 0, 0, 0, 428, 429, 5, 17, 0, 0,
		429, 430, 5, 38, 0, 0, 430, 431, 3, 96, 48, 0, 431, 432, 5, 44, 0, 0, 432,
		433, 3, 96, 48, 0, 433, 434, 5, 39, 0, 0, 434, 95, 1, 0, 0, 0, 435, 439,
		3, 98, 49, 0, 436, 439, 3, 42, 21, 0, 437, 439, 3, 114, 57, 0, 438, 435,
		1, 0, 0, 0, 438, 436, 1, 0, 0, 0, 438, 437, 1, 0, 0, 0, 439, 97, 1, 0,
		0, 0, 440, 443, 3, 100, 50, 0, 441, 443, 3, 102, 51, 0, 442, 440, 1, 0,
		0, 0, 442, 441, 1, 0, 0, 0, 443, 99, 1, 0, 0, 0, 444, 445, 5, 64, 0, 0,
		445, 446, 5, 38, 0, 0, 446, 447, 5, 67, 0, 0, 447, 455, 5, 39, 0, 0, 448,
		449, 5, 65, 0, 0, 449, 450, 5, 38, 0, 0, 450, 451, 5, 68, 0, 0, 451, 455,
		5, 39, 0, 0, 452, 453, 5, 38, 0, 0, 453, 455, 5, 39, 0, 0, 454, 444, 1,
		0, 0, 0, 454, 448, 1, 0, 0, 0, 454, 452, 1, 0, 0, 0, 455, 101, 1, 0, 0,
		0, 456, 457, 5, 66, 0, 0, 457, 458, 5, 38, 0, 0, 458, 459, 3, 104, 52,
		0, 459, 460, 5, 44, 0, 0, 460, 461, 3, 104, 52, 0, 461, 462, 5, 39, 0,
		0, 462, 103, 1, 0, 0, 0, 463, 471, 3, 42, 21, 0, 464, 471, 5, 67, 0, 0,
		465, 471, 5, 68, 0, 0, 466, 467, 5, 38, 0, 0, 467, 471, 5, 39, 0, 0, 468,
		471, 5, 69, 0, 0, 469, 471, 3, 114, 57, 0, 470, 463, 1, 0, 0, 0, 470, 464,
		1, 0, 0, 0, 470, 465, 1, 0, 0, 0, 470, 466, 1, 0, 0, 0, 470, 468, 1, 0,
		0, 0, 470, 469, 1, 0, 0, 0, 471, 105, 1, 0, 0, 0, 472, 473, 5, 18, 0, 0,
		473, 474, 5, 38, 0, 0, 474, 475, 3, 108, 54, 0, 475, 476, 5, 44, 0, 0,
		476, 477, 3, 108, 54, 0, 477, 478, 5, 39, 0, 0, 478, 107, 1, 0, 0, 0, 479,
		483, 3, 42, 21, 0, 480, 483, 3, 110, 55, 0, 481, 483, 3, 114, 57, 0, 482,
		479, 1, 0, 0, 0, 482, 480, 1, 0, 0, 0, 482, 481, 1, 0, 0, 0, 483, 109,
		1, 0, 0, 0, 484, 485, 5, 38, 0, 0, 485, 498, 5, 39, 0, 0, 486, 487, 5,
		38, 0, 0, 487, 492, 3, 112, 56, 0, 488, 489, 5, 44, 0, 0, 489, 491, 3,
		112, 56, 0, 490, 488, 1, 0, 0, 0, 491, 494, 1, 0, 0, 0, 492, 490, 1, 0,
		0, 0, 492, 493, 1, 0, 0, 0, 493, 495, 1, 0, 0, 0, 494, 492, 1, 0, 0, 0,
		495, 496, 5, 39, 0, 0, 496, 498, 1, 0, 0, 0, 497, 484, 1, 0, 0, 0, 497,
		486, 1, 0, 0, 0, 498, 111, 1, 0, 0, 0, 499, 507, 3, 32, 16, 0, 500, 507,
		3, 38, 19, 0, 501, 507, 3, 40, 20, 0, 502, 507, 3, 98, 49, 0, 503, 507,
		3, 110, 55, 0, 504, 507, 3, 42, 21, 0, 505, 507, 3, 114, 57, 0, 506, 499,
		1, 0, 0, 0, 506, 500, 1, 0, 0, 0, 506, 501, 1, 0, 0, 0, 506, 502, 1, 0,
		0, 0, 506, 503, 1, 0, 0, 0, 506, 504, 1, 0, 0, 0, 506, 505, 1, 0, 0, 0,
		507, 113, 1, 0, 0, 0, 508, 509, 5, 79, 0, 0, 509, 510, 3, 116, 58, 0, 510,
		115, 1, 0, 0, 0, 511, 513, 5, 38, 0, 0, 512, 514, 3, 118, 59, 0, 513, 512,
		1, 0, 0, 0, 513, 514, 1, 0, 0, 0, 514, 515, 1, 0, 0, 0, 515, 516, 5, 39,
		0, 0, 516, 117, 1, 0, 0, 0, 517, 522, 3, 120, 60, 0, 518, 519, 5, 44, 0,
		0, 519, 521, 3, 120, 60, 0, 520, 518, 1, 0, 0, 0, 521, 524, 1, 0, 0, 0,
		522, 520, 1, 0, 0, 0, 522, 523, 1, 0, 0, 0, 523, 119, 1, 0, 0, 0, 524,
		522, 1, 0, 0, 0, 525, 534, 3, 32, 16, 0, 526, 534, 3, 38, 19, 0, 527, 534,
		3, 40, 20, 0, 528, 534, 3, 50, 25, 0, 529, 534, 3, 98, 49, 0, 530, 534,
		3, 110, 55, 0, 531, 534, 3, 42, 21, 0, 532, 534, 3, 114, 57, 0, 533, 525,
		1, 0, 0, 0, 533, 526, 1, 0, 0, 0, 533, 527, 1, 0, 0, 0, 533, 528, 1, 0,
		0, 0, 533, 529, 1, 0, 0, 0, 533, 530, 1, 0, 0, 0, 533, 531, 1, 0, 0, 0,
		533, 532, 1, 0, 0, 0, 534, 121, 1, 0, 0, 0, 41, 130, 138, 142, 153, 159,
		166, 174, 181, 190, 199, 207, 218, 226, 231, 244, 257, 264, 282, 287, 295,
		311, 325, 337, 347, 356, 368, 380, 394, 401, 408, 438, 442, 454, 470, 482,
		492, 497, 506, 513, 522, 533,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// CqlParserInit initializes any static state used to implement CqlParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewCqlParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func CqlParserInit() {
	staticData := &CqlParserParserStaticData
	staticData.once.Do(cqlparserParserInit)
}

// NewCqlParser produces a new parser instance for the optional input antlr.TokenStream.
func NewCqlParser(input antlr.TokenStream) *CqlParser {
	CqlParserInit()
	this := new(CqlParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &CqlParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "CqlParser.g4"

	return this
}

// CqlParser tokens.
const (
	CqlParserEOF                       = antlr.TokenEOF
	CqlParserComparisonOperator        = 1
	CqlParserLT                        = 2
	CqlParserEQ                        = 3
	CqlParserGT                        = 4
	CqlParserNEQ                       = 5
	CqlParserGTEQ                      = 6
	CqlParserLTEQ                      = 7
	CqlParserBooleanLiteral            = 8
	CqlParserAND                       = 9
	CqlParserOR                        = 10
	CqlParserNOT                       = 11
	CqlParserLIKE                      = 12
	CqlParserBETWEEN                   = 13
	CqlParserIS                        = 14
	CqlParserNULL                      = 15
	CqlParserSpatialFunction           = 16
	CqlParserTemporalFunction          = 17
	CqlParserArrayFunction             = 18
	CqlParserIN                        = 19
	CqlParserPOINT                     = 20
	CqlParserLINESTRING                = 21
	CqlParserPOLYGON                   = 22
	CqlParserMULTIPOINT                = 23
	CqlParserMULTILINESTRING           = 24
	CqlParserMULTIPOLYGON              = 25
	CqlParserGEOMETRYCOLLECTION        = 26
	CqlParserBBOX                      = 27
	CqlParserCASEI                     = 28
	CqlParserACCENTI                   = 29
	CqlParserLOWER                     = 30
	CqlParserUPPER                     = 31
	CqlParserNumericLiteral            = 32
	CqlParserDIGIT                     = 33
	CqlParserDOLLAR                    = 34
	CqlParserUNDERSCORE                = 35
	CqlParserDOUBLEQUOTE               = 36
	CqlParserQUOTE                     = 37
	CqlParserLEFTPAREN                 = 38
	CqlParserRIGHTPAREN                = 39
	CqlParserLEFTSQUAREBRACKET         = 40
	CqlParserRIGHTSQUAREBRACKET        = 41
	CqlParserASTERISK                  = 42
	CqlParserPLUS                      = 43
	CqlParserCOMMA                     = 44
	CqlParserCARET                     = 45
	CqlParserMINUS                     = 46
	CqlParserPERIOD                    = 47
	CqlParserSOLIDUS                   = 48
	CqlParserCOLON                     = 49
	CqlParserPERCENT                   = 50
	CqlParserDIV                       = 51
	CqlParserALPHA                     = 52
	CqlParserIdentifierStart           = 53
	CqlParserIdentifierPart            = 54
	CqlParserUnsignedNumericLiteral    = 55
	CqlParserSignedNumericLiteral      = 56
	CqlParserExactNumericLiteral       = 57
	CqlParserApproximateNumericLiteral = 58
	CqlParserMantissa                  = 59
	CqlParserExponent                  = 60
	CqlParserSignedInteger             = 61
	CqlParserUnsignedInteger           = 62
	CqlParserSign                      = 63
	CqlParserDATE                      = 64
	CqlParserTIMESTAMP                 = 65
	CqlParserINTERVAL                  = 66
	CqlParserDateString                = 67
	CqlParserTimestampString           = 68
	CqlParserDotDotString              = 69
	CqlParserInstant                   = 70
	CqlParserFullDate                  = 71
	CqlParserDateYear                  = 72
	CqlParserDateMonth                 = 73
	CqlParserDateDay                   = 74
	CqlParserUtcTime                   = 75
	CqlParserTimeHour                  = 76
	CqlParserTimeMinute                = 77
	CqlParserTimeSecond                = 78
	CqlParserIdentifier                = 79
	CqlParserIdentifierBare            = 80
	CqlParserWS                        = 81
	CqlParserCharacterStringLiteral    = 82
	CqlParserQuotedQuote               = 83
)

// CqlParser rules.
const (
	CqlParserRULE_cqlFilter                 = 0
	CqlParserRULE_booleanExpression         = 1
	CqlParserRULE_booleanTerm               = 2
	CqlParserRULE_booleanFactor             = 3
	CqlParserRULE_booleanPrimary            = 4
	CqlParserRULE_predicate                 = 5
	CqlParserRULE_comparisonPredicate       = 6
	CqlParserRULE_binaryComparisonPredicate = 7
	CqlParserRULE_isLikePredicate           = 8
	CqlParserRULE_isBetweenPredicate        = 9
	CqlParserRULE_isInListPredicate         = 10
	CqlParserRULE_isNullPredicate           = 11
	CqlParserRULE_isNullOperand             = 12
	CqlParserRULE_scalarExpression          = 13
	CqlParserRULE_characterExpression       = 14
	CqlParserRULE_patternExpression         = 15
	CqlParserRULE_characterClause           = 16
	CqlParserRULE_characterLiteral          = 17
	CqlParserRULE_numericExpression         = 18
	CqlParserRULE_numericLiteral            = 19
	CqlParserRULE_booleanLiteral            = 20
	CqlParserRULE_propertyName              = 21
	CqlParserRULE_spatialPredicate          = 22
	CqlParserRULE_geomExpression            = 23
	CqlParserRULE_spatialInstance           = 24
	CqlParserRULE_geometryLiteral           = 25
	CqlParserRULE_point                     = 26
	CqlParserRULE_linestring                = 27
	CqlParserRULE_linestringDef             = 28
	CqlParserRULE_polygon                   = 29
	CqlParserRULE_polygonDef                = 30
	CqlParserRULE_multiPoint                = 31
	CqlParserRULE_multiPointDef             = 32
	CqlParserRULE_multiLinestring           = 33
	CqlParserRULE_multiPolygon              = 34
	CqlParserRULE_geometryCollection        = 35
	CqlParserRULE_bbox                      = 36
	CqlParserRULE_coordinate                = 37
	CqlParserRULE_xCoord                    = 38
	CqlParserRULE_yCoord                    = 39
	CqlParserRULE_zCoord                    = 40
	CqlParserRULE_westBoundLon              = 41
	CqlParserRULE_eastBoundLon              = 42
	CqlParserRULE_northBoundLat             = 43
	CqlParserRULE_southBoundLat             = 44
	CqlParserRULE_minElev                   = 45
	CqlParserRULE_maxElev                   = 46
	CqlParserRULE_temporalPredicate         = 47
	CqlParserRULE_temporalExpression        = 48
	CqlParserRULE_temporalClause            = 49
	CqlParserRULE_instantInstance           = 50
	CqlParserRULE_interval                  = 51
	CqlParserRULE_intervalParameter         = 52
	CqlParserRULE_arrayPredicate            = 53
	CqlParserRULE_arrayExpression           = 54
	CqlParserRULE_arrayClause               = 55
	CqlParserRULE_arrayElement              = 56
	CqlParserRULE_function                  = 57
	CqlParserRULE_argumentList              = 58
	CqlParserRULE_positionalArgument        = 59
	CqlParserRULE_argument                  = 60
)

// ICqlFilterContext is an interface to support dynamic dispatch.
type ICqlFilterContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BooleanExpression() IBooleanExpressionContext
	EOF() antlr.TerminalNode

	// IsCqlFilterContext differentiates from other interfaces.
	IsCqlFilterContext()
}

type CqlFilterContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCqlFilterContext() *CqlFilterContext {
	var p = new(CqlFilterContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_cqlFilter
	return p
}

func InitEmptyCqlFilterContext(p *CqlFilterContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_cqlFilter
}

func (*CqlFilterContext) IsCqlFilterContext() {}

func NewCqlFilterContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CqlFilterContext {
	var p = new(CqlFilterContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_cqlFilter

	return p
}

func (s *CqlFilterContext) GetParser() antlr.Parser { return s.parser }

func (s *CqlFilterContext) BooleanExpression() IBooleanExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanExpressionContext)
}

func (s *CqlFilterContext) EOF() antlr.TerminalNode {
	return s.GetToken(CqlParserEOF, 0)
}

func (s *CqlFilterContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CqlFilterContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CqlFilterContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterCqlFilter(s)
	}
}

func (s *CqlFilterContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitCqlFilter(s)
	}
}

func (p *CqlParser) CqlFilter() (localctx ICqlFilterContext) {
	localctx = NewCqlFilterContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, CqlParserRULE_cqlFilter)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(122)
		p.BooleanExpression()
	}
	{
		p.SetState(123)
		p.Match(CqlParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBooleanExpressionContext is an interface to support dynamic dispatch.
type IBooleanExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllBooleanTerm() []IBooleanTermContext
	BooleanTerm(i int) IBooleanTermContext
	AllOR() []antlr.TerminalNode
	OR(i int) antlr.TerminalNode

	// IsBooleanExpressionContext differentiates from other interfaces.
	IsBooleanExpressionContext()
}

type BooleanExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBooleanExpressionContext() *BooleanExpressionContext {
	var p = new(BooleanExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanExpression
	return p
}

func InitEmptyBooleanExpressionContext(p *BooleanExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanExpression
}

func (*BooleanExpressionContext) IsBooleanExpressionContext() {}

func NewBooleanExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BooleanExpressionContext {
	var p = new(BooleanExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_booleanExpression

	return p
}

func (s *BooleanExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *BooleanExpressionContext) AllBooleanTerm() []IBooleanTermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IBooleanTermContext); ok {
			len++
		}
	}

	tst := make([]IBooleanTermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IBooleanTermContext); ok {
			tst[i] = t.(IBooleanTermContext)
			i++
		}
	}

	return tst
}

func (s *BooleanExpressionContext) BooleanTerm(i int) IBooleanTermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanTermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanTermContext)
}

func (s *BooleanExpressionContext) AllOR() []antlr.TerminalNode {
	return s.GetTokens(CqlParserOR)
}

func (s *BooleanExpressionContext) OR(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserOR, i)
}

func (s *BooleanExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BooleanExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BooleanExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterBooleanExpression(s)
	}
}

func (s *BooleanExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitBooleanExpression(s)
	}
}

func (p *CqlParser) BooleanExpression() (localctx IBooleanExpressionContext) {
	localctx = NewBooleanExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, CqlParserRULE_booleanExpression)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(125)
		p.BooleanTerm()
	}
	p.SetState(130)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserOR {
		{
			p.SetState(126)
			p.Match(CqlParserOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(127)
			p.BooleanTerm()
		}

		p.SetState(132)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBooleanTermContext is an interface to support dynamic dispatch.
type IBooleanTermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllBooleanFactor() []IBooleanFactorContext
	BooleanFactor(i int) IBooleanFactorContext
	AllAND() []antlr.TerminalNode
	AND(i int) antlr.TerminalNode

	// IsBooleanTermContext differentiates from other interfaces.
	IsBooleanTermContext()
}

type BooleanTermContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBooleanTermContext() *BooleanTermContext {
	var p = new(BooleanTermContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanTerm
	return p
}

func InitEmptyBooleanTermContext(p *BooleanTermContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanTerm
}

func (*BooleanTermContext) IsBooleanTermContext() {}

func NewBooleanTermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BooleanTermContext {
	var p = new(BooleanTermContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_booleanTerm

	return p
}

func (s *BooleanTermContext) GetParser() antlr.Parser { return s.parser }

func (s *BooleanTermContext) AllBooleanFactor() []IBooleanFactorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IBooleanFactorContext); ok {
			len++
		}
	}

	tst := make([]IBooleanFactorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IBooleanFactorContext); ok {
			tst[i] = t.(IBooleanFactorContext)
			i++
		}
	}

	return tst
}

func (s *BooleanTermContext) BooleanFactor(i int) IBooleanFactorContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanFactorContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanFactorContext)
}

func (s *BooleanTermContext) AllAND() []antlr.TerminalNode {
	return s.GetTokens(CqlParserAND)
}

func (s *BooleanTermContext) AND(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserAND, i)
}

func (s *BooleanTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BooleanTermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BooleanTermContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterBooleanTerm(s)
	}
}

func (s *BooleanTermContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitBooleanTerm(s)
	}
}

func (p *CqlParser) BooleanTerm() (localctx IBooleanTermContext) {
	localctx = NewBooleanTermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, CqlParserRULE_booleanTerm)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(133)
		p.BooleanFactor()
	}
	p.SetState(138)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserAND {
		{
			p.SetState(134)
			p.Match(CqlParserAND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(135)
			p.BooleanFactor()
		}

		p.SetState(140)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBooleanFactorContext is an interface to support dynamic dispatch.
type IBooleanFactorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BooleanPrimary() IBooleanPrimaryContext
	NOT() antlr.TerminalNode

	// IsBooleanFactorContext differentiates from other interfaces.
	IsBooleanFactorContext()
}

type BooleanFactorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBooleanFactorContext() *BooleanFactorContext {
	var p = new(BooleanFactorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanFactor
	return p
}

func InitEmptyBooleanFactorContext(p *BooleanFactorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanFactor
}

func (*BooleanFactorContext) IsBooleanFactorContext() {}

func NewBooleanFactorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BooleanFactorContext {
	var p = new(BooleanFactorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_booleanFactor

	return p
}

func (s *BooleanFactorContext) GetParser() antlr.Parser { return s.parser }

func (s *BooleanFactorContext) BooleanPrimary() IBooleanPrimaryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanPrimaryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanPrimaryContext)
}

func (s *BooleanFactorContext) NOT() antlr.TerminalNode {
	return s.GetToken(CqlParserNOT, 0)
}

func (s *BooleanFactorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BooleanFactorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BooleanFactorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterBooleanFactor(s)
	}
}

func (s *BooleanFactorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitBooleanFactor(s)
	}
}

func (p *CqlParser) BooleanFactor() (localctx IBooleanFactorContext) {
	localctx = NewBooleanFactorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, CqlParserRULE_booleanFactor)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(142)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CqlParserNOT {
		{
			p.SetState(141)
			p.Match(CqlParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(144)
		p.BooleanPrimary()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBooleanPrimaryContext is an interface to support dynamic dispatch.
type IBooleanPrimaryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Predicate() IPredicateContext
	BooleanLiteral() IBooleanLiteralContext
	LEFTPAREN() antlr.TerminalNode
	BooleanExpression() IBooleanExpressionContext
	RIGHTPAREN() antlr.TerminalNode
	Function() IFunctionContext

	// IsBooleanPrimaryContext differentiates from other interfaces.
	IsBooleanPrimaryContext()
}

type BooleanPrimaryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBooleanPrimaryContext() *BooleanPrimaryContext {
	var p = new(BooleanPrimaryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanPrimary
	return p
}

func InitEmptyBooleanPrimaryContext(p *BooleanPrimaryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanPrimary
}

func (*BooleanPrimaryContext) IsBooleanPrimaryContext() {}

func NewBooleanPrimaryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BooleanPrimaryContext {
	var p = new(BooleanPrimaryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_booleanPrimary

	return p
}

func (s *BooleanPrimaryContext) GetParser() antlr.Parser { return s.parser }

func (s *BooleanPrimaryContext) Predicate() IPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPredicateContext)
}

func (s *BooleanPrimaryContext) BooleanLiteral() IBooleanLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanLiteralContext)
}

func (s *BooleanPrimaryContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *BooleanPrimaryContext) BooleanExpression() IBooleanExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanExpressionContext)
}

func (s *BooleanPrimaryContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *BooleanPrimaryContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *BooleanPrimaryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BooleanPrimaryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BooleanPrimaryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterBooleanPrimary(s)
	}
}

func (s *BooleanPrimaryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitBooleanPrimary(s)
	}
}

func (p *CqlParser) BooleanPrimary() (localctx IBooleanPrimaryContext) {
	localctx = NewBooleanPrimaryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, CqlParserRULE_booleanPrimary)
	p.SetState(153)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(146)
			p.Predicate()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(147)
			p.BooleanLiteral()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(148)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(149)
			p.BooleanExpression()
		}
		{
			p.SetState(150)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(152)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPredicateContext is an interface to support dynamic dispatch.
type IPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ComparisonPredicate() IComparisonPredicateContext
	SpatialPredicate() ISpatialPredicateContext
	TemporalPredicate() ITemporalPredicateContext
	ArrayPredicate() IArrayPredicateContext

	// IsPredicateContext differentiates from other interfaces.
	IsPredicateContext()
}

type PredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPredicateContext() *PredicateContext {
	var p = new(PredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_predicate
	return p
}

func InitEmptyPredicateContext(p *PredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_predicate
}

func (*PredicateContext) IsPredicateContext() {}

func NewPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PredicateContext {
	var p = new(PredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_predicate

	return p
}

func (s *PredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *PredicateContext) ComparisonPredicate() IComparisonPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IComparisonPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IComparisonPredicateContext)
}

func (s *PredicateContext) SpatialPredicate() ISpatialPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISpatialPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISpatialPredicateContext)
}

func (s *PredicateContext) TemporalPredicate() ITemporalPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITemporalPredicateContext)
}

func (s *PredicateContext) ArrayPredicate() IArrayPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayPredicateContext)
}

func (s *PredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterPredicate(s)
	}
}

func (s *PredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitPredicate(s)
	}
}

func (p *CqlParser) Predicate() (localctx IPredicateContext) {
	localctx = NewPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, CqlParserRULE_predicate)
	p.SetState(159)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CqlParserBooleanLiteral, CqlParserPOINT, CqlParserLINESTRING, CqlParserPOLYGON, CqlParserMULTIPOINT, CqlParserMULTILINESTRING, CqlParserMULTIPOLYGON, CqlParserCASEI, CqlParserACCENTI, CqlParserNumericLiteral, CqlParserLEFTPAREN, CqlParserDATE, CqlParserTIMESTAMP, CqlParserIdentifier, CqlParserCharacterStringLiteral:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(155)
			p.ComparisonPredicate()
		}

	case CqlParserSpatialFunction:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(156)
			p.SpatialPredicate()
		}

	case CqlParserTemporalFunction:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(157)
			p.TemporalPredicate()
		}

	case CqlParserArrayFunction:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(158)
			p.ArrayPredicate()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IComparisonPredicateContext is an interface to support dynamic dispatch.
type IComparisonPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BinaryComparisonPredicate() IBinaryComparisonPredicateContext
	IsLikePredicate() IIsLikePredicateContext
	IsBetweenPredicate() IIsBetweenPredicateContext
	IsInListPredicate() IIsInListPredicateContext
	IsNullPredicate() IIsNullPredicateContext

	// IsComparisonPredicateContext differentiates from other interfaces.
	IsComparisonPredicateContext()
}

type ComparisonPredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyComparisonPredicateContext() *ComparisonPredicateContext {
	var p = new(ComparisonPredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_comparisonPredicate
	return p
}

func InitEmptyComparisonPredicateContext(p *ComparisonPredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_comparisonPredicate
}

func (*ComparisonPredicateContext) IsComparisonPredicateContext() {}

func NewComparisonPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ComparisonPredicateContext {
	var p = new(ComparisonPredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_comparisonPredicate

	return p
}

func (s *ComparisonPredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *ComparisonPredicateContext) BinaryComparisonPredicate() IBinaryComparisonPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBinaryComparisonPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBinaryComparisonPredicateContext)
}

func (s *ComparisonPredicateContext) IsLikePredicate() IIsLikePredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIsLikePredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIsLikePredicateContext)
}

func (s *ComparisonPredicateContext) IsBetweenPredicate() IIsBetweenPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIsBetweenPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIsBetweenPredicateContext)
}

func (s *ComparisonPredicateContext) IsInListPredicate() IIsInListPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIsInListPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIsInListPredicateContext)
}

func (s *ComparisonPredicateContext) IsNullPredicate() IIsNullPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIsNullPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIsNullPredicateContext)
}

func (s *ComparisonPredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ComparisonPredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ComparisonPredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterComparisonPredicate(s)
	}
}

func (s *ComparisonPredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitComparisonPredicate(s)
	}
}

func (p *CqlParser) ComparisonPredicate() (localctx IComparisonPredicateContext) {
	localctx = NewComparisonPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, CqlParserRULE_comparisonPredicate)
	p.SetState(166)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(161)
			p.BinaryComparisonPredicate()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(162)
			p.IsLikePredicate()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(163)
			p.IsBetweenPredicate()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(164)
			p.IsInListPredicate()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(165)
			p.IsNullPredicate()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBinaryComparisonPredicateContext is an interface to support dynamic dispatch.
type IBinaryComparisonPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllScalarExpression() []IScalarExpressionContext
	ScalarExpression(i int) IScalarExpressionContext
	ComparisonOperator() antlr.TerminalNode

	// IsBinaryComparisonPredicateContext differentiates from other interfaces.
	IsBinaryComparisonPredicateContext()
}

type BinaryComparisonPredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBinaryComparisonPredicateContext() *BinaryComparisonPredicateContext {
	var p = new(BinaryComparisonPredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_binaryComparisonPredicate
	return p
}

func InitEmptyBinaryComparisonPredicateContext(p *BinaryComparisonPredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_binaryComparisonPredicate
}

func (*BinaryComparisonPredicateContext) IsBinaryComparisonPredicateContext() {}

func NewBinaryComparisonPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BinaryComparisonPredicateContext {
	var p = new(BinaryComparisonPredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_binaryComparisonPredicate

	return p
}

func (s *BinaryComparisonPredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *BinaryComparisonPredicateContext) AllScalarExpression() []IScalarExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IScalarExpressionContext); ok {
			len++
		}
	}

	tst := make([]IScalarExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IScalarExpressionContext); ok {
			tst[i] = t.(IScalarExpressionContext)
			i++
		}
	}

	return tst
}

func (s *BinaryComparisonPredicateContext) ScalarExpression(i int) IScalarExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IScalarExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IScalarExpressionContext)
}

func (s *BinaryComparisonPredicateContext) ComparisonOperator() antlr.TerminalNode {
	return s.GetToken(CqlParserComparisonOperator, 0)
}

func (s *BinaryComparisonPredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BinaryComparisonPredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BinaryComparisonPredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterBinaryComparisonPredicate(s)
	}
}

func (s *BinaryComparisonPredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitBinaryComparisonPredicate(s)
	}
}

func (p *CqlParser) BinaryComparisonPredicate() (localctx IBinaryComparisonPredicateContext) {
	localctx = NewBinaryComparisonPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, CqlParserRULE_binaryComparisonPredicate)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(168)
		p.ScalarExpression()
	}
	{
		p.SetState(169)
		p.Match(CqlParserComparisonOperator)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(170)
		p.ScalarExpression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIsLikePredicateContext is an interface to support dynamic dispatch.
type IIsLikePredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CharacterExpression() ICharacterExpressionContext
	LIKE() antlr.TerminalNode
	PatternExpression() IPatternExpressionContext
	NOT() antlr.TerminalNode

	// IsIsLikePredicateContext differentiates from other interfaces.
	IsIsLikePredicateContext()
}

type IsLikePredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIsLikePredicateContext() *IsLikePredicateContext {
	var p = new(IsLikePredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isLikePredicate
	return p
}

func InitEmptyIsLikePredicateContext(p *IsLikePredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isLikePredicate
}

func (*IsLikePredicateContext) IsIsLikePredicateContext() {}

func NewIsLikePredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IsLikePredicateContext {
	var p = new(IsLikePredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_isLikePredicate

	return p
}

func (s *IsLikePredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *IsLikePredicateContext) CharacterExpression() ICharacterExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICharacterExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICharacterExpressionContext)
}

func (s *IsLikePredicateContext) LIKE() antlr.TerminalNode {
	return s.GetToken(CqlParserLIKE, 0)
}

func (s *IsLikePredicateContext) PatternExpression() IPatternExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPatternExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPatternExpressionContext)
}

func (s *IsLikePredicateContext) NOT() antlr.TerminalNode {
	return s.GetToken(CqlParserNOT, 0)
}

func (s *IsLikePredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IsLikePredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IsLikePredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterIsLikePredicate(s)
	}
}

func (s *IsLikePredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitIsLikePredicate(s)
	}
}

func (p *CqlParser) IsLikePredicate() (localctx IIsLikePredicateContext) {
	localctx = NewIsLikePredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, CqlParserRULE_isLikePredicate)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(172)
		p.CharacterExpression()
	}
	p.SetState(174)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CqlParserNOT {
		{
			p.SetState(173)
			p.Match(CqlParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(176)
		p.Match(CqlParserLIKE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(177)
		p.PatternExpression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIsBetweenPredicateContext is an interface to support dynamic dispatch.
type IIsBetweenPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllNumericExpression() []INumericExpressionContext
	NumericExpression(i int) INumericExpressionContext
	BETWEEN() antlr.TerminalNode
	AND() antlr.TerminalNode
	NOT() antlr.TerminalNode

	// IsIsBetweenPredicateContext differentiates from other interfaces.
	IsIsBetweenPredicateContext()
}

type IsBetweenPredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIsBetweenPredicateContext() *IsBetweenPredicateContext {
	var p = new(IsBetweenPredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isBetweenPredicate
	return p
}

func InitEmptyIsBetweenPredicateContext(p *IsBetweenPredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isBetweenPredicate
}

func (*IsBetweenPredicateContext) IsIsBetweenPredicateContext() {}

func NewIsBetweenPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IsBetweenPredicateContext {
	var p = new(IsBetweenPredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_isBetweenPredicate

	return p
}

func (s *IsBetweenPredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *IsBetweenPredicateContext) AllNumericExpression() []INumericExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(INumericExpressionContext); ok {
			len++
		}
	}

	tst := make([]INumericExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(INumericExpressionContext); ok {
			tst[i] = t.(INumericExpressionContext)
			i++
		}
	}

	return tst
}

func (s *IsBetweenPredicateContext) NumericExpression(i int) INumericExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INumericExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(INumericExpressionContext)
}

func (s *IsBetweenPredicateContext) BETWEEN() antlr.TerminalNode {
	return s.GetToken(CqlParserBETWEEN, 0)
}

func (s *IsBetweenPredicateContext) AND() antlr.TerminalNode {
	return s.GetToken(CqlParserAND, 0)
}

func (s *IsBetweenPredicateContext) NOT() antlr.TerminalNode {
	return s.GetToken(CqlParserNOT, 0)
}

func (s *IsBetweenPredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IsBetweenPredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IsBetweenPredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterIsBetweenPredicate(s)
	}
}

func (s *IsBetweenPredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitIsBetweenPredicate(s)
	}
}

func (p *CqlParser) IsBetweenPredicate() (localctx IIsBetweenPredicateContext) {
	localctx = NewIsBetweenPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, CqlParserRULE_isBetweenPredicate)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(179)
		p.NumericExpression()
	}
	p.SetState(181)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CqlParserNOT {
		{
			p.SetState(180)
			p.Match(CqlParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(183)
		p.Match(CqlParserBETWEEN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(184)
		p.NumericExpression()
	}
	{
		p.SetState(185)
		p.Match(CqlParserAND)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(186)
		p.NumericExpression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIsInListPredicateContext is an interface to support dynamic dispatch.
type IIsInListPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllScalarExpression() []IScalarExpressionContext
	ScalarExpression(i int) IScalarExpressionContext
	IN() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	RIGHTPAREN() antlr.TerminalNode
	NOT() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsIsInListPredicateContext differentiates from other interfaces.
	IsIsInListPredicateContext()
}

type IsInListPredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIsInListPredicateContext() *IsInListPredicateContext {
	var p = new(IsInListPredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isInListPredicate
	return p
}

func InitEmptyIsInListPredicateContext(p *IsInListPredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isInListPredicate
}

func (*IsInListPredicateContext) IsIsInListPredicateContext() {}

func NewIsInListPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IsInListPredicateContext {
	var p = new(IsInListPredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_isInListPredicate

	return p
}

func (s *IsInListPredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *IsInListPredicateContext) AllScalarExpression() []IScalarExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IScalarExpressionContext); ok {
			len++
		}
	}

	tst := make([]IScalarExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IScalarExpressionContext); ok {
			tst[i] = t.(IScalarExpressionContext)
			i++
		}
	}

	return tst
}

func (s *IsInListPredicateContext) ScalarExpression(i int) IScalarExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IScalarExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IScalarExpressionContext)
}

func (s *IsInListPredicateContext) IN() antlr.TerminalNode {
	return s.GetToken(CqlParserIN, 0)
}

func (s *IsInListPredicateContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *IsInListPredicateContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *IsInListPredicateContext) NOT() antlr.TerminalNode {
	return s.GetToken(CqlParserNOT, 0)
}

func (s *IsInListPredicateContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *IsInListPredicateContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *IsInListPredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IsInListPredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IsInListPredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterIsInListPredicate(s)
	}
}

func (s *IsInListPredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitIsInListPredicate(s)
	}
}

func (p *CqlParser) IsInListPredicate() (localctx IIsInListPredicateContext) {
	localctx = NewIsInListPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, CqlParserRULE_isInListPredicate)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(188)
		p.ScalarExpression()
	}
	p.SetState(190)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CqlParserNOT {
		{
			p.SetState(189)
			p.Match(CqlParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(192)
		p.Match(CqlParserIN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(193)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(194)
		p.ScalarExpression()
	}
	p.SetState(199)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserCOMMA {
		{
			p.SetState(195)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(196)
			p.ScalarExpression()
		}

		p.SetState(201)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(202)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIsNullPredicateContext is an interface to support dynamic dispatch.
type IIsNullPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IsNullOperand() IIsNullOperandContext
	IS() antlr.TerminalNode
	NULL() antlr.TerminalNode
	NOT() antlr.TerminalNode

	// IsIsNullPredicateContext differentiates from other interfaces.
	IsIsNullPredicateContext()
}

type IsNullPredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIsNullPredicateContext() *IsNullPredicateContext {
	var p = new(IsNullPredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isNullPredicate
	return p
}

func InitEmptyIsNullPredicateContext(p *IsNullPredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isNullPredicate
}

func (*IsNullPredicateContext) IsIsNullPredicateContext() {}

func NewIsNullPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IsNullPredicateContext {
	var p = new(IsNullPredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_isNullPredicate

	return p
}

func (s *IsNullPredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *IsNullPredicateContext) IsNullOperand() IIsNullOperandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIsNullOperandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIsNullOperandContext)
}

func (s *IsNullPredicateContext) IS() antlr.TerminalNode {
	return s.GetToken(CqlParserIS, 0)
}

func (s *IsNullPredicateContext) NULL() antlr.TerminalNode {
	return s.GetToken(CqlParserNULL, 0)
}

func (s *IsNullPredicateContext) NOT() antlr.TerminalNode {
	return s.GetToken(CqlParserNOT, 0)
}

func (s *IsNullPredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IsNullPredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IsNullPredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterIsNullPredicate(s)
	}
}

func (s *IsNullPredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitIsNullPredicate(s)
	}
}

func (p *CqlParser) IsNullPredicate() (localctx IIsNullPredicateContext) {
	localctx = NewIsNullPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, CqlParserRULE_isNullPredicate)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(204)
		p.IsNullOperand()
	}
	{
		p.SetState(205)
		p.Match(CqlParserIS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(207)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CqlParserNOT {
		{
			p.SetState(206)
			p.Match(CqlParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(209)
		p.Match(CqlParserNULL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIsNullOperandContext is an interface to support dynamic dispatch.
type IIsNullOperandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CharacterClause() ICharacterClauseContext
	NumericLiteral() INumericLiteralContext
	BooleanLiteral() IBooleanLiteralContext
	InstantInstance() IInstantInstanceContext
	GeometryLiteral() IGeometryLiteralContext
	PropertyName() IPropertyNameContext
	Function() IFunctionContext

	// IsIsNullOperandContext differentiates from other interfaces.
	IsIsNullOperandContext()
}

type IsNullOperandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIsNullOperandContext() *IsNullOperandContext {
	var p = new(IsNullOperandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isNullOperand
	return p
}

func InitEmptyIsNullOperandContext(p *IsNullOperandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_isNullOperand
}

func (*IsNullOperandContext) IsIsNullOperandContext() {}

func NewIsNullOperandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IsNullOperandContext {
	var p = new(IsNullOperandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_isNullOperand

	return p
}

func (s *IsNullOperandContext) GetParser() antlr.Parser { return s.parser }

func (s *IsNullOperandContext) CharacterClause() ICharacterClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICharacterClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICharacterClauseContext)
}

func (s *IsNullOperandContext) NumericLiteral() INumericLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INumericLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INumericLiteralContext)
}

func (s *IsNullOperandContext) BooleanLiteral() IBooleanLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanLiteralContext)
}

func (s *IsNullOperandContext) InstantInstance() IInstantInstanceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInstantInstanceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInstantInstanceContext)
}

func (s *IsNullOperandContext) GeometryLiteral() IGeometryLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGeometryLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGeometryLiteralContext)
}

func (s *IsNullOperandContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *IsNullOperandContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *IsNullOperandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IsNullOperandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IsNullOperandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterIsNullOperand(s)
	}
}

func (s *IsNullOperandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitIsNullOperand(s)
	}
}

func (p *CqlParser) IsNullOperand() (localctx IIsNullOperandContext) {
	localctx = NewIsNullOperandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, CqlParserRULE_isNullOperand)
	p.SetState(218)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(211)
			p.CharacterClause()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(212)
			p.NumericLiteral()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(213)
			p.BooleanLiteral()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(214)
			p.InstantInstance()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(215)
			p.GeometryLiteral()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(216)
			p.PropertyName()
		}

	case 7:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(217)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IScalarExpressionContext is an interface to support dynamic dispatch.
type IScalarExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CharacterClause() ICharacterClauseContext
	NumericLiteral() INumericLiteralContext
	InstantInstance() IInstantInstanceContext
	BooleanLiteral() IBooleanLiteralContext
	PropertyName() IPropertyNameContext
	Function() IFunctionContext

	// IsScalarExpressionContext differentiates from other interfaces.
	IsScalarExpressionContext()
}

type ScalarExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyScalarExpressionContext() *ScalarExpressionContext {
	var p = new(ScalarExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_scalarExpression
	return p
}

func InitEmptyScalarExpressionContext(p *ScalarExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_scalarExpression
}

func (*ScalarExpressionContext) IsScalarExpressionContext() {}

func NewScalarExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ScalarExpressionContext {
	var p = new(ScalarExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_scalarExpression

	return p
}

func (s *ScalarExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ScalarExpressionContext) CharacterClause() ICharacterClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICharacterClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICharacterClauseContext)
}

func (s *ScalarExpressionContext) NumericLiteral() INumericLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INumericLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INumericLiteralContext)
}

func (s *ScalarExpressionContext) InstantInstance() IInstantInstanceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInstantInstanceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInstantInstanceContext)
}

func (s *ScalarExpressionContext) BooleanLiteral() IBooleanLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanLiteralContext)
}

func (s *ScalarExpressionContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *ScalarExpressionContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *ScalarExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ScalarExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ScalarExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterScalarExpression(s)
	}
}

func (s *ScalarExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitScalarExpression(s)
	}
}

func (p *CqlParser) ScalarExpression() (localctx IScalarExpressionContext) {
	localctx = NewScalarExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, CqlParserRULE_scalarExpression)
	p.SetState(226)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(220)
			p.CharacterClause()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(221)
			p.NumericLiteral()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(222)
			p.InstantInstance()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(223)
			p.BooleanLiteral()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(224)
			p.PropertyName()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(225)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICharacterExpressionContext is an interface to support dynamic dispatch.
type ICharacterExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CharacterClause() ICharacterClauseContext
	PropertyName() IPropertyNameContext
	Function() IFunctionContext

	// IsCharacterExpressionContext differentiates from other interfaces.
	IsCharacterExpressionContext()
}

type CharacterExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCharacterExpressionContext() *CharacterExpressionContext {
	var p = new(CharacterExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_characterExpression
	return p
}

func InitEmptyCharacterExpressionContext(p *CharacterExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_characterExpression
}

func (*CharacterExpressionContext) IsCharacterExpressionContext() {}

func NewCharacterExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CharacterExpressionContext {
	var p = new(CharacterExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_characterExpression

	return p
}

func (s *CharacterExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *CharacterExpressionContext) CharacterClause() ICharacterClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICharacterClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICharacterClauseContext)
}

func (s *CharacterExpressionContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *CharacterExpressionContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *CharacterExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CharacterExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CharacterExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterCharacterExpression(s)
	}
}

func (s *CharacterExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitCharacterExpression(s)
	}
}

func (p *CqlParser) CharacterExpression() (localctx ICharacterExpressionContext) {
	localctx = NewCharacterExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, CqlParserRULE_characterExpression)
	p.SetState(231)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(228)
			p.CharacterClause()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(229)
			p.PropertyName()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(230)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPatternExpressionContext is an interface to support dynamic dispatch.
type IPatternExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CharacterLiteral() ICharacterLiteralContext
	CASEI() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	PatternExpression() IPatternExpressionContext
	RIGHTPAREN() antlr.TerminalNode
	ACCENTI() antlr.TerminalNode

	// IsPatternExpressionContext differentiates from other interfaces.
	IsPatternExpressionContext()
}

type PatternExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPatternExpressionContext() *PatternExpressionContext {
	var p = new(PatternExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_patternExpression
	return p
}

func InitEmptyPatternExpressionContext(p *PatternExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_patternExpression
}

func (*PatternExpressionContext) IsPatternExpressionContext() {}

func NewPatternExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PatternExpressionContext {
	var p = new(PatternExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_patternExpression

	return p
}

func (s *PatternExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *PatternExpressionContext) CharacterLiteral() ICharacterLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICharacterLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICharacterLiteralContext)
}

func (s *PatternExpressionContext) CASEI() antlr.TerminalNode {
	return s.GetToken(CqlParserCASEI, 0)
}

func (s *PatternExpressionContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *PatternExpressionContext) PatternExpression() IPatternExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPatternExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPatternExpressionContext)
}

func (s *PatternExpressionContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *PatternExpressionContext) ACCENTI() antlr.TerminalNode {
	return s.GetToken(CqlParserACCENTI, 0)
}

func (s *PatternExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PatternExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PatternExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterPatternExpression(s)
	}
}

func (s *PatternExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitPatternExpression(s)
	}
}

func (p *CqlParser) PatternExpression() (localctx IPatternExpressionContext) {
	localctx = NewPatternExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, CqlParserRULE_patternExpression)
	p.SetState(244)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CqlParserCharacterStringLiteral:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(233)
			p.CharacterLiteral()
		}

	case CqlParserCASEI:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(234)
			p.Match(CqlParserCASEI)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(235)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(236)
			p.PatternExpression()
		}
		{
			p.SetState(237)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CqlParserACCENTI:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(239)
			p.Match(CqlParserACCENTI)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(240)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(241)
			p.PatternExpression()
		}
		{
			p.SetState(242)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICharacterClauseContext is an interface to support dynamic dispatch.
type ICharacterClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CharacterLiteral() ICharacterLiteralContext
	CASEI() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	CharacterExpression() ICharacterExpressionContext
	RIGHTPAREN() antlr.TerminalNode
	ACCENTI() antlr.TerminalNode

	// IsCharacterClauseContext differentiates from other interfaces.
	IsCharacterClauseContext()
}

type CharacterClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCharacterClauseContext() *CharacterClauseContext {
	var p = new(CharacterClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_characterClause
	return p
}

func InitEmptyCharacterClauseContext(p *CharacterClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_characterClause
}

func (*CharacterClauseContext) IsCharacterClauseContext() {}

func NewCharacterClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CharacterClauseContext {
	var p = new(CharacterClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_characterClause

	return p
}

func (s *CharacterClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *CharacterClauseContext) CharacterLiteral() ICharacterLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICharacterLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICharacterLiteralContext)
}

func (s *CharacterClauseContext) CASEI() antlr.TerminalNode {
	return s.GetToken(CqlParserCASEI, 0)
}

func (s *CharacterClauseContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *CharacterClauseContext) CharacterExpression() ICharacterExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICharacterExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICharacterExpressionContext)
}

func (s *CharacterClauseContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *CharacterClauseContext) ACCENTI() antlr.TerminalNode {
	return s.GetToken(CqlParserACCENTI, 0)
}

func (s *CharacterClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CharacterClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CharacterClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterCharacterClause(s)
	}
}

func (s *CharacterClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitCharacterClause(s)
	}
}

func (p *CqlParser) CharacterClause() (localctx ICharacterClauseContext) {
	localctx = NewCharacterClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, CqlParserRULE_characterClause)
	p.SetState(257)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CqlParserCharacterStringLiteral:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(246)
			p.CharacterLiteral()
		}

	case CqlParserCASEI:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(247)
			p.Match(CqlParserCASEI)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(248)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(249)
			p.CharacterExpression()
		}
		{
			p.SetState(250)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CqlParserACCENTI:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(252)
			p.Match(CqlParserACCENTI)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(253)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(254)
			p.CharacterExpression()
		}
		{
			p.SetState(255)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICharacterLiteralContext is an interface to support dynamic dispatch.
type ICharacterLiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CharacterStringLiteral() antlr.TerminalNode

	// IsCharacterLiteralContext differentiates from other interfaces.
	IsCharacterLiteralContext()
}

type CharacterLiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCharacterLiteralContext() *CharacterLiteralContext {
	var p = new(CharacterLiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_characterLiteral
	return p
}

func InitEmptyCharacterLiteralContext(p *CharacterLiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_characterLiteral
}

func (*CharacterLiteralContext) IsCharacterLiteralContext() {}

func NewCharacterLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CharacterLiteralContext {
	var p = new(CharacterLiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_characterLiteral

	return p
}

func (s *CharacterLiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *CharacterLiteralContext) CharacterStringLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserCharacterStringLiteral, 0)
}

func (s *CharacterLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CharacterLiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CharacterLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterCharacterLiteral(s)
	}
}

func (s *CharacterLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitCharacterLiteral(s)
	}
}

func (p *CqlParser) CharacterLiteral() (localctx ICharacterLiteralContext) {
	localctx = NewCharacterLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, CqlParserRULE_characterLiteral)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(259)
		p.Match(CqlParserCharacterStringLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INumericExpressionContext is an interface to support dynamic dispatch.
type INumericExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() INumericLiteralContext
	PropertyName() IPropertyNameContext
	Function() IFunctionContext

	// IsNumericExpressionContext differentiates from other interfaces.
	IsNumericExpressionContext()
}

type NumericExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNumericExpressionContext() *NumericExpressionContext {
	var p = new(NumericExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_numericExpression
	return p
}

func InitEmptyNumericExpressionContext(p *NumericExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_numericExpression
}

func (*NumericExpressionContext) IsNumericExpressionContext() {}

func NewNumericExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NumericExpressionContext {
	var p = new(NumericExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_numericExpression

	return p
}

func (s *NumericExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *NumericExpressionContext) NumericLiteral() INumericLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INumericLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INumericLiteralContext)
}

func (s *NumericExpressionContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *NumericExpressionContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *NumericExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NumericExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NumericExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterNumericExpression(s)
	}
}

func (s *NumericExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitNumericExpression(s)
	}
}

func (p *CqlParser) NumericExpression() (localctx INumericExpressionContext) {
	localctx = NewNumericExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, CqlParserRULE_numericExpression)
	p.SetState(264)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 16, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(261)
			p.NumericLiteral()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(262)
			p.PropertyName()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(263)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INumericLiteralContext is an interface to support dynamic dispatch.
type INumericLiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsNumericLiteralContext differentiates from other interfaces.
	IsNumericLiteralContext()
}

type NumericLiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNumericLiteralContext() *NumericLiteralContext {
	var p = new(NumericLiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_numericLiteral
	return p
}

func InitEmptyNumericLiteralContext(p *NumericLiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_numericLiteral
}

func (*NumericLiteralContext) IsNumericLiteralContext() {}

func NewNumericLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NumericLiteralContext {
	var p = new(NumericLiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_numericLiteral

	return p
}

func (s *NumericLiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *NumericLiteralContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *NumericLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NumericLiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NumericLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterNumericLiteral(s)
	}
}

func (s *NumericLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitNumericLiteral(s)
	}
}

func (p *CqlParser) NumericLiteral() (localctx INumericLiteralContext) {
	localctx = NewNumericLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, CqlParserRULE_numericLiteral)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(266)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBooleanLiteralContext is an interface to support dynamic dispatch.
type IBooleanLiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BooleanLiteral() antlr.TerminalNode

	// IsBooleanLiteralContext differentiates from other interfaces.
	IsBooleanLiteralContext()
}

type BooleanLiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBooleanLiteralContext() *BooleanLiteralContext {
	var p = new(BooleanLiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanLiteral
	return p
}

func InitEmptyBooleanLiteralContext(p *BooleanLiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_booleanLiteral
}

func (*BooleanLiteralContext) IsBooleanLiteralContext() {}

func NewBooleanLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BooleanLiteralContext {
	var p = new(BooleanLiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_booleanLiteral

	return p
}

func (s *BooleanLiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *BooleanLiteralContext) BooleanLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserBooleanLiteral, 0)
}

func (s *BooleanLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BooleanLiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BooleanLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterBooleanLiteral(s)
	}
}

func (s *BooleanLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitBooleanLiteral(s)
	}
}

func (p *CqlParser) BooleanLiteral() (localctx IBooleanLiteralContext) {
	localctx = NewBooleanLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 40, CqlParserRULE_booleanLiteral)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(268)
		p.Match(CqlParserBooleanLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPropertyNameContext is an interface to support dynamic dispatch.
type IPropertyNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() antlr.TerminalNode

	// IsPropertyNameContext differentiates from other interfaces.
	IsPropertyNameContext()
}

type PropertyNameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPropertyNameContext() *PropertyNameContext {
	var p = new(PropertyNameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_propertyName
	return p
}

func InitEmptyPropertyNameContext(p *PropertyNameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_propertyName
}

func (*PropertyNameContext) IsPropertyNameContext() {}

func NewPropertyNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PropertyNameContext {
	var p = new(PropertyNameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_propertyName

	return p
}

func (s *PropertyNameContext) GetParser() antlr.Parser { return s.parser }

func (s *PropertyNameContext) Identifier() antlr.TerminalNode {
	return s.GetToken(CqlParserIdentifier, 0)
}

func (s *PropertyNameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PropertyNameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PropertyNameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterPropertyName(s)
	}
}

func (s *PropertyNameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitPropertyName(s)
	}
}

func (p *CqlParser) PropertyName() (localctx IPropertyNameContext) {
	localctx = NewPropertyNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 42, CqlParserRULE_propertyName)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(270)
		p.Match(CqlParserIdentifier)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISpatialPredicateContext is an interface to support dynamic dispatch.
type ISpatialPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SpatialFunction() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	AllGeomExpression() []IGeomExpressionContext
	GeomExpression(i int) IGeomExpressionContext
	COMMA() antlr.TerminalNode
	RIGHTPAREN() antlr.TerminalNode

	// IsSpatialPredicateContext differentiates from other interfaces.
	IsSpatialPredicateContext()
}

type SpatialPredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySpatialPredicateContext() *SpatialPredicateContext {
	var p = new(SpatialPredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_spatialPredicate
	return p
}

func InitEmptySpatialPredicateContext(p *SpatialPredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_spatialPredicate
}

func (*SpatialPredicateContext) IsSpatialPredicateContext() {}

func NewSpatialPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SpatialPredicateContext {
	var p = new(SpatialPredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_spatialPredicate

	return p
}

func (s *SpatialPredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *SpatialPredicateContext) SpatialFunction() antlr.TerminalNode {
	return s.GetToken(CqlParserSpatialFunction, 0)
}

func (s *SpatialPredicateContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *SpatialPredicateContext) AllGeomExpression() []IGeomExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IGeomExpressionContext); ok {
			len++
		}
	}

	tst := make([]IGeomExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IGeomExpressionContext); ok {
			tst[i] = t.(IGeomExpressionContext)
			i++
		}
	}

	return tst
}

func (s *SpatialPredicateContext) GeomExpression(i int) IGeomExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGeomExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGeomExpressionContext)
}

func (s *SpatialPredicateContext) COMMA() antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, 0)
}

func (s *SpatialPredicateContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *SpatialPredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SpatialPredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SpatialPredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterSpatialPredicate(s)
	}
}

func (s *SpatialPredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitSpatialPredicate(s)
	}
}

func (p *CqlParser) SpatialPredicate() (localctx ISpatialPredicateContext) {
	localctx = NewSpatialPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 44, CqlParserRULE_spatialPredicate)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(272)
		p.Match(CqlParserSpatialFunction)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(273)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(274)
		p.GeomExpression()
	}
	{
		p.SetState(275)
		p.Match(CqlParserCOMMA)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(276)
		p.GeomExpression()
	}
	{
		p.SetState(277)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IGeomExpressionContext is an interface to support dynamic dispatch.
type IGeomExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SpatialInstance() ISpatialInstanceContext
	PropertyName() IPropertyNameContext
	Function() IFunctionContext

	// IsGeomExpressionContext differentiates from other interfaces.
	IsGeomExpressionContext()
}

type GeomExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyGeomExpressionContext() *GeomExpressionContext {
	var p = new(GeomExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_geomExpression
	return p
}

func InitEmptyGeomExpressionContext(p *GeomExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_geomExpression
}

func (*GeomExpressionContext) IsGeomExpressionContext() {}

func NewGeomExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *GeomExpressionContext {
	var p = new(GeomExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_geomExpression

	return p
}

func (s *GeomExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *GeomExpressionContext) SpatialInstance() ISpatialInstanceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISpatialInstanceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISpatialInstanceContext)
}

func (s *GeomExpressionContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *GeomExpressionContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *GeomExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *GeomExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *GeomExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterGeomExpression(s)
	}
}

func (s *GeomExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitGeomExpression(s)
	}
}

func (p *CqlParser) GeomExpression() (localctx IGeomExpressionContext) {
	localctx = NewGeomExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 46, CqlParserRULE_geomExpression)
	p.SetState(282)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 17, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(279)
			p.SpatialInstance()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(280)
			p.PropertyName()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(281)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISpatialInstanceContext is an interface to support dynamic dispatch.
type ISpatialInstanceContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	GeometryLiteral() IGeometryLiteralContext
	GeometryCollection() IGeometryCollectionContext
	Bbox() IBboxContext

	// IsSpatialInstanceContext differentiates from other interfaces.
	IsSpatialInstanceContext()
}

type SpatialInstanceContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySpatialInstanceContext() *SpatialInstanceContext {
	var p = new(SpatialInstanceContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_spatialInstance
	return p
}

func InitEmptySpatialInstanceContext(p *SpatialInstanceContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_spatialInstance
}

func (*SpatialInstanceContext) IsSpatialInstanceContext() {}

func NewSpatialInstanceContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SpatialInstanceContext {
	var p = new(SpatialInstanceContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_spatialInstance

	return p
}

func (s *SpatialInstanceContext) GetParser() antlr.Parser { return s.parser }

func (s *SpatialInstanceContext) GeometryLiteral() IGeometryLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGeometryLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGeometryLiteralContext)
}

func (s *SpatialInstanceContext) GeometryCollection() IGeometryCollectionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGeometryCollectionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGeometryCollectionContext)
}

func (s *SpatialInstanceContext) Bbox() IBboxContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBboxContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBboxContext)
}

func (s *SpatialInstanceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SpatialInstanceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SpatialInstanceContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterSpatialInstance(s)
	}
}

func (s *SpatialInstanceContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitSpatialInstance(s)
	}
}

func (p *CqlParser) SpatialInstance() (localctx ISpatialInstanceContext) {
	localctx = NewSpatialInstanceContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 48, CqlParserRULE_spatialInstance)
	p.SetState(287)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CqlParserPOINT, CqlParserLINESTRING, CqlParserPOLYGON, CqlParserMULTIPOINT, CqlParserMULTILINESTRING, CqlParserMULTIPOLYGON:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(284)
			p.GeometryLiteral()
		}

	case CqlParserGEOMETRYCOLLECTION:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(285)
			p.GeometryCollection()
		}

	case CqlParserBBOX:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(286)
			p.Bbox()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IGeometryLiteralContext is an interface to support dynamic dispatch.
type IGeometryLiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Point() IPointContext
	Linestring() ILinestringContext
	Polygon() IPolygonContext
	MultiPoint() IMultiPointContext
	MultiLinestring() IMultiLinestringContext
	MultiPolygon() IMultiPolygonContext

	// IsGeometryLiteralContext differentiates from other interfaces.
	IsGeometryLiteralContext()
}

type GeometryLiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyGeometryLiteralContext() *GeometryLiteralContext {
	var p = new(GeometryLiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_geometryLiteral
	return p
}

func InitEmptyGeometryLiteralContext(p *GeometryLiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_geometryLiteral
}

func (*GeometryLiteralContext) IsGeometryLiteralContext() {}

func NewGeometryLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *GeometryLiteralContext {
	var p = new(GeometryLiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_geometryLiteral

	return p
}

func (s *GeometryLiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *GeometryLiteralContext) Point() IPointContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPointContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPointContext)
}

func (s *GeometryLiteralContext) Linestring() ILinestringContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILinestringContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILinestringContext)
}

func (s *GeometryLiteralContext) Polygon() IPolygonContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPolygonContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPolygonContext)
}

func (s *GeometryLiteralContext) MultiPoint() IMultiPointContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMultiPointContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMultiPointContext)
}

func (s *GeometryLiteralContext) MultiLinestring() IMultiLinestringContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMultiLinestringContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMultiLinestringContext)
}

func (s *GeometryLiteralContext) MultiPolygon() IMultiPolygonContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMultiPolygonContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMultiPolygonContext)
}

func (s *GeometryLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *GeometryLiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *GeometryLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterGeometryLiteral(s)
	}
}

func (s *GeometryLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitGeometryLiteral(s)
	}
}

func (p *CqlParser) GeometryLiteral() (localctx IGeometryLiteralContext) {
	localctx = NewGeometryLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 50, CqlParserRULE_geometryLiteral)
	p.SetState(295)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CqlParserPOINT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(289)
			p.Point()
		}

	case CqlParserLINESTRING:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(290)
			p.Linestring()
		}

	case CqlParserPOLYGON:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(291)
			p.Polygon()
		}

	case CqlParserMULTIPOINT:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(292)
			p.MultiPoint()
		}

	case CqlParserMULTILINESTRING:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(293)
			p.MultiLinestring()
		}

	case CqlParserMULTIPOLYGON:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(294)
			p.MultiPolygon()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPointContext is an interface to support dynamic dispatch.
type IPointContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	POINT() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	Coordinate() ICoordinateContext
	RIGHTPAREN() antlr.TerminalNode

	// IsPointContext differentiates from other interfaces.
	IsPointContext()
}

type PointContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPointContext() *PointContext {
	var p = new(PointContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_point
	return p
}

func InitEmptyPointContext(p *PointContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_point
}

func (*PointContext) IsPointContext() {}

func NewPointContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PointContext {
	var p = new(PointContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_point

	return p
}

func (s *PointContext) GetParser() antlr.Parser { return s.parser }

func (s *PointContext) POINT() antlr.TerminalNode {
	return s.GetToken(CqlParserPOINT, 0)
}

func (s *PointContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *PointContext) Coordinate() ICoordinateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICoordinateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICoordinateContext)
}

func (s *PointContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *PointContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PointContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PointContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterPoint(s)
	}
}

func (s *PointContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitPoint(s)
	}
}

func (p *CqlParser) Point() (localctx IPointContext) {
	localctx = NewPointContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 52, CqlParserRULE_point)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(297)
		p.Match(CqlParserPOINT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(298)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(299)
		p.Coordinate()
	}
	{
		p.SetState(300)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILinestringContext is an interface to support dynamic dispatch.
type ILinestringContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LINESTRING() antlr.TerminalNode
	LinestringDef() ILinestringDefContext

	// IsLinestringContext differentiates from other interfaces.
	IsLinestringContext()
}

type LinestringContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLinestringContext() *LinestringContext {
	var p = new(LinestringContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_linestring
	return p
}

func InitEmptyLinestringContext(p *LinestringContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_linestring
}

func (*LinestringContext) IsLinestringContext() {}

func NewLinestringContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LinestringContext {
	var p = new(LinestringContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_linestring

	return p
}

func (s *LinestringContext) GetParser() antlr.Parser { return s.parser }

func (s *LinestringContext) LINESTRING() antlr.TerminalNode {
	return s.GetToken(CqlParserLINESTRING, 0)
}

func (s *LinestringContext) LinestringDef() ILinestringDefContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILinestringDefContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILinestringDefContext)
}

func (s *LinestringContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LinestringContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LinestringContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterLinestring(s)
	}
}

func (s *LinestringContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitLinestring(s)
	}
}

func (p *CqlParser) Linestring() (localctx ILinestringContext) {
	localctx = NewLinestringContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 54, CqlParserRULE_linestring)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(302)
		p.Match(CqlParserLINESTRING)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(303)
		p.LinestringDef()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILinestringDefContext is an interface to support dynamic dispatch.
type ILinestringDefContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LEFTPAREN() antlr.TerminalNode
	AllCoordinate() []ICoordinateContext
	Coordinate(i int) ICoordinateContext
	RIGHTPAREN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsLinestringDefContext differentiates from other interfaces.
	IsLinestringDefContext()
}

type LinestringDefContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLinestringDefContext() *LinestringDefContext {
	var p = new(LinestringDefContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_linestringDef
	return p
}

func InitEmptyLinestringDefContext(p *LinestringDefContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_linestringDef
}

func (*LinestringDefContext) IsLinestringDefContext() {}

func NewLinestringDefContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LinestringDefContext {
	var p = new(LinestringDefContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_linestringDef

	return p
}

func (s *LinestringDefContext) GetParser() antlr.Parser { return s.parser }

func (s *LinestringDefContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *LinestringDefContext) AllCoordinate() []ICoordinateContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ICoordinateContext); ok {
			len++
		}
	}

	tst := make([]ICoordinateContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ICoordinateContext); ok {
			tst[i] = t.(ICoordinateContext)
			i++
		}
	}

	return tst
}

func (s *LinestringDefContext) Coordinate(i int) ICoordinateContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICoordinateContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICoordinateContext)
}

func (s *LinestringDefContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *LinestringDefContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *LinestringDefContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *LinestringDefContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LinestringDefContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LinestringDefContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterLinestringDef(s)
	}
}

func (s *LinestringDefContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitLinestringDef(s)
	}
}

func (p *CqlParser) LinestringDef() (localctx ILinestringDefContext) {
	localctx = NewLinestringDefContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 56, CqlParserRULE_linestringDef)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(305)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(306)
		p.Coordinate()
	}
	p.SetState(311)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserCOMMA {
		{
			p.SetState(307)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(308)
			p.Coordinate()
		}

		p.SetState(313)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(314)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPolygonContext is an interface to support dynamic dispatch.
type IPolygonContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	POLYGON() antlr.TerminalNode
	PolygonDef() IPolygonDefContext

	// IsPolygonContext differentiates from other interfaces.
	IsPolygonContext()
}

type PolygonContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPolygonContext() *PolygonContext {
	var p = new(PolygonContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_polygon
	return p
}

func InitEmptyPolygonContext(p *PolygonContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_polygon
}

func (*PolygonContext) IsPolygonContext() {}

func NewPolygonContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PolygonContext {
	var p = new(PolygonContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_polygon

	return p
}

func (s *PolygonContext) GetParser() antlr.Parser { return s.parser }

func (s *PolygonContext) POLYGON() antlr.TerminalNode {
	return s.GetToken(CqlParserPOLYGON, 0)
}

func (s *PolygonContext) PolygonDef() IPolygonDefContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPolygonDefContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPolygonDefContext)
}

func (s *PolygonContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PolygonContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PolygonContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterPolygon(s)
	}
}

func (s *PolygonContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitPolygon(s)
	}
}

func (p *CqlParser) Polygon() (localctx IPolygonContext) {
	localctx = NewPolygonContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 58, CqlParserRULE_polygon)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(316)
		p.Match(CqlParserPOLYGON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(317)
		p.PolygonDef()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPolygonDefContext is an interface to support dynamic dispatch.
type IPolygonDefContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LEFTPAREN() antlr.TerminalNode
	AllLinestringDef() []ILinestringDefContext
	LinestringDef(i int) ILinestringDefContext
	RIGHTPAREN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsPolygonDefContext differentiates from other interfaces.
	IsPolygonDefContext()
}

type PolygonDefContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPolygonDefContext() *PolygonDefContext {
	var p = new(PolygonDefContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_polygonDef
	return p
}

func InitEmptyPolygonDefContext(p *PolygonDefContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_polygonDef
}

func (*PolygonDefContext) IsPolygonDefContext() {}

func NewPolygonDefContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PolygonDefContext {
	var p = new(PolygonDefContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_polygonDef

	return p
}

func (s *PolygonDefContext) GetParser() antlr.Parser { return s.parser }

func (s *PolygonDefContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *PolygonDefContext) AllLinestringDef() []ILinestringDefContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILinestringDefContext); ok {
			len++
		}
	}

	tst := make([]ILinestringDefContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILinestringDefContext); ok {
			tst[i] = t.(ILinestringDefContext)
			i++
		}
	}

	return tst
}

func (s *PolygonDefContext) LinestringDef(i int) ILinestringDefContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILinestringDefContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILinestringDefContext)
}

func (s *PolygonDefContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *PolygonDefContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *PolygonDefContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *PolygonDefContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PolygonDefContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PolygonDefContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterPolygonDef(s)
	}
}

func (s *PolygonDefContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitPolygonDef(s)
	}
}

func (p *CqlParser) PolygonDef() (localctx IPolygonDefContext) {
	localctx = NewPolygonDefContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 60, CqlParserRULE_polygonDef)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(319)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(320)
		p.LinestringDef()
	}
	p.SetState(325)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserCOMMA {
		{
			p.SetState(321)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(322)
			p.LinestringDef()
		}

		p.SetState(327)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(328)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMultiPointContext is an interface to support dynamic dispatch.
type IMultiPointContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	MULTIPOINT() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	AllMultiPointDef() []IMultiPointDefContext
	MultiPointDef(i int) IMultiPointDefContext
	RIGHTPAREN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsMultiPointContext differentiates from other interfaces.
	IsMultiPointContext()
}

type MultiPointContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiPointContext() *MultiPointContext {
	var p = new(MultiPointContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_multiPoint
	return p
}

func InitEmptyMultiPointContext(p *MultiPointContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_multiPoint
}

func (*MultiPointContext) IsMultiPointContext() {}

func NewMultiPointContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MultiPointContext {
	var p = new(MultiPointContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_multiPoint

	return p
}

func (s *MultiPointContext) GetParser() antlr.Parser { return s.parser }

func (s *MultiPointContext) MULTIPOINT() antlr.TerminalNode {
	return s.GetToken(CqlParserMULTIPOINT, 0)
}

func (s *MultiPointContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *MultiPointContext) AllMultiPointDef() []IMultiPointDefContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IMultiPointDefContext); ok {
			len++
		}
	}

	tst := make([]IMultiPointDefContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IMultiPointDefContext); ok {
			tst[i] = t.(IMultiPointDefContext)
			i++
		}
	}

	return tst
}

func (s *MultiPointContext) MultiPointDef(i int) IMultiPointDefContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMultiPointDefContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMultiPointDefContext)
}

func (s *MultiPointContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *MultiPointContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *MultiPointContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *MultiPointContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiPointContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MultiPointContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterMultiPoint(s)
	}
}

func (s *MultiPointContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitMultiPoint(s)
	}
}

func (p *CqlParser) MultiPoint() (localctx IMultiPointContext) {
	localctx = NewMultiPointContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 62, CqlParserRULE_multiPoint)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(330)
		p.Match(CqlParserMULTIPOINT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(331)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(332)
		p.MultiPointDef()
	}
	p.SetState(337)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserCOMMA {
		{
			p.SetState(333)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(334)
			p.MultiPointDef()
		}

		p.SetState(339)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(340)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMultiPointDefContext is an interface to support dynamic dispatch.
type IMultiPointDefContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LEFTPAREN() antlr.TerminalNode
	Coordinate() ICoordinateContext
	RIGHTPAREN() antlr.TerminalNode

	// IsMultiPointDefContext differentiates from other interfaces.
	IsMultiPointDefContext()
}

type MultiPointDefContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiPointDefContext() *MultiPointDefContext {
	var p = new(MultiPointDefContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_multiPointDef
	return p
}

func InitEmptyMultiPointDefContext(p *MultiPointDefContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_multiPointDef
}

func (*MultiPointDefContext) IsMultiPointDefContext() {}

func NewMultiPointDefContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MultiPointDefContext {
	var p = new(MultiPointDefContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_multiPointDef

	return p
}

func (s *MultiPointDefContext) GetParser() antlr.Parser { return s.parser }

func (s *MultiPointDefContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *MultiPointDefContext) Coordinate() ICoordinateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICoordinateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICoordinateContext)
}

func (s *MultiPointDefContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *MultiPointDefContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiPointDefContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MultiPointDefContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterMultiPointDef(s)
	}
}

func (s *MultiPointDefContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitMultiPointDef(s)
	}
}

func (p *CqlParser) MultiPointDef() (localctx IMultiPointDefContext) {
	localctx = NewMultiPointDefContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 64, CqlParserRULE_multiPointDef)
	p.SetState(347)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CqlParserLEFTPAREN:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(342)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(343)
			p.Coordinate()
		}
		{
			p.SetState(344)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CqlParserNumericLiteral:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(346)
			p.Coordinate()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMultiLinestringContext is an interface to support dynamic dispatch.
type IMultiLinestringContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	MULTILINESTRING() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	AllLinestringDef() []ILinestringDefContext
	LinestringDef(i int) ILinestringDefContext
	RIGHTPAREN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsMultiLinestringContext differentiates from other interfaces.
	IsMultiLinestringContext()
}

type MultiLinestringContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiLinestringContext() *MultiLinestringContext {
	var p = new(MultiLinestringContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_multiLinestring
	return p
}

func InitEmptyMultiLinestringContext(p *MultiLinestringContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_multiLinestring
}

func (*MultiLinestringContext) IsMultiLinestringContext() {}

func NewMultiLinestringContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MultiLinestringContext {
	var p = new(MultiLinestringContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_multiLinestring

	return p
}

func (s *MultiLinestringContext) GetParser() antlr.Parser { return s.parser }

func (s *MultiLinestringContext) MULTILINESTRING() antlr.TerminalNode {
	return s.GetToken(CqlParserMULTILINESTRING, 0)
}

func (s *MultiLinestringContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *MultiLinestringContext) AllLinestringDef() []ILinestringDefContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILinestringDefContext); ok {
			len++
		}
	}

	tst := make([]ILinestringDefContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILinestringDefContext); ok {
			tst[i] = t.(ILinestringDefContext)
			i++
		}
	}

	return tst
}

func (s *MultiLinestringContext) LinestringDef(i int) ILinestringDefContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILinestringDefContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILinestringDefContext)
}

func (s *MultiLinestringContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *MultiLinestringContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *MultiLinestringContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *MultiLinestringContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiLinestringContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MultiLinestringContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterMultiLinestring(s)
	}
}

func (s *MultiLinestringContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitMultiLinestring(s)
	}
}

func (p *CqlParser) MultiLinestring() (localctx IMultiLinestringContext) {
	localctx = NewMultiLinestringContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 66, CqlParserRULE_multiLinestring)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(349)
		p.Match(CqlParserMULTILINESTRING)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(350)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(351)
		p.LinestringDef()
	}
	p.SetState(356)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserCOMMA {
		{
			p.SetState(352)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(353)
			p.LinestringDef()
		}

		p.SetState(358)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(359)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMultiPolygonContext is an interface to support dynamic dispatch.
type IMultiPolygonContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	MULTIPOLYGON() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	AllPolygonDef() []IPolygonDefContext
	PolygonDef(i int) IPolygonDefContext
	RIGHTPAREN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsMultiPolygonContext differentiates from other interfaces.
	IsMultiPolygonContext()
}

type MultiPolygonContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiPolygonContext() *MultiPolygonContext {
	var p = new(MultiPolygonContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_multiPolygon
	return p
}

func InitEmptyMultiPolygonContext(p *MultiPolygonContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_multiPolygon
}

func (*MultiPolygonContext) IsMultiPolygonContext() {}

func NewMultiPolygonContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MultiPolygonContext {
	var p = new(MultiPolygonContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_multiPolygon

	return p
}

func (s *MultiPolygonContext) GetParser() antlr.Parser { return s.parser }

func (s *MultiPolygonContext) MULTIPOLYGON() antlr.TerminalNode {
	return s.GetToken(CqlParserMULTIPOLYGON, 0)
}

func (s *MultiPolygonContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *MultiPolygonContext) AllPolygonDef() []IPolygonDefContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IPolygonDefContext); ok {
			len++
		}
	}

	tst := make([]IPolygonDefContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IPolygonDefContext); ok {
			tst[i] = t.(IPolygonDefContext)
			i++
		}
	}

	return tst
}

func (s *MultiPolygonContext) PolygonDef(i int) IPolygonDefContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPolygonDefContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPolygonDefContext)
}

func (s *MultiPolygonContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *MultiPolygonContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *MultiPolygonContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *MultiPolygonContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiPolygonContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MultiPolygonContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterMultiPolygon(s)
	}
}

func (s *MultiPolygonContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitMultiPolygon(s)
	}
}

func (p *CqlParser) MultiPolygon() (localctx IMultiPolygonContext) {
	localctx = NewMultiPolygonContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 68, CqlParserRULE_multiPolygon)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(361)
		p.Match(CqlParserMULTIPOLYGON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(362)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(363)
		p.PolygonDef()
	}
	p.SetState(368)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserCOMMA {
		{
			p.SetState(364)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(365)
			p.PolygonDef()
		}

		p.SetState(370)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(371)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IGeometryCollectionContext is an interface to support dynamic dispatch.
type IGeometryCollectionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	GEOMETRYCOLLECTION() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	AllGeometryLiteral() []IGeometryLiteralContext
	GeometryLiteral(i int) IGeometryLiteralContext
	RIGHTPAREN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsGeometryCollectionContext differentiates from other interfaces.
	IsGeometryCollectionContext()
}

type GeometryCollectionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyGeometryCollectionContext() *GeometryCollectionContext {
	var p = new(GeometryCollectionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_geometryCollection
	return p
}

func InitEmptyGeometryCollectionContext(p *GeometryCollectionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_geometryCollection
}

func (*GeometryCollectionContext) IsGeometryCollectionContext() {}

func NewGeometryCollectionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *GeometryCollectionContext {
	var p = new(GeometryCollectionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_geometryCollection

	return p
}

func (s *GeometryCollectionContext) GetParser() antlr.Parser { return s.parser }

func (s *GeometryCollectionContext) GEOMETRYCOLLECTION() antlr.TerminalNode {
	return s.GetToken(CqlParserGEOMETRYCOLLECTION, 0)
}

func (s *GeometryCollectionContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *GeometryCollectionContext) AllGeometryLiteral() []IGeometryLiteralContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IGeometryLiteralContext); ok {
			len++
		}
	}

	tst := make([]IGeometryLiteralContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IGeometryLiteralContext); ok {
			tst[i] = t.(IGeometryLiteralContext)
			i++
		}
	}

	return tst
}

func (s *GeometryCollectionContext) GeometryLiteral(i int) IGeometryLiteralContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGeometryLiteralContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGeometryLiteralContext)
}

func (s *GeometryCollectionContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *GeometryCollectionContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *GeometryCollectionContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *GeometryCollectionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *GeometryCollectionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *GeometryCollectionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterGeometryCollection(s)
	}
}

func (s *GeometryCollectionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitGeometryCollection(s)
	}
}

func (p *CqlParser) GeometryCollection() (localctx IGeometryCollectionContext) {
	localctx = NewGeometryCollectionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 70, CqlParserRULE_geometryCollection)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(373)
		p.Match(CqlParserGEOMETRYCOLLECTION)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(374)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(375)
		p.GeometryLiteral()
	}
	p.SetState(380)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserCOMMA {
		{
			p.SetState(376)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(377)
			p.GeometryLiteral()
		}

		p.SetState(382)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(383)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBboxContext is an interface to support dynamic dispatch.
type IBboxContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BBOX() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	WestBoundLon() IWestBoundLonContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	SouthBoundLat() ISouthBoundLatContext
	EastBoundLon() IEastBoundLonContext
	NorthBoundLat() INorthBoundLatContext
	RIGHTPAREN() antlr.TerminalNode
	MinElev() IMinElevContext
	MaxElev() IMaxElevContext

	// IsBboxContext differentiates from other interfaces.
	IsBboxContext()
}

type BboxContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBboxContext() *BboxContext {
	var p = new(BboxContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_bbox
	return p
}

func InitEmptyBboxContext(p *BboxContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_bbox
}

func (*BboxContext) IsBboxContext() {}

func NewBboxContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BboxContext {
	var p = new(BboxContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_bbox

	return p
}

func (s *BboxContext) GetParser() antlr.Parser { return s.parser }

func (s *BboxContext) BBOX() antlr.TerminalNode {
	return s.GetToken(CqlParserBBOX, 0)
}

func (s *BboxContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *BboxContext) WestBoundLon() IWestBoundLonContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWestBoundLonContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IWestBoundLonContext)
}

func (s *BboxContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *BboxContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *BboxContext) SouthBoundLat() ISouthBoundLatContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISouthBoundLatContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISouthBoundLatContext)
}

func (s *BboxContext) EastBoundLon() IEastBoundLonContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEastBoundLonContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEastBoundLonContext)
}

func (s *BboxContext) NorthBoundLat() INorthBoundLatContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INorthBoundLatContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INorthBoundLatContext)
}

func (s *BboxContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *BboxContext) MinElev() IMinElevContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMinElevContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMinElevContext)
}

func (s *BboxContext) MaxElev() IMaxElevContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMaxElevContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMaxElevContext)
}

func (s *BboxContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BboxContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BboxContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterBbox(s)
	}
}

func (s *BboxContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitBbox(s)
	}
}

func (p *CqlParser) Bbox() (localctx IBboxContext) {
	localctx = NewBboxContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 72, CqlParserRULE_bbox)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(385)
		p.Match(CqlParserBBOX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(386)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(387)
		p.WestBoundLon()
	}
	{
		p.SetState(388)
		p.Match(CqlParserCOMMA)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(389)
		p.SouthBoundLat()
	}
	{
		p.SetState(390)
		p.Match(CqlParserCOMMA)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(394)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 27, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(391)
			p.MinElev()
		}
		{
			p.SetState(392)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	{
		p.SetState(396)
		p.EastBoundLon()
	}
	{
		p.SetState(397)
		p.Match(CqlParserCOMMA)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(398)
		p.NorthBoundLat()
	}
	p.SetState(401)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CqlParserCOMMA {
		{
			p.SetState(399)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(400)
			p.MaxElev()
		}

	}
	{
		p.SetState(403)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICoordinateContext is an interface to support dynamic dispatch.
type ICoordinateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	XCoord() IXCoordContext
	YCoord() IYCoordContext
	ZCoord() IZCoordContext

	// IsCoordinateContext differentiates from other interfaces.
	IsCoordinateContext()
}

type CoordinateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCoordinateContext() *CoordinateContext {
	var p = new(CoordinateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_coordinate
	return p
}

func InitEmptyCoordinateContext(p *CoordinateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_coordinate
}

func (*CoordinateContext) IsCoordinateContext() {}

func NewCoordinateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CoordinateContext {
	var p = new(CoordinateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_coordinate

	return p
}

func (s *CoordinateContext) GetParser() antlr.Parser { return s.parser }

func (s *CoordinateContext) XCoord() IXCoordContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IXCoordContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IXCoordContext)
}

func (s *CoordinateContext) YCoord() IYCoordContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IYCoordContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IYCoordContext)
}

func (s *CoordinateContext) ZCoord() IZCoordContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IZCoordContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IZCoordContext)
}

func (s *CoordinateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CoordinateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CoordinateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterCoordinate(s)
	}
}

func (s *CoordinateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitCoordinate(s)
	}
}

func (p *CqlParser) Coordinate() (localctx ICoordinateContext) {
	localctx = NewCoordinateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 74, CqlParserRULE_coordinate)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(405)
		p.XCoord()
	}
	{
		p.SetState(406)
		p.YCoord()
	}
	p.SetState(408)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CqlParserNumericLiteral {
		{
			p.SetState(407)
			p.ZCoord()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IXCoordContext is an interface to support dynamic dispatch.
type IXCoordContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsXCoordContext differentiates from other interfaces.
	IsXCoordContext()
}

type XCoordContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyXCoordContext() *XCoordContext {
	var p = new(XCoordContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_xCoord
	return p
}

func InitEmptyXCoordContext(p *XCoordContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_xCoord
}

func (*XCoordContext) IsXCoordContext() {}

func NewXCoordContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *XCoordContext {
	var p = new(XCoordContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_xCoord

	return p
}

func (s *XCoordContext) GetParser() antlr.Parser { return s.parser }

func (s *XCoordContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *XCoordContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *XCoordContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *XCoordContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterXCoord(s)
	}
}

func (s *XCoordContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitXCoord(s)
	}
}

func (p *CqlParser) XCoord() (localctx IXCoordContext) {
	localctx = NewXCoordContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 76, CqlParserRULE_xCoord)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(410)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IYCoordContext is an interface to support dynamic dispatch.
type IYCoordContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsYCoordContext differentiates from other interfaces.
	IsYCoordContext()
}

type YCoordContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyYCoordContext() *YCoordContext {
	var p = new(YCoordContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_yCoord
	return p
}

func InitEmptyYCoordContext(p *YCoordContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_yCoord
}

func (*YCoordContext) IsYCoordContext() {}

func NewYCoordContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *YCoordContext {
	var p = new(YCoordContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_yCoord

	return p
}

func (s *YCoordContext) GetParser() antlr.Parser { return s.parser }

func (s *YCoordContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *YCoordContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *YCoordContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *YCoordContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterYCoord(s)
	}
}

func (s *YCoordContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitYCoord(s)
	}
}

func (p *CqlParser) YCoord() (localctx IYCoordContext) {
	localctx = NewYCoordContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 78, CqlParserRULE_yCoord)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(412)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IZCoordContext is an interface to support dynamic dispatch.
type IZCoordContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsZCoordContext differentiates from other interfaces.
	IsZCoordContext()
}

type ZCoordContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyZCoordContext() *ZCoordContext {
	var p = new(ZCoordContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_zCoord
	return p
}

func InitEmptyZCoordContext(p *ZCoordContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_zCoord
}

func (*ZCoordContext) IsZCoordContext() {}

func NewZCoordContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ZCoordContext {
	var p = new(ZCoordContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_zCoord

	return p
}

func (s *ZCoordContext) GetParser() antlr.Parser { return s.parser }

func (s *ZCoordContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *ZCoordContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ZCoordContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ZCoordContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterZCoord(s)
	}
}

func (s *ZCoordContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitZCoord(s)
	}
}

func (p *CqlParser) ZCoord() (localctx IZCoordContext) {
	localctx = NewZCoordContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 80, CqlParserRULE_zCoord)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(414)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IWestBoundLonContext is an interface to support dynamic dispatch.
type IWestBoundLonContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsWestBoundLonContext differentiates from other interfaces.
	IsWestBoundLonContext()
}

type WestBoundLonContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWestBoundLonContext() *WestBoundLonContext {
	var p = new(WestBoundLonContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_westBoundLon
	return p
}

func InitEmptyWestBoundLonContext(p *WestBoundLonContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_westBoundLon
}

func (*WestBoundLonContext) IsWestBoundLonContext() {}

func NewWestBoundLonContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WestBoundLonContext {
	var p = new(WestBoundLonContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_westBoundLon

	return p
}

func (s *WestBoundLonContext) GetParser() antlr.Parser { return s.parser }

func (s *WestBoundLonContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *WestBoundLonContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WestBoundLonContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WestBoundLonContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterWestBoundLon(s)
	}
}

func (s *WestBoundLonContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitWestBoundLon(s)
	}
}

func (p *CqlParser) WestBoundLon() (localctx IWestBoundLonContext) {
	localctx = NewWestBoundLonContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 82, CqlParserRULE_westBoundLon)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(416)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IEastBoundLonContext is an interface to support dynamic dispatch.
type IEastBoundLonContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsEastBoundLonContext differentiates from other interfaces.
	IsEastBoundLonContext()
}

type EastBoundLonContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyEastBoundLonContext() *EastBoundLonContext {
	var p = new(EastBoundLonContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_eastBoundLon
	return p
}

func InitEmptyEastBoundLonContext(p *EastBoundLonContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_eastBoundLon
}

func (*EastBoundLonContext) IsEastBoundLonContext() {}

func NewEastBoundLonContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *EastBoundLonContext {
	var p = new(EastBoundLonContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_eastBoundLon

	return p
}

func (s *EastBoundLonContext) GetParser() antlr.Parser { return s.parser }

func (s *EastBoundLonContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *EastBoundLonContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EastBoundLonContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *EastBoundLonContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterEastBoundLon(s)
	}
}

func (s *EastBoundLonContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitEastBoundLon(s)
	}
}

func (p *CqlParser) EastBoundLon() (localctx IEastBoundLonContext) {
	localctx = NewEastBoundLonContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 84, CqlParserRULE_eastBoundLon)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(418)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INorthBoundLatContext is an interface to support dynamic dispatch.
type INorthBoundLatContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsNorthBoundLatContext differentiates from other interfaces.
	IsNorthBoundLatContext()
}

type NorthBoundLatContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNorthBoundLatContext() *NorthBoundLatContext {
	var p = new(NorthBoundLatContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_northBoundLat
	return p
}

func InitEmptyNorthBoundLatContext(p *NorthBoundLatContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_northBoundLat
}

func (*NorthBoundLatContext) IsNorthBoundLatContext() {}

func NewNorthBoundLatContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NorthBoundLatContext {
	var p = new(NorthBoundLatContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_northBoundLat

	return p
}

func (s *NorthBoundLatContext) GetParser() antlr.Parser { return s.parser }

func (s *NorthBoundLatContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *NorthBoundLatContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NorthBoundLatContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NorthBoundLatContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterNorthBoundLat(s)
	}
}

func (s *NorthBoundLatContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitNorthBoundLat(s)
	}
}

func (p *CqlParser) NorthBoundLat() (localctx INorthBoundLatContext) {
	localctx = NewNorthBoundLatContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 86, CqlParserRULE_northBoundLat)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(420)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISouthBoundLatContext is an interface to support dynamic dispatch.
type ISouthBoundLatContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsSouthBoundLatContext differentiates from other interfaces.
	IsSouthBoundLatContext()
}

type SouthBoundLatContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySouthBoundLatContext() *SouthBoundLatContext {
	var p = new(SouthBoundLatContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_southBoundLat
	return p
}

func InitEmptySouthBoundLatContext(p *SouthBoundLatContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_southBoundLat
}

func (*SouthBoundLatContext) IsSouthBoundLatContext() {}

func NewSouthBoundLatContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SouthBoundLatContext {
	var p = new(SouthBoundLatContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_southBoundLat

	return p
}

func (s *SouthBoundLatContext) GetParser() antlr.Parser { return s.parser }

func (s *SouthBoundLatContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *SouthBoundLatContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SouthBoundLatContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SouthBoundLatContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterSouthBoundLat(s)
	}
}

func (s *SouthBoundLatContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitSouthBoundLat(s)
	}
}

func (p *CqlParser) SouthBoundLat() (localctx ISouthBoundLatContext) {
	localctx = NewSouthBoundLatContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 88, CqlParserRULE_southBoundLat)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(422)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMinElevContext is an interface to support dynamic dispatch.
type IMinElevContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsMinElevContext differentiates from other interfaces.
	IsMinElevContext()
}

type MinElevContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMinElevContext() *MinElevContext {
	var p = new(MinElevContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_minElev
	return p
}

func InitEmptyMinElevContext(p *MinElevContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_minElev
}

func (*MinElevContext) IsMinElevContext() {}

func NewMinElevContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MinElevContext {
	var p = new(MinElevContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_minElev

	return p
}

func (s *MinElevContext) GetParser() antlr.Parser { return s.parser }

func (s *MinElevContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *MinElevContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MinElevContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MinElevContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterMinElev(s)
	}
}

func (s *MinElevContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitMinElev(s)
	}
}

func (p *CqlParser) MinElev() (localctx IMinElevContext) {
	localctx = NewMinElevContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 90, CqlParserRULE_minElev)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(424)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMaxElevContext is an interface to support dynamic dispatch.
type IMaxElevContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NumericLiteral() antlr.TerminalNode

	// IsMaxElevContext differentiates from other interfaces.
	IsMaxElevContext()
}

type MaxElevContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMaxElevContext() *MaxElevContext {
	var p = new(MaxElevContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_maxElev
	return p
}

func InitEmptyMaxElevContext(p *MaxElevContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_maxElev
}

func (*MaxElevContext) IsMaxElevContext() {}

func NewMaxElevContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MaxElevContext {
	var p = new(MaxElevContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_maxElev

	return p
}

func (s *MaxElevContext) GetParser() antlr.Parser { return s.parser }

func (s *MaxElevContext) NumericLiteral() antlr.TerminalNode {
	return s.GetToken(CqlParserNumericLiteral, 0)
}

func (s *MaxElevContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MaxElevContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MaxElevContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterMaxElev(s)
	}
}

func (s *MaxElevContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitMaxElev(s)
	}
}

func (p *CqlParser) MaxElev() (localctx IMaxElevContext) {
	localctx = NewMaxElevContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 92, CqlParserRULE_maxElev)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(426)
		p.Match(CqlParserNumericLiteral)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITemporalPredicateContext is an interface to support dynamic dispatch.
type ITemporalPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TemporalFunction() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	AllTemporalExpression() []ITemporalExpressionContext
	TemporalExpression(i int) ITemporalExpressionContext
	COMMA() antlr.TerminalNode
	RIGHTPAREN() antlr.TerminalNode

	// IsTemporalPredicateContext differentiates from other interfaces.
	IsTemporalPredicateContext()
}

type TemporalPredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTemporalPredicateContext() *TemporalPredicateContext {
	var p = new(TemporalPredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_temporalPredicate
	return p
}

func InitEmptyTemporalPredicateContext(p *TemporalPredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_temporalPredicate
}

func (*TemporalPredicateContext) IsTemporalPredicateContext() {}

func NewTemporalPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TemporalPredicateContext {
	var p = new(TemporalPredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_temporalPredicate

	return p
}

func (s *TemporalPredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *TemporalPredicateContext) TemporalFunction() antlr.TerminalNode {
	return s.GetToken(CqlParserTemporalFunction, 0)
}

func (s *TemporalPredicateContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *TemporalPredicateContext) AllTemporalExpression() []ITemporalExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITemporalExpressionContext); ok {
			len++
		}
	}

	tst := make([]ITemporalExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITemporalExpressionContext); ok {
			tst[i] = t.(ITemporalExpressionContext)
			i++
		}
	}

	return tst
}

func (s *TemporalPredicateContext) TemporalExpression(i int) ITemporalExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITemporalExpressionContext)
}

func (s *TemporalPredicateContext) COMMA() antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, 0)
}

func (s *TemporalPredicateContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *TemporalPredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TemporalPredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TemporalPredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterTemporalPredicate(s)
	}
}

func (s *TemporalPredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitTemporalPredicate(s)
	}
}

func (p *CqlParser) TemporalPredicate() (localctx ITemporalPredicateContext) {
	localctx = NewTemporalPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 94, CqlParserRULE_temporalPredicate)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(428)
		p.Match(CqlParserTemporalFunction)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(429)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(430)
		p.TemporalExpression()
	}
	{
		p.SetState(431)
		p.Match(CqlParserCOMMA)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(432)
		p.TemporalExpression()
	}
	{
		p.SetState(433)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITemporalExpressionContext is an interface to support dynamic dispatch.
type ITemporalExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TemporalClause() ITemporalClauseContext
	PropertyName() IPropertyNameContext
	Function() IFunctionContext

	// IsTemporalExpressionContext differentiates from other interfaces.
	IsTemporalExpressionContext()
}

type TemporalExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTemporalExpressionContext() *TemporalExpressionContext {
	var p = new(TemporalExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_temporalExpression
	return p
}

func InitEmptyTemporalExpressionContext(p *TemporalExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_temporalExpression
}

func (*TemporalExpressionContext) IsTemporalExpressionContext() {}

func NewTemporalExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TemporalExpressionContext {
	var p = new(TemporalExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_temporalExpression

	return p
}

func (s *TemporalExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *TemporalExpressionContext) TemporalClause() ITemporalClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITemporalClauseContext)
}

func (s *TemporalExpressionContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *TemporalExpressionContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *TemporalExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TemporalExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TemporalExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterTemporalExpression(s)
	}
}

func (s *TemporalExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitTemporalExpression(s)
	}
}

func (p *CqlParser) TemporalExpression() (localctx ITemporalExpressionContext) {
	localctx = NewTemporalExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 96, CqlParserRULE_temporalExpression)
	p.SetState(438)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 30, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(435)
			p.TemporalClause()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(436)
			p.PropertyName()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(437)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITemporalClauseContext is an interface to support dynamic dispatch.
type ITemporalClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	InstantInstance() IInstantInstanceContext
	Interval() IIntervalContext

	// IsTemporalClauseContext differentiates from other interfaces.
	IsTemporalClauseContext()
}

type TemporalClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTemporalClauseContext() *TemporalClauseContext {
	var p = new(TemporalClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_temporalClause
	return p
}

func InitEmptyTemporalClauseContext(p *TemporalClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_temporalClause
}

func (*TemporalClauseContext) IsTemporalClauseContext() {}

func NewTemporalClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TemporalClauseContext {
	var p = new(TemporalClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_temporalClause

	return p
}

func (s *TemporalClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *TemporalClauseContext) InstantInstance() IInstantInstanceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInstantInstanceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInstantInstanceContext)
}

func (s *TemporalClauseContext) Interval() IIntervalContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIntervalContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIntervalContext)
}

func (s *TemporalClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TemporalClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TemporalClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterTemporalClause(s)
	}
}

func (s *TemporalClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitTemporalClause(s)
	}
}

func (p *CqlParser) TemporalClause() (localctx ITemporalClauseContext) {
	localctx = NewTemporalClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 98, CqlParserRULE_temporalClause)
	p.SetState(442)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CqlParserLEFTPAREN, CqlParserDATE, CqlParserTIMESTAMP:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(440)
			p.InstantInstance()
		}

	case CqlParserINTERVAL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(441)
			p.Interval()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IInstantInstanceContext is an interface to support dynamic dispatch.
type IInstantInstanceContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DATE() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	DateString() antlr.TerminalNode
	RIGHTPAREN() antlr.TerminalNode
	TIMESTAMP() antlr.TerminalNode
	TimestampString() antlr.TerminalNode

	// IsInstantInstanceContext differentiates from other interfaces.
	IsInstantInstanceContext()
}

type InstantInstanceContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInstantInstanceContext() *InstantInstanceContext {
	var p = new(InstantInstanceContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_instantInstance
	return p
}

func InitEmptyInstantInstanceContext(p *InstantInstanceContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_instantInstance
}

func (*InstantInstanceContext) IsInstantInstanceContext() {}

func NewInstantInstanceContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InstantInstanceContext {
	var p = new(InstantInstanceContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_instantInstance

	return p
}

func (s *InstantInstanceContext) GetParser() antlr.Parser { return s.parser }

func (s *InstantInstanceContext) DATE() antlr.TerminalNode {
	return s.GetToken(CqlParserDATE, 0)
}

func (s *InstantInstanceContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *InstantInstanceContext) DateString() antlr.TerminalNode {
	return s.GetToken(CqlParserDateString, 0)
}

func (s *InstantInstanceContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *InstantInstanceContext) TIMESTAMP() antlr.TerminalNode {
	return s.GetToken(CqlParserTIMESTAMP, 0)
}

func (s *InstantInstanceContext) TimestampString() antlr.TerminalNode {
	return s.GetToken(CqlParserTimestampString, 0)
}

func (s *InstantInstanceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InstantInstanceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InstantInstanceContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterInstantInstance(s)
	}
}

func (s *InstantInstanceContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitInstantInstance(s)
	}
}

func (p *CqlParser) InstantInstance() (localctx IInstantInstanceContext) {
	localctx = NewInstantInstanceContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 100, CqlParserRULE_instantInstance)
	p.SetState(454)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CqlParserDATE:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(444)
			p.Match(CqlParserDATE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(445)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(446)
			p.Match(CqlParserDateString)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(447)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CqlParserTIMESTAMP:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(448)
			p.Match(CqlParserTIMESTAMP)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(449)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(450)
			p.Match(CqlParserTimestampString)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(451)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CqlParserLEFTPAREN:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(452)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(453)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIntervalContext is an interface to support dynamic dispatch.
type IIntervalContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	INTERVAL() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	AllIntervalParameter() []IIntervalParameterContext
	IntervalParameter(i int) IIntervalParameterContext
	COMMA() antlr.TerminalNode
	RIGHTPAREN() antlr.TerminalNode

	// IsIntervalContext differentiates from other interfaces.
	IsIntervalContext()
}

type IntervalContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIntervalContext() *IntervalContext {
	var p = new(IntervalContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_interval
	return p
}

func InitEmptyIntervalContext(p *IntervalContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_interval
}

func (*IntervalContext) IsIntervalContext() {}

func NewIntervalContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IntervalContext {
	var p = new(IntervalContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_interval

	return p
}

func (s *IntervalContext) GetParser() antlr.Parser { return s.parser }

func (s *IntervalContext) INTERVAL() antlr.TerminalNode {
	return s.GetToken(CqlParserINTERVAL, 0)
}

func (s *IntervalContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *IntervalContext) AllIntervalParameter() []IIntervalParameterContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIntervalParameterContext); ok {
			len++
		}
	}

	tst := make([]IIntervalParameterContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIntervalParameterContext); ok {
			tst[i] = t.(IIntervalParameterContext)
			i++
		}
	}

	return tst
}

func (s *IntervalContext) IntervalParameter(i int) IIntervalParameterContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIntervalParameterContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIntervalParameterContext)
}

func (s *IntervalContext) COMMA() antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, 0)
}

func (s *IntervalContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *IntervalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IntervalContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IntervalContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterInterval(s)
	}
}

func (s *IntervalContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitInterval(s)
	}
}

func (p *CqlParser) Interval() (localctx IIntervalContext) {
	localctx = NewIntervalContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 102, CqlParserRULE_interval)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(456)
		p.Match(CqlParserINTERVAL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(457)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(458)
		p.IntervalParameter()
	}
	{
		p.SetState(459)
		p.Match(CqlParserCOMMA)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(460)
		p.IntervalParameter()
	}
	{
		p.SetState(461)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIntervalParameterContext is an interface to support dynamic dispatch.
type IIntervalParameterContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	PropertyName() IPropertyNameContext
	DateString() antlr.TerminalNode
	TimestampString() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	RIGHTPAREN() antlr.TerminalNode
	DotDotString() antlr.TerminalNode
	Function() IFunctionContext

	// IsIntervalParameterContext differentiates from other interfaces.
	IsIntervalParameterContext()
}

type IntervalParameterContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIntervalParameterContext() *IntervalParameterContext {
	var p = new(IntervalParameterContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_intervalParameter
	return p
}

func InitEmptyIntervalParameterContext(p *IntervalParameterContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_intervalParameter
}

func (*IntervalParameterContext) IsIntervalParameterContext() {}

func NewIntervalParameterContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IntervalParameterContext {
	var p = new(IntervalParameterContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_intervalParameter

	return p
}

func (s *IntervalParameterContext) GetParser() antlr.Parser { return s.parser }

func (s *IntervalParameterContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *IntervalParameterContext) DateString() antlr.TerminalNode {
	return s.GetToken(CqlParserDateString, 0)
}

func (s *IntervalParameterContext) TimestampString() antlr.TerminalNode {
	return s.GetToken(CqlParserTimestampString, 0)
}

func (s *IntervalParameterContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *IntervalParameterContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *IntervalParameterContext) DotDotString() antlr.TerminalNode {
	return s.GetToken(CqlParserDotDotString, 0)
}

func (s *IntervalParameterContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *IntervalParameterContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IntervalParameterContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IntervalParameterContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterIntervalParameter(s)
	}
}

func (s *IntervalParameterContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitIntervalParameter(s)
	}
}

func (p *CqlParser) IntervalParameter() (localctx IIntervalParameterContext) {
	localctx = NewIntervalParameterContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 104, CqlParserRULE_intervalParameter)
	p.SetState(470)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 33, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(463)
			p.PropertyName()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(464)
			p.Match(CqlParserDateString)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(465)
			p.Match(CqlParserTimestampString)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(466)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(467)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(468)
			p.Match(CqlParserDotDotString)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(469)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArrayPredicateContext is an interface to support dynamic dispatch.
type IArrayPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ArrayFunction() antlr.TerminalNode
	LEFTPAREN() antlr.TerminalNode
	AllArrayExpression() []IArrayExpressionContext
	ArrayExpression(i int) IArrayExpressionContext
	COMMA() antlr.TerminalNode
	RIGHTPAREN() antlr.TerminalNode

	// IsArrayPredicateContext differentiates from other interfaces.
	IsArrayPredicateContext()
}

type ArrayPredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArrayPredicateContext() *ArrayPredicateContext {
	var p = new(ArrayPredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_arrayPredicate
	return p
}

func InitEmptyArrayPredicateContext(p *ArrayPredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_arrayPredicate
}

func (*ArrayPredicateContext) IsArrayPredicateContext() {}

func NewArrayPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArrayPredicateContext {
	var p = new(ArrayPredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_arrayPredicate

	return p
}

func (s *ArrayPredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *ArrayPredicateContext) ArrayFunction() antlr.TerminalNode {
	return s.GetToken(CqlParserArrayFunction, 0)
}

func (s *ArrayPredicateContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *ArrayPredicateContext) AllArrayExpression() []IArrayExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IArrayExpressionContext); ok {
			len++
		}
	}

	tst := make([]IArrayExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IArrayExpressionContext); ok {
			tst[i] = t.(IArrayExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ArrayPredicateContext) ArrayExpression(i int) IArrayExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayExpressionContext)
}

func (s *ArrayPredicateContext) COMMA() antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, 0)
}

func (s *ArrayPredicateContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *ArrayPredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArrayPredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArrayPredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterArrayPredicate(s)
	}
}

func (s *ArrayPredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitArrayPredicate(s)
	}
}

func (p *CqlParser) ArrayPredicate() (localctx IArrayPredicateContext) {
	localctx = NewArrayPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 106, CqlParserRULE_arrayPredicate)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(472)
		p.Match(CqlParserArrayFunction)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(473)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(474)
		p.ArrayExpression()
	}
	{
		p.SetState(475)
		p.Match(CqlParserCOMMA)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(476)
		p.ArrayExpression()
	}
	{
		p.SetState(477)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArrayExpressionContext is an interface to support dynamic dispatch.
type IArrayExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	PropertyName() IPropertyNameContext
	ArrayClause() IArrayClauseContext
	Function() IFunctionContext

	// IsArrayExpressionContext differentiates from other interfaces.
	IsArrayExpressionContext()
}

type ArrayExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArrayExpressionContext() *ArrayExpressionContext {
	var p = new(ArrayExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_arrayExpression
	return p
}

func InitEmptyArrayExpressionContext(p *ArrayExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_arrayExpression
}

func (*ArrayExpressionContext) IsArrayExpressionContext() {}

func NewArrayExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArrayExpressionContext {
	var p = new(ArrayExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_arrayExpression

	return p
}

func (s *ArrayExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ArrayExpressionContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *ArrayExpressionContext) ArrayClause() IArrayClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayClauseContext)
}

func (s *ArrayExpressionContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *ArrayExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArrayExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArrayExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterArrayExpression(s)
	}
}

func (s *ArrayExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitArrayExpression(s)
	}
}

func (p *CqlParser) ArrayExpression() (localctx IArrayExpressionContext) {
	localctx = NewArrayExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 108, CqlParserRULE_arrayExpression)
	p.SetState(482)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 34, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(479)
			p.PropertyName()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(480)
			p.ArrayClause()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(481)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArrayClauseContext is an interface to support dynamic dispatch.
type IArrayClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LEFTPAREN() antlr.TerminalNode
	RIGHTPAREN() antlr.TerminalNode
	AllArrayElement() []IArrayElementContext
	ArrayElement(i int) IArrayElementContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsArrayClauseContext differentiates from other interfaces.
	IsArrayClauseContext()
}

type ArrayClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArrayClauseContext() *ArrayClauseContext {
	var p = new(ArrayClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_arrayClause
	return p
}

func InitEmptyArrayClauseContext(p *ArrayClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_arrayClause
}

func (*ArrayClauseContext) IsArrayClauseContext() {}

func NewArrayClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArrayClauseContext {
	var p = new(ArrayClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_arrayClause

	return p
}

func (s *ArrayClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *ArrayClauseContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *ArrayClauseContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *ArrayClauseContext) AllArrayElement() []IArrayElementContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IArrayElementContext); ok {
			len++
		}
	}

	tst := make([]IArrayElementContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IArrayElementContext); ok {
			tst[i] = t.(IArrayElementContext)
			i++
		}
	}

	return tst
}

func (s *ArrayClauseContext) ArrayElement(i int) IArrayElementContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayElementContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayElementContext)
}

func (s *ArrayClauseContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *ArrayClauseContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *ArrayClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArrayClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArrayClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterArrayClause(s)
	}
}

func (s *ArrayClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitArrayClause(s)
	}
}

func (p *CqlParser) ArrayClause() (localctx IArrayClauseContext) {
	localctx = NewArrayClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 110, CqlParserRULE_arrayClause)
	var _la int

	p.SetState(497)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 36, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(484)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(485)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(486)
			p.Match(CqlParserLEFTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(487)
			p.ArrayElement()
		}
		p.SetState(492)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CqlParserCOMMA {
			{
				p.SetState(488)
				p.Match(CqlParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(489)
				p.ArrayElement()
			}

			p.SetState(494)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(495)
			p.Match(CqlParserRIGHTPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArrayElementContext is an interface to support dynamic dispatch.
type IArrayElementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CharacterClause() ICharacterClauseContext
	NumericLiteral() INumericLiteralContext
	BooleanLiteral() IBooleanLiteralContext
	TemporalClause() ITemporalClauseContext
	ArrayClause() IArrayClauseContext
	PropertyName() IPropertyNameContext
	Function() IFunctionContext

	// IsArrayElementContext differentiates from other interfaces.
	IsArrayElementContext()
}

type ArrayElementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArrayElementContext() *ArrayElementContext {
	var p = new(ArrayElementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_arrayElement
	return p
}

func InitEmptyArrayElementContext(p *ArrayElementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_arrayElement
}

func (*ArrayElementContext) IsArrayElementContext() {}

func NewArrayElementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArrayElementContext {
	var p = new(ArrayElementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_arrayElement

	return p
}

func (s *ArrayElementContext) GetParser() antlr.Parser { return s.parser }

func (s *ArrayElementContext) CharacterClause() ICharacterClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICharacterClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICharacterClauseContext)
}

func (s *ArrayElementContext) NumericLiteral() INumericLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INumericLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INumericLiteralContext)
}

func (s *ArrayElementContext) BooleanLiteral() IBooleanLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanLiteralContext)
}

func (s *ArrayElementContext) TemporalClause() ITemporalClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITemporalClauseContext)
}

func (s *ArrayElementContext) ArrayClause() IArrayClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayClauseContext)
}

func (s *ArrayElementContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *ArrayElementContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *ArrayElementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArrayElementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArrayElementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterArrayElement(s)
	}
}

func (s *ArrayElementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitArrayElement(s)
	}
}

func (p *CqlParser) ArrayElement() (localctx IArrayElementContext) {
	localctx = NewArrayElementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 112, CqlParserRULE_arrayElement)
	p.SetState(506)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 37, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(499)
			p.CharacterClause()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(500)
			p.NumericLiteral()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(501)
			p.BooleanLiteral()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(502)
			p.TemporalClause()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(503)
			p.ArrayClause()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(504)
			p.PropertyName()
		}

	case 7:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(505)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFunctionContext is an interface to support dynamic dispatch.
type IFunctionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() antlr.TerminalNode
	ArgumentList() IArgumentListContext

	// IsFunctionContext differentiates from other interfaces.
	IsFunctionContext()
}

type FunctionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunctionContext() *FunctionContext {
	var p = new(FunctionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_function
	return p
}

func InitEmptyFunctionContext(p *FunctionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_function
}

func (*FunctionContext) IsFunctionContext() {}

func NewFunctionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FunctionContext {
	var p = new(FunctionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_function

	return p
}

func (s *FunctionContext) GetParser() antlr.Parser { return s.parser }

func (s *FunctionContext) Identifier() antlr.TerminalNode {
	return s.GetToken(CqlParserIdentifier, 0)
}

func (s *FunctionContext) ArgumentList() IArgumentListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArgumentListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArgumentListContext)
}

func (s *FunctionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FunctionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterFunction(s)
	}
}

func (s *FunctionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitFunction(s)
	}
}

func (p *CqlParser) Function() (localctx IFunctionContext) {
	localctx = NewFunctionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 114, CqlParserRULE_function)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(508)
		p.Match(CqlParserIdentifier)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(509)
		p.ArgumentList()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArgumentListContext is an interface to support dynamic dispatch.
type IArgumentListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LEFTPAREN() antlr.TerminalNode
	RIGHTPAREN() antlr.TerminalNode
	PositionalArgument() IPositionalArgumentContext

	// IsArgumentListContext differentiates from other interfaces.
	IsArgumentListContext()
}

type ArgumentListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArgumentListContext() *ArgumentListContext {
	var p = new(ArgumentListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_argumentList
	return p
}

func InitEmptyArgumentListContext(p *ArgumentListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_argumentList
}

func (*ArgumentListContext) IsArgumentListContext() {}

func NewArgumentListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArgumentListContext {
	var p = new(ArgumentListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_argumentList

	return p
}

func (s *ArgumentListContext) GetParser() antlr.Parser { return s.parser }

func (s *ArgumentListContext) LEFTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserLEFTPAREN, 0)
}

func (s *ArgumentListContext) RIGHTPAREN() antlr.TerminalNode {
	return s.GetToken(CqlParserRIGHTPAREN, 0)
}

func (s *ArgumentListContext) PositionalArgument() IPositionalArgumentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPositionalArgumentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPositionalArgumentContext)
}

func (s *ArgumentListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArgumentListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArgumentListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterArgumentList(s)
	}
}

func (s *ArgumentListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitArgumentList(s)
	}
}

func (p *CqlParser) ArgumentList() (localctx IArgumentListContext) {
	localctx = NewArgumentListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 116, CqlParserRULE_argumentList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(511)
		p.Match(CqlParserLEFTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(513)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&280044241152) != 0) || ((int64((_la-64)) & ^0x3f) == 0 && ((int64(1)<<(_la-64))&294919) != 0) {
		{
			p.SetState(512)
			p.PositionalArgument()
		}

	}
	{
		p.SetState(515)
		p.Match(CqlParserRIGHTPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPositionalArgumentContext is an interface to support dynamic dispatch.
type IPositionalArgumentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllArgument() []IArgumentContext
	Argument(i int) IArgumentContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsPositionalArgumentContext differentiates from other interfaces.
	IsPositionalArgumentContext()
}

type PositionalArgumentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPositionalArgumentContext() *PositionalArgumentContext {
	var p = new(PositionalArgumentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_positionalArgument
	return p
}

func InitEmptyPositionalArgumentContext(p *PositionalArgumentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_positionalArgument
}

func (*PositionalArgumentContext) IsPositionalArgumentContext() {}

func NewPositionalArgumentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PositionalArgumentContext {
	var p = new(PositionalArgumentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_positionalArgument

	return p
}

func (s *PositionalArgumentContext) GetParser() antlr.Parser { return s.parser }

func (s *PositionalArgumentContext) AllArgument() []IArgumentContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IArgumentContext); ok {
			len++
		}
	}

	tst := make([]IArgumentContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IArgumentContext); ok {
			tst[i] = t.(IArgumentContext)
			i++
		}
	}

	return tst
}

func (s *PositionalArgumentContext) Argument(i int) IArgumentContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArgumentContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArgumentContext)
}

func (s *PositionalArgumentContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CqlParserCOMMA)
}

func (s *PositionalArgumentContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CqlParserCOMMA, i)
}

func (s *PositionalArgumentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PositionalArgumentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PositionalArgumentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterPositionalArgument(s)
	}
}

func (s *PositionalArgumentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitPositionalArgument(s)
	}
}

func (p *CqlParser) PositionalArgument() (localctx IPositionalArgumentContext) {
	localctx = NewPositionalArgumentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 118, CqlParserRULE_positionalArgument)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(517)
		p.Argument()
	}
	p.SetState(522)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CqlParserCOMMA {
		{
			p.SetState(518)
			p.Match(CqlParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(519)
			p.Argument()
		}

		p.SetState(524)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArgumentContext is an interface to support dynamic dispatch.
type IArgumentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CharacterClause() ICharacterClauseContext
	NumericLiteral() INumericLiteralContext
	BooleanLiteral() IBooleanLiteralContext
	GeometryLiteral() IGeometryLiteralContext
	TemporalClause() ITemporalClauseContext
	ArrayClause() IArrayClauseContext
	PropertyName() IPropertyNameContext
	Function() IFunctionContext

	// IsArgumentContext differentiates from other interfaces.
	IsArgumentContext()
}

type ArgumentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArgumentContext() *ArgumentContext {
	var p = new(ArgumentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_argument
	return p
}

func InitEmptyArgumentContext(p *ArgumentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CqlParserRULE_argument
}

func (*ArgumentContext) IsArgumentContext() {}

func NewArgumentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArgumentContext {
	var p = new(ArgumentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CqlParserRULE_argument

	return p
}

func (s *ArgumentContext) GetParser() antlr.Parser { return s.parser }

func (s *ArgumentContext) CharacterClause() ICharacterClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICharacterClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICharacterClauseContext)
}

func (s *ArgumentContext) NumericLiteral() INumericLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INumericLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INumericLiteralContext)
}

func (s *ArgumentContext) BooleanLiteral() IBooleanLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanLiteralContext)
}

func (s *ArgumentContext) GeometryLiteral() IGeometryLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGeometryLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGeometryLiteralContext)
}

func (s *ArgumentContext) TemporalClause() ITemporalClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITemporalClauseContext)
}

func (s *ArgumentContext) ArrayClause() IArrayClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayClauseContext)
}

func (s *ArgumentContext) PropertyName() IPropertyNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyNameContext)
}

func (s *ArgumentContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *ArgumentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArgumentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArgumentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.EnterArgument(s)
	}
}

func (s *ArgumentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CqlParserListener); ok {
		listenerT.ExitArgument(s)
	}
}

func (p *CqlParser) Argument() (localctx IArgumentContext) {
	localctx = NewArgumentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 120, CqlParserRULE_argument)
	p.SetState(533)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 40, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(525)
			p.CharacterClause()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(526)
			p.NumericLiteral()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(527)
			p.BooleanLiteral()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(528)
			p.GeometryLiteral()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(529)
			p.TemporalClause()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(530)
			p.ArrayClause()
		}

	case 7:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(531)
			p.PropertyName()
		}

	case 8:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(532)
			p.Function()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
