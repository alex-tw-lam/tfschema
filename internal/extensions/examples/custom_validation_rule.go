// Package examples demonstrates how to add new features by creating new files
// This file shows how to add a custom validation rule without modifying any existing code
package examples

import (
	"fmt"
	"strings"

	"github.com/atwlam/tfschema/internal/extensions"
	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/atwlam/tfschema/internal/validation"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func init() {
	// Register our custom validation rule when this package is imported
	extensions.RegisterLegacyValidationRule(parseContainsSubstringRule)
}

// ContainsSubstringRule represents a validation rule that checks if a string contains a specific substring
type ContainsSubstringRule struct {
	Substring string
	Pattern   string // JSON Schema pattern equivalent
}

// Apply applies the contains substring validation rule to a JSON schema
func (r *ContainsSubstringRule) Apply(schema *jsonschema.Schema) error {
	// Convert substring requirement to a regex pattern
	// This is a simple example - in practice, you'd want more robust pattern generation
	if r.Pattern != "" {
		schema.Pattern = r.Pattern
	}
	return nil
}

// parseContainsSubstringRule parses expressions like: contains(var.my_string, "required_text")
func parseContainsSubstringRule(expr hcl.Expression, varName string) (validation.Rule, []string, error) {
	call, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok || call.Name != "contains" {
		return nil, nil, nil // Not our rule
	}

	if len(call.Args) != 2 {
		return nil, nil, nil // Wrong number of arguments, might be enum validation
	}

	// Check if the second argument is a string literal (not a list like in enum validation)
	if _, ok := call.Args[0].(*hclsyntax.TupleConsExpr); ok {
		return nil, nil, nil // This is enum validation, not our substring validation
	}

	// Extract the variable path from the first argument
	pathHandler := &pathExpressionHandler{}
	path, err := pathHandler.ExtractPathFromExpression(call.Args[0], varName)
	if err != nil {
		return nil, nil, err
	}

	// Extract the substring from the second argument
	substringExpr := call.Args[1]
	val, diags := substringExpr.Value(nil)
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("failed to evaluate substring: %s", diags.Error())
	}

	if val.Type() != cty.String {
		return nil, nil, nil // Not a string, not our rule
	}

	substring := val.AsString()

	// Create a regex pattern that matches strings containing the substring
	// Escape special regex characters in the substring
	escapedSubstring := strings.ReplaceAll(substring, ".", "\\.")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "+", "\\+")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "*", "\\*")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "?", "\\?")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "^", "\\^")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "$", "\\$")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "[", "\\[")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "]", "\\]")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "{", "\\{")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "}", "\\}")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "(", "\\(")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, ")", "\\)")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "|", "\\|")
	escapedSubstring = strings.ReplaceAll(escapedSubstring, "\\", "\\\\")

	pattern := fmt.Sprintf(".*%s.*", escapedSubstring)

	rule := &ContainsSubstringRule{
		Substring: substring,
		Pattern:   pattern,
	}

	return rule, path, nil
}

// pathExpressionHandler is a minimal implementation for this example
type pathExpressionHandler struct{}

func (p *pathExpressionHandler) ExtractPathFromExpression(expr hcl.Expression, varName string) ([]string, error) {
	// This is a simplified implementation for the example
	// In practice, you'd use the full path expression handler from the validation package

	if scopeTraversal, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
		if len(scopeTraversal.Traversal) >= 2 {
			// Extract path segments after "var.variable_name"
			var path []string
			for i := 2; i < len(scopeTraversal.Traversal); i++ {
				if attr, ok := scopeTraversal.Traversal[i].(hcl.TraverseAttr); ok {
					path = append(path, attr.Name)
				}
			}
			return path, nil
		}
	}

	return []string{}, nil
}
