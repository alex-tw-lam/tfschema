package types

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// MapTypeConverter handles conversion of map and set types to JSON Schema
type MapTypeConverter struct {
	mainConverter TypeConverterWithIsOptional // Reference to main converter for recursive type conversion
}

// NewMapTypeConverter creates a new map type converter
func NewMapTypeConverter(mainConverter TypeConverterWithIsOptional) *MapTypeConverter {
	return &MapTypeConverter{
		mainConverter: mainConverter,
	}
}

// Convert converts a map() or set() type expression to a JSON Schema
func (m *MapTypeConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
	funcExpr, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok {
		return nil, fmt.Errorf("expression is not a map() or set() function call")
	}

	return m.ConvertMapType(funcExpr)
}

// ConvertMapType converts a map(TYPE) function call to JSON Schema
func (m *MapTypeConverter) ConvertMapType(funcExpr *hclsyntax.FunctionCallExpr) (*jsonschema.Schema, error) {
	if len(funcExpr.Args) != 1 {
		return nil, fmt.Errorf("map() expects exactly one argument")
	}

	// Convert the value type
	valueSchema, err := m.mainConverter.ConvertType(funcExpr.Args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert map value type: %w", err)
	}

	// Map types in JSON Schema are objects with additionalProperties
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: valueSchema,
	}, nil
}
