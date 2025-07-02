package types

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// PrimitiveTypeConverter handles conversion of primitive types (string, number, bool)
type PrimitiveTypeConverter struct{}

// NewPrimitiveTypeConverter creates a new primitive type converter
func NewPrimitiveTypeConverter() *PrimitiveTypeConverter {
	return &PrimitiveTypeConverter{}
}

// Convert converts a primitive type expression to a JSON Schema
func (p *PrimitiveTypeConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
	trav, ok := expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return nil, fmt.Errorf("expression is not a primitive type traversal")
	}

	typeName := trav.Traversal.RootName()
	schema := &jsonschema.Schema{}
	switch typeName {
	case "string":
		schema.Type = "string"
	case "number":
		schema.Type = "number"
	case "bool":
		schema.Type = "boolean"
	default:
		return nil, fmt.Errorf("unsupported primitive type: %s", typeName)
	}
	return schema, nil
}
