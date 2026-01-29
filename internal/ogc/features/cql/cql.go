//go:generate ../../../../hack/generate-cql-parser.sh
package cql

import (
	"github.com/PDOK/gokoala/internal/ogc/features/cql/parser"
	"github.com/antlr4-go/antlr/v4"
)

func ParseToSQL(cql string) string {
	if cql == "" {
		return ""
	}
	input := antlr.NewInputStream(cql)

	// lexer
	cqlLexer := parser.NewCqlLexer(input)

	// parser
	tokens := antlr.NewCommonTokenStream(cqlLexer, antlr.TokenDefaultChannel)
	cqlParser := parser.NewCqlParser(tokens)

	// result
	tree := cqlParser.CqlFilter()
	return tree.ToStringTree(nil, cqlParser)
}
