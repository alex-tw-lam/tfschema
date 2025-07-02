package converter

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
)

// AttributeProcessor handles the processing of Terraform variable attributes.
type AttributeProcessor struct {
	appliers []AttributeApplier
}

// NewAttributeProcessor creates a new AttributeProcessor with its required appliers.
func NewAttributeProcessor(appliers ...AttributeApplier) *AttributeProcessor {
	return &AttributeProcessor{
		appliers: appliers,
	}
}

// ApplyAttributes processes all relevant attributes from a variable block and applies them to the schema.
func (p *AttributeProcessor) ApplyAttributes(schema *jsonschema.Schema, attrs map[string]*hcl.Attribute) error {
	for _, applier := range p.appliers {
		if err := applier.Apply(schema, attrs); err != nil {
			return fmt.Errorf("failed to apply attribute: %w", err)
		}
	}
	return nil
}
