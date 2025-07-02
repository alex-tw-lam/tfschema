package validation

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func init() {
	RegisterRuleParserWithPriority(parseRegexRule, 10)
}

// findRegexCall recursively searches for a regex function call within complex expressions
func findRegexCall(expr hcl.Expression) *hclsyntax.FunctionCallExpr {
	switch e := expr.(type) {
	case *hclsyntax.FunctionCallExpr:
		if e.Name == "regex" {
			return e
		}
		// Check inside can() functions
		if e.Name == "can" && len(e.Args) == 1 {
			return findRegexCall(e.Args[0])
		}
	case *hclsyntax.BinaryOpExpr:
		// Search both sides of binary operations (like ||, &&)
		if left := findRegexCall(e.LHS); left != nil {
			return left
		}
		if right := findRegexCall(e.RHS); right != nil {
			return right
		}
	}
	return nil
}

// RegexRule represents a regex validation rule.
type RegexRule struct {
	Pattern string
}

// Apply applies the regex validation rule to a JSON schema.
func (r *RegexRule) Apply(schema *jsonschema.Schema) error {
	schema.Pattern = r.Pattern
	return nil
}

// regexRuleVisitor implements the hclsyntax.Walker interface to traverse
// an expression tree and build a RegexRule.
type regexRuleVisitor struct {
	varName string
	rule    *RegexRule
	path    []string
	err     error
}

func (v *regexRuleVisitor) Enter(node hclsyntax.Node) hcl.Diagnostics {
	// Handle different node types during tree traversal
	if call, ok := node.(*hclsyntax.FunctionCallExpr); ok {
		return v.visitFunctionCallExpr(call)
	}
	return nil
}

func (v *regexRuleVisitor) Exit(node hclsyntax.Node) hcl.Diagnostics {
	// No special processing needed on exit for regex rules
	return nil
}

func (v *regexRuleVisitor) visitFunctionCallExpr(expr *hclsyntax.FunctionCallExpr) hcl.Diagnostics {
	if expr.Name != "can" {
		return nil
	}

	if len(expr.Args) != 1 {
		v.err = fmt.Errorf("'can' function for regex should have one argument")
		return nil
	}

	regexCall, ok := expr.Args[0].(*hclsyntax.FunctionCallExpr)
	if !ok || regexCall.Name != "regex" {
		return nil // Not a regex call inside can()
	}

	if len(regexCall.Args) != 2 {
		v.err = fmt.Errorf("expected 'regex' function to have two arguments")
		return nil
	}

	path, err := pathHandler.ExtractPathFromExpression(regexCall.Args[1], v.varName)
	if err != nil {
		v.err = err
		return nil
	}
	if path == nil {
		v.path = []string{}
	} else {
		v.path = path
	}

	var pattern string
	if templateExpr, ok := regexCall.Args[0].(*hclsyntax.TemplateExpr); ok {
		if len(templateExpr.Parts) == 1 {
			if lit, ok := templateExpr.Parts[0].(*hclsyntax.LiteralValueExpr); ok && lit.Val.Type() == cty.String {
				pattern = lit.Val.AsString()
			}
		}
	} else if lit, ok := regexCall.Args[0].(*hclsyntax.LiteralValueExpr); ok && lit.Val.Type() == cty.String {
		pattern = lit.Val.AsString()
	}

	if pattern == "" {
		v.err = fmt.Errorf("regex pattern must be a string literal")
		return nil
	}

	v.rule = &RegexRule{
		Pattern: pattern,
	}
	return nil
}

func parseRegexRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	// Use the helper to find regex calls in complex expressions
	regexCall := findRegexCall(expr)
	if regexCall == nil {
		return nil, nil, nil
	}

	if len(regexCall.Args) != 2 {
		return nil, nil, fmt.Errorf("expected 'regex' function to have two arguments")
	}

	path, err := pathHandler.ExtractPathFromExpression(regexCall.Args[1], varName)
	if err != nil {
		return nil, nil, err
	}
	if path == nil {
		path = []string{}
	}

	var pattern string
	if templateExpr, ok := regexCall.Args[0].(*hclsyntax.TemplateExpr); ok {
		if len(templateExpr.Parts) == 1 {
			if lit, ok := templateExpr.Parts[0].(*hclsyntax.LiteralValueExpr); ok && lit.Val.Type() == cty.String {
				pattern = lit.Val.AsString()
			}
		}
	} else if lit, ok := regexCall.Args[0].(*hclsyntax.LiteralValueExpr); ok && lit.Val.Type() == cty.String {
		pattern = lit.Val.AsString()
	}

	if pattern == "" {
		return nil, nil, fmt.Errorf("regex pattern must be a string literal")
	}

	rule := &RegexRule{
		Pattern: pattern,
	}

	return rule, path, nil
}
