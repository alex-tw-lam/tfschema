package converter

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
)

// DefaultAttributeApplier handles the 'default' attribute for variables.
type DefaultAttributeApplier struct {
	defaultParser *DefaultParser
}

// NewDefaultAttributeApplier creates a new DefaultAttributeApplier.
func NewDefaultAttributeApplier(parser *DefaultParser) *DefaultAttributeApplier {
	return &DefaultAttributeApplier{
		defaultParser: parser,
	}
}

// Apply inspects the 'default' attribute and sets the schema's Default field.
func (a *DefaultAttributeApplier) Apply(schema *jsonschema.Schema, attrs map[string]*hcl.Attribute) error {
	attr, exists := attrs["default"]
	if !exists || attr == nil {
		return nil // Attribute not present
	}

	defaultValue, err := a.defaultParser.ParseDefaultValue(attr.Expr)
	if err != nil {
		return fmt.Errorf("failed to parse default value: %w", err)
	}

	schema.Default = defaultValue
	return nil
}
