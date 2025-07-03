package types

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// SetConverter handles conversion of set() types.
type SetConverter struct {
	mainConverter TypeConverterWithIsOptional
}

// NewSetConverter creates a new set type converter.
func NewSetConverter(mainConverter TypeConverterWithIsOptional) *SetConverter {
	return &SetConverter{
		mainConverter: mainConverter,
	}
}

// Convert converts a set type expression to a JSON schema.
func (c *SetConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
	funcCall, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok {
		return nil, fmt.Errorf("set converter expects a function call expression")
	}

	if len(funcCall.Args) != 1 {
		return nil, fmt.Errorf("set() expects exactly one argument")
	}

	elemExpr := funcCall.Args[0]

	itemsSchema, err := c.mainConverter.ConvertType(elemExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert set element type: %w", err)
	}

	uniqueItems := true
	return &jsonschema.Schema{
		Type:        "array",
		Items:       itemsSchema,
		UniqueItems: &uniqueItems,
	}, nil
}
