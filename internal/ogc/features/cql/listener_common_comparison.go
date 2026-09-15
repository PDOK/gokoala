package cql

import (
	"fmt"
	"strings"

	"github.com/PDOK/gokoala/internal/ogc/features/cql/parser"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
)

// ExitBooleanExpression Boolean expression
func (cl *CommonListener) ExitBooleanExpression(ctx *parser.BooleanExpressionContext) {
	count := len(ctx.AllBooleanTerm())
	if count > 1 {
		items := cl.stack.PopMany(count)
		cl.stack.Push("(" + strings.Join(items, " OR ") + ")")
	}
}

// ExitBooleanTerm Boolean expression
func (cl *CommonListener) ExitBooleanTerm(ctx *parser.BooleanTermContext) {
	count := len(ctx.AllBooleanFactor())
	if count > 1 {
		items := cl.stack.PopMany(count)
		cl.stack.Push("(" + strings.Join(items, " AND ") + ")")
	}
}

// ExitBooleanFactor Boolean expression
func (cl *CommonListener) ExitBooleanFactor(ctx *parser.BooleanFactorContext) {
	if ctx.NOT() != nil {
		expr := cl.stack.Pop()
		cl.stack.Push("NOT (" + expr + ")")
	}
}

// ExitIsNullPredicate Comparison expressions (IS NULL, IS NOT NULL)
func (cl *CommonListener) ExitIsNullPredicate(ctx *parser.IsNullPredicateContext) {
	if !cl.cqlConfig.IsAdvancedComparisonOperatorsEnabled() {
		cl.errorListener.Error(errAdvancedComparisonNotEnabled)
		return
	}
	expr := cl.stack.Pop()

	operator := "IS NULL"
	if ctx.NOT() != nil {
		operator = "IS NOT NULL"
	}
	cl.stack.Push(fmt.Sprintf("%s %s", expr, operator))
}

// ExitPatternExpression handles CASEI and ACCENTI.
func (cl *CommonListener) ExitPatternExpression(ctx *parser.PatternExpressionContext) {
	if ctx.CASEI() != nil {
		if !cl.cqlConfig.IsCaseInsensitiveComparisonEnabled() {
			cl.errorListener.Error(errCaseInsensitiveOperatorNotEnabled)
			return
		}
		cl.stack.Push(addCollation(cl.stack.Pop(), common.IgnoreCaseCollation))
	} else if ctx.ACCENTI() != nil {
		if !cl.cqlConfig.IsAccentInsensitiveComparisonEnabled() {
			cl.errorListener.Error(errAccentInsensitiveOperatorNotEnabled)
			return
		}
		cl.stack.Push(addCollation(cl.stack.Pop(), common.IgnoreAccentCollation))
	}
}

// ExitCharacterClause handles CASEI and ACCENTI.
func (cl *CommonListener) ExitCharacterClause(ctx *parser.CharacterClauseContext) {
	if ctx.CASEI() != nil {
		if !cl.cqlConfig.IsCaseInsensitiveComparisonEnabled() {
			cl.errorListener.Error(errCaseInsensitiveOperatorNotEnabled)
			return
		}
		cl.stack.Push(addCollation(cl.stack.Pop(), common.IgnoreCaseCollation))
	} else if ctx.ACCENTI() != nil {
		if !cl.cqlConfig.IsAccentInsensitiveComparisonEnabled() {
			cl.errorListener.Error(errAccentInsensitiveOperatorNotEnabled)
			return
		}
		cl.stack.Push(addCollation(cl.stack.Pop(), common.IgnoreAccentCollation))
	}
}

// ExitIsBetweenPredicate Comparison expressions (BETWEEN, NOT BETWEEN)
func (cl *CommonListener) ExitIsBetweenPredicate(ctx *parser.IsBetweenPredicateContext) {
	if !cl.cqlConfig.IsAdvancedComparisonOperatorsEnabled() {
		cl.errorListener.Error(errAdvancedComparisonNotEnabled)
		return
	}
	high := cl.stack.Pop()
	low := cl.stack.Pop()
	expr := cl.stack.Pop()

	operator := "BETWEEN"
	if ctx.NOT() != nil {
		operator = "NOT " + operator
	}
	cl.stack.Push(fmt.Sprintf("%s %s %s AND %s", expr, operator, low, high))
}
