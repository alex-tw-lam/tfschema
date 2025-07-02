package validation

// Range rules: numeric comparisons become number bounds (minimum/maximum and
// their exclusive variants); equality on a number becomes exact bounds.

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// RangeRule carries numeric bounds; whichever fields are set are stamped
// onto the target schema.
type RangeRule struct {
	Minimum          *float64
	Maximum          *float64
	ExclusiveMinimum *float64
	ExclusiveMaximum *float64
}

// Apply stamps the bounds onto a JSON Schema.
func (r *RangeRule) Apply(schema *jsonschema.Schema) {
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
}

func parseRangeRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	binary, ok := unwrapParen(expr).(*hclsyntax.BinaryOpExpr)
	// OR chains are handled by parseOrChainRule; other comparisons become ranges.
	if !ok || binary.Op == hclsyntax.OpLogicalOr || !isBareVariableComparison(binary, varName) {
		return nil, nil, nil // not a range operation
	}

	path, err := extractPathFromExpression(expr, varName)
	if err != nil {
		return nil, nil, err
	}

	if binary.Op == hclsyntax.OpLogicalAnd {
		// e.g. var.value >= 1 && var.value <= 10: merge both bounds.
		left, err := parseComparison(binary.LHS, varName)
		if err != nil {
			return nil, nil, err
		}
		right, err := parseComparison(binary.RHS, varName)
		if err != nil {
			return nil, nil, err
		}
		leftBounds, leftIsRange := left.(*RangeRule)
		rightBounds, rightIsRange := right.(*RangeRule)
		if !leftIsRange || !rightIsRange {
			return nil, nil, fmt.Errorf("both sides of '&&' must be numeric comparisons")
		}
		return mergeRangeRules(leftBounds, rightBounds), path, nil
	}

	rule, err := parseComparison(binary, varName)
	if err != nil {
		return nil, nil, err
	}
	return rule, path, nil
}

// parseComparison turns one `var.x OP value` comparison into a rule.
func parseComparison(expr hcl.Expression, varName string) (Rule, error) {
	comparison, ok := unwrapParen(expr).(*hclsyntax.BinaryOpExpr)
	if !ok {
		return nil, fmt.Errorf("not a comparison expression")
	}

	op, valueExpr, err := comparisonParts(comparison, varName)
	if err != nil {
		return nil, err
	}
	if op == hclsyntax.OpEqual {
		return parseEquality(valueExpr)
	}
	return parseRangeBounds(op, valueExpr)
}

// comparisonParts finds the variable side of a comparison and returns the
// operator (reversed for value-first forms like 3 < var.x) and the value side.
func comparisonParts(comparison *hclsyntax.BinaryOpExpr, varName string) (*hclsyntax.Operation, hcl.Expression, error) {
	if isVariableReferenceForVar(comparison.RHS, varName) {
		return reverseOp(comparison.Op), comparison.LHS, nil
	}
	if isVariableReferenceForVar(comparison.LHS, varName) {
		return comparison.Op, comparison.RHS, nil
	}
	return nil, nil, fmt.Errorf("no variable reference found in comparison")
}

// parseRangeBounds turns one numeric comparison into bounds.
func parseRangeBounds(op *hclsyntax.Operation, valueExpr hcl.Expression) (*RangeRule, error) {
	num, err := extractNumber(valueExpr)
	if err != nil {
		return nil, err
	}
	switch op {
	case hclsyntax.OpGreaterThan:
		return &RangeRule{ExclusiveMinimum: &num}, nil
	case hclsyntax.OpGreaterThanOrEqual:
		return &RangeRule{Minimum: &num}, nil
	case hclsyntax.OpLessThan:
		return &RangeRule{ExclusiveMaximum: &num}, nil
	case hclsyntax.OpLessThanOrEqual:
		return &RangeRule{Maximum: &num}, nil
	}
	return nil, fmt.Errorf("unsupported operator: %v", op)
}

// mergeRangeRules combines the bounds of two comparisons (e.g. >= 1 && <= 10):
// each field takes whichever side sets it, the right side winning conflicts.
func mergeRangeRules(left, right *RangeRule) *RangeRule {
	merged := *right
	if left.Minimum != nil && merged.Minimum == nil {
		merged.Minimum = left.Minimum
	}
	if left.Maximum != nil && merged.Maximum == nil {
		merged.Maximum = left.Maximum
	}
	if left.ExclusiveMinimum != nil && merged.ExclusiveMinimum == nil {
		merged.ExclusiveMinimum = left.ExclusiveMinimum
	}
	if left.ExclusiveMaximum != nil && merged.ExclusiveMaximum == nil {
		merged.ExclusiveMaximum = left.ExclusiveMaximum
	}
	return &merged
}

// reverseOp flips an operator written value-first (3 < var.x).
func reverseOp(op *hclsyntax.Operation) *hclsyntax.Operation {
	switch op {
	case hclsyntax.OpGreaterThan:
		return hclsyntax.OpLessThan
	case hclsyntax.OpGreaterThanOrEqual:
		return hclsyntax.OpLessThanOrEqual
	case hclsyntax.OpLessThan:
		return hclsyntax.OpGreaterThan
	case hclsyntax.OpLessThanOrEqual:
		return hclsyntax.OpGreaterThanOrEqual
	}
	return op
}
