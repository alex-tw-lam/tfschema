package validation

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// LengthRule applies length(var.x) OP n constraints: a pair of inclusive
// bounds, stamped onto whichever fields match the target schema's type.
type LengthRule struct {
	MinValue *int
	MaxValue *int
}

func parseLengthRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	collector := &lengthCollector{varName: varName}
	if err := collector.walk(expr); err != nil {
		return nil, nil, err
	}
	if collector.minValue == nil && collector.maxValue == nil {
		return nil, nil, nil // no length comparison found
	}
	return &LengthRule{MinValue: collector.minValue, MaxValue: collector.maxValue}, collector.path, nil
}

// Apply stamps the bounds onto the fields matching the schema's type.
func (r *LengthRule) Apply(schema *jsonschema.Schema) {
	switch schema.Type {
	case "string":
		schema.MinLength, schema.MaxLength = r.MinValue, r.MaxValue
	case "array":
		schema.MinItems, schema.MaxItems = r.MinValue, r.MaxValue
	case "object":
		schema.MinProperties, schema.MaxProperties = r.MinValue, r.MaxValue
	}
}

// lengthCollector gathers length(var.x) OP n comparisons into one pair of bounds.
type lengthCollector struct {
	varName  string
	path     []string
	minValue *int
	maxValue *int
}

func (c *lengthCollector) walk(expr hcl.Expression) error {
	binary, ok := expr.(*hclsyntax.BinaryOpExpr)
	if !ok {
		return nil
	}
	if err := c.foldIn(binary); err != nil {
		return err
	}
	if err := c.walk(binary.LHS); err != nil {
		return err
	}
	return c.walk(binary.RHS)
}

// foldIn folds one `length(...) OP n` comparison in; other nodes are skipped.
func (c *lengthCollector) foldIn(expr *hclsyntax.BinaryOpExpr) error {
	call, ok := expr.LHS.(*hclsyntax.FunctionCallExpr)
	if !ok || call.Name != "length" || len(call.Args) != 1 {
		return nil
	}

	if c.path == nil {
		path, err := extractPathFromExpression(call.Args[0], c.varName)
		if err != nil {
			return err
		}
		c.path = path
	}

	lit, ok := expr.RHS.(*hclsyntax.LiteralValueExpr)
	if !ok || lit.Val.Type() != cty.Number {
		return fmt.Errorf("rhs of length validation must be a number literal")
	}
	val, _ := lit.Val.AsBigFloat().Int64()
	n := int(val)
	switch expr.Op {
	case hclsyntax.OpGreaterThan: // > n means at least n+1
		c.minValue = intPtr(n + 1)
	case hclsyntax.OpGreaterThanOrEqual:
		c.minValue = intPtr(n)
	case hclsyntax.OpLessThan: // < n means at most n-1
		c.maxValue = intPtr(n - 1)
	case hclsyntax.OpLessThanOrEqual:
		c.maxValue = intPtr(n)
	case hclsyntax.OpEqual:
		c.minValue = intPtr(n)
		c.maxValue = intPtr(n)
	default:
		return fmt.Errorf("unsupported operator for length validation: %v", expr.Op)
	}
	return nil
}

func intPtr(i int) *int { return &i }
