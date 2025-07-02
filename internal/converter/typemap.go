package converter

// Type expressions -> cty.Type (official typeexpr library) -> JSON Schema.

import (
	"fmt"
	"sort"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/zclconf/go-cty/cty"
)

// ConvertTypeExpr parses a type expression via the official typeexpr library
// and maps the resulting cty.Type onto a JSON Schema.
func ConvertTypeExpr(expr hcl.Expression) (*jsonschema.Schema, error) {
	ty, _, diags := typeexpr.TypeConstraintWithDefaults(expr)
	if diags.HasErrors() {
		return nil, fmt.Errorf("invalid type expression: %s", diags)
	}
	return ctyTypeToSchema(ty), nil
}

// ctyTypeToSchema recursively converts a cty.Type to the equivalent JSON Schema.
func ctyTypeToSchema(ty cty.Type) *jsonschema.Schema {
	switch {
	case ty == cty.DynamicPseudoType: // `any`: no constraints
		return &jsonschema.Schema{}
	case ty == cty.String:
		return &jsonschema.Schema{Type: "string"}
	case ty == cty.Number:
		return &jsonschema.Schema{Type: "number"}
	case ty == cty.Bool:
		return &jsonschema.Schema{Type: "boolean"}
	case ty.IsListType():
		return &jsonschema.Schema{Type: "array", Items: &jsonschema.Items{Single: ctyTypeToSchema(ty.ElementType())}}
	case ty.IsSetType():
		return &jsonschema.Schema{Type: "array", Items: &jsonschema.Items{Single: ctyTypeToSchema(ty.ElementType())}, UniqueItems: boolPtr(true)}
	case ty.IsMapType():
		return &jsonschema.Schema{Type: "object", AdditionalProperties: &jsonschema.AdditionalProperties{ValueSchema: ctyTypeToSchema(ty.ElementType())}}
	case ty.IsTupleType():
		return tupleSchema(ty)
	case ty.IsObjectType():
		return objectSchema(ty)
	}
	return &jsonschema.Schema{} // defensive fallback; unreachable via ConvertTypeExpr
}

// tupleSchema maps a tuple type: fixed-length array with per-position item schemas.
func tupleSchema(ty cty.Type) *jsonschema.Schema {
	elementTypes := ty.TupleElementTypes()
	items := make([]*jsonschema.Schema, 0, len(elementTypes))
	for _, elemType := range elementTypes {
		items = append(items, ctyTypeToSchema(elemType))
	}
	n := intPtr(len(items))
	return &jsonschema.Schema{Type: "array", Items: &jsonschema.Items{Tuple: items}, MinItems: n, MaxItems: n}
}

// objectSchema maps attributes to properties; non-optional ones become required.
func objectSchema(ty cty.Type) *jsonschema.Schema {
	schema := &jsonschema.Schema{
		Type:                 "object",
		Properties:           make(map[string]*jsonschema.Schema),
		Required:             &[]string{},                                      // pointer to empty list: "required":[] is still emitted (terraschema compatibility)
		AdditionalProperties: &jsonschema.AdditionalProperties{AllowAny: true}, // terraschema's permissive approach
	}
	for name, attrType := range ty.AttributeTypes() {
		schema.Properties[name] = ctyTypeToSchema(attrType)
		if !ty.AttributeOptional(name) {
			*schema.Required = append(*schema.Required, name)
		}
	}
	sort.Strings(*schema.Required)
	return schema
}

func boolPtr(b bool) *bool { return &b }

func intPtr(i int) *int { return &i }
