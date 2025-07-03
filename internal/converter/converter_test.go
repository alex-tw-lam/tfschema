package converter

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertSchemasEqual(t *testing.T, expected, actual *jsonschema.Schema) {
	t.Helper()

	// Marshal both schemas to JSON
	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	require.NoError(t, err)

	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err)

	// Compare the JSON strings
	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}

func TestConvertStringWithLengthValidation(t *testing.T) {
	input := `
variable "string_with_length" {
  type = string
  validation {
    condition     = length(var.value) > 10
    error_message = "The length must be greater than 10."
  }
}`

	converter := New()
	schema, err := converter.ConvertString(input)
	require.NoError(t, err)

	expectedSchema := &jsonschema.Schema{
		Schema:               "http://json-schema.org/draft-07/schema#",
		Type:                 "object",
		AdditionalProperties: &[]bool{true}[0],
		Properties: map[string]*jsonschema.Schema{
			"string_with_length": {
				Type:      "string",
				MinLength: func() *int { i := 11; return &i }(),
			},
		},
		Required: &[]string{"string_with_length"},
	}

	assertSchemasEqual(t, expectedSchema, schema)
}

func TestConvertStringWithAllLengthValidations(t *testing.T) {
	tests := []struct {
		name           string
		condition      string
		expectedSchema func() *jsonschema.Schema
	}{
		{
			name:      "greater_than",
			condition: `length(var.value) > 10`,
			expectedSchema: func() *jsonschema.Schema {
				return &jsonschema.Schema{Type: "string", MinLength: func() *int { i := 11; return &i }()}
			},
		},
		{
			name:      "greater_than_or_equal",
			condition: `length(var.value) >= 10`,
			expectedSchema: func() *jsonschema.Schema {
				return &jsonschema.Schema{Type: "string", MinLength: func() *int { i := 10; return &i }()}
			},
		},
		{
			name:      "less_than",
			condition: `length(var.value) < 10`,
			expectedSchema: func() *jsonschema.Schema {
				return &jsonschema.Schema{Type: "string", MaxLength: func() *int { i := 9; return &i }()}
			},
		},
		{
			name:      "less_than_or_equal",
			condition: `length(var.value) <= 10`,
			expectedSchema: func() *jsonschema.Schema {
				return &jsonschema.Schema{Type: "string", MaxLength: func() *int { i := 10; return &i }()}
			},
		},
		{
			name:      "equal",
			condition: `length(var.value) == 10`,
			expectedSchema: func() *jsonschema.Schema {
				return &jsonschema.Schema{Type: "string", MinLength: func() *int { i := 10; return &i }(), MaxLength: func() *int { i := 10; return &i }()}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := fmt.Sprintf(`
variable "string_with_length_%s" {
  type = string
  validation {
    condition     = %s
    error_message = "The length validation failed."
  }
}`, tt.name, tt.condition)

			converter := New()
			schema, err := converter.ConvertString(input)
			require.NoError(t, err)

			expectedRootSchema := &jsonschema.Schema{
				Schema:               "http://json-schema.org/draft-07/schema#",
				Type:                 "object",
				AdditionalProperties: &[]bool{true}[0],
				Properties: map[string]*jsonschema.Schema{
					fmt.Sprintf("string_with_length_%s", tt.name): tt.expectedSchema(),
				},
				Required: &[]string{fmt.Sprintf("string_with_length_%s", tt.name)},
			}

			assertSchemasEqual(t, expectedRootSchema, schema)
		})
	}
}

func TestConvertStringWithRegexValidation(t *testing.T) {
	input := `
variable "string_with_regex" {
  type = string
  validation {
    condition     = can(regex("^[a-zA-Z0-9]*$", var.value))
    error_message = "The value must be alphanumeric."
  }
}`

	converter := New()
	schema, err := converter.ConvertString(input)
	require.NoError(t, err)

	expectedSchema := &jsonschema.Schema{
		Schema:               "http://json-schema.org/draft-07/schema#",
		Type:                 "object",
		AdditionalProperties: &[]bool{true}[0],
		Properties: map[string]*jsonschema.Schema{
			"string_with_regex": {
				Type:    "string",
				Pattern: "^[a-zA-Z0-9]*$",
			},
		},
		Required: &[]string{"string_with_regex"},
	}

	assertSchemasEqual(t, expectedSchema, schema)
}

func TestConvertStringWithEnumValidation(t *testing.T) {
	input := `
variable "string_with_enum" {
  type = string
  validation {
    condition     = contains(["a", "b", "c"], var.value)
    error_message = "The value must be one of a, b, or c."
  }
}`

	converter := New()
	schema, err := converter.ConvertString(input)
	require.NoError(t, err)

	expectedSchema := &jsonschema.Schema{
		Schema:               "http://json-schema.org/draft-07/schema#",
		Type:                 "object",
		AdditionalProperties: &[]bool{true}[0],
		Properties: map[string]*jsonschema.Schema{
			"string_with_enum": {
				Type: "string",
				Enum: []interface{}{"a", "b", "c"},
			},
		},
		Required: &[]string{"string_with_enum"},
	}

	assertSchemasEqual(t, expectedSchema, schema)
}

func TestConvertObject(t *testing.T) {
	input := `
variable "test_object" {
  type = object({
    name = string
    age  = number
  })
}`

	converter := New()
	schema, err := converter.ConvertString(input)
	require.NoError(t, err)

	expectedSchema := &jsonschema.Schema{
		Schema: "http://json-schema.org/draft-07/schema#",
		Type:   "object",
		Properties: map[string]*jsonschema.Schema{
			"test_object": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"age":  {Type: "number"},
					"name": {Type: "string"},
				},
				Required:             &[]string{"age", "name"},
				AdditionalProperties: &[]bool{true}[0],
			},
		},
		Required:             &[]string{"test_object"},
		AdditionalProperties: &[]bool{true}[0],
	}

	assertSchemasEqual(t, expectedSchema, schema)
}

func TestConvertObjectWithNestedValidation(t *testing.T) {
	input := `
variable "test_object_validated" {
  type = object({
    name = string
  })

  validation {
    condition     = length(var.test_object_validated.name) > 5
    error_message = "Name must be longer than 5 chars."
  }
}`

	converter := New()
	schema, err := converter.ConvertString(input)
	require.NoError(t, err)

	expectedSchema := &jsonschema.Schema{
		Schema: "http://json-schema.org/draft-07/schema#",
		Type:   "object",
		Properties: map[string]*jsonschema.Schema{
			"test_object_validated": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name": {
						Type:      "string",
						MinLength: &[]int{6}[0],
					},
				},
				Required:             &[]string{"name"},
				AdditionalProperties: &[]bool{true}[0],
			},
		},
		Required:             &[]string{"test_object_validated"},
		AdditionalProperties: &[]bool{true}[0],
	}

	assertSchemasEqual(t, expectedSchema, schema)
}
