package validation

import (
	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
)

// Global path expression handler for all validation rules
var pathHandler = NewPathExpressionHandler()

// Rule is the interface that all validation rules must implement.
type Rule interface {
	// Apply applies the validation rule to a JSON schema.
	Apply(schema *jsonschema.Schema) error
}

// ExtractValidationRules extracts the validation rules from a variable's blocks.
func ExtractValidationRules(blocks hcl.Blocks, varName string) ([]ScopedRule, error) {
	var scopedRules []ScopedRule

	for _, block := range blocks {
		if block.Type != "validation" {
			continue
		}

		content, diags := block.Body.Content(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{Name: "condition"},
				{Name: "error_message"},
			},
		})
		if diags.HasErrors() {
			continue
		}

		condition, ok := content.Attributes["condition"]
		if !ok {
			continue
		}

		// Try each registered parser in priority order
		for _, parser := range GetParsers() {
			rule, path, err := parser(condition.Expr, varName)
			if err != nil {
				return nil, err
			}
			if rule != nil {
				scopedRules = append(scopedRules, ScopedRule{
					Rule: rule,
					Path: path,
				})
				break // Move to next validation block
			}
		}
	}

	return scopedRules, nil
}
