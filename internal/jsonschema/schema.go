package jsonschema

import "encoding/json"

// Schema represents a JSON Schema object.
type Schema struct {
	Schema      string             `json:"$schema,omitempty"`
	Type        string             `json:"type,omitempty"`
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Default     interface{}        `json:"default,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`

	// Required is a pointer so an empty list is still emitted as
	// "required":[] (terraschema compatibility); a plain slice with
	// omitempty would drop the key entirely.
	Required *[]string `json:"required,omitempty"`

	// Items holds the element schemas of an array: one shared schema for
	// list/set, or one schema per tuple position.
	Items *Items `json:"items,omitempty"`

	// AdditionalProperties either allows extra properties freely or
	// constrains map values to a schema.
	AdditionalProperties *AdditionalProperties `json:"additionalProperties,omitempty"`
	MinLength            *int                  `json:"minLength,omitempty"`
	MaxLength            *int                  `json:"maxLength,omitempty"`
	MinItems             *int                  `json:"minItems,omitempty"`
	MaxItems             *int                  `json:"maxItems,omitempty"`
	MinProperties        *int                  `json:"minProperties,omitempty"`
	MaxProperties        *int                  `json:"maxProperties,omitempty"`
	Pattern              string                `json:"pattern,omitempty"`
	Minimum              *float64              `json:"minimum,omitempty"`
	Maximum              *float64              `json:"maximum,omitempty"`
	ExclusiveMinimum     *float64              `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     *float64              `json:"exclusiveMaximum,omitempty"`
	Enum                 []interface{}         `json:"enum,omitempty"`
	UniqueItems          *bool                 `json:"uniqueItems,omitempty"`
	Sensitive            *bool                 `json:"sensitive,omitempty"`
	AnyOf                []Schema              `json:"anyOf,omitempty"`
}

// Items is either a single element schema (list/set) or one schema per tuple
// position; whichever is set is emitted under the one "items" key.
type Items struct {
	Single *Schema
	Tuple  []*Schema
}

// MarshalJSON emits whichever variant is set under the single "items" key.
func (i *Items) MarshalJSON() ([]byte, error) {
	if i.Single != nil {
		return json.Marshal(i.Single)
	}
	return json.Marshal(i.Tuple)
}

// AdditionalProperties is either a blanket allow (bool) or the schema every
// map value must match; whichever is set is emitted under the one key.
type AdditionalProperties struct {
	AllowAny    bool
	ValueSchema *Schema
}

// MarshalJSON emits either the value schema or the boolean.
func (a *AdditionalProperties) MarshalJSON() ([]byte, error) {
	if a.ValueSchema != nil {
		return json.Marshal(a.ValueSchema)
	}
	return json.Marshal(a.AllowAny)
}
