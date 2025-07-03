package converter

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
)

// DescriptionAttributeApplier handles the 'description' attribute for variables.
type DescriptionAttributeApplier struct{}

// NewDescriptionAttributeApplier creates a new DescriptionAttributeApplier.
func NewDescriptionAttributeApplier() *DescriptionAttributeApplier {
	return &DescriptionAttributeApplier{}
}

// Apply inspects the 'description' attribute and sets the schema's Description field.
func (a *DescriptionAttributeApplier) Apply(schema *jsonschema.Schema, attrs map[string]*hcl.Attribute) error {
	attr, exists := attrs["description"]
	if !exists || attr == nil {
		return nil // Attribute not present
	}

	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return fmt.Errorf("failed to evaluate 'description' attribute: %w", diags)
	}

	schema.Description = val.AsString()
	return nil
}
