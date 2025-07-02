package types

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// OptionalTypeConverter handles conversion of optional() function calls to JSON Schema
type OptionalTypeConverter struct {
	mainConverter TypeConverterWithIsOptional // Reference to main converter for recursive type conversion
}

// NewOptionalTypeConverter creates a new optional type converter
func NewOptionalTypeConverter(mainConverter TypeConverterWithIsOptional) *OptionalTypeConverter {
	return &OptionalTypeConverter{
		mainConverter: mainConverter,
	}
}

// Convert converts an optional(TYPE, default?) function call to JSON Schema
func (o *OptionalTypeConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
	funcExpr, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok || funcExpr.Name != "optional" {
		return nil, fmt.Errorf("expression is not an optional() function call")
	}

	if len(funcExpr.Args) < 1 || len(funcExpr.Args) > 2 {
		return nil, fmt.Errorf("optional() expects 1 or 2 arguments, got %d", len(funcExpr.Args))
	}

	// Convert the base type
	baseSchema, err := o.mainConverter.ConvertType(funcExpr.Args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert optional base type: %w", err)
	}

	// For optional types, we don't include default values in the JSON Schema
	// The default is handled at the Terraform level, not the JSON Schema level
	// This keeps the schema clean and focuses on structure rather than values

	// For JSON Schema, optional types are just the base type
	// The "optional" nature is handled by not including the field in the required array
	// This is managed at the object level, not the individual property level
	return baseSchema, nil
}
