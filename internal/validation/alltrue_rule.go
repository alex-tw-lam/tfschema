package validation

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// parseAllTrueRule applies the inner comparison of
// alltrue([for x in var.list : ...]) to every element of the collection.
func parseAllTrueRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	call, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok || call.Name != "alltrue" || len(call.Args) != 1 {
		return nil, nil, nil // not an alltrue() call
	}
	forExpr, ok := call.Args[0].(*hclsyntax.ForExpr)
	if !ok {
		return nil, nil, nil
	}
	innerExpr := unwrapParen(forExpr.ValExpr)
	collectionPath, err := extractPathFromExpression(forExpr.CollExpr, varName)
	if err != nil {
		return nil, nil, fmt.Errorf("could not extract collection path from for expression: %w", err)
	}
	// Try each sibling rule parser on the loop body. This list is
	// ruleParsers minus alltrue itself; keep it in sync with ruleParsers.
	for _, parser := range []ParserFunc{parseLengthRule, parseRegexRule, parseEnumRule, parseOrChainRule, parseRangeRule} {
		rule, innerPath, err := parser(innerExpr, forExpr.ValVar)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse inner expression in alltrue: %w", err)
		}
		if rule != nil {
			fullPath := append(collectionPath, "*")
			return rule, append(fullPath, innerPath...), nil
		}
	}
	return nil, nil, nil
}
