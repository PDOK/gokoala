package cql

import (
	"github.com/PDOK/gokoala/internal/ogc/features/cql/parser"
)

func (cl *CommonListener) ExitFunction(ctx *parser.FunctionContext) {
	function := ctx.Identifier()
	if function != nil {
		functionName := function.GetText()
		cl.errorListener.Error("function " + functionName + " is unsupported")
	}
}

func (cl *CommonListener) ExitArrayPredicate(ctx *parser.ArrayPredicateContext) {
	if ctx.ArrayFunction() != nil {
		cl.errorListener.Error("array operators are not supported")
	}
}

func (cl *CommonListener) ExitArithmeticExpression(ctx *parser.ArithmeticExpressionContext) {
	if ctx.ArithmeticOperatorPlusMinus() != nil {
		cl.errorListener.Error("arithmetic operators are not supported")
	}
}

func (cl *CommonListener) ExitArithmeticTerm(ctx *parser.ArithmeticTermContext) {
	if ctx.ArithmeticOperatorMultDiv() != nil {
		cl.errorListener.Error("arithmetic operators are not supported")
	}
}
