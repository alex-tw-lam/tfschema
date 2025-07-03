package types

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ListTypeConverter handles conversion of list() types
type ListTypeConverter struct {
	mainConverter TypeConverterWithIsOptional // Reference to main converter for recursive type conversion
}

// NewListTypeConverter creates a new list type converter
func NewListTypeConverter(mainConverter TypeConverterWithIsOptional) *ListTypeConverter {
	return &ListTypeConverter{
		mainConverter: mainConverter,
	}
}

// Convert converts a list() type expression to a JSON Schema
func (l *ListTypeConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
	funcExpr, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok {
		return nil, fmt.Errorf("expected function call expression for list type, got %T", expr)
	}

	if funcExpr.Name != "list" {
		return nil, fmt.Errorf("expected list() function call, got %s()", funcExpr.Name)
	}

	if len(funcExpr.Args) != 1 {
		return nil, fmt.Errorf("list() expects exactly one argument, got %d", len(funcExpr.Args))
	}

	innerSchema, err := l.mainConverter.ConvertType(funcExpr.Args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert list element type: %w", err)
	}

	return &jsonschema.Schema{
		Type:  "array",
		Items: innerSchema,
	}, nil
}
