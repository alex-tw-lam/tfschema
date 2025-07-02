package validation

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func init() {
	RegisterRuleParserWithPriority(parseLengthRule, 10)
}

// LengthRule represents a length validation rule.
type LengthRule struct {
	Operator *hclsyntax.Operation
	Value    int
}

// getBaseType extracts the primary non-null type from a schema's Type field.
func getBaseType(t interface{}) string {
	if typeStr, ok := t.(string); ok {
		return typeStr
	}
	if typeSlice, ok := t.([]interface{}); ok {
		for _, v := range typeSlice {
			if typeStr, ok := v.(string); ok && typeStr != "null" {
				return typeStr
			}
		}
	}
	return ""
}

// Apply applies the length validation rule to a JSON schema.
func (r *LengthRule) Apply(schema *jsonschema.Schema) error {
	val := r.Value
	baseType := getBaseType(schema.Type)

	applyLength := func(min, max *int) {
		switch baseType {
		case "string":
			schema.MinLength = min
			schema.MaxLength = max
		case "array":
			schema.MinItems = min
			schema.MaxItems = max
		case "object":
			schema.MinProperties = min
			schema.MaxProperties = max
		}
	}

	switch r.Operator {
	case hclsyntax.OpGreaterThan:
		minVal := val + 1
		applyLength(&minVal, nil)
	case hclsyntax.OpGreaterThanOrEqual:
		applyLength(&val, nil)
	case hclsyntax.OpLessThan:
		maxVal := val - 1
		applyLength(nil, &maxVal)
	case hclsyntax.OpLessThanOrEqual:
		applyLength(nil, &val)
	case hclsyntax.OpEqual:
		applyLength(&val, &val)
	default:
		return fmt.Errorf("unsupported operator for length validation: %v", r.Operator)
	}
	return nil
}

// lengthRuleVisitor implements the hclsyntax.Walker interface to traverse
// an expression tree and build a LengthRule.
type lengthRuleVisitor struct {
	varName string
	path    []string
	rules   []*LengthRule
	err     error
}

func (v *lengthRuleVisitor) Enter(node hclsyntax.Node) hcl.Diagnostics {
	// Handle different node types during tree traversal
	if n, ok := node.(*hclsyntax.BinaryOpExpr); ok {
		return v.visitBinaryOpExpr(n)
	}
	return nil
}

func (v *lengthRuleVisitor) Exit(node hclsyntax.Node) hcl.Diagnostics {
	// No special processing needed on exit for length rules
	return nil
}

func (v *lengthRuleVisitor) visitBinaryOpExpr(expr *hclsyntax.BinaryOpExpr) hcl.Diagnostics {
	if expr.Op == hclsyntax.OpLogicalAnd {
		// Continue traversal for compound expressions
		return nil
	}

	call, ok := expr.LHS.(*hclsyntax.FunctionCallExpr)
	if !ok || call.Name != "length" {
		return nil // Not a length() call comparison.
	}

	path, err := pathHandler.ExtractPathFromExpression(call.Args[0], v.varName)
	if err != nil {
		v.err = err
		return nil
	}
	if v.path == nil {
		v.path = path
	}

	if len(call.Args) != 1 {
		v.err = fmt.Errorf("'length' function expects exactly one argument")
		return nil
	}

	lit, ok := expr.RHS.(*hclsyntax.LiteralValueExpr)
	if !ok {
		v.err = fmt.Errorf("rhs of length validation must be a literal value")
		return nil
	}

	if lit.Val.Type() != cty.Number {
		v.err = fmt.Errorf("rhs of length validation must be a number")
		return nil
	}

	val, _ := lit.Val.AsBigFloat().Int64()

	rule := &LengthRule{
		Operator: expr.Op,
		Value:    int(val),
	}
	v.rules = append(v.rules, rule)

	return nil
}

func parseLengthRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	if !containsLengthCall(expr) {
		return nil, nil, nil // No length() calls found
	}

	node, ok := expr.(hclsyntax.Node)
	if !ok {
		return nil, nil, fmt.Errorf("expression is not a syntax node")
	}

	visitor := &lengthRuleVisitor{
		varName: varName,
	}

	diags := hclsyntax.Walk(node, visitor)
	if diags.HasErrors() {
		return nil, nil, diags
	}
	if visitor.err != nil {
		return nil, nil, visitor.err
	}

	if len(visitor.rules) == 0 {
		return nil, nil, nil // No length rule found
	}

	path := visitor.path

	if len(visitor.rules) == 1 {
		return visitor.rules[0], path, nil
	}

	return mergeCompoundLengthRules(visitor.rules[0], visitor.rules[1]), path, nil
}

func containsLengthCall(expr hcl.Expression) bool {
	switch e := expr.(type) {
	case *hclsyntax.BinaryOpExpr:
		return containsLengthCall(e.LHS) || containsLengthCall(e.RHS)
	case *hclsyntax.FunctionCallExpr:
		return e.Name == "length"
	}
	return false
}

func mergeCompoundLengthRules(left, right *LengthRule) Rule {
	compound := &CompoundLengthRule{}

	// Determine which rule sets min vs max values
	for _, rule := range []*LengthRule{left, right} {
		switch rule.Operator {
		case hclsyntax.OpGreaterThan:
			minVal := rule.Value + 1
			compound.MinValue = &minVal
		case hclsyntax.OpGreaterThanOrEqual:
			compound.MinValue = &rule.Value
		case hclsyntax.OpLessThan:
			maxVal := rule.Value - 1
			compound.MaxValue = &maxVal
		case hclsyntax.OpLessThanOrEqual:
			compound.MaxValue = &rule.Value
		case hclsyntax.OpEqual:
			compound.MinValue = &rule.Value
			compound.MaxValue = &rule.Value
		}
	}

	return compound
}

// CompoundLengthRule represents a combined length validation rule with both min and max constraints.
type CompoundLengthRule struct {
	MinValue *int
	MaxValue *int
}

// Apply applies the compound length validation rule to a JSON schema.
func (r *CompoundLengthRule) Apply(schema *jsonschema.Schema) error {
	baseType := getBaseType(schema.Type)
	if r.MinValue != nil {
		switch baseType {
		case "string":
			schema.MinLength = r.MinValue
		case "array":
			schema.MinItems = r.MinValue
		case "object":
			schema.MinProperties = r.MinValue
		}
	}
	if r.MaxValue != nil {
		switch baseType {
		case "string":
			schema.MaxLength = r.MaxValue
		case "array":
			schema.MaxItems = r.MaxValue
		case "object":
			schema.MaxProperties = r.MaxValue
		}
	}
	return nil
}
