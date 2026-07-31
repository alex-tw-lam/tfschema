package converter

import (
	"sort"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
)

// TypeInferenceHandler infers a JSON Schema from a variable's default value
// when the variable declares no explicit type.
type TypeInferenceHandler struct{}

// NewTypeInferenceHandler creates a new TypeInferenceHandler.
func NewTypeInferenceHandler() *TypeInferenceHandler {
	return &TypeInferenceHandler{}
}

// InferSchemaFromDefault populates schema with the type and structure inferred
// from defaultValue. Attributes already applied to schema (description, default,
// sensitive, nullable) are preserved -- only type, properties, items and
// required are derived here.
//
// The default value is the Go representation produced by DefaultParser, i.e.:
//
//	string            -> string
//	float64           -> number   (cty numbers are decoded as float64)
//	bool              -> boolean
//	[]interface{}     -> array    (items inferred from the first element)
//	map[string]interface{} -> object (properties + required inferred per key)
//	nil               -> no constraints (null/unknown default)
func (t *TypeInferenceHandler) InferSchemaFromDefault(schema *jsonschema.Schema, defaultValue interface{}) *jsonschema.Schema {
	t.inferInto(schema, defaultValue)
	return schema
}

// inferInto derives JSON Schema type and structure from a parsed default value.
func (t *TypeInferenceHandler) inferInto(schema *jsonschema.Schema, v interface{}) {
	switch val := v.(type) {
	case nil:
		// null/unknown default: nothing to infer, schema stays unconstrained.
	case string:
		schema.Type = "string"
	case float64:
		schema.Type = "number"
	case bool:
		schema.Type = "boolean"
	case []interface{}:
		schema.Type = "array"
		// Infer the item schema from the first element. This assumes a
		// homogeneous list; a mixed-type default is approximated by its head.
		if len(val) > 0 {
			item := &jsonschema.Schema{}
			t.inferInto(item, val[0])
			schema.Items = item
		}
	case map[string]interface{}:
		schema.Type = "object"
		schema.Properties = make(map[string]*jsonschema.Schema, len(val))
		required := make([]string, 0, len(val))
		for key, elem := range val {
			prop := &jsonschema.Schema{}
			t.inferInto(prop, elem)
			schema.Properties[key] = prop
			required = append(required, key)
		}
		sort.Strings(required)
		schema.Required = &required
	default:
		// Unsupported value type: leave the schema unconstrained.
	}
}
