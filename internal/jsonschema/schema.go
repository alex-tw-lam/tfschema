package jsonschema

// Schema represents a JSON Schema object.
type Schema struct {
	Schema               string             `json:"$schema,omitempty"`
	Type                 interface{}        `json:"type,omitempty"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	Default              interface{}        `json:"default,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             *[]string          `json:"required,omitempty"`
	Items                interface{}        `json:"items,omitempty"` // Can be *Schema or []*Schema
	AdditionalProperties interface{}        `json:"additionalProperties,omitempty"`
	PrefixItems          []*Schema          `json:"prefixItems,omitempty"`
	AdditionalItems      *bool              `json:"additionalItems,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	MinItems             *int               `json:"minItems,omitempty"`
	MaxItems             *int               `json:"maxItems,omitempty"`
	MinProperties        *int               `json:"minProperties,omitempty"`
	MaxProperties        *int               `json:"maxProperties,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty"`
	ExclusiveMinimum     *float64           `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     *float64           `json:"exclusiveMaximum,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty"`
	UniqueItems          *bool              `json:"uniqueItems,omitempty"`
	Sensitive            *bool              `json:"sensitive,omitempty"`
	Nullable             *bool              `json:"nullable,omitempty"`
	AnyOf                []Schema           `json:"anyOf,omitempty"`
}
