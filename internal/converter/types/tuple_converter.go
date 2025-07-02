package types

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// TupleConverter handles conversion of tuple() types.
type TupleConverter struct {
	mainConverter TypeConverterWithIsOptional
}

// NewTupleConverter creates a new tuple type converter.
func NewTupleConverter(mainConverter TypeConverterWithIsOptional) *TupleConverter {
	return &TupleConverter{
		mainConverter: mainConverter,
	}
}

// Convert converts a tuple type expression to a JSON schema.
func (c *TupleConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
	funcCall, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok {
		return nil, fmt.Errorf("tuple converter expects a function call expression, got %T", expr)
	}

	if len(funcCall.Args) != 1 {
		return nil, fmt.Errorf("tuple() expects exactly one argument")
	}

	tupleExpr, ok := funcCall.Args[0].(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, fmt.Errorf("argument to tuple() must be a tuple constructor, got %T", funcCall.Args[0])
	}

	var items []*jsonschema.Schema
	for _, elemExpr := range tupleExpr.Exprs {
		itemSchema, err := c.mainConverter.ConvertType(elemExpr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tuple element type: %w", err)
		}
		items = append(items, itemSchema)
	}

	minMaxItems := len(items)
	schema := &jsonschema.Schema{
		Type:     "array",
		Items:    items,
		MinItems: &minMaxItems,
		MaxItems: &minMaxItems,
	}

	return schema, nil
}
