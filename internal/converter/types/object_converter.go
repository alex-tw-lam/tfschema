package types

import (
	"fmt"
	"sort"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ObjectTypeConverter handles conversion of object() types
type ObjectTypeConverter struct {
	mainConverter TypeConverterWithIsOptional // Reference to main converter for recursive type conversion
}

// NewObjectTypeConverter creates a new object type converter
func NewObjectTypeConverter(mainConverter TypeConverterWithIsOptional) *ObjectTypeConverter {
	return &ObjectTypeConverter{
		mainConverter: mainConverter,
	}
}

// Convert converts an object() type expression to a JSON Schema
func (o *ObjectTypeConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
	funcExpr, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok {
		return nil, fmt.Errorf("expected function call expression for object type, got %T", expr)
	}

	if funcExpr.Name != "object" {
		return nil, fmt.Errorf("expected object() function call, got %s()", funcExpr.Name)
	}

	if len(funcExpr.Args) != 1 {
		return nil, fmt.Errorf("object() expects exactly one argument, got %d", len(funcExpr.Args))
	}

	objExpr, ok := funcExpr.Args[0].(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil, fmt.Errorf("argument to object() must be an object constructor, got %T", funcExpr.Args[0])
	}

	schema := &jsonschema.Schema{
		Type:                 "object",
		Properties:           make(map[string]*jsonschema.Schema),
		Required:             &[]string{},      // Initialize as pointer to empty slice
		AdditionalProperties: &[]bool{true}[0], // Follow terraschema's permissive approach
	}

	for _, item := range objExpr.Items {
		key, diags := item.KeyExpr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate object key: %s", diags.Error())
		}

		propSchema, err := o.mainConverter.ConvertType(item.ValueExpr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert property type for key '%s': %w", key.AsString(), err)
		}

		schema.Properties[key.AsString()] = propSchema
		// Only add to required array if the property is not optional
		if !o.mainConverter.IsOptionalType(item.ValueExpr) {
			*schema.Required = append(*schema.Required, key.AsString())
		}
	}

	// Sort required fields alphabetically (terraschema compatibility)
	sort.Strings(*schema.Required)

	return schema, nil
}
