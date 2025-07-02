package converter

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// SensitiveAttributeApplier handles the 'sensitive' attribute.
type SensitiveAttributeApplier struct{}

// NewSensitiveAttributeApplier creates a new SensitiveAttributeApplier.
func NewSensitiveAttributeApplier() *SensitiveAttributeApplier {
	return &SensitiveAttributeApplier{}
}

// Apply inspects the 'sensitive' attribute of a Terraform variable
// and modifies the JSON Schema accordingly.
func (a *SensitiveAttributeApplier) Apply(schema *jsonschema.Schema, attrs map[string]*hcl.Attribute) error {
	attr, exists := attrs["sensitive"]
	if !exists || attr == nil {
		return nil // Attribute not present
	}

	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return fmt.Errorf("failed to evaluate 'sensitive' attribute: %w", diags)
	}

	if val.Type() == cty.Bool && val.True() {
		sensitive := true
		schema.Sensitive = &sensitive
	}

	return nil
}
