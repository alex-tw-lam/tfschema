package validation

import (
	"fmt"
	"log"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func init() {
	RegisterRuleParserWithPriority(parseRangeRule, 0)
}

// RangeRule represents numeric range constraints
type RangeRule struct {
	Minimum          *float64      `json:"minimum,omitempty"`
	Maximum          *float64      `json:"maximum,omitempty"`
	ExclusiveMinimum *float64      `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64      `json:"exclusiveMaximum,omitempty"`
	Enum             []interface{} `json:"enum,omitempty"`
}

// Apply applies the range validation rule to a JSON schema.
func (r *RangeRule) Apply(schema *jsonschema.Schema) error {
	if r.Minimum != nil {
		schema.Minimum = r.Minimum
	}
	if r.Maximum != nil {
		schema.Maximum = r.Maximum
	}
	if r.ExclusiveMinimum != nil {
		schema.ExclusiveMinimum = r.ExclusiveMinimum
	}
	if r.ExclusiveMaximum != nil {
		schema.ExclusiveMaximum = r.ExclusiveMaximum
	}
	if len(r.Enum) > 0 {
		schema.Enum = r.Enum
	}
	return nil
}

func parseRangeRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	expr = unwrapParen(expr)
	log.Printf("[range] called var=%s, expr=%T", varName, expr)
	binaryExpr, ok := expr.(*hclsyntax.BinaryOpExpr)
	if !ok {
		return nil, nil, nil // Not a binary operation.
	}
	log.Printf("[range] op=%v", binaryExpr.Op)

	// Check if this is a range comparison
	if !isRangeOperationForVar(binaryExpr, varName) {
		log.Printf("[range] isRangeOperation=false for var=%s", varName)
		return nil, nil, nil // Not a range operation.
	}

	// Extract the path from any variable reference in the expression
	path, err := extractPathFromRangeExpr(expr, varName)
	if err != nil {
		return nil, nil, err
	}

	rule := &RangeRule{}
	op := binaryExpr.Op

	switch op {
	case hclsyntax.OpLogicalAnd:
		// Handle compound expressions like: var.value >= 1 && var.value <= 10
		leftRule, err := parseSingleComparison(binaryExpr.LHS, varName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse left side: %w", err)
		}
		rightRule, err := parseSingleComparison(binaryExpr.RHS, varName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse right side: %w", err)
		}

		// Merge the rules
		rule = mergeRangeRules(leftRule, rightRule)

	case hclsyntax.OpLogicalOr:
		// If either side of the OR contains a function call, this isn't a simple enum.
		// Abort and let another parser (e.g., regex) handle it.
		if containsFunctionCall(binaryExpr.LHS) || containsFunctionCall(binaryExpr.RHS) {
			return nil, nil, nil
		}

		// Handle OR expressions like: var.value == 1 || var.value == 2
		enumValues, err := parseOrExpression(binaryExpr, varName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse OR expression: %w", err)
		}
		rule.Enum = enumValues

	default:
		// Handle single comparison
		singleRule, err := parseSingleComparison(expr, varName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse single comparison: %w", err)
		}
		rule = singleRule
	}

	log.Printf("[range] result rule=%#v path=%v", rule, path)
	return rule, path, nil
}

func isRangeOperation(expr *hclsyntax.BinaryOpExpr) bool {
	switch expr.Op {
	case hclsyntax.OpLogicalAnd, hclsyntax.OpLogicalOr:
		// For logical operations, check if it contains direct variable comparisons (not function calls)
		return hasDirectVariableComparison(expr)
	case hclsyntax.OpGreaterThan, hclsyntax.OpGreaterThanOrEqual,
		hclsyntax.OpLessThan, hclsyntax.OpLessThanOrEqual,
		hclsyntax.OpEqual:
		// For comparison operations, check if it has direct variable references
		return hasDirectVariableReference(expr)
	}
	return false
}

func isRangeOperationForVar(expr *hclsyntax.BinaryOpExpr, varName string) bool {
	switch expr.Op {
	case hclsyntax.OpLogicalAnd, hclsyntax.OpLogicalOr:
		// For logical operations, check if it contains direct variable comparisons (not function calls)
		return hasDirectVariableComparisonForVar(expr, varName)
	case hclsyntax.OpGreaterThan, hclsyntax.OpGreaterThanOrEqual,
		hclsyntax.OpLessThan, hclsyntax.OpLessThanOrEqual,
		hclsyntax.OpEqual:
		// For comparison operations, check if it has direct variable references
		return hasDirectVariableReferenceForVar(expr, varName)
	}
	return false
}

func hasDirectVariableComparison(expr *hclsyntax.BinaryOpExpr) bool {
	// Check if this is a compound expression with direct variable comparisons
	// (not function calls like length())
	return hasDirectVariableComparisonRecursive(expr.LHS) || hasDirectVariableComparisonRecursive(expr.RHS)
}

func hasDirectVariableComparisonForVar(expr *hclsyntax.BinaryOpExpr, varName string) bool {
	// Check if this is a compound expression with direct variable comparisons
	// (not function calls like length())
	return hasDirectVariableComparisonRecursiveForVar(expr.LHS, varName) || hasDirectVariableComparisonRecursiveForVar(expr.RHS, varName)
}

func hasDirectVariableComparisonRecursive(expr hcl.Expression) bool {
	if paren, ok := expr.(*hclsyntax.ParenthesesExpr); ok {
		return hasDirectVariableComparisonRecursive(paren.Expression)
	}
	if binaryExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok {
		switch binaryExpr.Op {
		case hclsyntax.OpGreaterThan, hclsyntax.OpGreaterThanOrEqual,
			hclsyntax.OpLessThan, hclsyntax.OpLessThanOrEqual,
			hclsyntax.OpEqual:
			// Check if this is a direct variable comparison (not a function call)
			return (isVariableReference(binaryExpr.LHS) && !isFunctionCall(binaryExpr.LHS)) ||
				(isVariableReference(binaryExpr.RHS) && !isFunctionCall(binaryExpr.RHS))
		case hclsyntax.OpLogicalAnd, hclsyntax.OpLogicalOr:
			// Recursively check nested logical operations
			return hasDirectVariableComparisonRecursive(binaryExpr.LHS) ||
				hasDirectVariableComparisonRecursive(binaryExpr.RHS)
		}
	}
	return false
}

func hasDirectVariableComparisonRecursiveForVar(expr hcl.Expression, varName string) bool {
	if paren, ok := expr.(*hclsyntax.ParenthesesExpr); ok {
		return hasDirectVariableComparisonRecursiveForVar(paren.Expression, varName)
	}
	if binaryExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok {
		switch binaryExpr.Op {
		case hclsyntax.OpGreaterThan, hclsyntax.OpGreaterThanOrEqual,
			hclsyntax.OpLessThan, hclsyntax.OpLessThanOrEqual,
			hclsyntax.OpEqual:
			// Check if this is a direct variable comparison (not a function call)
			return (isVariableReferenceForVar(binaryExpr.LHS, varName) && !isFunctionCall(binaryExpr.LHS)) ||
				(isVariableReferenceForVar(binaryExpr.RHS, varName) && !isFunctionCall(binaryExpr.RHS))
		case hclsyntax.OpLogicalAnd, hclsyntax.OpLogicalOr:
			// Recursively check nested logical operations
			return hasDirectVariableComparisonRecursiveForVar(binaryExpr.LHS, varName) ||
				hasDirectVariableComparisonRecursiveForVar(binaryExpr.RHS, varName)
		}
	}
	return false
}

func hasDirectVariableReference(expr *hclsyntax.BinaryOpExpr) bool {
	// Check if either side has a direct variable reference (not wrapped in function calls)
	return (isVariableReference(expr.LHS) && !isFunctionCall(expr.LHS)) ||
		(isVariableReference(expr.RHS) && !isFunctionCall(expr.RHS))
}

func hasDirectVariableReferenceForVar(expr *hclsyntax.BinaryOpExpr, varName string) bool {
	// Check if either side has a direct variable reference (not wrapped in function calls)
	return (isVariableReferenceForVar(expr.LHS, varName) && !isFunctionCall(expr.LHS)) ||
		(isVariableReferenceForVar(expr.RHS, varName) && !isFunctionCall(expr.RHS))
}

func isFunctionCall(expr hcl.Expression) bool {
	_, ok := expr.(*hclsyntax.FunctionCallExpr)
	return ok
}

func isVariableReference(expr hcl.Expression) bool {
	// First unwrap any parentheses
	if paren, ok := expr.(*hclsyntax.ParenthesesExpr); ok {
		return isVariableReference(paren.Expression)
	}

	if traversal, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
		rootName := traversal.Traversal.RootName()
		return rootName == "var" || rootName == "self"
	}
	return false
}

func isVariableReferenceForVar(expr hcl.Expression, varName string) bool {
	// First unwrap any parentheses
	if paren, ok := expr.(*hclsyntax.ParenthesesExpr); ok {
		return isVariableReferenceForVar(paren.Expression, varName)
	}

	if traversal, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
		rootName := traversal.Traversal.RootName()
		// Accept both var.* references and direct loop variable references
		return rootName == "var" || rootName == "self" || rootName == varName
	}
	return false
}

func extractPathFromRangeExpr(expr hcl.Expression, varName string) ([]string, error) {
	// Find the first variable reference in the expression
	switch e := expr.(type) {
	case *hclsyntax.BinaryOpExpr:
		if isVariableReferenceForVar(e.LHS, varName) {
			return pathHandler.ExtractPathFromExpression(e.LHS, varName)
		}
		if isVariableReferenceForVar(e.RHS, varName) {
			return pathHandler.ExtractPathFromExpression(e.RHS, varName)
		}
		// Try left side first, then right side for compound expressions
		if path, err := extractPathFromRangeExpr(e.LHS, varName); err == nil && path != nil {
			return path, nil
		}
		return extractPathFromRangeExpr(e.RHS, varName)
	case *hclsyntax.ScopeTraversalExpr:
		return pathHandler.ExtractPathFromExpression(e, varName)
	}

	return nil, nil // No variable reference found
}

func parseSingleComparison(expr hcl.Expression, varName string) (*RangeRule, error) {
	expr = unwrapParen(expr)
	comparison, ok := expr.(*hclsyntax.BinaryOpExpr)
	if !ok {
		return nil, fmt.Errorf("not a comparison expression")
	}

	// Determine which side is the variable and which is the value
	var valueExpr hcl.Expression
	var isReversed bool

	if isVariableReferenceForVar(comparison.LHS, varName) {
		valueExpr = comparison.RHS
		isReversed = false
	} else if isVariableReferenceForVar(comparison.RHS, varName) {
		valueExpr = comparison.LHS
		isReversed = true
	} else {
		return nil, fmt.Errorf("no variable reference found in comparison")
	}

	// Extract value (try numeric first, then any literal for enum support)
	var value interface{}

	if numVal, numErr := extractNumericValue(valueExpr); numErr == nil {
		value = numVal
	} else if litVal, litErr := extractLiteralValue(valueExpr); litErr == nil {
		value = litVal
	} else {
		return nil, fmt.Errorf("failed to extract value from expression")
	}

	rule := &RangeRule{}
	op := comparison.Op

	// Reverse the operator if the variable is on the right side
	if isReversed {
		switch op {
		case hclsyntax.OpGreaterThan:
			op = hclsyntax.OpLessThan
		case hclsyntax.OpGreaterThanOrEqual:
			op = hclsyntax.OpLessThanOrEqual
		case hclsyntax.OpLessThan:
			op = hclsyntax.OpGreaterThan
		case hclsyntax.OpLessThanOrEqual:
			op = hclsyntax.OpGreaterThanOrEqual
		}
	}

	switch op {
	case hclsyntax.OpGreaterThan:
		if numVal, ok := value.(float64); ok {
			rule.ExclusiveMinimum = &numVal
		} else {
			return nil, fmt.Errorf("comparison operators require numeric values")
		}
	case hclsyntax.OpGreaterThanOrEqual:
		if numVal, ok := value.(float64); ok {
			rule.Minimum = &numVal
		} else {
			return nil, fmt.Errorf("comparison operators require numeric values")
		}
	case hclsyntax.OpLessThan:
		if numVal, ok := value.(float64); ok {
			rule.ExclusiveMaximum = &numVal
		} else {
			return nil, fmt.Errorf("comparison operators require numeric values")
		}
	case hclsyntax.OpLessThanOrEqual:
		if numVal, ok := value.(float64); ok {
			rule.Maximum = &numVal
		} else {
			return nil, fmt.Errorf("comparison operators require numeric values")
		}
	case hclsyntax.OpEqual:
		// For numeric types, treat equality as a range with the same min and max
		if numVal, ok := value.(float64); ok {
			rule.Minimum = &numVal
			rule.Maximum = &numVal
		} else {
			// For non-numeric types, use enum
			rule.Enum = []interface{}{value}
		}
	default:
		return nil, fmt.Errorf("unsupported operator: %v", op)
	}

	return rule, nil
}

func parseOrExpression(expr *hclsyntax.BinaryOpExpr, varName string) ([]interface{}, error) {
	var values []interface{}

	// Recursively collect all values from OR expressions
	leftValues, err := collectOrValues(expr.LHS, varName)
	if err != nil {
		return nil, err
	}
	values = append(values, leftValues...)

	rightValues, err := collectOrValues(expr.RHS, varName)
	if err != nil {
		return nil, err
	}
	values = append(values, rightValues...)

	return values, nil
}

func collectOrValues(expr hcl.Expression, varName string) ([]interface{}, error) {
	if binaryExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok && binaryExpr.Op == hclsyntax.OpLogicalOr {
		// Recursively handle nested OR expressions
		return parseOrExpression(binaryExpr, varName)
	}

	if binaryExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok && binaryExpr.Op == hclsyntax.OpEqual {
		// Handle equality comparison
		var valueExpr hcl.Expression
		if isVariableReferenceForVar(binaryExpr.LHS, varName) {
			valueExpr = binaryExpr.RHS
		} else if isVariableReferenceForVar(binaryExpr.RHS, varName) {
			valueExpr = binaryExpr.LHS
		} else {
			return nil, fmt.Errorf("no variable reference found in equality comparison")
		}

		value, err := extractLiteralValue(valueExpr)
		if err != nil {
			return nil, err
		}
		return []interface{}{value}, nil
	}

	return nil, fmt.Errorf("expected equality or OR expression")
}

func mergeRangeRules(left, right *RangeRule) *RangeRule {
	result := &RangeRule{}

	if left.Minimum != nil {
		result.Minimum = left.Minimum
	}
	if right.Minimum != nil {
		result.Minimum = right.Minimum
	}

	if left.Maximum != nil {
		result.Maximum = left.Maximum
	}
	if right.Maximum != nil {
		result.Maximum = right.Maximum
	}

	if left.ExclusiveMinimum != nil {
		result.ExclusiveMinimum = left.ExclusiveMinimum
	}
	if right.ExclusiveMinimum != nil {
		result.ExclusiveMinimum = right.ExclusiveMinimum
	}

	if left.ExclusiveMaximum != nil {
		result.ExclusiveMaximum = left.ExclusiveMaximum
	}
	if right.ExclusiveMaximum != nil {
		result.ExclusiveMaximum = right.ExclusiveMaximum
	}

	return result
}

func extractNumericValue(expr hcl.Expression) (float64, error) {
	if literal, ok := expr.(*hclsyntax.LiteralValueExpr); ok {
		val := literal.Val
		if val.Type() == cty.Number {
			f, _ := val.AsBigFloat().Float64()
			return f, nil
		}
	}
	return 0, fmt.Errorf("expected numeric literal")
}

func extractLiteralValue(expr hcl.Expression) (interface{}, error) {
	// Try direct literal value first
	if literal, ok := expr.(*hclsyntax.LiteralValueExpr); ok {
		val := literal.Val
		if val.Type() == cty.Number {
			f, _ := val.AsBigFloat().Float64()
			return f, nil
		} else if val.Type() == cty.String {
			return val.AsString(), nil
		} else if val.Type() == cty.Bool {
			return val.True(), nil
		}
	}

	// Try evaluating the expression
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to evaluate expression: %s", diags.Error())
	}

	if val.Type() == cty.Number {
		f, _ := val.AsBigFloat().Float64()
		return f, nil
	} else if val.Type() == cty.String {
		return val.AsString(), nil
	} else if val.Type() == cty.Bool {
		return val.True(), nil
	}

	return nil, fmt.Errorf("unsupported literal type: %v", val.Type())
}

func containsFunctionCall(expr hcl.Expression) bool {
	if _, ok := expr.(*hclsyntax.FunctionCallExpr); ok {
		return true
	}
	if binary, ok := expr.(*hclsyntax.BinaryOpExpr); ok {
		return containsFunctionCall(binary.LHS) || containsFunctionCall(binary.RHS)
	}
	return false
}

// unwrapParen recursively removes any enclosing ParenthesesExpr
func unwrapParen(expr hcl.Expression) hcl.Expression {
	for {
		if p, ok := expr.(*hclsyntax.ParenthesesExpr); ok {
			expr = p.Expression
			continue
		}
		return expr
	}
}
