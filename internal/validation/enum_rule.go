package validation

// Enum rules: equality comparisons, contains([...], var.x) and
// `var.x == a || var.x == b` chains become JSON Schema "enum" constraints.

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// EnumRule restricts a value to a fixed set of allowed values.
type EnumRule struct {
	Values []interface{}
}

// Apply stamps the allowed values onto a JSON Schema.
func (r *EnumRule) Apply(schema *jsonschema.Schema) {
	schema.Enum = r.Values
}

// parseEnumRule turns contains(["a", "b"], var.x) into an enum rule.
func parseEnumRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	call, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok || call.Name != "contains" {
		return nil, nil, nil // not a 'contains' function call
	}
	if len(call.Args) != 2 {
		return nil, nil, fmt.Errorf("'contains' must have two arguments")
	}

	path, err := extractPathFromExpression(call.Args[1], varName)
	if err != nil {
		return nil, nil, err
	}

	listExpr, ok := call.Args[0].(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, nil, fmt.Errorf("first argument to 'contains' must be a literal list")
	}

	var values []interface{}
	for _, itemExpr := range listExpr.Exprs {
		value, err := extractLiteralValue(itemExpr)
		if err != nil {
			return nil, nil, err
		}
		values = append(values, value)
	}
	return &EnumRule{Values: values}, path, nil
}

// parseOrChainRule turns `var.x == a || var.x == b` chains into an enum rule.
func parseOrChainRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	binary, ok := unwrapParen(expr).(*hclsyntax.BinaryOpExpr)
	if !ok || binary.Op != hclsyntax.OpLogicalOr {
		return nil, nil, nil // not an OR chain
	}
	// An OR containing function calls is not an enum; let another parser try.
	if containsFunctionCall(binary) {
		return nil, nil, nil
	}

	values, err := enumFromOrChain(binary, varName)
	if err != nil {
		return nil, nil, err
	}
	path, err := extractPathFromExpression(expr, varName)
	if err != nil {
		return nil, nil, err
	}
	return &EnumRule{Values: values}, path, nil
}

// enumFromOrChain collects the literal values of `var.x == v || ...` leaves.
func enumFromOrChain(expr hcl.Expression, varName string) ([]interface{}, error) {
	binary, ok := unwrapParen(expr).(*hclsyntax.BinaryOpExpr)
	if !ok {
		return nil, fmt.Errorf("expected equality or OR expression")
	}

	if binary.Op == hclsyntax.OpLogicalOr {
		left, err := enumFromOrChain(binary.LHS, varName)
		if err != nil {
			return nil, err
		}
		right, err := enumFromOrChain(binary.RHS, varName)
		if err != nil {
			return nil, err
		}
		return append(left, right...), nil
	}

	if binary.Op != hclsyntax.OpEqual {
		return nil, fmt.Errorf("expected equality comparison in enum chain")
	}

	valueExpr := binary.RHS
	if isVariableReferenceForVar(binary.RHS, varName) {
		valueExpr = binary.LHS
	} else if !isVariableReferenceForVar(binary.LHS, varName) {
		return nil, fmt.Errorf("no variable reference found in equality comparison")
	}

	value, err := extractLiteralValue(valueExpr)
	if err != nil {
		return nil, err
	}
	return []interface{}{value}, nil
}

// parseEquality turns `var.x == value` into exact numeric bounds or an enum.
func parseEquality(valueExpr hcl.Expression) (Rule, error) {
	value, err := extractLiteralValue(valueExpr)
	if err != nil {
		return nil, err
	}
	if num, ok := value.(float64); ok {
		return &RangeRule{Minimum: &num, Maximum: &num}, nil
	}
	return &EnumRule{Values: []interface{}{value}}, nil
}
