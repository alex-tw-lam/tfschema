package validation

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const alltrueParserName = "alltrue"

func init() {
	RegisterRuleParserWithName(alltrueParserName, parseAllTrueRule, 20)
}

func parseAllTrueRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	call, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok || call.Name != "alltrue" {
		return nil, nil, nil // Not an alltrue() call.
	}

	if len(call.Args) != 1 {
		return nil, nil, nil
	}

	// The argument can be a tuple expression wrapping the ForExpr or the ForExpr directly
	var forExpr *hclsyntax.ForExpr
	if tuple, ok := call.Args[0].(*hclsyntax.TupleConsExpr); ok {
		if len(tuple.Exprs) != 1 {
			return nil, nil, nil
		}
		forExpr, ok = tuple.Exprs[0].(*hclsyntax.ForExpr)
		if !ok {
			return nil, nil, nil
		}
	} else if fe, ok := call.Args[0].(*hclsyntax.ForExpr); ok {
		forExpr = fe
	} else {
		return nil, nil, nil
	}

	innerExpr := forExpr.ValExpr

	// Unwrap any parentheses around the inner expression for simpler parsing.
	if paren, ok := innerExpr.(*hclsyntax.ParenthesesExpr); ok {
		innerExpr = paren.Expression
	}

	collectionPath, err := pathHandler.ExtractPathFromExpression(forExpr.CollExpr, varName)
	if err != nil {
		return nil, nil, fmt.Errorf("could not extract collection path from for expression: %w", err)
	}

	// Try every other registered parser against the loop body. alltrue is itself a
	// parser, so exclude it by name to avoid recursing into ourselves.
	for _, parser := range GetParsersExcept(alltrueParserName) {
		rule, innerPath, err := parser(innerExpr, forExpr.ValVar)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse inner expression in alltrue: %w", err)
		}
		if rule != nil {
			fullPath := append(collectionPath, "*")
			fullPath = append(fullPath, innerPath...)
			return rule, fullPath, nil
		}
	}

	return nil, nil, nil
}
