package validation

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// findRegexCall recursively searches the expression tree for a regex() call.
func findRegexCall(expr hcl.Expression) *hclsyntax.FunctionCallExpr {
	switch e := expr.(type) {
	case *hclsyntax.FunctionCallExpr:
		if e.Name == "regex" {
			return e
		}
		if e.Name == "can" && len(e.Args) == 1 { // look inside can()
			return findRegexCall(e.Args[0])
		}
	case *hclsyntax.BinaryOpExpr:
		if left := findRegexCall(e.LHS); left != nil {
			return left
		}
		return findRegexCall(e.RHS)
	}
	return nil
}

// parseRegexRule turns can(regex("pattern", var.x)) into a pattern rule.
func parseRegexRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	regexCall := findRegexCall(expr)
	if regexCall == nil {
		return nil, nil, nil // not a can(regex(...)) condition
	}
	if len(regexCall.Args) != 2 {
		return nil, nil, fmt.Errorf("expected 'regex' function to have two arguments")
	}
	path, err := extractPathFromExpression(regexCall.Args[1], varName)
	if err != nil {
		return nil, nil, err
	}
	if path == nil {
		path = []string{}
	}
	pattern, diags := regexCall.Args[0].Value(nil)
	if diags.HasErrors() || pattern.Type() != cty.String {
		return nil, nil, fmt.Errorf("regex pattern must be a string literal")
	}

	return &RegexRule{Pattern: pattern.AsString()}, path, nil
}

// RegexRule constrains strings to a regex pattern.
type RegexRule struct {
	Pattern string
}

// Apply stamps the pattern onto a JSON Schema.
func (r *RegexRule) Apply(schema *jsonschema.Schema) {
	schema.Pattern = r.Pattern
}
