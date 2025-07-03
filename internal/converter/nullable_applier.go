package converter

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// NullableAttributeApplier handles the 'nullable' attribute for variables.
// When nullable = true, it creates an anyOf schema with null and type options
// to match terraschema's behavior exactly.
type NullableAttributeApplier struct{}

// NewNullableAttributeApplier creates a new NullableAttributeApplier.
func NewNullableAttributeApplier() *NullableAttributeApplier {
	return &NullableAttributeApplier{}
}

// Apply inspects the 'nullable' attribute of a Terraform variable
// and modifies the JSON Schema to match terraschema's behavior.
func (a *NullableAttributeApplier) Apply(schema *jsonschema.Schema, attrs map[string]*hcl.Attribute) error {
	if attr, exists := attrs["nullable"]; exists && attr != nil {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return fmt.Errorf("failed to evaluate 'nullable' attribute: %w", diags)
		}

		if val.Type() == cty.Bool && val.True() {
			// Create anyOf schema to match terraschema behavior
			var originalType string
			if typeStr, ok := schema.Type.(string); ok {
				originalType = typeStr
			} else {
				// Fallback for non-string types
				originalType = "string" // Default assumption
			}

			var originalTitle string

			// Determine the type name for the title
			switch originalType {
			case "string":
				originalTitle = "string"
			case "number":
				originalTitle = "number"
			case "boolean":
				originalTitle = "boolean"
			case "array":
				originalTitle = "array"
			case "object":
				originalTitle = "object"
			default:
				originalTitle = originalType
			}

			// Create the anyOf structure
			schema.AnyOf = []jsonschema.Schema{
				{
					Type:  "null",
					Title: "null",
				},
				{
					Type:  originalType,
					Title: originalTitle,
				},
			}

			// Set the title for the variable selection
			// For now, use a generic title since we don't have variable name in this context
			schema.Title = "Select a type"

			// Clear the original type since we're using anyOf
			schema.Type = nil
		}
	}

	return nil
}
