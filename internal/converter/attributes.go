package converter

// Copies variable attributes (description/default/sensitive/nullable) onto the schema.

import (
	"encoding/json"
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// applyAttributes copies a variable block's description, default, sensitive
// and nullable attributes onto its JSON Schema.
func applyAttributes(schema *jsonschema.Schema, attrs map[string]*hcl.Attribute) error {
	for _, name := range []string{"description", "default", "sensitive", "nullable"} {
		attr, ok := attrs[name]
		if !ok {
			continue
		}
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return fmt.Errorf("failed to evaluate '%s' attribute: %w", name, diags)
		}
		switch name {
		case "description":
			schema.Description = val.AsString()
		case "default":
			value, err := ctyToGo(val)
			if err != nil {
				return err
			}
			schema.Default = value
		case "sensitive":
			if val.Type() == cty.Bool && val.True() {
				schema.Sensitive = boolPtr(true)
			}
		case "nullable":
			if val.Type() == cty.Bool && val.True() {
				applyNullableAnyOf(schema)
			}
		}
	}
	return nil
}

// applyNullableAnyOf rewrites a nullable variable as an anyOf choice between
// null and its type, matching terraschema's output.
func applyNullableAnyOf(schema *jsonschema.Schema) {
	typeName := schema.Type
	if typeName == "" {
		typeName = "string"
	}
	schema.Title = "Select a type"
	schema.Type = ""
	schema.AnyOf = []jsonschema.Schema{
		{Type: "null", Title: "null"},
		{Type: typeName, Title: typeName},
	}
}

// ctyToGo converts a cty value to plain Go values via the official ctyjson
// round-trip (numbers become float64, objects become map[string]interface{}).
func ctyToGo(val cty.Value) (interface{}, error) {
	if val.IsNull() || !val.IsKnown() {
		return nil, nil
	}
	data, err := ctyjson.Marshal(val, val.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to convert default value: %w", err)
	}
	var out interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("failed to convert default value: %w", err)
	}
	return out, nil
}
