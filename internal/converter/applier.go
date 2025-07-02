package converter

import (
	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
)

// AttributeApplier applies a specific Terraform variable attribute to a JSON Schema.
type AttributeApplier interface {
	// Apply applies the attribute to the schema.
	// It's given the entire map of attributes from the variable block.
	Apply(schema *jsonschema.Schema, attributes map[string]*hcl.Attribute) error
}
