package validation

// Shared helpers for reading variable references and literals out of HCL
// expressions; used by every rule parser.

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// isBareVariableComparison reports whether the expression compares the
// variable itself (e.g. var.port >= 1) rather than a function of it
// (e.g. length(var.name) > 0, which belongs to the length rule).
func isBareVariableComparison(expr hcl.Expression, varName string) bool {
	binary, ok := unwrapParen(expr).(*hclsyntax.BinaryOpExpr)
	if !ok {
		return false
	}

	switch binary.Op {
	case hclsyntax.OpGreaterThan, hclsyntax.OpGreaterThanOrEqual,
		hclsyntax.OpLessThan, hclsyntax.OpLessThanOrEqual,
		hclsyntax.OpEqual:
		// A comparison: one side must be the variable itself.
		return isVariableReferenceForVar(binary.LHS, varName) ||
			isVariableReferenceForVar(binary.RHS, varName)
	case hclsyntax.OpLogicalAnd, hclsyntax.OpLogicalOr:
		// A compound expression: check each side recursively.
		return isBareVariableComparison(binary.LHS, varName) ||
			isBareVariableComparison(binary.RHS, varName)
	}
	return false
}

func isVariableReferenceForVar(expr hcl.Expression, varName string) bool {
	if traversal, ok := unwrapParen(expr).(*hclsyntax.ScopeTraversalExpr); ok {
		// Accept both var.* references and direct loop variable references
		rootName := traversal.Traversal.RootName()
		return rootName == "var" || rootName == "self" || rootName == varName
	}
	return false
}

// extractLiteralValue evaluates a literal expression to float64/string/bool.
func extractLiteralValue(expr hcl.Expression) (interface{}, error) {
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to evaluate expression: %s", diags)
	}
	switch val.Type() {
	case cty.Number:
		f, _ := val.AsBigFloat().Float64()
		return f, nil
	case cty.String:
		return val.AsString(), nil
	case cty.Bool:
		return val.True(), nil
	}
	return nil, fmt.Errorf("unsupported literal type: %v", val.Type())
}

// extractNumber evaluates a number literal to a float64.
func extractNumber(expr hcl.Expression) (float64, error) {
	value, err := extractLiteralValue(expr)
	if err != nil {
		return 0, err
	}
	num, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("expected a number, got %T", value)
	}
	return num, nil
}

func containsFunctionCall(expr hcl.Expression) bool {
	if binary, ok := expr.(*hclsyntax.BinaryOpExpr); ok {
		return containsFunctionCall(binary.LHS) || containsFunctionCall(binary.RHS)
	}
	_, ok := expr.(*hclsyntax.FunctionCallExpr)
	return ok
}

// unwrapParen recursively removes any enclosing ParenthesesExpr.
func unwrapParen(expr hcl.Expression) hcl.Expression {
	for {
		if p, ok := expr.(*hclsyntax.ParenthesesExpr); ok {
			expr = p.Expression
			continue
		}
		return expr
	}
}
